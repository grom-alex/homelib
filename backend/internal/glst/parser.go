package glst

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// ParseReader parses a .glst genre tree file from the given reader.
// Format: each line is either a comment (starts with #), empty, or a genre entry.
// Genre entry format: <position> <code>;<name>
// Entries without semicolon (e.g., "0.0 Неотсортированное") get an auto-generated code.
func ParseReader(r io.Reader) (*ParseResult, error) {
	scanner := bufio.NewScanner(r)
	result := &ParseResult{}

	// Track positions for duplicate detection and orphan validation
	positionIndex := make(map[string]int) // position → index in result.Entries
	rootCodes := make(map[string]string)  // code → position (root level only)

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimRight(scanner.Text(), " \t\r")

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		entry, err := parseLine(line, lineNum)
		if err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("line %d: %s", lineNum, err.Error()))
			continue
		}

		// Validate orphaned children: parent position must exist (except root level)
		if entry.Level > 0 {
			if _, parentExists := positionIndex[entry.ParentPosition]; !parentExists {
				result.Warnings = append(result.Warnings,
					fmt.Sprintf("line %d: orphaned child %s (parent %s not found)", lineNum, entry.Position, entry.ParentPosition))
				continue
			}
		}

		// Check root-level code duplicates
		if entry.Level == 0 && entry.Code != "" {
			if prevPos, exists := rootCodes[entry.Code]; exists {
				result.Warnings = append(result.Warnings,
					fmt.Sprintf("line %d: duplicate root-level code %q (already at %s)", lineNum, entry.Code, prevPos))
			}
			rootCodes[entry.Code] = entry.Position
		}

		// Handle duplicate positions: last one wins
		if existingIdx, exists := positionIndex[entry.Position]; exists {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("line %d: duplicate position %s (overwriting previous entry)", lineNum, entry.Position))
			result.Entries[existingIdx] = entry
		} else {
			positionIndex[entry.Position] = len(result.Entries)
			result.Entries = append(result.Entries, entry)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading glst: %w", err)
	}

	return result, nil
}

// parseLine parses a single genre line: "<position> <code>;<name>" or "<position> <name>"
func parseLine(line string, lineNum int) (GenreEntry, error) {
	// Split by first space: position + rest
	spaceIdx := strings.IndexByte(line, ' ')
	if spaceIdx < 0 {
		return GenreEntry{}, fmt.Errorf("invalid format: no space separator")
	}

	position := line[:spaceIdx]
	rest := strings.TrimSpace(line[spaceIdx+1:])

	if position == "" || rest == "" {
		return GenreEntry{}, fmt.Errorf("invalid format: empty position or description")
	}

	// Validate position format: must be digits separated by dots, starting with "0."
	if !isValidPosition(position) {
		return GenreEntry{}, fmt.Errorf("invalid position format: %q", position)
	}

	// Parse code and name
	var code, name string
	if semiIdx := strings.IndexByte(rest, ';'); semiIdx >= 0 {
		code = rest[:semiIdx]
		name = rest[semiIdx+1:]
	} else {
		// No semicolon: auto-generate code from position
		name = rest
		code = "_root_" + strings.ReplaceAll(position[2:], ".", "_") // e.g., "0.0" → "_root_0", "0.1.2" → "_root_1_2"
	}

	name = strings.TrimSpace(name)
	code = strings.TrimSpace(code)

	// Compute level: count dots minus 1 (e.g., "0.1" → 1 dot → level 0, "0.1.2" → 2 dots → level 1)
	level := strings.Count(position, ".") - 1

	// Compute parent position
	var parentPosition string
	if level > 0 {
		lastDot := strings.LastIndexByte(position, '.')
		parentPosition = position[:lastDot]
	}

	return GenreEntry{
		Position:       position,
		Code:           code,
		Name:           name,
		Level:          level,
		ParentPosition: parentPosition,
	}, nil
}

// isValidPosition checks if position matches pattern like "0.1", "0.1.2", "0.9.13.0"
func isValidPosition(pos string) bool {
	if !strings.HasPrefix(pos, "0.") {
		return false
	}
	parts := strings.Split(pos, ".")
	if len(parts) < 2 {
		return false
	}
	for _, p := range parts {
		if p == "" {
			return false
		}
		for _, c := range p {
			if c < '0' || c > '9' {
				return false
			}
		}
	}
	return true
}
