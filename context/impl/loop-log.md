---
created: "2026-04-15T00:00:00Z"
last_edited: "2026-04-18T00:10:55+03:00"
---
# Loop Log

### Iteration 18 — 2026-04-18
- T-105: orientation-flip preserves both ratios — DONE. Files: app/model_test.go (+TestModel_OrientationFlip_PreservesBothRatios). Build P, Tests P.
- T-108: resize re-evals orientation + preserves ratios — DONE. Files: appshell/resize_test.go (+TestResizeModel_AutoFlipPreservesBothRatios). Build P, Tests P.
- T-106: detail pane soft wrap — DONE. Files: detailpane/wrap.go (new, ansi.HardwrapWc), wrap_test.go (new, 8 tests), detailpane/model.go (rawContent + SoftWrap on Open/SetWidth + borderRows fixed to 2). Build P, Tests P.
- Tier 9 complete. Next: Codex tier gate review for tier 9, then Tier 10 HUMAN sign-off (T-109/T-110 via tui-mcp).

### Iteration 17 — 2026-04-18
- T-095: click-to-focus on panes — DONE. Files: app/model.go handleMouse Press+Left zone-switch, model_test.go (+3 tests). Build P, Tests P.
- T-099: ratio live write-back — DONE. Files: app/model.go saveConfig + ratio key + drag release hooks, model_test.go (+2 tests). Build P, Tests P. Disk persisted ratio = in-memory ratio; height_ratio untouched on width_ratio writes.
- Next: T-105 (orientation-flip preserves ratios), T-106 (soft wrap), T-108 (resize extension).

### Iteration 16 — 2026-04-17
- T-094: right-split horizontal mouse zoning — DONE. Files: appshell/mouse.go (Zone() right-split branch with listEnd/divider/detailStart buffers), mouse_test.go (+2 tests).
- T-100: PaneStyle full-border DividerColor+UnfocusedBg+Faint when unfocused — DONE. Files: appshell/panestyle.go (new), entrylist/list.go (Focused/Alone fields + applyPaneStyle), detailpane/model.go (PaneStyle wiring).
- T-101: Alone forces focused treatment — DONE. Files: entrylist/list.go list_test.go (+TestView_Alone_UsesFocusedTreatment).
- T-102: cursor row keeps highlight unfocused — DONE. Files: entrylist/list.go View, list_test.go (TestView_CursorHighlight updated).
- T-103: detail pane top border in both orientations — DONE. Files: detailpane/model_test.go (TestPaneModel_TopBorder_InBothOrientationContexts via lipgloss.Width scan).
- T-104: divider-drag resizes width_ratio — DONE. Files: appshell/ratiokeys.go (RatioFromDragX), app/model.go (drag state machine Press→Motion→Release), model_test.go (TestModel_DividerDrag_UpdatesWidthRatio).
- T-107: lipgloss.Width-safe pane measurement — DONE. Files: detailpane/model.go (SetWidth + Width(w-2).MaxWidth(w) via PaneStyle), model_test.go (+emoji/CJK/ANSI tests). Wave commits 3a339e7, fa1cd66, b2a6c49.

### Iteration 12 — 2026-04-17
- T-084: DividerColor + UnfocusedBg theme tokens — DONE. Files: internal/theme/theme.go, theme_test.go. Build P, Tests P.
- T-085: DetailPane orientation config fields (width_ratio, position, threshold, wrap_mode) — DONE. Files: internal/config/config.go, config_test.go. Build P, Tests P.
- T-090: 60x15 minimum-viable terminal floor + fallback — DONE. Files: internal/ui/appshell/layout.go, layout_test.go. Build P, Tests P.
- T-093: Header narrow-mode drop-priority degradation — DONE. Files: internal/ui/appshell/header.go, header_test.go. Build P, Tests P.
- T-096: Tab focus cycle (NextFocus pure fn + app-shell wiring) — DONE. Files: internal/ui/appshell/focus.go (new), focus_test.go (new), internal/ui/app/model.go, model_test.go. Build P, Tests P.
- T-097: Esc priority chain + list ClearTransient — DONE. Files: internal/ui/entrylist/list.go, leveljump_test.go, internal/ui/app/model.go, model_test.go. Build P, Tests P. Wave commit ebedb3a.

### Iteration 13 — 2026-04-17
- T-086: ratio independence regression tests — DONE. Files: internal/config/writeback_test.go. Build P, Tests P.
- T-087: SelectOrientation + ResizeModel.WithConfig — DONE. Files: internal/ui/appshell/orientation.go (new), orientation_test.go (new), resize.go, internal/ui/app/model.go. Build P, Tests P.
- T-092: keyhints focus label (right-aligned, Bold+FocusBorder) — DONE. Files: internal/ui/appshell/keyhints.go, keyhints_test.go, internal/ui/app/model.go. Build P, Tests P. Wave commit df8d806.

### Iteration 14 — 2026-04-17
- T-088: right-split composition — DONE. Files: internal/ui/appshell/layout.go (Orientation+WidthRatio fields, ListContentWidth/DetailContentWidth, JoinHorizontal Render branch + inline divider stub), layout_test.go (+5 tests), internal/ui/app/model.go (WindowSizeMsg + relayout wire orientation/width ratio + use ListContentWidth). Build P, Tests P.
- Next: T-088 unblocks T-089, T-091, T-094, T-098, T-100, T-103, T-107 in tier 9.

### Iteration 15 — 2026-04-17
- T-089: vertical divider │ in DividerColor via JoinHorizontal — DONE. Files: appshell/divider.go (new), divider_test.go (new), layout.go renderInlineDivider. Build P, Tests P.
- T-098: ratio keymap +/-/=/| presets clamped [0.10,0.80] — DONE. Files: appshell/ratiokeys.go (new), ratiokeys_test.go (new), detailpane/height.go (SetRatio + cap), app/model.go handleKey. Build P, Tests P.
- T-091: auto-close pane on minimum underflow + 3s notice — DONE. Files: appshell/autoclose.go (new), autoclose_test.go (new), keyhints.go (notice + WithNotice/HasNotice), app/model.go (noticeClearMsg + tea.Tick + WindowSizeMsg wire), model_test.go (+3 tests). Build P, Tests P. Wave commit 70c5354.
- Next: T-094, T-100, T-103, T-107 unblocked in tier 9.

### Iteration 11 — 2026-04-16
- T-078: Theme tokens CursorHighlight/HeaderBg/FocusBorder — DONE. Files: internal/theme/theme.go, theme_test.go. Build P, Tests P.
- T-080: CursorPosition() on ListModel — DONE. Files: internal/ui/entrylist/list.go, list_test.go. Build P, Tests P.
- T-079: Cursor row highlight — DONE. Files: internal/ui/entrylist/list.go, list_test.go. Build P, Tests P.
- T-081: Header bg+bold+cursor pos — DONE. Files: internal/ui/appshell/header.go, header_test.go. Build P, Tests P.
- T-082: Detail pane top border — DONE. Files: internal/ui/detailpane/model.go, model_test.go. Build P, Tests P.
- T-083: Focus indicator on panes — DONE. Files: internal/ui/app/model.go, detailpane/model.go. Build P, Tests P.
- All Tier 8 tasks DONE. Next: completion.

### Iteration 10 — 2026-04-16
- T-069,T-071,T-074,T-077: Filter engine fixes — DONE. Files: internal/filter/{match,filter}.go + tests. Build P, Tests P.
- T-070,T-073: Log source fixes — DONE. Files: internal/logsource/{reader,loader,tail}.go + tests, cmd/gloggy/main.go. Build P, Tests P.
- T-072,T-075,T-076: UI fixes — DONE. Files: internal/ui/app/model.go, internal/ui/entrylist/list.go, tests/integration/tail_test.go. Build P, Tests P.
- Next: Tier 6 HUMAN sign-off (T-061..T-068) — cannot be automated

### Iteration 9 — 2026-04-15
- T-055..T-060: Tier 5 integration tests — DONE. Files: tests/integration/*.go. Build P, Tests P. Next: T-061..T-068 (HUMAN sign-off)

### Iteration 8 — 2026-04-15
- T-052,T-053,T-054: MouseRouter, ResizeModel, CopyMarkedEntries — DONE. Files: appshell/mouse.go, resize.go, clipboard.go + tests, go.mod. Build P, Tests P.

### Iteration 7 — 2026-04-15
- T-047..T-051: HeaderModel, LoadingModel, ParseArgs CLI, KeyHintBarModel, HelpOverlayModel — DONE. Files: appshell/header.go, loading.go, keyhints.go, helpoverlay.go + tests, cmd/gloggy/main.go. Build P, Tests P.
- T-044,T-046: PromptModel (filter add), LayoutModel (appshell) — DONE. Files: ui/filter/prompt.go, appshell/layout.go + tests. Build P, Tests P.
- T-045: FieldClickMsg, fieldAtLine, handleMouseClick — DONE. Files: detailpane/fieldclick.go + tests. Build P, Tests P.

### Iteration 6 — 2026-04-15
- T-042,T-043: HeightModel, SearchModel — DONE. Files: detailpane/height.go, search.go + tests. Build P, Tests P.
- T-041: PaneModel activation/dismissal — DONE. Files: detailpane/model.go + tests. Build P, Tests P.

### Iteration 5 — 2026-04-15
- T-040: Mouse handling in ListModel — DONE. Files: entrylist/list.go. Build P, Tests P.

### Iteration 4 — 2026-04-15
- T-029,T-030: ListModel virtual rendering, CursorModel two-level nav — DONE. Files: list.go, cursor.go + tests. Build P, Tests P.
- T-031,T-032,T-033,T-034: filtered view, level-jump, marks, selection signal — DONE. Files: leveljump.go, list.go updated. Build P, Tests P. Next: Tier 3 (T-040..T-048)

### Iteration 3 — 2026-04-15
- T-019,T-020: Apply()/FilteredIndex — DONE. Files: filter/index.go, index_test.go. Build P, Tests P.
- T-022,T-023: RenderCompactRow/level badge colors — DONE. Files: entrylist/row.go, row_test.go. Build P, Tests P.
- T-027,T-028: LoadFile/TailFile — DONE. Files: logsource/loader.go, tail.go + tests, go.mod. Build P, Tests P.
- T-037,T-038,T-039: ScrollModel/VisibilityModel/filter panel — DONE. Files: detailpane/scroll.go, visibility.go, filter/panel.go + tests. Build P, Tests P. Next: T-029..T-034 (Tier 2 remaining)

### Iteration 2 — 2026-04-15
- T-015,T-016,T-017: ReadFile/ReadStdin/tests — DONE. Files: internal/logsource/reader.go, reader_test.go. Build P, Tests P.
- T-018: filter/match.go Match() — DONE. Files: internal/filter/match.go, match_test.go. Build P, Tests P.
- T-026: FilterSet.ToggleAll() — DONE. Files: internal/filter/filter.go. Build P, Tests P.
- T-021: AbbreviateLogger — DONE. Files: internal/ui/entrylist/logger.go, logger_test.go. Build P, Tests P.
- T-024,T-025: config writeback tests — DONE. Files: internal/config/writeback_test.go. Build P, Tests P.
- T-035,T-036: RenderJSON/RenderRaw — DONE. Files: internal/ui/detailpane/render.go, render_test.go. Build P, Tests P (11/11). Next: T-019,T-020,T-022,T-023,T-027..T-034,T-037..T-039 (Tier 1-2)

### Iteration 1 — 2026-04-15
- T-001..T-014: Tier 0 — DONE. Files: go.mod, cmd/gloggy/main.go, internal/logsource/{entry,classify,parse}.go, internal/config/config.go, internal/theme/theme.go, internal/filter/filter.go, internal/ui/entrylist/{scroll,marks}.go, internal/ui/appshell/help.go + tests. Build P, Tests P (all packages). Next: T-015..T-026 (Tier 1)
