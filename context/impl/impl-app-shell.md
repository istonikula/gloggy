---
created: "2026-04-15T00:00:00Z"
last_edited: "2026-04-16T20:09:25+03:00"
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
