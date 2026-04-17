// Package config handles TOML-based configuration for gloggy.
package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	toml "github.com/pelletier/go-toml/v2"
)

// Config holds all gloggy configuration fields.
type Config struct {
	Theme        string     `toml:"theme"`
	CompactRow   CompactRow `toml:"compact_row"`
	HiddenFields []string   `toml:"hidden_fields"`
	LoggerDepth  int        `toml:"logger_depth"`
	DetailPane   DetailPane `toml:"detail_pane"`
}

// CompactRow holds settings for the compact table row display.
type CompactRow struct {
	Fields    []string `toml:"fields"`
	SubFields []string `toml:"sub_fields"`
}

// DetailPane holds settings for the detail pane.
type DetailPane struct {
	HeightRatio              float64 `toml:"height_ratio"`
	WidthRatio               float64 `toml:"width_ratio"`
	Position                 string  `toml:"position"`
	OrientationThresholdCols int     `toml:"orientation_threshold_cols"`
	WrapMode                 string  `toml:"wrap_mode"`
}

// DefaultConfig returns the built-in default configuration.
func DefaultConfig() Config {
	return Config{
		Theme: "tokyo-night",
		CompactRow: CompactRow{
			Fields:    []string{"time", "level", "logger", "msg"},
			SubFields: []string{},
		},
		HiddenFields: []string{},
		LoggerDepth:  2,
		DetailPane: DetailPane{
			HeightRatio:              0.30,
			WidthRatio:               0.30,
			Position:                 "auto",
			OrientationThresholdCols: 100,
			WrapMode:                 "soft",
		},
	}
}

// LoadResult contains the loaded config and any warnings encountered.
type LoadResult struct {
	Config   Config
	Warnings []string
	rawData  map[string]any
}

// DefaultConfigPath returns the platform-appropriate config file path.
func DefaultConfigPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("config dir: %w", err)
	}
	return filepath.Join(dir, "gloggy", "config.toml"), nil
}

// Load reads and parses a config file at the given path.
// If the file does not exist, defaults are returned without error.
func Load(path string) LoadResult {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return LoadResult{Config: DefaultConfig()}
		}
		return LoadResult{
			Config:   DefaultConfig(),
			Warnings: []string{fmt.Sprintf("cannot open config: %v", err)},
		}
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		return LoadResult{
			Config:   DefaultConfig(),
			Warnings: []string{fmt.Sprintf("cannot read config: %v", err)},
		}
	}
	return LoadFromBytes(data)
}

// LoadFromBytes parses config from raw TOML bytes.
func LoadFromBytes(data []byte) LoadResult {
	defaults := DefaultConfig()
	result := LoadResult{Config: defaults}

	var rawMap map[string]any
	if err := toml.Unmarshal(data, &rawMap); err != nil {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("TOML parse error, using defaults: %v", err))
		return result
	}
	result.rawData = rawMap

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("TOML decode error, using defaults: %v", err))
		return result
	}

	result.Config = mergeWithDefaults(cfg, rawMap, defaults)
	result.Config, result.Warnings = validateConfig(result.Config, defaults)
	return result
}

func mergeWithDefaults(cfg Config, raw map[string]any, defaults Config) Config {
	if _, ok := raw["theme"]; !ok {
		cfg.Theme = defaults.Theme
	}
	if _, ok := raw["logger_depth"]; !ok {
		cfg.LoggerDepth = defaults.LoggerDepth
	}
	if _, ok := raw["hidden_fields"]; !ok {
		cfg.HiddenFields = defaults.HiddenFields
	}

	if cr, ok := raw["compact_row"]; ok {
		if cm, ok := cr.(map[string]any); ok {
			if _, ok := cm["fields"]; !ok {
				cfg.CompactRow.Fields = defaults.CompactRow.Fields
			}
			if _, ok := cm["sub_fields"]; !ok {
				cfg.CompactRow.SubFields = defaults.CompactRow.SubFields
			}
		}
	} else {
		cfg.CompactRow = defaults.CompactRow
	}

	if dp, ok := raw["detail_pane"]; ok {
		if dm, ok := dp.(map[string]any); ok {
			if _, ok := dm["height_ratio"]; !ok {
				cfg.DetailPane.HeightRatio = defaults.DetailPane.HeightRatio
			}
			if _, ok := dm["width_ratio"]; !ok {
				cfg.DetailPane.WidthRatio = defaults.DetailPane.WidthRatio
			}
			if _, ok := dm["position"]; !ok {
				cfg.DetailPane.Position = defaults.DetailPane.Position
			}
			if _, ok := dm["orientation_threshold_cols"]; !ok {
				cfg.DetailPane.OrientationThresholdCols = defaults.DetailPane.OrientationThresholdCols
			}
			if _, ok := dm["wrap_mode"]; !ok {
				cfg.DetailPane.WrapMode = defaults.DetailPane.WrapMode
			}
		}
	} else {
		cfg.DetailPane = defaults.DetailPane
	}

	if cfg.CompactRow.Fields == nil {
		cfg.CompactRow.Fields = []string{}
	}
	if cfg.CompactRow.SubFields == nil {
		cfg.CompactRow.SubFields = []string{}
	}
	if cfg.HiddenFields == nil {
		cfg.HiddenFields = []string{}
	}
	return cfg
}

func validateConfig(cfg Config, defaults Config) (Config, []string) {
	var warnings []string
	if cfg.Theme == "" {
		cfg.Theme = defaults.Theme
		warnings = append(warnings, "invalid theme (empty), using default")
	}
	if cfg.LoggerDepth < 0 {
		warnings = append(warnings, fmt.Sprintf("invalid logger_depth %d, using default %d", cfg.LoggerDepth, defaults.LoggerDepth))
		cfg.LoggerDepth = defaults.LoggerDepth
	}
	if cfg.DetailPane.HeightRatio <= 0 || cfg.DetailPane.HeightRatio >= 1 {
		warnings = append(warnings, fmt.Sprintf("invalid detail_pane.height_ratio %.2f, using default %.2f", cfg.DetailPane.HeightRatio, defaults.DetailPane.HeightRatio))
		cfg.DetailPane.HeightRatio = defaults.DetailPane.HeightRatio
	}
	if cfg.DetailPane.WidthRatio <= 0 || cfg.DetailPane.WidthRatio >= 1 {
		warnings = append(warnings, fmt.Sprintf("invalid detail_pane.width_ratio %.2f, using default %.2f", cfg.DetailPane.WidthRatio, defaults.DetailPane.WidthRatio))
		cfg.DetailPane.WidthRatio = defaults.DetailPane.WidthRatio
	}
	switch cfg.DetailPane.Position {
	case "auto", "below", "right":
		// valid
	default:
		warnings = append(warnings, fmt.Sprintf("invalid detail_pane.position %q, using default %q", cfg.DetailPane.Position, defaults.DetailPane.Position))
		cfg.DetailPane.Position = defaults.DetailPane.Position
	}
	if cfg.DetailPane.OrientationThresholdCols <= 0 {
		warnings = append(warnings, fmt.Sprintf("invalid detail_pane.orientation_threshold_cols %d, using default %d", cfg.DetailPane.OrientationThresholdCols, defaults.DetailPane.OrientationThresholdCols))
		cfg.DetailPane.OrientationThresholdCols = defaults.DetailPane.OrientationThresholdCols
	}
	switch cfg.DetailPane.WrapMode {
	case "soft", "scroll", "modal":
		// valid (only "soft" is shipping; others reserved per kit)
	default:
		warnings = append(warnings, fmt.Sprintf("invalid detail_pane.wrap_mode %q, using default %q", cfg.DetailPane.WrapMode, defaults.DetailPane.WrapMode))
		cfg.DetailPane.WrapMode = defaults.DetailPane.WrapMode
	}
	return cfg, warnings
}

// Save writes the config to path, preserving unknown keys from the original load.
func Save(path string, result LoadResult) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	output := make(map[string]any)
	for k, v := range result.rawData {
		output[k] = v
	}
	output["theme"] = result.Config.Theme
	output["logger_depth"] = result.Config.LoggerDepth
	output["hidden_fields"] = result.Config.HiddenFields

	compact := make(map[string]any)
	if existing, ok := output["compact_row"].(map[string]any); ok {
		for k, v := range existing {
			compact[k] = v
		}
	}
	compact["fields"] = result.Config.CompactRow.Fields
	compact["sub_fields"] = result.Config.CompactRow.SubFields
	output["compact_row"] = compact

	detail := make(map[string]any)
	if existing, ok := output["detail_pane"].(map[string]any); ok {
		for k, v := range existing {
			detail[k] = v
		}
	}
	detail["height_ratio"] = result.Config.DetailPane.HeightRatio
	detail["width_ratio"] = result.Config.DetailPane.WidthRatio
	detail["position"] = result.Config.DetailPane.Position
	detail["orientation_threshold_cols"] = result.Config.DetailPane.OrientationThresholdCols
	detail["wrap_mode"] = result.Config.DetailPane.WrapMode
	output["detail_pane"] = detail

	data, err := toml.Marshal(output)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}
