---
created: "2026-04-15T00:00:00Z"
last_edited: "2026-04-19T19:45:00+03:00"
---
# Implementation Tracking: config

Build site: context/plans/build-site.md

| Task | Status | Notes |
|------|--------|-------|
| T-006 | DONE | Config struct, DefaultConfig, Load, DefaultConfigPath |
| T-007 | DONE | TOML parse error → warn+defaults; per-field fallback via validateConfig |
| T-008 | DONE | rawData preserved in LoadResult; Save merges struct into raw map |
| T-009 | DONE | Theme struct + 3 themes in internal/theme/theme.go |
| T-010 | DONE | Default values: theme=tokyo-night, fields=[time,level,logger,msg], depth=2, ratio=0.30 |
| T-024 | DONE | writeback_test.go — HiddenFieldsUpdatePersists, ProducesValidTOML (Save already existed) |
| T-025 | DONE | writeback_test.go — PreservesUnknownKeys, UnknownKeyDoesNotError (rawData mechanism already existed) |
| T-078 | DONE | CursorHighlight, HeaderBg, FocusBorder added to Theme; values for all 3 themes; test verifies non-empty |
| T-084 | DONE | DividerColor + UnfocusedBg theme tokens; distinct-from-Dim/FocusBorder assertion; values across 3 themes |
| T-085 | DONE | DetailPane.WidthRatio/Position/OrientationThresholdCols/WrapMode fields; defaults 0.30/auto/100/soft; enum validation |
| T-086 | DONE | Ratio independence: Save preserves both height/width keys; regression tests in writeback_test.go for both directions |
| T-130 | DONE | Scrolloff int, default 5, top-level key; missing→default, negative→0 warn; round-trips via Save |
| T-171 | DONE | DragHandle lipgloss.Color field on Theme (Tier 23 kit revision ed91d17). Mid-tone neutral populated per bundled theme: tokyo-night `#5a6475` (between DividerColor `#3b4261` and FocusBorder `#7aa2f7`); catppuccin-mocha `#6e7388` (between `#313244`/`#89b4fa`); material-dark `#65737e` (between `#37474f`/`#82aaff`). `theme_test.go` extended: per-theme non-empty DragHandle check (config R4 AC 9) + distinctness from DividerColor and FocusBorder (config R4 AC 10). Fan-in for T-172 (right divider) + T-173 (below top border). Closes cavekit-config.md R4 new AC 9 + AC 10. |
| T-175 | DONE | HUMAN sign-off via tui-mcp across all 3 bundled themes × both orientations (6 combinations) after resolving tui-mcp harness failure — root cause was node-pty's `spawn-helper` losing its executable bit (`-rw-r--r--`) on fresh npx-cache reinstall; `chmod +x ~/.npm/_npx/*/node_modules/node-pty/prebuilds/darwin-arm64/spawn-helper` restored spawning. (The `com.apple.provenance` xattr is benign metadata, not quarantine.) Screenshots captured for each combination: **tokyo-night 140x35 right** — `│` divider visibly mid-grey (`#5a6475`), clearly distinct from list's bright FocusBorder (`#7aa2f7`) and detail pane's very-dim DividerColor (`#3b4261`); **tokyo-night 80x24 below** — detail pane's top-border row (seam) mid-grey, distinct from its own dim left/right/bottom borders. **catppuccin-mocha 140x35 right** + **80x24 below** — same visual pattern with DragHandle `#6e7388` between FocusBorder `#89b4fa` and DividerColor `#313244`. **material-dark 140x35 right** + **80x24 below** — DragHandle `#65737e` between FocusBorder `#82aaff` and DividerColor `#37474f`. All three themes: drag seam reads as clearly-brighter-than-unfocused-border, clearly-dimmer-than-focused-border mid-tone neutral per the human AC. Belt-and-braces `TestDragHandle_LuminanceOrdering_AllThemes` in `internal/theme/theme_test.go` pins the objective WCAG-luminance invariant (tokyo-night Y=0.057/0.126/0.367, catppuccin-mocha Y=0.065/0.174/0.449, material-dark Y=0.055/0.165/0.407; all gaps ≫ 0.02 threshold) so future palette tuning cannot regress the ordering. Closes cavekit-config.md R4 AC 11 (human) + cavekit-app-shell.md R10 AC 4 re-confirm + R15 AC 16 (human). |
