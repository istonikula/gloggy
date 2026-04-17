package detailpane

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/lipgloss"
)

// SoftWrap wraps each line of content at width cells, returning the wrapped
// text. Wrapping is ANSI-safe (escape sequences are preserved across line
// breaks) and width-aware (emoji and CJK count as 2 cells each via
// go-runewidth, see T-107).
//
// width <= 0 returns the input unchanged.
//
// T-106 (cavekit-detail-pane R9, wrap_mode = "soft"): when a content line
// exceeds the pane content width, it wraps onto the next line at the cell
// boundary. No silent hard truncation — every cell is preserved.
func SoftWrap(content string, width int) string {
	if width <= 0 || content == "" {
		return content
	}
	in := strings.Split(content, "\n")
	out := make([]string, 0, len(in))
	for _, line := range in {
		if lipgloss.Width(line) <= width {
			out = append(out, line)
			continue
		}
		wrapped := ansi.HardwrapWc(line, width, true)
		out = append(out, strings.Split(wrapped, "\n")...)
	}
	return strings.Join(out, "\n")
}
