package detailpane

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/istonikula/gloggy/internal/logsource"
	"github.com/istonikula/gloggy/internal/theme"
)

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

