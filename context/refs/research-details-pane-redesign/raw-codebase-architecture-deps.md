## Agent: codebase-architecture-deps

**IMPORTANT CORRECTION from briefing:** gloggy is **Go + bubbletea + lipgloss**, not Rust + ratatui.

### 1. Overall Architecture
- Finding: Go TUI app using `charmbracelet/bubbletea v1.3.10`. Single binary log viewer with async loading, follow-mode tailing, filter engine.
- Evidence: `go.mod:5-11`; `cmd/gloggy/main.go:3-13`; `internal/ui/app/model.go:20-51`.
- Entry points: `cmd/gloggy/main.go` (CLI, config load, stream setup, TUI launch). File/stdin via `logsource.LoadFile()` / `logsource.ReadStdin()`. Tail via `logsource.TailFile()` using fsnotify v1.9.0.
- Core deps: bubbletea + lipgloss + termenv (muesli), fsnotify, pelletier/go-toml, atotto/clipboard.
- Confidence: HIGH

### 2. Current Details Pane Implementation
- Finding: Details pane is bottom-stacked vertical using `lipgloss.JoinVertical`. Implemented in `internal/ui/detailpane/` package.
- Main model: `detailpane.PaneModel` (model.go:17-26). Manages open/close state, delegates to `ScrollModel`.
- Layout: `header | entryList | detailPane | statusBar` (appshell/layout.go:96-102).
- Height: `HeightModel` stores ratio (default 0.30), adjustable via `+`/`-`, survives resize (height.go).
- Files:
  - `detailpane/model.go` — Open/close lifecycle, blur/focus: `Open(entry)`, `Close()`, `IsOpen()`
  - `detailpane/render.go` — JSON highlighting + raw: `RenderJSON()`, `RenderRaw()` (lines 17-60)
  - `detailpane/scroll.go` — Scrolling (j/k/wheel)
  - `detailpane/height.go` — Ratio height
  - `detailpane/search.go` — In-pane `/`/`n`/`N` search
  - `detailpane/visibility.go` — Per-field visibility + config writeback
  - `detailpane/fieldclick.go` — Mouse click field extraction
- Show/hide: `PaneModel.open` bool; `app/model.go:180-183` handles `SelectionMsg` → `pane.Open(entry)`; Enter or double-click emits `entrylist.OpenDetailPaneMsg` → `openPane()` (app/model.go:332-338). Esc/Enter emits `BlurredMsg`, closes pane (detailpane/model.go:73-82).
- Content: JSON entries `RenderJSON()` unmarshal + reorder (known-first, then alpha) + theme colors. Non-JSON `RenderRaw()` plain.
- Border: T-082 top border via `lipgloss.NormalBorder().BorderTop(true)` (model.go:118-126). T-083 left border when focused: `BorderLeft(m.Focused)` + `FocusBorder` color (model.go:122-124).
- Confidence: HIGH

### 3. Main List + Wide Row Handling
- Finding: Virtual rendering (only visible rows + buffer), one-line-per-entry, **truncation** not wrap.
- Files: `entrylist/list.go` (ListModel), `entrylist/row.go` (compact row), `entrylist/scroll.go`.
- Row format: `HH:MM:SS LEVEL LOGGER MSG` — 8+1+5+1+loggerLen+1, msg truncated to `width - visiblePrefixLen` (row.go:36-90).
- Newline flattening: `flattenNewlines()` collapses `\n\r\t` to single space (row.go:159-183).
- Virtual: renders only `ViewportHeight` rows, padded empty to exact height (list.go:350-393).
- Log types: JSON (parsed, `IsJSON=true`, fields extracted) and non-JSON (raw bytes, dim color).
- Confidence: HIGH

### 4. Layout Computation + Responsive Logic
- Finding: **No horizontal split exists.** Hardcoded vertical stacking. Below-vs-beside decision does not exist.
- Layout struct: `Width, Height, HeaderHeight, StatusBarHeight, DetailPaneOpen, DetailPaneHeight` (appshell/layout.go:16-103).
- `EntryListHeight()` = `Height - (Header + StatusBar + DetailPaneHeight if open)` (layout.go:28-38).
- Resize via `tea.WindowSizeMsg` → `ResizeModel.Update()` (resize.go:24-30).
- Height scaling: ratio preserved via `HeightModel.SetTerminalHeight()` (height.go:53-56).
- Relayout flow (app/model.go:340-348): `ApplyToLayout(resize, ratio, isOpen)` → `SetDetailPane(isOpen, paneHeight)` → list gets `tea.WindowSizeMsg{Width, Height:EntryListHeight}`.
- Confidence: HIGH

### 5. Dependencies (go.mod)
- Minimal charmbracelet-only TUI stack. No JSON tree widget, no textinput widget, no structured logging lib.
- Implications: existing deps sufficient for side-panel; `lipgloss.JoinHorizontal()` available. No new dep needed.
- Confidence: HIGH

### 6. Focus + Keyboard Routing
- Focus enum: `FocusTarget` = `FocusEntryList` | `FocusDetailPane` | `FocusFilterPanel` (layout.go:7-14).
- Routing in `app/model.go:211-280` `handleKey()`:
  - Global quit `q` only in entry list (line 213)
  - Help overlay intercepts all (110-115)
  - Switch on focus:
    - DetailPane (221-238): `+`/`-` resize, `/` search, else forward to `pane.Update()`
    - FilterPanel (240-248): forward; Esc returns focus
    - EntryList (default, 250-279): `f` filter, Enter open pane, `y` copy marks
- Transitions: `openPane()` sets `m.focus = FocusDetailPane` (334). `BlurredMsg` → `FocusEntryList` (192). `f` → `FocusFilterPanel` (253).
- Mouse: `MouseRouter` maps (x,y) to `ZoneEntryList | ZoneDetailPane | ZoneDivider | ZoneHeader | ZoneStatusBar` (mouse.go:19-85).
- Implication: side-panel needs mouse zone recalc (vertical slices, not horizontal rows).
- Confidence: HIGH

### 7. Theme/Style System
- Centralized `theme.Theme` struct (theme/theme.go:12-37). All tokens: LevelError/Warn/Info/Debug, SyntaxKey/String/Number/Boolean/Null, Mark, Dim, SearchHighlight, CursorHighlight, HeaderBg, FocusBorder.
- Built-in themes: tokyo-night (default), catppuccin-mocha, material-dark (lines 42-127).
- Usage: components receive `theme.Theme` on creation. Colors never hardcoded.
- Focus indicator: `Focused` field on pane → conditional `BorderLeft` + `FocusBorder` color.
- Config-driven: theme name in TOML; `theme.GetTheme(cfg.Theme)` (app/model.go:56) with fallback.
- Confidence: HIGH

### 8. Testing
- 44 `_test.go` files. Unit + integration. **No golden-file / snapshot tests.**
- Coverage: config, filter, logsource, theme, UI components, integration smoke tests.
- Unit example: `appshell/layout_test.go:9-46` checks line order without full snapshot.
- Integration: `tests/integration/smoke_test.go:23-100` full lifecycle, no-panic, no visual snapshot.
- Rendering tests: call `View()`, check non-empty. No ANSI/exact-layout verification.
- Gaps: no snapshot tests for layout, no mouse zone tests for side panel, no visual regression tool.
- Confidence: HIGH

### Tight Coupling Concerns
1. Layout hardcoded vertical (`appshell/layout.go`): `Render()` exclusively uses `JoinVertical()`. Needs `Orientation` or `DetailPaneWidth` field + horizontal branch.
2. Mouse zones assume vertical stack (`appshell/mouse.go:37-75`). Post-list pane hardcoded. Needs layout-aware zone calc.
3. Relayout passes only `EntryListHeight` (`app/model.go:340-348`). Needs width tracking.
4. Render order OK — `m.pane.IsOpen()` checked in `View()`, not state. No fix needed.

### Redesign Impact Matrix
| Aspect | Current | Side panel needs |
|--------|---------|------------------|
| Layout engine | JoinVertical | Add JoinHorizontal branch |
| Height tracking | HeightModel | Keep; add WidthModel |
| Focus routing | Centralized | Keep; update mouse zones |
| Theme | Tokenized | Keep, no changes |
| Lifecycle | Open/Close | Keep |
| Rendering | JSON/raw | Keep |
| Testing | No snapshots | **Add golden-file tests for layout** |

**Files to touch for side panel:**
1. `appshell/layout.go` — `Layout` struct + `Render()` method
2. `appshell/mouse.go` — `Zone()` calc for vertical zones
3. `app/model.go` — relayout width, View() composition
4. `detailpane/model.go` — accept width param / `SetWidth()`
5. `config/config.go` — add `detail_pane.width_ratio` or `position` field
