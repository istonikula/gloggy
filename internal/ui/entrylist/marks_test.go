package entrylist

import "testing"

func TestToggleMark(t *testing.T) {
	ms := NewMarkSet()
	ms.Toggle(42)
	if !ms.IsMarked(42) {
		t.Error("should be marked after toggle")
	}
	ms.Toggle(42)
	if ms.IsMarked(42) {
		t.Error("should be unmarked after second toggle")
	}
}

func TestMarkCount(t *testing.T) {
	ms := NewMarkSet()
	ms.Toggle(1)
	ms.Toggle(2)
	if ms.Count() != 2 {
		t.Errorf("count = %d, want 2", ms.Count())
	}
	ms.Toggle(1)
	if ms.Count() != 1 {
		t.Errorf("count = %d, want 1", ms.Count())
	}
}

func TestNextMark(t *testing.T) {
	ms := NewMarkSet()
	visible := []int{10, 20, 30, 40, 50}
	ms.Toggle(30) // index 2
	ms.Toggle(50) // index 4

	if ms.NextMark(0, visible) != 2 {
		t.Errorf("NextMark(0) = %d, want 2", ms.NextMark(0, visible))
	}
	if ms.NextMark(2, visible) != 4 {
		t.Errorf("NextMark(2) = %d, want 4", ms.NextMark(2, visible))
	}
}

func TestNextMarkWrap(t *testing.T) {
	ms := NewMarkSet()
	visible := []int{10, 20, 30, 40, 50}
	ms.Toggle(20) // index 1
	if ms.NextMark(3, visible) != 1 {
		t.Errorf("NextMark wrap = %d, want 1", ms.NextMark(3, visible))
	}
}

func TestPrevMark(t *testing.T) {
	ms := NewMarkSet()
	visible := []int{10, 20, 30, 40, 50}
	ms.Toggle(20) // index 1
	ms.Toggle(40) // index 3
	if ms.PrevMark(4, visible) != 3 {
		t.Errorf("PrevMark(4) = %d, want 3", ms.PrevMark(4, visible))
	}
}

func TestPrevMarkWrap(t *testing.T) {
	ms := NewMarkSet()
	visible := []int{10, 20, 30, 40, 50}
	ms.Toggle(40) // index 3
	if ms.PrevMark(1, visible) != 3 {
		t.Errorf("PrevMark wrap = %d, want 3", ms.PrevMark(1, visible))
	}
}

func TestMarkNoMarks(t *testing.T) {
	ms := NewMarkSet()
	visible := []int{10, 20, 30}
	if ms.NextMark(0, visible) != -1 {
		t.Error("NextMark with no marks should return -1")
	}
	if ms.PrevMark(0, visible) != -1 {
		t.Error("PrevMark with no marks should return -1")
	}
}

func TestMarkClear(t *testing.T) {
	ms := NewMarkSet()
	ms.Toggle(10)
	ms.Toggle(20)
	ms.Toggle(30)
	if ms.Count() != 3 {
		t.Fatalf("pre-Clear count = %d, want 3", ms.Count())
	}
	ms.Clear()
	if ms.Count() != 0 {
		t.Errorf("post-Clear count = %d, want 0", ms.Count())
	}
	for _, id := range []int{10, 20, 30} {
		if ms.IsMarked(id) {
			t.Errorf("IsMarked(%d) = true after Clear, want false", id)
		}
	}
	// Idempotent on empty set.
	ms.Clear()
	if ms.Count() != 0 {
		t.Errorf("second Clear count = %d, want 0", ms.Count())
	}
}

func TestMarkPersistsThroughFilterChange(t *testing.T) {
	ms := NewMarkSet()
	ms.Toggle(42)
	// After filtering, different visible list
	visible2 := []int{42, 50}
	if !ms.IsMarked(visible2[0]) {
		t.Error("mark should persist through filter change")
	}
}
