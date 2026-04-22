package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/istonikula/gloggy/internal/config"
	"github.com/istonikula/gloggy/internal/theme"
	"github.com/istonikula/gloggy/internal/ui/detailpane"
)

// newModelWithConfig builds a Model pointed at a real on-disk config file so
// Enter-commits-theme tests can observe the post-write contents. The file is
// seeded with `initialTOML` (empty string = no file yet).
func newModelWithConfig(t *testing.T, initialTOML string) (Model, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.toml")
	if initialTOML != "" {
		require.NoError(t, os.WriteFile(path, []byte(initialTOML), 0o644),
			"write initial config")
	}
	cfgResult := config.Load(path)
	m := New("", false, path, cfgResult)
	return m, path
}

// ---------- V29 (a): T is gated on pane-search input mode (V14) ----------

// TestModel_T_DoesNotOpenThemesel_DuringListSearchInput verifies V14/V29(a):
// while the list search is in input mode, `T` extends the query instead of
// opening the theme selector — same policy as `?` / `q`.
func TestModel_T_DoesNotOpenThemesel_DuringListSearchInput(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	m = m.SetEntries(makeEntries(5))
	m = key(m, "/")
	m = key(m, "a")
	m = key(m, "b")
	require.True(t, m.list.HasActiveSearch(), "precondition: list search active")
	require.True(t, m.list.Search().InputMode(), "precondition: list search in input mode")

	m = key(m, "T")

	assert.Falsef(t, m.themesel.IsOpen(),
		"T must NOT open theme selector while list search is in input mode (V14/V29a)")
	assert.Equalf(t, "abT", m.list.Search().Query(),
		"T should extend the query to %q, got %q", "abT", m.list.Search().Query())
}

// TestModel_T_DoesNotOpenThemesel_DuringPaneSearchInput verifies V14/V29(a)
// for the detail pane: with pane search in input mode, `T` is consumed as a
// query char and the selector stays closed.
func TestModel_T_DoesNotOpenThemesel_DuringPaneSearchInput(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	entries := makeEntries(3)
	m = m.SetEntries(entries)
	m = m.openPane(entries[0])
	m = key(m, "tab")
	m = key(m, "/")
	require.True(t, m.paneSearch.IsActive(), "precondition: pane search active")
	require.Equalf(t, detailpane.SearchModeInput, m.paneSearch.Mode(),
		"precondition: pane search in input mode")

	m = key(m, "T")

	assert.Falsef(t, m.themesel.IsOpen(),
		"T must NOT open theme selector while pane search is in input mode (V14/V29a)")
	assert.Equalf(t, "T", m.paneSearch.Query(),
		"T should extend the pane-search query, got %q", m.paneSearch.Query())
}

// TestModel_T_OpensThemesel_NoActiveSearch pins the happy path: `T` opens
// the selector when no pane search is in input mode, and the overlay's
// highlighted theme matches the current config.
func TestModel_T_OpensThemesel_NoActiveSearch(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	require.False(t, m.themesel.IsOpen(), "precondition: selector closed")

	m = key(m, "T")

	assert.Truef(t, m.themesel.IsOpen(), "T should open the theme selector")
	assert.Equalf(t, m.cfg.Config.Theme, m.themesel.Highlighted(),
		"opening highlights the current theme")
}

// ---------- V29 (b): navigation previews the highlighted theme ----------

// TestModel_ThemeSelector_Navigate_Previews verifies that pressing down in
// the selector swaps the app model's theme to the newly highlighted one,
// driving the whole-TUI repaint via the single theme.GetTheme path.
func TestModel_ThemeSelector_Navigate_Previews(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	m = m.SetEntries(makeEntries(3))

	require.Equalf(t, "tokyo-night", m.th.Name, "precondition: default theme")

	m = key(m, "T")
	require.Truef(t, m.themesel.IsOpen(), "selector should be open")
	m = send(m, tea.KeyMsg{Type: tea.KeyDown})

	want := theme.BuiltinNames()[1]
	assert.Equalf(t, want, m.th.Name,
		"down should preview the next theme, got %q", m.th.Name)
	assert.Equalf(t, want, m.themesel.Highlighted(),
		"highlight should track the preview")
}

// TestModel_ThemeSelector_View_UsesHighlightedTheme verifies V29(b): after
// navigating, the selector's View renders with the highlighted theme's
// FocusBorder color. We probe by rendering a known string with the same
// foreground and checking the SGR prefix appears in the View output — the
// terminal emits 24-bit RGB triplets, not raw hex strings, so asserting
// the hex literal does not match.
func TestModel_ThemeSelector_View_UsesHighlightedTheme(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)

	m = key(m, "T")
	m = send(m, tea.KeyMsg{Type: tea.KeyDown})

	view := m.View()
	highlightedTheme := theme.GetTheme(m.themesel.Highlighted())
	probe := lipgloss.NewStyle().
		Bold(true).
		Foreground(highlightedTheme.FocusBorder).
		Render("x")
	// The probe always carries the SGR prefix + fg bytes; drop the trailing
	// "x\x1b[0m" so we match only the style-open run.
	sgrPrefix := probe[:strings.IndexByte(probe, 'x')]
	require.NotEmpty(t, sgrPrefix, "probe must emit an SGR prefix")
	assert.Containsf(t, view, sgrPrefix,
		"View should render with the highlighted theme's FocusBorder SGR prefix %q",
		sgrPrefix)
}

// ---------- V29 (c): Enter commits + persists with unknown-key preservation ----------

// TestModel_ThemeSelector_Enter_PersistsTheme verifies V22 + V29(c): Enter
// writes the highlighted theme to config.toml via the existing write-back,
// and unknown keys in the existing config survive the round-trip.
func TestModel_ThemeSelector_Enter_PersistsTheme(t *testing.T) {
	initial := `theme = "tokyo-night"
future_plugin = "preserved"
[detail_pane]
height_ratio = 0.30
experimental_feature = true
`
	m, path := newModelWithConfig(t, initial)
	m = resize(m, 80, 24)

	m = key(m, "T")
	m = send(m, tea.KeyMsg{Type: tea.KeyDown}) // preview next theme
	want := theme.BuiltinNames()[1]
	require.Equalf(t, want, m.themesel.Highlighted(), "highlighted should be %q", want)

	m = send(m, tea.KeyMsg{Type: tea.KeyEnter})

	assert.Falsef(t, m.themesel.IsOpen(), "Enter should close the selector")
	assert.Equalf(t, want, m.cfg.Config.Theme,
		"Enter should update m.cfg.Config.Theme")
	assert.Equalf(t, want, m.th.Name,
		"Enter should leave the committed theme active")

	// Reload via config.Load — most robust way to verify persistence
	// without coupling to go-toml's quote style.
	reloaded := config.Load(path)
	assert.Equalf(t, want, reloaded.Config.Theme,
		"reloaded config should reflect the committed theme")

	// V22: unknown keys in the original file must survive the round-trip.
	data, err := os.ReadFile(path)
	require.NoError(t, err, "read persisted config")
	assert.Containsf(t, string(data), "future_plugin",
		"unknown top-level key must survive write-back (V22)")
	assert.Containsf(t, string(data), "experimental_feature",
		"unknown nested key must survive write-back (V22)")
}

// ---------- V29 (d): Esc reverts, no config write ----------

// TestModel_ThemeSelector_Esc_RevertsPreOpen_NoWrite verifies V29(d): Esc
// restores the pre-open theme and does NOT advance the config mtime (no
// on-disk write). We preview a non-default theme first, then Esc, and
// confirm the active theme is restored AND the config file is untouched
// (by checking the mtime and the serialised theme= value).
func TestModel_ThemeSelector_Esc_RevertsPreOpen_NoWrite(t *testing.T) {
	initial := `theme = "tokyo-night"
`
	m, path := newModelWithConfig(t, initial)
	m = resize(m, 80, 24)
	require.Equalf(t, "tokyo-night", m.cfg.Config.Theme, "precondition: tokyo-night")

	statBefore, err := os.Stat(path)
	require.NoError(t, err)
	dataBefore, err := os.ReadFile(path)
	require.NoError(t, err)

	m = key(m, "T")
	m = send(m, tea.KeyMsg{Type: tea.KeyDown}) // preview
	require.NotEqualf(t, "tokyo-night", m.th.Name,
		"preview should have swapped away from tokyo-night")

	m = send(m, tea.KeyMsg{Type: tea.KeyEsc})

	assert.Falsef(t, m.themesel.IsOpen(), "Esc should close the selector")
	assert.Equalf(t, "tokyo-night", m.th.Name,
		"Esc should restore the pre-open theme")
	assert.Equalf(t, "tokyo-night", m.cfg.Config.Theme,
		"Esc must not change m.cfg.Config.Theme")

	statAfter, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equalf(t, statBefore.ModTime(), statAfter.ModTime(),
		"Esc must not advance config mtime (no write-back)")
	dataAfter, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equalf(t, string(dataBefore), string(dataAfter),
		"config file bytes must be identical after Esc (no write)")
}

// ---------- closed-overlay invariant: no theme/config mutation ----------

// TestModel_ThemeSelector_Closed_NoMutation verifies V29: with the selector
// never opened, the active theme + config.Theme + config file are
// untouched. Guards against any accidental theme-swap path not funneled
// through the selector.
func TestModel_ThemeSelector_Closed_NoMutation(t *testing.T) {
	m, path := newModelWithConfig(t, `theme = "tokyo-night"
`)
	m = resize(m, 80, 24)

	statBefore, err := os.Stat(path)
	require.NoError(t, err)

	// A variety of non-T keys should not touch theme state.
	m = key(m, "j")
	m = key(m, "k")
	m = key(m, "f")
	m = key(m, "esc")

	assert.Falsef(t, m.themesel.IsOpen(), "selector must stay closed")
	assert.Equalf(t, "tokyo-night", m.th.Name, "active theme must not change")
	assert.Equalf(t, "tokyo-night", m.cfg.Config.Theme, "config theme must not change")

	statAfter, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equalf(t, statBefore.ModTime(), statAfter.ModTime(),
		"config file mtime must not advance")
}

