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

// T-124: R4 — g and Home jump to top.
func TestScrollModel_GJumpsToTop(t *testing.T) {
	m := makeScroll(50, 10)
	m = m.ScrollDown(25)
	if m.offset != 25 {
		t.Fatalf("precondition: offset = %d, want 25", m.offset)
	}
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
	if m2.offset != 0 {
		t.Errorf("g: offset = %d, want 0", m2.offset)
	}
	m3, _ := m.Update(tea.KeyMsg{Type: tea.KeyHome})
	if m3.offset != 0 {
		t.Errorf("home: offset = %d, want 0", m3.offset)
	}
}

// T-124: R4 — G and End jump to bottom (last line visible at bottom of viewport).
func TestScrollModel_GCapJumpsToBottom(t *testing.T) {
	m := makeScroll(50, 10)
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	if m2.offset != 40 {
		t.Errorf("G: offset = %d, want 40", m2.offset)
	}
	m3, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnd})
	if m3.offset != 40 {
		t.Errorf("end: offset = %d, want 40", m3.offset)
	}
}

// T-124: R4 — G on content shorter than viewport stays at 0 (clamp).
func TestScrollModel_GOnShortContent(t *testing.T) {
	m := makeScroll(5, 20)
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	if m2.offset != 0 {
		t.Errorf("G short: offset = %d, want 0", m2.offset)
	}
}

// T-124: R4 — PgDn / Ctrl+d / Space scroll down by height-1.
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
			if m2.offset != 9 {
				t.Errorf("%s: offset = %d, want 9", tc.name, m2.offset)
			}
		})
	}
}

// T-124: R4 — PgUp / Ctrl+u / b at top is no-op.
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
			if m2.offset != 0 {
				t.Errorf("%s: offset = %d, want 0", tc.name, m2.offset)
			}
		})
	}
}

// T-124: R4 — PgUp after G returns toward top by height-1.
func TestScrollModel_PageUpAfterEnd(t *testing.T) {
	m := makeScroll(50, 10)
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	if m2.offset != 40 {
		t.Fatalf("precondition: offset = %d, want 40", m2.offset)
	}
	m3, _ := m2.Update(tea.KeyMsg{Type: tea.KeyPgUp})
	if m3.offset != 31 {
		t.Errorf("pgup after G: offset = %d, want 31", m3.offset)
	}
}

// T-124: R4 — PgDn clamps at bottom (no overflow past last line).
func TestScrollModel_PageDownClampsAtBottom(t *testing.T) {
	m := makeScroll(15, 10)
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	if m2.offset != 5 {
		t.Errorf("first pgdown: offset = %d, want 5", m2.offset)
	}
	m3, _ := m2.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	if m3.offset != 5 {
		t.Errorf("second pgdown clamp: offset = %d, want 5", m3.offset)
	}
}

// T-124: R4 — Page keys with viewport height >= content size are a no-op.
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
			if m2.offset != 0 {
				t.Errorf("%s short: offset = %d, want 0", tc.name, m2.offset)
			}
		})
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
