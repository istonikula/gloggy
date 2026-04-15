---
created: "2026-04-15T00:00:00Z"
last_edited: "2026-04-15T21:49:00Z"
---
# Implementation Tracking: filter-engine

Build site: context/plans/build-site.md

| Task | Status | Notes |
|------|--------|-------|
| T-011 | DONE | Filter struct, FilterSet with Add/Remove/Enable/Disable/GetAll/GetEnabled; Mode as typed enum |
| T-018 | DONE | match.go — Match() with regex/literal dispatch; entryFieldValue handles known fields + Extra |
| T-026 | DONE | FilterSet.ToggleAll() — global disable saves/restores per-filter Enabled state |
| T-019 | DONE | filter/index.go — Apply() include/exclude logic; disabled filters ignored |
| T-020 | DONE | FilteredIndex type with Recompute() |
| T-039 | DONE | ui/filter/panel.go — Bubble Tea model: j/k nav, Space toggle, d delete, mouse click; FilterChangedMsg |
| T-044 | DONE | ui/filter/prompt.go — PromptModel: pre-fill field/pattern, Tab toggle mode, Enter confirm, Esc cancel |
| T-069 | DONE | Guard nil Extra map in entryFieldValue to prevent panic |
| T-071 | DONE | Regex cache via sync.Map in cachedRegexp(); 0 alloc/op on benchmark |
| T-074 | DONE | ToggleAll tracks saved state by filter ID (map[int]bool); Add/Remove sync |
| T-077 | DONE | JSON unquoting via json.Unmarshal for escaped strings in Extra values |
