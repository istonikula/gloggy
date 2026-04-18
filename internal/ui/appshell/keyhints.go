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
	paneOpen bool   // when true, the detail pane is visible → focus label is shown
	notice   string // transient one-shot notice that replaces the hints (T-091)
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

// WithNotice sets a transient status notice that replaces the key hints
// (T-091). Pass an empty string to clear. The caller is responsible for
// scheduling the clear (typically via a tea.Tick command).
func (m KeyHintBarModel) WithNotice(text string) KeyHintBarModel {
	m.notice = text
	return m
}

// HasNotice reports whether a transient notice is currently displayed.
func (m KeyHintBarModel) HasNotice() bool { return m.notice != "" }

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
	hintsStyle := lipgloss.NewStyle().Foreground(m.th.Dim)
	// Transient notice replaces the hints entirely (T-091).
	if m.notice != "" {
		return hintsStyle.MaxWidth(m.width).Render(m.notice)
	}

	domain := focusDomain(m.focus)
	bindings := m.registry[domain]

	var parts []string
	for _, kb := range bindings {
		// T-121 (app-shell R13): replace the static `/` description
		// when focused on the entry list so the keyhint reflects the
		// actual pane-state scope. Hide `/` entirely when the filter
		// panel is focused — `/` is a literal input there.
		if kb.Key == "/" {
			if m.focus == FocusFilterPanel {
				continue
			}
			if m.focus == FocusEntryList {
				if m.paneOpen {
					parts = append(parts, "/: search pane")
				} else {
					parts = append(parts, "/: search (open entry first)")
				}
				continue
			}
		}
		parts = append(parts, kb.Key+": "+kb.Desc)
	}

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
