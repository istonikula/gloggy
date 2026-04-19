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
//
// Background paint (T-179, config R4 AC 13):
//   - Focused / alone panes render `theme.BaseBg`.
//   - Unfocused-but-visible panes render `theme.UnfocusedBg`.
//
// No rendered pane falls through to the terminal's default background. If
// `UnfocusedBg` is empty on a custom theme, `BaseBg` is used as a defensive
// fallback so the invariant still holds.
func PaneStyle(th theme.Theme, state PaneVisualState) lipgloss.Style {
	border := lipgloss.NormalBorder()
	switch state {
	case PaneStateUnfocused:
		s := lipgloss.NewStyle().
			Border(border).
			BorderForeground(th.DividerColor).
			Faint(true)
		switch {
		case th.UnfocusedBg != "":
			s = s.Background(th.UnfocusedBg)
		case th.BaseBg != "":
			s = s.Background(th.BaseBg)
		}
		return s
	default: // PaneStateFocused
		s := lipgloss.NewStyle().
			Border(border).
			BorderForeground(th.FocusBorder)
		if th.BaseBg != "" {
			s = s.Background(th.BaseBg)
		}
		return s
	}
}

// WithDragSeamTop overrides the top border foreground with theme.DragHandle,
// preserving the left/right/bottom border colours from the focus-state base
// style. Applied to the below-mode detail pane so the row shared with the
// entry list renders as the draggable seam (T-173, Tier 23 kit revision;
// cavekit-app-shell R10 new AC 10, config R4 new AC 9). Focus-neutral per
// app-shell R15 — the seam colour does not shift when focus moves.
//
// This layers on top of PaneStyle rather than introducing a separate 1-row
// strip (the Tier-23 task's alternate "option a") so the pane's internal
// border arithmetic — `borderRows()`, `ContentHeight()`, `SetHeight()` —
// stays invariant. The visible row painted at the top of the pane still
// carries DragHandle's SGR, which is what T-174 asserts.
func WithDragSeamTop(s lipgloss.Style, th theme.Theme) lipgloss.Style {
	return s.BorderTopForeground(th.DragHandle)
}
