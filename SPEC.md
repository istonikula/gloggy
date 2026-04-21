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
- el: e/E w/W m u/U l/h Tab Enter / n/N
- dp: / n/N + - = | Esc
- global: f y ? q

### I.themes (`internal/theme`)

tokens: LevelError/Warn/Info/Debug, Key/String/Number/Boolean/Null, Mark, Dim, SearchHighlight, CursorHighlight, HeaderBg, FocusBorder, DividerColor, UnfocusedBg, DragHandle, BaseBg. each ctor cites upstream source (`TokyoNightSource` etc).

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
- **V14** (as:R14) global single-key reservations (`q`, `Tab`, `?`, `Esc`) MUST NOT preempt active in-pane search input. `q` → query char, NOT `tea.Quit`. exceptions: `Tab` still dismisses search via focus-cycle; `Esc` routed through pane search first.
- **V15** (as:R9) `y` never silent: marked-entries → notice w/ count; 0 marked → "no marked entries"; clipboard err → visible err notice. `//nolint:errcheck` on `CopyMarkedEntries` = kit violation.
- **V16** (as:R7, dp:R6) `height_ratio` + `width_ratio` preserved **independently** across orientation flips. below-mode uses `height_ratio` only for vertical; right-mode vertical height = full main slot, NOT `height_ratio × term_h` (content-loss bug).
- **V17** (as:R12) ratio clamp `[0.10, 0.80]`. at boundary = no-op (not wrap): value unchanged AND config write-back skipped. `handleRatioKey` MUST guard `saveConfig()` on `newR != current` — unconditional save = kit violation (B3): at ratio=0.1 repeated `-` advances mtime w/o value change, inflating disk I/O + log churn. detail closed → all 4 keys (`|+-=`) silent no-op.
- **V18** (as:R15) drag is focus-neutral + single-persist (one config write on Release, not per-Motion). bare Press+Release w/ no Motion → no config write. press-on-current-divider → current ratio unchanged (inverse math = exact inverse of forward `PaneHeight/Width`).
- **V19** (as:R15 drag-seam) drag-seam scope: right-mode = the 1-cell `│` divider glyph. below-mode = detail pane's TOP border row ONLY — list's bottom border is an adjacent row rendered in list's focus-state color, NOT shared (F-201).
- **V20** (cfg:R4) `DragHandle ≠ DividerColor` AND `DragHandle ≠ FocusBorder` in every bundled theme. `BaseBg ≠ UnfocusedBg` within each theme; `BaseBg` pairwise distinct across themes.
- **V21** (as:R7) on `position=auto`, every `WindowSizeMsg` crossing `orientation_threshold_cols` must refresh ALL pane-local orientation-dependent flags — both in resize handler AND in `relayout()`. miss → seam renders in pre-flip colors (F-200).
- **V22** (cfg:R3, cfg:R6) unknown TOML keys preserved on write-back. live writes preserve existing unrelated values.
- **V23** (ls:R8) tail mode available only for file inputs. stdin never tailed regardless of flags.
- **V24** (ls:R7) loading never blocks UI. progress signals stream; UI renders partial results mid-load.
- **V25** (as:R9) V15 coverage MUST include live-buffer verification. unit tests asserting `m.View()` contains the notice string are necessary but NOT sufficient — bubbletea's diff-renderer or keyhint line-replace edge cases can drop the row so the notice never reaches the pty. every `y`-feedback path (copied-N, no-marks, clipboard-err) requires either a tea.Program capture-renderer test OR a pty-driven integration test that reads the bottom row post-keypress and asserts the notice text is present. (gap revealed by B1: passing V15-aligned unit tests coexisted w/ fully-silent live TUI.)
- **V26** (el:R1/R9) list `View()` prefix slot MUST be accounted for in compact-row width math. list prepends a 2-cell prefix (`"* "` mark / `"⌀ "` pin / `"↻ "` wrap-indicator) based on row state; the width passed to `RenderCompactRow` MUST be `m.width - 2` when a prefix is present, OR a dedicated 2-cell prefix column reserved at leftmost position of every row (padded empty when absent). naive `prefix + RenderCompactRow(m.width)` concatenation yields a `width+2`-cell row that soft-wraps to 2 terminal lines — violates V4 + cascades into V5 by emitting an extra row that JoinVertical steals from the adjacent header slot.

## §T tasks

| id | status | desc | cites |
|----|--------|------|-------|
| T1 | x | human tui-mcp sign-off across all 3 themes × {80x24, 140x35} × {below, right} — README notes "not yet human tested" | V7,V10,V19,V20 |
| T2 | x | human verify tail-follow no-scroll-animation on `logs/big.log` w/ `gloggy -f` | V1,V2 |
| T3 | x | human verify mouse click-row resolver across orientations × focus states on `logs/small.log` | V8,V9 |
| T4 | ~ | human verify clipboard `y` notices (copied/empty/err) | V15 |
| T5 | ~ | human verify ratio drag — below + right, tui-mcp `send_mouse` press-hold-move-release | V17,V18 |
| T6 | . | guard `saveConfig()` in `handleRatioKey` on `newR != current`; add regression test for no-mtime-advance at ratio boundary across `-`/`+`/`=`/`\|` | V17 |

## §B bugs

| id | date | cause | fix |
|----|------|-------|-----|
| B1 | 2026-04-21 | `y` notice set on `m.keyhints` and present in `m.View()` output but never reaches rendered terminal buffer — bubbletea diff-renderer or KeyHintBar line-replace path silently drops the row. V15 violation observable only in live TUI (all unit tests pass). | pending: investigate keyhints line-replace vs bubbletea diff; add pty-integration test per V25 |
| B2 | 2026-04-21 | list View() prepends 2-cell mark/pin/wrap prefix onto a row already padded to full `m.width`; sum overflows pane inner width, soft-wraps to 2 terminal rows, pushes list output past ViewportHeight, JoinVertical displaces the filename/counter header slot. V4+V5 violated. Observable: mark any entry whose content fills the row → header row disappears from Y=0. Code site: `internal/ui/entrylist/list.go:714-739`. | subtract prefix len from RenderCompactRow width arg, or reserve a 2-cell prefix column inside RenderCompactRow's padding math — covered by V26 |
| B3 | 2026-04-21 | `handleRatioKey` (`internal/ui/app/model.go:808`) calls `saveConfig()` unconditionally after `NextDetailRatio`; at ratio boundary 0.10/0.80, repeated `-`/`+` keypresses produce no value change (clamp-pin) but advance config mtime every press, inflating disk I/O + log churn. V17 "no-op" was value-only and silent on write side-effects. | guard `saveConfig()` on `newR != current`; covered by V17 strengthening + T6 |

(historical backprops recorded in kit changelogs: F-013/F-015/F-016/F-017, F-101..F-109, F-121..F-129, F-132/F-133/F-134, F-200/F-201/F-202.)
