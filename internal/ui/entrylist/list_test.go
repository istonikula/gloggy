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

func makeEntries(n int) []logsource.Entry {
	entries := make([]logsource.Entry, n)
	for i := range entries {
		entries[i] = logsource.Entry{
			IsJSON:     true,
			LineNumber: i + 1,
			Time:       time.Now(),
			Level:      "INFO",
			Msg:        fmt.Sprintf("entry %d", i),
			Raw:        []byte(fmt.Sprintf(`{"msg":"entry %d"}`, i)),
		}
	}
	return entries
}

func defaultListModel(height int) ListModel {
	return NewListModel(theme.GetTheme("tokyo-night"), config.DefaultConfig(), 80, height)
}

// T-029: R6.1 — rendered rows ≤ visible+buffer for 100k entries
func TestListModel_VirtualRendering_100k(t *testing.T) {
	const total = 100_000
	const height = 40
	m := defaultListModel(height).SetEntries(makeEntries(total))

	count := m.RenderedRowCount()
	maxExpected := height + 2*renderBuffer
	if count > maxExpected {
		t.Errorf("rendered %d rows for 100k entries, max expected %d (height+2*buffer)", count, maxExpected)
	}
	if count == 0 {
		t.Error("rendered 0 rows — expected at least some")
	}
}

// T-029: R6.2 — render time < 16ms for large dataset
func TestListModel_RenderSpeed_100k(t *testing.T) {
	const total = 100_000
	const height = 40
	m := defaultListModel(height).SetEntries(makeEntries(total))

	start := time.Now()
	_ = m.View()
	elapsed := time.Since(start)

	if elapsed > 16*time.Millisecond {
		t.Errorf("View() took %v, expected < 16ms", elapsed)
	}
}

// Virtual rendering: scrolled to middle still only renders window+buffer
func TestListModel_VirtualRendering_Scrolled(t *testing.T) {
	const total = 1000
	const height = 20
	m := defaultListModel(height).SetEntries(makeEntries(total))

	// Scroll to middle.
	m.scroll.Cursor = 500
	m.scroll.Offset = 490

	count := m.RenderedRowCount()
	maxExpected := height + 2*renderBuffer
	if count > maxExpected {
		t.Errorf("rendered %d rows when scrolled, max expected %d", count, maxExpected)
	}
}

// No SelectionMsg when cursor doesn't move (already at boundary).
func TestListModel_NoSelectionMsg_AtBoundary(t *testing.T) {
	entries := makeEntries(5)
	m := defaultListModel(10).SetEntries(entries)

	// Already at top — k should not emit.
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	if cmd != nil {
		t.Error("expected nil cmd when cursor cannot move")
	}
}

// T-080: R11 — CursorPosition returns 1-based index
func TestCursorPosition_EmptyList(t *testing.T) {
	m := defaultListModel(10)
	if pos := m.CursorPosition(); pos != 0 {
		t.Fatalf("CursorPosition on empty = %d, want 0", pos)
	}
}

func TestCursorPosition_JK(t *testing.T) {
	m := defaultListModel(10).SetEntries(makeEntries(5))
	if pos := m.CursorPosition(); pos != 1 {
		t.Fatalf("initial CursorPosition = %d, want 1", pos)
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if pos := m.CursorPosition(); pos != 2 {
		t.Fatalf("after j: CursorPosition = %d, want 2", pos)
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	if pos := m.CursorPosition(); pos != 1 {
		t.Fatalf("after k: CursorPosition = %d, want 1", pos)
	}
}

func TestCursorPosition_gG(t *testing.T) {
	m := defaultListModel(10).SetEntries(makeEntries(5))
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	if pos := m.CursorPosition(); pos != 5 {
		t.Fatalf("after G: CursorPosition = %d, want 5", pos)
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
	if pos := m.CursorPosition(); pos != 1 {
		t.Fatalf("after g: CursorPosition = %d, want 1", pos)
	}
}

func TestCursorPosition_AfterFilter(t *testing.T) {
	m := defaultListModel(10).SetEntries(makeEntries(5))
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	// Filter to indices 0,1 → cursor clamps
	m = m.SetFilter([]int{0, 1})
	if pos := m.CursorPosition(); pos != 2 {
		t.Fatalf("after filter: CursorPosition = %d, want 2", pos)
	}
}
