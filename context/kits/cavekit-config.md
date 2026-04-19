---
created: "2026-04-15T00:00:00Z"
last_edited: "2026-04-19T19:42:23Z"
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
**Description:** The config specifies which theme is active. Three bundled themes are available: `tokyo-night` (default), `catppuccin-mocha`, and `material-dark`. Each theme defines color tokens for: level badges (error, warn, info, debug), syntax highlighting (key, string, number, boolean, null), marks, dim lines, search highlights, cursor highlight (background color for the selected row), header background, focus border/accent, divider color (right-split divider and the border of any unfocused visible pane), unfocused background (the dim tint painted behind an unfocused pane), and `BaseBg` (the primary pane background). `BaseBg` is painted on all pane backgrounds; the terminal's default background is not used for rendered pane surfaces. The two focus-state tokens (DividerColor, UnfocusedBg) play the roles defined in DESIGN.md §2. A third focus-neutral token, `DragHandle`, colours the pane-resize drag seam (DESIGN.md §4.5 and cavekit-app-shell R15) — distinct from `DividerColor` so the draggable seam is visually spottable against unfocused pane borders. Each theme's color tokens trace to a documented canonical upstream source. The theme constructor cites the source URL and — where applicable — the variant name (e.g. tokyo-night variant vs. storm).
**Acceptance Criteria:**
- [ ] [auto] The default config specifies `tokyo-night` as the active theme
- [ ] [auto] Setting `theme = "catppuccin-mocha"` in config causes that theme's color tokens to be active
- [ ] [auto] Setting `theme = "material-dark"` in config causes that theme's color tokens to be active
- [ ] [auto] Each bundled theme defines color tokens for all required categories: level badges (error, warn, info, debug), syntax highlighting (key, string, number, boolean, null), marks, dim, search highlight, cursor highlight, header background, and focus border
- [ ] [auto] Specifying an unknown theme name falls back to `tokyo-night` with a warning
- [ ] [human] One-time visual sign-off per bundled theme: all color tokens produce a coherent, readable theme when applied together
- [ ] [auto] Each bundled theme defines non-empty DividerColor and UnfocusedBg tokens
- [ ] [human] DividerColor reads as a quiet neutral (closer to Dim than to FocusBorder) and UnfocusedBg is a subtle background tint (per DESIGN.md §2)
- [ ] [auto] Each bundled theme defines a non-empty `DragHandle` token
- [ ] [auto] In every bundled theme, `DragHandle != DividerColor` (the drag seam must be visually distinct from unfocused pane borders) and `DragHandle != FocusBorder` (the drag seam must not compete with focus signalling)
- [ ] [human] `DragHandle` reads as a mid-tone neutral — clearly brighter than `DividerColor` on the theme's background but dimmer than `FocusBorder` (per DESIGN.md §2)
- [ ] [auto] Each bundled theme defines a non-empty `BaseBg` token
- [ ] [auto] `BaseBg` is painted on pane backgrounds wherever an `UnfocusedBg` overlay does not apply — no rendered pane falls through to the terminal's default background
- [ ] [auto] Across the three bundled themes, `BaseBg` values are pairwise distinct
- [ ] [auto] Within each bundled theme, `BaseBg != UnfocusedBg` (the primary pane fill must not collapse into the unfocused tint)
- [ ] [auto] Each theme constructor includes a canonical-source citation (upstream URL + variant name where applicable), discoverable at test time (e.g. as a comment or named constant adjacent to the constructor)
- [ ] [auto] Catppuccin-mocha's `FocusBorder` equals the upstream Lavender value (`#b4befe`), not Blue
- [ ] [human] One-time sign-off per bundled theme: the cited canonical source's palette is faithfully reflected in the theme's tokens (augments the existing coherence sign-off; does not replace it)
- [ ] [human] At a glance, the three bundled themes are perceptibly distinct from each other — not merely shade drift of the same dark palette
**Dependencies:** R1, R2

### R5: Field and Display Settings
**Description:** Config controls: which fields appear in the compact list row (default: time, level, logger, msg), which extra fields appear as sub-rows (default: none), which fields are hidden in the detail pane (default: none), logger abbreviation depth (default: 2), detail pane height ratio (default: 0.30), detail pane orientation position (`below` | `right` | `auto`, default `auto`), auto-orientation threshold in columns (default 100), detail pane width ratio for right-split mode (default 0.30), wrap mode for detail pane content (default `soft`), and a **shared top-level `scrolloff`** (default 5) honoured by both the entry list and the detail pane (see cavekit-entry-list.md R12, cavekit-detail-pane.md R11, DESIGN.md §4.3 + §4.4). The two pane ratios are independent settings — flipping orientation must never overwrite one with the other. `scrolloff` is a single top-level key so users tune one "context rows" value for both scrolling panes; do NOT split it into `entry_list.scrolloff` + `detail_pane.scrolloff`.
**Acceptance Criteria:**
- [ ] [auto] The default compact row fields are time, level, logger, and msg
- [ ] [auto] Setting sub-row fields in config causes those fields to appear as sub-rows in the entry list
- [ ] [auto] Setting hidden fields in config causes those fields to be omitted from the detail pane
- [ ] [auto] The default logger abbreviation depth is 2
- [ ] [auto] The default detail pane height ratio is 0.30
- [ ] [auto] Each of these settings can be overridden in the config file and the new values take effect
- [ ] [auto] The default config includes detail_pane.position = "auto"
- [ ] [auto] detail_pane.orientation_threshold_cols defaults to 100
- [ ] [auto] detail_pane.width_ratio defaults to 0.30
- [ ] [auto] detail_pane.wrap_mode defaults to "soft"
- [ ] [auto] detail_pane.height_ratio and detail_pane.width_ratio are preserved independently across orientation flips — changing one does not overwrite the other
- [ ] [auto] The default top-level `scrolloff` is 5
- [ ] [auto] `scrolloff` is exposed as a top-level TOML int key (not nested under `entry_list` or `detail_pane`)
- [ ] [auto] An integer `scrolloff` value in config file is read by both the entry-list model and the detail-pane model (same source of truth)
- [ ] [auto] A negative or non-integer `scrolloff` is clamped to 0 at load time with a warning, per R2 (invalid config handling)
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
- See also: cavekit-app-shell.md (reads theme; pane rendering honours `theme.BaseBg` per R4)
- See also: cavekit-filter-engine.md (no direct dependency, but future filter presets may use config)

## Changelog

- 2026-04-19: R4 extended for theme palette fidelity — added BaseBg token, canonical-source citation AC, catppuccin Lavender FocusBorder invariant, perceptual-distinctness sign-off. Grounded by context/refs/research-brief-theme-palettes.md.

### 2026-04-19 — Revision (DragHandle token)
- **Affected:** R4
- **Summary:** R4 extended to require a third focus-state-adjacent theme token, `DragHandle`, colouring the pane-resize drag seam distinctly from `DividerColor`. ACs enforce non-empty per-theme value, distinctness from both `DividerColor` and `FocusBorder`, and a human sign-off on mid-tone-neutral readability. Companion to cavekit-app-shell R10/R15 revision and DESIGN.md §2 / §4.5 edits.
- **Driven by:** /ck:sketch session 2026-04-19 — user reported the right-split divider was indistinguishable from unfocused pane borders, making the drag target hard to spot.

### 2026-04-16 — Revision
- **Affected:** R4
- **Summary:** R4 updated to require three new theme color tokens: cursor highlight (background for selected row), header background, and focus border/accent. All bundled themes must define these tokens. Driven by user observation that cursor row, header bar, and pane focus have no visual distinction.
- **Commits:** manual testing feedback (no commit)

### 2026-04-17 — Revision (details-pane redesign)
- **Affected:** R4, R5
- **Summary:** R4 extended to require DividerColor and UnfocusedBg tokens in every bundled theme, supporting the pane visual-state matrix. R5 extended to cover the new detail pane orientation keys (position, orientation_threshold_cols, width_ratio, wrap_mode) and to guarantee height_ratio and width_ratio are preserved independently across orientation flips.
- **Driven by:** DESIGN.md + research-brief-details-pane-redesign.md

### 2026-04-18 — Revision (shared top-level `scrolloff`)
- **Affected:** R5
- **Summary:** R5 extended to define a shared top-level `scrolloff` int key (default 5) consumed by both the entry list and the detail pane. Single source of truth for the "context rows between cursor and edge" setting — must NOT be split into two pane-specific keys. Negative values clamped to 0 at load time per R2. Companion to cavekit-entry-list.md R12 + cavekit-detail-pane.md R11 + DESIGN.md §4.3 "Shared scrolloff".
- **Driven by:** `/ck:check` run 2026-04-18 after user report on row highlight + scrolloff behaviour.
