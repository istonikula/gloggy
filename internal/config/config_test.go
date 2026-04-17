package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultConfigPath(t *testing.T) {
	path, err := DefaultConfigPath()
	if err != nil {
		t.Fatalf("DefaultConfigPath() error: %v", err)
	}
	if !strings.HasSuffix(path, filepath.Join("gloggy", "config.toml")) {
		t.Errorf("path should end with gloggy/config.toml, got %s", path)
	}
}

func TestDefaultConfig_Values(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Theme != "tokyo-night" {
		t.Errorf("theme: want tokyo-night, got %s", cfg.Theme)
	}
	if cfg.LoggerDepth != 2 {
		t.Errorf("logger_depth: want 2, got %d", cfg.LoggerDepth)
	}
	if cfg.DetailPane.HeightRatio != 0.30 {
		t.Errorf("height_ratio: want 0.30, got %f", cfg.DetailPane.HeightRatio)
	}
	if cfg.DetailPane.WidthRatio != 0.30 {
		t.Errorf("width_ratio: want 0.30, got %f", cfg.DetailPane.WidthRatio)
	}
	if cfg.DetailPane.Position != "auto" {
		t.Errorf("position: want auto, got %s", cfg.DetailPane.Position)
	}
	if cfg.DetailPane.OrientationThresholdCols != 100 {
		t.Errorf("orientation_threshold_cols: want 100, got %d", cfg.DetailPane.OrientationThresholdCols)
	}
	if cfg.DetailPane.WrapMode != "soft" {
		t.Errorf("wrap_mode: want soft, got %s", cfg.DetailPane.WrapMode)
	}
	wantFields := []string{"time", "level", "logger", "msg"}
	if len(cfg.CompactRow.Fields) != len(wantFields) {
		t.Fatalf("compact_row.fields len: want %d, got %d", len(wantFields), len(cfg.CompactRow.Fields))
	}
	for i, f := range wantFields {
		if cfg.CompactRow.Fields[i] != f {
			t.Errorf("compact_row.fields[%d]: want %s, got %s", i, f, cfg.CompactRow.Fields[i])
		}
	}
}

func TestLoad_NoFile_ReturnsDefaults(t *testing.T) {
	result := Load("/nonexistent/path/config.toml")
	defaults := DefaultConfig()
	if result.Config.Theme != defaults.Theme {
		t.Errorf("theme: want %s, got %s", defaults.Theme, result.Config.Theme)
	}
	if len(result.Warnings) != 0 {
		t.Errorf("expected no warnings for missing file, got %v", result.Warnings)
	}
}

func TestLoad_InvalidTOML_ReturnsDefaults(t *testing.T) {
	result := LoadFromBytes([]byte(`not valid toml [[[`))
	defaults := DefaultConfig()
	if result.Config.Theme != defaults.Theme {
		t.Errorf("expected default theme, got %s", result.Config.Theme)
	}
	if len(result.Warnings) == 0 {
		t.Error("expected warnings for invalid TOML")
	}
}

func TestLoad_InvalidFieldValues_FallbackPerField(t *testing.T) {
	result := LoadFromBytes([]byte(`
theme = ""
logger_depth = -5

[detail_pane]
height_ratio = 2.0
`))
	defaults := DefaultConfig()
	if result.Config.Theme != defaults.Theme {
		t.Errorf("invalid theme should fallback, got %s", result.Config.Theme)
	}
	if result.Config.LoggerDepth != defaults.LoggerDepth {
		t.Errorf("invalid logger_depth should fallback, got %d", result.Config.LoggerDepth)
	}
	if result.Config.DetailPane.HeightRatio != defaults.DetailPane.HeightRatio {
		t.Errorf("invalid height_ratio should fallback, got %f", result.Config.DetailPane.HeightRatio)
	}
}

func TestLoad_PartialOverride_DefaultsForRest(t *testing.T) {
	result := LoadFromBytes([]byte(`logger_depth = 5`))
	if result.Config.LoggerDepth != 5 {
		t.Errorf("overridden field: want 5, got %d", result.Config.LoggerDepth)
	}
	if result.Config.Theme != DefaultConfig().Theme {
		t.Errorf("non-overridden theme should be default, got %s", result.Config.Theme)
	}
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
	if err := Save(path, result); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	saved, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read saved: %v", err)
	}
	if !strings.Contains(string(saved), "future_feature") {
		t.Error("unknown key 'future_feature' not preserved")
	}

	reloaded := Load(path)
	if reloaded.Config.Theme != "catppuccin-mocha" {
		t.Errorf("theme not preserved, got %s", reloaded.Config.Theme)
	}
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
	if len(result.Warnings) != 0 {
		t.Fatalf("unexpected warnings on valid overrides: %v", result.Warnings)
	}
	dp := result.Config.DetailPane
	if dp.HeightRatio != 0.40 || dp.WidthRatio != 0.55 || dp.Position != "right" ||
		dp.OrientationThresholdCols != 80 || dp.WrapMode != "soft" {
		t.Errorf("override values not preserved: %+v", dp)
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := Save(path, result); err != nil {
		t.Fatalf("Save: %v", err)
	}
	reloaded := Load(path)
	dp2 := reloaded.Config.DetailPane
	if dp2.HeightRatio != 0.40 || dp2.WidthRatio != 0.55 || dp2.Position != "right" ||
		dp2.OrientationThresholdCols != 80 || dp2.WrapMode != "soft" {
		t.Errorf("override values not preserved through round-trip: %+v", dp2)
	}
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
	if dp.WidthRatio != defaults.WidthRatio {
		t.Errorf("invalid width_ratio should fallback, got %f", dp.WidthRatio)
	}
	if dp.Position != defaults.Position {
		t.Errorf("invalid position should fallback to %s, got %s", defaults.Position, dp.Position)
	}
	if dp.OrientationThresholdCols != defaults.OrientationThresholdCols {
		t.Errorf("invalid orientation_threshold_cols should fallback, got %d", dp.OrientationThresholdCols)
	}
	if dp.WrapMode != defaults.WrapMode {
		t.Errorf("invalid wrap_mode should fallback to %s, got %s", defaults.WrapMode, dp.WrapMode)
	}
	if len(result.Warnings) < 4 {
		t.Errorf("expected 4+ warnings for 4 invalid values, got %d: %v", len(result.Warnings), result.Warnings)
	}
}

func TestSave_CreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "config.toml")
	if err := Save(path, LoadResult{Config: DefaultConfig()}); err != nil {
		t.Fatalf("Save() error: %v", err)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("config file not created")
	}
}
