package entrylist

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToggleMark(t *testing.T) {
	ms := NewMarkSet()
	ms.Toggle(42)
	assert.True(t, ms.IsMarked(42), "should be marked after toggle")
	ms.Toggle(42)
	assert.False(t, ms.IsMarked(42), "should be unmarked after second toggle")
}

func TestMarkCount(t *testing.T) {
	ms := NewMarkSet()
	ms.Toggle(1)
	ms.Toggle(2)
	assert.Equal(t, 2, ms.Count())
	ms.Toggle(1)
	assert.Equal(t, 1, ms.Count())
}

func TestNextMark(t *testing.T) {
	ms := NewMarkSet()
	visible := []int{10, 20, 30, 40, 50}
	ms.Toggle(30) // index 2
	ms.Toggle(50) // index 4

	assert.Equal(t, 2, ms.NextMark(0, visible), "NextMark(0)")
	assert.Equal(t, 4, ms.NextMark(2, visible), "NextMark(2)")
}

func TestNextMarkWrap(t *testing.T) {
	ms := NewMarkSet()
	visible := []int{10, 20, 30, 40, 50}
	ms.Toggle(20) // index 1
	assert.Equal(t, 1, ms.NextMark(3, visible), "NextMark wrap")
}

func TestPrevMark(t *testing.T) {
	ms := NewMarkSet()
	visible := []int{10, 20, 30, 40, 50}
	ms.Toggle(20) // index 1
	ms.Toggle(40) // index 3
	assert.Equal(t, 3, ms.PrevMark(4, visible), "PrevMark(4)")
}

func TestPrevMarkWrap(t *testing.T) {
	ms := NewMarkSet()
	visible := []int{10, 20, 30, 40, 50}
	ms.Toggle(40) // index 3
	assert.Equal(t, 3, ms.PrevMark(1, visible), "PrevMark wrap")
}

func TestMarkNoMarks(t *testing.T) {
	ms := NewMarkSet()
	visible := []int{10, 20, 30}
	assert.Equal(t, -1, ms.NextMark(0, visible), "NextMark with no marks")
	assert.Equal(t, -1, ms.PrevMark(0, visible), "PrevMark with no marks")
}

func TestMarkClear(t *testing.T) {
	ms := NewMarkSet()
	ms.Toggle(10)
	ms.Toggle(20)
	ms.Toggle(30)
	require.Equal(t, 3, ms.Count(), "pre-Clear count")
	ms.Clear()
	assert.Equal(t, 0, ms.Count(), "post-Clear count")
	for _, id := range []int{10, 20, 30} {
		assert.False(t, ms.IsMarked(id), "IsMarked(%d) after Clear", id)
	}
	// Idempotent on empty set.
	ms.Clear()
	assert.Equal(t, 0, ms.Count(), "second Clear count")
}

func TestMarkPersistsThroughFilterChange(t *testing.T) {
	ms := NewMarkSet()
	ms.Toggle(42)
	// After filtering, different visible list
	visible2 := []int{42, 50}
	assert.True(t, ms.IsMarked(visible2[0]), "mark should persist through filter change")
}
