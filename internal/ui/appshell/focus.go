package appshell

// NextFocus computes the next focus target after a Tab key press.
//
// Rules (DESIGN.md §6, kit app-shell/R11):
//   - When any overlay is open (filter panel or help), Tab is inert; focus
//     stays where it is.
//   - Otherwise, focus cycles through the visible panes in the order they
//     appear in `visible`. If `current` is the last entry, it wraps to the
//     first. If `current` is not in `visible` (e.g. the pane was just
//     closed), focus jumps to the first visible pane.
//   - Tab never closes a pane.
//
// `visible` is the ordered list of panes eligible for cycling. The caller
// derives this from the layout state — typically [FocusEntryList] when the
// detail pane is closed and [FocusEntryList, FocusDetailPane] when it is
// open. The filter panel and help overlay are NOT in `visible` because
// they are overlays, not panes.
func NextFocus(current FocusTarget, visible []FocusTarget, overlayOpen bool) FocusTarget {
	if overlayOpen {
		return current
	}
	if len(visible) == 0 {
		return current
	}
	for i, t := range visible {
		if t == current {
			return visible[(i+1)%len(visible)]
		}
	}
	return visible[0]
}
