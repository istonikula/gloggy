// Package entrylist provides pure-logic functions for entry list navigation.
package entrylist

// ScrollState holds cursor and viewport state.
type ScrollState struct {
	Cursor         int
	Offset         int
	TotalEntries   int
	ViewportHeight int
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

func ensureVisible(cursor, offset, height int) int {
	if cursor < offset {
		return cursor
	}
	if cursor >= offset+height {
		return cursor - height + 1
	}
	return offset
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
	s.Offset = ensureVisible(s.Cursor, s.Offset, s.ViewportHeight)
	s.Offset = clampOffset(s.Offset, s.TotalEntries, s.ViewportHeight)
	return s
}

// HalfPageDown moves cursor down by half the viewport height.
func HalfPageDown(s ScrollState) ScrollState {
	half := s.ViewportHeight / 2
	if half < 1 {
		half = 1
	}
	s.Cursor = clampCursor(s.Cursor+half, s.TotalEntries)
	s.Offset = ensureVisible(s.Cursor, s.Offset, s.ViewportHeight)
	s.Offset = clampOffset(s.Offset, s.TotalEntries, s.ViewportHeight)
	return s
}

// HalfPageUp moves cursor up by half the viewport height.
func HalfPageUp(s ScrollState) ScrollState {
	half := s.ViewportHeight / 2
	if half < 1 {
		half = 1
	}
	s.Cursor = clampCursor(s.Cursor-half, s.TotalEntries)
	s.Offset = ensureVisible(s.Cursor, s.Offset, s.ViewportHeight)
	s.Offset = clampOffset(s.Offset, s.TotalEntries, s.ViewportHeight)
	return s
}
