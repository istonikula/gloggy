---
created: "2026-04-15T00:00:00Z"
last_edited: "2026-04-16T20:09:25+03:00"
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
