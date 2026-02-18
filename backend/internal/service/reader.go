package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/grom-alex/homelib/backend/internal/archive"
	"github.com/grom-alex/homelib/backend/internal/bookfile"
	"github.com/grom-alex/homelib/backend/internal/config"
	"github.com/grom-alex/homelib/backend/internal/repository"
)

// bookDownloadInfoProvider abstracts the book repo dependency for testing.
type bookDownloadInfoProvider interface {
	GetBookForDownload(ctx context.Context, id int64) (archiveName, fileInArchive, format string, err error)
}

type ReaderService struct {
	bookRepo  bookDownloadInfoProvider
	libCfg    config.LibraryConfig
	cachePath string
	cacheTTL  time.Duration
}

func NewReaderService(bookRepo *repository.BookRepo, libCfg config.LibraryConfig, readerCfg config.ReaderConfig) *ReaderService {
	return &ReaderService{
		bookRepo:  bookRepo,
		libCfg:    libCfg,
		cachePath: readerCfg.CachePath,
		cacheTTL:  readerCfg.CacheTTL,
	}
}

// GetBookContent returns the book metadata and structure. Uses file cache.
func (s *ReaderService) GetBookContent(ctx context.Context, bookID int64) (*bookfile.BookContent, error) {
	// Try cache
	cached, err := s.getCachedContent(bookID)
	if err == nil {
		s.touchCache(bookID)
		return cached, nil
	}

	// Parse book
	conv, err := s.parseBook(ctx, bookID)
	if err != nil {
		return nil, err
	}

	content := conv.Content()

	// Cache content
	_ = s.cacheContent(bookID, content)

	return content, nil
}

// GetChapter returns the HTML content of a specific chapter. Uses file cache.
func (s *ReaderService) GetChapter(ctx context.Context, bookID int64, chapterID string) (*bookfile.ChapterContent, error) {
	// Try cache
	cached, err := s.getCachedChapter(bookID, chapterID)
	if err == nil {
		s.touchCache(bookID)
		return cached, nil
	}

	// Parse book
	conv, err := s.parseBook(ctx, bookID)
	if err != nil {
		return nil, err
	}

	ch, err := conv.Chapter(chapterID)
	if err != nil {
		return nil, fmt.Errorf("chapter %q: %w", chapterID, err)
	}

	// Cache chapter
	_ = s.cacheChapter(bookID, ch)

	return ch, nil
}

// GetBookImage returns an embedded image from the book. Uses file cache.
func (s *ReaderService) GetBookImage(ctx context.Context, bookID int64, imageID string) (*bookfile.ImageData, error) {
	// Try cache
	cached, err := s.getCachedImage(bookID, imageID)
	if err == nil {
		s.touchCache(bookID)
		return cached, nil
	}

	// Parse book
	conv, err := s.parseBook(ctx, bookID)
	if err != nil {
		return nil, err
	}

	img, err := conv.Image(imageID)
	if err != nil {
		return nil, fmt.Errorf("image %q: %w", imageID, err)
	}

	// Cache image
	_ = s.cacheImage(bookID, img)

	return img, nil
}

// parseBook extracts the book from archive and parses it.
func (s *ReaderService) parseBook(ctx context.Context, bookID int64) (bookfile.BookConverter, error) {
	archiveName, fileInArchive, format, err := s.bookRepo.GetBookForDownload(ctx, bookID)
	if err != nil {
		return nil, fmt.Errorf("book not found: %w", err)
	}

	conv, err := bookfile.GetConverter(format)
	if err != nil {
		return nil, err
	}

	archivePath := filepath.Join(s.libCfg.ArchivesPath, archiveName)

	// Path traversal prevention
	absBasePath, err := filepath.EvalSymlinks(s.libCfg.ArchivesPath)
	if err != nil {
		return nil, fmt.Errorf("invalid base path: %w", err)
	}
	absArchivePath, err := filepath.EvalSymlinks(archivePath)
	if err != nil {
		return nil, fmt.Errorf("book not found: invalid archive path")
	}
	if !strings.HasPrefix(absArchivePath, absBasePath+string(filepath.Separator)) && absArchivePath != absBasePath {
		return nil, fmt.Errorf("book not found: invalid archive path")
	}

	rc, _, err := archive.ExtractFile(archivePath, fileInArchive)
	if err != nil {
		return nil, fmt.Errorf("extract file: %w", err)
	}
	defer func() { _ = rc.Close() }()

	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	if err := conv.Parse(data, bookID); err != nil {
		return nil, fmt.Errorf("parse book: %w", err)
	}

	return conv, nil
}

// --- File cache ---

func (s *ReaderService) bookCacheDir(bookID int64) string {
	return filepath.Join(s.cachePath, fmt.Sprintf("%d", bookID))
}

func (s *ReaderService) ensureCacheDir(bookID int64) error {
	return os.MkdirAll(s.bookCacheDir(bookID), 0o755)
}

// Content cache

func (s *ReaderService) getCachedContent(bookID int64) (*bookfile.BookContent, error) {
	path := filepath.Join(s.bookCacheDir(bookID), "content.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var content bookfile.BookContent
	if err := json.Unmarshal(data, &content); err != nil {
		return nil, err
	}
	return &content, nil
}

func (s *ReaderService) cacheContent(bookID int64, content *bookfile.BookContent) error {
	if err := s.ensureCacheDir(bookID); err != nil {
		return err
	}
	data, err := json.Marshal(content)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(s.bookCacheDir(bookID), "content.json"), data, 0o644)
}

// Chapter cache

func (s *ReaderService) getCachedChapter(bookID int64, chapterID string) (*bookfile.ChapterContent, error) {
	path := filepath.Join(s.bookCacheDir(bookID), fmt.Sprintf("ch_%s.html", chapterID))
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var ch bookfile.ChapterContent
	if err := json.Unmarshal(data, &ch); err != nil {
		return nil, err
	}
	return &ch, nil
}

func (s *ReaderService) cacheChapter(bookID int64, ch *bookfile.ChapterContent) error {
	if err := s.ensureCacheDir(bookID); err != nil {
		return err
	}
	data, err := json.Marshal(ch)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(s.bookCacheDir(bookID), fmt.Sprintf("ch_%s.html", ch.ID)), data, 0o644)
}

// Image cache

func (s *ReaderService) getCachedImage(bookID int64, imageID string) (*bookfile.ImageData, error) {
	dir := s.bookCacheDir(bookID)

	// Read metadata
	metaPath := filepath.Join(dir, fmt.Sprintf("img_%s.meta", imageID))
	contentType, err := os.ReadFile(metaPath)
	if err != nil {
		return nil, err
	}

	// Read binary
	dataPath := filepath.Join(dir, fmt.Sprintf("img_%s.bin", imageID))
	data, err := os.ReadFile(dataPath)
	if err != nil {
		return nil, err
	}

	return &bookfile.ImageData{
		ID:          imageID,
		ContentType: string(contentType),
		Data:        data,
	}, nil
}

func (s *ReaderService) cacheImage(bookID int64, img *bookfile.ImageData) error {
	if err := s.ensureCacheDir(bookID); err != nil {
		return err
	}
	dir := s.bookCacheDir(bookID)

	// Write metadata
	metaPath := filepath.Join(dir, fmt.Sprintf("img_%s.meta", img.ID))
	if err := os.WriteFile(metaPath, []byte(img.ContentType), 0o644); err != nil {
		return err
	}

	// Write binary
	dataPath := filepath.Join(dir, fmt.Sprintf("img_%s.bin", img.ID))
	return os.WriteFile(dataPath, img.Data, 0o644)
}

// touchCache updates the modification time of the book cache directory on access.
func (s *ReaderService) touchCache(bookID int64) {
	dir := s.bookCacheDir(bookID)
	now := time.Now()
	_ = os.Chtimes(dir, now, now)
}

// CleanupExpiredCache removes book cache directories not accessed within cacheTTL.
func (s *ReaderService) CleanupExpiredCache() (int, error) {
	if s.cacheTTL <= 0 {
		return 0, nil
	}

	entries, err := os.ReadDir(s.cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("read cache dir: %w", err)
	}

	cutoff := time.Now().Add(-s.cacheTTL)
	removed := 0

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			dirPath := filepath.Join(s.cachePath, entry.Name())
			if err := os.RemoveAll(dirPath); err != nil {
				log.Printf("cache cleanup: failed to remove %s: %v", dirPath, err)
				continue
			}
			removed++
		}
	}

	return removed, nil
}

// StartCacheCleanup runs periodic cache cleanup in a background goroutine.
// It stops when the context is cancelled.
func (s *ReaderService) StartCacheCleanup(ctx context.Context) {
	if s.cacheTTL <= 0 {
		return
	}

	// Run cleanup once on startup
	if removed, err := s.CleanupExpiredCache(); err != nil {
		log.Printf("cache cleanup error: %v", err)
	} else if removed > 0 {
		log.Printf("cache cleanup: removed %d expired entries", removed)
	}

	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if removed, err := s.CleanupExpiredCache(); err != nil {
					log.Printf("cache cleanup error: %v", err)
				} else if removed > 0 {
					log.Printf("cache cleanup: removed %d expired entries", removed)
				}
			}
		}
	}()
}
