package appshell

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/istonikula/gloggy/internal/theme"
)

// bgSGRSentinel is a sentinel character used only to split lipgloss's bg-only
// rendered output into opening/closing SGR halves. NUL is chosen because it
// cannot appear in normal TUI text.
const bgSGRSentinel = "\x00"

// BgSGROpen returns the ANSI SGR opening sequence that sets the terminal's
// background to c, e.g. `"\x1b[48;2;R;G;Bm"` under TrueColor. Returns "" when
// the active color profile has no bg representation (e.g. the Ascii profile
// with monochrome terminals), or when c is empty.
func BgSGROpen(c lipgloss.Color) string {
	if string(c) == "" {
		return ""
	}
	rendered := lipgloss.NewStyle().Background(c).Render(bgSGRSentinel)
	idx := strings.Index(rendered, bgSGRSentinel)
	if idx <= 0 {
		return ""
	}
	return rendered[:idx]
}

// RepaintBg rewrites body so every embedded SGR reset (`\x1b[0m`) is
// immediately followed by a bg-set sequence for color bg. Without this,
// per-token syntax-highlight Foreground styles emit a full reset that
// "punches out" any outer lipgloss Background — the outer bg visually
// terminates at the first inner reset, leaving the rest of the row on the
// terminal default bg.
//
// This generalizes the single-row fix in `detailpane.PaneModel.paintCursorRow`
// (T-141, F-105) to entire pane bodies. Apply right before the outer
// PaneStyle wrap, once the body is fully assembled (post-indicator,
// post-cursor-highlight, post-search-prompt).
//
// Returns body unchanged when bg produces no SGR opening sequence (ASCII
// profile, empty color) so monochrome environments are not disturbed.
func RepaintBg(body string, bg lipgloss.Color) string {
	open := BgSGROpen(bg)
	if open == "" {
		return body
	}
	return strings.ReplaceAll(body, "\x1b[0m", "\x1b[0m"+open)
}

// PaneBg returns the background color that PaneStyle paints for the given
// visual state, mirroring the logic in `PaneStyle` itself. Use this to feed
// `RepaintBg` without reaching into theme internals at every call site.
func PaneBg(th theme.Theme, state PaneVisualState) lipgloss.Color {
	switch state {
	case PaneStateUnfocused:
		if th.UnfocusedBg != "" {
			return th.UnfocusedBg
		}
		return th.BaseBg
	default: // PaneStateFocused
		return th.BaseBg
	}
}
