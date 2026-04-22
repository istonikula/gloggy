# SPEC — gloggy

distilled 2026-04-21 from code + `context/kits/cavekit-*.md`. kit files remain authoritative for per-AC detail; this spec is compressed single-source for drift checks. identifiers preserved verbatim (R#, T-###, F-###).

## §G goal

TUI for interactive JSONL log analysis. single binary. file or stdin. go + bubbletea.

## §C constraints

- go ≥1.26. macOS + linux only (fsnotify). windows out of scope.
- TUI only — bubbletea/lipgloss/termenv stack. UTF-8 assumed.
- single file input OR stdin. no multi-file, no remote, no compressed, no format-switch.
- TrueColor honoured via COLORTERM=truecolor|24bit (else termenv downsample).
- terminal floor 60x15 → "too small" message (as:R2 AC7).
- deps pinned: bubbletea v1.3.10, lipgloss v1.1.0, fsnotify v1.9.0, go-toml v2.3.0, atotto/clipboard v0.1.4.

## §I interfaces

### I.cli

- `gloggy <file>`            — load file
- `gloggy -f <file>`         — tail/follow (NOT stdin)
- `gloggy` (piped stdin)     — read stdin synchronously pre-TUI
- invalid args → stderr err + exit 1 (as:R1)

### I.cfg

- path: `$XDG_CONFIG_HOME/gloggy/config.toml` (via `os.UserConfigDir`)
- created w/ defaults on first run. unknown keys preserved on write-back.
- schema (`internal/config.Config`):
  - `theme` ∈ {`tokyo-night`(def), `catppuccin-mocha`, `material-dark`}
  - `compact_row.fields` def=[time,level,logger,msg]; `.sub_fields` def=[]
  - `hidden_fields` def=[]
  - `logger_depth` def=2
  - `detail_pane.height_ratio` def=0.30; `.width_ratio` def=0.30
  - `detail_pane.position` def=`auto` ∈ {`below`,`right`,`auto`}
  - `detail_pane.orientation_threshold_cols` def=100
  - `detail_pane.wrap_mode` def=`soft` (only mode shipped; `scroll`/`modal` future)
  - `scrolloff` def=5 — **top-level shared key**, NOT nested (cfg:R5)

### I.entry (`internal/logsource.Entry`)

`{LineNumber int, IsJSON bool, Time time.Time, Level, Msg, Logger, Thread string, Extra map[string]json.RawMessage, Raw []byte}`. RFC3339Nano. unparseable time → zero (ls:R5).

### I.msgs (bubbletea stream)

- `EntryBatchMsg{Entries []}` + `LoadProgressMsg{Count}` + `LoadDoneMsg` — load stream (drain via `LoadFileStreamMsg.Next()`)
- `TailMsg{Entries []}` + `TailStopMsg{Err}` — tail stream, **batched per fsnotify Write event** (ls:R8)

### I.keys (canonical — full list in README + as:R5 help overlay)

- nav: j/k g/G Ctrl-d/u PgDn/PgUp Space/b (dp) Home/End (dp)
- el: e/E w/W m/M u/U l/h Tab Enter / n/N  (M = clear all marks, silent no-op on zero)
- dp: / n/N + - = | Esc
- global: f y ? q T

### I.themes (`internal/theme`)

tokens: LevelError/Warn/Info/Debug, Key/String/Number/Boolean/Null, Mark, Dim, SearchHighlight, CursorHighlight, HeaderBg, FocusBorder, DividerColor, UnfocusedBg, DragHandle, BaseBg. each ctor cites upstream source (`TokyoNightSource` etc).

### I.themesel (theme selector overlay — see V29)

- trigger: global `T` (V14-gated)
- form: full-screen modal overlay (per as:R5 help pattern)
- nav: ↑/↓ or k/j cycle highlight; Enter commits; Esc reverts
- shows: bundled themes only (`tokyo-night`, `catppuccin-mocha`, `material-dark`); user-defined themes out of scope
- live preview: each highlight change → whole-TUI repaint in highlighted theme
- commit (Enter): writes `theme=<name>` to config.toml via existing write-back (V22-preserves unknown keys) + closes
- revert (Esc): restores pre-open theme + closes; no config write

## §V invariants

- **V1** (ls:R8) tail batched per FS event: 1 `TailMsg` per drain, `len(Entries)` = lines made available. initial pre-watcher drain = 1 batch. per-line emission = kit violation (causes row-by-row scroll animation on `gloggy -f bigfile`).
- **V2** (el:R14) `AppendEntries(K≥1)` ⇒ EXACTLY one cursor-snap + one viewport-adjust + one selection-signal, regardless of K. symmetric across `TailMsg` + `EntryBatchMsg`.
- **V3** (el:R14) tail-follow pre-cond: `Cursor == Total-1` before append → auto-advance + scrolloff bottom-edge. else cursor/viewport UNCHANGED. `IsAtTail()` drives `[FOLLOW]` badge live per render.
- **V4** (el:R1) each compact row = exactly 1 terminal line. embedded `\n` in msg/raw flattened to space pre-render.
- **V5** (el:R6) list `View()` outputs EXACTLY `ViewportHeight` rows (padded w/ blanks if short). no buffer, no overflow — bubbletea `JoinVertical` does not tolerate extra rows.
- **V6** (cfg:R5) `scrolloff` is **top-level** TOML int. single source of truth for el:R12 + dp:R11. NOT splittable per-pane. negative/non-int clamped to 0 + warn. at use: clamped `[0, floor(VisibleRows/2)]`.
- **V7** (el:R3, dp:R2, as:R3, as:R10) NO hardcoded colors anywhere. every fg/bg/attribute resolves through active theme token. render output ANSI must contain theme's token value.
- **V8** (el:R10, dp:R10) layout math has **single owner** (as:R2). pane mouse-handlers must NOT re-derive terminal-Y→row or subtract borders a 2nd time. violations: 2-row-offset click bug (F-127); detail underfill/overflow (F-103/F-104).
- **V9** (el:R10 AC6) list click-row resolver MUST reject clicks when `contentTopY` unset — either panic-on-unset or "wired" flag returning no-row. zero-default → silent 2-row-offset regression (F-127).
- **V10** (dp:R11) detail-pane cursor-row bg is **contiguous** across col 0..content-width. must survive inner `\x1b[0m` SGR resets from per-token syntax highlighting — strip inner resets on cursor row OR cell-level bg paint after reflow. byte-concat of `Style.Render()` is kit violation.
- **V11** (dp:R9) soft-wrap preserves SGR state across wrap boundary. continuation line reopens same style. raw `ansi.HardwrapWc(preserveSGR=false)` forbidden.
- **V12** (dp:R1, as:R11) opening detail pane does NOT transfer focus — stays on list for live-preview. focus transfer only via Tab, mouse-click-in-pane, or `/` (as:R13).
- **V13** (as:R13) `/` routes by **current focus**, not pane-open state. list-focused → list search (el:R13); dp-focused → dp search (dp:R7); filter-panel-focused → literal char. never silent no-op.
- **V14** (as:R14) global single-key reservations (`q`, `Tab`, `?`, `Esc`) MUST NOT preempt active in-pane search input. `q`/`?` → query char, NOT `tea.Quit`/help-open. exceptions: `Tab` still dismisses search via focus-cycle; `Esc` routed through pane search first. **at-impl**: global/overlay interceptors (`HelpOverlayModel.Update` + any future pre-`handleKey` consumer) MUST gate on "no pane search in input mode" before consuming a key. top-of-`Update` unconditional consumption = kit violation (B5). tests asserting a reserved key (`q`/`?`) fires its global action while pane-search is in input mode = kit violation — flip the precondition (no active search) or assert the key becomes a query char.
- **V15** (as:R9) `y` never silent: marked-entries → notice w/ count; 0 marked → "no marked entries"; clipboard err → visible err notice. `//nolint:errcheck` on `CopyMarkedEntries` = kit violation.
- **V16** (as:R7, dp:R6) `height_ratio` + `width_ratio` preserved **independently** across orientation flips. below-mode uses `height_ratio` only for vertical; right-mode vertical height = full main slot, NOT `height_ratio × term_h` (content-loss bug).
- **V17** (as:R12) ratio clamp `[0.10, 0.80]`. per-key semantics (`NextDetailRatio`, `ratiokeys.go`): `+` grows focused pane by `RatioStep` (0.05); `-` shrinks focused pane by `RatioStep`; `=` resets to `RatioDefault` (0.30) regardless of focus; `|` cycles focused pane's share through presets {0.30, 0.50} — **off-preset ratio jumps to first preset (0.30)**, not wrap (B4 gap). at boundary = no-op: value unchanged AND config write-back skipped. `handleRatioKey` MUST guard `saveConfig()` on `newR != current` — unconditional save = kit violation (B3). detail closed → all 4 keys (`|+-=`) silent no-op.
- **V18** (as:R15) drag is focus-neutral + single-persist (one config write on Release, not per-Motion). bare Press+Release w/ no Motion → no config write. press-on-current-divider → current ratio unchanged (inverse math = exact inverse of forward `PaneHeight/Width`).
- **V19** (as:R15 drag-seam) drag-seam scope: right-mode = the 1-cell `│` divider glyph. below-mode = detail pane's TOP border row ONLY — list's bottom border is an adjacent row rendered in list's focus-state color, NOT shared (F-201).
- **V20** (cfg:R4) `DragHandle ≠ DividerColor` AND `DragHandle ≠ FocusBorder` in every bundled theme. `BaseBg ≠ UnfocusedBg` within each theme; `BaseBg` pairwise distinct across themes.
- **V21** (as:R7) on `position=auto`, every `WindowSizeMsg` crossing `orientation_threshold_cols` must refresh ALL pane-local orientation-dependent flags — both in resize handler AND in `relayout()`. miss → seam renders in pre-flip colors (F-200).
- **V22** (cfg:R3, cfg:R6) unknown TOML keys preserved on write-back. live writes preserve existing unrelated values.
- **V23** (ls:R8) tail mode available only for file inputs. stdin never tailed regardless of flags.
- **V24** (ls:R7) loading never blocks UI. progress signals stream; UI renders partial results mid-load.
- **V25** (as:R9) V15 coverage MUST include live-buffer verification. unit tests asserting `m.View()` contains the notice string are necessary but NOT sufficient: (a) renderer-level issues (diff-renderer skips, line-replace width drift) could drop the row before it reaches the pty; (b) perceptual issues (notice style indistinguishable from keyhints) render bytes into the pty that humans cannot parse as feedback. every `y`-feedback path (copied-N, no-marks, clipboard-err) requires a tea.Program capture-renderer test OR a pty-driven integration test. a capture-renderer test catches (a); catching (b) also requires a tui-mcp / pty screenshot check that asserts the notice is visually distinct from the keyhints row (style, contrast, or bold). (gap revealed by B1: passing V15-aligned unit tests coexisted w/ a live TUI where the notice bytes reached the pty but blended into the Dim keyhints row.)
- **V26** (el:R1/R9) list `View()` prefix slot MUST be accounted for in compact-row width math. list prepends a 2-cell prefix (`"* "` mark / `"⌀ "` pin / `"↻ "` wrap-indicator) based on row state; the width passed to `RenderCompactRow` MUST be `m.width - 2` when a prefix is present, OR a dedicated 2-cell prefix column reserved at leftmost position of every row (padded empty when absent). naive `prefix + RenderCompactRow(m.width)` concatenation yields a `width+2`-cell row that soft-wraps to 2 terminal lines — violates V4 + cascades into V5 by emitting an extra row that JoinVertical steals from the adjacent header slot.
- **V27** (tests) `*_test.go` files MUST use `github.com/stretchr/testify` — `require` for fatal assertions, `assert` for non-fatal. stdlib `t.Errorf`/`t.Fatalf`/`t.Error`/`t.Fatal` forbidden in `*_test.go`. rationale: consistent failure messaging, diff-style equality output, less `if err != nil { t.Fatalf(...) }` boilerplate. at-use: `require.NoError(t, err)`, `require.Equal(t, want, got)`, `assert.Len(t, xs, 3)`. kit violation: bare `if got != want { t.Errorf(...) }` or `if err != nil { t.Fatal(err) }`. exception: `t.Helper`, `t.Skip*`, `t.Log*`, `t.Run`, `t.Cleanup`, `t.TempDir`, etc remain stdlib (not failure reporters).
- **V28** (rendering) full-screen replacement views (help overlay + any future overlay returning its own `View()`) MUST NOT use `\t` for alignment. bubbletea's line-diff renderer does not clear cells that `\t` skips over — they retain bytes from the previous frame, bleeding prior content into the overlay. lipgloss does not expand tabs either. at-impl: pad with spaces (column width = max key length across all domains + constant gap). tests MUST assert `View()` contains zero `\t` for any full-screen-replacement view. kit violation: `sb.WriteString("\t")` (or any raw `\t` emission) in a view builder.
- **V29** (I.themesel) theme selector overlay: trigger `T` MUST gate on V14 — active pane-search input mode → `T` becomes query char, NOT overlay-open. open → highlight = current theme. ↑/↓/k/j cycle; Enter commits + persists `theme=` via config write-back (V22 preserves unknown keys); Esc reverts to pre-open theme + no config write. live-preview: each highlight change repaints whole TUI in highlighted theme via single `theme.GetTheme` resolution path (no parallel palette code). closed overlay → zero mutation (no theme swap, no config mtime advance). overlay full-screen-replacement view → V28 applies (no `\t` padding). tests MUST cover: (a) `T` during pane-search input mode → query char (V14); (b) navigate → `View()` ANSI bytes carry highlighted theme tokens; (c) Enter → `config.toml` `theme` key updated + unrelated keys preserved; (d) Esc → pre-open theme restored + no config mtime advance.

## §T tasks

| id | status | desc | cites |
|----|--------|------|-------|
| T1 | x | human tui-mcp sign-off across all 3 themes × {80x24, 140x35} × {below, right} — README notes "not yet human tested" | V7,V10,V19,V20 |
| T2 | x | human verify tail-follow no-scroll-animation on `logs/big.log` w/ `gloggy -f` | V1,V2 |
| T3 | x | human verify mouse click-row resolver across orientations × focus states on `logs/small.log` | V8,V9 |
| T4 | x | human verify clipboard `y` notices (copied/empty/err) | V15 |
| T5 | x | human verify drag visuals (below+right) via tui-mcp screenshot + V18 bare-Press+Release no-config-write via `send_mouse`; motion-driven path covered by unit `TestModel_T156_*`/`_T164_*` — tui-mcp `send_mouse` exposes no Motion action, raw SGR motion via `send_text` not decoded | V17,V18,V19 |
| T6 | x | guard `saveConfig()` in `handleRatioKey` on `newR != current`; add regression test for no-mtime-advance at ratio boundary across `-`/`+`/`=`/`\|` | V17 |
| T7 | x | fix B1: diagnose + repair `y`-notice drop (keyhints line-replace vs bubbletea diff-renderer); add tea.Program capture-renderer OR pty-driven test for copied-N / no-marks / clipboard-err paths per V25 | V15,V25 |
| T8 | x | automate V25 class-(b) coverage — tui-mcp / pty-driven golden-frame or contrast-check test that launches gloggy, presses `y` on no-marks, reads the bottom row, and asserts the notice cells are visually distinct from the keyhints row (e.g. differing SGR style: Bold or different fg). repeat for all 3 themes × all 3 y-feedback paths (copied-N / no-marks / clipboard-err). | V15,V25 |
| T9 | x | impl `M` clear-all-marks in entrylist: add `MarkSet.Clear()`, wire `case "M"` in `list.go` (drops pin + resets mark-nav state, per u/U pattern), unit test empties MarkSet + verifies zero-count after Clear. 0-marks → silent no-op, no confirm. update help overlay (as:R5) + README keymap. | I.keys |
| T10 | x | fix B2/V26: list `View()` overflows `m.width` by 2 cells when a prefix glyph is present (`* ` mark / `⌀ ` pin / `↻ ` wrap). `list.go:740/743/745` concat `prefix + RenderCompactRow*(m.width)` → row = `m.width + 2` cells, soft-wraps to 2 terminal lines, cascades into V5 by displacing the header row from Y=0. impl option A: reserve a dedicated 2-cell prefix column inside `RenderCompactRow` + `RenderCompactRowWithBg` padding math (prefix glyph at cells [0:2], content at [2:width], padded empty when no prefix — single contract, no call-site branching). option B: pass `m.width - 2` to the renderer at the three call sites when `prefix != ""`. prefer A. add regression test: mark an entry whose msg fills the width, assert `m.View()` emits exactly `ViewportHeight` `\n`-separated rows + header visible at Y=0 (V4+V5+V26 guard). | V4,V5,V26 |
| T11 | x | migrate all `*_test.go` to testify: add `github.com/stretchr/testify` to `go.mod`; replace ~1253 `t.Errorf`/`t.Fatalf`/`t.Error`/`t.Fatal` across 58 files with `require.*`/`assert.*`. package-by-package commits for reviewability; `go test ./...` green after each package. | V27 |
| T12 | x | shrink + split `internal/ui/app/model_test.go` (~2444 LOC post-T11, was 2787 pre-testify) into topic-focused files (focus transitions, ratio keys, divider drag, search, resize/lifecycle, misc). **two-phase**: (1) simplification sweep — dedupe setup/fixtures via shared helpers, table-drive related one-offs w/ `t.Run`, drop redundant coverage (V-cited paths ! intact); (2) pure-move relocation of remaining tests into topic files (no func renames, no logic edits). one commit per phase-group so blame + coverage regressions stay reviewable. `go test ./internal/ui/app/...` green after each commit. do AFTER T11 lands. | V27 |
| T13 | x | audit `internal/ui/detailpane/model_test.go` (~664 LOC) — siblings (scroll/search/wrap/visibility/render) already topic-owned, so `model_test.go` holds Init/Update/View + focus/size integration. method: simplification sweep first (dedupe + table-drive per T12 phase 1), then split only if >~400 LOC remain. prefer fewer files + denser tables over aggressive splitting. `go test ./internal/ui/detailpane/...` green each commit. | V27 |
| T14 | x | audit `internal/ui/entrylist/list_test.go` (~624 LOC) — siblings (cursor/scroll/row/marks/search/leveljump) already topic-owned; remainder ! View + Update integration. same method as T13: simplification sweep first, split only if still oversized. `go test ./internal/ui/entrylist/...` green each commit. | V27 |
| T15 | x | fix B5: gate `HelpOverlayModel.Update` on pane-search-input-mode so `?` is consumed as a query char when list or detail search is active in input mode. impl option A: invert ordering — move `m.help.Update(msg)` *after* `handleKey` routes the key to the focused pane's search (cleanest — no predicate plumbing). option B: thread a "search-active" predicate through `Update` and short-circuit `m.help.Update` when true. prefer A. flip `TestModel_HelpOverlay_PreservesListSearchState` to assert `?` becomes a query char during active list search (query = `"abc?"`); add a new test covering `?` opens help when no pane-search in input mode. repeat both for detail-pane search. `go test ./internal/ui/app/... ./internal/ui/appshell/...` green. | V14 |
| T16 | x | fix B6: replace `\t` in `HelpOverlayModel.View()` with fixed-column space padding. impl: compute max key-string width across all domains (use `lipgloss.Width` for correct East-Asian / arrow-glyph widths) + a 2-space gap, then pad each key row to that column before writing the description. add unit test asserting `View()` contains no `\t` when open. optional tui-mcp regression: launch gloggy on `logs/tiny.log`, press `?`, `snapshot`, assert no background digits/words leak between key and desc. `go test ./internal/ui/appshell/...` green. | V28 |
| T17 | . | impl theme selector overlay per I.themesel: new `internal/ui/appshell/themeselector.go` (model + Update + View per `HelpOverlayModel` pattern); wire global `T` AFTER pane-search routing in `Model.Update` (V14-safe — same ordering lesson as T15/B5); threading: overlay holds pre-open theme + active highlight; highlight change → swap-theme msg to app model for whole-TUI repaint; Enter → existing config write-back + close; Esc → restore pre-open theme + close. update help overlay (as:R5) + README keymap. tests per V29 (a)-(d). | V14,V22,V28,V29 |
| T18 | . | [HUMAN via tui-mcp] verify theme selector live-preview + persist + revert across all 3 themes × {below, right} orientations on `logs/small.log`: launch gloggy, press `T`, navigate ↑/↓ asserting visual repaint, Enter to commit, kill+relaunch to verify persistence, then press `T` + Esc to verify revert + no config mtime advance. | V29 |

## §B bugs

| id | date | cause | fix |
|----|------|-------|-----|
| B1 | 2026-04-21 | `y` notice bytes DO reach the pty (confirmed via tui-mcp `read_region` immediately after `send_keys y`). Root cause is visual contrast: `KeyHintBarModel.View()` rendered the notice via `Foreground(m.th.Dim)` — same style as the keyhints row — so text was present in the buffer but visually blended with the dim keyhints during the 2s auto-clear window. Earlier diagnosis ("diff-renderer or line-replace drops the row") was wrong; the bug is perceptual, not a dropped write. V15 violation observable only in live TUI. | fixed in T7 (commit `89cb940`): notice branch in `internal/ui/appshell/keyhints.go:86-89` now uses `lipgloss.NewStyle().Bold(true).Foreground(m.th.FocusBorder)` — distinct from Dim keyhints on all 3 themes. V25 live-buffer tests added in `internal/ui/app/live_buffer_y_notice_test.go` for copied-N / no-marks / clipboard-err paths. |
| B2 | 2026-04-21 | list View() prepends 2-cell mark/pin/wrap prefix onto a row already padded to full `m.width`; sum overflows pane inner width, soft-wraps to 2 terminal rows, pushes list output past ViewportHeight, JoinVertical displaces the filename/counter header slot. V4+V5 violated. Observable: mark any entry whose content fills the row → header row disappears from Y=0. Code site: `internal/ui/entrylist/list.go:714-739`. | subtract prefix len from RenderCompactRow width arg, or reserve a 2-cell prefix column inside RenderCompactRow's padding math — covered by V26 |
| B3 | 2026-04-21 | `handleRatioKey` (`internal/ui/app/model.go:808`) calls `saveConfig()` unconditionally after `NextDetailRatio`; at ratio boundary 0.10/0.80, repeated `-`/`+` keypresses produce no value change (clamp-pin) but advance config mtime every press, inflating disk I/O + log churn. V17 "no-op" was value-only and silent on write side-effects. | guard `saveConfig()` on `newR != current`; covered by V17 strengthening + T6 |
| B4 | 2026-04-21 | SPEC V17 + §I.keys under-documented `\|` as just "ratio" alongside `+/-/=`; actual `cycleDetailPreset` semantics (preset-cycle with off-preset→first-preset fallback) surprised operator at `height_ratio=0.1` → `\|` snapped to 0.30 without orientation flip. Code is intentional (ratiokeys.go:78-85, unit test `TestNextDetailRatio_PipeOffPreset_JumpsToFirst`). No code change. | V17 expanded to enumerate per-key semantics + off-preset→first-preset fallback |
| B6 | 2026-04-21 | `HelpOverlayModel.View()` (`internal/ui/appshell/helpoverlay.go:78-80`) writes `\t` between key and description. bubbletea's line-diff renderer leaves `\t`-skipped cells unchanged from the previous frame, so opening `?` from the list view leaves background bytes at those columns. altscreen (`cmd/gloggy/main.go:111`) does not help — diff is per-frame inside the alt buffer. Observable via tui-mcp snapshot after `?`: `  Ctrl+d:56 INFOHalf page down`, `  Tab:17Cycle focus...`, etc., where `:17`, `:56 INFO`, `13:17` are bytes from the list row that was at that column before the overlay opened. | replace `\t` with space-padding to a fixed column; covered by V28 + T16 |
| B5 | 2026-04-21 | `?` preempts active in-pane search — violates V14. `HelpOverlayModel.Update` (`internal/ui/appshell/helpoverlay.go:52-55`) opens help on `?` unconditionally; `Model.Update` (`internal/ui/app/model.go:158-161`) calls `m.help.Update(msg)` at the top, before `handleKey` can route `?` to the active pane search. test `TestModel_HelpOverlay_PreservesListSearchState` (`internal/ui/app/model_focus_test.go:375-397`) codifies the violation: list search in input mode w/ query `"abc"` → press `?` → assertion `m.help.IsOpen()` passes + search preserved on `esc` cycle. V14 reserves `q/Tab/?/Esc` + exempts only `Tab`/`Esc`; `?` has no exemption, so it MUST append to the query like `q` does. | move help-overlay interception after pane-search routing OR gate overlay `Update` on search-input-mode; flip the test to assert `?` → query `"abc?"` + split the help-open path into a no-active-search scenario — covered by V14 strengthening + T15 |

(historical backprops recorded in kit changelogs: F-013/F-015/F-016/F-017, F-101..F-109, F-121..F-129, F-132/F-133/F-134, F-200/F-201/F-202.)
