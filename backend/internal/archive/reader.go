package archive

import (
	"archive/zip"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

// ContentTypes maps file extensions to MIME types.
var ContentTypes = map[string]string{
	"fb2":  "application/x-fictionbook+xml",
	"epub": "application/epub+zip",
	"pdf":  "application/pdf",
	"djvu": "image/vnd.djvu",
	"doc":  "application/msword",
	"txt":  "text/plain; charset=utf-8",
	"rtf":  "application/rtf",
	"htm":  "text/html; charset=utf-8",
	"html": "text/html; charset=utf-8",
	"mobi": "application/x-mobipocket-ebook",
	"azw3": "application/vnd.amazon.ebook",
}

// GetContentType returns the MIME type for a file extension.
func GetContentType(ext string) string {
	ext = strings.ToLower(strings.TrimPrefix(ext, "."))
	if ct, ok := ContentTypes[ext]; ok {
		return ct
	}
	return "application/octet-stream"
}

// ExtractFile opens a ZIP archive and returns a reader for the specified file.
// The caller must close the returned ReadCloser.
func ExtractFile(archivePath, fileInArchive string) (io.ReadCloser, int64, error) {
	zr, err := zip.OpenReader(archivePath)
	if err != nil {
		return nil, 0, fmt.Errorf("open archive %s: %w", archivePath, err)
	}

	for _, f := range zr.File {
		if f.Name == fileInArchive {
			rc, err := f.Open()
			if err != nil {
				_ = zr.Close()
				return nil, 0, fmt.Errorf("open file in archive: %w", err)
			}
			// Wrap to close both the file and the archive
			return &archiveFileReader{rc: rc, zr: zr}, int64(f.UncompressedSize64), nil
		}
	}

	_ = zr.Close()
	return nil, 0, fmt.Errorf("file %s not found in archive %s", fileInArchive, filepath.Base(archivePath))
}

type archiveFileReader struct {
	rc io.ReadCloser
	zr *zip.ReadCloser
}

func (r *archiveFileReader) Read(p []byte) (int, error) {
	return r.rc.Read(p)
}

func (r *archiveFileReader) Close() error {
	err1 := r.rc.Close()
	err2 := r.zr.Close()
	if err1 != nil {
		return err1
	}
	return err2
}
