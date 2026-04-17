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
	cursorPos    int    // 1-based cursor position in visible set
	focusLabel   string // optional focus label (T-092 may set; first to drop)
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

// WithFocusLabel sets an optional focus label (e.g. "focus: list").
// When the header is too narrow, the focus label is the first thing dropped.
func (m HeaderModel) WithFocusLabel(label string) HeaderModel {
	m.focusLabel = label
	return m
}

// WithWidth updates the render width.
func (m HeaderModel) WithWidth(w int) HeaderModel {
	m.width = w
	return m
}

// View renders the header bar as a single padded line, applying narrow-mode
// degradation per DESIGN.md §4.1: drop focus label → counts → cursor-pos →
// FOLLOW. Source is always kept; if it alone overflows, truncate with `…`.
func (m HeaderModel) View() string {
	source := m.sourceName
	if source == "" {
		source = "stdin"
	}

	// Drop priority order — first entry is dropped first when too narrow.
	type comp struct {
		name string
		text string
	}
	droppable := []comp{
		{"focus_label", ""},
		{"counts", ""},
		{"cursor_pos", ""},
		{"follow_badge", ""},
	}
	if m.focusLabel != "" {
		droppable[0].text = "  " + m.focusLabel
	}
	if m.totalCount > 0 || m.visibleCount > 0 {
		droppable[1].text = fmt.Sprintf("  %d/%d entries", m.visibleCount, m.totalCount)
	}
	if m.cursorPos > 0 {
		droppable[2].text = fmt.Sprintf("  %d/%d", m.cursorPos, m.visibleCount)
	}
	if m.followMode {
		droppable[3].text = " [FOLLOW]"
	}

	build := func(included [4]bool) string {
		// Render order (left-to-right, distinct from drop priority):
		// source + follow_badge + cursor_pos + counts + focus_label
		out := source
		if included[3] {
			out += droppable[3].text
		}
		if included[2] {
			out += droppable[2].text
		}
		if included[1] {
			out += droppable[1].text
		}
		if included[0] {
			out += droppable[0].text
		}
		return out
	}

	included := [4]bool{true, true, true, true}
	for i := 0; i < 4; i++ {
		if !included[i] {
			continue
		}
		if droppable[i].text == "" {
			included[i] = false
		}
	}
	content := build(included)

	if m.width > 0 && lipgloss.Width(content) > m.width {
		// Drop in priority order: 0 (focus_label) first, then 1, 2, 3.
		for i := 0; i < 4; i++ {
			if !included[i] {
				continue
			}
			included[i] = false
			content = build(included)
			if lipgloss.Width(content) <= m.width {
				break
			}
		}
	}

	// If even the source alone overflows, truncate with `…`.
	if m.width > 0 && lipgloss.Width(content) > m.width {
		content = truncateToWidth(source, m.width)
	}

	style := lipgloss.NewStyle().
		Background(m.th.HeaderBg).
		Bold(true).
		Width(m.width)
	return style.Render(content)
}

// truncateToWidth shortens s so that its rendered cell width is <= max.
// When truncated, an ellipsis `…` is appended (consuming 1 cell).
// Uses lipgloss.Width so emoji/CJK/ANSI are accounted for correctly.
func truncateToWidth(s string, max int) string {
	if max <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= max {
		return s
	}
	if max == 1 {
		return "…"
	}
	// Reserve 1 cell for the ellipsis; binary-search the rune prefix that fits.
	runes := []rune(s)
	lo, hi := 0, len(runes)
	for lo < hi {
		mid := (lo + hi + 1) / 2
		if lipgloss.Width(string(runes[:mid]))+1 <= max {
			lo = mid
		} else {
			hi = mid - 1
		}
	}
	return string(runes[:lo]) + "…"
}
