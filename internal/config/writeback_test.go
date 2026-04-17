package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
	if err := os.WriteFile(path, []byte(initial), 0o644); err != nil {
		t.Fatalf("write initial config: %v", err)
	}

	// Load it.
	result := Load(path)

	// Interactively hide a field.
	result.Config.HiddenFields = []string{"thread", "logger"}

	// Save back.
	if err := Save(path, result); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Reload and verify.
	result2 := Load(path)
	if len(result2.Config.HiddenFields) != 2 {
		t.Fatalf("expected 2 hidden fields, got %v", result2.Config.HiddenFields)
	}
	if result2.Config.Theme != "tokyo-night" {
		t.Errorf("theme changed unexpectedly: %q", result2.Config.Theme)
	}
	if result2.Config.LoggerDepth != 2 {
		t.Errorf("logger_depth changed unexpectedly: %d", result2.Config.LoggerDepth)
	}

	// R6.2: saved file must be valid TOML — reload succeeded above.
	// R6.3: other fields not affected — checked theme and logger_depth above.
}

// T-024: Save produces valid TOML (R6.2).
func TestSave_ProducesValidTOML(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	result := LoadResult{Config: DefaultConfig(), rawData: map[string]any{}}

	if err := Save(path, result); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Loading the saved file must not produce warnings.
	result2 := Load(path)
	if len(result2.Warnings) != 0 {
		t.Errorf("reloaded config has warnings: %v", result2.Warnings)
	}
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
	if err := os.WriteFile(path, []byte(initial), 0o644); err != nil {
		t.Fatalf("write initial: %v", err)
	}

	result := Load(path)
	if err := Save(path, result); err != nil {
		t.Fatalf("Save: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read saved: %v", err)
	}
	content := string(data)

	for _, key := range []string{"future_plugin", "some_unknown_section"} {
		if !strings.Contains(content, key) {
			t.Errorf("unknown key %q was not preserved after save", key)
		}
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
	if err := os.WriteFile(path, []byte(initial), 0o644); err != nil {
		t.Fatalf("write initial: %v", err)
	}

	result := Load(path)
	if result.Config.DetailPane.WidthRatio != 0.20 {
		t.Fatalf("preload width_ratio: got %.2f, want 0.20", result.Config.DetailPane.WidthRatio)
	}

	// Change only height_ratio.
	result.Config.DetailPane.HeightRatio = 0.50
	if err := Save(path, result); err != nil {
		t.Fatalf("Save: %v", err)
	}

	reloaded := Load(path)
	if reloaded.Config.DetailPane.HeightRatio != 0.50 {
		t.Errorf("height_ratio after save: got %.2f, want 0.50", reloaded.Config.DetailPane.HeightRatio)
	}
	if reloaded.Config.DetailPane.WidthRatio != 0.20 {
		t.Errorf("width_ratio mutated by height_ratio write-back: got %.2f, want 0.20",
			reloaded.Config.DetailPane.WidthRatio)
	}
}

// T-086: Reverse direction — changing width_ratio must not mutate height_ratio.
func TestSave_RatioIndependence_WidthChange(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	initial := `theme = "tokyo-night"

[detail_pane]
height_ratio = 0.60
width_ratio = 0.30
`
	if err := os.WriteFile(path, []byte(initial), 0o644); err != nil {
		t.Fatalf("write initial: %v", err)
	}

	result := Load(path)
	result.Config.DetailPane.WidthRatio = 0.45
	if err := Save(path, result); err != nil {
		t.Fatalf("Save: %v", err)
	}

	reloaded := Load(path)
	if reloaded.Config.DetailPane.WidthRatio != 0.45 {
		t.Errorf("width_ratio after save: got %.2f, want 0.45", reloaded.Config.DetailPane.WidthRatio)
	}
	if reloaded.Config.DetailPane.HeightRatio != 0.60 {
		t.Errorf("height_ratio mutated by width_ratio write-back: got %.2f, want 0.60",
			reloaded.Config.DetailPane.HeightRatio)
	}
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
	if err := os.WriteFile(path, []byte(initial), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	result := Load(path)
	// Must load without warnings.
	if len(result.Warnings) != 0 {
		t.Errorf("unexpected warnings: %v", result.Warnings)
	}
	if result.Config.Theme != "catppuccin-mocha" {
		t.Errorf("theme = %q, want catppuccin-mocha", result.Config.Theme)
	}
}
