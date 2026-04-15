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

// SelectionMsg emitted when cursor moves.
func TestListModel_SelectionMsg(t *testing.T) {
	entries := makeEntries(5)
	m := defaultListModel(10).SetEntries(entries)

	m2, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	_ = m2
	if cmd == nil {
		t.Fatal("expected SelectionMsg cmd after j")
	}
	msg := cmd()
	sel, ok := msg.(SelectionMsg)
	if !ok {
		t.Fatalf("expected SelectionMsg, got %T", msg)
	}
	if sel.Entry.Msg != "entry 1" {
		t.Errorf("SelectionMsg.Entry.Msg = %q, want %q", sel.Entry.Msg, "entry 1")
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
