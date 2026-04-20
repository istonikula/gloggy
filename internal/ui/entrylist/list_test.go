package entrylist

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

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

// Regression for F-202 (color-profile-collapse, Tier 24): a previous
// T-180 sign-off concluded themes were perceptually distinct based on
// tui-mcp screenshots. That conclusion was wrong — the tui-mcp PTY
// environment downgraded termenv's color profile (stdout not detected
// as a true TTY), which caused lipgloss to downsample every theme's
// TrueColor BaseBg hex to the same xterm-256 palette slot (`48;5;232m`).
// All three themes rendered identical dark bg in the screenshots.
//
// This test runs under the test-package `init()` that forces
// `termenv.TrueColor`, so we're asserting the RENDERER-LEVEL outcome:
// when the color profile IS TrueColor, do the three bundled themes
// produce DISTINCT BaseBg SGR in the real `View()` output path?
//
// Why this test exists: the tui-mcp visual verification is unreliable
// for fine-grained bg distinction because the harness environment is
// not guaranteed to have TrueColor. Users in real terminals with
// COLORTERM=truecolor should see the distinctness (and `main.go`'s
// `forceTrueColorIfSupported()` guards that path). This test is the
// objective floor: in a TrueColor profile, themes MUST NOT collapse.
//
// cavekit-config.md R4 AC 19 (cross-theme perceptibly-distinct).
func TestView_BaseBg_ThemesProduceDistinctSGR_UnderTrueColor(t *testing.T) {
	names := []string{"tokyo-night", "catppuccin-mocha", "material-dark"}
	seen := map[string]string{} // baseBgSGR -> theme name

	for _, name := range names {
		th := theme.GetTheme(name)
		m := NewListModel(th, config.DefaultConfig(), 60, 8).SetEntries(makeEntries(3))
		m.Focused = true
		out := m.View()

		baseBgSGR := bgColorANSI(th.BaseBg)
		if baseBgSGR == "" {
			t.Fatalf("%s: empty BaseBg SGR probe — is TrueColor forced in init()?", name)
		}
		if !strings.Contains(out, baseBgSGR) {
			t.Errorf("%s: View() output missing BaseBg SGR %q", name, baseBgSGR)
		}
		if prev, clash := seen[baseBgSGR]; clash {
			t.Errorf("%s and %s collapse to the same BaseBg SGR %q — profile downsampling likely active",
				prev, name, baseBgSGR)
		}
		seen[baseBgSGR] = name
	}
}

// Regression for F-203 (two-tone row bug, Tier 24 follow-up): the default
// (non-cursor, non-match) row path in `RenderCompactRow` styles the level
// segment with `lipgloss.NewStyle().Foreground(levelColor)` and nothing
// else. That emits a full `\x1b[0m` reset after the level — which, when
// the outer `PaneStyle` paints `BaseBg`, punches a hole through the pane's
// bg: TIME+LEVEL render on BaseBg, LOGGER+MSG fall back to terminal default.
//
// `applyPaneStyle` now calls `appshell.RepaintBg` to re-assert `BaseBg`
// after every inner reset. This test pins the user-observed invariant
// directly: the `\x1b[0m` emitted after the level token must be
// immediately followed by the BaseBg reassert sequence so logger+message
// inherit BaseBg and not terminal default.
func TestView_LevelTokenReset_FollowedByBaseBg(t *testing.T) {
	th := theme.GetTheme("tokyo-night")
	m := NewListModel(th, config.DefaultConfig(), 60, 8).SetEntries(makeEntries(3))
	m.Focused = true
	out := m.View()

	baseBgOpen := bgColorANSI(th.BaseBg)
	if baseBgOpen == "" {
		t.Fatal("empty BaseBg SGR probe — is TrueColor forced in init()?")
	}

	// `RenderCompactRow` emits the level as `…\x1b[38;2;lvlFgm<pad>INFO <pad>\x1b[0m…`.
	// Find the INFO token and scan forward for the very next `\x1b[0m` — that
	// is the level segment's close. It must be followed by the BaseBg open.
	idx := strings.Index(out, "INFO")
	if idx < 0 {
		t.Fatal("INFO token missing from view output — test premise invalid")
	}
	tail := out[idx:]
	reset := strings.Index(tail, "\x1b[0m")
	if reset < 0 {
		t.Fatal("no reset after INFO — test premise invalid")
	}
	after := tail[reset+len("\x1b[0m"):]
	if !strings.HasPrefix(after, baseBgOpen) {
		previewEnd := 60
		if previewEnd > len(after) {
			previewEnd = len(after)
		}
		t.Fatalf("reset after INFO not followed by BaseBg reassert %q\nsaw: %q",
			baseBgOpen, after[:previewEnd])
	}
}

// bgColorANSI mirrors the appshell test helper — renders a probe with
// background color c and returns the SGR prefix. Local copy so this test
// doesn't cross package boundaries.
func bgColorANSI(c lipgloss.Color) string {
	rendered := lipgloss.NewStyle().Background(c).Render("x")
	end := strings.Index(rendered, "x")
	if end <= 0 {
		return ""
	}
	return rendered[:end]
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

// TestListModel_T163_RowForY_Rejects_When_ContentTopY_Unset pins the
// T-163 (F-127) regression vector: a ListModel that has never had
// `WithContentTopY` called on it must reject all rowForY lookups. Without
// the contentTopYSet guard, `contentTopY` defaults to 0 and rowForY
// silently resolves any terminal-Y in [0, viewportHeight) to a visible
// row — reintroducing the pre-T-158 2-row offset bug if the app-shell
// wiring is ever dropped.
func TestListModel_T163_RowForY_Rejects_When_ContentTopY_Unset(t *testing.T) {
	// Fresh list, never wires WithContentTopY. WindowSizeMsg still arrives
	// so ViewportHeight is set to a normal value — proving the rejection
	// is due to contentTopYSet, not a degenerate viewport.
	m := defaultListModel(20).SetEntries(makeEntries(30))
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 22})

	for _, y := range []int{0, 2, 5, 10, 15} {
		if got := m.rowForY(y); got != -1 {
			t.Errorf("rowForY(%d) without WithContentTopY = %d, want -1 (rejected)", y, got)
		}
	}

	// After injecting contentTopY the resolver comes back online.
	m = m.WithContentTopY(2)
	if got := m.rowForY(2); got != 0 {
		t.Errorf("after WithContentTopY(2), rowForY(2) = %d, want 0", got)
	}
	if got := m.rowForY(5); got != 3 {
		t.Errorf("after WithContentTopY(2), rowForY(5) = %d, want 3", got)
	}
}

// R14 AC1+AC2: cursor on last entry => AppendEntries advances cursor to new
// last and scrolls viewport (scrolloff at bottom edge applies).
func TestAppendEntries_TailFollow_AdvancesCursorAndViewport(t *testing.T) {
	const height = 10
	m := defaultListModel(height).SetEntries(makeEntries(20))
	// Land cursor on the last entry.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	if m.scroll.Cursor != 19 {
		t.Fatalf("setup: cursor=%d, want 19", m.scroll.Cursor)
	}

	before := m.scroll.Offset
	m = m.AppendEntries(makeEntries(5))

	if m.scroll.Cursor != 24 {
		t.Errorf("cursor after follow-append = %d, want 24 (new last)", m.scroll.Cursor)
	}
	// The last entry must be inside the viewport: Offset <= 24 < Offset+Height.
	if m.scroll.Offset > 24 || 24 >= m.scroll.Offset+height {
		t.Errorf("offset=%d with height=%d does not include cursor 24", m.scroll.Offset, height)
	}
	if m.scroll.Offset == before {
		t.Errorf("offset did not advance (before=%d, after=%d)", before, m.scroll.Offset)
	}
}

// R14 AC3+AC4: cursor not on last entry => AppendEntries leaves cursor and
// offset untouched.
func TestAppendEntries_NotAtTail_NoFollow(t *testing.T) {
	const height = 10
	m := defaultListModel(height).SetEntries(makeEntries(20))
	// Move cursor into the middle via j presses (3 times), which will sit at
	// cursor=3 with offset 0. (Cursor is definitely not on the last entry.)
	for i := 0; i < 3; i++ {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	}
	wantCursor := m.scroll.Cursor
	wantOffset := m.scroll.Offset
	if wantCursor >= m.scroll.TotalEntries-1 {
		t.Fatalf("setup: cursor at tail; expected mid-list; cursor=%d total=%d", wantCursor, m.scroll.TotalEntries)
	}

	m = m.AppendEntries(makeEntries(5))

	if m.scroll.Cursor != wantCursor {
		t.Errorf("cursor moved: was %d, now %d (must not change when not at tail)", wantCursor, m.scroll.Cursor)
	}
	if m.scroll.Offset != wantOffset {
		t.Errorf("offset moved: was %d, now %d (must not change when not at tail)", wantOffset, m.scroll.Offset)
	}
}

// R14 AC5: empty list => first append leaves Cursor=0, Offset=0.
func TestAppendEntries_EmptyList_FirstAppend(t *testing.T) {
	m := defaultListModel(10)
	if m.scroll.TotalEntries != 0 {
		t.Fatalf("setup: total=%d, want 0", m.scroll.TotalEntries)
	}
	m = m.AppendEntries(makeEntries(5))
	if m.scroll.Cursor != 0 {
		t.Errorf("cursor after empty-first-append = %d, want 0", m.scroll.Cursor)
	}
	if m.scroll.Offset != 0 {
		t.Errorf("offset after empty-first-append = %d, want 0", m.scroll.Offset)
	}
}

// R14 AC6: IsAtTail reflects cursor==last; k breaks it, G restores it.
func TestIsAtTail_ReflectsCursorAtLast(t *testing.T) {
	m := defaultListModel(10).SetEntries(makeEntries(20))
	if m.IsAtTail() {
		t.Error("IsAtTail on fresh list (cursor=0) should be false")
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	if !m.IsAtTail() {
		t.Error("IsAtTail after G should be true")
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	if m.IsAtTail() {
		t.Error("IsAtTail after k should be false")
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	if !m.IsAtTail() {
		t.Error("IsAtTail after return-to-tail (G) should be true again")
	}
}

// R14 AC6 edge: empty list is not "at tail".
func TestIsAtTail_EmptyIsFalse(t *testing.T) {
	m := defaultListModel(10)
	if m.IsAtTail() {
		t.Error("IsAtTail on empty list must be false")
	}
}

// R14 combined: after k, subsequent appends must NOT advance viewport even if
// the user had been following. Restoring tail via G must re-engage.
func TestAppendEntries_PauseWithK_ResumeWithG(t *testing.T) {
	const height = 10
	m := defaultListModel(height).SetEntries(makeEntries(20))
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	m = m.AppendEntries(makeEntries(3)) // following: advances
	if m.scroll.Cursor != 22 {
		t.Fatalf("setup: cursor=%d, want 22", m.scroll.Cursor)
	}

	// User presses k to pause.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	pausedCursor := m.scroll.Cursor
	pausedOffset := m.scroll.Offset

	// Appends arrive while paused — no follow.
	m = m.AppendEntries(makeEntries(3))
	if m.scroll.Cursor != pausedCursor {
		t.Errorf("paused: cursor moved %d -> %d", pausedCursor, m.scroll.Cursor)
	}
	if m.scroll.Offset != pausedOffset {
		t.Errorf("paused: offset moved %d -> %d", pausedOffset, m.scroll.Offset)
	}

	// User presses G to resume.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	m = m.AppendEntries(makeEntries(3))
	wantLast := 20 + 3 + 3 + 3 - 1
	if m.scroll.Cursor != wantLast {
		t.Errorf("resumed: cursor=%d, want %d", m.scroll.Cursor, wantLast)
	}
}
