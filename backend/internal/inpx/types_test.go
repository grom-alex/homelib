package inpx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFieldMapping(t *testing.T) {
	fm := NewFieldMapping([]string{"AUTHOR", "GENRE", "TITLE"})

	assert.Equal(t, 0, fm.Get("AUTHOR"))
	assert.Equal(t, 1, fm.Get("GENRE"))
	assert.Equal(t, 2, fm.Get("TITLE"))
	assert.Equal(t, -1, fm.Get("MISSING"))
}

func TestDefaultFieldMapping(t *testing.T) {
	assert.Equal(t, 0, DefaultFieldMapping.Get("AUTHOR"))
	assert.Equal(t, 1, DefaultFieldMapping.Get("GENRE"))
	assert.Equal(t, 2, DefaultFieldMapping.Get("TITLE"))
	assert.Equal(t, 3, DefaultFieldMapping.Get("SERIES"))
	assert.Equal(t, 4, DefaultFieldMapping.Get("SERNO"))
	assert.Equal(t, 5, DefaultFieldMapping.Get("FILE"))
	assert.Equal(t, 6, DefaultFieldMapping.Get("SIZE"))
	assert.Equal(t, 7, DefaultFieldMapping.Get("LIBID"))
	assert.Equal(t, 8, DefaultFieldMapping.Get("DEL"))
	assert.Equal(t, 9, DefaultFieldMapping.Get("EXT"))
	assert.Equal(t, 10, DefaultFieldMapping.Get("DATE"))
	assert.Equal(t, -1, DefaultFieldMapping.Get("LANG"))
}

func TestAuthor_FullName(t *testing.T) {
	tests := []struct {
		author   Author
		expected string
	}{
		{Author{"Булгаков", "Михаил", "Афанасьевич"}, "Булгаков Михаил Афанасьевич"},
		{Author{"Толстой", "Лев", ""}, "Толстой Лев"},
		{Author{"Достоевский", "", ""}, "Достоевский"},
		{Author{"", "", ""}, ""},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, tt.author.FullName())
	}
}

func TestAuthor_SortName(t *testing.T) {
	tests := []struct {
		author   Author
		expected string
	}{
		{Author{"Булгаков", "Михаил", "Афанасьевич"}, "Булгаков, Михаил Афанасьевич"},
		{Author{"Толстой", "Лев", ""}, "Толстой, Лев"},
		{Author{"Достоевский", "", ""}, "Достоевский"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, tt.author.SortName())
	}
}
