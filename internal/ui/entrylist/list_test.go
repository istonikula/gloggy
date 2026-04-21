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

// T-080: R11 — CursorPosition returns 1-based index, tracks j/k/g/G + filter.
func TestCursorPosition(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		m := defaultListModel(10)
		assert.Equal(t, 0, m.CursorPosition())
	})

	t.Run("initial + j + k", func(t *testing.T) {
		m := defaultListModel(10).SetEntries(makeEntries(5))
		assert.Equal(t, 1, m.CursorPosition(), "initial")
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
		assert.Equal(t, 2, m.CursorPosition(), "after j")
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
		assert.Equal(t, 1, m.CursorPosition(), "after k")
	})

	t.Run("G then g", func(t *testing.T) {
		m := defaultListModel(10).SetEntries(makeEntries(5))
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
		assert.Equal(t, 5, m.CursorPosition(), "after G")
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
		assert.Equal(t, 1, m.CursorPosition(), "after g")
	})

	t.Run("clamps after filter", func(t *testing.T) {
		m := defaultListModel(10).SetEntries(makeEntries(5))
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
		m = m.SetFilter([]int{0, 1})
		assert.Equal(t, 2, m.CursorPosition(), "after filter")
	})
}

// T-079: R1.8 + T-102 — cursor row carries CursorHighlight bg regardless of
// focus state. List output is wrapped in a pane border; cursor row is the
// first content line inside the top border (lines[1]).
func TestView_CursorRow_HasHighlightBg(t *testing.T) {
	for _, focused := range []bool{true, false} {
		name := "focused"
		if !focused {
			name = "unfocused"
		}
		t.Run(name, func(t *testing.T) {
			m := defaultListModel(10).SetEntries(makeEntries(3))
			m.Focused = focused
			lines := strings.Split(m.View(), "\n")
			require.GreaterOrEqualf(t, len(lines), 3, "expected at least 3 lines, got %d", len(lines))
			cursorRow := lines[1]
			// "48;" is the ANSI SGR background prefix (CursorHighlight).
			assert.Containsf(t, cursorRow, "48;",
				"cursor row missing bg escape (48;): %q", cursorRow)
			if focused && len(lines) > 2 {
				assert.NotEqual(t, lines[2], cursorRow, "cursor row should differ from non-cursor row")
			}
		})
	}
}

// WindowSizeMsg must be processed even when the entry list is empty.
// T-100: incoming width/height are the OUTER pane allocation; the list
// reserves 2 cells / 2 rows for its full border.
func TestWindowSizeMsg_ProcessedWhenEmpty(t *testing.T) {
	m := defaultListModel(10) // no entries
	m, _ = m.Update(tea.WindowSizeMsg{Width: 200, Height: 50})
	assert.Equal(t, 198, m.width, "width after WindowSizeMsg on empty list (200 - 2 borders)")
	assert.Equal(t, 48, m.scroll.ViewportHeight, "ViewportHeight after WindowSizeMsg on empty list (50 - 2 borders)")
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

// T-111 R8 #6 / R9 #5 — wrap indicator (↻) renders on cursor row when a
// level-jump or mark navigation wraps around; Esc (ClearTransient) hides it.
func TestListModel_View_WrapIndicator(t *testing.T) {
	// Build entries: 5 INFOs then 1 ERROR at the end. With cursor at 5 (the
	// ERROR), pressing `e` advances forward and wraps back to the same
	// ERROR — WrapForward direction.
	setup := func() ListModel {
		entries := makeEntries(6)
		entries[5].Level = "ERROR"
		m := defaultListModel(10).SetEntries(entries)
		m.Focused = true
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
		return m
	}

	t.Run("renders ↻ after wrap", func(t *testing.T) {
		m := setup()
		require.NotEqual(t, NoWrap, m.WrapDir(), "expected wrap after pressing e past last ERROR")
		require.True(t, m.HasTransient(), "HasTransient must be true while wrap indicator is active")
		assert.Contains(t, m.View(), "↻", "View() must render the wrap indicator glyph ↻")
	})

	t.Run("ClearTransient drops ↻", func(t *testing.T) {
		m := setup()
		require.NotEqual(t, NoWrap, m.WrapDir(), "precondition: expected wrap state before ClearTransient")
		m = m.ClearTransient()
		assert.False(t, m.HasTransient(), "ClearTransient must drop wrap state")
		assert.NotContains(t, m.View(), "↻", "wrap glyph must not render after ClearTransient")
	})
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

// T-158/T-163 (F-127): click-row resolver uses contentTopY injection. A
// ListModel without `WithContentTopY` called must reject all rowForY lookups
// (V9: silent 2-row-offset regression); once injected, y=contentTopY maps to
// row 0; y=contentTopY+k maps to row k; y<contentTopY is a no-op (top border).
func TestListModel_MouseClick_ContentTopY(t *testing.T) {
	t.Run("rowForY rejects when contentTopY unset", func(t *testing.T) {
		// Fresh list, never wires WithContentTopY. WindowSizeMsg still arrives
		// so ViewportHeight is set — proving rejection is due to contentTopYSet,
		// not a degenerate viewport.
		m := defaultListModel(20).SetEntries(makeEntries(30))
		m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 22})
		for _, y := range []int{0, 2, 5, 10, 15} {
			assert.Equalf(t, -1, m.rowForY(y), "rowForY(%d) without WithContentTopY should be rejected", y)
		}
		// After injection the resolver comes back online.
		m = m.WithContentTopY(2)
		assert.Equal(t, 0, m.rowForY(2), "after WithContentTopY(2), rowForY(2)")
		assert.Equal(t, 3, m.rowForY(5), "after WithContentTopY(2), rowForY(5)")
	})

	t.Run("click honours contentTopY offset", func(t *testing.T) {
		m := defaultListModel(20).SetEntries(makeEntries(30)).WithContentTopY(2)
		m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 22})
		m, _ = m.Update(tea.MouseMsg{X: 10, Y: 2, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
		assert.Equal(t, 0, m.Cursor(), "click y=2 with contentTopY=2: Cursor")
		m, _ = m.Update(tea.MouseMsg{X: 10, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
		assert.Equal(t, 3, m.Cursor(), "click y=5 with contentTopY=2: Cursor")
	})

	t.Run("click above contentTopY is a no-op", func(t *testing.T) {
		m := defaultListModel(20).SetEntries(makeEntries(30)).WithContentTopY(2)
		m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 22})
		for i := 0; i < 3; i++ {
			m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		}
		before := m.Cursor()
		m, _ = m.Update(tea.MouseMsg{X: 10, Y: 1, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
		assert.Equalf(t, before, m.Cursor(), "click on top border (y=1): Cursor should be unchanged (%d)", before)
	})
}

