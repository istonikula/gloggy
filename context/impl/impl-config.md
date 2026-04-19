---
created: "2026-04-15T00:00:00Z"
last_edited: "2026-04-19T12:30:00+03:00"
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
