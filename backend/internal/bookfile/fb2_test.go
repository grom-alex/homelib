package bookfile

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testdataPath(name string) string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "fb2_testdata", name)
}

func loadTestFB2(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(testdataPath(name))
	require.NoError(t, err)
	return data
}

func parseTestFB2(t *testing.T, name string, bookID int64) *FB2Converter {
	t.Helper()
	conv := &FB2Converter{}
	data := loadTestFB2(t, name)
	require.NoError(t, conv.Parse(data, bookID))
	return conv
}

// --- Parse & Metadata ---

func TestFB2Converter_Parse_SimpleMetadata(t *testing.T) {
	conv := parseTestFB2(t, "simple.fb2", 1)
	content := conv.Content()

	assert.Equal(t, "Простая книга", content.Metadata.Title)
	assert.Equal(t, "Иван Тестов", content.Metadata.Author)
	assert.Equal(t, "ru", content.Metadata.Language)
	assert.Equal(t, "fb2", content.Metadata.Format)
	assert.Empty(t, content.Metadata.Cover)
}

func TestFB2Converter_Parse_ComplexMetadata(t *testing.T) {
	conv := parseTestFB2(t, "complex.fb2", 42)
	content := conv.Content()

	assert.Equal(t, "Сборник с элементами", content.Metadata.Title)
	assert.Equal(t, "Анна Поэтова", content.Metadata.Author)
	assert.Equal(t, "ru", content.Metadata.Language)
	assert.Equal(t, "/api/books/42/image/cover.jpg", content.Metadata.Cover)
}

func TestFB2Converter_Parse_MalformedXML(t *testing.T) {
	conv := &FB2Converter{}
	data := loadTestFB2(t, "malformed.fb2")
	err := conv.Parse(data, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parse FB2 XML")
}

func TestFB2Converter_Parse_InvalidData(t *testing.T) {
	conv := &FB2Converter{}
	err := conv.Parse([]byte("not xml at all"), 1)
	assert.Error(t, err)
}

func TestFB2Converter_Parse_EmptyData(t *testing.T) {
	conv := &FB2Converter{}
	err := conv.Parse([]byte{}, 1)
	assert.Error(t, err)
}

// --- TOC & Chapters ---

func TestFB2Converter_TOC_SimpleBook(t *testing.T) {
	conv := parseTestFB2(t, "simple.fb2", 1)
	content := conv.Content()

	assert.Equal(t, 3, content.TotalChapters)
	require.Len(t, content.TOC, 3)
	require.Len(t, content.ChapterIDs, 3)

	assert.Equal(t, "Глава первая", content.TOC[0].Title)
	assert.Equal(t, 0, content.TOC[0].Level)

	assert.Equal(t, "Глава вторая", content.TOC[1].Title)
	assert.Equal(t, 0, content.TOC[1].Level)

	assert.Equal(t, "Глава третья", content.TOC[2].Title)
	assert.Equal(t, 0, content.TOC[2].Level)
}

func TestFB2Converter_TOC_NestedSections(t *testing.T) {
	conv := parseTestFB2(t, "complex.fb2", 1)
	content := conv.Content()

	// complex.fb2 has: Часть первая > (Глава 1, Глава 2), Часть вторая
	require.True(t, content.TotalChapters >= 4, "expected at least 4 TOC entries")

	// First level entries
	assert.Equal(t, "Часть первая", content.TOC[0].Title)
	assert.Equal(t, 0, content.TOC[0].Level)

	// Nested chapters
	assert.Equal(t, "Глава 1. Стихи", content.TOC[1].Title)
	assert.Equal(t, 1, content.TOC[1].Level)

	assert.Equal(t, "Глава 2. Цитаты и изображения", content.TOC[2].Title)
	assert.Equal(t, 1, content.TOC[2].Level)

	// Second part
	assert.Equal(t, "Часть вторая", content.TOC[3].Title)
	assert.Equal(t, 0, content.TOC[3].Level)
}

func TestFB2Converter_TOC_SingleSection_AutoTitle(t *testing.T) {
	conv := parseTestFB2(t, "single_section.fb2", 1)
	content := conv.Content()

	assert.Equal(t, 1, content.TotalChapters)
	require.Len(t, content.TOC, 1)
	// Section without title gets auto-generated "Глава N"
	assert.Equal(t, "Глава 1", content.TOC[0].Title)
}

func TestFB2Converter_ChapterIDs_Sequential(t *testing.T) {
	conv := parseTestFB2(t, "simple.fb2", 1)
	content := conv.Content()

	for _, id := range content.ChapterIDs {
		assert.True(t, strings.HasPrefix(id, "ch"), "chapter ID %q should start with 'ch'", id)
	}
	// IDs must be unique
	seen := make(map[string]bool)
	for _, id := range content.ChapterIDs {
		assert.False(t, seen[id], "duplicate chapter ID: %s", id)
		seen[id] = true
	}
}

// --- Chapter Content ---

func TestFB2Converter_Chapter_NotFound(t *testing.T) {
	conv := parseTestFB2(t, "simple.fb2", 1)
	_, err := conv.Chapter("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestFB2Converter_Chapter_Title(t *testing.T) {
	conv := parseTestFB2(t, "simple.fb2", 1)
	content := conv.Content()

	ch, err := conv.Chapter(content.ChapterIDs[0])
	require.NoError(t, err)

	assert.Equal(t, content.ChapterIDs[0], ch.ID)
	assert.Equal(t, "Глава первая", ch.Title)
}

func TestFB2Converter_Chapter_Paragraphs(t *testing.T) {
	conv := parseTestFB2(t, "simple.fb2", 1)
	content := conv.Content()

	ch, err := conv.Chapter(content.ChapterIDs[0])
	require.NoError(t, err)

	assert.Contains(t, ch.HTML, "<p>")
	assert.Contains(t, ch.HTML, "Первый параграф первой главы")
}

func TestFB2Converter_Chapter_HasTitleTag(t *testing.T) {
	conv := parseTestFB2(t, "simple.fb2", 1)
	content := conv.Content()

	ch, err := conv.Chapter(content.ChapterIDs[0])
	require.NoError(t, err)

	assert.Contains(t, ch.HTML, `<h2 class="chapter-title">`)
	assert.Contains(t, ch.HTML, "Глава первая")
}

// --- Inline Tag Mapping ---

func TestFB2Converter_InlineTags_Emphasis(t *testing.T) {
	conv := parseTestFB2(t, "simple.fb2", 1)
	content := conv.Content()

	ch, err := conv.Chapter(content.ChapterIDs[0])
	require.NoError(t, err)

	assert.Contains(t, ch.HTML, "<em>курсивом</em>")
}

func TestFB2Converter_InlineTags_Strong(t *testing.T) {
	conv := parseTestFB2(t, "simple.fb2", 1)
	content := conv.Content()

	ch, err := conv.Chapter(content.ChapterIDs[0])
	require.NoError(t, err)

	assert.Contains(t, ch.HTML, "<strong>жирным</strong>")
}

func TestFB2Converter_InlineTags_Strikethrough(t *testing.T) {
	conv := parseTestFB2(t, "simple.fb2", 1)
	content := conv.Content()

	ch, err := conv.Chapter(content.ChapterIDs[0])
	require.NoError(t, err)

	assert.Contains(t, ch.HTML, "<del>зачёркнутым</del>")
}

func TestFB2Converter_InlineTags_Code(t *testing.T) {
	conv := parseTestFB2(t, "simple.fb2", 1)
	content := conv.Content()

	ch, err := conv.Chapter(content.ChapterIDs[0])
	require.NoError(t, err)

	assert.Contains(t, ch.HTML, "<code>моноширинным</code>")
}

func TestFB2Converter_InlineTags_SubSup(t *testing.T) {
	conv := parseTestFB2(t, "simple.fb2", 1)
	content := conv.Content()

	ch, err := conv.Chapter(content.ChapterIDs[1])
	require.NoError(t, err)

	assert.Contains(t, ch.HTML, "H<sub>2</sub>O")
	assert.Contains(t, ch.HTML, "<sup>1</sup>")
}

// --- Subtitle ---

func TestFB2Converter_Subtitle(t *testing.T) {
	conv := parseTestFB2(t, "simple.fb2", 1)
	content := conv.Content()

	ch, err := conv.Chapter(content.ChapterIDs[2])
	require.NoError(t, err)

	assert.Contains(t, ch.HTML, `<p class="subtitle">`)
	assert.Contains(t, ch.HTML, "Подзаголовок третьей главы")
}

// --- Empty Line ---

func TestFB2Converter_EmptyLine(t *testing.T) {
	conv := parseTestFB2(t, "simple.fb2", 1)
	content := conv.Content()

	ch, err := conv.Chapter(content.ChapterIDs[1])
	require.NoError(t, err)

	assert.Contains(t, ch.HTML, "<br/>")
}

// --- Epigraph ---

func TestFB2Converter_Epigraph(t *testing.T) {
	conv := parseTestFB2(t, "complex.fb2", 1)
	content := conv.Content()

	// First chapter (Часть первая) has epigraph
	ch, err := conv.Chapter(content.ChapterIDs[0])
	require.NoError(t, err)

	assert.Contains(t, ch.HTML, `<blockquote class="epigraph">`)
	assert.Contains(t, ch.HTML, "Быть или не быть")
	assert.Contains(t, ch.HTML, `<p class="epigraph-author">`)
	assert.Contains(t, ch.HTML, "У. Шекспир")
}

// --- Poem ---

func TestFB2Converter_Poem(t *testing.T) {
	conv := parseTestFB2(t, "complex.fb2", 1)
	content := conv.Content()

	// "Глава 1. Стихи" contains the poem
	ch, err := conv.Chapter(content.ChapterIDs[1])
	require.NoError(t, err)

	assert.Contains(t, ch.HTML, `<div class="poem">`)
	assert.Contains(t, ch.HTML, `<div class="stanza">`)
	assert.Contains(t, ch.HTML, `<p class="verse">`)
	assert.Contains(t, ch.HTML, "Мороз и солнце; день чудесный!")
	assert.Contains(t, ch.HTML, `<p class="poem-author">`)
	assert.Contains(t, ch.HTML, "А.С. Пушкин")
}

func TestFB2Converter_Poem_HasTwoStanzas(t *testing.T) {
	conv := parseTestFB2(t, "complex.fb2", 1)
	content := conv.Content()

	ch, err := conv.Chapter(content.ChapterIDs[1])
	require.NoError(t, err)

	stanzaCount := strings.Count(ch.HTML, `<div class="stanza">`)
	assert.Equal(t, 2, stanzaCount)
}

func TestFB2Converter_Poem_Title(t *testing.T) {
	conv := parseTestFB2(t, "complex.fb2", 1)
	content := conv.Content()

	ch, err := conv.Chapter(content.ChapterIDs[1])
	require.NoError(t, err)

	// Poem title rendered as subtitle
	assert.Contains(t, ch.HTML, "Зимнее утро")
}

// --- Cite ---

func TestFB2Converter_Cite(t *testing.T) {
	conv := parseTestFB2(t, "complex.fb2", 1)
	content := conv.Content()

	// "Глава 2. Цитаты и изображения" contains cite
	ch, err := conv.Chapter(content.ChapterIDs[2])
	require.NoError(t, err)

	assert.Contains(t, ch.HTML, `<blockquote class="cite">`)
	assert.Contains(t, ch.HTML, "Все счастливые семьи похожи")
	assert.Contains(t, ch.HTML, "Л.Н. Толстой")
}

// --- Image References ---

func TestFB2Converter_InlineImage(t *testing.T) {
	conv := parseTestFB2(t, "complex.fb2", 42)
	content := conv.Content()

	// "Глава 2" has inline image
	ch, err := conv.Chapter(content.ChapterIDs[2])
	require.NoError(t, err)

	assert.Contains(t, ch.HTML, `<img src="/api/books/42/image/img1.png"`)
	assert.Contains(t, ch.HTML, `loading="lazy"`)
}

func TestFB2Converter_Image_ExtractBinary(t *testing.T) {
	conv := parseTestFB2(t, "complex.fb2", 1)

	img, err := conv.Image("cover.jpg")
	require.NoError(t, err)

	assert.Equal(t, "cover.jpg", img.ID)
	assert.Equal(t, "image/jpeg", img.ContentType)
	assert.NotEmpty(t, img.Data)
}

func TestFB2Converter_Image_PNG(t *testing.T) {
	conv := parseTestFB2(t, "complex.fb2", 1)

	img, err := conv.Image("img1.png")
	require.NoError(t, err)

	assert.Equal(t, "img1.png", img.ID)
	assert.Equal(t, "image/png", img.ContentType)
	assert.NotEmpty(t, img.Data)
	// PNG magic bytes
	assert.Equal(t, byte(0x89), img.Data[0])
	assert.Equal(t, byte('P'), img.Data[1])
}

func TestFB2Converter_Image_NotFound(t *testing.T) {
	conv := parseTestFB2(t, "complex.fb2", 1)

	_, err := conv.Image("nonexistent.png")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// --- Footnotes ---

func TestFB2Converter_FootnoteRef(t *testing.T) {
	conv := parseTestFB2(t, "complex.fb2", 1)
	content := conv.Content()

	// "Часть вторая" has footnotes
	ch, err := conv.Chapter(content.ChapterIDs[3])
	require.NoError(t, err)

	assert.Contains(t, ch.HTML, `class="footnote-ref"`)
	assert.Contains(t, ch.HTML, `data-note-id="note1"`)
	assert.Contains(t, ch.HTML, `data-note-id="note2"`)
}

func TestFB2Converter_FootnoteBody(t *testing.T) {
	conv := parseTestFB2(t, "complex.fb2", 1)
	content := conv.Content()

	ch, err := conv.Chapter(content.ChapterIDs[3])
	require.NoError(t, err)

	assert.Contains(t, ch.HTML, `<div class="footnote-body" id="note1">`)
	assert.Contains(t, ch.HTML, "Это текст первой сноски")
	assert.Contains(t, ch.HTML, `<div class="footnote-body" id="note2">`)
	assert.Contains(t, ch.HTML, "Текст второй сноски")
}

func TestFB2Converter_FootnoteBody_HasFormatting(t *testing.T) {
	conv := parseTestFB2(t, "complex.fb2", 1)
	content := conv.Content()

	ch, err := conv.Chapter(content.ChapterIDs[3])
	require.NoError(t, err)

	// note2 has <emphasis>форматированием</emphasis> which should be converted to <em>
	assert.Contains(t, ch.HTML, "<em>форматированием</em>")
}

func TestFB2Converter_FootnoteBody_NotInOtherChapters(t *testing.T) {
	conv := parseTestFB2(t, "complex.fb2", 1)
	content := conv.Content()

	// Chapter without footnote refs should not have footnote bodies
	ch, err := conv.Chapter(content.ChapterIDs[0])
	require.NoError(t, err)

	assert.NotContains(t, ch.HTML, `class="footnote-body"`)
}

// --- Single Section Book ---

func TestFB2Converter_SingleSection(t *testing.T) {
	conv := parseTestFB2(t, "single_section.fb2", 1)
	content := conv.Content()

	assert.Equal(t, "Книга без оглавления", content.Metadata.Title)
	assert.Equal(t, "Пётр Одинов", content.Metadata.Author)
	assert.Equal(t, 1, content.TotalChapters)

	ch, err := conv.Chapter(content.ChapterIDs[0])
	require.NoError(t, err)

	assert.Contains(t, ch.HTML, "Это книга с единственной секцией")
	// No title tag since section has no title
	assert.NotContains(t, ch.HTML, `<h2 class="chapter-title">`)
}

// --- GetConverter ---

func TestGetConverter_FB2(t *testing.T) {
	conv, err := GetConverter("fb2")
	require.NoError(t, err)
	assert.IsType(t, &FB2Converter{}, conv)
}

func TestGetConverter_Unsupported(t *testing.T) {
	_, err := GetConverter("doc")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported format")
}

// --- fb2Author.FullName ---

func TestFB2Author_FullName(t *testing.T) {
	tests := []struct {
		name     string
		author   fb2Author
		expected string
	}{
		{
			name:     "full name",
			author:   fb2Author{FirstName: "Иван", MiddleName: "Иванович", LastName: "Тестов"},
			expected: "Иван Иванович Тестов",
		},
		{
			name:     "no middle name",
			author:   fb2Author{FirstName: "Иван", LastName: "Тестов"},
			expected: "Иван Тестов",
		},
		{
			name:     "only last name",
			author:   fb2Author{LastName: "Тестов"},
			expected: "Тестов",
		},
		{
			name:     "empty",
			author:   fb2Author{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.author.FullName())
		})
	}
}

// --- fb2Title.Text ---

func TestFB2Title_Text_Nil(t *testing.T) {
	var title *fb2Title
	assert.Equal(t, "", title.Text())
}

// --- extractAttrValue ---

func TestExtractAttrValue(t *testing.T) {
	tests := []struct {
		name     string
		tag      string
		attr     string
		expected string
	}{
		{
			name:     "simple attribute",
			tag:      `<a href="#note1" type="note">`,
			attr:     "href",
			expected: "#note1",
		},
		{
			name:     "namespaced attribute",
			tag:      `<a l:href="#note1" type="note">`,
			attr:     "href",
			expected: "#note1",
		},
		{
			name:     "type attribute",
			tag:      `<a l:href="#note1" type="note">`,
			attr:     "type",
			expected: "note",
		},
		{
			name:     "missing attribute",
			tag:      `<a l:href="#note1">`,
			attr:     "type",
			expected: "",
		},
		{
			name:     "empty tag",
			tag:      "",
			attr:     "href",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, extractAttrValue(tt.tag, tt.attr))
		})
	}
}

// --- convertInline ---

func TestFB2Converter_ConvertInline_Empty(t *testing.T) {
	conv := &FB2Converter{}
	assert.Equal(t, "", conv.convertInline(""))
}

func TestFB2Converter_ConvertInline_PlainText(t *testing.T) {
	conv := &FB2Converter{}
	result := conv.convertInline("Просто текст")
	assert.Equal(t, "Просто текст", result)
}

func TestFB2Converter_ConvertInline_AllTags(t *testing.T) {
	conv := &FB2Converter{}
	input := "<emphasis>em</emphasis> <strong>st</strong> <strikethrough>del</strikethrough> <code>cd</code> <sup>s</sup> <sub>b</sub>"
	result := conv.convertInline(input)

	assert.Contains(t, result, "<em>em</em>")
	assert.Contains(t, result, "<strong>st</strong>")
	assert.Contains(t, result, "<del>del</del>")
	assert.Contains(t, result, "<code>cd</code>")
	assert.Contains(t, result, "<sup>s</sup>")
	assert.Contains(t, result, "<sub>b</sub>")
}
