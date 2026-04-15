---
created: "2026-04-15T00:00:00Z"
last_edited: "2026-04-15T00:00:00Z"
---
# Implementation Tracking: detail-pane

Build site: context/plans/build-site.md

| Task | Status | Notes |
|------|--------|-------|
| T-035 | DONE | RenderJSON() — indented, syntax-highlighted, ordered (known fields first then alpha), hidden fields skipped |
| T-036 | DONE | RenderRaw() — plain string(entry.Raw), no ANSI |
| T-037 | DONE | ui/detailpane/scroll.go — ScrollModel j/k/mouse-wheel, boundary clamping, View() |
| T-038 | DONE | ui/detailpane/visibility.go — VisibilityModel with ToggleField() + config.Save() writeback |
| T-041 | DONE | detailpane/model.go — PaneModel Open/Close, Esc/Enter emit BlurredMsg |
| T-042 | DONE | detailpane/height.go — HeightModel ratio-based height, +/- keys, resize via WindowSizeMsg |
| T-043 | DONE | detailpane/search.go — SearchModel /, n/N navigation, HighlightLines(), wrap indicator |
| T-045 | DONE | detailpane/fieldclick.go — fieldAtLine() parser, FieldClickMsg on left-click |
