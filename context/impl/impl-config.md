---
created: "2026-04-15T00:00:00Z"
last_edited: "2026-04-17T22:32:51+03:00"
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
