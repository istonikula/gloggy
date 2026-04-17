---
created: "2026-04-15T00:00:00Z"
last_edited: "2026-04-18T00:35:07+03:00"
---
# Implementation Tracking: app-shell

Build site: context/plans/build-site.md

| Task | Status | Notes |
|------|--------|-------|
| T-014 | DONE | KeybindingRegistry data structure in internal/ui/appshell/help.go; all 4 domains covered |
| T-046 | DONE | appshell/layout.go — LayoutModel, Layout struct, Render() via lipgloss.JoinVertical |
| T-047 | DONE | appshell/header.go — HeaderModel: source name, FOLLOW badge, entry counts |
| T-048 | DONE | appshell/loading.go — LoadingModel: show/hide indicator with entry count |
| T-049 | DONE | cmd/gloggy/main.go — ParseArgs: file/stdin/follow mode, clear error on invalid args |
| T-050 | DONE | appshell/keyhints.go — KeyHintBarModel: context-sensitive hints by FocusTarget |
| T-051 | DONE | appshell/helpoverlay.go — HelpOverlayModel: ? opens, Esc closes, intercepts keys |
| T-052 | DONE | appshell/mouse.go — MouseRouter: zones header/entrylist/divider/detailpane/statusbar |
| T-053 | DONE | appshell/resize.go — ResizeModel + ApplyToLayout: WindowSizeMsg, proportional pane |
| T-054 | DONE | appshell/clipboard.go — CopyMarkedEntries: JSONL, original order, no-op on empty marks |
| T-072 | DONE | cachedVisibleCount field; updated in refilter() and SetEntries(); no more O(n²) Apply |
| T-081 | DONE | Header bar: HeaderBg background, Bold, WithCursorPos(); cursor/visible + visible/total display |
| T-083 | DONE | Focus indicator: colored left border on focused pane via FocusBorder; updates on focus change |
| T-090 | DONE | layout.go MinTerminalWidth/Height=60/15, IsBelowMinFloor, RenderTooSmall via lipgloss.Place; Render short-circuits |
| T-093 | DONE | header.go drop-priority order focus/counts/cursor/follow; source always kept; truncateToWidth uses lipgloss.Width binary search |
| T-096 | DONE | appshell/focus.go NextFocus pure fn + wiring in app/model.go handleKey; Tab inert on overlay, no-op on single pane, never closes |
| T-097 | DONE | Esc priority chain: help intercept → filter panel self-close → detail pane forward → list ClearTransient (wrap indicator) |
| T-087 | DONE | appshell/orientation.go SelectOrientation; ResizeModel.WithConfig + Orientation; re-eval on every WindowSizeMsg |
| T-088 | DONE | Layout.Orientation+WidthRatio fields; ListContentWidth/DetailContentWidth (DESIGN.md §5 formula); Render right-split branch via JoinHorizontal with inline divider |
| T-092 | DONE | KeyHintBarModel.WithPaneOpen + right-aligned focus label (Bold + FocusBorder); omitted in single-pane state |
| T-089 | DONE | appshell/divider.go RenderDivider (│ glyph, DividerColor); inline join via lipgloss.JoinHorizontal in Render right-split |
| T-091 | DONE | appshell/autoclose.go ShouldAutoCloseDetail (MinDetailWidth=30 right, MinDetailHeight=3 below); KeyHintBarModel.WithNotice; app/model wires noticeClearMsg via tea.Tick(3s) |
| T-098 | DONE | appshell/ratiokeys.go NextRatio +/-/=/| presets [0.10,0.30,0.70] clamped [0.10,0.80]; routed via orientation in handleKey |
| T-094 | DONE | mouse.go right-split horizontal zoning: list/listEnd-buffer/divider/detailStart-buffer/detail; tests at width 100 |
| T-095 | DONE | app/model.handleMouse Press+Left → focus transfers to ZoneEntryList/ZoneDetailPane; tests cover both directions and pane-closed no-op |
| T-099 | DONE | app/model.saveConfig calls config.Save after ratio key + drag release; tests verify width_ratio persists to disk and height_ratio untouched |
| T-100 | DONE | appshell/panestyle.go PaneStyle(state); entrylist + detailpane use full-border DividerColor+UnfocusedBg+Faint when unfocused |
| T-101 | DONE | list.Alone forces focused treatment when pane closed; covered in entrylist visual-state tests |
| T-102 | DONE | entrylist cursor row keeps CursorHighlight bg unfocused; non-Bold weight; tests TestView_CursorHighlight |
| T-103 | DONE | detailpane top border verified in both orientations; lipgloss.Width-safe scan in tests |
| T-104 | DONE | appshell/ratiokeys.RatioFromDragX; app/model drag state machine Press→Motion→Release on right-split divider |
| T-107 | DONE | detailpane uses lipgloss.Width via PaneStyle; outer width matches allocation; emoji/CJK/ANSI tests |
| T-105 | DONE | model_test.go TestModel_OrientationFlip_PreservesBothRatios — right→below→right with height_ratio=0.60 width_ratio=0.20 verifies neither mutated |
| T-108 | DONE | resize_test.go TestResizeModel_AutoFlipPreservesBothRatios — position=auto, 120→90 flips to below, ratios preserved across two resizes |
| T-100-fix | DONE | entrylist/list.go WindowSizeMsg deducts 2 cells/2 rows for full pane border (matches detailpane borderRows=2 fix); list_test.go TestWindowSizeMsg_ProcessedWhenEmpty asserts 198x48 |
| T-109 | DONE | HUMAN sign-off via tui-mcp (140x35 + 80x35 resize) across tokyo-night/catppuccin-mocha/material-dark: DividerColor reads quiet (closer to Dim than FocusBorder), UnfocusedBg subtle bg tint, divider does not recolor on focus change |
| T-110 | DONE | HUMAN sign-off via tui-mcp across all 3 themes + both orientations: focused pane=FocusBorder+base bg+full fg; unfocused=DividerColor+UnfocusedBg+Faint fg; alone=focused treatment; cursor row keeps CursorHighlight when list unfocused; detail top border visible right + below |
