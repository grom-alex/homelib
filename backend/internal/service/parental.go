package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/sync/singleflight"

	"github.com/grom-alex/homelib/backend/internal/models"
)

const (
	metaKeyRestrictedGenreCodes = "restricted_genre_codes"
	metaKeyParentalPinHash      = "parental_pin_hash"
	restrictedIDsCacheTTL       = 5 * time.Minute
)

// ParentalGenreRepo abstracts genre repo methods needed by ParentalService.
type ParentalGenreRepo interface {
	GetIDsByCodes(ctx context.Context, codes []string) (map[string][]int, error)
	GetDescendantIDs(ctx context.Context, parentIDs []int) ([]int, error)
}

// ParentalUserRepo abstracts user repo methods needed by ParentalService.
type ParentalUserRepo interface {
	GetSettings(ctx context.Context, userID string) (json.RawMessage, error)
	UpdateSettings(ctx context.Context, userID string, patch json.RawMessage) (json.RawMessage, error)
	ListUsersWithAdultStatus(ctx context.Context) ([]models.UserAdultStatus, error)
}

// ParentalService manages parental controls: restricted genres, PIN, and per-user adult content access.
type ParentalService struct {
	meta      MetadataStore
	genreRepo ParentalGenreRepo
	userRepo  ParentalUserRepo

	// In-memory cache for resolved restricted genre IDs
	cacheMu   sync.RWMutex
	cachedIDs []int
	cacheTime time.Time

	// Prevents thundering herd on cache miss
	sfGroup singleflight.Group
}

func NewParentalService(meta MetadataStore, genreRepo ParentalGenreRepo, userRepo ParentalUserRepo) *ParentalService {
	return &ParentalService{
		meta:      meta,
		genreRepo: genreRepo,
		userRepo:  userRepo,
	}
}

// GetRestrictedGenreCodes returns the list of restricted genre codes from app_metadata.
func (s *ParentalService) GetRestrictedGenreCodes(ctx context.Context) ([]string, error) {
	val, err := s.meta.Get(ctx, metaKeyRestrictedGenreCodes)
	if err != nil {
		return nil, err
	}
	if val == "" {
		return nil, nil
	}

	var codes []string
	if err := json.Unmarshal([]byte(val), &codes); err != nil {
		return nil, fmt.Errorf("parse restricted genre codes: %w", err)
	}
	return codes, nil
}

// SetRestrictedGenreCodes updates the list of restricted genre codes and invalidates the cache.
func (s *ParentalService) SetRestrictedGenreCodes(ctx context.Context, codes []string) error {
	data, err := json.Marshal(codes)
	if err != nil {
		return fmt.Errorf("marshal restricted codes: %w", err)
	}
	if err := s.meta.Set(ctx, metaKeyRestrictedGenreCodes, string(data)); err != nil {
		return err
	}
	s.InvalidateCache()
	return nil
}

// SetPin sets the parental PIN (stored as bcrypt hash in app_metadata).
func (s *ParentalService) SetPin(ctx context.Context, pin string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(pin), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash pin: %w", err)
	}
	return s.meta.Set(ctx, metaKeyParentalPinHash, string(hash))
}

// VerifyPin checks whether the provided PIN matches the stored hash.
func (s *ParentalService) VerifyPin(ctx context.Context, pin string) (bool, error) {
	hash, err := s.meta.Get(ctx, metaKeyParentalPinHash)
	if err != nil {
		return false, err
	}
	if hash == "" {
		return false, nil // no PIN set
	}
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(pin))
	if err != nil {
		return false, nil // wrong PIN
	}
	return true, nil
}

// RemovePin deletes the parental PIN.
func (s *ParentalService) RemovePin(ctx context.Context) error {
	return s.meta.Set(ctx, metaKeyParentalPinHash, "")
}

// IsPinSet checks whether a parental PIN has been configured.
func (s *ParentalService) IsPinSet(ctx context.Context) (bool, error) {
	hash, err := s.meta.Get(ctx, metaKeyParentalPinHash)
	if err != nil {
		return false, err
	}
	return hash != "", nil
}

// GetRestrictedGenreIDs resolves restricted genre codes to IDs, including all descendants.
// Results are cached in memory for restrictedIDsCacheTTL.
// Uses singleflight to prevent thundering herd on cache miss.
func (s *ParentalService) GetRestrictedGenreIDs(ctx context.Context) ([]int, error) {
	s.cacheMu.RLock()
	if s.cachedIDs != nil && time.Since(s.cacheTime) < restrictedIDsCacheTTL {
		ids := make([]int, len(s.cachedIDs))
		copy(ids, s.cachedIDs)
		s.cacheMu.RUnlock()
		return ids, nil
	}
	s.cacheMu.RUnlock()

	// Singleflight: only one goroutine resolves IDs on cache miss
	val, err, _ := s.sfGroup.Do("restricted_ids", func() (interface{}, error) {
		// Double-check under singleflight (another goroutine may have populated)
		s.cacheMu.RLock()
		if s.cachedIDs != nil && time.Since(s.cacheTime) < restrictedIDsCacheTTL {
			ids := make([]int, len(s.cachedIDs))
			copy(ids, s.cachedIDs)
			s.cacheMu.RUnlock()
			return ids, nil
		}
		s.cacheMu.RUnlock()

		return s.resolveRestrictedIDs(ctx)
	})
	if err != nil {
		return nil, err
	}
	ids, _ := val.([]int)
	return ids, nil
}

// resolveRestrictedIDs performs the actual resolution of restricted genre codes to IDs.
func (s *ParentalService) resolveRestrictedIDs(ctx context.Context) ([]int, error) {
	codes, err := s.GetRestrictedGenreCodes(ctx)
	if err != nil {
		return nil, err
	}
	if len(codes) == 0 {
		s.cacheMu.Lock()
		s.cachedIDs = []int{}
		s.cacheTime = time.Now()
		s.cacheMu.Unlock()
		return nil, nil
	}

	// Resolve codes → IDs
	codeToIDs, err := s.genreRepo.GetIDsByCodes(ctx, codes)
	if err != nil {
		return nil, fmt.Errorf("resolve genre codes: %w", err)
	}

	// Collect all direct IDs
	idSet := make(map[int]bool)
	var directIDs []int
	for _, ids := range codeToIDs {
		for _, id := range ids {
			if !idSet[id] {
				idSet[id] = true
				directIDs = append(directIDs, id)
			}
		}
	}

	// Get all descendants
	descendantIDs, err := s.genreRepo.GetDescendantIDs(ctx, directIDs)
	if err != nil {
		return nil, fmt.Errorf("resolve descendant genres: %w", err)
	}
	for _, id := range descendantIDs {
		if !idSet[id] {
			idSet[id] = true
			directIDs = append(directIDs, id)
		}
	}

	// Cache result
	s.cacheMu.Lock()
	s.cachedIDs = directIDs
	s.cacheTime = time.Now()
	s.cacheMu.Unlock()

	return directIDs, nil
}

// InvalidateCache clears the cached restricted genre IDs.
func (s *ParentalService) InvalidateCache() {
	s.cacheMu.Lock()
	s.cachedIDs = nil
	s.cacheTime = time.Time{}
	s.cacheMu.Unlock()
}

// IsAdultContentEnabled checks whether a user has adult content enabled in their settings.
func (s *ParentalService) IsAdultContentEnabled(ctx context.Context, userID string) (bool, error) {
	settings, err := s.userRepo.GetSettings(ctx, userID)
	if err != nil {
		return false, err
	}
	if len(settings) == 0 {
		return false, nil
	}

	var parsed struct {
		Parental struct {
			AdultContentEnabled bool `json:"adult_content_enabled"`
		} `json:"parental"`
	}
	if err := json.Unmarshal(settings, &parsed); err != nil {
		return false, nil // corrupted settings → default to restricted
	}
	return parsed.Parental.AdultContentEnabled, nil
}

// SetAdultContentEnabled updates the adult_content_enabled flag in user settings.
func (s *ParentalService) SetAdultContentEnabled(ctx context.Context, userID string, enabled bool) error {
	patch := fmt.Sprintf(`{"parental":{"adult_content_enabled":%v}}`, enabled)
	_, err := s.userRepo.UpdateSettings(ctx, userID, json.RawMessage(patch))
	return err
}

// ListUsersAdultStatus returns all users with their adult content access status.
func (s *ParentalService) ListUsersAdultStatus(ctx context.Context) ([]models.UserAdultStatus, error) {
	return s.userRepo.ListUsersWithAdultStatus(ctx)
}
