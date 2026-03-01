package glst

// GenreEntry represents a single genre parsed from a .glst file.
type GenreEntry struct {
	Position       string // "0.1.2" — path in the tree
	Code           string // "sf_action" — INPX code (may be empty for auto-generated)
	Name           string // "Боевая фантастика"
	Level          int    // 0, 1, 2, 3 — computed from position (number of dots - 1)
	ParentPosition string // "0.1" — computed from position (empty for root level)
}

// ParseResult contains the result of parsing a .glst file.
type ParseResult struct {
	Entries  []GenreEntry
	Warnings []string // Warnings about skipped/invalid lines
}
