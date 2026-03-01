package glst

import (
	"strings"
	"testing"
)

func TestParseReader_ValidFile(t *testing.T) {
	input := `#genres list
0.0 Неотсортированное
0.1 sf_all;Фантастика
0.1.1 sf_history;Альтернативная история
0.1.2 sf_action;Боевая фантастика
`
	result, err := ParseReader(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Entries) != 4 {
		t.Fatalf("expected 4 entries, got %d", len(result.Entries))
	}

	// Check root entry without code
	e0 := result.Entries[0]
	if e0.Position != "0.0" {
		t.Errorf("entry[0].Position = %q, want %q", e0.Position, "0.0")
	}
	if e0.Code != "_root_0" {
		t.Errorf("entry[0].Code = %q, want %q", e0.Code, "_root_0")
	}
	if e0.Name != "Неотсортированное" {
		t.Errorf("entry[0].Name = %q, want %q", e0.Name, "Неотсортированное")
	}
	if e0.Level != 0 {
		t.Errorf("entry[0].Level = %d, want 0", e0.Level)
	}
	if e0.ParentPosition != "" {
		t.Errorf("entry[0].ParentPosition = %q, want empty", e0.ParentPosition)
	}

	// Check root entry with code
	e1 := result.Entries[1]
	if e1.Position != "0.1" {
		t.Errorf("entry[1].Position = %q, want %q", e1.Position, "0.1")
	}
	if e1.Code != "sf_all" {
		t.Errorf("entry[1].Code = %q, want %q", e1.Code, "sf_all")
	}
	if e1.Name != "Фантастика" {
		t.Errorf("entry[1].Name = %q, want %q", e1.Name, "Фантастика")
	}
	if e1.Level != 0 {
		t.Errorf("entry[1].Level = %d, want 0", e1.Level)
	}

	// Check level 1 entry
	e2 := result.Entries[2]
	if e2.Level != 1 {
		t.Errorf("entry[2].Level = %d, want 1", e2.Level)
	}
	if e2.ParentPosition != "0.1" {
		t.Errorf("entry[2].ParentPosition = %q, want %q", e2.ParentPosition, "0.1")
	}
}

func TestParseReader_SkipCommentsAndEmpty(t *testing.T) {
	input := `#comment line
0.1 sf_all;Фантастика

#another comment
0.2 det_all;Детективы
`
	result, err := ParseReader(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result.Entries))
	}
	if len(result.Warnings) != 0 {
		t.Errorf("expected 0 warnings, got %d: %v", len(result.Warnings), result.Warnings)
	}
}

func TestParseReader_EntryWithoutCode(t *testing.T) {
	input := `0.0 Неотсортированное
0.5 some_code;Приключения
`
	result, err := ParseReader(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result.Entries))
	}
	// Entry without semicolon should get auto-generated code
	if result.Entries[0].Code != "_root_0" {
		t.Errorf("expected auto-generated code %q, got %q", "_root_0", result.Entries[0].Code)
	}
	if result.Entries[0].Name != "Неотсортированное" {
		t.Errorf("expected name %q, got %q", "Неотсортированное", result.Entries[0].Name)
	}
}

func TestParseReader_SyntaxErrors(t *testing.T) {
	input := `0.1 sf_all;Фантастика
bad line without position
0.2 det_all;Детективы
`
	result, err := ParseReader(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Entries) != 2 {
		t.Fatalf("expected 2 entries (bad line skipped), got %d", len(result.Entries))
	}
	if len(result.Warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d", len(result.Warnings))
	}
	if !strings.Contains(result.Warnings[0], "line 2") {
		t.Errorf("warning should reference line 2, got: %s", result.Warnings[0])
	}
}

func TestParseReader_RootLevelCodeDuplicates(t *testing.T) {
	input := `0.1 sf_all;Фантастика
0.2 sf_all;Дубликат кода
`
	result, err := ParseReader(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should warn about root-level duplicate code
	if len(result.Warnings) == 0 {
		t.Fatal("expected warning about root-level duplicate code")
	}
	foundDupWarning := false
	for _, w := range result.Warnings {
		if strings.Contains(w, "duplicate") && strings.Contains(w, "sf_all") {
			foundDupWarning = true
			break
		}
	}
	if !foundDupWarning {
		t.Errorf("expected duplicate warning for sf_all, got: %v", result.Warnings)
	}
}

func TestParseReader_OrphanedChildren(t *testing.T) {
	input := `0.1 sf_all;Фантастика
0.27.0 orphan;Сирота
`
	result, err := ParseReader(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Orphan should be skipped with warning (parent 0.27 does not exist)
	if len(result.Entries) != 1 {
		t.Fatalf("expected 1 entry (orphan skipped), got %d", len(result.Entries))
	}
	if len(result.Warnings) == 0 {
		t.Fatal("expected warning about orphaned child")
	}
	foundOrphanWarning := false
	for _, w := range result.Warnings {
		if strings.Contains(w, "orphan") || strings.Contains(w, "parent") {
			foundOrphanWarning = true
			break
		}
	}
	if !foundOrphanWarning {
		t.Errorf("expected orphan warning, got: %v", result.Warnings)
	}
}

func TestParseReader_DuplicatePositions(t *testing.T) {
	input := `0.1 sf_all;Фантастика
0.1 sf_dup;Дубликат позиции
`
	result, err := ParseReader(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Last entry with same position wins
	if len(result.Entries) != 1 {
		t.Fatalf("expected 1 entry (duplicate replaced), got %d", len(result.Entries))
	}
	if result.Entries[0].Code != "sf_dup" {
		t.Errorf("expected last entry to win, got code %q", result.Entries[0].Code)
	}
	if len(result.Warnings) == 0 {
		t.Fatal("expected warning about duplicate position")
	}
}

func TestParseReader_LevelComputation(t *testing.T) {
	input := `0.1 sf_all;Фантастика
0.1.1 sf_history;Альтернативная история
0.9 sci_all;Наука
0.9.13 sci_chem;Химия
0.9.13.0 sci_hem_general;Общая химия
`
	result, err := ParseReader(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Entries) != 5 {
		t.Fatalf("expected 5 entries, got %d", len(result.Entries))
	}

	tests := []struct {
		idx            int
		position       string
		level          int
		parentPosition string
	}{
		{0, "0.1", 0, ""},
		{1, "0.1.1", 1, "0.1"},
		{2, "0.9", 0, ""},
		{3, "0.9.13", 1, "0.9"},
		{4, "0.9.13.0", 2, "0.9.13"},
	}

	for _, tc := range tests {
		e := result.Entries[tc.idx]
		if e.Level != tc.level {
			t.Errorf("entry %q: Level = %d, want %d", tc.position, e.Level, tc.level)
		}
		if e.ParentPosition != tc.parentPosition {
			t.Errorf("entry %q: ParentPosition = %q, want %q", tc.position, e.ParentPosition, tc.parentPosition)
		}
	}
}

func TestParseReader_EmptyInput(t *testing.T) {
	result, err := ParseReader(strings.NewReader(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(result.Entries))
	}
}

func TestParseReader_TrailingWhitespace(t *testing.T) {
	input := "0.1 sf_all;Фантастика \t\n0.2 det_all;Детективы\t\n"
	result, err := ParseReader(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result.Entries))
	}
	if result.Entries[0].Name != "Фантастика" {
		t.Errorf("expected trimmed name %q, got %q", "Фантастика", result.Entries[0].Name)
	}
}
