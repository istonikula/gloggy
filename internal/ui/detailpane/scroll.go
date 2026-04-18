package detailpane

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// ScrollModel holds the scrollable detail pane content state.
// T-131 (F-026): cursor is the 0-indexed document line currently under the
// cursor-tracking viewport. It exists whenever the pane is open regardless
// of focus; the View renders it with CursorHighlight bg.
type ScrollModel struct {
	lines  []string
	offset int // first visible line index
	height int // number of visible lines
	cursor int // 0-indexed document line; always >= 0
}

// NewScrollModel creates a ScrollModel with the given content and visible height.
func NewScrollModel(content string, height int) ScrollModel {
	lines := strings.Split(content, "\n")
	if height < 1 {
		height = 1
	}
	return ScrollModel{lines: lines, height: height, cursor: 0}
}

// SetContent replaces the content and resets the scroll position and cursor.
func (m ScrollModel) SetContent(content string, height int) ScrollModel {
	lines := strings.Split(content, "\n")
	if height < 1 {
		height = 1
	}
	return ScrollModel{lines: lines, height: height, offset: 0, cursor: 0}
}

// Cursor returns the current cursor document line (0-indexed).
func (m ScrollModel) Cursor() int { return m.cursor }

// Offset returns the current viewport offset (first visible line).
func (m ScrollModel) Offset() int { return m.offset }

// LineCount returns the total number of content lines.
func (m ScrollModel) LineCount() int { return len(m.lines) }

// ScrollDown moves down by n lines, clamped at the bottom.
func (m ScrollModel) ScrollDown(n int) ScrollModel {
	m.offset += n
	m.clamp()
	return m
}

// ScrollUp moves up by n lines, clamped at the top.
func (m ScrollModel) ScrollUp(n int) ScrollModel {
	m.offset -= n
	m.clamp()
	return m
}

func (m *ScrollModel) clamp() {
	if m.offset < 0 {
		m.offset = 0
	}
	max := len(m.lines) - m.height
	if max < 0 {
		max = 0
	}
	if m.offset > max {
		m.offset = max
	}
}

// Clamp returns a copy of m with offset clamped to the current lines/height.
// Call after mutating height or lines externally to keep the viewport valid
// (T-123, F-019).
func (m ScrollModel) Clamp() ScrollModel {
	m.clamp()
	return m
}

// AtTop returns true when already at the top.
func (m ScrollModel) AtTop() bool { return m.offset == 0 }

// AtBottom returns true when already at the bottom.
func (m ScrollModel) AtBottom() bool {
	max := len(m.lines) - m.height
	if max < 0 {
		max = 0
	}
	return m.offset >= max
}

// Update handles tea.KeyMsg and tea.MouseMsg for scrolling.
func (m ScrollModel) Update(msg tea.Msg) (ScrollModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			m = m.ScrollDown(1)
		case "k", "up":
			m = m.ScrollUp(1)
		case "g", "home":
			m.offset = 0
			m.clamp()
		case "G", "end":
			m.offset = len(m.lines) - m.height
			m.clamp()
		case "pgdown", "ctrl+d", " ":
			step := m.height - 1
			if step < 1 {
				step = 1
			}
			m = m.ScrollDown(step)
		case "pgup", "ctrl+u", "b":
			step := m.height - 1
			if step < 1 {
				step = 1
			}
			m = m.ScrollUp(step)
		}
	case tea.MouseMsg:
		switch msg.Button {
		case tea.MouseButtonWheelDown:
			m = m.ScrollDown(1)
		case tea.MouseButtonWheelUp:
			m = m.ScrollUp(1)
		}
	}
	return m, nil
}

// View renders the visible portion of the content. Always returns exactly
// m.height rows — short content is bottom-padded with empty lines so the
// pane keeps its allocated outer height (F-013 visual fix). When lines are
// empty, returns m.height blank rows so the surrounding border still draws
// at full size.
func (m ScrollModel) View() string {
	h := m.height
	if h < 1 {
		h = 1
	}
	if len(m.lines) == 0 {
		return strings.Repeat("\n", h-1)
	}
	end := m.offset + h
	if end > len(m.lines) {
		end = len(m.lines)
	}
	visible := m.lines[m.offset:end]
	if len(visible) >= h {
		return strings.Join(visible, "\n")
	}
	out := strings.Join(visible, "\n")
	out += strings.Repeat("\n", h-len(visible))
	return out
}
