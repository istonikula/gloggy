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

// Update handles key events while the overlay is open: Esc closes it, all
// other keys are consumed (not forwarded). Callers MUST only invoke Update
// when IsOpen() is true — opening via '?' is the caller's responsibility
// (see V14: gated on no pane-search in input mode).
// Returns (model, shouldForward) — while open, shouldForward is always false.
func (m HelpOverlayModel) Update(msg tea.Msg) (HelpOverlayModel, bool) {
	if !m.open {
		return m, true
	}
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.String() == "esc" {
			m = m.Close()
		}
	}
	// All messages consumed while overlay is open — do not forward.
	return m, false
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
