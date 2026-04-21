package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfigPath(t *testing.T) {
	path, err := DefaultConfigPath()
	require.NoError(t, err)
	assert.True(t, strings.HasSuffix(path, filepath.Join("gloggy", "config.toml")),
		"path should end with gloggy/config.toml, got %s", path)
}

func TestDefaultConfig_Values(t *testing.T) {
	cfg := DefaultConfig()
	assert.Equal(t, "tokyo-night", cfg.Theme)
	assert.Equal(t, 2, cfg.LoggerDepth)
	assert.InDelta(t, 0.30, cfg.DetailPane.HeightRatio, 1e-9)
	assert.InDelta(t, 0.30, cfg.DetailPane.WidthRatio, 1e-9)
	assert.Equal(t, "auto", cfg.DetailPane.Position)
	assert.Equal(t, 100, cfg.DetailPane.OrientationThresholdCols)
	assert.Equal(t, "soft", cfg.DetailPane.WrapMode)
	wantFields := []string{"time", "level", "logger", "msg"}
	require.Equal(t, wantFields, cfg.CompactRow.Fields)
}

func TestLoad_NoFile_ReturnsDefaults(t *testing.T) {
	result := Load("/nonexistent/path/config.toml")
	defaults := DefaultConfig()
	assert.Equal(t, defaults.Theme, result.Config.Theme)
	assert.Empty(t, result.Warnings, "expected no warnings for missing file")
}

func TestLoad_InvalidTOML_ReturnsDefaults(t *testing.T) {
	result := LoadFromBytes([]byte(`not valid toml [[[`))
	defaults := DefaultConfig()
	assert.Equal(t, defaults.Theme, result.Config.Theme)
	assert.NotEmpty(t, result.Warnings, "expected warnings for invalid TOML")
}

func TestLoad_InvalidFieldValues_FallbackPerField(t *testing.T) {
	result := LoadFromBytes([]byte(`
theme = ""
logger_depth = -5

[detail_pane]
height_ratio = 2.0
`))
	defaults := DefaultConfig()
	assert.Equal(t, defaults.Theme, result.Config.Theme, "invalid theme should fallback")
	assert.Equal(t, defaults.LoggerDepth, result.Config.LoggerDepth, "invalid logger_depth should fallback")
	assert.InDelta(t, defaults.DetailPane.HeightRatio, result.Config.DetailPane.HeightRatio, 1e-9,
		"invalid height_ratio should fallback")
}

func TestLoad_PartialOverride_DefaultsForRest(t *testing.T) {
	result := LoadFromBytes([]byte(`logger_depth = 5`))
	assert.Equal(t, 5, result.Config.LoggerDepth, "overridden field")
	assert.Equal(t, DefaultConfig().Theme, result.Config.Theme, "non-overridden theme should be default")
}

func TestRoundTrip_PreservesUnknownKeys(t *testing.T) {
	input := `
theme = "catppuccin-mocha"
future_feature = true
logger_depth = 3
`
	result := LoadFromBytes([]byte(input))

	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	require.NoError(t, Save(path, result))

	saved, err := os.ReadFile(path)
	require.NoError(t, err, "read saved")
	assert.Contains(t, string(saved), "future_feature", "unknown key 'future_feature' not preserved")

	reloaded := Load(path)
	assert.Equal(t, "catppuccin-mocha", reloaded.Config.Theme, "theme not preserved")
}

// T-085: detail_pane orientation/width keys round-trip.
func TestDetailPane_OrientationKeysRoundTrip(t *testing.T) {
	input := `
[detail_pane]
height_ratio = 0.40
width_ratio = 0.55
position = "right"
orientation_threshold_cols = 80
wrap_mode = "soft"
`
	result := LoadFromBytes([]byte(input))
	require.Empty(t, result.Warnings, "unexpected warnings on valid overrides")
	dp := result.Config.DetailPane
	assert.InDelta(t, 0.40, dp.HeightRatio, 1e-9, "HeightRatio preserved")
	assert.InDelta(t, 0.55, dp.WidthRatio, 1e-9, "WidthRatio preserved")
	assert.Equal(t, "right", dp.Position)
	assert.Equal(t, 80, dp.OrientationThresholdCols)
	assert.Equal(t, "soft", dp.WrapMode)

	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	require.NoError(t, Save(path, result))
	reloaded := Load(path)
	dp2 := reloaded.Config.DetailPane
	assert.InDelta(t, 0.40, dp2.HeightRatio, 1e-9, "HeightRatio round-trip")
	assert.InDelta(t, 0.55, dp2.WidthRatio, 1e-9, "WidthRatio round-trip")
	assert.Equal(t, "right", dp2.Position)
	assert.Equal(t, 80, dp2.OrientationThresholdCols)
	assert.Equal(t, "soft", dp2.WrapMode)
}

// T-085: invalid detail_pane keys fall back to defaults with warnings.
func TestDetailPane_InvalidValuesFallback(t *testing.T) {
	input := `
[detail_pane]
height_ratio = 0.30
width_ratio = 2.5
position = "diagonal"
orientation_threshold_cols = -10
wrap_mode = "telepathic"
`
	result := LoadFromBytes([]byte(input))
	defaults := DefaultConfig().DetailPane
	dp := result.Config.DetailPane
	assert.InDelta(t, defaults.WidthRatio, dp.WidthRatio, 1e-9, "invalid width_ratio should fallback")
	assert.Equal(t, defaults.Position, dp.Position, "invalid position should fallback")
	assert.Equal(t, defaults.OrientationThresholdCols, dp.OrientationThresholdCols,
		"invalid orientation_threshold_cols should fallback")
	assert.Equal(t, defaults.WrapMode, dp.WrapMode, "invalid wrap_mode should fallback")
	assert.GreaterOrEqual(t, len(result.Warnings), 4,
		"expected 4+ warnings for 4 invalid values, got %v", result.Warnings)
}

// T-130: shared top-level scrolloff config.
func TestDefaultConfig_ScrolloffIs5(t *testing.T) {
	cfg := DefaultConfig()
	assert.Equal(t, 5, cfg.Scrolloff, "default scrolloff")
}

func TestScrolloff_OverrideRoundTrips(t *testing.T) {
	input := `scrolloff = 12
`
	result := LoadFromBytes([]byte(input))
	require.Empty(t, result.Warnings, "unexpected warnings on valid scrolloff")
	assert.Equal(t, 12, result.Config.Scrolloff, "override scrolloff")
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	require.NoError(t, Save(path, result))
	reloaded := Load(path)
	assert.Equal(t, 12, reloaded.Config.Scrolloff, "reloaded scrolloff")
}

func TestScrolloff_NegativeClampedToZero(t *testing.T) {
	result := LoadFromBytes([]byte(`scrolloff = -3
`))
	assert.Equal(t, 0, result.Config.Scrolloff, "negative scrolloff should clamp to 0")
	found := false
	for _, w := range result.Warnings {
		if strings.Contains(w, "scrolloff") {
			found = true
			break
		}
	}
	assert.True(t, found, "expected warning about scrolloff, got %v", result.Warnings)
}

func TestScrolloff_MissingKeyUsesDefault(t *testing.T) {
	result := LoadFromBytes([]byte(`theme = "tokyo-night"
`))
	assert.Equal(t, 5, result.Config.Scrolloff, "missing scrolloff should default to 5")
}

func TestSave_CreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "config.toml")
	require.NoError(t, Save(path, LoadResult{Config: DefaultConfig()}))
	_, err := os.Stat(path)
	assert.False(t, os.IsNotExist(err), "config file not created")
}
