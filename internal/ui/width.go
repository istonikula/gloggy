// Package ui holds rendering helpers shared across the gloggy sub-models.
package ui

import "github.com/charmbracelet/lipgloss"

// SafeTruncate returns the longest rune-prefix of s whose rendered terminal
// width (lipgloss.Width — cells, NOT bytes) is <= cells. It never cuts a rune
// mid-byte and never over-truncates multibyte (CJK/emoji) input (V36).
//
// Unlike truncateToWidth in appshell, no ellipsis is appended — callers reserve
// exact cell budgets for compact-row columns, so the prefix must fit verbatim.
func SafeTruncate(s string, cells int) string {
	if cells <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= cells {
		return s
	}
	width := 0
	for i, r := range s {
		w := lipgloss.Width(string(r))
		if width+w > cells {
			return s[:i]
		}
		width += w
	}
	return s
}
