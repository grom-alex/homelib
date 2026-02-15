package models

type Genre struct {
	ID        int    `json:"id"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	ParentID  *int   `json:"parent_id,omitempty"`
	MetaGroup string `json:"meta_group,omitempty"`
}

type GenreTreeItem struct {
	ID         int             `json:"id"`
	Code       string          `json:"code"`
	Name       string          `json:"name"`
	MetaGroup  string          `json:"meta_group,omitempty"`
	BooksCount int             `json:"books_count"`
	Children   []GenreTreeItem `json:"children,omitempty"`
}
