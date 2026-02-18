package service

import (
	"archive/zip"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grom-alex/homelib/backend/internal/config"
)

// --- Mock BookRepo ---

type mockBookRepo struct {
	archiveName   string
	fileInArchive string
	format        string
	err           error
}

func (m *mockBookRepo) GetBookForDownload(_ context.Context, _ int64) (string, string, string, error) {
	return m.archiveName, m.fileInArchive, m.format, m.err
}

// --- Helpers ---

const simpleFB2 = `<?xml version="1.0" encoding="UTF-8"?>
<FictionBook xmlns="http://www.gribuser.ru/xml/fictionbook/2.0">
  <description>
    <title-info>
      <author><first-name>Test</first-name><last-name>Author</last-name></author>
      <book-title>Test Book</book-title>
      <lang>en</lang>
    </title-info>
  </description>
  <body>
    <section>
      <title><p>Chapter One</p></title>
      <p>Hello world</p>
    </section>
  </body>
</FictionBook>`

func createTestArchive(t *testing.T, dir string, archiveName, fileName, content string) string {
	t.Helper()
	archivePath := filepath.Join(dir, archiveName)
	f, err := os.Create(archivePath)
	require.NoError(t, err)

	w := zip.NewWriter(f)
	fw, err := w.Create(fileName)
	require.NoError(t, err)
	_, err = fw.Write([]byte(content))
	require.NoError(t, err)
	require.NoError(t, w.Close())
	require.NoError(t, f.Close())

	return archivePath
}

func setupReaderService(t *testing.T, repo bookDownloadInfoProvider) (*ReaderService, string) {
	t.Helper()
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "cache")
	archivesDir := filepath.Join(tmpDir, "archives")
	require.NoError(t, os.MkdirAll(archivesDir, 0o755))

	svc := &ReaderService{
		bookRepo:  repo,
		libCfg:    config.LibraryConfig{ArchivesPath: archivesDir},
		cachePath: cacheDir,
		cacheTTL:  30 * 24 * time.Hour,
	}
	return svc, archivesDir
}

// --- Tests ---

func TestReaderService_GetBookContent_Success(t *testing.T) {
	repo := &mockBookRepo{
		archiveName:   "test.zip",
		fileInArchive: "book.fb2",
		format:        "fb2",
	}
	svc, archivesDir := setupReaderService(t, repo)
	createTestArchive(t, archivesDir, "test.zip", "book.fb2", simpleFB2)

	content, err := svc.GetBookContent(context.Background(), 1)
	require.NoError(t, err)

	assert.Equal(t, "Test Book", content.Metadata.Title)
	assert.Equal(t, "Test Author", content.Metadata.Author)
	assert.Equal(t, "en", content.Metadata.Language)
	assert.Equal(t, "fb2", content.Metadata.Format)
	assert.Equal(t, 1, content.TotalChapters)
	require.Len(t, content.ChapterIDs, 1)
}

func TestReaderService_GetBookContent_CacheHit(t *testing.T) {
	repo := &mockBookRepo{
		archiveName:   "test.zip",
		fileInArchive: "book.fb2",
		format:        "fb2",
	}
	svc, archivesDir := setupReaderService(t, repo)
	createTestArchive(t, archivesDir, "test.zip", "book.fb2", simpleFB2)

	// First call — cache miss, parses from archive
	content1, err := svc.GetBookContent(context.Background(), 1)
	require.NoError(t, err)

	// Verify cache file exists
	cachePath := filepath.Join(svc.cachePath, "1", "content.json")
	_, statErr := os.Stat(cachePath)
	assert.NoError(t, statErr, "cache file should exist")

	// Delete archive to ensure second call uses cache
	require.NoError(t, os.Remove(filepath.Join(archivesDir, "test.zip")))

	// Second call — cache hit
	content2, err := svc.GetBookContent(context.Background(), 1)
	require.NoError(t, err)

	assert.Equal(t, content1.Metadata.Title, content2.Metadata.Title)
	assert.Equal(t, content1.TotalChapters, content2.TotalChapters)
}

func TestReaderService_GetBookContent_BookNotFound(t *testing.T) {
	repo := &mockBookRepo{err: fmt.Errorf("no rows")}
	svc, _ := setupReaderService(t, repo)

	_, err := svc.GetBookContent(context.Background(), 999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "book not found")
}

func TestReaderService_GetBookContent_UnsupportedFormat(t *testing.T) {
	repo := &mockBookRepo{
		archiveName:   "test.zip",
		fileInArchive: "book.epub",
		format:        "epub",
	}
	svc, _ := setupReaderService(t, repo)

	_, err := svc.GetBookContent(context.Background(), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported format")
}

func TestReaderService_GetChapter_Success(t *testing.T) {
	repo := &mockBookRepo{
		archiveName:   "test.zip",
		fileInArchive: "book.fb2",
		format:        "fb2",
	}
	svc, archivesDir := setupReaderService(t, repo)
	createTestArchive(t, archivesDir, "test.zip", "book.fb2", simpleFB2)

	// Get content to learn chapter IDs
	content, err := svc.GetBookContent(context.Background(), 1)
	require.NoError(t, err)
	require.Len(t, content.ChapterIDs, 1)

	ch, err := svc.GetChapter(context.Background(), 1, content.ChapterIDs[0])
	require.NoError(t, err)

	assert.Equal(t, content.ChapterIDs[0], ch.ID)
	assert.Equal(t, "Chapter One", ch.Title)
	assert.Contains(t, ch.HTML, "Hello world")
}

func TestReaderService_GetChapter_CacheHit(t *testing.T) {
	repo := &mockBookRepo{
		archiveName:   "test.zip",
		fileInArchive: "book.fb2",
		format:        "fb2",
	}
	svc, archivesDir := setupReaderService(t, repo)
	createTestArchive(t, archivesDir, "test.zip", "book.fb2", simpleFB2)

	content, err := svc.GetBookContent(context.Background(), 1)
	require.NoError(t, err)
	chID := content.ChapterIDs[0]

	// First call — parse
	ch1, err := svc.GetChapter(context.Background(), 1, chID)
	require.NoError(t, err)

	// Delete archive
	require.NoError(t, os.Remove(filepath.Join(archivesDir, "test.zip")))

	// Second call — cache hit
	ch2, err := svc.GetChapter(context.Background(), 1, chID)
	require.NoError(t, err)

	assert.Equal(t, ch1.HTML, ch2.HTML)
}

func TestReaderService_GetChapter_NotFound(t *testing.T) {
	repo := &mockBookRepo{
		archiveName:   "test.zip",
		fileInArchive: "book.fb2",
		format:        "fb2",
	}
	svc, archivesDir := setupReaderService(t, repo)
	createTestArchive(t, archivesDir, "test.zip", "book.fb2", simpleFB2)

	_, err := svc.GetChapter(context.Background(), 1, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestReaderService_GetBookImage_Success(t *testing.T) {
	fb2WithImage := `<?xml version="1.0" encoding="UTF-8"?>
<FictionBook xmlns="http://www.gribuser.ru/xml/fictionbook/2.0">
  <description>
    <title-info>
      <book-title>Image Book</book-title>
      <lang>en</lang>
    </title-info>
  </description>
  <body>
    <section><p>Text</p></section>
  </body>
  <binary id="img1.png" content-type="image/png">iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==</binary>
</FictionBook>`

	repo := &mockBookRepo{
		archiveName:   "test.zip",
		fileInArchive: "book.fb2",
		format:        "fb2",
	}
	svc, archivesDir := setupReaderService(t, repo)
	createTestArchive(t, archivesDir, "test.zip", "book.fb2", fb2WithImage)

	img, err := svc.GetBookImage(context.Background(), 1, "img1.png")
	require.NoError(t, err)

	assert.Equal(t, "img1.png", img.ID)
	assert.Equal(t, "image/png", img.ContentType)
	assert.NotEmpty(t, img.Data)
	// PNG magic bytes
	assert.Equal(t, byte(0x89), img.Data[0])
}

func TestReaderService_GetBookImage_CacheHit(t *testing.T) {
	fb2WithImage := `<?xml version="1.0" encoding="UTF-8"?>
<FictionBook xmlns="http://www.gribuser.ru/xml/fictionbook/2.0">
  <description>
    <title-info><book-title>Image Book</book-title><lang>en</lang></title-info>
  </description>
  <body><section><p>Text</p></section></body>
  <binary id="img1.png" content-type="image/png">iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==</binary>
</FictionBook>`

	repo := &mockBookRepo{
		archiveName:   "test.zip",
		fileInArchive: "book.fb2",
		format:        "fb2",
	}
	svc, archivesDir := setupReaderService(t, repo)
	createTestArchive(t, archivesDir, "test.zip", "book.fb2", fb2WithImage)

	// First call
	img1, err := svc.GetBookImage(context.Background(), 1, "img1.png")
	require.NoError(t, err)

	// Delete archive
	require.NoError(t, os.Remove(filepath.Join(archivesDir, "test.zip")))

	// Second call — cache hit
	img2, err := svc.GetBookImage(context.Background(), 1, "img1.png")
	require.NoError(t, err)

	assert.Equal(t, img1.ContentType, img2.ContentType)
	assert.Equal(t, img1.Data, img2.Data)
}

func TestReaderService_GetBookImage_NotFound(t *testing.T) {
	repo := &mockBookRepo{
		archiveName:   "test.zip",
		fileInArchive: "book.fb2",
		format:        "fb2",
	}
	svc, archivesDir := setupReaderService(t, repo)
	createTestArchive(t, archivesDir, "test.zip", "book.fb2", simpleFB2)

	_, err := svc.GetBookImage(context.Background(), 1, "nonexistent.png")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestReaderService_MalformedFB2(t *testing.T) {
	malformedFB2 := `<?xml version="1.0"?><FictionBook><body><section><p>unclosed`

	repo := &mockBookRepo{
		archiveName:   "test.zip",
		fileInArchive: "book.fb2",
		format:        "fb2",
	}
	svc, archivesDir := setupReaderService(t, repo)
	createTestArchive(t, archivesDir, "test.zip", "book.fb2", malformedFB2)

	_, err := svc.GetBookContent(context.Background(), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parse book")
}

func TestReaderService_CacheDir_Structure(t *testing.T) {
	repo := &mockBookRepo{
		archiveName:   "test.zip",
		fileInArchive: "book.fb2",
		format:        "fb2",
	}
	svc, archivesDir := setupReaderService(t, repo)
	createTestArchive(t, archivesDir, "test.zip", "book.fb2", simpleFB2)

	_, err := svc.GetBookContent(context.Background(), 42)
	require.NoError(t, err)

	// Cache dir: {cachePath}/42/
	expectedDir := filepath.Join(svc.cachePath, "42")
	info, statErr := os.Stat(expectedDir)
	require.NoError(t, statErr)
	assert.True(t, info.IsDir())

	// content.json exists
	_, statErr = os.Stat(filepath.Join(expectedDir, "content.json"))
	assert.NoError(t, statErr)
}

func TestNewReaderService(t *testing.T) {
	svc := NewReaderService(nil, config.LibraryConfig{}, config.ReaderConfig{
		CachePath: "/tmp/test",
		CacheTTL:  48 * time.Hour,
	})
	assert.Equal(t, "/tmp/test", svc.cachePath)
	assert.Equal(t, 48*time.Hour, svc.cacheTTL)
}

// --- Test bookfile.GetConverter via ReaderService ---

func TestReaderService_GetBookContent_CachesBookContent(t *testing.T) {
	repo := &mockBookRepo{
		archiveName:   "test.zip",
		fileInArchive: "book.fb2",
		format:        "fb2",
	}
	svc, archivesDir := setupReaderService(t, repo)
	createTestArchive(t, archivesDir, "test.zip", "book.fb2", simpleFB2)

	// Call once to populate cache
	content, err := svc.GetBookContent(context.Background(), 1)
	require.NoError(t, err)
	require.NotNil(t, content)

	// Verify cached content is valid JSON by reading it
	cached, err := svc.getCachedContent(1)
	require.NoError(t, err)
	assert.Equal(t, content.Metadata.Title, cached.Metadata.Title)
}

func TestReaderService_GetChapter_CachesChapterContent(t *testing.T) {
	repo := &mockBookRepo{
		archiveName:   "test.zip",
		fileInArchive: "book.fb2",
		format:        "fb2",
	}
	svc, archivesDir := setupReaderService(t, repo)
	createTestArchive(t, archivesDir, "test.zip", "book.fb2", simpleFB2)

	content, err := svc.GetBookContent(context.Background(), 1)
	require.NoError(t, err)

	ch, err := svc.GetChapter(context.Background(), 1, content.ChapterIDs[0])
	require.NoError(t, err)

	cached, cacheErr := svc.getCachedChapter(1, content.ChapterIDs[0])
	require.NoError(t, cacheErr)
	assert.Equal(t, ch.HTML, cached.HTML)
}

// --- Cache cleanup tests ---

func TestReaderService_CleanupExpiredCache_RemovesOld(t *testing.T) {
	repo := &mockBookRepo{
		archiveName:   "test.zip",
		fileInArchive: "book.fb2",
		format:        "fb2",
	}
	svc, archivesDir := setupReaderService(t, repo)
	svc.cacheTTL = 1 * time.Hour
	createTestArchive(t, archivesDir, "test.zip", "book.fb2", simpleFB2)

	// Populate cache for two books
	_, err := svc.GetBookContent(context.Background(), 1)
	require.NoError(t, err)
	_, err = svc.GetBookContent(context.Background(), 2)
	require.NoError(t, err)

	// Set book 1 cache dir mtime to 2 hours ago (expired)
	oldTime := time.Now().Add(-2 * time.Hour)
	require.NoError(t, os.Chtimes(svc.bookCacheDir(1), oldTime, oldTime))

	// Book 2 stays fresh (just created)

	removed, err := svc.CleanupExpiredCache()
	require.NoError(t, err)
	assert.Equal(t, 1, removed)

	// Book 1 cache should be gone
	_, statErr := os.Stat(svc.bookCacheDir(1))
	assert.True(t, os.IsNotExist(statErr))

	// Book 2 cache should remain
	_, statErr = os.Stat(svc.bookCacheDir(2))
	assert.NoError(t, statErr)
}

func TestReaderService_CleanupExpiredCache_KeepsFresh(t *testing.T) {
	repo := &mockBookRepo{
		archiveName:   "test.zip",
		fileInArchive: "book.fb2",
		format:        "fb2",
	}
	svc, archivesDir := setupReaderService(t, repo)
	svc.cacheTTL = 1 * time.Hour
	createTestArchive(t, archivesDir, "test.zip", "book.fb2", simpleFB2)

	_, err := svc.GetBookContent(context.Background(), 1)
	require.NoError(t, err)

	// Everything is fresh — nothing should be removed
	removed, err := svc.CleanupExpiredCache()
	require.NoError(t, err)
	assert.Equal(t, 0, removed)

	_, statErr := os.Stat(svc.bookCacheDir(1))
	assert.NoError(t, statErr)
}

func TestReaderService_CleanupExpiredCache_ZeroTTL(t *testing.T) {
	svc := &ReaderService{cachePath: t.TempDir(), cacheTTL: 0}

	removed, err := svc.CleanupExpiredCache()
	require.NoError(t, err)
	assert.Equal(t, 0, removed)
}

func TestReaderService_CleanupExpiredCache_NoCacheDir(t *testing.T) {
	svc := &ReaderService{cachePath: "/nonexistent/path", cacheTTL: 1 * time.Hour}

	removed, err := svc.CleanupExpiredCache()
	require.NoError(t, err)
	assert.Equal(t, 0, removed)
}

func TestReaderService_TouchCache_UpdatesMtime(t *testing.T) {
	repo := &mockBookRepo{
		archiveName:   "test.zip",
		fileInArchive: "book.fb2",
		format:        "fb2",
	}
	svc, archivesDir := setupReaderService(t, repo)
	createTestArchive(t, archivesDir, "test.zip", "book.fb2", simpleFB2)

	// Populate cache
	_, err := svc.GetBookContent(context.Background(), 1)
	require.NoError(t, err)

	// Set mtime to the past
	oldTime := time.Now().Add(-48 * time.Hour)
	require.NoError(t, os.Chtimes(svc.bookCacheDir(1), oldTime, oldTime))

	// Touch should update mtime
	svc.touchCache(1)

	info, err := os.Stat(svc.bookCacheDir(1))
	require.NoError(t, err)
	assert.True(t, info.ModTime().After(oldTime))
}
