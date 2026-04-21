package appshell

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

// T15 / V14: when the overlay is closed, Update forwards all keys and does
// NOT open the overlay on its own. Opening is the caller's responsibility
// (app layer gates on pane-search-input-mode per V14).
func TestHelpOverlay_Closed_QuestionMark_Forwarded(t *testing.T) {
	m := NewHelpOverlayModel()
	m2, forward := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	assert.Falsef(t, m2.IsOpen(), "? must NOT open the overlay from Update — caller opens")
	assert.Truef(t, forward, "? should be forwarded so the caller can route it")
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

// T16 / V28 / B6: View() MUST NOT emit `\t`. Tab characters leak bytes
// from the previous frame through bubbletea's line-diff renderer.
func TestHelpOverlay_View_NoTabs(t *testing.T) {
	m := NewHelpOverlayModel().Open()
	v := m.View()
	assert.NotContainsf(t, v, "\t",
		"help overlay View() must use space padding, not \\t (V28)")
}

// T16: description column aligns across all bindings — each key is
// padded to max(key width) + gap so descriptions start at the same
// terminal cell regardless of key length.
func TestHelpOverlay_View_DescriptionColumnAligned(t *testing.T) {
	m := NewHelpOverlayModel().Open()
	v := m.View()

	// Pull out lines of the form "  <key><pad><desc>". Domain headers
	// ("[global]" etc.) and section separators are ignored.
	reg := DefaultKeybindings()
	expectCol := 2 + maxKeyWidth(reg) + helpKeyDescGap // leading "  " + key col + gap

	for _, domain := range Domains() {
		for _, kb := range reg[domain] {
			// Find the line that contains this description and check
			// that the description starts at expectCol.
			var found bool
			for _, line := range strings.Split(v, "\n") {
				// Match a line that starts with "  " + key and contains desc.
				if !strings.HasPrefix(line, "  "+kb.Key) {
					continue
				}
				if !strings.Contains(line, kb.Desc) {
					continue
				}
				descStart := strings.Index(line, kb.Desc)
				// Compare display columns, not byte offsets — keys with
				// multi-byte glyphs (e.g. "j/↓") inflate byte indices.
				descCol := lipgloss.Width(line[:descStart])
				assert.Equalf(t, expectCol, descCol,
					"desc %q should start at col %d, got %d (line=%q)",
					kb.Desc, expectCol, descCol, line)
				found = true
				break
			}
			assert.Truef(t, found, "expected to find binding line for key=%q desc=%q", kb.Key, kb.Desc)
		}
	}
}

// maxKeyWidth + helpKeyDescGap are package-private helpers — assert the
// test can see them and they produce non-degenerate values so a future
// rewrite that replaces them silently triggers a compile error first.
func TestHelpOverlay_KeyColumnWidthSanity(t *testing.T) {
	reg := DefaultKeybindings()
	w := maxKeyWidth(reg)
	assert.Greaterf(t, w, 0, "maxKeyWidth should be > 0 for the default registry")
	// "Ctrl+d" is 6 cells — any narrower and lipgloss.Width is broken.
	assert.GreaterOrEqualf(t, w, lipgloss.Width("Ctrl+d"),
		"maxKeyWidth should cover the widest default key (e.g. Ctrl+d)")
}
