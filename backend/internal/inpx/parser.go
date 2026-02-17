package inpx

import (
	"archive/zip"
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// ParseResult holds the complete result of parsing an INPX file.
type ParseResult struct {
	Collection CollectionInfo
	Version    string
	Records    []BookRecord
}

// Parse reads an INPX file (ZIP archive) and returns all parsed book records.
func Parse(reader io.ReaderAt, size int64) (*ParseResult, error) {
	zr, err := zip.NewReader(reader, size)
	if err != nil {
		return nil, fmt.Errorf("open zip: %w", err)
	}

	result := &ParseResult{}

	// Read metadata files
	result.Collection, err = readCollectionInfo(zr)
	if err != nil {
		return nil, fmt.Errorf("read collection.info: %w", err)
	}

	result.Version, _ = readVersionInfo(zr)

	mapping := DefaultFieldMapping
	if sm, err := readStructureInfo(zr); err == nil {
		mapping = sm
	}

	// Parse all .inp files
	for _, f := range zr.File {
		if !strings.HasSuffix(f.Name, ".inp") {
			continue
		}

		// Default archive name derived from .inp filename
		defaultArchive := strings.TrimSuffix(f.Name, ".inp") + ".zip"

		records, err := parseInpFile(f, mapping, defaultArchive)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", f.Name, err)
		}
		result.Records = append(result.Records, records...)
	}

	return result, nil
}

func readCollectionInfo(zr *zip.Reader) (CollectionInfo, error) {
	f := findFile(zr, "collection.info")
	if f == nil {
		return CollectionInfo{}, fmt.Errorf("collection.info not found")
	}

	lines, err := readLines(f)
	if err != nil {
		return CollectionInfo{}, err
	}

	ci := CollectionInfo{}
	if len(lines) > 0 {
		ci.Name = lines[0]
	}
	if len(lines) > 1 {
		ci.Code = lines[1]
	}
	if len(lines) > 2 {
		ci.Type, _ = strconv.Atoi(lines[2])
	}
	if len(lines) > 3 {
		ci.Description = lines[3]
	}
	if len(lines) > 4 {
		ci.SourceURL = lines[4]
	}

	return ci, nil
}

func readVersionInfo(zr *zip.Reader) (string, error) {
	f := findFile(zr, "version.info")
	if f == nil {
		return "", fmt.Errorf("version.info not found")
	}

	lines, err := readLines(f)
	if err != nil {
		return "", err
	}

	if len(lines) > 0 {
		return strings.TrimSpace(lines[0]), nil
	}
	return "", nil
}

func readStructureInfo(zr *zip.Reader) (FieldMapping, error) {
	f := findFile(zr, "structure.info")
	if f == nil {
		return FieldMapping{}, fmt.Errorf("structure.info not found")
	}

	lines, err := readLines(f)
	if err != nil {
		return FieldMapping{}, err
	}

	if len(lines) == 0 {
		return FieldMapping{}, fmt.Errorf("structure.info is empty")
	}

	// Fields separated by semicolons
	raw := strings.TrimRight(lines[0], ";")
	fields := strings.Split(raw, ";")
	for i := range fields {
		fields[i] = strings.TrimSpace(fields[i])
	}

	return NewFieldMapping(fields), nil
}

func parseInpFile(f *zip.File, mapping FieldMapping, defaultArchive string) ([]BookRecord, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	defer func() { _ = rc.Close() }()

	var records []BookRecord
	scanner := bufio.NewScanner(rc)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024) // 1MB buffer for long lines

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimRight(line, "\r\n")
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Split(line, "\x04")
		rec := ParseRecord(fields, mapping, defaultArchive)

		// Skip records without essential fields
		if rec.Title == "" || rec.FileName == "" {
			continue
		}

		records = append(records, rec)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan: %w", err)
	}

	return records, nil
}

func findFile(zr *zip.Reader, name string) *zip.File {
	for _, f := range zr.File {
		if strings.EqualFold(f.Name, name) {
			return f
		}
	}
	return nil
}

func readLines(f *zip.File) ([]string, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer func() { _ = rc.Close() }()

	var lines []string
	scanner := bufio.NewScanner(rc)
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\r")
		lines = append(lines, line)
	}
	return lines, scanner.Err()
}
