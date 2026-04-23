// Package filter provides the filter panel overlay UI component.
package filter

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/istonikula/gloggy/internal/filter"
)

// FilterChangedMsg is emitted whenever the filter set changes (toggle, delete).
// The parent model should use FilterSet to recompute the filtered index.
type FilterChangedMsg struct {
	FilterSet *filter.FilterSet
}

// Model is the Bubble Tea model for the filter panel overlay.
type Model struct {
	fs     *filter.FilterSet
	cursor int
}

// New creates a panel model for the given FilterSet.
// The FilterSet is mutated in place by panel operations.
func New(fs *filter.FilterSet) Model {
	return Model{fs: fs}
}

// FilterSet returns the current filter set.
func (m Model) FilterSet() *filter.FilterSet { return m.fs }

// Init satisfies tea.Model.
func (m Model) Init() tea.Cmd { return nil }

// Update handles key and mouse events.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	n := len(m.fs.GetAll())

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if m.cursor < n-1 {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case " ":
			if n > 0 && m.cursor < n {
				ids := m.fs.GetIDs()
				id := ids[m.cursor]
				filters := m.fs.GetAll()
				if filters[m.cursor].Enabled {
					m.fs.Disable(id)
				} else {
					m.fs.Enable(id)
				}
				return m, func() tea.Msg { return FilterChangedMsg{FilterSet: m.fs} }
			}
		case "d":
			if n > 0 && m.cursor < n {
				ids := m.fs.GetIDs()
				id := ids[m.cursor]
				m.fs.Remove(id)
				newN := len(m.fs.GetAll())
				if m.cursor >= newN && newN > 0 {
					m.cursor = newN - 1
				}
				return m, func() tea.Msg { return FilterChangedMsg{FilterSet: m.fs} }
			}
		}
	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress {
			row := msg.Y
			if row >= 0 && row < n {
				m.cursor = row
			}
		}
	}
	return m, nil
}

// View renders the filter list as a string.
// Each line: "[x] field:pattern (include|exclude)"
func (m Model) View() string {
	filters := m.fs.GetAll()
	if len(filters) == 0 {
		return "(no filters — open an entry with Enter, click a field in the detail pane to add)"
	}
	var sb strings.Builder
	for i, f := range filters {
		check := "[ ]"
		if f.Enabled {
			check = "[x]"
		}
		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}
		line := fmt.Sprintf("%s%s %s:%s (%s)", cursor, check, f.Field, f.Pattern, f.Mode)
		sb.WriteString(line)
		if i < len(filters)-1 {
			sb.WriteByte('\n')
		}
	}
	return sb.String()
}
