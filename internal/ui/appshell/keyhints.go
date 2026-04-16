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

// View renders the key-hint bar for the current focus.
func (m KeyHintBarModel) View() string {
	domain := focusDomain(m.focus)
	bindings := m.registry[domain]

	var parts []string
	for _, kb := range bindings {
		parts = append(parts, kb.Key+": "+kb.Desc)
	}

	content := strings.Join(parts, "  ")
	// Truncate to exactly 1 row — never wrap. The layout reserves
	// StatusBarHeight=1, so wrapping would overflow into adjacent zones.
	style := lipgloss.NewStyle().
		Foreground(m.th.Dim).
		MaxWidth(m.width)
	return style.Render(content)
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
