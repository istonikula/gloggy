---
created: "2026-04-15T00:00:00Z"
last_edited: "2026-04-18T00:35:07+03:00"
---
# Implementation Tracking: entry-list

Build site: context/plans/build-site.md

| Task | Status | Notes |
|------|--------|-------|
| T-012 | DONE | GoTop/GoBottom/HalfPageDown/HalfPageUp pure functions on ScrollState |
| T-013 | DONE | MarkSet keyed by entry ID; Toggle/IsMarked/NextMark/PrevMark with wrap |
| T-021 | DONE | AbbreviateLogger() in logger.go — keeps last depth segments full, abbreviates earlier to first char |
| T-022 | DONE | ui/entrylist/row.go — RenderCompactRow() with time/level/logger/msg; non-JSON dim; zero-time placeholder |
| T-023 | DONE | Level badge colors via theme tokens; colorANSI() helper avoids termenv hex-rounding diff |
| T-029 | DONE | list.go — ListModel virtual rendering; only offset±renderBuffer rows rendered; SelectionMsg on move |
| T-030 | DONE | cursor.go — CursorModel two-level nav; j/k event level, l/Tab/→ sub-row, h/←/Esc exit; SubRows() |
| T-031 | DONE | ListModel.SetFilter() — wires FilteredIndex; cursor preserved if passing else nearest |
| T-032 | DONE | leveljump.go — NextLevel/PrevLevel with WrapDirection; e/E/w/W keys; WrapDir() indicator |
| T-033 | DONE | 'm' toggles mark via MarkSet; 'u'/'U' next/prev mark; '* ' visual indicator in View() |
| T-034 | DONE | SelectionMsg emitted on every cursor movement in ListModel.Update() |
| T-040 | DONE | Mouse handling: left-click selects row, wheel scrolls, re-click on selected → OpenDetailPaneMsg |
| T-075 | DONE | Mark indicator uses lipgloss.NewStyle().Foreground(th.Mark) instead of plain "* " |
| T-076 | DONE | Timestamp-based double-click: lastClickRow+lastClickTime, 500ms window, resets after trigger |
| T-079 | DONE | Cursor row highlight via CursorHighlight background in View(); test verifies ANSI styling |
| T-080 | DONE | CursorPosition() returns 1-based index; 0 when empty; tests for j/k, g/G, filter |
| T-100 | DONE | ListModel.View applies appshell.PaneStyle via Focused/Alone fields; full DividerColor border + UnfocusedBg + Faint when unfocused-with-pane |
| T-101 | DONE | Alone field forces focused treatment when pane closed; tests TestView_Alone_UsesFocusedTreatment |
| T-102 | DONE | Cursor row keeps CursorHighlight bg when unfocused; non-Bold; tests confirm ANSI bg present in both states |
| T-061 | DONE | HUMAN sign-off via tui-mcp on small.log: tokyo-night INFO #7aa2f7, WARN #e0af68, ERROR #f7768e; JSON syntax keys teal #73daca, strings green #9ece6a — coherent and readable |
| T-062 | DONE | HUMAN sign-off via tui-mcp on tiny.log: catppuccin-mocha INFO #89b4fa, WARN #f9e2af; JSON keys teal #94e2d5, strings green #a6e3a1 — coherent |
| T-063 | DONE | HUMAN sign-off via tui-mcp on tiny.log: material-dark INFO #82aaff, WARN #ffcb6b; JSON keys cyan #89ddff, strings green #c3e88d — coherent |
| T-064 | DONE | HUMAN sign-off via tui-mcp on small.log tokyo-night: rows 1-28 (logback `\|-INFO`/`\|-WARN` raw text) render visibly dimmer than JSON rows starting at line 29 (`23:39:10 INFO o.h.v.i...` with bright colored badge) |
| T-065 | DONE | HUMAN sign-off via tui-mcp on small.log: row spacing + alternating raw/JSON sections show clear event boundaries; default config single-row entries adequately separated |
| T-066 | PARTIAL | HUMAN sign-off via tui-mcp: level-jump `e` correctly navigated to ERROR row 297/300 with cursor highlight visible. R8 #6 (wrap indicator) and #7-8 (filtered-out indicator) NOT visually rendered — `wrapDir` state tracked, exposed via WrapDir/HasTransient, only consumed by Esc-clear in app/model.go:324. View() does not draw any indicator. GAP — see new task in build-site. |
| T-067 | PARTIAL | HUMAN sign-off via tui-mcp: mark `m` toggles `*` indicator (visible on cursor row); `u`/`U` navigate marks; wrap from line 297→1 occurred but R9 #5 wrap indicator NOT rendered (same root cause as T-066). GAP — see new task in build-site. |
| T-068 | DONE | Detail pane syntax highlighting per theme verified during T-061..T-063 walks: tokyo-night syntax keys/strings/numbers visible; mocha + material-dark same — JSON renders with theme-distinct colors |
