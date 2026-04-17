## Agent: codebase-patterns-tests

### 1. Coding Conventions + Error Handling
- Finding: Go stdlib `error` + `fmt.Errorf("%w", ...)` wrapping. No anyhow/thiserror-style libs.
- Evidence: `cmd/gloggy/main.go:16-19`. Errors bubble as `(T, error)`. Config returns `LoadResult` with separate `Warnings []string` (config.go:48-50).
- Implication: Keep Go-stdlib error style; no new deps.
- Confidence: HIGH

- Finding: Value-receiver immutable/chainable pattern for models.
- Evidence: `detailpane/model.go:17-127` — all methods `func (m PaneModel) ...` return new instance.
- Implication: New panes/models follow same immutable-chainable style, `(NewModel, tea.Cmd)` returns.
- Confidence: HIGH

- Finding: Package-by-domain organization under `internal/`.
- Evidence: `internal/logsource/`, `config/`, `filter/`, `ui/app/`, `ui/entrylist/`, `ui/detailpane/`, `ui/appshell/`, `ui/filter/`, `theme/`.
- Implication: Right-panel redesign belongs in its own package (or extends `detailpane`).
- Confidence: HIGH

### 2. State Management — Elm-architecture Bubble Tea
- Finding: Root `Update()` is giant type-switch dispatching to subsystems.
- Evidence: `app/model.go:108-209`. Root `Model` (lines 20-51) holds: `list, pane, paneHeight, paneSearch, visibility, filterSet, filterPanel, header, help, focus`, etc. Focus dispatch at 220-279: `switch m.focus { case FocusDetailPane, FocusFilterPanel, ... }`.
- Message types per subsystem: `entrylist.SelectionMsg`, `detailpane.BlurredMsg`, `uifilter.FilterConfirmedMsg`, `logsource.LoadFileStreamMsg`.
- Implication: Right-side pane = new focus target. Extend `FocusTarget` enum (appshell/layout.go:7-14). Create messages. Wire in Update().
- Confidence: HIGH

- Finding: Command bus `tea.Cmd` for async work (file load/tail).
- Evidence: `logsource/loader.go:13-98` returns cmds that produce `LoadFileStreamMsg`.
- Implication: Complex right-pane work (if any) uses `tea.Cmd`.
- Confidence: HIGH

### 3. Pane/Widget Composition
- Finding: **No Component trait/interface.** Each pane is independent model, manually wired in parent.
- Evidence: `app/model.go:20-51` holds 12 sub-models directly. Each implements `tea.Model` (Init, Update, View).
- New pane registration: add field to app Model, create msgs, route in Update(), call View() in layout Render().
- Implication: Create new `rightpane` package (or reshape `detailpane`). Add field, messages, routing, view.
- Confidence: HIGH

- Finding: Composition via `lipgloss.JoinVertical` (stack only).
- Evidence: `appshell/layout.go:96-103`. Current: `header(1) | entryList | detailPane(if open) | statusBar(1)`.
- Redesign: split middle zone using `lipgloss.JoinHorizontal`: `header(1) | [list | rightPane] | statusBar(1)`.
- Confidence: HIGH

### 4. Wide Text / Wrap / Scroll
- Finding: **No word-wrap in detail pane.** Content split by `\n`, vertical scroll only.
- Evidence: `detailpane/scroll.go:16-22` + line 103 `strings.Join(m.lines[m.offset:end], "\n")`.
- Implication: Reuse ScrollModel if right pane works line-by-line. Long fields may need horizontal scroll or wrap (new).
- Confidence: HIGH

- Finding: Entry list truncates messages; no wrap.
- Evidence: `entrylist/row.go:42-80` — `msg[:remaining]`.
- Implication: Right pane gets more width → less truncation issue for JSON display.
- Confidence: HIGH

- Finding: **No horizontal scroll logic anywhere.**
- Implication: Gap. If right pane must show very long field values without wrap, new code needed.
- Confidence: HIGH

### 5. Nested Data Rendering — Flat JSON
- Finding: `RenderJSON()` renders nested objects/arrays as plain indented text. **No tree/accordion folding.**
- Evidence: `detailpane/render.go:17-170` — recursive `renderValue()` outputs nested braces/brackets as text.
- Ordering: known-first (time, level, msg, logger, thread) then alphabetic (lines 64-84).
- Implication: Reuse as-is. Folding = new feature requiring per-path expand/collapse state.
- Confidence: HIGH

- Finding: Field visibility is global, not per-entry.
- Evidence: `detailpane/visibility.go:11-63` — `hiddenFields []string`, persisted via `config.Save()` (line 49).
- Implication: Reuse/extend `VisibilityModel` for right pane.
- Confidence: HIGH

### 6. Focus Indicators + Borders
- Finding: Colored left border on focused pane.
- Evidence: `detailpane/model.go:110-127` — `BorderLeft(m.Focused)` + `BorderForeground(m.th.FocusBorder)` (122-124).
- Assignment: `m.pane.Focused = (m.focus == appshell.FocusDetailPane)` (app/model.go:311).
- Test: `detailpane/model_test.go:113-125` checks "│" border chars count.
- Theme token: `theme/theme.go:36` — `FocusBorder lipgloss.Color` (tokyo-night `#7aa2f7`).
- Confidence: HIGH

- Finding: Top border on current detail pane.
- Evidence: `detailpane/model.go:119` — `BorderTop(true)`. Test at `model_test.go:99-110` verifies "─".
- Implication: In split layout, top border may be replaced by left border separator from list.
- Confidence: HIGH

- Finding: Header bar has `HeaderBg` background token.
- Evidence: `theme/theme.go:35` + `appshell/header.go` (not shown).
- Confidence: MEDIUM

### 7. Test Infrastructure
- Finding: Stdlib `testing` only. **No insta / tui-test / snapshot framework.**
- Colocated `_test.go` files. Run: `go test ./...` (+ Makefile targets).
- Confidence: HIGH

- Finding: Tests check substring presence + ANSI code extraction; no full-screen snapshots.
- Evidence: `detailpane/render_test.go:54-125` (ANSI color extraction `colorANSI()`), `model_test.go:29-125`.
- Gap: no snapshot testing for layout composition.
- Implication: Add snapshot (golden-file) tests for split layout verification.
- Confidence: HIGH

- Finding: Fixtures are inline builders; no `testdata/`.
- Evidence: `app/model_test.go:17-62` helpers: `testCfg`, `makeEntries(n)`, `newModel`, `send`, `key`, `resize`.
- Confidence: HIGH

- Finding: Integration tests in `tests/integration/`.
- Evidence: `tests/integration/smoke_test.go:23-135` — real temp JSONL, load → nav → open detail → filter → mark → resize.
- Confidence: HIGH

### 8. Existing Tests for List/Detail/Layout
- List: `entrylist/list_test.go:35-92` — virtual render count ≤ viewport; <16ms for 100k entries; no-op SelectionMsg; 1-based cursor.
- Detail: `detailpane/render_test.go:54-125` — indent, all fields present, ANSI colors per token, theme switch, non-JSON plain.
- Detail model: `model_test.go:29-125` — Open sets flag; Esc/Enter emit BlurredMsg; top border; focus → left border.
- Layout: `appshell/layout_test.go` (build-site T-046) — line order / composition.
- Smoke: `tests/integration/smoke_test.go:23-135` — full lifecycle.
- Implication: Add split-layout tests; extend smoke to cover right-pane + focus switching + split layout.
- Confidence: HIGH

### 9. Context Directory
- Kits: `context/kits/cavekit-overview.md`, `cavekit-detail-pane.md` (8 reqs, lines 14-94), `cavekit-entry-list.md`, `cavekit-app-shell.md`, etc.
- Impl tracking: `context/impl/impl-detail-pane.md`, `impl-app-shell.md`, `impl-entry-list.md` — task-ID tables with status.
- Refs: `context/refs/research-details-pane-redesign/findings-board.md` initialized.
- Build site: `context/plans/build-site.md:1-150` — 83 tasks in 9 tiers. Tier 8 visual polish (T-081 focus, T-083 border).
- Implication: Revise `cavekit-detail-pane.md` (or add `cavekit-right-pane.md`). Track in new `impl-right-pane.md`. Follow cavekit methodology: sketch → map → make.
- Confidence: HIGH

### Gaps + Risks
1. No snapshot tests — manual verification of split layout needed (or introduce golden-file tests).
2. No horizontal scroll — gap for very long field values on right pane.
3. No tree/accordion rendering — if folding desired, new code.
4. Layout complexity: vertical→horizontal split changes LayoutModel, mouse routing, resize logic, focus management. Off-by-one risk high.
5. Focus switch at split boundary: mouse-on-divider, Tab/h/l keys — new interactions needing tests.
