package models

import "time"

type Book struct {
	ID             int64      `json:"id"`
	CollectionID   *int       `json:"collection_id,omitempty"`
	Title          string     `json:"title"`
	Lang           string     `json:"lang"`
	Year           *int       `json:"year,omitempty"`
	Format         string     `json:"format"`
	FileSize       *int64     `json:"file_size,omitempty"`
	ArchiveName    string     `json:"archive_name"`
	FileInArchive  string     `json:"file_in_archive"`
	SeriesID       *int64     `json:"series_id,omitempty"`
	SeriesNum      *int       `json:"series_num,omitempty"`
	SeriesType     *string    `json:"series_type,omitempty"`
	LibID          string     `json:"lib_id,omitempty"`
	LibRate        *int16     `json:"lib_rate,omitempty"`
	IsDeleted      bool       `json:"is_deleted"`
	HasCover       bool       `json:"has_cover"`
	Description    *string    `json:"description,omitempty"`
	Keywords       []string   `json:"keywords,omitempty"`
	DateAdded      *time.Time `json:"date_added,omitempty"`
	AddedAt        time.Time  `json:"added_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type BookListItem struct {
	ID        int64             `json:"id"`
	Title     string            `json:"title"`
	Lang      string            `json:"lang"`
	Year      *int              `json:"year,omitempty"`
	Format    string            `json:"format"`
	FileSize  *int64            `json:"file_size,omitempty"`
	LibRate   *int16            `json:"lib_rate,omitempty"`
	IsDeleted bool              `json:"is_deleted"`
	Authors   []BookAuthorRef   `json:"authors"`
	Genres    []BookGenreRef    `json:"genres"`
	Series    *BookSeriesRef    `json:"series,omitempty"`
}

type BookAuthorRef struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type BookGenreRef struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type BookSeriesRef struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Num  *int   `json:"num,omitempty"`
}

type BookDetail struct {
	ID          int64             `json:"id"`
	Title       string            `json:"title"`
	Lang        string            `json:"lang"`
	Year        *int              `json:"year,omitempty"`
	Format      string            `json:"format"`
	FileSize    *int64            `json:"file_size,omitempty"`
	LibRate     *int16            `json:"lib_rate,omitempty"`
	IsDeleted   bool              `json:"is_deleted"`
	Description *string           `json:"description,omitempty"`
	Keywords    []string          `json:"keywords,omitempty"`
	DateAdded   *time.Time        `json:"date_added,omitempty"`
	Authors     []BookAuthorRef   `json:"authors"`
	Genres      []BookGenreDetailRef `json:"genres"`
	Series      *BookSeriesDetailRef `json:"series,omitempty"`
	Collection  *BookCollectionRef   `json:"collection,omitempty"`
}

type BookGenreDetailRef struct {
	ID   int    `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type BookSeriesDetailRef struct {
	ID   int64   `json:"id"`
	Name string  `json:"name"`
	Num  *int    `json:"num,omitempty"`
	Type *string `json:"type,omitempty"`
}

type BookCollectionRef struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type BookFilter struct {
	Query    string `form:"q"`
	AuthorID *int64 `form:"author_id"`
	GenreID  *int   `form:"genre_id"`
	SeriesID *int64 `form:"series_id"`
	Lang     string `form:"lang"`
	Format   string `form:"format"`
	Page     int    `form:"page"`
	Limit    int    `form:"limit"`
	Sort     string `form:"sort"`
	Order    string `form:"order"`
}

func (f *BookFilter) SetDefaults() {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 || f.Limit > 100 {
		f.Limit = 20
	}
	if f.Sort == "" {
		f.Sort = "title"
	}
	if f.Order == "" {
		f.Order = "asc"
	}
}

func (f *BookFilter) Offset() int {
	return (f.Page - 1) * f.Limit
}
