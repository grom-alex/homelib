package models

type Genre struct {
	ID        int    `json:"id"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	ParentID  *int   `json:"parent_id,omitempty"`
	MetaGroup string `json:"meta_group,omitempty"`
	Position  string `json:"position"`
	SortOrder int    `json:"sort_order"`
	IsActive  bool   `json:"is_active"`
}

type GenreTreeItem struct {
	ID         int             `json:"id"`
	Code       string          `json:"code"`
	Name       string          `json:"name"`
	Position   string          `json:"position"`
	BooksCount int             `json:"books_count"`
	Children   []GenreTreeItem `json:"children,omitempty"`
}
