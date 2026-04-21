package detailpane

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeScroll(lines int, height int) ScrollModel {
	content := make([]string, lines)
	for i := range content {
		content[i] = "line"
	}
	return NewScrollModel(strings.Join(content, "\n"), height)
}

// T-132 (F-026): j moves the cursor, not the offset. With default scrolloff=0
// on a 20-line doc with viewport=5, `j` x 1 moves cursor 0→1; cursor still
// inside viewport so offset stays at 0.
func TestScrollModel_JMovesCursor_NoViewportShift(t *testing.T) {
	m := makeScroll(20, 5)
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	assert.Equal(t, 1, m2.cursor, "cursor")
	assert.Equal(t, 0, m2.offset, "offset (cursor still inside viewport)")
}

// T-132: once cursor reaches the bottom edge of the viewport the viewport
// follows. scrolloff=0, height=5 → cursor must pass row 4 for offset to move.
func TestScrollModel_JAtViewportEdge_ShiftsOffset(t *testing.T) {
	m := makeScroll(20, 5)
	for i := 0; i < 4; i++ {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	}
	require.Equal(t, 4, m.cursor, "precondition: cursor")
	require.Equal(t, 0, m.offset, "precondition: offset")
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	assert.Equal(t, 5, m.cursor, "after 5th j: cursor")
	assert.Equal(t, 1, m.offset, "after 5th j: offset")
}

// T-132: k at top is no-op for cursor (0), offset stays 0.
func TestScrollModel_KAtTopIsNoop(t *testing.T) {
	m := makeScroll(20, 5)
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	assert.Equal(t, 0, m2.cursor, "cursor")
	assert.Equal(t, 0, m2.offset, "offset")
}

// T-132: scrolloff=3 on viewport=10 → cursor must pass line 6 before offset
// moves (viewport-1-scrolloff = 6).
func TestScrollModel_Scrolloff3_FollowsAtRow6(t *testing.T) {
	m := makeScroll(100, 10).WithScrolloff(3)
	for i := 0; i < 6; i++ {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	}
	assert.Equal(t, 6, m.cursor, "after 6 j: cursor")
	assert.Equal(t, 0, m.offset, "after 6 j: offset")
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	assert.Equal(t, 7, m.cursor, "after 7th j: cursor")
	assert.Equal(t, 1, m.offset, "after 7th j: offset")
}

// T-133: wheel scrolls offset first; cursor drags only when the margin is
// crossed. With default scrolloff=0 a single wheel tick moves offset 0→1 and
// drags cursor 0→1 (since cursor was at offset+0 row, now above top).
func TestScrollModel_MouseWheelScrolls(t *testing.T) {
	m := makeScroll(20, 5)
	m2, _ := m.Update(tea.MouseMsg{Button: tea.MouseButtonWheelDown})
	assert.Equal(t, 1, m2.offset, "WheelDown: offset")
	m3, _ := m2.Update(tea.MouseMsg{Button: tea.MouseButtonWheelUp})
	assert.Equal(t, 0, m3.offset, "WheelUp: offset")
}

// T-133: wheel in the middle of a doc with scrolloff=3 does NOT drag the
// cursor until the margin is crossed.
func TestScrollModel_WheelDown_DragsCursorAtScrolloffEdge(t *testing.T) {
	m := makeScroll(100, 10).WithScrolloff(3)
	m.offset = 45
	m.cursor = 50
	m, _ = m.Update(tea.MouseMsg{Button: tea.MouseButtonWheelDown})
	assert.Equal(t, 46, m.offset, "after 1 WheelDown: offset")
	assert.Equal(t, 50, m.cursor, "after 1 WheelDown: cursor")
	m, _ = m.Update(tea.MouseMsg{Button: tea.MouseButtonWheelDown})
	assert.Equal(t, 47, m.offset, "after 2 WheelDown: offset")
	assert.Equal(t, 50, m.cursor, "after 2 WheelDown: cursor")
	m, _ = m.Update(tea.MouseMsg{Button: tea.MouseButtonWheelDown})
	assert.Equal(t, 48, m.offset, "after 3 WheelDown: offset")
	assert.Equal(t, 51, m.cursor, "after 3 WheelDown: cursor (dragged at margin)")
}

// T-133 symmetric: WheelUp drags cursor when cursor would exceed the bottom
// margin.
func TestScrollModel_WheelUp_DragsCursorAtScrolloffEdge(t *testing.T) {
	m := makeScroll(100, 10).WithScrolloff(3)
	m.offset = 45
	m.cursor = 50
	m, _ = m.Update(tea.MouseMsg{Button: tea.MouseButtonWheelUp})
	assert.Equal(t, 44, m.offset, "after 1 WheelUp: offset")
	assert.Equal(t, 50, m.cursor, "after 1 WheelUp: cursor")
	m, _ = m.Update(tea.MouseMsg{Button: tea.MouseButtonWheelUp})
	assert.Equal(t, 43, m.offset, "after 2 WheelUp: offset")
	assert.Equal(t, 49, m.cursor, "after 2 WheelUp: cursor (dragged at margin)")
}

// T-132: g/Home → cursor=0, offset=0.
func TestScrollModel_GJumpsToTop(t *testing.T) {
	m := makeScroll(50, 10)
	m.cursor = 25
	m.offset = 20
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
	assert.Equal(t, 0, m2.cursor, "g: cursor")
	assert.Equal(t, 0, m2.offset, "g: offset")
	m3, _ := m.Update(tea.KeyMsg{Type: tea.KeyHome})
	assert.Equal(t, 0, m3.cursor, "home: cursor")
	assert.Equal(t, 0, m3.offset, "home: offset")
}

// T-132: G/End → cursor = last, offset = max so cursor visible at bottom.
// 50 lines, height=10 → cursor=49, offset=40.
func TestScrollModel_GCapJumpsToBottom(t *testing.T) {
	m := makeScroll(50, 10)
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	assert.Equal(t, 49, m2.cursor, "G: cursor")
	assert.Equal(t, 40, m2.offset, "G: offset")
	m3, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnd})
	assert.Equal(t, 49, m3.cursor, "end: cursor")
	assert.Equal(t, 40, m3.offset, "end: offset")
}

// T-132: G on content shorter than viewport — offset stays 0, cursor=last.
func TestScrollModel_GOnShortContent(t *testing.T) {
	m := makeScroll(5, 20)
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	assert.Equal(t, 0, m2.offset, "G short: offset")
	assert.Equal(t, 4, m2.cursor, "G short: cursor")
}

// T-132: PgDn from top with scrolloff=0 moves cursor by height-1; cursor
// reaches the far edge so offset follows.
func TestScrollModel_PageDownFromTop(t *testing.T) {
	cases := []struct {
		name string
		msg  tea.KeyMsg
	}{
		{"pgdown", tea.KeyMsg{Type: tea.KeyPgDown}},
		{"ctrl+d", tea.KeyMsg{Type: tea.KeyCtrlD}},
		{"space", tea.KeyMsg{Type: tea.KeySpace, Runes: []rune(" ")}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := makeScroll(50, 10)
			m2, _ := m.Update(tc.msg)
			assert.Equalf(t, 9, m2.cursor, "%s: cursor", tc.name)
			assert.Equalf(t, 0, m2.offset, "%s: offset (cursor at bottom of viewport)", tc.name)
		})
	}
}

// T-132: PgUp at top is no-op (cursor already at 0).
func TestScrollModel_PageUpAtTopIsNoop(t *testing.T) {
	cases := []struct {
		name string
		msg  tea.KeyMsg
	}{
		{"pgup", tea.KeyMsg{Type: tea.KeyPgUp}},
		{"ctrl+u", tea.KeyMsg{Type: tea.KeyCtrlU}},
		{"b", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := makeScroll(50, 10)
			m2, _ := m.Update(tc.msg)
			assert.Equalf(t, 0, m2.cursor, "%s: cursor", tc.name)
			assert.Equalf(t, 0, m2.offset, "%s: offset", tc.name)
		})
	}
}

// T-132: PgUp after G returns toward top by height-1 via cursor movement.
// After G: cursor=49, offset=40. PgUp by 9 → cursor=40, offset=31 (cursor
// sits at viewport top edge, so followCursor pulls offset down to 40-9=31).
func TestScrollModel_PageUpAfterEnd(t *testing.T) {
	m := makeScroll(50, 10)
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	require.Equal(t, 49, m2.cursor, "precondition: cursor")
	require.Equal(t, 40, m2.offset, "precondition: offset")
	m3, _ := m2.Update(tea.KeyMsg{Type: tea.KeyPgUp})
	assert.Equal(t, 40, m3.cursor, "pgup after G: cursor")
	assert.Equal(t, 40, m3.offset, "pgup after G: offset (cursor still inside)")
}

// T-132: PgDn clamps — cursor cannot exceed last line.
func TestScrollModel_PageDownClampsAtBottom(t *testing.T) {
	m := makeScroll(15, 10)
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	// cursor 0+9=9, viewport=10, still inside → offset=0
	assert.Equal(t, 9, m2.cursor, "first pgdown: cursor")
	assert.Equal(t, 0, m2.offset, "first pgdown: offset")
	m3, _ := m2.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	// cursor 9+9=18 clamped to 14 (last line index). offset follows: cursor=14
	// > offset+9 (0+9=9) → offset = 14-9 = 5.
	assert.Equal(t, 14, m3.cursor, "second pgdown: cursor")
	assert.Equal(t, 5, m3.offset, "second pgdown: offset")
}

// T-132: page keys on content shorter than viewport — no shift.
func TestScrollModel_PageKeysOnShortContent(t *testing.T) {
	m := makeScroll(5, 10)
	for _, tc := range []struct {
		name string
		msg  tea.KeyMsg
	}{
		{"pgdown", tea.KeyMsg{Type: tea.KeyPgDown}},
		{"pgup", tea.KeyMsg{Type: tea.KeyPgUp}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m2, _ := m.Update(tc.msg)
			assert.Equalf(t, 0, m2.offset, "%s short: offset", tc.name)
		})
	}
}

// T-132: scrolloff clamps to floor(height/2) so viewport movement stays
// possible even when configured value exceeds the viewport.
func TestScrollModel_ScrolloffClampedToHalfViewport(t *testing.T) {
	m := makeScroll(100, 6).WithScrolloff(10)
	// effective = min(10, 6/2) = 3
	// cursor must reach row 6-1-3 = 2 before offset shifts.
	for i := 0; i < 2; i++ {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	}
	assert.Equal(t, 2, m.cursor, "after 2 j: cursor")
	assert.Equal(t, 0, m.offset, "after 2 j: offset")
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	assert.Equal(t, 3, m.cursor, "after 3rd j: cursor")
	assert.Equal(t, 1, m.offset, "after 3rd j: offset")
}

// T-132: clamp at edges — scrolloff yields so cursor can reach line 0 and
// the last line regardless of margin.
func TestScrollModel_ScrolloffYieldsAtDocumentEdges(t *testing.T) {
	m := makeScroll(50, 10).WithScrolloff(3)
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	assert.Equal(t, 49, m2.cursor, "G with scrolloff: cursor")
	m3, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
	assert.Equal(t, 0, m3.cursor, "g with scrolloff: cursor")
	assert.Equal(t, 0, m3.offset, "g with scrolloff: offset")
}

// T-037: R4.4 — stop at top and bottom (offset clamping still holds).
func TestScrollModel_ClampedAtBoundaries(t *testing.T) {
	m := makeScroll(5, 3)

	// Scroll up at top via k: should stay at 0 (cursor at 0).
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	assert.Equal(t, 0, m2.cursor, "top clamp: cursor")
	assert.Equal(t, 0, m2.offset, "top clamp: offset")

	// Scroll to bottom via ScrollDown (offset-only helper — does not touch cursor).
	m3 := m.ScrollDown(100)
	maxOffset := 5 - 3
	assert.Equal(t, maxOffset, m3.offset, "bottom clamp: offset")
	assert.True(t, m3.AtBottom(), "expected AtBottom() = true")
	assert.True(t, m.AtTop(), "expected AtTop() = true for initial model")
}

// View returns only the visible lines.
func TestScrollModel_View(t *testing.T) {
	content := "A\nB\nC\nD\nE"
	m := NewScrollModel(content, 3)
	assert.Equal(t, "A\nB\nC", m.View())
	m = m.ScrollDown(2)
	assert.Equal(t, "C\nD\nE", m.View(), "after scroll")
}

// F-013 visual fix: View must always return exactly height rows so the
// detail pane keeps its allocated outer height when content is shorter
// than the viewport.
func TestScrollModel_View_PadsShortContentToFullHeight(t *testing.T) {
	m := NewScrollModel("A\nB\nC", 10)
	got := m.View()
	rows := strings.Count(got, "\n") + 1
	assert.Equalf(t, 10, rows, "short-content View() rows (got %q)", got)
	assert.Truef(t, strings.HasPrefix(got, "A\nB\nC\n"), "short-content View() must keep content prefix; got %q", got)
}

// Empty content still produces a full-height blank viewport so the
// surrounding pane border draws at full size.
func TestScrollModel_View_EmptyContentReturnsFullHeightBlank(t *testing.T) {
	m := NewScrollModel("", 5)
	got := m.View()
	rows := strings.Count(got, "\n") + 1
	assert.Equal(t, 5, rows, "empty View() rows")
}
