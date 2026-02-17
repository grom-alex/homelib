package inpx

import (
	"regexp"
	"strconv"
	"strings"
)

var seriesTypeRe = regexp.MustCompile(`^(.+?)\[([ap])\](\d*)$`)

// ParseAuthors parses an INPX AUTHOR field.
// Format: "LastName,FirstName,MiddleName:LastName2,FirstName2,MiddleName2:"
func ParseAuthors(s string) []Author {
	var authors []Author
	for _, part := range strings.Split(s, ":") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		names := strings.Split(part, ",")
		a := Author{}
		if len(names) > 0 {
			a.LastName = strings.TrimSpace(names[0])
		}
		if len(names) > 1 {
			a.FirstName = strings.TrimSpace(names[1])
		}
		if len(names) > 2 {
			a.MiddleName = strings.TrimSpace(names[2])
		}
		if a.LastName != "" || a.FirstName != "" {
			authors = append(authors, a)
		}
	}
	return authors
}

// ParseGenres parses an INPX GENRE field.
// Format: "genre1:genre2:genre3:"
func ParseGenres(s string) []string {
	var genres []string
	for _, g := range strings.Split(s, ":") {
		g = strings.TrimSpace(g)
		if g != "" {
			genres = append(genres, g)
		}
	}
	return genres
}

// ParseSeries parses an INPX SERIES field with optional type marker.
// Format: "Series Name[p]5" where [a]=author's, [p]=publisher's.
// Returns series name, type ("a"/"p"/""), and number.
func ParseSeries(s string) (name, seriesType string, num int) {
	s = strings.TrimSpace(s)
	if s == "" {
		return
	}
	if m := seriesTypeRe.FindStringSubmatch(s); m != nil {
		name = m[1]
		seriesType = m[2]
		if m[3] != "" {
			num, _ = strconv.Atoi(m[3])
		}
		return
	}
	name = s
	return
}

// ParseKeywords parses an INPX KEYWORDS field.
// Format: "kw1:kw2:kw3:"
func ParseKeywords(s string) []string {
	var keywords []string
	for _, kw := range strings.Split(s, ":") {
		kw = strings.TrimSpace(kw)
		if kw != "" {
			keywords = append(keywords, kw)
		}
	}
	return keywords
}

// ParseRecord parses a single .inp line (already split by \x04) into a BookRecord.
func ParseRecord(fields []string, mapping FieldMapping, defaultArchive string) BookRecord {
	get := func(name string) string {
		idx := mapping.Get(name)
		if idx < 0 || idx >= len(fields) {
			return ""
		}
		return strings.TrimSpace(fields[idx])
	}

	rec := BookRecord{}

	rec.Authors = ParseAuthors(get("AUTHOR"))
	rec.Genres = ParseGenres(get("GENRE"))
	rec.Title = get("TITLE")

	seriesRaw := get("SERIES")
	if seriesRaw != "" {
		rec.Series, rec.SeriesType, rec.SeriesNum = ParseSeries(seriesRaw)
	}

	// SERNO overrides series number from ParseSeries if present
	if serno := get("SERNO"); serno != "" {
		if n, err := strconv.Atoi(serno); err == nil && n > 0 {
			rec.SeriesNum = n
		}
	}

	rec.FileName = get("FILE")
	if sizeStr := get("SIZE"); sizeStr != "" {
		rec.FileSize, _ = strconv.ParseInt(sizeStr, 10, 64)
	}
	rec.LibID = get("LIBID")
	rec.IsDeleted = get("DEL") == "1"
	rec.Extension = strings.ToLower(get("EXT"))
	rec.Date = get("DATE")
	rec.Language = strings.ToLower(get("LANG"))

	if rateStr := get("LIBRATE"); rateStr != "" {
		if n, err := strconv.Atoi(rateStr); err == nil {
			rec.LibRate = n
		}
	}

	rec.Keywords = ParseKeywords(get("KEYWORDS"))

	if folder := get("FOLDER"); folder != "" {
		rec.ArchiveName = folder
	} else {
		rec.ArchiveName = defaultArchive
	}

	if insnoStr := get("INSNO"); insnoStr != "" {
		rec.InsNo, _ = strconv.Atoi(insnoStr)
	}

	return rec
}
