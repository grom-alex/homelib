package inpx

import (
	"archive/zip"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse_FullINPX(t *testing.T) {
	inpx := buildTestINPX(t, testINPXOptions{
		collectionInfo: "Test Library\ntest-lib\n0\nTest collection\nhttp://example.com\n",
		versionInfo:    "20220904\n",
		structureInfo:  "AUTHOR;GENRE;TITLE;SERIES;SERNO;FILE;SIZE;LIBID;DEL;EXT;DATE;LANG;LIBRATE;\n",
		inpFiles: map[string]string{
			"fb2-000001-000100.inp": "Булгаков,Михаил,Афанасьевич\x04sf_fantasy:dramaturgy\x04Мастер и Маргарита\x04\x04\x0494240\x04106260\x0494240\x04\x04fb2\x042007-06-20\x04ru\x045\r\n" +
				"Толстой,Лев,Николаевич\x04prose_classic\x04Война и мир\x04\x04\x0494241\x04500000\x0494241\x04\x04fb2\x042007-06-21\x04ru\x044\r\n",
		},
	})

	result, err := Parse(bytes.NewReader(inpx), int64(len(inpx)))
	require.NoError(t, err)

	assert.Equal(t, "Test Library", result.Collection.Name)
	assert.Equal(t, "test-lib", result.Collection.Code)
	assert.Equal(t, 0, result.Collection.Type)
	assert.Equal(t, "Test collection", result.Collection.Description)
	assert.Equal(t, "http://example.com", result.Collection.SourceURL)
	assert.Equal(t, "20220904", result.Version)

	assert.Len(t, result.Records, 2)

	rec1 := result.Records[0]
	assert.Equal(t, "Мастер и Маргарита", rec1.Title)
	assert.Len(t, rec1.Authors, 1)
	assert.Equal(t, "Булгаков", rec1.Authors[0].LastName)
	assert.Equal(t, []string{"sf_fantasy", "dramaturgy"}, rec1.Genres)
	assert.Equal(t, "94240", rec1.LibID)
	assert.Equal(t, "fb2", rec1.Extension)
	assert.Equal(t, "ru", rec1.Language)
	assert.Equal(t, 5, rec1.LibRate)
	assert.Equal(t, "fb2-000001-000100.zip", rec1.ArchiveName)
	assert.False(t, rec1.IsDeleted)

	rec2 := result.Records[1]
	assert.Equal(t, "Война и мир", rec2.Title)
	assert.Equal(t, "Толстой", rec2.Authors[0].LastName)
}

func TestParse_DeletedRecords(t *testing.T) {
	inpx := buildTestINPX(t, testINPXOptions{
		collectionInfo: "Test\ntest\n0\n\n\n",
		versionInfo:    "20220904\n",
		inpFiles: map[string]string{
			"test.inp": "Author,Name,\x04genre\x04Normal Book\x04\x04\x0412345\x04100\x0412345\x04\x04fb2\x042020-01-01\r\n" +
				"Author,Name,\x04genre\x04Deleted Book\x04\x04\x0412346\x04100\x0412346\x041\x04fb2\x042020-01-01\r\n",
		},
	})

	result, err := Parse(bytes.NewReader(inpx), int64(len(inpx)))
	require.NoError(t, err)

	assert.Len(t, result.Records, 2)
	assert.False(t, result.Records[0].IsDeleted)
	assert.True(t, result.Records[1].IsDeleted)
}

func TestParse_SkipsEmptyTitleAndFile(t *testing.T) {
	inpx := buildTestINPX(t, testINPXOptions{
		collectionInfo: "Test\ntest\n0\n\n\n",
		versionInfo:    "20220904\n",
		inpFiles: map[string]string{
			"test.inp": "\x04genre\x04\x04\x04\x04\x04100\x0411111\x04\x04fb2\x042020-01-01\r\n" + // no title, no file
				"Author,Name,\x04genre\x04Good Book\x04\x04\x0412345\x04100\x0412345\x04\x04fb2\x042020-01-01\r\n",
		},
	})

	result, err := Parse(bytes.NewReader(inpx), int64(len(inpx)))
	require.NoError(t, err)

	assert.Len(t, result.Records, 1)
	assert.Equal(t, "Good Book", result.Records[0].Title)
}

func TestParse_CustomStructureInfo(t *testing.T) {
	inpx := buildTestINPX(t, testINPXOptions{
		collectionInfo: "Test\ntest\n0\n\n\n",
		versionInfo:    "20220904\n",
		structureInfo:  "TITLE;AUTHOR;GENRE;FILE;SIZE;LIBID;DEL;EXT;DATE;\n",
		inpFiles: map[string]string{
			"test.inp": "Custom Title\x04Author,Name,\x04genre\x0412345\x04100\x0412345\x04\x04fb2\x042020-01-01\r\n",
		},
	})

	result, err := Parse(bytes.NewReader(inpx), int64(len(inpx)))
	require.NoError(t, err)

	assert.Len(t, result.Records, 1)
	assert.Equal(t, "Custom Title", result.Records[0].Title)
	assert.Equal(t, "Author", result.Records[0].Authors[0].LastName)
}

func TestParse_MissingCollectionInfo(t *testing.T) {
	inpx := buildTestINPX(t, testINPXOptions{
		versionInfo: "20220904\n",
		inpFiles:    map[string]string{"test.inp": "A,B,\x04g\x04Title\x04\x04\x04f\x0410\x04123\x04\x04fb2\x042020-01-01\r\n"},
	})

	_, err := Parse(bytes.NewReader(inpx), int64(len(inpx)))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "collection.info")
}

type testINPXOptions struct {
	collectionInfo string
	versionInfo    string
	structureInfo  string
	inpFiles       map[string]string
}

func buildTestINPX(t *testing.T, opts testINPXOptions) []byte {
	t.Helper()

	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	if opts.collectionInfo != "" {
		f, err := w.Create("collection.info")
		require.NoError(t, err)
		_, err = f.Write([]byte(opts.collectionInfo))
		require.NoError(t, err)
	}

	if opts.versionInfo != "" {
		f, err := w.Create("version.info")
		require.NoError(t, err)
		_, err = f.Write([]byte(opts.versionInfo))
		require.NoError(t, err)
	}

	if opts.structureInfo != "" {
		f, err := w.Create("structure.info")
		require.NoError(t, err)
		_, err = f.Write([]byte(opts.structureInfo))
		require.NoError(t, err)
	}

	for name, content := range opts.inpFiles {
		f, err := w.Create(name)
		require.NoError(t, err)
		_, err = f.Write([]byte(content))
		require.NoError(t, err)
	}

	require.NoError(t, w.Close())
	return buf.Bytes()
}
