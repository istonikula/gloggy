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
