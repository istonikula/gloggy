package appshell

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/istonikula/gloggy/internal/theme"
)

// ThemeSelectorAction signals the intent emitted by a ThemeSelectorModel.Update
// call so the caller (app.Model) can apply the corresponding effect — preview
// the highlighted theme on nav, persist + close on commit, or restore the
// pre-open theme on revert.
type ThemeSelectorAction int

const (
	// ThemeSelNone — no effect (unrelated key, or arrow press at clamp edge).
	ThemeSelNone ThemeSelectorAction = iota
	// ThemeSelPreview — highlight changed; caller should re-theme the TUI.
	ThemeSelPreview
	// ThemeSelCommit — Enter pressed; caller should persist theme + close.
	ThemeSelCommit
	// ThemeSelRevert — Esc pressed; caller should restore pre-open theme + close.
	ThemeSelRevert
)

// ThemeSelectorModel is the V29 theme-selector overlay. Opens on global `T`
// (gated on pane-search input mode per V14 — handled by the caller). Full-
// screen replacement view à la HelpOverlayModel. Nav: ↑/↓ or k/j cycle
// highlight; Enter commits; Esc reverts. The selector does NOT mutate theme
// or config itself — it emits ThemeSelectorAction signals and the caller
// drives the actual repaint + write-back.
type ThemeSelectorModel struct {
	open        bool
	themes      []string
	highlighted int
	preOpen     string
}

// NewThemeSelectorModel creates a ThemeSelectorModel populated with the
// bundled theme names. User-defined themes are out of scope (I.themesel).
func NewThemeSelectorModel() ThemeSelectorModel {
	return ThemeSelectorModel{themes: theme.BuiltinNames()}
}

// IsOpen reports whether the overlay is visible.
func (m ThemeSelectorModel) IsOpen() bool { return m.open }

// Open activates the overlay. The highlighted row starts on `current`;
// unknown names fall back to the first theme so Esc still has a meaningful
// pre-open to restore (GetTheme's fallback to DefaultThemeName is silent).
func (m ThemeSelectorModel) Open(current string) ThemeSelectorModel {
	m.open = true
	m.preOpen = current
	m.highlighted = 0
	for i, name := range m.themes {
		if name == current {
			m.highlighted = i
			break
		}
	}
	return m
}

// Close dismisses the overlay without signalling commit or revert. Used by
// Update after emitting the action.
func (m ThemeSelectorModel) Close() ThemeSelectorModel {
	m.open = false
	return m
}

// Highlighted returns the theme name at the current highlight row.
func (m ThemeSelectorModel) Highlighted() string {
	if m.highlighted < 0 || m.highlighted >= len(m.themes) {
		return ""
	}
	return m.themes[m.highlighted]
}

// PreOpen returns the theme name that was active when Open was called.
// Used by Esc to restore the pre-open theme.
func (m ThemeSelectorModel) PreOpen() string { return m.preOpen }

// Themes returns the list of selectable theme names in display order.
func (m ThemeSelectorModel) Themes() []string {
	out := make([]string, len(m.themes))
	copy(out, m.themes)
	return out
}

// Update handles key events while the overlay is open. Callers MUST only
// invoke Update when IsOpen() is true. Returns the updated model and the
// action the caller should perform. Keys other than arrows/k/j/Enter/Esc
// are consumed (no forward) so the TUI below the overlay never sees them.
func (m ThemeSelectorModel) Update(msg tea.Msg) (ThemeSelectorModel, ThemeSelectorAction) {
	if !m.open {
		return m, ThemeSelNone
	}
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, ThemeSelNone
	}
	switch keyMsg.String() {
	case "up", "k":
		if m.highlighted > 0 {
			m.highlighted--
			return m, ThemeSelPreview
		}
		return m, ThemeSelNone
	case "down", "j":
		if m.highlighted < len(m.themes)-1 {
			m.highlighted++
			return m, ThemeSelPreview
		}
		return m, ThemeSelNone
	case "enter":
		m.open = false
		return m, ThemeSelCommit
	case "esc":
		m.open = false
		return m, ThemeSelRevert
	}
	return m, ThemeSelNone
}

// View renders the theme selector overlay, or empty string when closed.
// V28: pads with spaces only — no `\t`. Highlight row is rendered with
// FocusBorder fg so the selected entry is visible regardless of the
// terminal's default fg.
func (m ThemeSelectorModel) View() string {
	if !m.open {
		return ""
	}
	// Resolve theme tokens via the highlighted theme so the selector itself
	// live-previews (alongside the whole-TUI repaint driven by the caller).
	th := theme.GetTheme(m.Highlighted())
	var sb strings.Builder
	sb.WriteString("Theme\n")
	sb.WriteString(strings.Repeat("─", 40))
	sb.WriteByte('\n')
	sb.WriteByte('\n')
	selStyle := lipgloss.NewStyle().Bold(true).Foreground(th.FocusBorder)
	for i, name := range m.themes {
		if i == m.highlighted {
			sb.WriteString(selStyle.Render("> " + name))
		} else {
			sb.WriteString("  " + name)
		}
		sb.WriteByte('\n')
	}
	sb.WriteByte('\n')
	sb.WriteString("↑/↓ or k/j  Navigate\n")
	sb.WriteString("Enter       Apply + save\n")
	sb.WriteString("Esc         Cancel (restore previous theme)\n")
	return sb.String()
}
