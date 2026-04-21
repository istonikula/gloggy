package detailpane

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/istonikula/gloggy/internal/logsource"
)

func longEntry(raw []byte) logsource.Entry {
	return logsource.Entry{IsJSON: true, LineNumber: 1, Raw: raw}
}

// T-106: SoftWrap splits a line wider than the content budget across
// multiple lines. No byte is dropped.
func TestSoftWrap_LongLineWrapsAtWidth(t *testing.T) {
	long := strings.Repeat("A", 120)
	got := SoftWrap(long, 40)
	lines := strings.Split(got, "\n")
	require.GreaterOrEqualf(t, len(lines), 3, "expected at least 3 wrapped lines, got %d (%q)", len(lines), got)
	for i, line := range lines {
		if w := lipgloss.Width(line); w > 40 {
			assert.Failf(t, "line exceeds width 40", "line %d: got %d cells (%q)", i, w, line)
		}
	}
	// No bytes lost.
	joined := strings.ReplaceAll(got, "\n", "")
	assert.Equalf(t, long, joined, "wrapped text dropped content:\n got=%q\n want=%q", joined, long)
}

// T-106: short lines are returned unchanged.
func TestSoftWrap_ShortLinePassesThrough(t *testing.T) {
	in := "hello world"
	got := SoftWrap(in, 40)
	assert.Equalf(t, in, got, "short line mutated")
}

// T-106: width ≤ 0 returns the input unchanged (no wrap attempted).
func TestSoftWrap_ZeroWidthPasses(t *testing.T) {
	in := strings.Repeat("x", 50)
	got := SoftWrap(in, 0)
	assert.Equalf(t, in, got, "width=0 must pass through")
}

// T-106: multi-line input keeps its newline structure; each line wrapped
// independently at the width budget.
func TestSoftWrap_MultipleLinesIndependent(t *testing.T) {
	short := "short"
	long := strings.Repeat("B", 50)
	in := short + "\n" + long + "\n" + short
	got := SoftWrap(in, 20)
	lines := strings.Split(got, "\n")
	// 1 (short) + >=3 (long/20) + 1 (short)
	require.GreaterOrEqualf(t, len(lines), 5, "expected at least 5 lines, got %d (%q)", len(lines), got)
	assert.Equal(t, "short", lines[0], "first line changed")
	assert.Equal(t, "short", lines[len(lines)-1], "last line changed")
}

// T-106: ANSI-styled content survives wrapping — the escape sequences are
// preserved across the break, not cut mid-sequence.
func TestSoftWrap_PreservesANSIAcrossBreaks(t *testing.T) {
	styled := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff00ff")).Render(strings.Repeat("Z", 80))
	got := SoftWrap(styled, 20)
	// The wrapped text must still contain the escape sequence prefix
	// somewhere and render every original Z.
	assert.Containsf(t, got, "\x1b[", "expected ANSI escape preserved, got: %q", got)
	// Measure each line's printable cell width ≤ 20.
	for i, line := range strings.Split(got, "\n") {
		if w := lipgloss.Width(line); w > 20 {
			assert.Failf(t, "wrapped ANSI line exceeds width 20", "line %d: got %d cells", i, w)
		}
	}
}

// T-106: CJK/emoji double-width characters wrap by cells, not bytes.
func TestSoftWrap_CJKEmojiByCells(t *testing.T) {
	// Each of these glyphs is 2 cells wide.
	content := strings.Repeat("日", 20)
	got := SoftWrap(content, 10) // should wrap at 5 glyphs (= 10 cells)
	for i, line := range strings.Split(got, "\n") {
		if w := lipgloss.Width(line); w > 10 {
			assert.Failf(t, "line exceeds 10 cells", "line %d width %d", i, w)
		}
	}
}

// T-106 / T-139 (F-103): PaneModel's View wraps content to `contentWidth()`
// cells so no output line exceeds the pane's outer budget (= content + 2
// border cells). `SetWidth(n)` takes CONTENT width under single-owner
// border accounting (the layout publishes post-border DetailContentWidth).
func TestPaneModel_View_WrapsWithinAllocation(t *testing.T) {
	m := defaultPane(10)
	// A single very long field renders into a JSON line wider than any
	// reasonable pane width.
	long := strings.Repeat("x", 300)
	raw := []byte(`{"msg":"` + long + `"}`)

	m = m.Open(longEntry(raw))
	m = m.SetWidth(48) // content 48 → outer 50

	out := m.View()
	for i, line := range strings.Split(out, "\n") {
		if w := lipgloss.Width(line); w > 50 {
			assert.Failf(t, "pane line exceeds outer width 50", "line %d: got %d cells", i, w)
		}
	}
}

// T-106: changing the width re-wraps so narrower allocations produce more
// underlying scroll lines than wider ones (re-wrap happened, not a no-op).
func TestPaneModel_SetWidth_Rewraps(t *testing.T) {
	m := defaultPane(20)
	long := strings.Repeat("y", 300)
	raw := []byte(`{"msg":"` + long + `"}`)
	m = m.Open(longEntry(raw))

	m = m.SetWidth(80)
	wideLines := len(m.scroll.lines)

	m = m.SetWidth(30)
	narrowLines := len(m.scroll.lines)

	assert.Greaterf(t, narrowLines, wideLines,
		"expected narrower pane to produce MORE wrapped scroll lines: wide=%d narrow=%d", wideLines, narrowLines)
}

// T-140 (F-104): SoftWrap must re-emit the active SGR at the start of each
// continuation line so a wrapped colored value keeps its color across the
// wrap boundary. Regression: `ansi.HardwrapWc` preserved escape bytes but
// did NOT re-emit the active style after a break, so a long string value
// ended up partially uncolored on wrap (observed on tiny.log:45).
func TestSoftWrap_T140_SGRRestoredOnContinuation(t *testing.T) {
	// Key in green, value in cyan, one long run that forces wrap inside
	// the value. Width 12 forces the value (`longstringvalue`) to break.
	line := "\x1b[32mkey\x1b[0m: \x1b[36mlongstringvalue\x1b[0m"
	got := SoftWrap(line, 12)
	parts := strings.Split(got, "\n")
	require.GreaterOrEqualf(t, len(parts), 2, "expected wrap to produce >=2 lines, got 1: %q", got)
	// Continuation lines (everything after the first) must contain a
	// cyan-SGR opener so the value keeps its color; they must NOT leave
	// the terminal in the previous (key) style.
	for i, p := range parts[1:] {
		assert.Containsf(t, p, "\x1b[36m", "continuation line %d missing cyan SGR reopen: %q", i+1, p)
	}
}

// T-139 (F-103): after the single-owner border fix, a line of exactly
// `contentWidth` cells fits the pane without wrapping; one cell wider wraps.
func TestPaneModel_T139_ExactContentWidthFits(t *testing.T) {
	m := defaultPane(10)
	line40 := strings.Repeat("x", 40)
	m = m.Open(logsource.Entry{IsJSON: false, LineNumber: 1, Raw: []byte(line40)})
	m = m.SetWidth(40)
	// Raw rendering yields the body with the 40-char line; the scroll
	// model's `lines` must still be 1 entry (no wrap).
	assert.Lenf(t, m.scroll.lines, 1, "content width 40 + line 40: expected 1 scroll line (no wrap)")
	// Same with 41 must wrap to 2.
	m2 := defaultPane(10).Open(logsource.Entry{IsJSON: false, LineNumber: 1, Raw: []byte(strings.Repeat("x", 41))}).SetWidth(40)
	assert.Lenf(t, m2.scroll.lines, 2, "content width 40 + line 41: expected 2 scroll lines (wrap)")
}

// T-106: total rendered height never exceeds the allocated pane height,
// even when the wrapped content overflows. The scroll model caps the
// visible viewport at ContentHeight().
func TestPaneModel_View_HeightWithinAllocation(t *testing.T) {
	m := defaultPane(8)
	long := strings.Repeat("z", 500)
	raw := []byte(`{"msg":"` + long + `"}`)
	m = m.Open(longEntry(raw))
	m = m.SetWidth(40)

	out := m.View()
	gotLines := len(strings.Split(out, "\n"))
	// Outer height is content + top border = ContentHeight()+1 = 8.
	assert.LessOrEqualf(t, gotLines, 8, "rendered height %d exceeds allocated 8 rows", gotLines)
}
