package entrylist

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/istonikula/gloggy/internal/config"
	"github.com/istonikula/gloggy/internal/theme"
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

// T9 (I.keys): `M` clears all marks. Cursor/viewport unchanged, no
// SelectionMsg emitted. Empty MarkSet → silent no-op.
func TestListModel_M_ClearAllMarks(t *testing.T) {
	t.Run("clears both marks without moving cursor/offset", func(t *testing.T) {
		m := defaultListModel(10).SetEntries(makeEntries(5))
		// Mark two entries: cursor at 0 → `m`, move to 2 → `m`.
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("m")})
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("m")})
		require.Equal(t, 2, m.Marks().Count(), "pre-M count")
		cursorBefore := m.scroll.Cursor
		offsetBefore := m.scroll.Offset
		m2, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("M")})
		assert.Equal(t, 0, m2.Marks().Count(), "post-M count")
		assert.Equalf(t, cursorBefore, m2.scroll.Cursor, "M moved cursor: %d → %d", cursorBefore, m2.scroll.Cursor)
		assert.Equalf(t, offsetBefore, m2.scroll.Offset, "M moved offset: %d → %d", offsetBefore, m2.scroll.Offset)
		assert.Nilf(t, cmd, "M emitted cmd %v, want nil (no SelectionMsg on clear-all-marks)", cmd)
	})

	t.Run("silent no-op on empty MarkSet", func(t *testing.T) {
		m := defaultListModel(10).SetEntries(makeEntries(5))
		require.Equal(t, 0, m.Marks().Count(), "precondition: Count")
		cursorBefore := m.scroll.Cursor
		m2, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("M")})
		assert.Equal(t, 0, m2.Marks().Count(), "post-M count")
		assert.Equalf(t, cursorBefore, m2.scroll.Cursor, "M on empty moved cursor: %d → %d", cursorBefore, m2.scroll.Cursor)
		assert.Nilf(t, cmd, "M on empty emitted cmd %v, want nil", cmd)
	})
}

// T10 (V4, V5, V26): a marked row must not overflow `m.width`. Previously
// `list.View()` rendered `prefix + RenderCompactRow(m.width)` so a 2-cell
// mark glyph pushed the total row to `m.width+2`, which soft-wraps to 2
// terminal lines and cascades into V5 by displacing the header (B2).
// Post-fix: a 2-cell prefix column is reserved on every row; content is
// sized to `m.width-2`; total row = m.width.
func TestListModel_V26_MarkedRow_NoWidthOverflow(t *testing.T) {
	const innerWidth = 60
	const innerHeight = 5
	entries := makeEntries(innerHeight)
	for i := range entries {
		entries[i].Msg = strings.Repeat("x", 200)
	}
	m := NewListModel(theme.GetTheme("tokyo-night"), config.DefaultConfig(), innerWidth, innerHeight)
	m = m.SetEntries(entries)
	// Mark the cursor-row entry so `* ` prefix is prepended.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("m")})

	view := m.View()
	lines := strings.Split(view, "\n")

	// V5: applyPaneStyle wraps the inner ViewportHeight rows with top +
	// bottom border → exactly innerHeight + 2 output lines.
	assert.Lenf(t, lines, innerHeight+2,
		"View line count should be %d (top border + %d inner rows + bottom border)", innerHeight+2, innerHeight)

	// V4 + V26: each line's visible width must fit in innerWidth + 2
	// (inner content + left-/right- border). Pre-fix: marked row = innerWidth+2
	// content + border = innerWidth+4 → fails here.
	maxAllowed := innerWidth + 2
	for i, line := range lines {
		w := lipgloss.Width(line)
		assert.LessOrEqualf(t, w, maxAllowed, "line %d visible width = %d, want ≤ %d: %q", i, w, maxAllowed, line)
	}
}
