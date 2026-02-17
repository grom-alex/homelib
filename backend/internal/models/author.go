package models

import "time"

type Author struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	NameSort  string    `json:"name_sort"`
	CreatedAt time.Time `json:"created_at"`
}

type AuthorListItem struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	BooksCount int    `json:"books_count"`
}

type AuthorDetail struct {
	ID         int64          `json:"id"`
	Name       string         `json:"name"`
	Books      []BookListItem `json:"books"`
	BooksCount int            `json:"books_count"`
}
