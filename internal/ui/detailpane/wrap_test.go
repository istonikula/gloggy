package detailpane

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

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
	if len(lines) < 3 {
		t.Fatalf("expected at least 3 wrapped lines, got %d (%q)", len(lines), got)
	}
	for i, line := range lines {
		if w := lipgloss.Width(line); w > 40 {
			t.Errorf("line %d exceeds width 40: got %d cells (%q)", i, w, line)
		}
	}
	// No bytes lost.
	if joined := strings.ReplaceAll(got, "\n", ""); joined != long {
		t.Errorf("wrapped text dropped content:\n got=%q\n want=%q", joined, long)
	}
}

// T-106: short lines are returned unchanged.
func TestSoftWrap_ShortLinePassesThrough(t *testing.T) {
	in := "hello world"
	got := SoftWrap(in, 40)
	if got != in {
		t.Errorf("short line mutated: got %q want %q", got, in)
	}
}

// T-106: width ≤ 0 returns the input unchanged (no wrap attempted).
func TestSoftWrap_ZeroWidthPasses(t *testing.T) {
	in := strings.Repeat("x", 50)
	if got := SoftWrap(in, 0); got != in {
		t.Errorf("width=0 must pass through: got %q want %q", got, in)
	}
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
	if len(lines) < 5 {
		t.Fatalf("expected at least 5 lines, got %d (%q)", len(lines), got)
	}
	if lines[0] != "short" {
		t.Errorf("first line changed: got %q", lines[0])
	}
	if lines[len(lines)-1] != "short" {
		t.Errorf("last line changed: got %q", lines[len(lines)-1])
	}
}

// T-106: ANSI-styled content survives wrapping — the escape sequences are
// preserved across the break, not cut mid-sequence.
func TestSoftWrap_PreservesANSIAcrossBreaks(t *testing.T) {
	styled := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff00ff")).Render(strings.Repeat("Z", 80))
	got := SoftWrap(styled, 20)
	// The wrapped text must still contain the escape sequence prefix
	// somewhere and render every original Z.
	if !strings.Contains(got, "\x1b[") {
		t.Errorf("expected ANSI escape preserved, got: %q", got)
	}
	// Measure each line's printable cell width ≤ 20.
	for i, line := range strings.Split(got, "\n") {
		if w := lipgloss.Width(line); w > 20 {
			t.Errorf("wrapped ANSI line %d exceeds width 20: got %d cells", i, w)
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
			t.Errorf("line %d width %d exceeds 10 cells", i, w)
		}
	}
}

// T-106: PaneModel's View wraps content to the allocated width so no output
// line exceeds the pane's content width.
func TestPaneModel_View_WrapsWithinAllocation(t *testing.T) {
	m := defaultPane(10)
	// A single very long field renders into a JSON line wider than any
	// reasonable pane width.
	long := strings.Repeat("x", 300)
	raw := []byte(`{"msg":"` + long + `"}`)

	m = m.Open(longEntry(raw))
	m = m.SetWidth(50) // outer 50 → inner 48

	out := m.View()
	for i, line := range strings.Split(out, "\n") {
		if w := lipgloss.Width(line); w > 50 {
			t.Errorf("pane line %d exceeds outer width 50: got %d cells", i, w)
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

	if narrowLines <= wideLines {
		t.Errorf("expected narrower pane to produce MORE wrapped scroll lines: wide=%d narrow=%d", wideLines, narrowLines)
	}
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
	if gotLines > 8 {
		t.Errorf("rendered height %d exceeds allocated 8 rows", gotLines)
	}
}
