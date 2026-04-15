---
created: "2026-04-15T00:00:00Z"
last_edited: "2026-04-15T00:00:00Z"
---

# Cavekit: Config

## Scope

Loading, validating, and persisting application configuration from a TOML file. Covers default generation, theme selection, field visibility settings, logger abbreviation depth, detail pane height, and live write-back of interactive changes.

## Requirements

### R1: Config File Location and Defaults
**Description:** Configuration is read from `~/.config/gloggy/config.toml`. If the file does not exist on first run, it is created with default values.
**Acceptance Criteria:**
- [ ] [auto] When no config file exists, one is created at `~/.config/gloggy/config.toml` with default values
- [ ] [auto] The created default config file is valid TOML
- [ ] [auto] When the config file exists, its values are loaded and used
**Dependencies:** none

### R2: Invalid Config Handling
**Description:** If the config file contains invalid TOML or invalid values, the application warns the user and falls back to default values for the invalid portions. The application never crashes due to config errors.
**Acceptance Criteria:**
- [ ] [auto] Given a config file with invalid TOML syntax, the application starts with default values and produces a warning
- [ ] [auto] Given a config file with a valid TOML structure but an invalid value (e.g. negative pane height), the invalid value falls back to its default
- [ ] [auto] The application does not crash or exit due to any config file content
**Dependencies:** R1

### R3: Forward Compatibility
**Description:** Unknown keys in the config file are ignored and preserved. They do not cause errors or warnings.
**Acceptance Criteria:**
- [ ] [auto] A config file containing keys not defined by the application loads without error
- [ ] [auto] Unknown keys are not removed when the config file is rewritten by the application
**Dependencies:** R1

### R4: Theme Selection
**Description:** The config specifies which theme is active. Three bundled themes are available: `tokyo-night` (default), `catppuccin-mocha`, and `material-dark`. Each theme defines color tokens for: level badges (error, warn, info, debug), syntax highlighting (key, string, number, boolean, null), marks, dim lines, and search highlights.
**Acceptance Criteria:**
- [ ] [auto] The default config specifies `tokyo-night` as the active theme
- [ ] [auto] Setting `theme = "catppuccin-mocha"` in config causes that theme's color tokens to be active
- [ ] [auto] Setting `theme = "material-dark"` in config causes that theme's color tokens to be active
- [ ] [auto] Each bundled theme defines color tokens for all required categories: level badges (error, warn, info, debug), syntax highlighting (key, string, number, boolean, null), marks, dim, and search highlight
- [ ] [auto] Specifying an unknown theme name falls back to `tokyo-night` with a warning
- [ ] [human] One-time visual sign-off per bundled theme: all color tokens produce a coherent, readable theme when applied together
**Dependencies:** R1, R2

### R5: Field and Display Settings
**Description:** Config controls: which fields appear in the compact list row (default: time, level, logger, msg), which extra fields appear as sub-rows (default: none), which fields are hidden in the detail pane (default: none), logger abbreviation depth (default: 2), and detail pane height ratio (default: 0.30).
**Acceptance Criteria:**
- [ ] [auto] The default compact row fields are time, level, logger, and msg
- [ ] [auto] Setting sub-row fields in config causes those fields to appear as sub-rows in the entry list
- [ ] [auto] Setting hidden fields in config causes those fields to be omitted from the detail pane
- [ ] [auto] The default logger abbreviation depth is 2
- [ ] [auto] The default detail pane height ratio is 0.30
- [ ] [auto] Each of these settings can be overridden in the config file and the new values take effect
**Dependencies:** R1

### R6: Live Write-Back
**Description:** Interactive changes made during a session (e.g. hiding a field in the detail pane, resizing the pane) are written to the config file immediately, so they persist for future sessions.
**Acceptance Criteria:**
- [ ] [auto] When a field is hidden interactively in the detail pane, the config file is updated to reflect the change
- [ ] [auto] After interactive write-back, the config file remains valid TOML
- [ ] [auto] Existing config values not affected by the change are preserved
**Dependencies:** R1, R3

### R7: Extensibility
**Description:** The config schema is designed so that future features can add new keys without restructuring the existing schema. Existing keys remain stable.
**Acceptance Criteria:**
- [ ] [auto] Adding a new top-level key or section to the config does not require changing the schema of existing keys
- [ ] [auto] A config file written by the current version can be read by a future version that adds new keys (forward-compatible by R3)
**Dependencies:** R3

## Out of Scope

- Key binding remapping (v2)
- Multiple named configuration profiles
- Filter presets stored in config (v2)
- Theme authoring or custom theme files
- Config file format other than TOML

## Cross-References

- See also: cavekit-entry-list.md (reads field visibility, sub-row fields, logger depth, theme)
- See also: cavekit-detail-pane.md (reads/writes field visibility, reads theme and pane height)
- See also: cavekit-app-shell.md (reads theme)
- See also: cavekit-filter-engine.md (no direct dependency, but future filter presets may use config)

## Changelog
