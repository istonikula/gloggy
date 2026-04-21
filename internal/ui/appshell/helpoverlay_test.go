package appshell

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

// T-051: R5.1 — ? opens the help overlay.
func TestHelpOverlay_QuestionMark_Opens(t *testing.T) {
	m := NewHelpOverlayModel()
	m2, forward := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	assert.Truef(t, m2.IsOpen(), "? should open the help overlay")
	assert.Falsef(t, forward, "? should not be forwarded to other components")
}

// T-051: R5.2 — overlay lists all keybindings by domain.
func TestHelpOverlay_View_ContainsDomains(t *testing.T) {
	m := NewHelpOverlayModel().Open()
	v := m.View()
	for _, domain := range Domains() {
		assert.Containsf(t, v, string(domain), "expected domain %q in help overlay view", domain)
	}
}

// T-051: R5.3 — Esc closes the overlay.
func TestHelpOverlay_Esc_Closes(t *testing.T) {
	m := NewHelpOverlayModel().Open()
	m2, forward := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	assert.Falsef(t, m2.IsOpen(), "Esc should close the help overlay")
	assert.Falsef(t, forward, "Esc should not be forwarded when overlay is open")
}

// T-051: R5.4 — other keys are intercepted while overlay is open.
func TestHelpOverlay_OtherKeys_Intercepted(t *testing.T) {
	m := NewHelpOverlayModel().Open()
	_, forward := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	assert.Falsef(t, forward, "j should be intercepted while help overlay is open")
}

// Closed overlay view is empty.
func TestHelpOverlay_Closed_ViewEmpty(t *testing.T) {
	m := NewHelpOverlayModel()
	assert.Emptyf(t, m.View(), "expected empty view when overlay is closed")
}
