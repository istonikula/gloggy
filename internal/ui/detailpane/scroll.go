package detailpane

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// ScrollModel holds the scrollable detail pane content state.
// T-131 (F-026): cursor is the 0-indexed document line currently under the
// cursor-tracking viewport. It exists whenever the pane is open regardless
// of focus; the View renders it with CursorHighlight bg.
// T-132 (F-026): navigation operates on cursor first, then followCursor
// adjusts offset so cursor stays ≥ scrolloff rows from viewport edges
// (nvim-style). Document edges yield — cursor can reach line 0 / last line.
type ScrollModel struct {
	lines     []string
	offset    int // first visible line index
	height    int // number of visible lines
	cursor    int // 0-indexed document line; always >= 0
	scrolloff int // context rows kept around cursor (clamped to [0, height/2] at use)
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
// Preserves scrolloff so config-driven value survives re-renders.
func (m ScrollModel) SetContent(content string, height int) ScrollModel {
	lines := strings.Split(content, "\n")
	if height < 1 {
		height = 1
	}
	return ScrollModel{lines: lines, height: height, offset: 0, cursor: 0, scrolloff: m.scrolloff}
}

// Cursor returns the current cursor document line (0-indexed).
func (m ScrollModel) Cursor() int { return m.cursor }

// Offset returns the current viewport offset (first visible line).
func (m ScrollModel) Offset() int { return m.offset }

// LineCount returns the total number of content lines.
func (m ScrollModel) LineCount() int { return len(m.lines) }

// Scrolloff returns the configured scrolloff margin.
func (m ScrollModel) Scrolloff() int { return m.scrolloff }

// WithScrolloff sets the scrolloff margin (T-132, F-026). Negative values
// clamp to 0. The effective margin is further clamped to floor(viewport/2)
// at use time so scrolloff never exceeds half the viewport.
func (m ScrollModel) WithScrolloff(n int) ScrollModel {
	if n < 0 {
		n = 0
	}
	m.scrolloff = n
	return m
}

// effectiveScrolloff returns the scrolloff clamped to [0, floor(height/2)]
// so it can never exceed half the viewport — keeps cursor movement possible
// in small viewports.
func (m ScrollModel) effectiveScrolloff() int {
	max := m.height / 2
	if m.scrolloff > max {
		return max
	}
	return m.scrolloff
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

// clampCursor keeps cursor in [0, len(lines)-1].
func (m *ScrollModel) clampCursor() {
	if m.cursor < 0 {
		m.cursor = 0
	}
	if len(m.lines) == 0 {
		m.cursor = 0
		return
	}
	last := len(m.lines) - 1
	if m.cursor > last {
		m.cursor = last
	}
}

// followCursor adjusts offset so cursor stays ≥ scrolloff rows from viewport
// edges. Document edges yield — cursor can reach the very first or last row.
// T-132 (F-026).
func (m *ScrollModel) followCursor() {
	so := m.effectiveScrolloff()
	top := m.offset + so
	bottom := m.offset + m.height - 1 - so
	if m.cursor < top {
		m.offset = m.cursor - so
	} else if m.cursor > bottom {
		m.offset = m.cursor - m.height + 1 + so
	}
	m.clamp()
}

// MoveCursor shifts cursor by delta (positive = down, negative = up),
// clamps to document bounds, then follows with scrolloff.
func (m ScrollModel) MoveCursor(delta int) ScrollModel {
	m.cursor += delta
	m.clampCursor()
	m.followCursor()
	return m
}

// SetCursor sets cursor to an absolute document line index and follows.
// Clamped to document bounds.
func (m ScrollModel) SetCursor(idx int) ScrollModel {
	m.cursor = idx
	m.clampCursor()
	m.followCursor()
	return m
}

// Update handles tea.KeyMsg and tea.MouseMsg.
// T-132 (F-026): keyboard navigation operates on cursor first; offset
// follows via scrolloff. T-133 handles mouse wheel (offset-first with
// cursor drag at scrolloff edges).
func (m ScrollModel) Update(msg tea.Msg) (ScrollModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			m = m.MoveCursor(1)
		case "k", "up":
			m = m.MoveCursor(-1)
		case "g", "home":
			m = m.SetCursor(0)
		case "G", "end":
			if len(m.lines) > 0 {
				m = m.SetCursor(len(m.lines) - 1)
			}
		case "pgdown", "ctrl+d", " ":
			step := m.height - 1
			if step < 1 {
				step = 1
			}
			m = m.MoveCursor(step)
		case "pgup", "ctrl+u", "b":
			step := m.height - 1
			if step < 1 {
				step = 1
			}
			m = m.MoveCursor(-step)
		}
	case tea.MouseMsg:
		switch msg.Button {
		case tea.MouseButtonWheelDown:
			m = m.wheelDown(1)
		case tea.MouseButtonWheelUp:
			m = m.wheelUp(1)
		}
	}
	return m, nil
}

// wheelDown scrolls the viewport down by n; if the cursor would enter the
// top-scrolloff margin (i.e. cursor < offset + scrolloff), drag cursor along
// so it sits exactly at the scrolloff-th visible row. T-133 (F-026).
func (m ScrollModel) wheelDown(n int) ScrollModel {
	m.offset += n
	m.clamp()
	so := m.effectiveScrolloff()
	minCursor := m.offset + so
	if m.cursor < minCursor {
		m.cursor = minCursor
		m.clampCursor()
	}
	return m
}

// wheelUp scrolls the viewport up by n; if the cursor would leave the
// bottom-scrolloff margin (i.e. cursor > offset + viewport - 1 - scrolloff),
// drag cursor along. T-133 (F-026).
func (m ScrollModel) wheelUp(n int) ScrollModel {
	m.offset -= n
	m.clamp()
	so := m.effectiveScrolloff()
	maxCursor := m.offset + m.height - 1 - so
	if m.cursor > maxCursor {
		m.cursor = maxCursor
		m.clampCursor()
	}
	return m
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
