package entrylist

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
	assert.LessOrEqualf(t, count, height,
		"rendered %d rows for 100k entries, max expected %d (viewport height)", count, height)
	assert.Greater(t, count, 0, "rendered 0 rows — expected at least some")
}

// T-029: R6.2 — render time < 16ms for large dataset
func TestListModel_RenderSpeed_100k(t *testing.T) {
	const total = 100_000
	const height = 40
	m := defaultListModel(height).SetEntries(makeEntries(total))

	start := time.Now()
	_ = m.View()
	elapsed := time.Since(start)

	assert.Lessf(t, elapsed, 16*time.Millisecond, "View() took %v, expected < 16ms", elapsed)
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
	assert.LessOrEqualf(t, count, height, "rendered %d rows when scrolled, max expected %d", count, height)
}

// No SelectionMsg when cursor doesn't move (already at boundary).
func TestListModel_NoSelectionMsg_AtBoundary(t *testing.T) {
	entries := makeEntries(5)
	m := defaultListModel(10).SetEntries(entries)

	// Already at top — k should not emit.
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	assert.Nil(t, cmd, "expected nil cmd when cursor cannot move")
}

// T-080: R11 — CursorPosition returns 1-based index
func TestCursorPosition_EmptyList(t *testing.T) {
	m := defaultListModel(10)
	require.Equal(t, 0, m.CursorPosition(), "CursorPosition on empty")
}

func TestCursorPosition_JK(t *testing.T) {
	m := defaultListModel(10).SetEntries(makeEntries(5))
	require.Equal(t, 1, m.CursorPosition(), "initial CursorPosition")
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	require.Equal(t, 2, m.CursorPosition(), "after j: CursorPosition")
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	require.Equal(t, 1, m.CursorPosition(), "after k: CursorPosition")
}

func TestCursorPosition_gG(t *testing.T) {
	m := defaultListModel(10).SetEntries(makeEntries(5))
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	require.Equal(t, 5, m.CursorPosition(), "after G: CursorPosition")
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
	require.Equal(t, 1, m.CursorPosition(), "after g: CursorPosition")
}

func TestCursorPosition_AfterFilter(t *testing.T) {
	m := defaultListModel(10).SetEntries(makeEntries(5))
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	// Filter to indices 0,1 → cursor clamps
	m = m.SetFilter([]int{0, 1})
	require.Equal(t, 2, m.CursorPosition(), "after filter: CursorPosition")
}

// T-079: R1.8 — cursor row rendered with CursorHighlight background.
// T-100: list output is wrapped in a pane border; the cursor row is now the
// first content line inside the top border.
func TestView_CursorHighlight(t *testing.T) {
	m := defaultListModel(10).SetEntries(makeEntries(3))
	m.Focused = true
	v := m.View()
	lines := strings.Split(v, "\n")
	require.GreaterOrEqualf(t, len(lines), 3,
		"expected at least 3 lines (top border + content): got %d", len(lines))
	// First line is top border; cursor row is the next content line.
	cursorRow := lines[1]
	assert.Containsf(t, cursorRow, "\x1b[", "cursor row should have ANSI styling: %q", cursorRow)
	// Non-cursor rows should not match the cursor row.
	if len(lines) > 2 {
		assert.NotEqual(t, lines[2], cursorRow, "cursor row should differ from non-cursor row")
	}
	// "48;" is the ANSI SGR background prefix (CursorHighlight).
	if !strings.Contains(cursorRow, "48;") {
		t.Logf("cursor row (line 1): %q", cursorRow)
		if len(lines) > 2 {
			t.Logf("non-cursor row (line 2): %q", lines[2])
		}
		assert.Fail(t, "cursor row should contain background color escape (48;)")
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
	assert.Equal(t, 198, m.width, "width after WindowSizeMsg on empty list (200 - 2 borders)")
	assert.Equal(t, 48, m.scroll.ViewportHeight, "ViewportHeight after WindowSizeMsg on empty list (50 - 2 borders)")
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
		assert.Equalf(t, tt.want, got, "flattenNewlines(%q)", tt.in)
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
	assert.Equalf(t, len(uLines), len(fLines), "row count mismatch: focused=%d unfocused=%d", len(fLines), len(uLines))

	assert.NotEqual(t, unfocused, focused, "focused and unfocused output must differ (border color + bg)")

	// Unfocused render must include a background SGR (48;) for UnfocusedBg.
	assert.Containsf(t, unfocused, "48;", "unfocused render missing background SGR (48;): %q", unfocused)
	// Focused render must NOT include a background SGR on the border.
	// (Cursor row uses 48; for CursorHighlight; the border lines should
	// not — a structural test would inspect line[0] only.)
	assert.NotContainsf(t, fLines[0], "48;", "focused border line must not have UnfocusedBg: %q", fLines[0])
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
		require.NotEmptyf(t, baseBgSGR, "%s: empty BaseBg SGR probe — is TrueColor forced in init()?", name)
		assert.Containsf(t, out, baseBgSGR, "%s: View() output missing BaseBg SGR %q", name, baseBgSGR)
		if prev, clash := seen[baseBgSGR]; clash {
			assert.Failf(t, "theme BaseBg collapse",
				"%s and %s collapse to the same BaseBg SGR %q — profile downsampling likely active",
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
	require.NotEmpty(t, baseBgOpen, "empty BaseBg SGR probe — is TrueColor forced in init()?")

	// `RenderCompactRow` emits the level as `…\x1b[38;2;lvlFgm<pad>INFO <pad>\x1b[0m…`.
	// Find the INFO token and scan forward for the very next `\x1b[0m` — that
	// is the level segment's close. It must be followed by the BaseBg open.
	idx := strings.Index(out, "INFO")
	require.GreaterOrEqual(t, idx, 0, "INFO token missing from view output — test premise invalid")
	tail := out[idx:]
	reset := strings.Index(tail, "\x1b[0m")
	require.GreaterOrEqual(t, reset, 0, "no reset after INFO — test premise invalid")
	after := tail[reset+len("\x1b[0m"):]
	if !strings.HasPrefix(after, baseBgOpen) {
		previewEnd := 60
		if previewEnd > len(after) {
			previewEnd = len(after)
		}
		require.Failf(t, "reset after INFO not followed by BaseBg reassert",
			"want %q\nsaw: %q", baseBgOpen, after[:previewEnd])
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
	assert.NotEqual(t, unfocused, alone, "alone pane must differ from unfocused (alone uses focused treatment)")

	m.Alone = false
	m.Focused = true
	focused := m.View()
	assert.Equalf(t, focused, alone, "alone treatment must equal focused treatment\nalone:    %q\nfocused:  %q", alone, focused)
}

// T-102: the cursor row is rendered with CursorHighlight regardless of
// focus state. When unfocused, the surrounding pane is Faint-dimmed but
// the cursor row remains identifiable.
func TestView_CursorRow_AlwaysRendered_WhenUnfocused(t *testing.T) {
	m := defaultListModel(10).SetEntries(makeEntries(3))
	m.Focused = false
	v := m.View()
	lines := strings.Split(v, "\n")
	require.GreaterOrEqualf(t, len(lines), 2, "expected at least 2 lines: got %d", len(lines))
	// First content line (after top border) is the cursor row at index 0.
	cursorRow := lines[1]
	assert.Containsf(t, cursorRow, "48;", "cursor row must carry CursorHighlight bg even when unfocused: %q", cursorRow)
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
	require.NotEqual(t, NoWrap, m.WrapDir(), "expected wrap after pressing e past last ERROR")
	require.True(t, m.HasTransient(), "HasTransient must be true while wrap indicator is active")
	v := m.View()
	assert.Containsf(t, v, "↻", "View() must render the wrap indicator glyph ↻; got %q", v)
}

// T-111 — pressing Esc (ClearTransient) hides the wrap indicator on next View.
func TestListModel_View_NoIndicator_AfterClearTransient(t *testing.T) {
	entries := makeEntries(6)
	entries[5].Level = "ERROR"
	m := defaultListModel(10).SetEntries(entries)
	m.Focused = true
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	require.NotEqual(t, NoWrap, m.WrapDir(), "precondition: expected wrap state before ClearTransient")
	m = m.ClearTransient()
	assert.False(t, m.HasTransient(), "ClearTransient must drop wrap state")
	assert.NotContains(t, m.View(), "↻", "wrap glyph must not render after ClearTransient")
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
	require.Len(t, m.visibleEntries(), 5, "precondition: filtered visible len")
	// Press `e` — must navigate to the filtered-out ERROR and pin it.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	require.Equal(t, 5, m.PinnedFullIdx(), "PinnedFullIdx after `e` to filtered-out ERROR")
	vis := m.visibleEntries()
	assert.Len(t, vis, 6, "visible entries after pin (5 filtered + 1 pinned)")
	v := m.View()
	assert.Containsf(t, v, "⌀", "View() must render outside-filter indicator ⌀ when level-jump pins; got %q", v)
	// Subsequent j-nav clears the pin.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	assert.Equal(t, -1, m.PinnedFullIdx(), "PinnedFullIdx after j (pin cleared)")
	assert.NotContains(t, m.View(), "⌀", "⌀ glyph must not render after pin is cleared")
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
	assert.NotContainsf(t, row, "\n", "compact row contains newline: %q", row)
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
	assert.Equal(t, 0, m.Cursor(), "click y=2 with contentTopY=2: Cursor")

	m, _ = m.Update(tea.MouseMsg{X: 10, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	assert.Equal(t, 3, m.Cursor(), "click y=5 with contentTopY=2: Cursor")
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

	assert.Equalf(t, before, m.Cursor(), "click on top border (y=1): Cursor should be unchanged (%d)", before)
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
		assert.Equalf(t, -1, m.rowForY(y), "rowForY(%d) without WithContentTopY should be rejected", y)
	}

	// After injecting contentTopY the resolver comes back online.
	m = m.WithContentTopY(2)
	assert.Equal(t, 0, m.rowForY(2), "after WithContentTopY(2), rowForY(2)")
	assert.Equal(t, 3, m.rowForY(5), "after WithContentTopY(2), rowForY(5)")
}

// R14 AC1+AC2: cursor on last entry => AppendEntries advances cursor to new
// last and scrolls viewport (scrolloff at bottom edge applies).
func TestAppendEntries_TailFollow_AdvancesCursorAndViewport(t *testing.T) {
	const height = 10
	m := defaultListModel(height).SetEntries(makeEntries(20))
	// Land cursor on the last entry.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	require.Equal(t, 19, m.scroll.Cursor, "setup: cursor")

	before := m.scroll.Offset
	m = m.AppendEntries(makeEntries(5))

	assert.Equal(t, 24, m.scroll.Cursor, "cursor after follow-append (new last)")
	// The last entry must be inside the viewport: Offset <= 24 < Offset+Height.
	assert.Truef(t, m.scroll.Offset <= 24 && 24 < m.scroll.Offset+height,
		"offset=%d with height=%d does not include cursor 24", m.scroll.Offset, height)
	assert.NotEqualf(t, before, m.scroll.Offset, "offset did not advance (before=%d, after=%d)", before, m.scroll.Offset)
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
	require.Lessf(t, wantCursor, m.scroll.TotalEntries-1,
		"setup: cursor at tail; expected mid-list; cursor=%d total=%d", wantCursor, m.scroll.TotalEntries)

	m = m.AppendEntries(makeEntries(5))

	assert.Equalf(t, wantCursor, m.scroll.Cursor, "cursor moved: was %d, now %d (must not change when not at tail)", wantCursor, m.scroll.Cursor)
	assert.Equalf(t, wantOffset, m.scroll.Offset, "offset moved: was %d, now %d (must not change when not at tail)", wantOffset, m.scroll.Offset)
}

// R14 AC5: empty list => first append leaves Cursor=0, Offset=0.
func TestAppendEntries_EmptyList_FirstAppend(t *testing.T) {
	m := defaultListModel(10)
	require.Equal(t, 0, m.scroll.TotalEntries, "setup: total")
	m = m.AppendEntries(makeEntries(5))
	assert.Equal(t, 0, m.scroll.Cursor, "cursor after empty-first-append")
	assert.Equal(t, 0, m.scroll.Offset, "offset after empty-first-append")
}

// R14 AC6: IsAtTail reflects cursor==last; k breaks it, G restores it.
func TestIsAtTail_ReflectsCursorAtLast(t *testing.T) {
	m := defaultListModel(10).SetEntries(makeEntries(20))
	assert.False(t, m.IsAtTail(), "IsAtTail on fresh list (cursor=0) should be false")
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	assert.True(t, m.IsAtTail(), "IsAtTail after G should be true")
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	assert.False(t, m.IsAtTail(), "IsAtTail after k should be false")
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	assert.True(t, m.IsAtTail(), "IsAtTail after return-to-tail (G) should be true again")
}

// R14 AC6 edge: empty list is not "at tail".
func TestIsAtTail_EmptyIsFalse(t *testing.T) {
	m := defaultListModel(10)
	assert.False(t, m.IsAtTail(), "IsAtTail on empty list must be false")
}

// R14 combined: after k, subsequent appends must NOT advance viewport even if
// the user had been following. Restoring tail via G must re-engage.
func TestAppendEntries_PauseWithK_ResumeWithG(t *testing.T) {
	const height = 10
	m := defaultListModel(height).SetEntries(makeEntries(20))
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	m = m.AppendEntries(makeEntries(3)) // following: advances
	require.Equal(t, 22, m.scroll.Cursor, "setup: cursor")

	// User presses k to pause.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	pausedCursor := m.scroll.Cursor
	pausedOffset := m.scroll.Offset

	// Appends arrive while paused — no follow.
	m = m.AppendEntries(makeEntries(3))
	assert.Equalf(t, pausedCursor, m.scroll.Cursor, "paused: cursor moved %d -> %d", pausedCursor, m.scroll.Cursor)
	assert.Equalf(t, pausedOffset, m.scroll.Offset, "paused: offset moved %d -> %d", pausedOffset, m.scroll.Offset)

	// User presses G to resume.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	m = m.AppendEntries(makeEntries(3))
	wantLast := 20 + 3 + 3 + 3 - 1
	assert.Equal(t, wantLast, m.scroll.Cursor, "resumed: cursor")
}

// T9 (I.keys): `M` clears all marks. Cursor/viewport unchanged, no
// SelectionMsg emitted.
func TestListModel_M_ClearsAllMarks(t *testing.T) {
	m := defaultListModel(10).SetEntries(makeEntries(5))
	// Mark two entries: cursor at 0 → `m`, move to 2 → `m`.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("m")})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("m")})
	require.Equal(t, 2, m.Marks().Count(), "pre-M count")
	cursorBefore := m.scroll.Cursor
	offsetBefore := m.scroll.Offset
	m2, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("M")})
	assert.Equal(t, 0, m2.Marks().Count(), "post-M count")
	assert.Equalf(t, cursorBefore, m2.scroll.Cursor, "M moved cursor: %d → %d", cursorBefore, m2.scroll.Cursor)
	assert.Equalf(t, offsetBefore, m2.scroll.Offset, "M moved offset: %d → %d", offsetBefore, m2.scroll.Offset)
	assert.Nilf(t, cmd, "M emitted cmd %v, want nil (no SelectionMsg on clear-all-marks)", cmd)
}

// T10 (V4, V5, V26): a marked row must not overflow `m.width`. Previously
// `list.View()` rendered `prefix + RenderCompactRow(m.width)` so a 2-cell
// mark glyph pushed the total row to `m.width+2`, which soft-wraps to 2
// terminal lines and cascades into V5 by displacing the header (B2).
// Post-fix: a 2-cell prefix column is reserved on every row; content is
// sized to `m.width-2`; total row = m.width.
func TestListModel_V26_MarkedRow_NoWidthOverflow(t *testing.T) {
	const innerWidth = 60
	const innerHeight = 5
	entries := makeEntries(innerHeight)
	for i := range entries {
		entries[i].Msg = strings.Repeat("x", 200)
	}
	m := NewListModel(theme.GetTheme("tokyo-night"), config.DefaultConfig(), innerWidth, innerHeight)
	m = m.SetEntries(entries)
	// Mark the cursor-row entry so `* ` prefix is prepended.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("m")})

	view := m.View()
	lines := strings.Split(view, "\n")

	// V5: applyPaneStyle wraps the inner ViewportHeight rows with top +
	// bottom border → exactly innerHeight + 2 output lines.
	assert.Lenf(t, lines, innerHeight+2,
		"View line count should be %d (top border + %d inner rows + bottom border)", innerHeight+2, innerHeight)

	// V4 + V26: each line's visible width must fit in innerWidth + 2
	// (inner content + left-/right- border). Pre-fix: marked row = innerWidth+2
	// content + border = innerWidth+4 → fails here.
	maxAllowed := innerWidth + 2
	for i, line := range lines {
		w := lipgloss.Width(line)
		assert.LessOrEqualf(t, w, maxAllowed, "line %d visible width = %d, want ≤ %d: %q", i, w, maxAllowed, line)
	}
}

// T9 (I.keys): `M` with zero marks is a silent no-op — no panic, no state
// change, no command.
func TestListModel_M_EmptyNoop(t *testing.T) {
	m := defaultListModel(10).SetEntries(makeEntries(5))
	require.Equal(t, 0, m.Marks().Count(), "precondition: Count")
	cursorBefore := m.scroll.Cursor
	m2, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("M")})
	assert.Equal(t, 0, m2.Marks().Count(), "post-M count")
	assert.Equalf(t, cursorBefore, m2.scroll.Cursor, "M on empty moved cursor: %d → %d", cursorBefore, m2.scroll.Cursor)
	assert.Nilf(t, cmd, "M on empty emitted cmd %v, want nil", cmd)
}
