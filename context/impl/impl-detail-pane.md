---
created: "2026-04-15T00:00:00Z"
last_edited: "2026-04-18T00:10:55+03:00"
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
| T-082 | DONE | Top border separator via NormalBorder BorderTop + FocusBorder color; test verifies "─" |
| T-100 | DONE | PaneModel.View uses appshell.PaneStyle(state); unfocused → DividerColor border + UnfocusedBg + Faint; focused → FocusBorder |
| T-103 | DONE | Top border verified in both orientations (right + below); lipgloss.Width scan over first View line |
| T-107 | DONE | PaneModel uses lipgloss.Width via styling; SetWidth(w) caps outer with Width(w-2).MaxWidth(w); emoji+CJK+ANSI tests |
