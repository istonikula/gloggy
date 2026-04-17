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
// T-100: list output is wrapped in a pane border; the cursor row is now the
// first content line inside the top border.
func TestView_CursorHighlight(t *testing.T) {
	m := defaultListModel(10).SetEntries(makeEntries(3))
	m.Focused = true
	v := m.View()
	lines := strings.Split(v, "\n")
	if len(lines) < 3 {
		t.Fatalf("expected at least 3 lines (top border + content): got %d", len(lines))
	}
	// First line is top border; cursor row is the next content line.
	cursorRow := lines[1]
	if !strings.Contains(cursorRow, "\x1b[") {
		t.Errorf("cursor row should have ANSI styling: %q", cursorRow)
	}
	// Non-cursor rows should not match the cursor row.
	if len(lines) > 2 && cursorRow == lines[2] {
		t.Error("cursor row should differ from non-cursor row")
	}
	// "48;" is the ANSI SGR background prefix (CursorHighlight).
	if !strings.Contains(cursorRow, "48;") {
		t.Logf("cursor row (line 1): %q", cursorRow)
		if len(lines) > 2 {
			t.Logf("non-cursor row (line 2): %q", lines[2])
		}
		t.Errorf("cursor row should contain background color escape (48;)")
	}
}

// WindowSizeMsg must be processed even when the entry list is empty.
// The initial resize arrives before async loading finishes.
//
// T-100: incoming width/height are the OUTER pane allocation; the list
// reserves 2 cells / 2 rows for its full border.
func TestWindowSizeMsg_ProcessedWhenEmpty(t *testing.T) {
	m := defaultListModel(10) // no entries
	m, _ = m.Update(tea.WindowSizeMsg{Width: 200, Height: 50})
	if m.width != 198 {
		t.Errorf("width = %d after WindowSizeMsg on empty list, want 198 (200 - 2 borders)", m.width)
	}
	if m.scroll.ViewportHeight != 48 {
		t.Errorf("ViewportHeight = %d after WindowSizeMsg on empty list, want 48 (50 - 2 borders)", m.scroll.ViewportHeight)
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

// T-100: list rendered with focus elsewhere uses DividerColor borders +
// UnfocusedBg background. Outer dimensions match the focused render so the
// focus toggle does not reflow the layout.
func TestView_VisualState_Unfocused_DiffersFromFocused_AndHasBg(t *testing.T) {
	m := defaultListModel(10).SetEntries(makeEntries(3))

	m.Focused = true
	focused := m.View()

	m.Focused = false
	unfocused := m.View()

	// Outer dimensions must match between focused and unfocused.
	fLines := strings.Split(focused, "\n")
	uLines := strings.Split(unfocused, "\n")
	if len(fLines) != len(uLines) {
		t.Errorf("row count mismatch: focused=%d unfocused=%d", len(fLines), len(uLines))
	}

	if focused == unfocused {
		t.Errorf("focused and unfocused output must differ (border color + bg)")
	}

	// Unfocused render must include a background SGR (48;) for UnfocusedBg.
	if !strings.Contains(unfocused, "48;") {
		t.Errorf("unfocused render missing background SGR (48;): %q", unfocused)
	}
	// Focused render must NOT include a background SGR on the border.
	// (Cursor row uses 48; for CursorHighlight; the border lines should
	// not — a structural test would inspect line[0] only.)
	if strings.Contains(fLines[0], "48;") {
		t.Errorf("focused border line must not have UnfocusedBg: %q", fLines[0])
	}
}

// T-101: when the pane is alone (no other pane visible), it uses the
// focused styling regardless of the Focused flag.
func TestView_VisualState_Alone_UsesFocusedTreatment(t *testing.T) {
	m := defaultListModel(10).SetEntries(makeEntries(3))

	m.Focused = false
	m.Alone = false
	unfocused := m.View()

	m.Alone = true
	alone := m.View()

	// Alone output must match the focused treatment, not the unfocused one.
	if alone == unfocused {
		t.Errorf("alone pane must differ from unfocused (alone uses focused treatment)")
	}

	m.Alone = false
	m.Focused = true
	focused := m.View()
	if alone != focused {
		t.Errorf("alone treatment must equal focused treatment\nalone:    %q\nfocused:  %q", alone, focused)
	}
}

// T-102: the cursor row is rendered with CursorHighlight regardless of
// focus state. When unfocused, the surrounding pane is Faint-dimmed but
// the cursor row remains identifiable.
func TestView_CursorRow_AlwaysRendered_WhenUnfocused(t *testing.T) {
	m := defaultListModel(10).SetEntries(makeEntries(3))
	m.Focused = false
	v := m.View()
	lines := strings.Split(v, "\n")
	if len(lines) < 2 {
		t.Fatalf("expected at least 2 lines: got %d", len(lines))
	}
	// First content line (after top border) is the cursor row at index 0.
	cursorRow := lines[1]
	if !strings.Contains(cursorRow, "48;") {
		t.Errorf("cursor row must carry CursorHighlight bg even when unfocused: %q", cursorRow)
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
