// Package filter provides the filter panel overlay UI component.
package filter

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/istonikula/gloggy/internal/filter"
	"github.com/istonikula/gloggy/internal/theme"
)

// FilterChangedMsg is emitted whenever the filter set changes (toggle, delete).
// The parent model should use FilterSet to recompute the filtered index.
type FilterChangedMsg struct {
	FilterSet *filter.FilterSet
}

// OpenPromptMsg is emitted when the panel wants the parent to open the
// shared filter prompt. Mode=Add uses OpenBlank; Mode=Edit uses OpenEdit.
type OpenPromptMsg struct {
	IsEdit  bool
	Filter  filter.Filter // set when IsEdit=true
	FilterID int          // set when IsEdit=true
}

// Model is the Bubble Tea model for the filter panel overlay.
type Model struct {
	fs     *filter.FilterSet
	cursor int
	th     theme.Theme
}

// New creates a panel model for the given FilterSet.
// The FilterSet is mutated in place by panel operations.
func New(fs *filter.FilterSet) Model {
	return Model{fs: fs}
}

// WithTheme sets the theme used for V35 row formatting.
func (m Model) WithTheme(th theme.Theme) Model {
	m.th = th
	return m
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
		case "a":
			// Emit OpenPromptMsg so parent opens the shared prompt in blank/add mode.
			return m, func() tea.Msg { return OpenPromptMsg{IsEdit: false} }
		case "e":
			if n > 0 && m.cursor < n {
				filters := m.fs.GetAll()
				ids := m.fs.GetIDs()
				f := filters[m.cursor]
				id := ids[m.cursor]
				return m, func() tea.Msg {
					return OpenPromptMsg{IsEdit: true, Filter: f, FilterID: id}
				}
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

// View renders the filter list as a string per V35 Direction A format.
// Each line: "[x|·] +|−  field|«line»:pattern [·glob|·regex]"
func (m Model) View() string {
	filters := m.fs.GetAll()
	if len(filters) == 0 {
		return "(no filters — press 'a' to add, or click a field in the detail pane)"
	}
	var sb strings.Builder
	for i, f := range filters {
		line := m.formatRow(f, i == m.cursor)
		sb.WriteString(line)
		if i < len(filters)-1 {
			sb.WriteByte('\n')
		}
	}
	return sb.String()
}

// formatRow renders a single filter row per V35 Direction A format.
//
// Layout: [cursor_prefix][enabled][+|-]  [scope]:[pattern] [·syntax_tag]
//   - cursor prefix: "> " when selected, "  " otherwise.
//   - enabled: "[x]" / "[ ]"
//   - mode: "+" (include, LevelInfo token) / "−" (exclude, LevelError token).
//   - scope: field name or "«line»" for whole-line (Field == "").
//   - syntax tag: "·glob" / "·regex" in Dim, omitted for Literal.
//   - disabled row: entire line rendered in Dim token.
func (m Model) formatRow(f filter.Filter, cursor bool) string {
	prefix := "  "
	if cursor {
		prefix = "> "
	}

	enabled := "[ ]"
	if f.Enabled {
		enabled = "[x]"
	}

	scope := f.Field
	if scope == "" {
		scope = "«line»"
	}

	modeGlyph := "+"
	modeColor := m.th.LevelInfo
	if f.Mode == filter.Exclude {
		modeGlyph = "-"
		modeColor = m.th.LevelError
	}

	body := fmt.Sprintf("%s:%s", scope, f.Pattern)

	var synTag string
	switch f.Syntax {
	case filter.Glob:
		synTag = "  ·glob"
	case filter.Regex:
		synTag = "  ·regex"
	}

	if m.th.Dim != "" && !f.Enabled {
		// Disabled: entire row in Dim.
		line := fmt.Sprintf("%s%s %s  %s%s", prefix, enabled, modeGlyph, body, synTag)
		return lipgloss.NewStyle().Foreground(m.th.Dim).Render(line)
	}

	// Enabled: color the mode glyph; rest of line in default fg.
	modeStyled := lipgloss.NewStyle().Foreground(modeColor).Render(modeGlyph)
	synStyled := ""
	if synTag != "" {
		synStyled = lipgloss.NewStyle().Foreground(m.th.Dim).Render(synTag)
	}

	// Build with cursor bg on selected row.
	if cursor && m.th.CursorHighlight != "" {
		full := fmt.Sprintf("%s%s %s  %s%s", prefix, enabled, modeGlyph, body, synTag)
		return lipgloss.NewStyle().Background(m.th.CursorHighlight).Render(full)
	}

	return fmt.Sprintf("%s%s %s  %s%s", prefix, enabled, modeStyled, body, synStyled)
}
