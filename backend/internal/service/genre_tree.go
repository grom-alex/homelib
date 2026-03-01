package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"sync"

	"github.com/grom-alex/homelib/backend/internal/glst"
)

const metaKeyGenreTreeHash = "genre_tree_hash"

// GenreTreeResult contains the outcome of a genre tree load operation.
type GenreTreeResult struct {
	GenresLoaded  int      `json:"genres_loaded"`
	BooksRemapped int      `json:"books_remapped"`
	Warnings      []string `json:"warnings"`
	Skipped       bool     `json:"skipped"`
}

// MetadataStore abstracts app_metadata access for testability.
type MetadataStore interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string) error
}

// GenreTreeRepository abstracts genre repo methods needed by GenreTreeService.
type GenreTreeRepository interface {
	LoadTree(ctx context.Context, entries []glst.GenreEntry) (int, error)
	GetIDsByCodes(ctx context.Context, codes []string) (map[string][]int, error)
	GetUnsortedGenreID(ctx context.Context) (int, error)
}

// BookRemapper abstracts book_genres remapping for testability.
type BookRemapper interface {
	RemapBookGenres(ctx context.Context, codeToIDs map[string][]int, fallbackID int, batchSize int) (int, error)
}

// GenreTreeService handles loading and reloading the genre tree from .glst data.
type GenreTreeService struct {
	genreData []byte
	meta      MetadataStore
	genreRepo GenreTreeRepository
	bookRepo  BookRemapper

	mu      sync.Mutex
	running bool
}

var ErrGenreReloadAlreadyRunning = fmt.Errorf("genre reload is already running")

func NewGenreTreeService(
	genreData []byte,
	meta MetadataStore,
	genreRepo GenreTreeRepository,
	bookRepo BookRemapper,
) *GenreTreeService {
	return &GenreTreeService{
		genreData: genreData,
		meta:      meta,
		genreRepo: genreRepo,
		bookRepo:  bookRepo,
	}
}

// LoadIfNeeded loads the genre tree only if the .glst data hash has changed.
func (s *GenreTreeService) LoadIfNeeded(ctx context.Context) (*GenreTreeResult, error) {
	return s.load(ctx, false)
}

// ForceReload loads the genre tree regardless of hash.
func (s *GenreTreeService) ForceReload(ctx context.Context) (*GenreTreeResult, error) {
	return s.load(ctx, true)
}

func (s *GenreTreeService) load(ctx context.Context, force bool) (*GenreTreeResult, error) {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return nil, ErrGenreReloadAlreadyRunning
	}
	s.running = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
	}()

	currentHash := computeSHA256(s.genreData)

	if !force {
		storedHash, err := s.meta.Get(ctx, metaKeyGenreTreeHash)
		if err != nil {
			return nil, fmt.Errorf("get genre tree hash: %w", err)
		}
		if storedHash == currentHash {
			return &GenreTreeResult{Skipped: true}, nil
		}
	}

	// Parse GLST data
	parsed, err := glst.ParseReader(bytes.NewReader(s.genreData))
	if err != nil {
		return nil, fmt.Errorf("parse genre file: %w", err)
	}

	result := &GenreTreeResult{
		Warnings: parsed.Warnings,
	}

	// Load tree into DB
	loaded, err := s.genreRepo.LoadTree(ctx, parsed.Entries)
	if err != nil {
		return result, fmt.Errorf("load genre tree: %w", err)
	}
	result.GenresLoaded = loaded
	log.Printf("Genre tree loaded: %d genres, %d warnings", loaded, len(parsed.Warnings))

	// Remap book_genres
	codeToIDs, err := s.genreRepo.GetIDsByCodes(ctx, s.collectCodes(parsed.Entries))
	if err != nil {
		return result, fmt.Errorf("get genre IDs by codes: %w", err)
	}

	fallbackID, err := s.genreRepo.GetUnsortedGenreID(ctx)
	if err != nil {
		return result, fmt.Errorf("get unsorted genre ID: %w", err)
	}

	remapped, err := s.bookRepo.RemapBookGenres(ctx, codeToIDs, fallbackID, 3000)
	if err != nil {
		return result, fmt.Errorf("remap book genres: %w", err)
	}
	result.BooksRemapped = remapped

	// Update hash
	if err := s.meta.Set(ctx, metaKeyGenreTreeHash, currentHash); err != nil {
		return result, fmt.Errorf("set genre tree hash: %w", err)
	}

	return result, nil
}

func (s *GenreTreeService) collectCodes(entries []glst.GenreEntry) []string {
	seen := make(map[string]bool, len(entries))
	var codes []string
	for _, e := range entries {
		if e.Code != "" && !seen[e.Code] {
			seen[e.Code] = true
			codes = append(codes, e.Code)
		}
	}
	return codes
}

func computeSHA256(data []byte) string {
	h := sha256.Sum256(data)
	return fmt.Sprintf("%x", h[:])
}
