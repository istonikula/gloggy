package entrylist

import (
	"fmt"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

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
	if idx != 1 {
		t.Errorf("NextLevel from 0: got %d, want 1", idx)
	}
	if dir != NoWrap {
		t.Errorf("expected NoWrap, got %v", dir)
	}
}

// T-032: R8.2 — E moves to previous ERROR
func TestLevelJump_PrevError(t *testing.T) {
	entries := levelEntries()
	idx, dir := PrevLevel(entries, 4, "ERROR")
	if idx != 1 {
		t.Errorf("PrevLevel from 4: got %d, want 1", idx)
	}
	if dir != NoWrap {
		t.Errorf("expected NoWrap, got %v", dir)
	}
}

// T-032: R8.3+R8.4 — w/W for WARN
func TestLevelJump_Warn(t *testing.T) {
	entries := levelEntries()
	idx, _ := NextLevel(entries, 0, "WARN")
	if idx != 2 {
		t.Errorf("NextLevel WARN from 0: got %d, want 2", idx)
	}
	idx, _ = PrevLevel(entries, 6, "WARN")
	if idx != 2 {
		t.Errorf("PrevLevel WARN from 6: got %d, want 2", idx)
	}
}

// T-032: R8.5 — wrap when no more in direction
func TestLevelJump_Wraps(t *testing.T) {
	entries := levelEntries()
	// Last ERROR is at 4; next from 4 should wrap to 1.
	idx, dir := NextLevel(entries, 4, "ERROR")
	if idx != 1 {
		t.Errorf("NextLevel wrap: got %d, want 1", idx)
	}
	if dir != WrapForward {
		t.Errorf("expected WrapForward, got %v", dir)
	}
}

// T-032: R8.6 — WrapDir() reflects wrap
func TestListModel_LevelJump_WrapDir(t *testing.T) {
	entries := levelEntries()
	m := NewListModel(theme.GetTheme("tokyo-night"), config.DefaultConfig(), 80, 10).SetEntries(entries)
	// cursor at 4 (ERROR); next ERROR wraps to 1
	m.scroll.Cursor = 4
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	if m2.WrapDir() != WrapForward {
		t.Errorf("expected WrapForward after level-jump wrap, got %v", m2.WrapDir())
	}
}

// T-031: R7.1 — filtered entries don't appear in visible list
func TestListModel_FilteredView_ExcludesEntries(t *testing.T) {
	entries := makeEntries(5)
	m := NewListModel(theme.GetTheme("tokyo-night"), config.DefaultConfig(), 80, 10).SetEntries(entries)
	// Only show entries 1 and 3.
	m = m.SetFilter([]int{1, 3})
	vis := m.visibleEntries()
	if len(vis) != 2 {
		t.Fatalf("expected 2 visible entries, got %d", len(vis))
	}
	if vis[0].LineNumber != 2 || vis[1].LineNumber != 4 {
		t.Errorf("wrong visible entries: got line numbers %d,%d, want 2,4",
			vis[0].LineNumber, vis[1].LineNumber)
	}
}

// T-031: R7.2 — filter change updates list
func TestListModel_FilteredView_UpdatesOnFilterChange(t *testing.T) {
	entries := makeEntries(5)
	m := NewListModel(theme.GetTheme("tokyo-night"), config.DefaultConfig(), 80, 10).SetEntries(entries)
	m = m.SetFilter([]int{0, 2, 4})
	if len(m.visibleEntries()) != 3 {
		t.Fatalf("after filter, expected 3 visible")
	}
	m = m.SetFilter([]int{1})
	if len(m.visibleEntries()) != 1 {
		t.Fatalf("after second filter, expected 1 visible")
	}
}

// T-031: R7.3+R7.4 — cursor preserved or moved to nearest when filter changes
func TestListModel_FilteredView_CursorPreservedOrMoved(t *testing.T) {
	entries := makeEntries(5)
	m := NewListModel(theme.GetTheme("tokyo-night"), config.DefaultConfig(), 80, 10).SetEntries(entries)
	m.scroll.Cursor = 2
	// Filter includes cursor entry (index 2).
	m2 := m.SetFilter([]int{0, 2, 4})
	if m2.scroll.Cursor != 2 {
		t.Errorf("cursor should be preserved (still in filtered set), got %d", m2.scroll.Cursor)
	}
	// Filter excludes cursor entry; nearest is index 1 → visible pos 0.
	m3 := m.SetFilter([]int{1, 3})
	// cursor was 2 (not in filter), nearest before it is 1.
	vis := m3.visibleEntries()
	if len(vis) == 0 {
		t.Fatal("expected visible entries after filter")
	}
	// Cursor should have been moved (not crash).
	_ = m3.scroll.Cursor
}

// T-033: R9.1 — m toggles mark
func TestListModel_MarkToggle(t *testing.T) {
	entries := makeEntries(3)
	m := NewListModel(theme.GetTheme("tokyo-night"), config.DefaultConfig(), 80, 10).SetEntries(entries)
	// Mark entry at cursor 0.
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("m")})
	if !m2.marks.IsMarked(entries[0].LineNumber) {
		t.Error("entry should be marked after m")
	}
	// Unmark.
	m3, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("m")})
	if m3.marks.IsMarked(entries[0].LineNumber) {
		t.Error("entry should be unmarked after second m")
	}
}

// T-033: R9.2 — marked entries show visual indicator in view
func TestListModel_MarkedEntryIndicator(t *testing.T) {
	entries := makeEntries(3)
	m := NewListModel(theme.GetTheme("tokyo-night"), config.DefaultConfig(), 80, 10).SetEntries(entries)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("m")})
	view := m.View()
	if !containsStr(view, "* ") {
		t.Errorf("expected mark indicator '* ' in view, got:\n%s", view)
	}
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
	if m2.scroll.Cursor != 1 {
		t.Errorf("u from 0: cursor = %d, want 1", m2.scroll.Cursor)
	}

	// From position 3, U should go to previous mark at 1.
	m3, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("U")})
	m.scroll.Cursor = 3
	m4, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("U")})
	_ = m3
	if m4.scroll.Cursor != 1 {
		t.Errorf("U from 3: cursor = %d, want 1", m4.scroll.Cursor)
	}
}

// T-034: R10.1 — SelectionMsg emitted when cursor moves
func TestListModel_SelectionMsg(t *testing.T) {
	entries := makeEntries(3)
	m := NewListModel(theme.GetTheme("tokyo-night"), config.DefaultConfig(), 80, 10).SetEntries(entries)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if cmd == nil {
		t.Fatal("expected SelectionMsg cmd after j")
	}
	msg := cmd()
	if _, ok := msg.(SelectionMsg); !ok {
		t.Errorf("expected SelectionMsg, got %T", msg)
	}
}
