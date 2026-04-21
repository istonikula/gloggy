package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T-024: When hidden_fields is updated interactively and saved, the config
// file reflects the change (R6.1, R6.2, R6.3).
func TestSave_HiddenFieldsUpdatePersists(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")

	// Start with a config file that has some known and unknown keys.
	initial := `theme = "tokyo-night"
logger_depth = 2
unknown_future_key = "preserved"

[compact_row]
fields = ["time", "level", "msg"]

[detail_pane]
height_ratio = 0.30
`
	require.NoError(t, os.WriteFile(path, []byte(initial), 0o644), "write initial config")

	// Load it.
	result := Load(path)

	// Interactively hide a field.
	result.Config.HiddenFields = []string{"thread", "logger"}

	// Save back.
	require.NoError(t, Save(path, result))

	// Reload and verify.
	result2 := Load(path)
	require.Len(t, result2.Config.HiddenFields, 2)
	assert.Equal(t, "tokyo-night", result2.Config.Theme, "theme changed unexpectedly")
	assert.Equal(t, 2, result2.Config.LoggerDepth, "logger_depth changed unexpectedly")

	// R6.2: saved file must be valid TOML — reload succeeded above.
	// R6.3: other fields not affected — checked theme and logger_depth above.
}

// T-024: Save produces valid TOML (R6.2).
func TestSave_ProducesValidTOML(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	result := LoadResult{Config: DefaultConfig(), rawData: map[string]any{}}

	require.NoError(t, Save(path, result))

	// Loading the saved file must not produce warnings.
	result2 := Load(path)
	assert.Empty(t, result2.Warnings, "reloaded config has warnings")
}

// T-025: Unknown keys in the config survive a load-save round-trip (R7 / R3.2).
func TestSave_PreservesUnknownKeys(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")

	initial := `theme = "tokyo-night"
future_plugin = "enabled"
some_unknown_section = { value = 42 }

[detail_pane]
height_ratio = 0.30
experimental_feature = true
`
	require.NoError(t, os.WriteFile(path, []byte(initial), 0o644), "write initial")

	result := Load(path)
	require.NoError(t, Save(path, result))

	data, err := os.ReadFile(path)
	require.NoError(t, err, "read saved")
	content := string(data)

	for _, key := range []string{"future_plugin", "some_unknown_section"} {
		assert.Contains(t, content, key, "unknown key %q was not preserved after save", key)
	}
}

// T-086: Updating height_ratio via Save must not mutate width_ratio, and vice
// versa. The two ratios live as independent keys in the TOML file.
func TestSave_RatioIndependence_HeightChange(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	initial := `theme = "tokyo-night"

[detail_pane]
height_ratio = 0.30
width_ratio = 0.20
`
	require.NoError(t, os.WriteFile(path, []byte(initial), 0o644), "write initial")

	result := Load(path)
	require.InDelta(t, 0.20, result.Config.DetailPane.WidthRatio, 1e-9, "preload width_ratio")

	// Change only height_ratio.
	result.Config.DetailPane.HeightRatio = 0.50
	require.NoError(t, Save(path, result))

	reloaded := Load(path)
	assert.InDelta(t, 0.50, reloaded.Config.DetailPane.HeightRatio, 1e-9, "height_ratio after save")
	assert.InDelta(t, 0.20, reloaded.Config.DetailPane.WidthRatio, 1e-9,
		"width_ratio mutated by height_ratio write-back")
}

// T-086: Reverse direction — changing width_ratio must not mutate height_ratio.
func TestSave_RatioIndependence_WidthChange(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	initial := `theme = "tokyo-night"

[detail_pane]
height_ratio = 0.60
width_ratio = 0.30
`
	require.NoError(t, os.WriteFile(path, []byte(initial), 0o644), "write initial")

	result := Load(path)
	result.Config.DetailPane.WidthRatio = 0.45
	require.NoError(t, Save(path, result))

	reloaded := Load(path)
	assert.InDelta(t, 0.45, reloaded.Config.DetailPane.WidthRatio, 1e-9, "width_ratio after save")
	assert.InDelta(t, 0.60, reloaded.Config.DetailPane.HeightRatio, 1e-9,
		"height_ratio mutated by width_ratio write-back")
}

// T-025: Adding a new top-level key to config does not require schema changes (R7.1).
// This is structural: the current code handles all known fields + forwards unknown
// ones. A future version that adds a new key can be loaded by the current version
// without error (it lands in rawData).
func TestLoad_UnknownKeyDoesNotError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")

	initial := `theme = "catppuccin-mocha"
new_version_field = "hello from the future"
`
	require.NoError(t, os.WriteFile(path, []byte(initial), 0o644), "write")

	result := Load(path)
	// Must load without warnings.
	assert.Empty(t, result.Warnings, "unexpected warnings")
	assert.Equal(t, "catppuccin-mocha", result.Config.Theme)
}
