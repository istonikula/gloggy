package detailpane

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func makeScroll(lines int, height int) ScrollModel {
	content := make([]string, lines)
	for i := range content {
		content[i] = "line"
	}
	return NewScrollModel(strings.Join(content, "\n"), height)
}

// T-037: R4.1 — j scrolls down
func TestScrollModel_JScrollsDown(t *testing.T) {
	m := makeScroll(20, 5)
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if m2.offset != 1 {
		t.Errorf("offset = %d, want 1", m2.offset)
	}
}

// T-037: R4.2 — k scrolls up
func TestScrollModel_KScrollsUp(t *testing.T) {
	m := makeScroll(20, 5)
	m = m.ScrollDown(5)
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	if m2.offset != 4 {
		t.Errorf("offset = %d, want 4", m2.offset)
	}
}

// T-037: R4.3 — mouse wheel scrolls
func TestScrollModel_MouseWheelScrolls(t *testing.T) {
	m := makeScroll(20, 5)
	m2, _ := m.Update(tea.MouseMsg{Button: tea.MouseButtonWheelDown})
	if m2.offset != 1 {
		t.Errorf("WheelDown: offset = %d, want 1", m2.offset)
	}
	m3, _ := m2.Update(tea.MouseMsg{Button: tea.MouseButtonWheelUp})
	if m3.offset != 0 {
		t.Errorf("WheelUp: offset = %d, want 0", m3.offset)
	}
}

// T-037: R4.4 — stop at top and bottom
func TestScrollModel_ClampedAtBoundaries(t *testing.T) {
	m := makeScroll(5, 3)

	// Scroll up at top: should stay at 0.
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	if m2.offset != 0 {
		t.Errorf("top clamp: offset = %d, want 0", m2.offset)
	}

	// Scroll to bottom and beyond.
	m3 := m.ScrollDown(100)
	maxOffset := 5 - 3 // totalLines - height
	if m3.offset != maxOffset {
		t.Errorf("bottom clamp: offset = %d, want %d", m3.offset, maxOffset)
	}

	// AtBottom should be true.
	if !m3.AtBottom() {
		t.Error("expected AtBottom() = true")
	}
	if !m.AtTop() {
		t.Error("expected AtTop() = true for initial model")
	}
}

// View returns only the visible lines.
func TestScrollModel_View(t *testing.T) {
	content := "A\nB\nC\nD\nE"
	m := NewScrollModel(content, 3)
	if m.View() != "A\nB\nC" {
		t.Errorf("View() = %q, want %q", m.View(), "A\nB\nC")
	}
	m = m.ScrollDown(2)
	if m.View() != "C\nD\nE" {
		t.Errorf("after scroll View() = %q, want %q", m.View(), "C\nD\nE")
	}
}
