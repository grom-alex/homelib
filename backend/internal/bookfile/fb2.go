package bookfile

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"html"
	"strings"
)

// FB2 XML structures

type fb2FictionBook struct {
	XMLName     xml.Name        `xml:"FictionBook"`
	Description fb2Description  `xml:"description"`
	Bodies      []fb2Body       `xml:"body"`
	Binaries    []fb2Binary     `xml:"binary"`
}

type fb2Description struct {
	TitleInfo fb2TitleInfo `xml:"title-info"`
}

type fb2TitleInfo struct {
	Genres    []string       `xml:"genre"`
	Authors   []fb2Author    `xml:"author"`
	BookTitle string         `xml:"book-title"`
	Lang      string         `xml:"lang"`
	Coverpage *fb2Coverpage  `xml:"coverpage"`
}

type fb2Author struct {
	FirstName  string `xml:"first-name"`
	MiddleName string `xml:"middle-name"`
	LastName   string `xml:"last-name"`
}

func (a fb2Author) FullName() string {
	parts := make([]string, 0, 3)
	if a.FirstName != "" {
		parts = append(parts, a.FirstName)
	}
	if a.MiddleName != "" {
		parts = append(parts, a.MiddleName)
	}
	if a.LastName != "" {
		parts = append(parts, a.LastName)
	}
	return strings.Join(parts, " ")
}

type fb2Coverpage struct {
	Images []fb2Image `xml:"image"`
}

type fb2Body struct {
	Name     string       `xml:"name,attr"`
	Title    *fb2Title    `xml:"title"`
	Sections []fb2Section `xml:"section"`
}

type fb2Section struct {
	ID        string        `xml:"id,attr"`
	Title     *fb2Title     `xml:"title"`
	Epigraphs []fb2Epigraph `xml:"epigraph"`
	Content   []fb2Element  `xml:",any"`
	Sections  []fb2Section  `xml:"section"`
}

type fb2Title struct {
	Paragraphs []fb2Paragraph `xml:"p"`
}

func (t *fb2Title) Text() string {
	if t == nil {
		return ""
	}
	parts := make([]string, 0, len(t.Paragraphs))
	for _, p := range t.Paragraphs {
		parts = append(parts, p.Text())
	}
	return strings.Join(parts, " ")
}

type fb2Epigraph struct {
	Paragraphs []fb2Paragraph `xml:"p"`
	TextAuthor string         `xml:"text-author"`
}

type fb2Paragraph struct {
	XMLName xml.Name
	Content string `xml:",innerxml"`
}

func (p fb2Paragraph) Text() string {
	// Strip XML tags for plain text
	s := p.Content
	for {
		start := strings.Index(s, "<")
		if start == -1 {
			break
		}
		end := strings.Index(s[start:], ">")
		if end == -1 {
			break
		}
		s = s[:start] + s[start+end+1:]
	}
	return strings.TrimSpace(html.UnescapeString(s))
}

type fb2Element struct {
	XMLName xml.Name
	Content string `xml:",innerxml"`
	Attrs   []xml.Attr `xml:",any,attr"`
}

type fb2Poem struct {
	Title      *fb2Title       `xml:"title"`
	Stanzas    []fb2Stanza     `xml:"stanza"`
	TextAuthor string          `xml:"text-author"`
}

type fb2Stanza struct {
	Verses []fb2Paragraph `xml:"v"`
}

type fb2Cite struct {
	Paragraphs []fb2Paragraph `xml:"p"`
	TextAuthor string         `xml:"text-author"`
}

type fb2Image struct {
	Href string `xml:"href,attr"`
}

type fb2Binary struct {
	ID          string `xml:"id,attr"`
	ContentType string `xml:"content-type,attr"`
	Data        string `xml:",chardata"`
}

type fb2Link struct {
	Href    string `xml:"href,attr"`
	Type    string `xml:"type,attr"`
	Content string `xml:",innerxml"`
}

// FB2Converter implements BookConverter for FB2 format.
type FB2Converter struct {
	book    *fb2FictionBook
	bookID  int64
	content *BookContent
	// Pre-parsed chapter data: chapterID -> section
	chapters map[string]*fb2Section
	// Note bodies from <body name="notes">
	notes map[string]*fb2Section
}

func (c *FB2Converter) Parse(data []byte, bookID int64) error {
	c.bookID = bookID
	c.book = &fb2FictionBook{}
	if err := xml.Unmarshal(data, c.book); err != nil {
		return fmt.Errorf("parse FB2 XML: %w", err)
	}

	c.chapters = make(map[string]*fb2Section)
	c.notes = make(map[string]*fb2Section)

	// Extract notes from <body name="notes">
	for i := range c.book.Bodies {
		if c.book.Bodies[i].Name == "notes" {
			for j := range c.book.Bodies[i].Sections {
				sec := &c.book.Bodies[i].Sections[j]
				if sec.ID != "" {
					c.notes[sec.ID] = sec
				}
			}
		}
	}

	// Build TOC and chapter index from main body
	var toc []TOCEntry
	var chapterIDs []string
	chapterCounter := 0

	for i := range c.book.Bodies {
		if c.book.Bodies[i].Name != "" {
			continue // skip notes and other named bodies
		}
		c.buildTOC(&c.book.Bodies[i].Sections, 0, &toc, &chapterIDs, &chapterCounter)
	}

	// Build metadata
	ti := c.book.Description.TitleInfo
	author := ""
	if len(ti.Authors) > 0 {
		author = ti.Authors[0].FullName()
	}

	var coverURL string
	if ti.Coverpage != nil && len(ti.Coverpage.Images) > 0 {
		href := strings.TrimPrefix(ti.Coverpage.Images[0].Href, "#")
		if href != "" {
			coverURL = fmt.Sprintf("/api/books/%d/image/%s", bookID, href)
		}
	}

	c.content = &BookContent{
		Metadata: BookMetadata{
			Title:    ti.BookTitle,
			Author:   author,
			Cover:    coverURL,
			Language: ti.Lang,
			Format:   "fb2",
		},
		TOC:           toc,
		ChapterIDs:    chapterIDs,
		TotalChapters: len(chapterIDs),
	}

	// Pre-compute chapter HTML sizes for page estimation
	sizes := make(map[string]int, len(chapterIDs))
	for _, id := range chapterIDs {
		if sec, ok := c.chapters[id]; ok {
			sizes[id] = len(c.convertSection(sec))
		}
	}
	// Add cover image size to first chapter estimate
	if coverURL != "" && len(chapterIDs) > 0 {
		sizes[chapterIDs[0]] += 2000 // approximate cover <img> rendered size
	}
	c.content.ChapterSizes = sizes

	return nil
}

func (c *FB2Converter) buildTOC(sections *[]fb2Section, level int, toc *[]TOCEntry, ids *[]string, counter *int) {
	for i := range *sections {
		sec := &(*sections)[i]
		*counter++
		if sec.ID == "" {
			sec.ID = fmt.Sprintf("ch%d", *counter)
		}

		title := ""
		if sec.Title != nil {
			title = sec.Title.Text()
		}
		if title == "" {
			title = fmt.Sprintf("Глава %d", *counter)
		}

		*toc = append(*toc, TOCEntry{
			ID:    sec.ID,
			Title: title,
			Level: level,
		})
		*ids = append(*ids, sec.ID)
		c.chapters[sec.ID] = sec

		if len(sec.Sections) > 0 {
			c.buildTOC(&sec.Sections, level+1, toc, ids, counter)
		}
	}
}

func (c *FB2Converter) Content() *BookContent {
	return c.content
}

func (c *FB2Converter) Chapter(chapterID string) (*ChapterContent, error) {
	sec, ok := c.chapters[chapterID]
	if !ok {
		return nil, fmt.Errorf("chapter %q not found", chapterID)
	}

	title := ""
	if sec.Title != nil {
		title = sec.Title.Text()
	}

	htmlContent := c.convertSection(sec)

	// Prepend cover image to the first chapter
	if c.content != nil && len(c.content.ChapterIDs) > 0 &&
		chapterID == c.content.ChapterIDs[0] && c.content.Metadata.Cover != "" {
		cover := fmt.Sprintf("<div class=\"book-cover\"><img src=\"%s\" alt=\"%s\" /></div>\n",
			c.content.Metadata.Cover, c.content.Metadata.Title)
		htmlContent = cover + htmlContent
	}

	return &ChapterContent{
		ID:    chapterID,
		Title: title,
		HTML:  htmlContent,
	}, nil
}

func (c *FB2Converter) Image(imageID string) (*ImageData, error) {
	for _, bin := range c.book.Binaries {
		if bin.ID == imageID {
			data, err := base64.StdEncoding.DecodeString(strings.TrimSpace(bin.Data))
			if err != nil {
				return nil, fmt.Errorf("decode image %s: %w", imageID, err)
			}
			return &ImageData{
				ID:          imageID,
				ContentType: bin.ContentType,
				Data:        data,
			}, nil
		}
	}
	return nil, fmt.Errorf("image %q not found", imageID)
}

// convertSection renders a section to HTML.
func (c *FB2Converter) convertSection(sec *fb2Section) string {
	var b strings.Builder

	// Title
	if sec.Title != nil {
		b.WriteString(`<h2 class="chapter-title">`)
		for _, p := range sec.Title.Paragraphs {
			b.WriteString(html.EscapeString(p.Text()))
			b.WriteString(" ")
		}
		b.WriteString("</h2>\n")
	}

	// Epigraphs
	for _, ep := range sec.Epigraphs {
		b.WriteString(c.convertEpigraph(&ep))
	}

	// Content elements
	for _, elem := range sec.Content {
		switch elem.XMLName.Local {
		case "p":
			b.WriteString("<p>")
			b.WriteString(c.convertInline(elem.Content))
			b.WriteString("</p>\n")
		case "poem":
			b.WriteString(c.convertPoemFromXML(elem.Content))
		case "cite":
			b.WriteString(c.convertCiteFromXML(elem.Content))
		case "subtitle":
			b.WriteString(`<p class="subtitle">`)
			b.WriteString(c.convertInline(elem.Content))
			b.WriteString("</p>\n")
		case "empty-line":
			b.WriteString("<br/>\n")
		case "image":
			b.WriteString(c.convertImageElem(&elem))
		case "epigraph":
			// already handled above via sec.Epigraphs
		case "section":
			// nested sections handled via TOC building
		case "title":
			// already handled above
		}
	}

	// Append footnote bodies referenced in this chapter
	c.appendFootnoteBodies(&b, sec)

	return b.String()
}

func (c *FB2Converter) convertEpigraph(ep *fb2Epigraph) string {
	var b strings.Builder
	b.WriteString(`<blockquote class="epigraph">`)
	for _, p := range ep.Paragraphs {
		b.WriteString("<p>")
		b.WriteString(c.convertInline(p.Content))
		b.WriteString("</p>")
	}
	if ep.TextAuthor != "" {
		b.WriteString(`<p class="epigraph-author">`)
		b.WriteString(html.EscapeString(ep.TextAuthor))
		b.WriteString("</p>")
	}
	b.WriteString("</blockquote>\n")
	return b.String()
}

func (c *FB2Converter) convertPoemFromXML(innerXML string) string {
	var poem fb2Poem
	wrapped := "<poem>" + innerXML + "</poem>"
	if err := xml.Unmarshal([]byte(wrapped), &poem); err != nil {
		return "<p>" + html.EscapeString(innerXML) + "</p>"
	}

	var b strings.Builder
	b.WriteString(`<div class="poem">`)
	if poem.Title != nil {
		b.WriteString(`<p class="subtitle">`)
		b.WriteString(html.EscapeString(poem.Title.Text()))
		b.WriteString("</p>")
	}
	for _, st := range poem.Stanzas {
		b.WriteString(`<div class="stanza">`)
		for _, v := range st.Verses {
			b.WriteString(`<p class="verse">`)
			b.WriteString(c.convertInline(v.Content))
			b.WriteString("</p>")
		}
		b.WriteString("</div>")
	}
	if poem.TextAuthor != "" {
		b.WriteString(`<p class="poem-author">`)
		b.WriteString(html.EscapeString(poem.TextAuthor))
		b.WriteString("</p>")
	}
	b.WriteString("</div>\n")
	return b.String()
}

func (c *FB2Converter) convertCiteFromXML(innerXML string) string {
	var cite fb2Cite
	wrapped := "<cite>" + innerXML + "</cite>"
	if err := xml.Unmarshal([]byte(wrapped), &cite); err != nil {
		return "<p>" + html.EscapeString(innerXML) + "</p>"
	}

	var b strings.Builder
	b.WriteString(`<blockquote class="cite">`)
	for _, p := range cite.Paragraphs {
		b.WriteString("<p>")
		b.WriteString(c.convertInline(p.Content))
		b.WriteString("</p>")
	}
	if cite.TextAuthor != "" {
		b.WriteString(`<p class="epigraph-author">`)
		b.WriteString(html.EscapeString(cite.TextAuthor))
		b.WriteString("</p>")
	}
	b.WriteString("</blockquote>\n")
	return b.String()
}

func (c *FB2Converter) convertImageElem(elem *fb2Element) string {
	for _, attr := range elem.Attrs {
		if attr.Name.Local == "href" {
			href := strings.TrimPrefix(attr.Value, "#")
			return fmt.Sprintf(`<img src="/api/books/%d/image/%s" alt="" loading="lazy"/>`, c.bookID, href) + "\n"
		}
	}
	return ""
}

// convertInline converts FB2 inline markup to HTML.
func (c *FB2Converter) convertInline(content string) string {
	if content == "" {
		return ""
	}

	// Replace FB2 inline tags with HTML equivalents per §8.3 tag mapping
	r := strings.NewReplacer(
		"<emphasis>", "<em>",
		"</emphasis>", "</em>",
		"<strong>", "<strong>",
		"</strong>", "</strong>",
		"<strikethrough>", "<del>",
		"</strikethrough>", "</del>",
		"<code>", "<code>",
		"</code>", "</code>",
		"<sup>", "<sup>",
		"</sup>", "</sup>",
		"<sub>", "<sub>",
		"</sub>", "</sub>",
	)
	result := r.Replace(content)

	// Convert footnote references: <a ... type="note" ...> → <a class="footnote-ref" data-note-id="...">
	result = c.convertFootnoteRefs(result)

	// Convert image references inside paragraphs
	result = c.convertInlineImages(result)

	return result
}

// convertFootnoteRefs converts <a type="note" l:href="#noteID">text</a> to footnote-ref links.
func (c *FB2Converter) convertFootnoteRefs(content string) string {
	// Simple parser for <a ...type="note"...> elements
	var result strings.Builder
	remaining := content

	for {
		idx := strings.Index(remaining, "<a ")
		if idx == -1 {
			result.WriteString(remaining)
			break
		}

		result.WriteString(remaining[:idx])
		remaining = remaining[idx:]

		endTag := strings.Index(remaining, ">")
		if endTag == -1 {
			result.WriteString(remaining)
			break
		}

		tagContent := remaining[:endTag+1]
		remaining = remaining[endTag+1:]

		// Check if this is a note link
		if strings.Contains(tagContent, `type="note"`) {
			href := extractAttrValue(tagContent, "href")
			noteID := strings.TrimPrefix(href, "#")
			result.WriteString(fmt.Sprintf(`<a class="footnote-ref" data-note-id="%s">`, html.EscapeString(noteID)))
		} else {
			result.WriteString(tagContent)
		}
	}

	return result.String()
}

// convertInlineImages replaces <image l:href="#id"/> with <img> tags inside paragraphs.
func (c *FB2Converter) convertInlineImages(content string) string {
	var result strings.Builder
	remaining := content

	for {
		idx := strings.Index(remaining, "<image ")
		if idx == -1 {
			result.WriteString(remaining)
			break
		}

		result.WriteString(remaining[:idx])
		remaining = remaining[idx:]

		endTag := strings.Index(remaining, "/>")
		if endTag == -1 {
			endTag = strings.Index(remaining, ">")
			if endTag == -1 {
				result.WriteString(remaining)
				break
			}
		}

		tagContent := remaining[:endTag]
		if strings.Contains(remaining[endTag:endTag+2], "/>") {
			remaining = remaining[endTag+2:]
		} else {
			remaining = remaining[endTag+1:]
		}

		href := extractAttrValue(tagContent, "href")
		imgID := strings.TrimPrefix(href, "#")
		if imgID != "" {
			result.WriteString(fmt.Sprintf(`<img src="/api/books/%d/image/%s" alt="" loading="lazy"/>`, c.bookID, imgID))
		}
	}

	return result.String()
}

// appendFootnoteBodies appends hidden footnote body divs for all notes referenced in the section.
func (c *FB2Converter) appendFootnoteBodies(b *strings.Builder, sec *fb2Section) {
	if len(c.notes) == 0 {
		return
	}

	// Scan the section content for footnote references
	sectionHTML := b.String()
	for noteID, noteSec := range c.notes {
		marker := fmt.Sprintf(`data-note-id="%s"`, noteID)
		if strings.Contains(sectionHTML, marker) {
			b.WriteString(fmt.Sprintf(`<div class="footnote-body" id="%s">`, html.EscapeString(noteID)))
			for _, elem := range noteSec.Content {
				if elem.XMLName.Local == "p" {
					b.WriteString("<p>")
					b.WriteString(c.convertInline(elem.Content))
					b.WriteString("</p>")
				}
			}
			b.WriteString("</div>\n")
		}
	}
}

// extractAttrValue extracts the value of an attribute from an XML opening tag string.
func extractAttrValue(tag, attrName string) string {
	// Look for attrName="value" or :attrName="value" (namespaced)
	patterns := []string{
		attrName + `="`,
		":" + attrName + `="`,
	}
	for _, pattern := range patterns {
		idx := strings.Index(tag, pattern)
		if idx == -1 {
			continue
		}
		start := idx + len(pattern)
		end := strings.Index(tag[start:], `"`)
		if end == -1 {
			continue
		}
		return tag[start : start+end]
	}
	return ""
}
