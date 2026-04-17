package appshell

import "github.com/istonikula/gloggy/internal/config"

// Orientation describes how the detail pane is composed alongside the entry
// list.
type Orientation int

const (
	// OrientationBelow stacks the detail pane below the entry list (the
	// original Tier 1-8 layout).
	OrientationBelow Orientation = iota
	// OrientationRight splits the main area horizontally with the detail
	// pane on the right (Tier 9 redesign).
	OrientationRight
)

// String returns the canonical token for the orientation.
func (o Orientation) String() string {
	switch o {
	case OrientationRight:
		return "right"
	default:
		return "below"
	}
}

// SelectOrientation resolves the effective pane orientation for a given
// terminal width and config.
//
// Rules (DESIGN.md §5, kit app-shell/R7):
//   - position = "below"         → OrientationBelow
//   - position = "right"         → OrientationRight
//   - position = "auto" (default)→ OrientationRight when
//     width >= OrientationThresholdCols, else OrientationBelow
//
// Invalid position values are treated as "auto" so the caller cannot corrupt
// the layout by passing a malformed config (Load already validates, this is a
// belt-and-braces fallback).
func SelectOrientation(width int, cfg config.Config) Orientation {
	switch cfg.DetailPane.Position {
	case "below":
		return OrientationBelow
	case "right":
		return OrientationRight
	}
	threshold := cfg.DetailPane.OrientationThresholdCols
	if threshold <= 0 {
		threshold = 100
	}
	if width >= threshold {
		return OrientationRight
	}
	return OrientationBelow
}
