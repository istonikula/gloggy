package entrylist

import "sort"

// MarkSet tracks marked entry IDs. Keyed by entry identity so marks persist through filter changes.
type MarkSet struct {
	marks map[int]bool
}

// NewMarkSet creates an empty MarkSet.
func NewMarkSet() *MarkSet {
	return &MarkSet{marks: make(map[int]bool)}
}

// Toggle adds a mark if absent, removes it if present.
func (ms *MarkSet) Toggle(id int) {
	if ms.marks[id] {
		delete(ms.marks, id)
	} else {
		ms.marks[id] = true
	}
}

// IsMarked returns whether the given ID is marked.
func (ms *MarkSet) IsMarked(id int) bool {
	return ms.marks[id]
}

// Count returns the number of marked entries.
func (ms *MarkSet) Count() int {
	return len(ms.marks)
}

// All returns all marked IDs in sorted order.
func (ms *MarkSet) All() []int {
	ids := make([]int, 0, len(ms.marks))
	for id := range ms.marks {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	return ids
}

// NextMark finds the next marked index after currentIdx in visibleIDs, with wrap.
// Returns -1 if no marks exist in visibleIDs.
func (ms *MarkSet) NextMark(currentIdx int, visibleIDs []int) int {
	n := len(visibleIDs)
	if n == 0 || ms.Count() == 0 {
		return -1
	}
	for i := 1; i <= n; i++ {
		idx := (currentIdx + i) % n
		if ms.marks[visibleIDs[idx]] {
			return idx
		}
	}
	return -1
}

// PrevMark finds the previous marked index before currentIdx in visibleIDs, with wrap.
// Returns -1 if no marks exist in visibleIDs.
func (ms *MarkSet) PrevMark(currentIdx int, visibleIDs []int) int {
	n := len(visibleIDs)
	if n == 0 || ms.Count() == 0 {
		return -1
	}
	for i := 1; i <= n; i++ {
		idx := (currentIdx - i%n + n) % n
		if ms.marks[visibleIDs[idx]] {
			return idx
		}
	}
	return -1
}
