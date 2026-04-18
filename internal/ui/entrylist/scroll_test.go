package entrylist

import "testing"

func TestGoTop(t *testing.T) {
	s := ScrollState{Cursor: 50, Offset: 40, TotalEntries: 100, ViewportHeight: 20}
	s = GoTop(s)
	if s.Cursor != 0 || s.Offset != 0 {
		t.Errorf("GoTop: cursor=%d offset=%d, want 0,0", s.Cursor, s.Offset)
	}
}

func TestGoBottom(t *testing.T) {
	s := ScrollState{Cursor: 0, Offset: 0, TotalEntries: 100, ViewportHeight: 20}
	s = GoBottom(s)
	if s.Cursor != 99 {
		t.Errorf("GoBottom cursor = %d, want 99", s.Cursor)
	}
	if s.Offset != 80 {
		t.Errorf("GoBottom offset = %d, want 80", s.Offset)
	}
}

func TestGoBottomFewerThanViewport(t *testing.T) {
	s := ScrollState{Cursor: 0, Offset: 0, TotalEntries: 5, ViewportHeight: 20}
	s = GoBottom(s)
	if s.Cursor != 4 || s.Offset != 0 {
		t.Errorf("GoBottom small list: cursor=%d offset=%d, want 4,0", s.Cursor, s.Offset)
	}
}

func TestHalfPageDown(t *testing.T) {
	s := ScrollState{Cursor: 10, Offset: 0, TotalEntries: 100, ViewportHeight: 20}
	s = HalfPageDown(s)
	if s.Cursor != 20 {
		t.Errorf("HalfPageDown cursor = %d, want 20", s.Cursor)
	}
}

func TestHalfPageDownClamp(t *testing.T) {
	s := ScrollState{Cursor: 95, Offset: 80, TotalEntries: 100, ViewportHeight: 20}
	s = HalfPageDown(s)
	if s.Cursor != 99 {
		t.Errorf("HalfPageDown clamped cursor = %d, want 99", s.Cursor)
	}
}

func TestHalfPageUp(t *testing.T) {
	s := ScrollState{Cursor: 30, Offset: 20, TotalEntries: 100, ViewportHeight: 20}
	s = HalfPageUp(s)
	if s.Cursor != 20 {
		t.Errorf("HalfPageUp cursor = %d, want 20", s.Cursor)
	}
}

func TestHalfPageUpClampToZero(t *testing.T) {
	s := ScrollState{Cursor: 3, Offset: 0, TotalEntries: 100, ViewportHeight: 20}
	s = HalfPageUp(s)
	if s.Cursor != 0 || s.Offset != 0 {
		t.Errorf("HalfPageUp clamped: cursor=%d offset=%d, want 0,0", s.Cursor, s.Offset)
	}
}

func TestScrollKeepsCursorVisible(t *testing.T) {
	s := ScrollState{Cursor: 19, Offset: 0, TotalEntries: 100, ViewportHeight: 20}
	s = HalfPageDown(s)
	if s.Cursor < s.Offset || s.Cursor >= s.Offset+s.ViewportHeight {
		t.Errorf("cursor %d not visible in [%d, %d)", s.Cursor, s.Offset, s.Offset+s.ViewportHeight)
	}
}

// T-135 (F-026): followCursor respects scrolloff margin when cursor moves
// toward the bottom edge. viewport=20, scrolloff=5 → cursor must exceed
// viewport-1-scrolloff = 14 before offset shifts.
func TestFollowCursor_Scrolloff_BottomEdge(t *testing.T) {
	s := ScrollState{Cursor: 14, Offset: 0, TotalEntries: 200, ViewportHeight: 20, Scrolloff: 5}
	s = followCursor(s)
	if s.Offset != 0 {
		t.Errorf("cursor at row 14 (=bottom-margin) should not shift offset; got %d", s.Offset)
	}
	s.Cursor = 15
	s = followCursor(s)
	if s.Offset != 1 {
		t.Errorf("cursor at row 15 crosses margin; offset = %d, want 1", s.Offset)
	}
}

// T-135: scrolloff yields at document edges — cursor can reach line 0 / last.
func TestFollowCursor_Scrolloff_YieldsAtEdges(t *testing.T) {
	s := ScrollState{Cursor: 0, Offset: 0, TotalEntries: 200, ViewportHeight: 20, Scrolloff: 5}
	s = followCursor(s)
	if s.Offset != 0 {
		t.Errorf("at top: offset = %d, want 0 (margin yields)", s.Offset)
	}
	s = ScrollState{Cursor: 199, Offset: 180, TotalEntries: 200, ViewportHeight: 20, Scrolloff: 5}
	s = followCursor(s)
	if s.Offset != 180 {
		t.Errorf("at bottom: offset = %d, want 180 (margin yields)", s.Offset)
	}
}

// T-135: scrolloff > viewport/2 clamps to floor(viewport/2).
func TestFollowCursor_ScrolloffClamped(t *testing.T) {
	// viewport=6, scrolloff=10 → effective = 3. viewport-1-so = 2.
	s := ScrollState{Cursor: 2, Offset: 0, TotalEntries: 100, ViewportHeight: 6, Scrolloff: 10}
	s = followCursor(s)
	if s.Offset != 0 {
		t.Errorf("effective scrolloff keeps offset=0 when cursor at margin, got %d", s.Offset)
	}
	s.Cursor = 3
	s = followCursor(s)
	if s.Offset != 1 {
		t.Errorf("effective scrolloff shifts offset after cursor>=3, got %d", s.Offset)
	}
}

// T-135: HalfPageDown now follows scrolloff — cursor lands at center, offset
// shifts to keep margin. viewport=20, scrolloff=5 — HalfPageDown from cursor=0
// moves cursor to 10; cursor=10 is row 10 with offset=0 still inside (10 <= 14).
func TestHalfPageDown_WithScrolloff(t *testing.T) {
	s := ScrollState{Cursor: 0, Offset: 0, TotalEntries: 200, ViewportHeight: 20, Scrolloff: 5}
	s = HalfPageDown(s)
	if s.Cursor != 10 {
		t.Errorf("HalfPageDown cursor = %d, want 10", s.Cursor)
	}
	if s.Offset != 0 {
		t.Errorf("HalfPageDown offset = %d, want 0 (cursor still inside margin)", s.Offset)
	}
}

// T-135: WheelDown — offset +1; cursor drags only when it would enter margin.
// cursor=50, offset=45, viewport=20, scrolloff=5 — cursor at row 5; top margin
// is 5, so cursor is already at the top-margin edge.
// WheelDown 1 → offset=46, minCursor=51, cursor=50<51 → drag to 51.
func TestWheelDown_DragsCursorAtTopMargin(t *testing.T) {
	s := ScrollState{Cursor: 50, Offset: 45, TotalEntries: 200, ViewportHeight: 20, Scrolloff: 5}
	s = WheelDown(s)
	if s.Offset != 46 || s.Cursor != 51 {
		t.Errorf("WheelDown at margin: offset=%d cursor=%d, want 46/51", s.Offset, s.Cursor)
	}
}

// T-135: WheelDown mid-viewport (cursor far from margin) leaves cursor alone.
// cursor=60, offset=50, viewport=20, scrolloff=5 — cursor at row 10, margin 5.
// WheelDown 1 → offset=51, minCursor=56, cursor=60 > 56 → no drag.
func TestWheelDown_NoDragWhenCursorInsideMargin(t *testing.T) {
	s := ScrollState{Cursor: 60, Offset: 50, TotalEntries: 200, ViewportHeight: 20, Scrolloff: 5}
	s = WheelDown(s)
	if s.Offset != 51 || s.Cursor != 60 {
		t.Errorf("WheelDown mid-viewport: offset=%d cursor=%d, want 51/60", s.Offset, s.Cursor)
	}
}

// T-135: WheelUp symmetric — cursor dragged up when leaving bottom margin.
// cursor=60, offset=50, viewport=20, scrolloff=5 — cursor row 10. max allowed
// row = 14. WheelUp 1 → offset=49, maxCursor=63, cursor=60 <= 63 → no drag.
// WheelUp until cursor row > 14: offset must drop enough that cursor-offset > 14.
// After 3 WheelUp: offset=47. maxCursor = 47+14 = 61. cursor=60 <= 61 → no drag.
// After 4 WheelUp: offset=46. maxCursor = 46+14 = 60. cursor=60 <= 60 → no drag.
// After 5 WheelUp: offset=45. maxCursor = 59. cursor=60 > 59 → drag to 59.
func TestWheelUp_DragsCursorAtBottomMargin(t *testing.T) {
	s := ScrollState{Cursor: 60, Offset: 50, TotalEntries: 200, ViewportHeight: 20, Scrolloff: 5}
	for i := 0; i < 4; i++ {
		s = WheelUp(s)
	}
	if s.Offset != 46 || s.Cursor != 60 {
		t.Errorf("after 4 WheelUp: offset=%d cursor=%d, want 46/60", s.Offset, s.Cursor)
	}
	s = WheelUp(s)
	if s.Offset != 45 || s.Cursor != 59 {
		t.Errorf("after 5 WheelUp: offset=%d cursor=%d, want 45/59 (dragged)", s.Offset, s.Cursor)
	}
}

// T-135: GoBottom applies scrolloff — cursor=last, offset at max (margin yields).
func TestGoBottom_WithScrolloff(t *testing.T) {
	s := ScrollState{Cursor: 0, Offset: 0, TotalEntries: 100, ViewportHeight: 20, Scrolloff: 5}
	s = GoBottom(s)
	if s.Cursor != 99 {
		t.Errorf("GoBottom cursor = %d, want 99", s.Cursor)
	}
	if s.Offset != 80 {
		t.Errorf("GoBottom offset = %d, want 80 (at-bottom, margin yields)", s.Offset)
	}
}
