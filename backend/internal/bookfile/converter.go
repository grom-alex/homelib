package bookfile

import "fmt"

// BookMetadata contains metadata extracted from a book file.
type BookMetadata struct {
	Title    string `json:"title"`
	Author   string `json:"author"`
	Cover    string `json:"cover,omitempty"`
	Language string `json:"language"`
	Format   string `json:"format"`
}

// TOCEntry represents a single entry in the table of contents.
type TOCEntry struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Level int    `json:"level"`
}

// BookContent is the result of converting a book (metadata + structure, no chapter text).
type BookContent struct {
	Metadata      BookMetadata   `json:"metadata"`
	TOC           []TOCEntry     `json:"toc"`
	ChapterIDs    []string       `json:"chapters"`
	TotalChapters int            `json:"totalChapters"`
	ChapterSizes  map[string]int `json:"chapterSizes,omitempty"`
}

// ChapterContent holds the HTML content of a single chapter.
type ChapterContent struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	HTML  string `json:"html"`
}

// ImageData holds binary image data extracted from a book.
type ImageData struct {
	ID          string
	ContentType string
	Data        []byte
}

// BookConverter is the interface for converting book files to HTML.
// Each format (FB2, EPUB, etc.) implements this interface.
type BookConverter interface {
	// Parse reads raw book data and prepares internal state.
	Parse(data []byte, bookID int64) error

	// Content returns the book metadata and structure.
	Content() *BookContent

	// Chapter returns the HTML content of a specific chapter.
	Chapter(chapterID string) (*ChapterContent, error)

	// Image returns binary data for an embedded image.
	Image(imageID string) (*ImageData, error)
}

// GetConverter returns the appropriate converter for the given book format.
func GetConverter(format string) (BookConverter, error) {
	switch format {
	case "fb2":
		return &FB2Converter{}, nil
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}
