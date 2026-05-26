package filter

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/istonikula/gloggy/internal/filter"
	"github.com/istonikula/gloggy/internal/theme"
)

// FilterConfirmedMsg is emitted when the user confirms a filter prompt.
type FilterConfirmedMsg struct {
	FilterID  int
	FilterSet *filter.FilterSet
	IsNew     bool // true = Add, false = Edit (Update)
}

// FilterCancelledMsg is emitted when the user cancels the filter prompt.
type FilterCancelledMsg struct{}

// FilterRejectedMsg is emitted when validation fails at commit time.
// The Reason is a short human-readable message suitable for display in the
// keyhints notice (V15-pattern: never silent).
type FilterRejectedMsg struct {
	Reason string
}

// row indices within the multi-row form.
const (
	rowField   = 0
	rowPattern = 1
	rowSyntax  = 2
	rowMode    = 3
	numRows    = 4
)

// PromptModel is a multi-row Add/Edit form for filter creation and editing.
//
// Rows: Field (text input), Pattern (text input), Syntax (enum cycle),
// Mode (enum cycle).
//
// Navigation:
//
//	Tab / Shift+Tab   cycle row focus (4-way wrap)
//	←/→               cycle enum value on Syntax/Mode rows
//	Backspace         delete last char on text-input rows
//	Enter             validate + commit (emits FilterConfirmedMsg or FilterRejectedMsg)
//	Esc               cancel (emits FilterCancelledMsg)
type PromptModel struct {
	active         bool
	focusRow       int
	field          string
	pattern        string
	syntax         filter.Syntax
	mode           filter.Mode
	fs             *filter.FilterSet
	editID         *int // nil = Add; non-nil = Edit (id of filter to Update)
	th             theme.Theme
	rejectedReason string // non-empty after validation failure; shown in View
}

// NewPromptModel creates a PromptModel that operates on the given FilterSet.
func NewPromptModel(fs *filter.FilterSet) PromptModel {
	return PromptModel{fs: fs}
}

// WithTheme sets the theme used for styled rendering (V35 cursor visibility).
func (m PromptModel) WithTheme(th theme.Theme) PromptModel {
	m.th = th
	return m
}

// IsActive returns true when the prompt is showing.
func (m PromptModel) IsActive() bool { return m.active }

// Field returns the current field input text.
func (m PromptModel) Field() string { return m.field }

// Pattern returns the current pattern input text.
func (m PromptModel) Pattern() string { return m.pattern }

// Mode returns the currently selected Mode.
func (m PromptModel) Mode() filter.Mode { return m.mode }

// Syntax returns the currently selected Syntax.
func (m PromptModel) Syntax() filter.Syntax { return m.syntax }

// OpenBlank opens the prompt for a new whole-line filter. The cursor lands
// on the Pattern row (most common edit target).
func (m PromptModel) OpenBlank() PromptModel {
	m.active = true
	m.focusRow = rowPattern
	m.field = ""
	m.pattern = ""
	m.syntax = filter.Literal
	m.mode = filter.Include
	m.editID = nil
	return m
}

// OpenFromPaneClick opens the prompt pre-filled with a field-scoped filter
// from a detail-pane field click. Cursor lands on Pattern row.
func (m PromptModel) OpenFromPaneClick(field, pattern string) PromptModel {
	m.active = true
	m.focusRow = rowPattern
	m.field = field
	m.pattern = pattern
	m.syntax = filter.Literal
	m.mode = filter.Include
	m.editID = nil
	return m
}

// OpenEdit opens the prompt pre-filled with an existing filter for editing.
// The filter's id is remembered so Enter commits via Update instead of Add.
// Cursor lands on Pattern row.
func (m PromptModel) OpenEdit(f filter.Filter, id int) PromptModel {
	m.active = true
	m.focusRow = rowPattern
	m.field = f.Field
	m.pattern = f.Pattern
	m.syntax = f.Syntax
	m.mode = f.Mode
	idCopy := id
	m.editID = &idCopy
	return m
}

// Close deactivates the prompt without committing.
func (m PromptModel) Close() PromptModel {
	m.active = false
	return m
}

// Update handles key events when the prompt is active.
func (m PromptModel) Update(msg tea.Msg) (PromptModel, tea.Cmd) {
	if !m.active {
		return m, nil
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m = m.Close()
			return m, func() tea.Msg { return FilterCancelledMsg{} }

		case "enter":
			f := filter.Filter{
				Field:   m.field,
				Pattern: m.pattern,
				Syntax:  m.syntax,
				Mode:    m.mode,
				Enabled: true,
			}
			if err := m.fs.Validate(f); err != nil {
				reason := err.Error()
				m.rejectedReason = reason
				return m, func() tea.Msg { return FilterRejectedMsg{Reason: reason} }
			}
			m.rejectedReason = ""
			if m.editID == nil {
				id := m.fs.Add(f)
				m = m.Close()
				fs := m.fs
				return m, func() tea.Msg {
					return FilterConfirmedMsg{FilterID: id, FilterSet: fs, IsNew: true}
				}
			}
			id := *m.editID
			m.fs.Update(id, f)
			m = m.Close()
			fs := m.fs
			return m, func() tea.Msg {
				return FilterConfirmedMsg{FilterID: id, FilterSet: fs, IsNew: false}
			}

		case "tab":
			m.rejectedReason = ""
			m.focusRow = (m.focusRow + 1) % numRows
		case "shift+tab":
			m.rejectedReason = ""
			m.focusRow = (m.focusRow + numRows - 1) % numRows

		case "left":
			m.rejectedReason = ""
			switch m.focusRow {
			case rowSyntax:
				m.syntax = cycleSyntaxPrev(m.syntax)
			case rowMode:
				m.mode = cycleModeToggle(m.mode)
			}
		case "right":
			m.rejectedReason = ""
			switch m.focusRow {
			case rowSyntax:
				m.syntax = cycleSyntaxNext(m.syntax)
			case rowMode:
				m.mode = cycleModeToggle(m.mode)
			}

		case "backspace", "ctrl+h":
			m.rejectedReason = ""
			switch m.focusRow {
			case rowField:
				if len(m.field) > 0 {
					m.field = m.field[:len([]rune(m.field))-1]
				}
			case rowPattern:
				if len(m.pattern) > 0 {
					m.pattern = m.pattern[:len([]rune(m.pattern))-1]
				}
			}

		default:
			// Printable chars go to the focused text-input row.
			// V14: reserved globals (q ? T F) become literal chars here.
			runes := msg.Runes
			if len(runes) == 1 {
				m.rejectedReason = ""
				switch m.focusRow {
				case rowField:
					m.field += string(runes)
				case rowPattern:
					m.pattern += string(runes)
				}
			}
		}
	}
	return m, nil
}

// View renders the multi-row form per V35 layout with V28 space-padding.
// Returns "" when the prompt is not active.
func (m PromptModel) View() string {
	if !m.active {
		return ""
	}
	title := "Add filter"
	if m.editID != nil {
		title = "Edit filter"
	}
	ruler := strings.Repeat("─", 40)

	var sb strings.Builder
	sb.WriteString(title + "\n")
	sb.WriteString(ruler + "\n\n")
	sb.WriteString(m.renderRow(rowField, "Field:  ", m.field, ""))
	sb.WriteByte('\n')
	sb.WriteString(m.renderRow(rowPattern, "Pattern:", m.pattern, ""))
	sb.WriteByte('\n')
	sb.WriteString(m.renderSyntaxRow())
	sb.WriteByte('\n')
	sb.WriteString(m.renderModeRow())
	sb.WriteString("\n\n")
	if m.rejectedReason != "" {
		sb.WriteString("  ! " + m.rejectedReason + "\n")
	}
	sb.WriteString(ruler + "\n")
	sb.WriteString(m.footerHints())
	return sb.String()
}

// renderRow renders a text-input row (Field or Pattern).
// When focused, prepends "> " and appends "█" caret; unfocused uses "  ".
func (m PromptModel) renderRow(row int, label, value, _ string) string {
	focused := m.focusRow == row
	prefix := "  "
	if focused {
		prefix = "> "
	}
	text := prefix + label + "  " + value
	if focused {
		text += "█"
	}
	if focused && m.th.CursorHighlight != "" {
		return lipgloss.NewStyle().Background(m.th.CursorHighlight).Render(text)
	}
	return text
}

// renderSyntaxRow renders the Syntax enum-cycle row.
func (m PromptModel) renderSyntaxRow() string {
	focused := m.focusRow == rowSyntax
	prefix := "  "
	if focused {
		prefix = "> "
	}
	opts := []filter.Syntax{filter.Literal, filter.Glob, filter.Regex}
	var parts []string
	for _, s := range opts {
		label := s.String()
		if s == m.syntax {
			label = "[" + label + "]"
		}
		if !focused && s != m.syntax {
			label = lipgloss.NewStyle().Foreground(m.th.Dim).Render(label)
		}
		parts = append(parts, label)
	}
	middle := strings.Join(parts, "  ")
	var text string
	if focused {
		text = prefix + "Syntax:  " + "◂  " + middle + "  ▸"
	} else {
		text = prefix + "Syntax:  " + middle
	}
	if focused && m.th.CursorHighlight != "" {
		return lipgloss.NewStyle().Background(m.th.CursorHighlight).Render(text)
	}
	return text
}

// renderModeRow renders the Mode enum-cycle row.
func (m PromptModel) renderModeRow() string {
	focused := m.focusRow == rowMode
	prefix := "  "
	if focused {
		prefix = "> "
	}
	opts := []filter.Mode{filter.Include, filter.Exclude}
	var parts []string
	for _, mo := range opts {
		label := mo.String()
		if mo == m.mode {
			label = "[" + label + "]"
		}
		if !focused && mo != m.mode {
			label = lipgloss.NewStyle().Foreground(m.th.Dim).Render(label)
		}
		parts = append(parts, label)
	}
	middle := strings.Join(parts, "  ")
	var text string
	if focused {
		text = prefix + "Mode:    " + "◂  " + middle + "  ▸"
	} else {
		text = prefix + "Mode:    " + middle
	}
	if focused && m.th.CursorHighlight != "" {
		return lipgloss.NewStyle().Background(m.th.CursorHighlight).Render(text)
	}
	return text
}

// footerHints returns context-sensitive key hints depending on focused row.
func (m PromptModel) footerHints() string {
	base := "  Tab next field  ·  Enter confirm  ·  Esc cancel"
	if m.focusRow == rowSyntax || m.focusRow == rowMode {
		base = "  Tab next field  ·  ←/→ cycle  ·  Enter confirm  ·  Esc cancel"
	}
	return base
}

func cycleSyntaxNext(s filter.Syntax) filter.Syntax {
	return filter.Syntax((int(s) + 1) % 3)
}

func cycleSyntaxPrev(s filter.Syntax) filter.Syntax {
	return filter.Syntax((int(s) + 2) % 3)
}

func cycleModeToggle(m filter.Mode) filter.Mode {
	if m == filter.Include {
		return filter.Exclude
	}
	return filter.Include
}
