---
created: "2026-04-15T00:00:00Z"
last_edited: "2026-04-18T12:07:49+03:00"
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
| T-043 | DONE | Integration closed by T-113..T-122 (2026-04-18): `PaneModel.View()` now consumes `SearchModel`, renders the prompt row with `(cur/total)` / `No matches` / wrap arrows, highlights matches via `HighlightLines(ContentLines())`, reserves a content row, and the app-shell wires cross-pane `/` activation (F-001) + dismissal on pane close (F-006) + input/navigate mode split (F-008) + UTF-8-safe backspace (F-009). Originally downgraded PARTIAL by the 2026-04-18 `/ck:check`; now upgraded back to DONE. |
| T-045 | DONE | detailpane/fieldclick.go — fieldAtLine() parser, FieldClickMsg on left-click |
| T-082 | DONE | Top border separator via NormalBorder BorderTop + FocusBorder color; test verifies "─" |
| T-100 | DONE | PaneModel.View uses appshell.PaneStyle(state); unfocused → DividerColor border + UnfocusedBg + Faint; focused → FocusBorder |
| T-103 | DONE | Top border verified in both orientations (right + below); lipgloss.Width scan over first View line |
| T-107 | DONE | PaneModel uses lipgloss.Width via styling; SetWidth(w) caps outer with Width(w-2).MaxWidth(w); emoji+CJK+ANSI tests |
| T-106 | DONE | wrap.go SoftWrap via ansi.HardwrapWc (ANSI-safe + cell-aware); PaneModel.rawContent + Open/SetWidth re-wrap to contentWidth; borderRows fixed to 2 (top+bottom); 8 wrap tests |
| T-113 | DONE | `PaneModel.ContentLines()` splits `rawContent` by `\n` and ANSI-strips each line via `x/ansi.Strip`. Two tests: no-borders-no-ANSI and closed-returns-nil. Closes F-003. |
| T-114 | DONE | `PaneModel.WithSearch(s)` stores a SearchModel; `View()` reserves 1 content row for a prompt showing `/<query>`, `(cur/total)`, `No matches`, or wrap ↓/↑; active match set drives highlight render from `ContentLines() → HighlightLines()`. App swapped to `pane.ContentLines()` as the match line source. Closes F-002, F-004, F-010. |
| T-115 | DONE | `PaneModel.ScrollToLine(idx)` minimal scroll into window (top or bottom alignment); app calls it after every search update when a match exists (covers typing + n/N + wrap). Closes F-005. |
| T-117 | DONE | `app.Update` dismisses `paneSearch` on `BlurredMsg`; `openPane` dismisses on each new entry. 2 tests verifying query/active cleared across open/close and across entries. Closes F-006. |
| T-118 | DONE | `SearchModel.mode` (Input|Navigate) with `Mode()` accessor; Enter commits input→navigate; `/` re-enters input preserving query; n/N literal-append in input, navigate in navigate; app forwards non-search keys to `pane.Update` when navigate. 8 tests. Closes F-008. |
| T-119 | DONE | Backspace rune-slices `m.query` via `[]rune(m.query)[:len-1]`. Unit test covers ascii / café / 日本語 / 🚀x. Closes F-009. |
| T-120 | DONE | `TestModel_TwoStepEsc_DismissesSearchThenClosesPane` in app/model_test.go — open pane → `/` → type → first Esc (search dismissed, pane stays open); second Esc (pane closes + BlurredMsg + focus returns to list). Closes F-007. |
| T-122 | DONE | HUMAN sign-off via tui-mcp (tokyo-night @ 140x35 right-split, small.log): `/INFO` in detail pane renders `/INFO  (1/1)` prompt row, Enter commits to navigate mode, two-step Esc chain dismisses search then closes pane, cross-pane `/` from list focus transfers to detail pane + activates search in one keystroke, help overlay + keyhint bar show correct scope text. Notice-padding bug (T-121-fix) discovered and fixed during this sign-off. |
| T-123 | DONE | `appshell.DetailPaneVerticalRows(l)` returns full main slot (height-header-status) in right-mode, `l.DetailPaneHeight` in below. Wired into `app.Update` WindowSizeMsg branch + `relayout()` via `m.pane.SetHeight(appshell.DetailPaneVerticalRows(l))`. Fixed `PaneModel.Open/SetWidth` to seed scroll with `ContentHeight()` (not outer `m.height`), `SetHeight` re-clamps via new public `ScrollModel.Clamp()`, `View()` clamps after its local height shrink. 4 layout tests (right-full-slot / below-ratio / closed-zero / floor-one) + 4 app tests (right→≥20 content rows / below→PaneHeight=7 / `+` still adjusts below / orientation flip preserves ratio). Closes F-013 (P0), F-014 (P1), F-018 (P2), F-019 (P2), F-022 (P3). |
| T-124 | DONE | `ScrollModel.Update()` extended: `g`/Home → offset 0; `G`/End → last-line anchored at bottom; PgDn/Ctrl+d/Space → height-1 down; PgUp/Ctrl+u/`b` → height-1 up. 8 tests (g jumps top, G caps bottom, G on short content, PgDn table, PgUp at top no-op, PgUp after End, PgDn clamp at bottom, page-keys on short content). Search-input routing already prevents binding leak (T-118 split). Closes F-015 (P1). |
