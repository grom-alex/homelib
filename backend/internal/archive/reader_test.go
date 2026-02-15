package archive

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetContentType(t *testing.T) {
	tests := []struct {
		ext      string
		expected string
	}{
		{"fb2", "application/x-fictionbook+xml"},
		{"epub", "application/epub+zip"},
		{"pdf", "application/pdf"},
		{"djvu", "image/vnd.djvu"},
		{"txt", "text/plain; charset=utf-8"},
		{".epub", "application/epub+zip"},
		{"EPUB", "application/epub+zip"},
		{"unknown", "application/octet-stream"},
		{"", "application/octet-stream"},
	}

	for _, tt := range tests {
		t.Run(tt.ext, func(t *testing.T) {
			assert.Equal(t, tt.expected, GetContentType(tt.ext))
		})
	}
}

func TestExtractFile_Success(t *testing.T) {
	archivePath := createTestZip(t, map[string]string{
		"12345.fb2": "book content here",
		"67890.epub": "epub content",
	})

	rc, size, err := ExtractFile(archivePath, "12345.fb2")
	require.NoError(t, err)
	defer func() { _ = rc.Close() }()

	assert.Equal(t, int64(17), size)

	data, err := io.ReadAll(rc)
	require.NoError(t, err)
	assert.Equal(t, "book content here", string(data))
}

func TestExtractFile_NotFound(t *testing.T) {
	archivePath := createTestZip(t, map[string]string{
		"12345.fb2": "content",
	})

	_, _, err := ExtractFile(archivePath, "nonexistent.fb2")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestExtractFile_BadArchive(t *testing.T) {
	dir := t.TempDir()
	badPath := filepath.Join(dir, "bad.zip")
	require.NoError(t, os.WriteFile(badPath, []byte("not a zip"), 0644))

	_, _, err := ExtractFile(badPath, "file.fb2")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "open archive")
}

func TestExtractFile_MissingArchive(t *testing.T) {
	_, _, err := ExtractFile("/nonexistent/archive.zip", "file.fb2")
	assert.Error(t, err)
}

func createTestZip(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.zip")

	f, err := os.Create(path)
	require.NoError(t, err)

	w := zip.NewWriter(f)
	for name, content := range files {
		fw, err := w.Create(name)
		require.NoError(t, err)
		_, err = fw.Write([]byte(content))
		require.NoError(t, err)
	}
	require.NoError(t, w.Close())
	require.NoError(t, f.Close())

	return path
}
