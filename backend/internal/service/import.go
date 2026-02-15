package service

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/grom-alex/homelib/backend/internal/config"
	"github.com/grom-alex/homelib/backend/internal/inpx"
	"github.com/grom-alex/homelib/backend/internal/models"
	"github.com/grom-alex/homelib/backend/internal/repository"
)

type ImportService struct {
	pool           *pgxpool.Pool
	cfg            config.ImportConfig
	libCfg         config.LibraryConfig
	bookRepo       *repository.BookRepo
	authorRepo     *repository.AuthorRepo
	genreRepo      *repository.GenreRepo
	seriesRepo     *repository.SeriesRepo
	collectionRepo *repository.CollectionRepo

	mu     sync.Mutex
	status models.ImportStatus
}

func NewImportService(
	pool *pgxpool.Pool,
	cfg config.ImportConfig,
	libCfg config.LibraryConfig,
	bookRepo *repository.BookRepo,
	authorRepo *repository.AuthorRepo,
	genreRepo *repository.GenreRepo,
	seriesRepo *repository.SeriesRepo,
	collectionRepo *repository.CollectionRepo,
) *ImportService {
	return &ImportService{
		pool:           pool,
		cfg:            cfg,
		libCfg:         libCfg,
		bookRepo:       bookRepo,
		authorRepo:     authorRepo,
		genreRepo:      genreRepo,
		seriesRepo:     seriesRepo,
		collectionRepo: collectionRepo,
		status:         models.ImportStatus{Status: "idle"},
	}
}

// StartImport begins an INPX import in the background.
// Returns an error if an import is already running.
func (s *ImportService) StartImport() error {
	s.mu.Lock()
	if s.status.Status == "running" {
		s.mu.Unlock()
		return fmt.Errorf("import is already running")
	}
	now := time.Now()
	s.status = models.ImportStatus{
		Status:    "running",
		StartedAt: &now,
	}
	s.mu.Unlock()

	go s.runImport()
	return nil
}

// GetStatus returns the current import status.
func (s *ImportService) GetStatus() models.ImportStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.status
}

// SetStatusForTest sets the import status (for testing only).
func (s *ImportService) SetStatusForTest(status models.ImportStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status = status
}

func (s *ImportService) runImport() {
	start := time.Now()
	stats, err := s.importINPX(context.Background())

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	if err != nil {
		errStr := err.Error()
		s.status = models.ImportStatus{
			Status:     "failed",
			StartedAt:  s.status.StartedAt,
			FinishedAt: &now,
			Stats:      stats,
			Error:      &errStr,
		}
		log.Printf("Import failed: %v", err)
		return
	}

	stats.DurationMs = time.Since(start).Milliseconds()
	s.status = models.ImportStatus{
		Status:     "completed",
		StartedAt:  s.status.StartedAt,
		FinishedAt: &now,
		Stats:      stats,
	}
	log.Printf("Import completed: %+v", stats)
}

func (s *ImportService) importINPX(ctx context.Context) (*models.ImportStats, error) {
	stats := &models.ImportStats{}

	f, err := os.Open(s.libCfg.INPXPath)
	if err != nil {
		return stats, fmt.Errorf("open INPX file: %w", err)
	}
	defer func() { _ = f.Close() }()

	fi, err := f.Stat()
	if err != nil {
		return stats, fmt.Errorf("stat INPX file: %w", err)
	}

	result, err := inpx.Parse(f, fi.Size())
	if err != nil {
		return stats, fmt.Errorf("parse INPX: %w", err)
	}

	// Update progress: total records known
	batchSize := s.cfg.BatchSize
	if batchSize <= 0 {
		batchSize = 3000
	}
	totalBatches := (len(result.Records) + batchSize - 1) / batchSize

	s.mu.Lock()
	s.status.TotalRecords = len(result.Records)
	s.status.TotalBatches = totalBatches
	s.mu.Unlock()

	log.Printf("INPX parsed: %d records, %d batches", len(result.Records), totalBatches)

	// Upsert collection
	coll := &models.Collection{
		Name:           result.Collection.Name,
		Code:           result.Collection.Code,
		CollectionType: result.Collection.Type,
		Description:    result.Collection.Description,
		SourceURL:      result.Collection.SourceURL,
		Version:        result.Version,
		BooksCount:     len(result.Records),
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return stats, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := s.collectionRepo.Upsert(ctx, tx, coll); err != nil {
		return stats, err
	}

	if err := tx.Commit(ctx); err != nil {
		return stats, fmt.Errorf("commit collection: %w", err)
	}

	// Caches for dedup within import
	authorCache := make(map[string]int64)
	genreCache := make(map[string]int)
	seriesCache := make(map[string]int64)

	batchNum := 0
	for i := 0; i < len(result.Records); i += batchSize {
		end := i + batchSize
		if end > len(result.Records) {
			end = len(result.Records)
		}
		batch := result.Records[i:end]

		batchStats, err := s.processBatch(ctx, batch, coll.ID, authorCache, genreCache, seriesCache)
		if err != nil {
			stats.Errors++
			log.Printf("Batch %d-%d error: %v", i, end, err)
		} else {
			stats.BooksAdded += batchStats.BooksAdded
			stats.BooksUpdated += batchStats.BooksUpdated
			stats.AuthorsAdded += batchStats.AuthorsAdded
			stats.GenresAdded += batchStats.GenresAdded
			stats.SeriesAdded += batchStats.SeriesAdded
		}

		batchNum++
		s.mu.Lock()
		s.status.ProcessedBatch = batchNum
		s.mu.Unlock()
	}

	// Count deleted
	for _, rec := range result.Records {
		if rec.IsDeleted {
			stats.BooksDeleted++
		}
	}

	return stats, nil
}

func (s *ImportService) processBatch(
	ctx context.Context,
	records []inpx.BookRecord,
	collectionID int,
	authorCache map[string]int64,
	genreCache map[string]int,
	seriesCache map[string]int64,
) (*models.ImportStats, error) {
	stats := &models.ImportStats{}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return stats, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Collect unique authors, genres, series for this batch
	authorSet := make(map[string]models.Author)
	genreSet := make(map[string]bool)
	seriesSet := make(map[string]bool)

	for _, rec := range records {
		for _, a := range rec.Authors {
			fullName := a.FullName()
			sortName := a.SortName()
			if _, ok := authorCache[sortName]; !ok {
				authorSet[sortName] = models.Author{Name: fullName, NameSort: sortName}
			}
		}
		for _, g := range rec.Genres {
			if _, ok := genreCache[g]; !ok {
				genreSet[g] = true
			}
		}
		if rec.Series != "" {
			if _, ok := seriesCache[rec.Series]; !ok {
				seriesSet[rec.Series] = true
			}
		}
	}

	// Upsert authors
	if len(authorSet) > 0 {
		authors := make([]models.Author, 0, len(authorSet))
		for _, a := range authorSet {
			authors = append(authors, a)
		}
		ids, err := s.authorRepo.UpsertAuthors(ctx, tx, authors)
		if err != nil {
			return stats, err
		}
		for k, v := range ids {
			if _, existed := authorCache[k]; !existed {
				stats.AuthorsAdded++
			}
			authorCache[k] = v
		}
	}

	// Upsert genres
	if len(genreSet) > 0 {
		codes := make([]string, 0, len(genreSet))
		for code := range genreSet {
			codes = append(codes, code)
		}
		ids, err := s.genreRepo.UpsertGenres(ctx, tx, codes)
		if err != nil {
			return stats, err
		}
		for k, v := range ids {
			if _, existed := genreCache[k]; !existed {
				stats.GenresAdded++
			}
			genreCache[k] = v
		}
	}

	// Upsert series
	if len(seriesSet) > 0 {
		names := make([]string, 0, len(seriesSet))
		for name := range seriesSet {
			names = append(names, name)
		}
		ids, err := s.seriesRepo.UpsertSeries(ctx, tx, names)
		if err != nil {
			return stats, err
		}
		for k, v := range ids {
			if _, existed := seriesCache[k]; !existed {
				stats.SeriesAdded++
			}
			seriesCache[k] = v
		}
	}

	// Build and upsert books
	books := make([]models.Book, 0, len(records))
	type bookMeta struct {
		authorIDs []int64
		genreIDs  []int
	}
	metas := make([]bookMeta, 0, len(records))

	for _, rec := range records {
		collIDPtr := &collectionID
		var seriesID *int64
		var seriesNum *int
		var seriesType *string

		if rec.Series != "" {
			if id, ok := seriesCache[rec.Series]; ok {
				seriesID = &id
			}
			if rec.SeriesNum > 0 {
				n := rec.SeriesNum
				seriesNum = &n
			}
			if rec.SeriesType != "" {
				seriesType = &rec.SeriesType
			}
		}

		var year *int
		// Year not directly in BookRecord — will be nil for now

		var fileSize *int64
		if rec.FileSize > 0 {
			fileSize = &rec.FileSize
		}

		var libRate *int16
		if rec.LibRate > 0 {
			r := int16(rec.LibRate)
			libRate = &r
		}

		var description *string
		// Description not in INPX — will be nil

		var keywords []string
		if len(rec.Keywords) > 0 {
			keywords = rec.Keywords
		}

		var dateAdded *time.Time
		if rec.Date != "" {
			if t, err := time.Parse("2006-01-02", rec.Date); err == nil {
				dateAdded = &t
			}
		}

		fileInArchive := rec.FileName
		if rec.Extension != "" {
			fileInArchive = rec.FileName + "." + rec.Extension
		}

		book := models.Book{
			CollectionID:  collIDPtr,
			Title:         rec.Title,
			Lang:          rec.Language,
			Year:          year,
			Format:        rec.Extension,
			FileSize:      fileSize,
			ArchiveName:   rec.ArchiveName,
			FileInArchive: fileInArchive,
			SeriesID:      seriesID,
			SeriesNum:     seriesNum,
			SeriesType:    seriesType,
			LibID:         rec.LibID,
			LibRate:       libRate,
			IsDeleted:     rec.IsDeleted,
			Description:   description,
			Keywords:      keywords,
			DateAdded:     dateAdded,
		}
		books = append(books, book)

		// Collect author/genre IDs
		var aIDs []int64
		for _, a := range rec.Authors {
			if id, ok := authorCache[a.SortName()]; ok {
				aIDs = append(aIDs, id)
			}
		}
		var gIDs []int
		for _, g := range rec.Genres {
			if id, ok := genreCache[g]; ok {
				gIDs = append(gIDs, id)
			}
		}
		metas = append(metas, bookMeta{authorIDs: aIDs, genreIDs: gIDs})
	}

	inserted, updated, err := s.bookRepo.BatchUpsert(ctx, tx, books)
	if err != nil {
		return stats, err
	}
	stats.BooksAdded = inserted
	stats.BooksUpdated = updated

	// Set M:N relationships
	for i, book := range books {
		if book.ID == 0 {
			continue
		}
		if len(metas[i].authorIDs) > 0 {
			if err := s.bookRepo.SetBookAuthors(ctx, tx, book.ID, metas[i].authorIDs); err != nil {
				return stats, err
			}
		}
		if len(metas[i].genreIDs) > 0 {
			if err := s.bookRepo.SetBookGenres(ctx, tx, book.ID, metas[i].genreIDs); err != nil {
				return stats, err
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return stats, fmt.Errorf("commit batch: %w", err)
	}

	return stats, nil
}
