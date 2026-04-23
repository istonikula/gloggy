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

	// A variety of non-T / non-t keys should not touch theme state.
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

// ---------- V30 (a): t is gated on pane-search input mode (V14) ----------

// TestModel_t_DoesNotCycleTheme_DuringListSearchInput verifies V14/V30(a):
// while the list search is in input mode, `t` extends the query instead of
// cycling the theme — same policy as `T` / `?` / `q`.
func TestModel_t_DoesNotCycleTheme_DuringListSearchInput(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	m = m.SetEntries(makeEntries(5))
	m = key(m, "/")
	m = key(m, "a")
	m = key(m, "b")
	require.True(t, m.list.HasActiveSearch(), "precondition: list search active")
	require.True(t, m.list.Search().InputMode(), "precondition: list search in input mode")
	before := m.th.Name

	m = key(m, "t")

	assert.Equalf(t, before, m.th.Name,
		"t must NOT cycle theme while list search is in input mode (V14/V30a)")
	assert.Equalf(t, "abt", m.list.Search().Query(),
		"t should extend the query to %q, got %q", "abt", m.list.Search().Query())
}

// TestModel_t_DoesNotCycleTheme_DuringPaneSearchInput verifies V14/V30(a)
// for the detail pane: with pane search in input mode, `t` is consumed as
// a query char and the theme is not swapped.
func TestModel_t_DoesNotCycleTheme_DuringPaneSearchInput(t *testing.T) {
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
	before := m.th.Name

	m = key(m, "t")

	assert.Equalf(t, before, m.th.Name,
		"t must NOT cycle theme while pane search is in input mode (V14/V30a)")
	assert.Equalf(t, "t", m.paneSearch.Query(),
		"t should extend the pane-search query, got %q", m.paneSearch.Query())
}

// ---------- V30 (b, c): sequential t cycles + persists each step ----------

// TestModel_t_CyclesThroughBundledThemes_AndPersists verifies V30(b)+(c):
// sequential `t` presses advance through bundled themes in declaration
// order (tokyo-night → catppuccin-mocha → material-dark → wrap), each
// press updates config.toml's theme= via write-back, and unrelated keys
// in the file survive the round-trip (V22).
func TestModel_t_CyclesThroughBundledThemes_AndPersists(t *testing.T) {
	initial := `theme = "tokyo-night"
future_plugin = "preserved"
[detail_pane]
height_ratio = 0.30
experimental_feature = true
`
	m, path := newModelWithConfig(t, initial)
	m = resize(m, 80, 24)
	require.Equalf(t, "tokyo-night", m.th.Name, "precondition: tokyo-night active")

	names := theme.BuiltinNames()
	// Press t three times; after N presses we should be at names[N % len].
	// Starting from names[0], the sequence is: names[1], names[2], names[0].
	want := []string{names[1], names[2], names[0]}
	for i, w := range want {
		m = key(m, "t")
		assert.Equalf(t, w, m.th.Name,
			"press %d: active theme should be %q, got %q", i+1, w, m.th.Name)
		assert.Equalf(t, w, m.cfg.Config.Theme,
			"press %d: cfg.Config.Theme should be %q, got %q", i+1, w, m.cfg.Config.Theme)

		// Verify persistence + V22 unknown-key preservation after each press.
		reloaded := config.Load(path)
		assert.Equalf(t, w, reloaded.Config.Theme,
			"press %d: reloaded config.Theme should be %q", i+1, w)

		data, err := os.ReadFile(path)
		require.NoError(t, err, "read persisted config (press %d)", i+1)
		assert.Containsf(t, string(data), "future_plugin",
			"press %d: unknown top-level key must survive write-back (V22)", i+1)
		assert.Containsf(t, string(data), "experimental_feature",
			"press %d: unknown nested key must survive write-back (V22)", i+1)
	}

	// The selector must not be involved in the cycle path.
	assert.Falsef(t, m.themesel.IsOpen(),
		"t must not open the theme-selector overlay")
}

// ---------- V30 (d): overlay-open consumes t (no parallel cycle) ----------

// TestModel_ThemeSelector_Open_t_IsOverlayDomain verifies V30(d): while
// the selector overlay is open, `t` is the overlay's domain. It must not
// swap the theme via the direct-cycle path, nor advance the config mtime.
// (The overlay's Update consumes unknown keys as no-ops.)
func TestModel_ThemeSelector_Open_t_IsOverlayDomain(t *testing.T) {
	m, path := newModelWithConfig(t, `theme = "tokyo-night"
`)
	m = resize(m, 80, 24)

	m = key(m, "T")
	require.Truef(t, m.themesel.IsOpen(), "precondition: selector open")
	require.Equalf(t, "tokyo-night", m.th.Name, "precondition: tokyo-night active")

	statBefore, err := os.Stat(path)
	require.NoError(t, err)

	m = key(m, "t")

	assert.Truef(t, m.themesel.IsOpen(),
		"selector should remain open — t is overlay's domain, no side-effect")
	assert.Equalf(t, "tokyo-night", m.th.Name,
		"theme must NOT swap via direct cycle while selector is open (V30d)")
	assert.Equalf(t, "tokyo-night", m.cfg.Config.Theme,
		"cfg.Config.Theme must NOT change while selector is open")

	statAfter, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equalf(t, statBefore.ModTime(), statAfter.ModTime(),
		"config file mtime must not advance (no write-back from t while overlay open)")
}

