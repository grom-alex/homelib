package inpx

import "strings"

// CollectionInfo holds metadata parsed from collection.info inside INPX.
type CollectionInfo struct {
	Name        string
	Code        string
	Type        int
	Description string
	SourceURL   string
}

// FieldMapping maps field names to their positions in .inp records.
type FieldMapping struct {
	Fields []string
	Index  map[string]int
}

// Get returns the index for a field name, or -1 if not found.
func (fm FieldMapping) Get(name string) int {
	if idx, ok := fm.Index[name]; ok {
		return idx
	}
	return -1
}

// NewFieldMapping creates a FieldMapping from a list of field names.
func NewFieldMapping(fields []string) FieldMapping {
	idx := make(map[string]int, len(fields))
	for i, f := range fields {
		idx[strings.TrimSpace(f)] = i
	}
	return FieldMapping{Fields: fields, Index: idx}
}

// DefaultFieldMapping is used when structure.info is absent from the INPX.
var DefaultFieldMapping = NewFieldMapping([]string{
	"AUTHOR", "GENRE", "TITLE", "SERIES", "SERNO",
	"FILE", "SIZE", "LIBID", "DEL", "EXT", "DATE",
})

// Author represents a parsed book author with name parts.
type Author struct {
	LastName   string
	FirstName  string
	MiddleName string
}

// FullName returns the displayable author name.
func (a Author) FullName() string {
	parts := []string{}
	if a.LastName != "" {
		parts = append(parts, a.LastName)
	}
	if a.FirstName != "" {
		parts = append(parts, a.FirstName)
	}
	if a.MiddleName != "" {
		parts = append(parts, a.MiddleName)
	}
	return strings.Join(parts, " ")
}

// SortName returns the name in "LastName, FirstName MiddleName" form for sorting and dedup.
func (a Author) SortName() string {
	if a.FirstName == "" {
		return a.LastName
	}
	name := a.LastName + ", " + a.FirstName
	if a.MiddleName != "" {
		name += " " + a.MiddleName
	}
	return name
}

// BookRecord represents a single book entry parsed from an .inp file.
type BookRecord struct {
	Authors     []Author
	Genres      []string
	Title       string
	Series      string
	SeriesType  string // "a" = author's, "p" = publisher's, "" = unknown
	SeriesNum   int
	FileName    string
	FileSize    int64
	LibID       string
	IsDeleted   bool
	Extension   string
	Date        string
	Language    string
	LibRate     int
	Keywords    []string
	ArchiveName string
	InsNo       int
}
