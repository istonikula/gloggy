---
created: "2026-04-15T00:00:00Z"
last_edited: "2026-04-17T22:52:48+03:00"
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
