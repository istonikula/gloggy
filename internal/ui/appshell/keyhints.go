package appshell

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/istonikula/gloggy/internal/theme"
)

// KeyHintBarModel renders the bottom status bar with context-sensitive keybindings.
// It updates immediately when the focus target changes.
type KeyHintBarModel struct {
	focus    FocusTarget
	registry KeybindingRegistry
	th       theme.Theme
	width    int
	paneOpen bool // when true, the detail pane is visible → focus label is shown
}

// NewKeyHintBarModel creates a KeyHintBarModel.
func NewKeyHintBarModel(th theme.Theme, width int) KeyHintBarModel {
	return KeyHintBarModel{
		focus:    FocusEntryList,
		registry: DefaultKeybindings(),
		th:       th,
		width:    width,
	}
}

// WithFocus updates the focused component, changing which keybindings are shown.
func (m KeyHintBarModel) WithFocus(f FocusTarget) KeyHintBarModel {
	m.focus = f
	return m
}

// WithWidth updates the render width.
func (m KeyHintBarModel) WithWidth(w int) KeyHintBarModel {
	m.width = w
	return m
}

// WithPaneOpen signals whether the detail pane is currently visible. When
// the detail pane is open (more than one pane visible), the status bar
// appends a right-aligned focus label (T-092).
func (m KeyHintBarModel) WithPaneOpen(open bool) KeyHintBarModel {
	m.paneOpen = open
	return m
}

// focusLabelText returns the label string for the current focus ("focus: X")
// or empty when the label should be omitted (single-pane state).
func (m KeyHintBarModel) focusLabelText() string {
	if !m.paneOpen {
		return ""
	}
	switch m.focus {
	case FocusDetailPane:
		return "focus: details"
	case FocusFilterPanel:
		return "focus: filter"
	default:
		return "focus: list"
	}
}

// View renders the key-hint bar for the current focus.
func (m KeyHintBarModel) View() string {
	domain := focusDomain(m.focus)
	bindings := m.registry[domain]

	var parts []string
	for _, kb := range bindings {
		parts = append(parts, kb.Key+": "+kb.Desc)
	}

	hintsStyle := lipgloss.NewStyle().Foreground(m.th.Dim)
	hints := strings.Join(parts, "  ")

	label := m.focusLabelText()
	if label == "" {
		// Single-pane state: status bar is hints only, truncated to 1 row.
		return hintsStyle.MaxWidth(m.width).Render(hints)
	}

	labelStyle := lipgloss.NewStyle().Bold(true).Foreground(m.th.FocusBorder)
	renderedLabel := labelStyle.Render(label)
	labelWidth := lipgloss.Width(renderedLabel)

	// Room for the hints on the left, a 2-cell gutter, then the label on
	// the right. When the total width is too narrow, truncate hints first.
	hintsBudget := m.width - labelWidth - 2
	if hintsBudget < 0 {
		hintsBudget = 0
	}
	hintsTruncated := hintsStyle.MaxWidth(hintsBudget).Render(hints)
	hintsRendered := lipgloss.Width(hintsTruncated)

	gap := m.width - hintsRendered - labelWidth
	if gap < 1 {
		gap = 1
	}
	return hintsTruncated + strings.Repeat(" ", gap) + renderedLabel
}

func focusDomain(f FocusTarget) Domain {
	switch f {
	case FocusDetailPane:
		return DomainDetailPane
	case FocusFilterPanel:
		return DomainFilterPanel
	default:
		return DomainEntryList
	}
}
