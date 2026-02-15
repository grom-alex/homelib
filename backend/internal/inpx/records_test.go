package inpx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseAuthors(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Author
	}{
		{
			"single author",
			"Булгаков,Михаил,Афанасьевич",
			[]Author{{LastName: "Булгаков", FirstName: "Михаил", MiddleName: "Афанасьевич"}},
		},
		{
			"multiple authors",
			"Бомонт,Френсис,:Флетчер,Джон,:",
			[]Author{
				{LastName: "Бомонт", FirstName: "Френсис"},
				{LastName: "Флетчер", FirstName: "Джон"},
			},
		},
		{
			"author without middle name",
			"Толстой,Лев,",
			[]Author{{LastName: "Толстой", FirstName: "Лев"}},
		},
		{
			"last name only",
			"Unknown",
			[]Author{{LastName: "Unknown"}},
		},
		{
			"empty string",
			"",
			nil,
		},
		{
			"colons only",
			":::",
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseAuthors(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseGenres(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"single genre", "sf_fantasy", []string{"sf_fantasy"}},
		{"multiple genres", "tragedy:drama:", []string{"tragedy", "drama"}},
		{"empty string", "", nil},
		{"only colons", "::", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseGenres(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseSeries(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		expectedName     string
		expectedType     string
		expectedNum      int
	}{
		{"publisher series with number", "Библиотека поэта[p]5", "Библиотека поэта", "p", 5},
		{"author series with number", "Гарри Поттер[a]3", "Гарри Поттер", "a", 3},
		{"publisher series no number", "Серия книг[p]", "Серия книг", "p", 0},
		{"plain series", "Обычная серия", "Обычная серия", "", 0},
		{"empty string", "", "", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, seriesType, num := ParseSeries(tt.input)
			assert.Equal(t, tt.expectedName, name)
			assert.Equal(t, tt.expectedType, seriesType)
			assert.Equal(t, tt.expectedNum, num)
		})
	}
}

func TestParseKeywords(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"single keyword", "fantasy", []string{"fantasy"}},
		{"multiple", "fantasy:magic:dragons:", []string{"fantasy", "magic", "dragons"}},
		{"empty", "", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseKeywords(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseRecord(t *testing.T) {
	// Standard field mapping with extended fields
	mapping := NewFieldMapping([]string{
		"AUTHOR", "GENRE", "TITLE", "SERIES", "SERNO",
		"FILE", "SIZE", "LIBID", "DEL", "EXT", "DATE",
		"INSNO", "FOLDER", "LANG", "LIBRATE", "KEYWORDS",
	})

	fields := []string{
		"Булгаков,Михаил,Афанасьевич",
		"dramaturgy",
		"Мастер и Маргарита",
		"",
		"0",
		"94240",
		"106260",
		"94240",
		"",
		"fb2",
		"2007-06-20",
		"27",
		"fb2-000024-030559.zip",
		"ru",
		"5",
		"dramaturgy",
	}

	rec := ParseRecord(fields, mapping, "default.zip")

	assert.Len(t, rec.Authors, 1)
	assert.Equal(t, "Булгаков", rec.Authors[0].LastName)
	assert.Equal(t, "Михаил", rec.Authors[0].FirstName)
	assert.Equal(t, "Афанасьевич", rec.Authors[0].MiddleName)
	assert.Equal(t, []string{"dramaturgy"}, rec.Genres)
	assert.Equal(t, "Мастер и Маргарита", rec.Title)
	assert.Equal(t, "", rec.Series)
	assert.Equal(t, 0, rec.SeriesNum)
	assert.Equal(t, "94240", rec.FileName)
	assert.Equal(t, int64(106260), rec.FileSize)
	assert.Equal(t, "94240", rec.LibID)
	assert.False(t, rec.IsDeleted)
	assert.Equal(t, "fb2", rec.Extension)
	assert.Equal(t, "2007-06-20", rec.Date)
	assert.Equal(t, "ru", rec.Language)
	assert.Equal(t, 5, rec.LibRate)
	assert.Equal(t, []string{"dramaturgy"}, rec.Keywords)
	assert.Equal(t, "fb2-000024-030559.zip", rec.ArchiveName)
}

func TestParseRecord_DefaultArchive(t *testing.T) {
	fields := []string{
		"Author,Name,",
		"genre",
		"Title",
		"",
		"",
		"12345",
		"100",
		"12345",
		"",
		"fb2",
		"2020-01-01",
	}

	rec := ParseRecord(fields, DefaultFieldMapping, "default-archive.zip")
	assert.Equal(t, "default-archive.zip", rec.ArchiveName)
}

func TestParseRecord_DeletedBook(t *testing.T) {
	fields := []string{
		"Author,Name,",
		"genre",
		"Deleted Book",
		"",
		"",
		"99999",
		"50",
		"99999",
		"1",
		"fb2",
		"2020-01-01",
	}

	rec := ParseRecord(fields, DefaultFieldMapping, "archive.zip")
	assert.True(t, rec.IsDeleted)
}

func TestParseRecord_WithSeries(t *testing.T) {
	fields := []string{
		"Роулинг,Джоан,",
		"sf_fantasy",
		"Гарри Поттер и Узник Азкабана",
		"Гарри Поттер[a]3",
		"",
		"12345",
		"500000",
		"12345",
		"",
		"fb2",
		"2020-01-01",
	}

	rec := ParseRecord(fields, DefaultFieldMapping, "archive.zip")
	assert.Equal(t, "Гарри Поттер", rec.Series)
	assert.Equal(t, "a", rec.SeriesType)
	assert.Equal(t, 3, rec.SeriesNum)
}

func TestParseRecord_SernoOverride(t *testing.T) {
	fields := []string{
		"Author,Name,",
		"genre",
		"Title",
		"Series[p]",
		"7",
		"12345",
		"100",
		"12345",
		"",
		"fb2",
		"2020-01-01",
	}

	rec := ParseRecord(fields, DefaultFieldMapping, "archive.zip")
	assert.Equal(t, "Series", rec.Series)
	assert.Equal(t, "p", rec.SeriesType)
	assert.Equal(t, 7, rec.SeriesNum) // SERNO overrides
}
