---
created: "2026-04-15T00:00:00Z"
last_edited: "2026-04-16T20:09:25+03:00"
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
