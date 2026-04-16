package appshell

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/istonikula/gloggy/internal/theme"
)

// HeaderModel renders the top bar showing source name, follow mode, and entry counts.
type HeaderModel struct {
	sourceName   string // file path or "stdin"
	followMode   bool   // tail/follow active
	totalCount   int
	visibleCount int
	cursorPos    int // 1-based cursor position in visible set
	th           theme.Theme
	width        int
}

// NewHeaderModel creates a HeaderModel.
func NewHeaderModel(th theme.Theme, width int) HeaderModel {
	return HeaderModel{th: th, width: width}
}

// WithSource sets the source name.
func (m HeaderModel) WithSource(name string) HeaderModel {
	m.sourceName = name
	return m
}

// WithFollow sets the tail/follow mode flag.
func (m HeaderModel) WithFollow(follow bool) HeaderModel {
	m.followMode = follow
	return m
}

// WithCounts updates the entry counts.
func (m HeaderModel) WithCounts(total, visible int) HeaderModel {
	m.totalCount = total
	m.visibleCount = visible
	return m
}

// WithCursorPos sets the 1-based cursor position in the visible list.
func (m HeaderModel) WithCursorPos(pos int) HeaderModel {
	m.cursorPos = pos
	return m
}

// WithWidth updates the render width.
func (m HeaderModel) WithWidth(w int) HeaderModel {
	m.width = w
	return m
}

// View renders the header bar as a single padded line.
func (m HeaderModel) View() string {
	source := m.sourceName
	if source == "" {
		source = "stdin"
	}

	followBadge := ""
	if m.followMode {
		followBadge = " [FOLLOW]"
	}

	counts := fmt.Sprintf("  %d/%d  %d/%d entries", m.cursorPos, m.visibleCount, m.visibleCount, m.totalCount)

	content := source + followBadge + counts

	style := lipgloss.NewStyle().
		Background(m.th.HeaderBg).
		Bold(true).
		Width(m.width)
	return style.Render(content)
}
