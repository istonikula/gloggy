package entrylist

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGoTop(t *testing.T) {
	s := ScrollState{Cursor: 50, Offset: 40, TotalEntries: 100, ViewportHeight: 20}
	s = GoTop(s)
	assert.Equal(t, 0, s.Cursor, "GoTop cursor")
	assert.Equal(t, 0, s.Offset, "GoTop offset")
}

func TestGoBottom(t *testing.T) {
	s := ScrollState{Cursor: 0, Offset: 0, TotalEntries: 100, ViewportHeight: 20}
	s = GoBottom(s)
	assert.Equal(t, 99, s.Cursor, "GoBottom cursor")
	assert.Equal(t, 80, s.Offset, "GoBottom offset")
}

func TestGoBottomFewerThanViewport(t *testing.T) {
	s := ScrollState{Cursor: 0, Offset: 0, TotalEntries: 5, ViewportHeight: 20}
	s = GoBottom(s)
	assert.Equal(t, 4, s.Cursor, "GoBottom small list cursor")
	assert.Equal(t, 0, s.Offset, "GoBottom small list offset")
}

func TestHalfPageDown(t *testing.T) {
	s := ScrollState{Cursor: 10, Offset: 0, TotalEntries: 100, ViewportHeight: 20}
	s = HalfPageDown(s)
	assert.Equal(t, 20, s.Cursor)
}

func TestHalfPageDownClamp(t *testing.T) {
	s := ScrollState{Cursor: 95, Offset: 80, TotalEntries: 100, ViewportHeight: 20}
	s = HalfPageDown(s)
	assert.Equal(t, 99, s.Cursor, "HalfPageDown clamped cursor")
}

func TestHalfPageUp(t *testing.T) {
	s := ScrollState{Cursor: 30, Offset: 20, TotalEntries: 100, ViewportHeight: 20}
	s = HalfPageUp(s)
	assert.Equal(t, 20, s.Cursor)
}

func TestHalfPageUpClampToZero(t *testing.T) {
	s := ScrollState{Cursor: 3, Offset: 0, TotalEntries: 100, ViewportHeight: 20}
	s = HalfPageUp(s)
	assert.Equal(t, 0, s.Cursor, "HalfPageUp clamped cursor")
	assert.Equal(t, 0, s.Offset, "HalfPageUp clamped offset")
}

func TestScrollKeepsCursorVisible(t *testing.T) {
	s := ScrollState{Cursor: 19, Offset: 0, TotalEntries: 100, ViewportHeight: 20}
	s = HalfPageDown(s)
	assert.True(t, s.Cursor >= s.Offset && s.Cursor < s.Offset+s.ViewportHeight,
		"cursor %d not visible in [%d, %d)", s.Cursor, s.Offset, s.Offset+s.ViewportHeight)
}

// T-135 (F-026): followCursor respects scrolloff margin when cursor moves
// toward the bottom edge. viewport=20, scrolloff=5 → cursor must exceed
// viewport-1-scrolloff = 14 before offset shifts.
func TestFollowCursor_Scrolloff_BottomEdge(t *testing.T) {
	s := ScrollState{Cursor: 14, Offset: 0, TotalEntries: 200, ViewportHeight: 20, Scrolloff: 5}
	s = followCursor(s)
	assert.Equal(t, 0, s.Offset, "cursor at row 14 (=bottom-margin) should not shift offset")
	s.Cursor = 15
	s = followCursor(s)
	assert.Equal(t, 1, s.Offset, "cursor at row 15 crosses margin")
}

// T-135: scrolloff yields at document edges — cursor can reach line 0 / last.
func TestFollowCursor_Scrolloff_YieldsAtEdges(t *testing.T) {
	s := ScrollState{Cursor: 0, Offset: 0, TotalEntries: 200, ViewportHeight: 20, Scrolloff: 5}
	s = followCursor(s)
	assert.Equal(t, 0, s.Offset, "at top: margin yields")
	s = ScrollState{Cursor: 199, Offset: 180, TotalEntries: 200, ViewportHeight: 20, Scrolloff: 5}
	s = followCursor(s)
	assert.Equal(t, 180, s.Offset, "at bottom: margin yields")
}

// T-135: scrolloff > viewport/2 clamps to floor(viewport/2).
func TestFollowCursor_ScrolloffClamped(t *testing.T) {
	// viewport=6, scrolloff=10 → effective = 3. viewport-1-so = 2.
	s := ScrollState{Cursor: 2, Offset: 0, TotalEntries: 100, ViewportHeight: 6, Scrolloff: 10}
	s = followCursor(s)
	assert.Equal(t, 0, s.Offset, "effective scrolloff keeps offset=0 when cursor at margin")
	s.Cursor = 3
	s = followCursor(s)
	assert.Equal(t, 1, s.Offset, "effective scrolloff shifts offset after cursor>=3")
}

// T-135: HalfPageDown now follows scrolloff — cursor lands at center, offset
// shifts to keep margin. viewport=20, scrolloff=5 — HalfPageDown from cursor=0
// moves cursor to 10; cursor=10 is row 10 with offset=0 still inside (10 <= 14).
func TestHalfPageDown_WithScrolloff(t *testing.T) {
	s := ScrollState{Cursor: 0, Offset: 0, TotalEntries: 200, ViewportHeight: 20, Scrolloff: 5}
	s = HalfPageDown(s)
	assert.Equal(t, 10, s.Cursor, "HalfPageDown cursor")
	assert.Equal(t, 0, s.Offset, "HalfPageDown offset (cursor still inside margin)")
}

// T-135: WheelDown — offset +1; cursor drags only when it would enter margin.
func TestWheelDown_DragsCursorAtTopMargin(t *testing.T) {
	s := ScrollState{Cursor: 50, Offset: 45, TotalEntries: 200, ViewportHeight: 20, Scrolloff: 5}
	s = WheelDown(s)
	assert.Equal(t, 46, s.Offset, "WheelDown at margin offset")
	assert.Equal(t, 51, s.Cursor, "WheelDown at margin cursor")
}

// T-135: WheelDown mid-viewport (cursor far from margin) leaves cursor alone.
func TestWheelDown_NoDragWhenCursorInsideMargin(t *testing.T) {
	s := ScrollState{Cursor: 60, Offset: 50, TotalEntries: 200, ViewportHeight: 20, Scrolloff: 5}
	s = WheelDown(s)
	assert.Equal(t, 51, s.Offset, "WheelDown mid-viewport offset")
	assert.Equal(t, 60, s.Cursor, "WheelDown mid-viewport cursor")
}

// T-135: WheelUp symmetric — cursor dragged up when leaving bottom margin.
func TestWheelUp_DragsCursorAtBottomMargin(t *testing.T) {
	s := ScrollState{Cursor: 60, Offset: 50, TotalEntries: 200, ViewportHeight: 20, Scrolloff: 5}
	for i := 0; i < 4; i++ {
		s = WheelUp(s)
	}
	assert.Equal(t, 46, s.Offset, "after 4 WheelUp offset")
	assert.Equal(t, 60, s.Cursor, "after 4 WheelUp cursor")
	s = WheelUp(s)
	assert.Equal(t, 45, s.Offset, "after 5 WheelUp offset")
	assert.Equal(t, 59, s.Cursor, "after 5 WheelUp cursor (dragged)")
}

// T-135: GoBottom applies scrolloff — cursor=last, offset at max (margin yields).
func TestGoBottom_WithScrolloff(t *testing.T) {
	s := ScrollState{Cursor: 0, Offset: 0, TotalEntries: 100, ViewportHeight: 20, Scrolloff: 5}
	s = GoBottom(s)
	assert.Equal(t, 99, s.Cursor, "GoBottom cursor")
	assert.Equal(t, 80, s.Offset, "GoBottom offset (at-bottom, margin yields)")
}
