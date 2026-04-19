package appshell

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/istonikula/gloggy/internal/theme"
)

// dividerGlyph is the 1-cell vertical separator used in right-split mode
// (DESIGN.md §4.5).
const dividerGlyph = "│"

// RenderDivider returns `height` rows of a single vertical glyph styled with
// `theme.DragHandle` (T-172 — Tier 23 drag-seam token). The divider renders
// the draggable seam distinctly from the unfocused pane borders (which use
// `DividerColor`). Focus-neutral per app-shell/R15 — the glyph colour does
// not change on focus shift. Used by LayoutModel to compose right-split
// panes (DESIGN.md §4.5, kit app-shell/R2).
func RenderDivider(height int, th theme.Theme) string {
	if height <= 0 {
		return ""
	}
	style := lipgloss.NewStyle().Foreground(th.DragHandle)
	cell := style.Render(dividerGlyph)
	rows := make([]string, height)
	for i := range rows {
		rows[i] = cell
	}
	return strings.Join(rows, "\n")
}
