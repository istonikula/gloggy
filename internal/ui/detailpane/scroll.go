package detailpane

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// ScrollModel holds the scrollable detail pane content state.
type ScrollModel struct {
	lines  []string
	offset int // first visible line index
	height int // number of visible lines
}

// NewScrollModel creates a ScrollModel with the given content and visible height.
func NewScrollModel(content string, height int) ScrollModel {
	lines := strings.Split(content, "\n")
	if height < 1 {
		height = 1
	}
	return ScrollModel{lines: lines, height: height}
}

// SetContent replaces the content and resets the scroll position.
func (m ScrollModel) SetContent(content string, height int) ScrollModel {
	lines := strings.Split(content, "\n")
	if height < 1 {
		height = 1
	}
	return ScrollModel{lines: lines, height: height, offset: 0}
}

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

// View renders the visible portion of the content.
func (m ScrollModel) View() string {
	if len(m.lines) == 0 {
		return ""
	}
	end := m.offset + m.height
	if end > len(m.lines) {
		end = len(m.lines)
	}
	return strings.Join(m.lines[m.offset:end], "\n")
}
