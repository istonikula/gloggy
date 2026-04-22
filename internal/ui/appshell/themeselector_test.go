package appshell

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/istonikula/gloggy/internal/theme"
)

// T17 / V29: closed selector renders empty and Update emits no action
// (nothing to preview / commit / revert when overlay isn't open).
func TestThemeSelector_Closed_ViewEmpty(t *testing.T) {
	m := NewThemeSelectorModel()
	assert.Emptyf(t, m.View(), "closed selector view must be empty")
	_, action := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	assert.Equalf(t, ThemeSelNone, action, "closed selector must emit no action")
}

// T17 / V29: Open seeds PreOpen + highlights the current theme.
func TestThemeSelector_Open_HighlightsCurrent(t *testing.T) {
	m := NewThemeSelectorModel().Open("catppuccin-mocha")
	assert.Truef(t, m.IsOpen(), "selector should be open")
	assert.Equalf(t, "catppuccin-mocha", m.Highlighted(), "highlight should match current theme")
	assert.Equalf(t, "catppuccin-mocha", m.PreOpen(), "PreOpen should capture the open-time theme")
}

// T17 / V29: unknown theme name falls back to the first entry so Esc still
// has a non-empty PreOpen to restore.
func TestThemeSelector_Open_UnknownThemeFallsBackToFirst(t *testing.T) {
	m := NewThemeSelectorModel().Open("not-a-real-theme")
	require.NotEmpty(t, m.Themes())
	assert.Equalf(t, m.Themes()[0], m.Highlighted(),
		"unknown current theme should highlight the first entry")
	assert.Equalf(t, "not-a-real-theme", m.PreOpen(),
		"PreOpen must preserve whatever name was passed, even unknown")
}

// T17 / V29: j / down moves highlight forward; clamped at last row.
func TestThemeSelector_Navigate_Down(t *testing.T) {
	m := NewThemeSelectorModel().Open(theme.BuiltinNames()[0])
	names := m.Themes()
	for i := 1; i < len(names); i++ {
		var action ThemeSelectorAction
		m, action = m.Update(tea.KeyMsg{Type: tea.KeyDown})
		assert.Equalf(t, ThemeSelPreview, action, "step %d: down must emit Preview", i)
		assert.Equalf(t, names[i], m.Highlighted(), "step %d: highlight should advance", i)
	}
	// Clamp at last row — further down is a no-op, no Preview emitted.
	_, action := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equalf(t, ThemeSelNone, action,
		"down at last row must be a no-op (no Preview, no wrap)")
}

// T17 / V29: j is an alias for down (vim-style nav).
func TestThemeSelector_Navigate_Down_JKey(t *testing.T) {
	m := NewThemeSelectorModel().Open(theme.BuiltinNames()[0])
	m2, action := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	assert.Equalf(t, ThemeSelPreview, action, "j must emit Preview like down")
	assert.Equalf(t, theme.BuiltinNames()[1], m2.Highlighted(),
		"j should advance the highlight")
}

// T17 / V29: k / up moves highlight backward; clamped at first row.
func TestThemeSelector_Navigate_Up(t *testing.T) {
	names := theme.BuiltinNames()
	m := NewThemeSelectorModel().Open(names[len(names)-1])
	for i := len(names) - 2; i >= 0; i-- {
		var action ThemeSelectorAction
		m, action = m.Update(tea.KeyMsg{Type: tea.KeyUp})
		assert.Equalf(t, ThemeSelPreview, action, "step %d: up must emit Preview", i)
		assert.Equalf(t, names[i], m.Highlighted(), "step %d: highlight should move up", i)
	}
	// Clamp at first row.
	_, action := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	assert.Equalf(t, ThemeSelNone, action,
		"up at first row must be a no-op (no Preview, no wrap)")
}

// T17 / V29: Enter emits Commit and closes the overlay.
func TestThemeSelector_Enter_Commits(t *testing.T) {
	m := NewThemeSelectorModel().Open("tokyo-night")
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	highlighted := m.Highlighted()
	m2, action := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Equalf(t, ThemeSelCommit, action, "Enter must emit Commit")
	assert.Falsef(t, m2.IsOpen(), "Enter must close the overlay")
	assert.Equalf(t, highlighted, m2.Highlighted(),
		"Commit preserves the Highlighted value so the caller can persist it")
}

// T17 / V29: Esc emits Revert and closes. Caller is responsible for
// restoring the pre-open theme using PreOpen().
func TestThemeSelector_Esc_Reverts(t *testing.T) {
	m := NewThemeSelectorModel().Open("tokyo-night")
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m2, action := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	assert.Equalf(t, ThemeSelRevert, action, "Esc must emit Revert")
	assert.Falsef(t, m2.IsOpen(), "Esc must close the overlay")
	assert.Equalf(t, "tokyo-night", m2.PreOpen(),
		"PreOpen must remain the open-time theme after Esc")
}

// T17 / V28 / V29: the selector is a full-screen replacement view; it MUST
// NOT emit `\t`. Tab chars leak prior-frame bytes through bubbletea's
// line-diff renderer.
func TestThemeSelector_View_NoTabs(t *testing.T) {
	m := NewThemeSelectorModel().Open("tokyo-night")
	v := m.View()
	assert.NotContainsf(t, v, "\t",
		"theme selector View() must use space padding, not \\t (V28)")
}

// T17 / V29: View lists every bundled theme exactly once.
func TestThemeSelector_View_ListsAllBuiltinThemes(t *testing.T) {
	m := NewThemeSelectorModel().Open("tokyo-night")
	v := m.View()
	for _, name := range theme.BuiltinNames() {
		assert.Containsf(t, v, name,
			"View() should list theme %q; got:\n%s", name, v)
	}
	// The highlighted row uses a "> " prefix; non-highlighted rows use
	// "  " padding. Verify the current highlight is rendered distinctly.
	assert.Truef(t, strings.Contains(v, "> tokyo-night"),
		"highlighted theme should be rendered with a > marker; got:\n%s", v)
}

// T17 / V29 (b): navigating changes which theme's tokens the View carries.
// We assert the selector's own View swaps the "> " marker onto a different
// theme name after a down press — the whole-TUI repaint happens in the app
// model via applyTheme and is covered by app-level tests.
func TestThemeSelector_Navigate_ChangesHighlightedRow(t *testing.T) {
	m := NewThemeSelectorModel().Open("tokyo-night")
	before := m.View()
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	after := m2.View()
	assert.NotEqualf(t, before, after,
		"View should change after navigation so the caller knows to repaint")
	assert.Containsf(t, after, "> "+theme.BuiltinNames()[1],
		"the new highlight row should carry the > marker")
}
