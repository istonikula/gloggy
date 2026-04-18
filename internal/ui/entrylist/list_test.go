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

// T-111 R8 #6 / R9 #5 — wrap indicator (↻) renders on cursor row when a
// level-jump or mark navigation wraps around.
func TestListModel_View_RendersWrapIndicator(t *testing.T) {
	// Build entries: 5 INFOs then 1 ERROR at the end. With cursor at 5 (the
	// ERROR), pressing `e` advances forward and wraps back to the same
	// ERROR — WrapForward direction.
	entries := makeEntries(6)
	entries[5].Level = "ERROR"
	m := defaultListModel(10).SetEntries(entries)
	m.Focused = true
	// Move cursor to the ERROR row (last entry).
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	// Press `e` — there is only one ERROR (the current one), so NextLevel
	// finds it again and reports WrapForward.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	if m.WrapDir() == NoWrap {
		t.Fatalf("expected wrap after pressing e past last ERROR; got NoWrap")
	}
	if !m.HasTransient() {
		t.Fatalf("HasTransient must be true while wrap indicator is active")
	}
	v := m.View()
	if !strings.Contains(v, "↻") {
		t.Errorf("View() must render the wrap indicator glyph ↻; got %q", v)
	}
}

// T-111 — pressing Esc (ClearTransient) hides the wrap indicator on next View.
func TestListModel_View_NoIndicator_AfterClearTransient(t *testing.T) {
	entries := makeEntries(6)
	entries[5].Level = "ERROR"
	m := defaultListModel(10).SetEntries(entries)
	m.Focused = true
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	if m.WrapDir() == NoWrap {
		t.Fatalf("precondition: expected wrap state before ClearTransient")
	}
	m = m.ClearTransient()
	if m.HasTransient() {
		t.Errorf("ClearTransient must drop wrap state")
	}
	if strings.Contains(m.View(), "↻") {
		t.Errorf("wrap glyph must not render after ClearTransient")
	}
}

// T-112 R8 #7-8 — when level-jump lands on an entry excluded by the active
// filter, the entry is pinned into the visible list with the ⌀ glyph.
func TestListModel_LevelJump_LandsOnFilteredOutEntry_RendersIndicator(t *testing.T) {
	// 5 INFOs and 1 ERROR (at index 5).
	entries := makeEntries(6)
	entries[5].Level = "ERROR"
	m := defaultListModel(10).SetEntries(entries)
	m.Focused = true
	// Apply filter that hides the ERROR (only INFOs pass).
	m = m.SetFilter([]int{0, 1, 2, 3, 4})
	if got := len(m.visibleEntries()); got != 5 {
		t.Fatalf("precondition: filtered visible len = %d, want 5", got)
	}
	// Press `e` — must navigate to the filtered-out ERROR and pin it.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	if m.PinnedFullIdx() != 5 {
		t.Fatalf("PinnedFullIdx after `e` to filtered-out ERROR = %d, want 5", m.PinnedFullIdx())
	}
	vis := m.visibleEntries()
	if len(vis) != 6 {
		t.Errorf("visible entries after pin = %d, want 6 (5 filtered + 1 pinned)", len(vis))
	}
	v := m.View()
	if !strings.Contains(v, "⌀") {
		t.Errorf("View() must render outside-filter indicator ⌀ when level-jump pins; got %q", v)
	}
	// Subsequent j-nav clears the pin.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if m.PinnedFullIdx() != -1 {
		t.Errorf("PinnedFullIdx after j = %d, want -1 (pin cleared)", m.PinnedFullIdx())
	}
	if strings.Contains(m.View(), "⌀") {
		t.Errorf("⌀ glyph must not render after pin is cleared")
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

// ---------- T-158: click-row resolver uses contentTopY injection ----------

// TestListModel_MouseClick_HonorsContentTopY_Offset2 verifies that with
// contentTopY = 2 (header 1 + top border 1), a click at terminal y=2
// lands on visible row 0 — NOT row 2 as the pre-T-158 resolver did.
func TestListModel_MouseClick_HonorsContentTopY_Offset2(t *testing.T) {
	m := defaultListModel(20).SetEntries(makeEntries(30)).WithContentTopY(2)
	// Deliver a sized window so ViewportHeight is meaningful.
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 22})

	m, _ = m.Update(tea.MouseMsg{X: 10, Y: 2, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	if got := m.Cursor(); got != 0 {
		t.Errorf("click y=2 with contentTopY=2: Cursor = %d, want 0", got)
	}

	m, _ = m.Update(tea.MouseMsg{X: 10, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	if got := m.Cursor(); got != 3 {
		t.Errorf("click y=5 with contentTopY=2: Cursor = %d, want 3", got)
	}
}

// TestListModel_MouseClick_TopBorderBelowContentTopY_NoOp: click at y=1
// (above the first content row) is a no-op — rowForY returns -1 and the
// Press handler does not touch the cursor.
func TestListModel_MouseClick_TopBorderBelowContentTopY_NoOp(t *testing.T) {
	m := defaultListModel(20).SetEntries(makeEntries(30)).WithContentTopY(2)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 22})
	// Move cursor off row 0 so a bug would be observable.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	before := m.Cursor()

	m, _ = m.Update(tea.MouseMsg{X: 10, Y: 1, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})

	if got := m.Cursor(); got != before {
		t.Errorf("click on top border (y=1): Cursor = %d, want %d (unchanged)", got, before)
	}
}
