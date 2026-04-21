package entrylist

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// R14 AC1+AC2: cursor on last entry => AppendEntries advances cursor to new
// last and scrolls viewport (scrolloff at bottom edge applies).
func TestAppendEntries_TailFollow_AdvancesCursorAndViewport(t *testing.T) {
	const height = 10
	m := defaultListModel(height).SetEntries(makeEntries(20))
	// Land cursor on the last entry.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	require.Equal(t, 19, m.scroll.Cursor, "setup: cursor")

	before := m.scroll.Offset
	m = m.AppendEntries(makeEntries(5))

	assert.Equal(t, 24, m.scroll.Cursor, "cursor after follow-append (new last)")
	// The last entry must be inside the viewport: Offset <= 24 < Offset+Height.
	assert.Truef(t, m.scroll.Offset <= 24 && 24 < m.scroll.Offset+height,
		"offset=%d with height=%d does not include cursor 24", m.scroll.Offset, height)
	assert.NotEqualf(t, before, m.scroll.Offset, "offset did not advance (before=%d, after=%d)", before, m.scroll.Offset)
}

// R14 AC3+AC4: cursor not on last entry => AppendEntries leaves cursor and
// offset untouched.
func TestAppendEntries_NotAtTail_NoFollow(t *testing.T) {
	const height = 10
	m := defaultListModel(height).SetEntries(makeEntries(20))
	// Move cursor into the middle via j presses (3 times), which will sit at
	// cursor=3 with offset 0. (Cursor is definitely not on the last entry.)
	for i := 0; i < 3; i++ {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	}
	wantCursor := m.scroll.Cursor
	wantOffset := m.scroll.Offset
	require.Lessf(t, wantCursor, m.scroll.TotalEntries-1,
		"setup: cursor at tail; expected mid-list; cursor=%d total=%d", wantCursor, m.scroll.TotalEntries)

	m = m.AppendEntries(makeEntries(5))

	assert.Equalf(t, wantCursor, m.scroll.Cursor, "cursor moved: was %d, now %d (must not change when not at tail)", wantCursor, m.scroll.Cursor)
	assert.Equalf(t, wantOffset, m.scroll.Offset, "offset moved: was %d, now %d (must not change when not at tail)", wantOffset, m.scroll.Offset)
}

// R14 AC5: empty list => first append leaves Cursor=0, Offset=0.
func TestAppendEntries_EmptyList_FirstAppend(t *testing.T) {
	m := defaultListModel(10)
	require.Equal(t, 0, m.scroll.TotalEntries, "setup: total")
	m = m.AppendEntries(makeEntries(5))
	assert.Equal(t, 0, m.scroll.Cursor, "cursor after empty-first-append")
	assert.Equal(t, 0, m.scroll.Offset, "offset after empty-first-append")
}

// R14 AC6: IsAtTail reflects cursor==last; k breaks it, G restores it.
// Edge: empty list is not "at tail".
func TestIsAtTail(t *testing.T) {
	t.Run("empty list is false", func(t *testing.T) {
		m := defaultListModel(10)
		assert.False(t, m.IsAtTail(), "IsAtTail on empty list must be false")
	})

	t.Run("tracks cursor position through j/k/g/G", func(t *testing.T) {
		m := defaultListModel(10).SetEntries(makeEntries(20))
		assert.False(t, m.IsAtTail(), "IsAtTail on fresh list (cursor=0)")
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
		assert.True(t, m.IsAtTail(), "IsAtTail after G")
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
		assert.False(t, m.IsAtTail(), "IsAtTail after k")
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
		assert.True(t, m.IsAtTail(), "IsAtTail after return-to-tail (G)")
	})
}

// R14 combined: after k, subsequent appends must NOT advance viewport even if
// the user had been following. Restoring tail via G must re-engage.
func TestAppendEntries_PauseWithK_ResumeWithG(t *testing.T) {
	const height = 10
	m := defaultListModel(height).SetEntries(makeEntries(20))
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	m = m.AppendEntries(makeEntries(3)) // following: advances
	require.Equal(t, 22, m.scroll.Cursor, "setup: cursor")

	// User presses k to pause.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	pausedCursor := m.scroll.Cursor
	pausedOffset := m.scroll.Offset

	// Appends arrive while paused — no follow.
	m = m.AppendEntries(makeEntries(3))
	assert.Equalf(t, pausedCursor, m.scroll.Cursor, "paused: cursor moved %d -> %d", pausedCursor, m.scroll.Cursor)
	assert.Equalf(t, pausedOffset, m.scroll.Offset, "paused: offset moved %d -> %d", pausedOffset, m.scroll.Offset)

	// User presses G to resume.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	m = m.AppendEntries(makeEntries(3))
	wantLast := 20 + 3 + 3 + 3 - 1
	assert.Equal(t, wantLast, m.scroll.Cursor, "resumed: cursor")
}
