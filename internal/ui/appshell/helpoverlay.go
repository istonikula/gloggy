package appshell

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
)

// helpKeyDescGap is the space between the key column and description column.
const helpKeyDescGap = 2

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
	// V28: pad key column with spaces, never `\t`. bubbletea's line-diff
	// renderer does not clear cells that `\t` skips over — they retain
	// bytes from the previous frame, bleeding prior content into the
	// overlay. Compute the max key width across all registered bindings
	// so the description column aligns uniformly.
	keyCol := maxKeyWidth(m.registry) + helpKeyDescGap
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
			pad := keyCol - lipgloss.Width(kb.Key)
			if pad < 1 {
				pad = 1
			}
			sb.WriteString(strings.Repeat(" ", pad))
			sb.WriteString(kb.Desc)
			sb.WriteByte('\n')
		}
	}
	sb.WriteString("\nEsc  Close this overlay\n")
	return sb.String()
}

// maxKeyWidth returns the widest rendered key string across the registry,
// measured in terminal cells via lipgloss.Width so wide/East-Asian glyphs
// and arrow chars count correctly.
func maxKeyWidth(reg KeybindingRegistry) int {
	max := 0
	for _, bindings := range reg {
		for _, kb := range bindings {
			if w := lipgloss.Width(kb.Key); w > max {
				max = w
			}
		}
	}
	return max
}
