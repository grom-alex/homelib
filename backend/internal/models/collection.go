package models

import "time"

type Collection struct {
	ID             int        `json:"id"`
	Name           string     `json:"name"`
	Code           string     `json:"code"`
	CollectionType int        `json:"collection_type"`
	Description    string     `json:"description,omitempty"`
	SourceURL      string     `json:"source_url,omitempty"`
	Version        string     `json:"version,omitempty"`
	VersionDate    *time.Time `json:"version_date,omitempty"`
	BooksCount     int        `json:"books_count"`
	LastImportAt   *time.Time `json:"last_import_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}
