package service

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/grom-alex/homelib/backend/internal/archive"
	"github.com/grom-alex/homelib/backend/internal/config"
	"github.com/grom-alex/homelib/backend/internal/repository"
)

type DownloadService struct {
	bookRepo *repository.BookRepo
	libCfg   config.LibraryConfig
}

func NewDownloadService(bookRepo *repository.BookRepo, libCfg config.LibraryConfig) *DownloadService {
	return &DownloadService{bookRepo: bookRepo, libCfg: libCfg}
}

type DownloadResult struct {
	Reader      io.ReadCloser
	Filename    string
	ContentType string
	Size        int64
}

// DownloadBook returns a stream for the book file extracted from a ZIP archive.
func (s *DownloadService) DownloadBook(ctx context.Context, bookID int64) (*DownloadResult, error) {
	archiveName, fileInArchive, format, err := s.bookRepo.GetBookForDownload(ctx, bookID)
	if err != nil {
		return nil, fmt.Errorf("book not found: %w", err)
	}

	if archiveName == "" || fileInArchive == "" {
		return nil, fmt.Errorf("book not found: missing archive info")
	}

	// Prevent path traversal: ensure resolved path stays within ArchivesPath
	archivePath := filepath.Join(s.libCfg.ArchivesPath, archiveName)
	absArchivePath, err := filepath.Abs(archivePath)
	if err != nil {
		return nil, fmt.Errorf("invalid archive path: %w", err)
	}
	absBasePath, err := filepath.Abs(s.libCfg.ArchivesPath)
	if err != nil {
		return nil, fmt.Errorf("invalid base path: %w", err)
	}
	if !strings.HasPrefix(absArchivePath, absBasePath+string(filepath.Separator)) && absArchivePath != absBasePath {
		return nil, fmt.Errorf("book not found: invalid archive path")
	}

	reader, size, err := archive.ExtractFile(archivePath, fileInArchive)
	if err != nil {
		return nil, fmt.Errorf("extract file: %w", err)
	}

	return &DownloadResult{
		Reader:      reader,
		Filename:    fileInArchive,
		ContentType: archive.GetContentType(format),
		Size:        size,
	}, nil
}
