package detailpane

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/lipgloss"
)

// sgrSeq matches an SGR escape sequence (CSI … m).
var sgrSeq = regexp.MustCompile(`\x1b\[[0-9;]*m`)

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
//
// T-140 (F-104): `ansi.HardwrapWc` preserves escape bytes in-order but does
// NOT close and re-open the active SGR state at wrap boundaries. Downstream
// consumers (lipgloss borders / pane style) render each line as an
// independent row, so a continuation line whose styling was opened on the
// previous line renders in the terminal default colours — the "second half
// uncolored" observation on `logs/tiny.log` line 45. We post-process the
// hardwrap output: scan each produced line, track the active SGR across
// newlines, emit `\x1b[0m` at the end of each wrapped segment, and re-emit
// the active SGR at the start of the next.
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
		out = append(out, strings.Split(preserveSGRAcrossBreaks(wrapped), "\n")...)
	}
	return strings.Join(out, "\n")
}

// preserveSGRAcrossBreaks walks `wrapped` (a hardwrap output potentially
// containing `\n` in the middle of a styled region) and re-emits the active
// SGR state across each wrap boundary. Active state accumulates on
// non-reset SGR sequences and clears on `\x1b[0m` / `\x1b[m`. At each `\n`
// where active != "", emit `\x1b[0m` before the newline and the accumulated
// active SGR after it.
func preserveSGRAcrossBreaks(wrapped string) string {
	if !strings.Contains(wrapped, "\n") {
		return wrapped
	}
	var out strings.Builder
	out.Grow(len(wrapped) + 16)
	active := "" // last-seen non-reset SGR; concatenated until reset
	i := 0
	for i < len(wrapped) {
		if wrapped[i] == '\n' {
			if active != "" {
				out.WriteString("\x1b[0m\n")
				out.WriteString(active)
			} else {
				out.WriteByte('\n')
			}
			i++
			continue
		}
		if loc := sgrSeq.FindStringIndex(wrapped[i:]); loc != nil && loc[0] == 0 {
			seq := wrapped[i : i+loc[1]]
			out.WriteString(seq)
			if seq == "\x1b[0m" || seq == "\x1b[m" {
				active = ""
			} else {
				active += seq
			}
			i += loc[1]
			continue
		}
		out.WriteByte(wrapped[i])
		i++
	}
	return out.String()
}
