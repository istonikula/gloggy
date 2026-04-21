package detailpane

import (
	"regexp"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/istonikula/gloggy/internal/logsource"
	"github.com/istonikula/gloggy/internal/theme"
)

// test helpers (T-107).
func lipglossWidth(s string) int { return lipgloss.Width(s) }
func lipglossStyle(s string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000")).Render(s)
}

func testEntry() logsource.Entry {
	return logsource.Entry{
		IsJSON:     true,
		LineNumber: 1,
		Time:       time.Now(),
		Level:      "INFO",
		Msg:        "hello",
		Raw:        []byte(`{"level":"INFO","msg":"hello"}`),
	}
}

func defaultPane(height int) PaneModel {
	return NewPaneModel(theme.GetTheme("tokyo-night"), height)
}

// paneWithNLines opens a pane at (height, width) pre-populated with n copies
// of "ln" as rawContent. Collapses the hand-rolled setup that otherwise
// repeats across scroll/cursor-highlight/indicator tests.
func paneWithNLines(height, width, n int) PaneModel {
	m := defaultPane(height).SetWidth(width)
	m.open = true
	lines := make([]string, n)
	for i := range lines {
		lines[i] = "ln"
	}
	m.rawContent = strings.Join(lines, "\n")
	m.scroll = NewScrollModel(m.rawContent, m.ContentHeight())
	return m
}

// Regression for F-203 (two-tone detail-pane bug, Tier 24 follow-up):
// `render.go` styles each JSON token (key, string, number, boolean, null)
// with `lipgloss.NewStyle().Foreground(...)` only — no Background. Every
// token ends in `\x1b[0m` which, when the outer `PaneStyle` paints
// `BaseBg`, punches a hole through the pane bg: the screenshot showed
// BaseBg up through the quoted key and terminal-default bg starting at
// the `:` separator.
//
// `PaneModel.View` now calls `appshell.RepaintBg` to re-assert BaseBg
// after every inner reset. This test pins the user-observed invariant
// directly: the `\x1b[0m` emitted after the `"time"` key must be
// immediately followed by the BaseBg reassert sequence so the subsequent
// `": "` separator and value inherit BaseBg.
func TestPaneModel_View_KeyTokenReset_FollowedByBaseBg(t *testing.T) {
	th := theme.GetTheme("tokyo-night")
	m := NewPaneModel(th, 20)
	m.Focused = true
	m = m.SetWidth(60).Open(logsource.Entry{
		IsJSON:     true,
		LineNumber: 1,
		Time:       time.Now(),
		Level:      "INFO",
		Msg:        "hello",
		Raw: []byte(`{"time":"2026-04-14T23:39:10.868Z",` +
			`"level":"INFO","msg":"hello","count":42,"active":true,"data":null}`),
	})
	out := m.View()

	baseBgOpen := bgColorANSI(th.BaseBg)
	require.NotEmpty(t, baseBgOpen, "empty BaseBg SGR probe — is TrueColor forced in init()?")

	// Find the first `"time"` key token — its trailing `\x1b[0m` is the one
	// that historically "punched out" the pane bg before the `": "` separator.
	idx := strings.Index(out, "\"time\"")
	require.GreaterOrEqual(t, idx, 0, "`\"time\"` key missing from view output — test premise invalid")
	tail := out[idx:]
	reset := strings.Index(tail, "\x1b[0m")
	require.GreaterOrEqual(t, reset, 0, "no reset after `\"time\"` — test premise invalid")
	after := tail[reset+len("\x1b[0m"):]
	if !strings.HasPrefix(after, baseBgOpen) {
		previewEnd := 60
		if previewEnd > len(after) {
			previewEnd = len(after)
		}
		require.Failf(t, "reset after `\"time\"` not followed by BaseBg reassert",
			"want %q\nsaw: %q", baseBgOpen, after[:previewEnd])
	}
}

// bgColorANSI is the detailpane-local mirror of the entrylist helper:
// renders a bg probe and returns the opening SGR sequence. Package-local
// so the test doesn't depend on another test package's helpers.
func bgColorANSI(c lipgloss.Color) string {
	rendered := lipgloss.NewStyle().Background(c).Render("x")
	end := strings.Index(rendered, "x")
	if end <= 0 {
		return ""
	}
	return rendered[:end]
}

// T-041: R1.1 — Enter on entry opens detail pane (caller opens via Open(); here we test Open sets state).
func TestPaneModel_Open_SetsOpen(t *testing.T) {
	m := defaultPane(10).Open(testEntry())
	assert.True(t, m.IsOpen(), "expected pane to be open after Open()")
}

// T-041: R1.2 — Double-click handled by ListModel; PaneModel.Open() is the activation path.
// Just verify Open() renders non-empty content.
func TestPaneModel_Open_RendersContent(t *testing.T) {
	m := defaultPane(10).Open(testEntry())
	assert.NotEmpty(t, m.View(), "expected non-empty view after Open()")
}

// T-041: R1.3/R1.4 — Esc and Enter both close the pane and emit BlurredMsg.
func TestPaneModel_DismissKey_ClosesAndEmitsBlurred(t *testing.T) {
	for _, tc := range []struct {
		name    string
		keyType tea.KeyType
	}{
		{"Esc", tea.KeyEsc},
		{"Enter", tea.KeyEnter},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := defaultPane(10).Open(testEntry())
			m2, cmd := m.Update(tea.KeyMsg{Type: tc.keyType})
			assert.False(t, m2.IsOpen(), "expected pane to be closed after %s", tc.name)
			require.NotNil(t, cmd, "expected BlurredMsg cmd")
			msg := cmd()
			_, ok := msg.(BlurredMsg)
			assert.Truef(t, ok, "expected BlurredMsg, got %T", msg)
		})
	}
}

// When pane is closed, Update is a no-op.
func TestPaneModel_Closed_UpdateNoop(t *testing.T) {
	m := defaultPane(10)
	m2, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	assert.False(t, m2.IsOpen(), "should remain closed")
	assert.Nil(t, cmd, "expected nil cmd when pane is closed")
}

// View returns empty string when closed.
func TestPaneModel_Closed_ViewEmpty(t *testing.T) {
	m := defaultPane(10)
	assert.Empty(t, m.View(), "expected empty view when pane is closed")
}

// T-082: R1.5 — open pane View starts with a top border character.
func TestPaneModel_TopBorder(t *testing.T) {
	m := defaultPane(10).Open(testEntry())
	v := m.View()
	require.NotEmpty(t, v, "expected non-empty view")
	// NormalBorder top uses "─" characters.
	assert.Containsf(t, v, "─", "expected top border character '─' in view: %q", v)
}

// T-100: focused vs unfocused panes use the DESIGN.md §4 matrix —
// borders render in BOTH states, only the color differs (FocusBorder vs
// DividerColor). Vertical bar count therefore matches; the discriminator
// is the rendered ANSI color of the border foreground.
func TestPaneModel_Focused_VsUnfocused_DifferentBorderColor(t *testing.T) {
	m := defaultPane(10).Open(testEntry())
	m.Focused = true
	focused := m.View()
	m.Focused = false
	unfocused := m.View()
	assert.Greaterf(t, strings.Count(focused, "│"), 0, "focused pane should render vertical border: %q", focused)
	assert.Greaterf(t, strings.Count(unfocused, "│"), 0, "unfocused pane should render vertical border: %q", unfocused)
	assert.NotEqualf(t, unfocused, focused, "focused and unfocused outputs must differ (border color): %q", focused)
}

// T-107: lipgloss.Width measures cell width correctly across emoji, CJK,
// and ANSI-styled text — verifying our chosen primitive is safe for the
// pane's width-aware code paths.
func TestPaneModel_LipglossWidth_HandlesEmojiCJKAnsi(t *testing.T) {
	for _, tc := range []struct {
		name string
		s    string
		want int
	}{
		{"emoji", "🔥", 2},
		{"cjk", "日本語", 6},
		{"ansi-wrapped ascii", lipglossStyle("X"), 1},
		{"mixed", "a🔥b", 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equalf(t, tc.want, lipglossWidth(tc.s), "lipgloss.Width(%q)", tc.s)
		})
	}
}

// T-107 / T-139 (F-103): pane outer width equals CONTENT width + 2 border
// cells — emoji/CJK content must not push the pane past its budget. With
// single-owner border accounting, `SetWidth(n)` receives CONTENT width and
// the outer rendered block is exactly n + 2 cells wide.
func TestPaneModel_View_OuterWidth_MatchesAllocation(t *testing.T) {
	const contentW = 24
	const outerW = contentW + 2
	entry := logsource.Entry{
		IsJSON: true,
		Time:   time.Now(),
		Level:  "INFO",
		Msg:    "🔥 fire — 日本語 — long enough to overflow naive budgets",
		Raw:    []byte(`{"msg":"🔥 fire 日本語"}`),
	}
	m := defaultPane(8).Open(entry).SetWidth(contentW)
	v := m.View()
	require.NotEmpty(t, v, "expected non-empty view")
	for i, line := range strings.Split(v, "\n") {
		w := lipglossWidth(line)
		if w > outerW {
			assert.Failf(t, "line exceeds outer width",
				"line %d width=%d exceeds outer=%d (content=%d + 2 borders): %q", i, w, outerW, contentW, line)
		}
	}
}

// T-103: the detail pane top border renders in both orientations. The pane
// itself is orientation-agnostic — the layout composes it via either
// JoinVertical (below) or JoinHorizontal (right). The pane's first View()
// line must always be the top border row.
func TestPaneModel_TopBorder_InBothOrientationContexts(t *testing.T) {
	for _, focused := range []bool{true, false} {
		m := defaultPane(10).Open(testEntry())
		m.Focused = focused
		v := m.View()
		lines := strings.Split(v, "\n")
		require.GreaterOrEqualf(t, len(lines), 2, "focused=%v: expected at least 2 lines (top border + content), got %d", focused, len(lines))
		// The first line is the top border. Strip ANSI escapes by
		// scanning for the box-drawing horizontal glyph; lipgloss.Width
		// returns cell width regardless of escape sequences, so a top
		// border line cell-width must equal the rendered output width.
		assert.Truef(t, strings.ContainsRune(lines[0], '─'),
			"focused=%v: first line missing top border glyph '─': %q", focused, lines[0])
	}
}

// T-113 (cavekit-detail-pane R7 / F-003): ContentLines() returns the
// pre-render soft-wrapped content — no border glyphs, no ANSI escapes.
// This is the authoritative match source for in-pane search.
func TestPaneModel_ContentLines_NoBordersNoANSI(t *testing.T) {
	m := defaultPane(10).SetWidth(40).Open(testEntry())
	lines := m.ContentLines()
	require.NotEmpty(t, lines, "expected non-empty ContentLines after Open()")
	borderGlyphs := []string{"│", "─", "╭", "╮", "╰", "╯", "┌", "┐", "└", "┘"}
	for i, line := range lines {
		assert.NotContainsf(t, line, "\x1b[", "line %d contains ANSI escape: %q", i, line)
		for _, g := range borderGlyphs {
			assert.NotContainsf(t, line, g, "line %d contains border glyph %q: %q", i, g, line)
		}
	}
}

// T-113: ContentLines() returns nil when pane is closed or has no content.
func TestPaneModel_ContentLines_ClosedReturnsNil(t *testing.T) {
	m := defaultPane(10)
	assert.Nil(t, m.ContentLines(), "expected nil from closed pane")
	m = m.Open(testEntry()).Close()
	assert.Nil(t, m.ContentLines(), "expected nil from closed-after-open pane")
}

// T-114 (F-002, F-004, F-010): the search prompt row, (cur/total) counter,
// bare-prompt and "No matches" indicator all render via WithSearch() — and
// inactive search leaves the pane's normal rendering unchanged.
func TestPaneModel_WithSearch(t *testing.T) {
	// Entry with multiple 'hello' matches so MatchCount >= 1.
	multiHit := logsource.Entry{
		IsJSON: true,
		Time:   time.Now(),
		Level:  "INFO",
		Msg:    "hello",
		Raw:    []byte(`{"level":"INFO","msg":"hello world","tag":"hello there"}`),
	}
	for _, tc := range []struct {
		name     string
		entry    logsource.Entry
		activate bool
		query    string
		want     []string
		notWant  []string
	}{
		{"prompt+counter on match", multiHit, true, "hello", []string{"/hello", "(1/"}, nil},
		{"no matches indicator", testEntry(), true, "zzz-nope", []string{"No matches"}, nil},
		{"bare prompt", testEntry(), true, "", []string{"/"}, nil},
		{"inactive no prompt", testEntry(), false, "", nil, []string{"No matches"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pane := defaultPane(12).SetWidth(60).Open(tc.entry)
			search := NewSearchModel(theme.GetTheme("tokyo-night"))
			if tc.activate {
				search = search.Activate()
				if tc.query != "" {
					search = search.SetQuery(tc.query, pane.ContentLines())
				}
			}
			view := pane.WithSearch(search).View()
			for _, w := range tc.want {
				assert.Containsf(t, view, w, "view missing %q: %q", w, view)
			}
			for _, nw := range tc.notWant {
				assert.NotContainsf(t, view, nw, "view unexpectedly contained %q: %q", nw, view)
			}
		})
	}
}

// T-115 (F-005): ScrollToLine brings an out-of-window line into view.
// Uses many-line content so the viewport can actually be smaller than
// the line count.
func TestPaneModel_ScrollToLine_BringsMatchIntoView(t *testing.T) {
	// Build an entry whose rendered JSON will have plenty of fields so
	// the wrapped content exceeds the pane height.
	raw := `{"a":"1","b":"2","c":"3","d":"4","e":"5","f":"6","g":"7","h":"8","i":"9","j":"10","target":"find-me"}`
	entry := logsource.Entry{IsJSON: true, Time: time.Now(), Level: "INFO", Msg: "x", Raw: []byte(raw)}
	pane := defaultPane(6).SetWidth(40).Open(entry) // content height ~4
	lines := pane.ContentLines()
	// Locate the line containing "target".
	targetIdx := -1
	for i, l := range lines {
		if strings.Contains(l, "target") {
			targetIdx = i
			break
		}
	}
	require.GreaterOrEqualf(t, targetIdx, 0, "setup: no 'target' in content: %v", lines)
	if targetIdx <= pane.ContentHeight()-1 {
		t.Skipf("target at idx=%d already in initial viewport (content height %d); test not applicable", targetIdx, pane.ContentHeight())
	}
	scrolled := pane.ScrollToLine(targetIdx)
	view := scrolled.View()
	assert.Containsf(t, view, "target", "after ScrollToLine(%d), view should contain 'target': %q", targetIdx, view)
}

// T-134 (F-026, cavekit R11): ScrollToLine moves the cursor AND scrolls so
// the cursor has scrolloff context.
func TestPaneModel_ScrollToLine_MovesCursorWithScrolloffContext(t *testing.T) {
	m := paneWithNLines(12, 40, 100).WithScrolloff(5) // ContentHeight = 10
	m = m.ScrollToLine(40)
	assert.Equal(t, 40, m.scroll.Cursor(), "cursor")
	assert.Equal(t, 36, m.scroll.Offset(), "offset (cursor at bottom-margin)")
}

// T-134: cursor-row render still has CursorHighlight bg when search is
// active — the bg is the last paint in View() so it composes on top of
// SearchHighlight fg.
func TestPaneModel_View_CursorBgOverSearchActive(t *testing.T) {
	th := theme.GetTheme("tokyo-night")
	m := paneWithNLines(12, 40, 100).WithScrolloff(5)
	m.Focused = true
	m = m.ScrollToLine(40)
	view := m.View()
	bgSGR := bgSGRFor(th.CursorHighlight)
	assert.Containsf(t, view, bgSGR, "cursor-row bg missing after ScrollToLine; view=%q", view)
}

// T-125 (F-016): scroll indicator reports percentage when content exceeds
// viewport, 100 at the bottom, and sentinel -1 when the content fits.
func TestPaneModel_ScrollPercent(t *testing.T) {
	for _, tc := range []struct {
		name    string
		nLines  int
		offset  int
		want    int
		comment string
	}{
		{"top of long doc", 200, 0, 10, "offset 0, height 20, total 200 → (0+20)/200 = 10%"},
		{"bottom of long doc", 200, 180, 100, "max offset = 200-20 → 100%"},
		{"fits viewport", 4, 0, -1, "short content → sentinel -1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := paneWithNLines(22, 0, tc.nLines) // width irrelevant for ScrollPercent
			m.scroll.offset = tc.offset
			assert.Equalf(t, tc.want, m.ScrollPercent(), "%s", tc.comment)
		})
	}
}

// T-125: View overlays the indicator on the last body line (dim text).
func TestPaneModel_View_IncludesScrollIndicator(t *testing.T) {
	m := paneWithNLines(22, 30, 200)
	view := m.View()
	assert.Containsf(t, view, "10%", "view should contain \"10%%\" indicator, got: %q", view)
}

// T-125: View omits the indicator when content fits (no "0%" noise).
func TestPaneModel_View_OmitsIndicatorOnShortContent(t *testing.T) {
	m := defaultPane(22).SetWidth(30).Open(testEntry())
	view := m.View()
	assert.NotContainsf(t, view, "%", "short-content view should not render a percentage indicator, got: %q", view)
}

// T-127 (F-020): hidden fields set via WithHiddenFields reach the JSON
// renderer through Open — the suppressed key must not appear in rawContent.
func TestPaneModel_Open_HonorsHiddenFields(t *testing.T) {
	entry := logsource.Entry{
		IsJSON:     true,
		LineNumber: 1,
		Time:       time.Now(),
		Level:      "INFO",
		Msg:        "hello",
		Raw:        []byte(`{"level":"INFO","msg":"hello","secret":"hunter2"}`),
	}
	pane := defaultPane(20).WithHiddenFields([]string{"secret"}).Open(entry)
	assert.NotContainsf(t, pane.rawContent, "secret", "raw content should not include suppressed field `secret`, got: %q", pane.rawContent)
	assert.NotContainsf(t, pane.rawContent, "hunter2", "raw content should not include suppressed value, got: %q", pane.rawContent)
	assert.Containsf(t, pane.rawContent, "hello", "raw content should still include non-suppressed fields, got: %q", pane.rawContent)
}

// T-127 (F-020): Rerender with an updated hiddenFields set re-renders
// the current entry without the newly suppressed field.
func TestPaneModel_Rerender_RemovesNewlyHiddenField(t *testing.T) {
	entry := logsource.Entry{
		IsJSON:     true,
		LineNumber: 1,
		Time:       time.Now(),
		Level:      "INFO",
		Msg:        "hello",
		Raw:        []byte(`{"level":"INFO","msg":"hello","secret":"hunter2"}`),
	}
	pane := defaultPane(20).Open(entry)
	require.Containsf(t, pane.rawContent, "secret", "precondition: raw content should include `secret` before hide, got: %q", pane.rawContent)
	pane = pane.WithHiddenFields([]string{"secret"}).Rerender()
	assert.NotContainsf(t, pane.rawContent, "secret", "raw content should not include suppressed field after Rerender, got: %q", pane.rawContent)
}

// T-127 (F-020): Rerender preserves scroll offset so toggling a field's
// visibility does not jump the viewport back to the top.
func TestPaneModel_Rerender_PreservesScrollOffset(t *testing.T) {
	// Build an entry with enough fields to guarantee scrolling room.
	raw := `{"a":"1","b":"2","c":"3","d":"4","e":"5","f":"6","g":"7","h":"8","i":"9","j":"10","k":"11","l":"12"}`
	entry := logsource.Entry{IsJSON: true, Time: time.Now(), Level: "INFO", Msg: "x", Raw: []byte(raw)}
	pane := defaultPane(6).SetWidth(40).Open(entry)
	// Scroll down a few lines so we can detect offset preservation.
	pane.scroll.offset = 3
	pane.scroll = pane.scroll.Clamp()
	offBefore := pane.scroll.offset
	if offBefore == 0 {
		t.Skipf("content is too short to scroll; test not applicable")
	}

	pane = pane.WithHiddenFields([]string{"a"}).Rerender()
	offAfter := pane.scroll.offset
	assert.NotEqualf(t, 0, offAfter, "Rerender jumped to top; expected to preserve offset ~%d, got 0", offBefore)
}

// T-127: Rerender on a closed pane is a safe no-op.
func TestPaneModel_Rerender_ClosedPaneNoOp(t *testing.T) {
	pane := defaultPane(20)
	out := pane.Rerender()
	assert.False(t, out.IsOpen(), "Rerender on closed pane must not open it")
}

// bgSGRFor extracts the background SGR prefix that lipgloss emits for a
// given color under the current render profile. Keeps the tests robust
// against termenv's color-profile rounding (truecolor RGB can drift by 1
// when lipgloss routes through 256-color intermediates).
func bgSGRFor(c lipgloss.Color) string {
	rendered := lipgloss.NewStyle().Background(c).Render("x")
	// The SGR is everything from "48;" up to the closing "m" before the
	// character payload. Extract the "48;2;r;g;b" substring — the caller
	// just needs a reliable sentinel.
	i := strings.Index(rendered, "48;")
	if i < 0 {
		return ""
	}
	end := strings.Index(rendered[i:], "m")
	if end < 0 {
		return ""
	}
	return rendered[i : i+end]
}

// T-131: cursor row keeps CursorHighlight bg in both focus states; focused
// combines Bold (SGR 1) with bg, unfocused combines Faint (SGR 2) with bg —
// never both.
func TestPaneModel_View_CursorHighlight_FocusAttr(t *testing.T) {
	th := theme.GetTheme("tokyo-night")
	bgSGR := bgSGRFor(th.CursorHighlight)
	require.NotEmpty(t, bgSGR, "could not synthesize CursorHighlight bg SGR")

	hasAttrBg := func(view string, attr int) bool {
		prefix := "\x1b[" + string(rune('0'+attr)) + ";48;"
		infix := ";" + string(rune('0'+attr)) + ";48;"
		return strings.Contains(view, prefix) || strings.Contains(view, infix)
	}

	for _, tc := range []struct {
		name    string
		focused bool
		wantBg  bool // always true; kept for clarity
		// Exactly one of wantBold/wantFaint is true; the other must be absent.
		wantBold, wantFaint bool
	}{
		{"focused bold", true, true, true, false},
		{"unfocused faint", false, true, false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := paneWithNLines(12, 30, 40)
			m.Focused = tc.focused
			view := m.View()
			assert.Containsf(t, view, bgSGR, "expected CursorHighlight bg SGR %q: %q", bgSGR, view)
			assert.Equalf(t, tc.wantBold, hasAttrBg(view, 1), "bold+bg presence: %q", view)
			assert.Equalf(t, tc.wantFaint, hasAttrBg(view, 2), "faint+bg presence: %q", view)
		})
	}
}

// T-131: cursor row paints at the correct visible position when offset > 0.
func TestPaneModel_View_CursorHighlight_HonorsOffset(t *testing.T) {
	th := theme.GetTheme("tokyo-night")
	m := paneWithNLines(10, 20, 50)
	m.scroll.cursor = 7
	m.scroll.offset = 3
	m.Focused = true
	view := m.View()
	bgSGR := bgSGRFor(th.CursorHighlight)
	assert.Containsf(t, view, bgSGR, "expected CursorHighlight bg in view with cursor=7, offset=3; view=%q", view)
	// visible row = cursor - offset = 4. Rendered content rows start at
	// index 1 (after top border row), so cursor visually lands at row 5.
	rows := strings.Split(view, "\n")
	require.GreaterOrEqualf(t, len(rows), 6, "expected >=6 rows, got %d", len(rows))
	// The cursor row must contain the CursorHighlight bg; other content rows
	// (index 1..4 and 6..contentH) must not.
	for i := 1; i <= 4; i++ {
		assert.NotContainsf(t, rows[i], bgSGR, "row %d should not have CursorHighlight bg; row=%q", i, rows[i])
	}
	assert.Containsf(t, rows[5], bgSGR, "row 5 (visible cursor row) should have CursorHighlight bg; row=%q", rows[5])
}

// T-131: closed pane renders nothing at all — no cursor paint possible.
func TestPaneModel_View_Closed_NoCursor(t *testing.T) {
	m := defaultPane(10)
	assert.Empty(t, m.View(), "closed pane should render empty string")
}

// T-131: SetContent resets cursor to 0.
func TestScrollModel_SetContent_ResetsCursor(t *testing.T) {
	m := NewScrollModel("a\nb\nc\nd\ne\nf", 3)
	m.cursor = 4
	m.offset = 2
	m = m.SetContent("x\ny\nz", 3)
	assert.Equal(t, 0, m.Cursor(), "SetContent should reset cursor to 0")
	assert.Equal(t, 0, m.Offset(), "SetContent should reset offset to 0")
}

// T-141 (F-105): paintCursorRow strips inner `\x1b[0m` resets from the
// cursor row before applying the outer CursorHighlight bg.
func TestPaneModel_PaintCursorRow_T141_StripsInnerResets(t *testing.T) {
	m := defaultPane(10).SetWidth(40)
	m.open = true
	m.Focused = true
	m.scroll = NewScrollModel("", 4)
	m.scroll.cursor = 0
	m.scroll.offset = 0

	body := "\x1b[32m\"key\"\x1b[0m: \x1b[36m\"value\"\x1b[0m,"
	got := m.paintCursorRow(body, 4)

	require.Containsf(t, got, "\x1b[", "painted row missing any SGR: %q", got)
	// Drop trailing terminator.
	trimmed := strings.TrimSuffix(got, "\x1b[0m")
	// Remove legitimate "reset + bg-reopen" transitions from inside the
	// output; any leftover `\x1b[0m` is a bare reset that would visually
	// break the bg run (F-105).
	scrubbed := innerResetAfterBgReopen.ReplaceAllString(trimmed, "")
	assert.NotContainsf(t, scrubbed, "\x1b[0m", "painted cursor row contains a bare inner `\\x1b[0m` (F-105): %q", got)
	// Also assert the outer bold+bg opens exactly once — regression guard
	// for the strip logic. lipgloss's bold-bg opener starts with `\x1b[1;`.
	openers := strings.Count(got, "\x1b[1;")
	assert.Equalf(t, 1, openers, "expected 1 outer bold+bg opener, got %d: %q", openers, got)
}

// innerResetAfterBgReopen matches `\x1b[0m` IMMEDIATELY followed by a bg-
// opening SGR (`\x1b[48…m`) — the legitimate lipgloss text→padding boundary
// that does NOT visually break the bg. Used by the T-141 test to scrub out
// these benign transitions before checking for bare resets.
var innerResetAfterBgReopen = regexp.MustCompile(`\x1b\[0m\x1b\[48[;\d]*m`)

// T-125 / T-139 (F-103): Overlay preserves overall pane width — indicator
// does NOT add columns. Under single-owner border accounting, `SetWidth(n)`
// is content width; outer rendered width = n + 2 border cells.
func TestPaneModel_View_IndicatorDoesNotExpandWidth(t *testing.T) {
	m := paneWithNLines(22, 30, 200)
	view := m.View()
	// Inspect each row's cell width; all should equal outer pane width
	// (content 30 + 2 border cells = 32).
	for i, row := range strings.Split(view, "\n") {
		w := lipgloss.Width(row)
		assert.Equalf(t, 32, w, "row %d cell width (want 32); row=%q", i, row)
	}
}
