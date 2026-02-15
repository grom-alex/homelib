package models

import "time"

type Series struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type SeriesListItem struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	BooksCount int    `json:"books_count"`
}
