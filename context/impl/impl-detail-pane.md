---
created: "2026-04-15T00:00:00Z"
last_edited: "2026-04-18T09:40:17+03:00"
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
| T-043 | PARTIAL | detailpane/search.go — `SearchModel` implemented and unit-tested, but `detailpane/model.go.View()` never calls `HighlightLines` and renders no search prompt. `/` from entry-list focus is a silent no-op. Downgraded 2026-04-18 by `/ck:check` (F-002, F-004, F-005). Integration closed by T-113..T-122. |
| T-045 | DONE | detailpane/fieldclick.go — fieldAtLine() parser, FieldClickMsg on left-click |
| T-082 | DONE | Top border separator via NormalBorder BorderTop + FocusBorder color; test verifies "─" |
| T-100 | DONE | PaneModel.View uses appshell.PaneStyle(state); unfocused → DividerColor border + UnfocusedBg + Faint; focused → FocusBorder |
| T-103 | DONE | Top border verified in both orientations (right + below); lipgloss.Width scan over first View line |
| T-107 | DONE | PaneModel uses lipgloss.Width via styling; SetWidth(w) caps outer with Width(w-2).MaxWidth(w); emoji+CJK+ANSI tests |
| T-106 | DONE | wrap.go SoftWrap via ansi.HardwrapWc (ANSI-safe + cell-aware); PaneModel.rawContent + Open/SetWidth re-wrap to contentWidth; borderRows fixed to 2 (top+bottom); 8 wrap tests |
| T-113 | NEW | Expose `PaneModel.ContentLines()` returning soft-wrapped unstyled content; unblocks T-114 search wiring. Closes F-003. |
| T-114 | NEW | Wire `SearchModel` into `PaneModel.View()` via `WithSearch()`; render prompt row with `(cur/total)` and "No matches" state; compose highlights. Closes F-002, F-004, F-010. |
| T-115 | NEW | `PaneModel.ScrollToLine()` + call after `n`/`N`. Closes F-005. |
| T-117 | NEW | Dismiss `paneSearch` on `BlurredMsg` and `openPane`. Closes F-006. |
| T-118 | NEW | Split input-mode vs navigation-mode in `SearchModel`; let `j`/`k`/`g`/`G`/`Ctrl+d`/`Ctrl+u` pass through to pane scroll in navigation-mode. Closes F-008. |
| T-119 | NEW | UTF-8-safe backspace in `SearchModel.Update`. Closes F-009. |
| T-120 | NEW | Integration test for two-step Esc (dismiss search → close pane). Closes F-007. |
| T-122 | NEW | [HUMAN] `/` search end-to-end sign-off via tui-mcp per overview Verification Conventions. |
