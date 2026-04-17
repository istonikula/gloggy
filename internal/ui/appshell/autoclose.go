package appshell

// Pane minimum-viable dimensions (DESIGN.md §4.4 & §8). When the computed
// pane content size falls below these thresholds the pane is auto-closed
// and a one-time status notice replaces the key hints (T-091).
const (
	MinDetailWidth  = 30 // right-split mode
	MinDetailHeight = 3  // below mode
)

// ShouldAutoCloseDetail reports whether the detail pane is currently below
// the minimum-viable dimensions for its orientation. The caller should
// auto-close the pane when this returns true.
func ShouldAutoCloseDetail(layout Layout) bool {
	if !layout.DetailPaneOpen {
		return false
	}
	if layout.Orientation == OrientationRight {
		return layout.DetailContentWidth() < MinDetailWidth
	}
	// Below mode: pane height must support 1 border row + at least 2 content rows.
	return layout.DetailPaneHeight < MinDetailHeight
}
