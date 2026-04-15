package appshell

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// HelpOverlayModel shows a full-screen keybinding reference.
// While open, it intercepts all key events so other components do not process them.
type HelpOverlayModel struct {
	open     bool
	registry KeybindingRegistry
}

// NewHelpOverlayModel creates a HelpOverlayModel.
func NewHelpOverlayModel() HelpOverlayModel {
	return HelpOverlayModel{registry: DefaultKeybindings()}
}

// IsOpen returns true when the overlay is visible.
func (m HelpOverlayModel) IsOpen() bool { return m.open }

// Open activates the overlay.
func (m HelpOverlayModel) Open() HelpOverlayModel {
	m.open = true
	return m
}

// Close dismisses the overlay.
func (m HelpOverlayModel) Close() HelpOverlayModel {
	m.open = false
	return m
}

// Update handles key events. When the overlay is open, only Esc closes it; all other
// keys are consumed (not forwarded). When closed, '?' opens it.
// Returns (model, shouldForward) — callers should only forward the message to other
// components when shouldForward is true.
func (m HelpOverlayModel) Update(msg tea.Msg) (HelpOverlayModel, bool) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, !m.open // forward non-key msgs unless overlay intercepts
	}
	if m.open {
		if keyMsg.String() == "esc" {
			m = m.Close()
		}
		// All keys consumed while overlay is open — do not forward.
		return m, false
	}
	if keyMsg.String() == "?" {
		m = m.Open()
		return m, false
	}
	return m, true
}

// View renders the help overlay, or empty string when closed.
func (m HelpOverlayModel) View() string {
	if !m.open {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("Keybindings\n")
	sb.WriteString(strings.Repeat("─", 40))
	sb.WriteByte('\n')
	for _, domain := range Domains() {
		bindings := m.registry[domain]
		if len(bindings) == 0 {
			continue
		}
		sb.WriteString("\n[")
		sb.WriteString(string(domain))
		sb.WriteString("]\n")
		for _, kb := range bindings {
			sb.WriteString("  ")
			sb.WriteString(kb.Key)
			sb.WriteString("\t")
			sb.WriteString(kb.Desc)
			sb.WriteByte('\n')
		}
	}
	sb.WriteString("\nEsc  Close this overlay\n")
	return sb.String()
}
