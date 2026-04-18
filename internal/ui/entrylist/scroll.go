// Package entrylist provides pure-logic functions for entry list navigation.
package entrylist

// ScrollState holds cursor and viewport state.
// T-135 (F-026): Scrolloff is the shared top-level cursor margin (nvim-style).
// Clamped to [0, floor(ViewportHeight/2)] at use time so the margin never
// exceeds half the viewport — keeps cursor movement possible in tiny windows.
type ScrollState struct {
	Cursor         int
	Offset         int
	TotalEntries   int
	ViewportHeight int
	Scrolloff      int
}

func clampCursor(cursor, total int) int {
	if total <= 0 {
		return 0
	}
	if cursor < 0 {
		return 0
	}
	if cursor >= total {
		return total - 1
	}
	return cursor
}

func clampOffset(offset, total, height int) int {
	if offset < 0 {
		return 0
	}
	max := total - height
	if max < 0 {
		max = 0
	}
	if offset > max {
		return max
	}
	return offset
}

// effectiveScrolloff returns the scrolloff clamped to [0, floor(height/2)].
func effectiveScrolloff(scrolloff, height int) int {
	if scrolloff < 0 {
		return 0
	}
	max := height / 2
	if scrolloff > max {
		return max
	}
	return scrolloff
}

// ensureVisible keeps cursor inside the viewport (margin=0 baseline).
// Retained for mouse-click selection and filter reshaping — those paths
// do not want scrolloff margins applied because the click already names
// the target row.
func ensureVisible(cursor, offset, height int) int {
	if cursor < offset {
		return cursor
	}
	if cursor >= offset+height {
		return cursor - height + 1
	}
	return offset
}

// followCursor adjusts offset so cursor stays >= scrolloff rows from each
// viewport edge (T-135, F-026). Document edges yield — cursor can reach
// row 0 or the last row. Mirrors detailpane.ScrollModel.followCursor.
func followCursor(s ScrollState) ScrollState {
	so := effectiveScrolloff(s.Scrolloff, s.ViewportHeight)
	top := s.Offset + so
	bottom := s.Offset + s.ViewportHeight - 1 - so
	if s.Cursor < top {
		s.Offset = s.Cursor - so
	} else if s.Cursor > bottom {
		s.Offset = s.Cursor - s.ViewportHeight + 1 + so
	}
	s.Offset = clampOffset(s.Offset, s.TotalEntries, s.ViewportHeight)
	return s
}

// GoTop moves cursor to the first entry.
func GoTop(s ScrollState) ScrollState {
	s.Cursor = 0
	s.Offset = 0
	return s
}

// GoBottom moves cursor to the last entry.
func GoBottom(s ScrollState) ScrollState {
	s.Cursor = clampCursor(s.TotalEntries-1, s.TotalEntries)
	s = followCursor(s)
	return s
}

// HalfPageDown moves cursor down by half the viewport height.
func HalfPageDown(s ScrollState) ScrollState {
	half := s.ViewportHeight / 2
	if half < 1 {
		half = 1
	}
	s.Cursor = clampCursor(s.Cursor+half, s.TotalEntries)
	s = followCursor(s)
	return s
}

// HalfPageUp moves cursor up by half the viewport height.
func HalfPageUp(s ScrollState) ScrollState {
	half := s.ViewportHeight / 2
	if half < 1 {
		half = 1
	}
	s.Cursor = clampCursor(s.Cursor-half, s.TotalEntries)
	s = followCursor(s)
	return s
}

// WheelDown scrolls offset by 1 and drags cursor along if it would enter
// the top scrolloff margin. T-135 (F-026).
func WheelDown(s ScrollState) ScrollState {
	s.Offset = clampOffset(s.Offset+1, s.TotalEntries, s.ViewportHeight)
	so := effectiveScrolloff(s.Scrolloff, s.ViewportHeight)
	minCursor := s.Offset + so
	if s.Cursor < minCursor {
		s.Cursor = clampCursor(minCursor, s.TotalEntries)
	}
	return s
}

// WheelUp scrolls offset up by 1 and drags cursor along if it would leave
// the bottom scrolloff margin. T-135 (F-026).
func WheelUp(s ScrollState) ScrollState {
	s.Offset = clampOffset(s.Offset-1, s.TotalEntries, s.ViewportHeight)
	so := effectiveScrolloff(s.Scrolloff, s.ViewportHeight)
	maxCursor := s.Offset + s.ViewportHeight - 1 - so
	if s.Cursor > maxCursor {
		s.Cursor = clampCursor(maxCursor, s.TotalEntries)
	}
	return s
}
