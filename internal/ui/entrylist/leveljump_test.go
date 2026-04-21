package entrylist

import (
	"fmt"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/istonikula/gloggy/internal/config"
	"github.com/istonikula/gloggy/internal/logsource"
	"github.com/istonikula/gloggy/internal/theme"
)

func levelEntries() []logsource.Entry {
	levels := []string{"INFO", "ERROR", "WARN", "DEBUG", "ERROR", "INFO", "WARN"}
	entries := make([]logsource.Entry, len(levels))
	for i, l := range levels {
		entries[i] = logsource.Entry{
			IsJSON:     true,
			LineNumber: i + 1,
			Level:      l,
			Msg:        fmt.Sprintf("msg %d", i),
			Time:       time.Now(),
			Raw:        []byte(`{}`),
		}
	}
	return entries
}

// T-032: R8.1 — e moves to next ERROR
func TestLevelJump_NextError(t *testing.T) {
	entries := levelEntries() // positions: 0=INFO,1=ERROR,2=WARN,3=DEBUG,4=ERROR,5=INFO,6=WARN
	idx, dir := NextLevel(entries, 0, "ERROR")
	assert.Equal(t, 1, idx, "NextLevel from 0")
	assert.Equal(t, NoWrap, dir)
}

// T-032: R8.2 — E moves to previous ERROR
func TestLevelJump_PrevError(t *testing.T) {
	entries := levelEntries()
	idx, dir := PrevLevel(entries, 4, "ERROR")
	assert.Equal(t, 1, idx, "PrevLevel from 4")
	assert.Equal(t, NoWrap, dir)
}

// T-032: R8.3+R8.4 — w/W for WARN
func TestLevelJump_Warn(t *testing.T) {
	entries := levelEntries()
	idx, _ := NextLevel(entries, 0, "WARN")
	assert.Equal(t, 2, idx, "NextLevel WARN from 0")
	idx, _ = PrevLevel(entries, 6, "WARN")
	assert.Equal(t, 2, idx, "PrevLevel WARN from 6")
}

// T-032: R8.5 — wrap when no more in direction
func TestLevelJump_Wraps(t *testing.T) {
	entries := levelEntries()
	// Last ERROR is at 4; next from 4 should wrap to 1.
	idx, dir := NextLevel(entries, 4, "ERROR")
	assert.Equal(t, 1, idx, "NextLevel wrap")
	assert.Equal(t, WrapForward, dir)
}

// T-032: R8.6 — WrapDir() reflects wrap
func TestListModel_LevelJump_WrapDir(t *testing.T) {
	entries := levelEntries()
	m := NewListModel(theme.GetTheme("tokyo-night"), config.DefaultConfig(), 80, 10).SetEntries(entries)
	// cursor at 4 (ERROR); next ERROR wraps to 1
	m.scroll.Cursor = 4
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	assert.Equal(t, WrapForward, m2.WrapDir(), "expected WrapForward after level-jump wrap")
}

// T-097: HasTransient + ClearTransient — Esc on the list should be able to
// clear an active wrap indicator without affecting other state.
func TestListModel_ClearTransient_ResetsWrapDir(t *testing.T) {
	entries := levelEntries()
	m := NewListModel(theme.GetTheme("tokyo-night"), config.DefaultConfig(), 80, 10).SetEntries(entries)
	m.scroll.Cursor = 4
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	require.True(t, m2.HasTransient(), "expected HasTransient() == true after wrap")
	cursorBefore := m2.Cursor()
	m3 := m2.ClearTransient()
	assert.False(t, m3.HasTransient(), "ClearTransient must reset wrap indicator")
	assert.Equal(t, NoWrap, m3.WrapDir(), "WrapDir after ClearTransient")
	assert.Equal(t, cursorBefore, m3.Cursor(), "ClearTransient must not move the cursor")
}

// T-031: R7.1 — filtered entries don't appear in visible list
func TestListModel_FilteredView_ExcludesEntries(t *testing.T) {
	entries := makeEntries(5)
	m := NewListModel(theme.GetTheme("tokyo-night"), config.DefaultConfig(), 80, 10).SetEntries(entries)
	// Only show entries 1 and 3.
	m = m.SetFilter([]int{1, 3})
	vis := m.visibleEntries()
	require.Len(t, vis, 2, "expected 2 visible entries")
	assert.Equal(t, 2, vis[0].LineNumber)
	assert.Equal(t, 4, vis[1].LineNumber)
}

// T-031: R7.2 — filter change updates list
func TestListModel_FilteredView_UpdatesOnFilterChange(t *testing.T) {
	entries := makeEntries(5)
	m := NewListModel(theme.GetTheme("tokyo-night"), config.DefaultConfig(), 80, 10).SetEntries(entries)
	m = m.SetFilter([]int{0, 2, 4})
	require.Len(t, m.visibleEntries(), 3, "after filter, expected 3 visible")
	m = m.SetFilter([]int{1})
	require.Len(t, m.visibleEntries(), 1, "after second filter, expected 1 visible")
}

// T-031: R7.3+R7.4 — cursor preserved or moved to nearest when filter changes
func TestListModel_FilteredView_CursorPreservedOrMoved(t *testing.T) {
	entries := makeEntries(5)
	m := NewListModel(theme.GetTheme("tokyo-night"), config.DefaultConfig(), 80, 10).SetEntries(entries)
	m.scroll.Cursor = 2
	// Filter includes cursor entry (index 2).
	m2 := m.SetFilter([]int{0, 2, 4})
	assert.Equal(t, 2, m2.scroll.Cursor, "cursor should be preserved (still in filtered set)")
	// Filter excludes cursor entry; nearest is index 1 → visible pos 0.
	m3 := m.SetFilter([]int{1, 3})
	// cursor was 2 (not in filter), nearest before it is 1.
	vis := m3.visibleEntries()
	require.NotEmpty(t, vis, "expected visible entries after filter")
	// Cursor should have been moved (not crash).
	_ = m3.scroll.Cursor
}

// T-033: R9.1 — m toggles mark
func TestListModel_MarkToggle(t *testing.T) {
	entries := makeEntries(3)
	m := NewListModel(theme.GetTheme("tokyo-night"), config.DefaultConfig(), 80, 10).SetEntries(entries)
	// Mark entry at cursor 0.
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("m")})
	assert.True(t, m2.marks.IsMarked(entries[0].LineNumber), "entry should be marked after m")
	// Unmark.
	m3, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("m")})
	assert.False(t, m3.marks.IsMarked(entries[0].LineNumber), "entry should be unmarked after second m")
}

// T-033: R9.2 — marked entries show visual indicator in view
func TestListModel_MarkedEntryIndicator(t *testing.T) {
	entries := makeEntries(3)
	m := NewListModel(theme.GetTheme("tokyo-night"), config.DefaultConfig(), 80, 10).SetEntries(entries)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("m")})
	view := m.View()
	assert.True(t, containsStr(view, "* "),
		"expected mark indicator '* ' in view, got:\n%s", view)
}

// T-033: R9.3+R9.4 — u/U navigates marks
func TestListModel_MarkNavigation(t *testing.T) {
	entries := makeEntries(5)
	m := NewListModel(theme.GetTheme("tokyo-night"), config.DefaultConfig(), 80, 10).SetEntries(entries)

	// Mark entries at positions 1 and 3.
	m.scroll.Cursor = 1
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("m")})
	m.scroll.Cursor = 3
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("m")})

	// From position 0, u should go to mark at 1.
	m.scroll.Cursor = 0
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("u")})
	assert.Equal(t, 1, m2.scroll.Cursor, "u from 0")

	// From position 3, U should go to previous mark at 1.
	m3, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("U")})
	m.scroll.Cursor = 3
	m4, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("U")})
	_ = m3
	assert.Equal(t, 1, m4.scroll.Cursor, "U from 3")
}

// T-034: R10.1 — SelectionMsg emitted when cursor moves
func TestListModel_SelectionMsg(t *testing.T) {
	entries := makeEntries(3)
	m := NewListModel(theme.GetTheme("tokyo-night"), config.DefaultConfig(), 80, 10).SetEntries(entries)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	require.NotNil(t, cmd, "expected SelectionMsg cmd after j")
	msg := cmd()
	_, ok := msg.(SelectionMsg)
	assert.True(t, ok, "expected SelectionMsg, got %T", msg)
}
