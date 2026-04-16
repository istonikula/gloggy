package entrylist

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/istonikula/gloggy/internal/config"
	"github.com/istonikula/gloggy/internal/logsource"
	"github.com/istonikula/gloggy/internal/theme"
)

func makeEntries(n int) []logsource.Entry {
	entries := make([]logsource.Entry, n)
	for i := range entries {
		entries[i] = logsource.Entry{
			IsJSON:     true,
			LineNumber: i + 1,
			Time:       time.Now(),
			Level:      "INFO",
			Msg:        fmt.Sprintf("entry %d", i),
			Raw:        []byte(fmt.Sprintf(`{"msg":"entry %d"}`, i)),
		}
	}
	return entries
}

func defaultListModel(height int) ListModel {
	return NewListModel(theme.GetTheme("tokyo-night"), config.DefaultConfig(), 80, height)
}

// T-029: R6.1 — rendered rows ≤ viewport height for 100k entries
func TestListModel_VirtualRendering_100k(t *testing.T) {
	const total = 100_000
	const height = 40
	m := defaultListModel(height).SetEntries(makeEntries(total))

	count := m.RenderedRowCount()
	if count > height {
		t.Errorf("rendered %d rows for 100k entries, max expected %d (viewport height)", count, height)
	}
	if count == 0 {
		t.Error("rendered 0 rows — expected at least some")
	}
}

// T-029: R6.2 — render time < 16ms for large dataset
func TestListModel_RenderSpeed_100k(t *testing.T) {
	const total = 100_000
	const height = 40
	m := defaultListModel(height).SetEntries(makeEntries(total))

	start := time.Now()
	_ = m.View()
	elapsed := time.Since(start)

	if elapsed > 16*time.Millisecond {
		t.Errorf("View() took %v, expected < 16ms", elapsed)
	}
}

// Virtual rendering: scrolled to middle still only renders viewport height
func TestListModel_VirtualRendering_Scrolled(t *testing.T) {
	const total = 1000
	const height = 20
	m := defaultListModel(height).SetEntries(makeEntries(total))

	// Scroll to middle.
	m.scroll.Cursor = 500
	m.scroll.Offset = 490

	count := m.RenderedRowCount()
	if count > height {
		t.Errorf("rendered %d rows when scrolled, max expected %d", count, height)
	}
}

// No SelectionMsg when cursor doesn't move (already at boundary).
func TestListModel_NoSelectionMsg_AtBoundary(t *testing.T) {
	entries := makeEntries(5)
	m := defaultListModel(10).SetEntries(entries)

	// Already at top — k should not emit.
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	if cmd != nil {
		t.Error("expected nil cmd when cursor cannot move")
	}
}

// T-080: R11 — CursorPosition returns 1-based index
func TestCursorPosition_EmptyList(t *testing.T) {
	m := defaultListModel(10)
	if pos := m.CursorPosition(); pos != 0 {
		t.Fatalf("CursorPosition on empty = %d, want 0", pos)
	}
}

func TestCursorPosition_JK(t *testing.T) {
	m := defaultListModel(10).SetEntries(makeEntries(5))
	if pos := m.CursorPosition(); pos != 1 {
		t.Fatalf("initial CursorPosition = %d, want 1", pos)
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if pos := m.CursorPosition(); pos != 2 {
		t.Fatalf("after j: CursorPosition = %d, want 2", pos)
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	if pos := m.CursorPosition(); pos != 1 {
		t.Fatalf("after k: CursorPosition = %d, want 1", pos)
	}
}

func TestCursorPosition_gG(t *testing.T) {
	m := defaultListModel(10).SetEntries(makeEntries(5))
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	if pos := m.CursorPosition(); pos != 5 {
		t.Fatalf("after G: CursorPosition = %d, want 5", pos)
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
	if pos := m.CursorPosition(); pos != 1 {
		t.Fatalf("after g: CursorPosition = %d, want 1", pos)
	}
}

func TestCursorPosition_AfterFilter(t *testing.T) {
	m := defaultListModel(10).SetEntries(makeEntries(5))
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	// Filter to indices 0,1 → cursor clamps
	m = m.SetFilter([]int{0, 1})
	if pos := m.CursorPosition(); pos != 2 {
		t.Fatalf("after filter: CursorPosition = %d, want 2", pos)
	}
}

// T-079: R1.8 — cursor row rendered with CursorHighlight background.
func TestView_CursorHighlight(t *testing.T) {
	m := defaultListModel(10).SetEntries(makeEntries(3))
	v := m.View()
	lines := strings.Split(v, "\n")
	if len(lines) < 1 {
		t.Fatal("expected at least 1 line")
	}
	// Cursor is at row 0. That row should have ANSI background styling.
	if !strings.Contains(lines[0], "\x1b[") {
		t.Errorf("cursor row should have ANSI styling: %q", lines[0])
	}
	// Non-cursor rows should not have CursorHighlight background.
	// (They may have other styling, but not the same background.)
	if len(lines) > 1 && lines[0] == lines[1] {
		t.Error("cursor row should differ from non-cursor row")
	}

	// Verify background escape code is present for cursor row.
	// "48;" is the ANSI SGR background prefix.
	if !strings.Contains(lines[0], "48;") {
		t.Logf("cursor row (line 0): %q", lines[0])
		t.Logf("non-cursor row (line 1): %q", lines[1])
		t.Errorf("cursor row should contain background color escape (48;)")
	}
}

// WindowSizeMsg must be processed even when the entry list is empty.
// The initial resize arrives before async loading finishes.
func TestWindowSizeMsg_ProcessedWhenEmpty(t *testing.T) {
	m := defaultListModel(10) // no entries
	m, _ = m.Update(tea.WindowSizeMsg{Width: 200, Height: 50})
	if m.width != 200 {
		t.Errorf("width = %d after WindowSizeMsg on empty list, want 200", m.width)
	}
	if m.scroll.ViewportHeight != 50 {
		t.Errorf("ViewportHeight = %d after WindowSizeMsg on empty list, want 50", m.scroll.ViewportHeight)
	}
}

// Messages with embedded newlines must render as exactly one terminal line.
func TestFlattenNewlines(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"no newlines", "no newlines"},
		{"line1\nline2", "line1 line2"},
		{"line1\r\nline2", "line1 line2"},
		{"tabs\t\there", "tabs\t\there"},
		{"mixed\n\t\tindent", "mixed indent"},
		{"trailing\n", "trailing "},
		{"\nleading", " leading"},
		{"multi\n\n\nnewlines", "multi newlines"},
	}
	for _, tt := range tests {
		got := flattenNewlines(tt.in)
		if got != tt.want {
			t.Errorf("flattenNewlines(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

// An entry with embedded newlines in its message renders as one line in the list.
func TestRenderCompactRow_FlattenNewlines(t *testing.T) {
	entry := logsource.Entry{
		IsJSON: true,
		Time:   time.Now(),
		Level:  "INFO",
		Logger: "test",
		Msg:    "line1\n\tline2\n\tline3",
		Raw:    []byte(`{}`),
	}
	row := RenderCompactRow(entry, 120, theme.GetTheme("tokyo-night"), config.DefaultConfig())
	if strings.Contains(row, "\n") {
		t.Errorf("compact row contains newline: %q", row)
	}
}
