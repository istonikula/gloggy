# Review findings ledger — 2026-05-30

Input for `/ck:spec`. Full-repo audit (5 reviewers) + filter-redesign delta (3 reviewers), commits through `6337856`.

**For ck:spec:** these are raw triaged findings. YOU decide how to apply them — which earn a new §V invariant vs. a one-off §T fix vs. a §B backprop row, how to cluster, V/B/T numbering, cites, caveman encoding per FORMAT.md. Skip every `cleared-FP` row (verified false positive — do NOT spec it). For `reviewer-claimed` rows, re-read the cited code before encoding (line numbers may have drifted; mechanism may be mis-stated). For `confirmed` rows, I read the code and the bug is real as described.

Severity = my triage (CRIT|HIGH|MED|LOW). verify = confirmed | reviewer-claimed | cleared-FP. Where my read corrected the reviewer's mechanism, the corrected root-cause is in the row.

## Findings

| file:line | sev | verify | root-cause (corrected where noted) | fix-direction |
|-----------|-----|--------|------------------------------------|---------------|
| config/config.go:273 | CRIT | confirmed | `os.WriteFile(path,data,0o644)` truncates in-place; crash mid-write corrupts/empties config. | tempfile + fsync + `os.Rename`. |
| cmd/gloggy/main.go:59 | CRIT | confirmed | `stdinStat,_ := os.Stdin.Stat()` then `stdinStat.Mode()`; on Stat err stdinStat is nil → **nil-deref panic** (reviewer said "silently 0" — wrong, it panics). | check err; treat err as non-pipe. |
| cmd/gloggy/main.go:64-66 | HIGH | confirmed | `case fs.NArg()==1` matches even when `fromStdin` true (first case needs NArg==0) → piped-stdin + file arg silently ignores stdin, reads file. | reject conflict w/ stderr + nonzero exit. |
| cmd/gloggy/main_test.go:40 | HIGH | reviewer-claimed | pipe write-end `pw` never closed; fd leak across `-count=N` runs. | close `pw` after capturing `pr`. |
| filter/filter.go (Add, ~line drifted) | CRIT→HIGH | reviewer-claimed | `savedEnabled[id]=...` written when `globallyDisabled` w/o nil-guard; map only init'd in `ToggleAll`. Safe in current call-graph (only ToggleAll flips flag); latent for future refactor. Downgrade CRIT→HIGH — not currently reachable. | init `savedEnabled` in `NewFilterSet`. |
| ui/entrylist/row.go:68 | CRIT→HIGH | confirmed (mechanism corrected) | `visiblePrefixLen` uses `len(loggerStr)` = **byte-len of raw (unstyled) AbbreviateLogger output**; reviewer said "ANSI-styled" — wrong, it's raw. Real bug is multibyte rune (CJK/emoji) byte-len > cell-width → msg over-truncated. | width via `lipgloss.Width`/rune-count. |
| ui/entrylist/row.go:122 | HIGH | reviewer-claimed | same byte-vs-cell prefix width in `RenderCompactRowWithBg`; verify whether styled/unstyled. | same. |
| ui/entrylist/row.go:42-79 | MED | reviewer-claimed | raw-text trunc `raw[:width]` cuts mid-rune → invalid UTF-8. | rune-aware trunc. |
| ui/filter/prompt.go:207-214 | MED | reviewer-claimed | backspace counts runes; insert path (~224,226) appends bytes → buffer rune/byte drift on multibyte input. | unify on `[]rune`. |
| logsource/tail.go:76 | HIGH | reviewer-claimed | fresh `bufio.Reader` per drain caches bytes → stale file offset, re-read/skip on rotation. | seek to offset, or persistent reader. |
| logsource/tail.go:92 | HIGH | reviewer-claimed | `pending` partial-line buffer survives file truncation/rotation → corrupt/lost line. | detect size-shrink, reset pending, reopen. |
| logsource/tail.go:113 | HIGH | reviewer-claimed | drain IO err closes tail, no retry/reopen; lines lost, no user signal. | TailErrMsg{Retryable} + backoff reopen. |
| logsource/loader.go:30 | HIGH | reviewer-claimed | `os.Open` err discarded → emits LoadDoneMsg(0); missing file looks "loaded empty". | emit LoadErrMsg. |
| logsource/loader.go:86 | LOW | reviewer-claimed | scanner err logged, not returned; partial load shows "done". | return/surface err. |
| logsource/tail_stdin.go:106 | MED | reviewer-claimed | `readerDone` err only read in `==context.Canceled` branch; if `ctx.Done()` wins race, err leaked. | drain readerDone unconditionally. |
| logsource/classify.go:15 | LOW | reviewer-claimed | no empty-line bounds-check before `line[0]`; not a known panic, invariant unverified. | defensive test. |
| ui/entrylist/cursor.go:176 | HIGH | reviewer-claimed | `CursorPosition` returns 1 on empty filtered list (should be 0); no bounds guard. | empty-list guard → 0. |
| ui/entrylist/list.go:550-571 | HIGH | reviewer-claimed | `entryIndexForVisible`/`visibleIndexForEntry` asymmetric across pin-out-of-filter → round-trip identity broken. | round-trip property-test + fix. |
| ui/detailpane/scroll.go:255 | HIGH | reviewer-claimed | empty content renders `h-1` blanks (off-by-one) → adjacent panes shift up. | fill exactly h. |
| ui/detailpane/model.go:261-282 | HIGH | reviewer-claimed | `ScrollToLine` temp-mutates `s.height`, `followCursor` sees wrong viewport → cursor lands out of bounds. | thread viewport via param. |
| ui/appshell/layout.go:109 | CRIT→SUSPECT | reviewer-claimed | reviewer alleges ClickToPaneRow boundary off-by-one. BUT ClickToListRow comment documents `[start, start+viewportRows)` half-open as **intentional**. Likely FP, or applies to a different func. **ck: re-read ClickToPaneRow specifically before specing.** | verify first; likely no-op. |
| ui/entrylist/marks.go:73 | HIGH→cleared-FP | cleared-FP | `(currentIdx - i%n + n)%n`: `i%n` differs from `i` only at `i==n`, where both formulas yield `currentIdx`. For `i∈1..n-1`, `i%n==i`. **Harmless — identical output to `(currentIdx-i+n)%n`.** Confusing code, not a bug. | optional: drop `%n` for clarity (cosmetic, no §B). |
| ui/filter/panel.go:53-57 | HIGH | reviewer-claimed | cursor reads `GetIDs()`+`GetAll()` separately; concurrent mutation → mismatched filter target on toggle/delete. | snapshot once via single read. |
| ui/filter/prompt.go:157-160 | HIGH | reviewer-claimed | `Validate` checks `Pattern==""` only; whitespace-only passes → filter matches ~everything. | TrimSpace reject. |
| ui/filter/prompt.go:163-173 | MED | reviewer-claimed | Edit ignores `fs.Update(id,f) bool`; if id removed between OpenEdit+Submit, prompt closes w/ no mutation, no signal. | check return + notice. |
| ui/filter/prompt.go:361 | LOW | reviewer-claimed | `cycleSyntaxPrev` hardcodes `%3`; Syntax enum growth breaks cycle. | derive modulo from enum count. |
| ui/app/model.go:683-747 | HIGH | reviewer-claimed | pane auto-close during drag doesn't resync `m.focus`; focus left on hidden detail pane → next key mis-routed. | sync focus to visible peer on close. |
| ui/app/model.go:840-852 | MED | reviewer-claimed | `View()` writes `m.list.Focused`/`Alone` before render; bubbletea calls View >1×/frame → stale-state risk. Purity violation. | move writes to Update. |
| ui/app/model.go:938-943 | MED | reviewer-claimed | `appendToList` re-syncs pane on new entry, clobbers active pane-search query/scroll. | preserve paneSearch across re-sync. |
| ui/app/model.go:517-524 | MED | reviewer-claimed | ratio keys silent no-op when pane closed; no feedback. | notice on no-op. |
| ui/app/model.go:810 | LOW | reviewer-claimed | `saveConfig` swallows err; disk-full/perm err vanishes. | log + notice. |
| ui/app/model.go:393-400 | MED | reviewer-claimed | dead duplicate `OpenPromptMsg` branch (FocusFilterPanel case 583-598 resolves it first). Maintenance trap. | delete dead branch. |
| ui/appshell/ratiokeys.go:121-136 | MED | reviewer-claimed | `RatioFromDragX` div-by-zero when `usable==0` → +Inf, clamp masks it. | early-return guard. |
| ui/appshell/mouse.go:82-90 | MED | reviewer-claimed | `ListContentWidth()-1` → -1 when width 0 → list zone collapses. | guard width<1 → ZoneUnknown. |
| ui/appshell/help.go:53 | LOW | reviewer-claimed | "Delete highlighted filter" vs model.go:918 "Delete filter" wording drift. | align. |
| filter/filter.go:68 (Remove) | LOW | reviewer-claimed | `Remove` leaves stale `savedEnabled[id]`; cosmetic leak (ids not reused). | `delete(savedEnabled,id)`. |
| config/config.go:80 | LOW | reviewer-claimed | `Load(path)` no symlink/`..` resolution; matters only if custom `--config`/env path added. | harden or document caller-trust. |
| tests/integration/tail_test.go:68 | MED | reviewer-claimed | 3s timeout silently "skips" (passes) on fsnotify delay/unsupported env → masks failure. | hard-fail or explicit t.Skip. |

## Cleared false positives — do NOT spec

| finding | why cleared |
|---------|-------------|
| ui/app/model.go:238-279 "WindowSizeMsg missing return" | re-review confirmed cmd IS returned; reviewer line-ranges off. |
| ui/app/model.go:186-234 "themesel overlay drops Cmds" | re-review: render path correct; FP. |
| ui/app/model.go:656 "markedIDs empty-Cmd race" | current code guards len==0; only revisit if overlay-cancel path added. |
| ui/entrylist/marks.go:73 "PrevMark off-by-one" | math proven harmless (see ledger row); identical output. |

## Cross-cutting themes (hints, not directives — ck decides if any earn a §V)

1. ANSI/multibyte width: row.go (×2-3) + prompt.go buffer. Recurring class — candidate single render-width invariant.
2. Tail rotation/truncation/err-survival: tail.go ×3 + loader.go + tail_stdin.go. Candidate tail-robustness invariant + integration suite.
3. Error-swallow: config save, loader, tail, stdin — several silent-fail paths. Candidate broadened "never silent" (V15 already exists for `y`).
4. Layout/viewport off-by-ones + idx round-trip: cursor, list, detailpane, (layout suspect). V34 already governs ground-truth tests; may want round-trip identity invariant.
5. tea purity / focus-pane sync: model.go View-mutation + focus-on-close + paneSearch-clobber.
6. Filter validation thinness: prompt whitespace + Update-return. V35 already covers filter UX — likely amend, not new V.
7. Dead code / degenerate-size guards: model.go dead branch, ratiokeys/mouse div-by-zero/negative-width. Pure §T tidy, probably no §V.
