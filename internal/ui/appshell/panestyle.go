package appshell

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/istonikula/gloggy/internal/theme"
)

// PaneVisualState selects the pane styling per DESIGN.md §4 matrix (T-100).
// A pane is "focused" when it currently owns keyboard focus OR when it is the
// only visible pane (alone treatment, T-101). Otherwise it is "unfocused".
type PaneVisualState int

const (
	// PaneStateFocused: FocusBorder borders, base background, full-contrast fg.
	PaneStateFocused PaneVisualState = iota
	// PaneStateUnfocused: DividerColor borders, UnfocusedBg background,
	// Faint fg blend toward Dim.
	PaneStateUnfocused
)

// PaneStyle returns the lipgloss style for a pane with the given visual
// state. The style includes a complete rectangular border on all four sides.
// Callers should reserve 2 cells of width and 2 rows of height for the
// border when sizing pane content.
func PaneStyle(th theme.Theme, state PaneVisualState) lipgloss.Style {
	border := lipgloss.NormalBorder()
	switch state {
	case PaneStateUnfocused:
		s := lipgloss.NewStyle().
			Border(border).
			BorderForeground(th.DividerColor).
			Faint(true)
		if th.UnfocusedBg != "" {
			s = s.Background(th.UnfocusedBg)
		}
		return s
	default: // PaneStateFocused
		return lipgloss.NewStyle().
			Border(border).
			BorderForeground(th.FocusBorder)
	}
}
