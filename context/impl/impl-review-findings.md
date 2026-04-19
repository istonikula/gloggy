---
created: "2026-04-18T09:40:17+03:00"
last_edited: "2026-04-19T19:45:00+03:00"
---

# Review Findings

Source: `/ck:check` run on 2026-04-18 after user report "pressing / does not do anything".
Scope: in-pane search flow (`internal/ui/detailpane/search.go`, `internal/ui/app/model.go` key routing, `internal/ui/appshell/help.go` keyhints).

| Finding | Severity | File:line | Status | Addressed by |
|---|---|---|---|---|
| F-001: `/` with entry-list focus is a silent no-op | P0 | `internal/ui/app/model.go:296-301, 316-357` | NEW | T-116 |
| F-002: Search highlights + prompt never rendered (`HighlightLines` dead) | P0 | `internal/ui/detailpane/model.go` (View has no `SearchModel` ref); `internal/ui/detailpane/search.go:127-141` | NEW | T-114 |
| F-003: `strings.Split(m.pane.View(), "\n")` includes border + ANSI — off-by-one + corruption | P1 | `internal/ui/app/model.go:298` | NEW | T-113, T-114 |
| F-004: No search prompt / query / match counter rendered | P1 | `internal/ui/detailpane/model.go`; `internal/ui/detailpane/search.go` | NEW | T-114 |
| F-005: `n`/`N` do not scroll viewport to current match | P1 | `internal/ui/detailpane/search.go:96-123`; `internal/ui/app/model.go:296-301` | NEW | T-115 |
| F-006: Search state survives pane close / reopen — stale query corrupts next input | P2 | `internal/ui/app/model.go:227-232`; `internal/ui/detailpane/model.go:85-89` | NEW | T-117 |
| F-007: Two-step Esc (dismiss search → close pane) untested; fragile against branch reorder | P2 | `internal/ui/app/model.go:296-304`; `internal/ui/detailpane/model.go:108-114` | NEW | T-120 |
| F-008: Active search hijacks all keys — `j`/`k`/`g`/`G` appended to query instead of scrolling | P2 | `internal/ui/app/model.go:296-301`; `internal/ui/detailpane/search.go:185-192` | NEW | T-118 |
| F-009: Backspace on multi-byte query corrupts UTF-8 (byte slice vs rune count) | P2 | `internal/ui/detailpane/search.go:181-184` | NEW | T-119 |
| F-010: No feedback when query yields zero matches | P3 | `internal/ui/detailpane/search.go:78-93, 96-123` | NEW | T-114 |
| F-011: Keyhint bar does not advertise `/` when list focused; help overlay gives no scope note | P3 | `internal/ui/appshell/help.go:26-44`; `internal/ui/appshell/keyhints.go` | NEW | T-121 |
| F-012: Cavekit R7 does not specify content-source contract for search | P3 (cavekit gap) | `context/kits/cavekit-detail-pane.md:81-91` | CLOSED | Kit revision 2026-04-18 (R7 now mandates soft-wrapped unstyled content) |

## Context carried forward for the next `/ck:make`

- Tier 12 tasks T-113..T-122 address F-001..F-011. Execute in listed order: T-113 must land before T-114..T-121 (provides `ContentLines()`); T-122 is HUMAN sign-off and must be last.
- `context/impl/impl-detail-pane.md` T-043 row downgraded from DONE to PARTIAL — `SearchModel` exists but not integrated. Do not treat as complete.
- All HUMAN sign-off on this feature must follow the `cavekit-overview.md` "Verification Conventions" (tui-mcp, `logs/small.log`, all three themes, both orientations, both small/large geometries).

## Unrelated UX issue noted (not acted on)

- Iteration 20 loop-log recorded "help overlay (`?`) shows text bleed-through from underlying view (overlay does not opaquely cover content cells)". Out of search scope — does not have a task in Tier 12. File under a future UX-polish tier if/when addressed.

---

## /ck:check run 2026-04-18 (Tier 13 — detail-pane nav + height + focus)

Source: user report — "details pane only portion of content visible even if space below (tiny.log:34)", "no cursor position shown", "cannot navigate to beginning/end with g/G", "no need to autofocus on details when opening". Scope: detail-pane scroll + height + open-focus flow (`internal/ui/detailpane/{model,scroll,height,wrap}.go`, `internal/ui/app/model.go` relayout + openPane + handleKey, `internal/ui/appshell/layout.go`).

| Finding | Severity | File:line | Status | Addressed by |
|---|---|---|---|---|
| F-013: Detail pane vertical height uses below-mode `height_ratio` in right-split → content clipped | P0 | `internal/ui/app/model.go:162, 523-536`; `internal/ui/appshell/layout.go:43-86`; `internal/ui/detailpane/height.go:29-35` | NEW | T-123 |
| F-014: `relayout()` never calls `SetHeight` on the pane | P1 | `internal/ui/app/model.go:523-536` | NEW | T-123 |
| F-015: No g/G/PgUp/PgDn/Home/End/Ctrl+d/Ctrl+u bindings in detail pane scroll | P1 | `internal/ui/detailpane/scroll.go:74-92` | NEW | T-124 |
| F-016: No scroll-position indicator rendered in detail pane | P1 | `internal/ui/detailpane/model.go:235-283`; `internal/ui/detailpane/scroll.go:95-104` | NEW | T-125 |
| F-017: Opening the pane unconditionally steals focus from the list | P1 | `internal/ui/app/model.go:512-521` | NEW | T-126 |
| F-018: `SoftWrap` re-wrap passes outer pane height (not content height) to `scroll.SetContent` | P2 | `internal/ui/detailpane/model.go:53, 87` | NEW | T-123 |
| F-019: `PaneModel.View()` mutates `scroll.height` without re-clamping offset | P2 | `internal/ui/detailpane/model.go:256-261` | NEW | T-123 |
| F-020: `PaneModel.Open()` hardcodes `hiddenFields=nil` — R5 compliance gap | P2 | `internal/ui/detailpane/model.go:77-89`; `internal/ui/app/model.go:56, 102, 512-521` | NEW | T-127 |
| F-021: DESIGN.md §4.4 + §9 keymap missing nav keys + scroll-position indicator spec | P2 | `DESIGN.md:224-244, 573-588` | NEW | T-128 |
| F-022: `ScrollToLine` scroll math assumes `viewport == scroll.height` (self-resolves with F-013) | P3 | `internal/ui/detailpane/model.go:152-170` | NEW | T-123 (secondary) |
| F-023: `paneHeight.PaneHeight()` can return 1 on tiny terminals (auto-close catches — noted only) | P3 | `internal/ui/detailpane/height.go:29-35` | NOTED | — |
| F-024: After F-017 fix, Esc from list-focus with pane open must close pane | P3 | `internal/ui/app/model.go` handleKey FocusEntryList branch | NEW | T-126 |
| F-025: Clamp must run after every height/content change on scroll model | P3 | `internal/ui/detailpane/scroll.go` + `model.go:256-261` | NEW | T-123 |

### Context carried forward for the next `/ck:make`

- Tier 13 tasks T-123..T-129 address F-013..F-022, F-024, F-025. Execute T-123 first — it is the P0 content-loss fix and provides the Layout-derived height helper the other tasks assume works.
- T-127 closes an unrelated R5 compliance gap discovered during review (VisibilityModel never wired into `PaneModel.Open`). Not user-reported, but same codepath — pick up in the same cycle.
- T-128 updates DESIGN.md §4.4 + §9; run it AFTER T-124/T-125 so the doc matches the implementation rather than leading it.
- T-129 is the HUMAN sign-off gate — must use tui-mcp per cavekit-overview.md "Verification Conventions" with `logs/tiny.log` line 34 as the primary reproducer across all three themes and both orientations at 80x24 + 140x35.
- All of Tier 12 is complete per loop-log iteration 20; Tier 13 starts from a clean base at commit `75cd5d3`.

---

## /ck:check run 2026-04-18 (Tier 14 — cursor-tracking viewport + shared scrolloff)

Source: user report — "I still see no row highlight where cursor is when focused on details pane" + follow-up "scrolloff should be implemented on the list as well" + "in nvim if the cursor in the middle the window and I start scrolling down using mouse, the cursor like goes up until it's the 5th topmost row, when scrolling continues that 4 top rows before cursor is maintained — same happens at bottom — so there's a sort of context of row around the highlighted row". Scope: detail-pane viewport semantics (`internal/ui/detailpane/{model,scroll}.go`), entry-list cursor+scroll coupling (`internal/ui/entrylist/list.go`), shared config key (`internal/config/config.go`), DESIGN.md §4 + §4.3 + §4.4 + §6 + §9.

Verdict: REVISE — code is spec-compliant against pre-revision kits, but DESIGN.md §4 matrix "Cursor row (list only)" explicitly excluded the detail pane from cursor-row semantics. User expects a vim-like cursor-tracking viewport with nvim `scrolloff` drag on BOTH panes. Amendment: drop the list-only scope, redefine detail pane as cursor-tracking, add shared top-level `scrolloff` config, integrate search `n`/`N` with cursor.

| Finding | Severity | File:line | Status | Addressed by |
|---|---|---|---|---|
| F-026: Detail pane has no per-line cursor when focused; no shared scrolloff on list or detail pane; search n/N scrolls viewport but does not place a cursor | P1 | `internal/ui/detailpane/scroll.go` (offset-only; no `cursor` field); `internal/ui/detailpane/model.go:358-411` (View has no cursor-row render); `internal/ui/entrylist/list.go` (scroll.Cursor + scroll.Offset but no scrolloff follow); `internal/config/config.go:14-20, 38-55` (no top-level `scrolloff` key); `DESIGN.md §4 matrix line 167` (pre-revision: "Cursor row (list only)") | NEW | T-130, T-131, T-132, T-133, T-134, T-135, T-136, T-137 |

### Context carried forward for the next `/ck:make`

- Tier 14 tasks T-130..T-137 address F-026. Execute T-130 first — it is the shared config foundation all other tasks depend on.
- T-131..T-134 are detail-pane tasks (cursor render, nav, mouse wheel, search). T-135 is the entry-list mirror (list already has `scroll.Cursor` so this is adding scrolloff follow + wheel drag). T-136 is a doc-consistency pass after implementation; T-137 is HUMAN sign-off.
- DESIGN.md was amended in this `/ck:check` run (§4 matrix, §4.3 "Shared scrolloff", §4.4 "Cursor and scrolloff", §6 cue #4, §9 keymap `j`/`k` + Mouse wheel rows).
- Kit revisions in this run: cavekit-detail-pane.md R11 (NEW), cavekit-entry-list.md R12 (NEW), cavekit-config.md R5 (extended with shared `scrolloff`).
- README was also updated in this run to reflect the new cursor + scrolloff behaviour and to replace the stale "Sonnet only" mention with "Claude" (multi-model).
- All of Tier 13 is complete per loop-log iteration 25; Tier 14 starts from a clean base at commit `5fe9184`.

---

## /ck:check run 2026-04-18 (post-Tier 14 — clipboard feedback, list search, detail-pane rendering)

Source: user notes — (list) "no search", (list) "copy of marked message does not copy the entries", (details pane) "on small screen when details opened with max preset width, lines overflow (tiny.log:1, logger property)", (details pane) "strange prop value coloring (tiny.log:45, msg prop value when cursor moved into wrapped lines)", (details pane) "cursor background color not shown after prop — `\"prop\": \"value\",` coloring ends before `:` and continues after `,`". Scope: clipboard flow (`internal/ui/app/model.go:408-418`, `internal/ui/appshell/clipboard.go`), entry-list search (missing feature), detail-pane rendering (`internal/ui/detailpane/render.go:42-52, 88-144`, `internal/ui/detailpane/model.go:55-74, 342-365, 459-468`, `internal/ui/detailpane/wrap.go:20-35`, `internal/ui/appshell/layout.go:66-86`).

Verdict: REVISE — two P1 findings (F-102 silent clipboard failure, F-105 non-contiguous cursor bg). F-103 cannot be confirmed from code alone (code shows underfill, not overflow; requires live tui-mcp repro to distinguish overflow-from-double-subtract vs underfill-from-double-subtract depending on which width is fed in). Amendments: R9 app-shell (clipboard feedback ACs), R9/R10/R11 detail-pane (wrap SGR, single-owner border, contiguous cursor bg), R13 entry-list (NEW list search), R13 app-shell (focus-based `/` routing).

| Finding | Severity | File:line | Status | Addressed by |
|---|---|---|---|---|
| F-101: Entry list has no search — `/` falls through to detail-pane routing only (cavekit gap) | P2 | `internal/ui/entrylist/list.go` (no search field); `internal/ui/app/model.go` `/` routing to detail pane | NEW | T-143, T-144, T-145 |
| F-102: `y` clipboard copy is silent — error from `clipboard.WriteAll` discarded via `//nolint:errcheck`; `ClipboardCopiedMsg` dropped; no status-bar notice on success, no notice on error, no notice on zero marks | P1 | `internal/ui/app/model.go:408-418`; `internal/ui/appshell/clipboard.go` (ClipboardCopiedMsg / ClipboardErrorMsg defined but unused) | NEW | T-138 |
| F-103: Detail-pane width accounting is broken in BOTH orientations — **right-orientation overflow**: pane extends past terminal right edge (no right border rendered, content clipped at terminal col); **below-orientation underfill**: pane is narrower than the list pane with a 14-col dead zone on the right (`DetailContentWidth()` returns 0 in below-mode → `SetWidth(0)` → pane falls back to prior/stale width). Root cause: layout publishes a CONTENT width via `DetailContentWidth()` but `SetWidth`/`contentWidth()` treats `m.width` as OUTER width and subtracts borders again (`m.width - 2`). Single-owner border accounting is missing. **Live-repro confirmed via tui-mcp** 2026-04-18: on 110×24 right-orientation, detail pane `┌──────...` has no closing `┐` (clips at col 110); below-orientation pane is 66 cols vs list 80 cols (14-col underfill). | P1 (visually confirmed — **right-split content loss past right edge**) | `internal/ui/detailpane/model.go:55-74, 459-468`; `internal/ui/appshell/layout.go:66-86`; `internal/ui/app/model.go:169, 574` (call sites passing `DetailContentWidth` — content width — into `SetWidth` which expects outer) | NEW | T-139 |
| F-104: Soft-wrap does not preserve SGR across break — `ansi.HardwrapWc(line, width, true)` with default `preserveSGR=false`; styled value spanning wrap point loses bg/fg on continuation line | P1 | `internal/ui/detailpane/wrap.go:20-35`; `internal/ui/detailpane/render.go:42-52` (per-token `Style.Render()` emits `\x1b[0m`) | NEW | T-140 |
| F-105: Cursor-row bg non-contiguous — per-token `lipgloss.Style.Render()` emits `\x1b[…m … \x1b[0m` segments; outer `paintCursorRow` applies bg + width but lipgloss does NOT re-inject outer SGR across inner resets; bg visually terminates at `:`/`,`/spaces between styled segments | P1 (directly matches user note "coloring ends before `:` and continues after `,`") | `internal/ui/detailpane/model.go:342-365` (paintCursorRow); `internal/ui/detailpane/render.go:42-52, 88-144` (per-token render sources the inner resets) | NEW | T-141 |

### Context carried forward for the next `/ck:make`

- Tier 15 tasks T-138..T-142 address F-102, F-103, F-104, F-105. Execute T-138 first (independent clipboard fix); T-139/T-140/T-141 detail-pane render fixes can proceed in parallel; T-142 is HUMAN sign-off and must be last.
- Tier 16 tasks T-143..T-145 add the new list-search feature (cavekit-entry-list.md R13 + cavekit-app-shell.md R13 re-routing). T-145 is HUMAN sign-off.
- F-103 live-repro completed via tui-mcp 2026-04-18 on `logs/tiny.log`: BOTH directions of the bug are present simultaneously. Right-orientation at 110×24: pane extends past terminal right edge, no right border, content clipped. Below-orientation at 80×24: pane is 66 cols vs list 80 cols. The T-139 fix must: (a) standardize on either outer-width or content-width as the single contract between layout and pane (pick one — outer is more common in lipgloss patterns), (b) remove the second border subtract from `contentWidth()`/`View()`, (c) make `DetailContentWidth()` return a non-zero value in below-orientation (currently returns 0, breaking the below-mode path entirely). HUMAN sign-off T-142 must verify right-split at 110×24 + 140×35 shows the closing `┐` border visible and no content past it, AND below-mode at 80×24 shows the detail pane outer width equal to list outer width.
- Secondary observation during repro: the `|` ratio-preset cycle in right-orientation does not visibly change the pane widths in the rendered output even though the internal ratio state updates (both 0.10 and 0.70 presets render at the same visible widths). Likely adjacent to F-103 — once width accounting is single-owner, the ratio math should also take effect. If not, surface as F-106 in a follow-up check.
- Kit revisions in this run: cavekit-app-shell.md R9 (expanded to 10 ACs), cavekit-app-shell.md R13 (rewritten to focus-based routing), cavekit-entry-list.md R13 (NEW), cavekit-detail-pane.md R9 (SGR-across-wrap AC added), cavekit-detail-pane.md R10 (single-owner border AC added), cavekit-detail-pane.md R11 (contiguous cursor bg AC added + human sign-off on `tiny.log:45`).
- DESIGN.md §4 matrix + §9 keymap must be amended to mention the cursor-row contiguity clause and list-search keymap row respectively.

---

## /ck:check run 2026-04-18 (post-Tier 15/16 — intercept ordering, focus-loss, streaming matches)

Source: automated `/ck:check` after Tier 15 + Tier 16 HUMAN sign-off. Goal-backward verifier confirmed 52/55 Tier-15/16 ACs MET, 0 STUB, 0 falsely_complete tasks, 485/485 tests P. Surveyor reported 50/56 reqs COMPLETE; 5 PARTIAL + 1 MISSING + 1 DESIGN VIOLATION are **pre-existing** filter-subsystem latency (not introduced this loop) and the help-overlay chrome. Inspector flagged 10 review findings below.

Verdict: **REVISE** — P1 F-106 is a user-visible data-loss regression introduced by T-144-fix (global `q`-quit intercept pre-empts active list-search input). Kit gaps codified as cavekit-app-shell R14 (NEW) + cavekit-entry-list R13 AC 7 broadening + R13 streaming AC.

| Finding | Severity | File:line | Status | Addressed by |
|---|---|---|---|---|
| F-106: Global `q`-quit intercept pre-empts list-search input mode — user typing query containing `q` quits app with data loss | P1 | `internal/ui/app/model.go:292-297` (fires before FocusEntryList search redirect at `:400`) | NEW | T-146 |
| F-107: Dead function `SearchModel.searchHighlightStyle()` + unused `lipgloss` import | P2 | `internal/ui/entrylist/search.go:208-213` | NEW | T-149 |
| F-108: `f` focus-transfer to filter panel does NOT clear list-search state; keyhint suppresses but View() still injects search prompt | P2 | `internal/ui/app/model.go:427-431`; kit AC 7 only enumerated Tab + filter-change | NEW | T-147 |
| F-109: Streaming entries arriving during active list search don't update match set or `(cur/total)` counter | P2 | `internal/ui/entrylist/list.go:92-97` (`AppendEntries`); `internal/ui/app/model.go:199-237` (EntryBatchMsg/TailMsg) | NEW | T-148 |
| F-110: Clipboard notice hides list-search prompt for 2-3s when both coincide — contradicts R13 AC 3 literal reading | P2 | `internal/ui/app/model.go:573-575` gates on `!HasNotice()` | NEW | deferred (kit amendment proposal or UX compose) |
| F-111: `composeSearchRow` byte-indexed msg truncation can split multi-byte runes (mirrored pre-existing bug from `row.go:76`) | P3 | `internal/ui/entrylist/search.go:192-194` + `internal/ui/entrylist/row.go:76-78` | NEW | deferred — scope beyond T-143 |
| F-112: `SearchModel` shape diverges between entrylist (`Deactivate`/`InputMode()`) and detailpane (`Dismiss`/`Mode()`) | P3 | `internal/ui/entrylist/search.go` vs `internal/ui/detailpane/search.go` | NEW | deferred — quality only |
| F-113: `preserveSGRAcrossBreaks` accumulates unbounded SGR opens without dedup | P3 | `internal/ui/detailpane/wrap.go:62-91` | NEW | NOTED — not observable at current scale |
| F-114: `computeMatches` O(rows × query-composition) per keystroke, no benchmark / debounce | P3 | `internal/ui/entrylist/search.go:149-165` | NEW | NOTED — defer until large-log lag reported |
| F-115: No test pins behavior for "list-search active → `f` → ???" | P3 | `internal/ui/app/model_test.go` | NEW | bundled with T-147 |

### Kit revisions in this run

- **cavekit-app-shell.md R14 (NEW)** — "Global Key-Intercept Ordering under Active In-Pane Search". Codifies that `q`/`Tab`/`?`/`Esc` global reservations must not pre-empt pane search in input mode. Adds 5 ACs covering the `q`-quit exemption (both list + pane search), `Tab` dismissal-via-focus-cycle, `?` help preserving + restoring search state, and `q` reverting to Quit after search dismissed. Driven by F-106 (P1).
- **cavekit-entry-list.md R13 (AC expansion)** — AC 7 broadened from "Tab cycle OR filter change" to enumerate ALL focus-loss triggers (Tab, `f` transfer, mouse click off list). New AC added for streaming entries: `AppendEntries` with active search recomputes match set for appended slice and keeps `(current/total)` live. Driven by F-108 + F-109 (both P2).

### Context carried forward for the next `/ck:make`

- Tier 17 (T-146..T-150) addresses F-106, F-107, F-108, F-109, F-115. Execute T-146 first — P1 data-loss blocker. T-147/T-148/T-149 are disjoint-file and parallelizable. T-150 HUMAN gate last.
- F-110 (clipboard notice hides search prompt) — pick up later; likely a UX compose fix + an AC 3 kit carve-out. Not a build blocker.
- F-111 (byte-indexed truncation) mirrors pre-existing `row.go:76` bug — fix both sites or adopt `ansi.Truncate` uniformly in a future quality tier.
- F-112/F-113/F-114 are quality/perf polish items; defer unless user reports lag or inconsistency.
- Surveyor-flagged pre-existing gaps (filter subsystem invisibility: filter-engine R4/R5/R6, detail-pane R5/R8, entry-list R4 two-level cursor, app-shell R5 help-overlay chrome) are NOT included in Tier 17 — they predate this loop. Candidates for a future "filter UX" tier if/when user surfaces them.

---

## /ck:check run 2026-04-18 (post-Tier 17 — test-polish + DRY helper)

Source: automated `/ck:check` after Tier 17 code landed (T-146 + T-147 + T-148 + T-149). Frontier is T-150 HUMAN sign-off via tui-mcp.

Verdict: **APPROVE**. Goal-backward verifier reports 12/12 ACs MET (R14 five ACs + R13 AC 7 four focus-loss branches + R13 streaming three ACs). Surveyor 10/12 COMPLETE, 2 PARTIAL (test-pin only, not implementation). Inspector flags 2 quality findings. Build P, tests P (493/493 across 11 packages). No regressions, no kit-semantic gaps, no DESIGN violations.

| Finding | Severity | File:line | Status | Addressed by |
|---|---|---|---|---|
| F-116: `ExtendMatches` duplicates case-fold + substring logic from `computeMatches` — must stay in lock-step forever | P2 | `internal/ui/entrylist/search.go:154, ~179` | FIXED | T-151 — shared `matchRow(lowerNeedle, entry, width, cfg)` helper in `internal/ui/entrylist/search.go`; both paths call it |
| F-117: `TestModel_Q_ListSearchInputMode_DoesNotQuit` asserts only one keystroke (`q`) — no chained-keystroke coverage | P3 | `internal/ui/app/model_test.go:1351-1378` | FIXED | T-152 — extended test with chained `u`/`i`/`t`, asserts `Query()=="quit"` and no QuitMsg across sequence |
| F-118: R14 AC 5 (`?`-Esc preserves search state) has no regression test — preservation is correct by construction but silent-drift risk | P3 | `internal/ui/app/model_test.go` (no such test) | FIXED | T-153 — new `TestModel_HelpOverlay_PreservesListSearchState` pins Active+Query+InputMode survival |
| F-119: R13 AC 7 mouse-click-off-list branch (`ZoneDetailPane` press clears search) has no test — only focus-change is pinned | P3 | `internal/ui/app/model_test.go:591` pins focus but not search-clear | FIXED | T-154 — new `TestModel_MouseClick_DetailZone_ClearsActiveListSearch` pins paired focus-transfer + search-deactivate |
| F-120: R13 AC 7 wording "mouse click in any non-list zone" is broader than implementation — Header/StatusBar/Divider clicks do NOT clear search (pre-existing, not Tier 17 regression). True semantic likely "focus-loss" | P3 (kit-wording) | `internal/ui/app/model.go` handleMouse (only `ZoneDetailPane` handled) | NEW | deferred — kit tightening, no code change |

### Kit revisions in this run
None. R14 and R13 expansions are MET in code; the test-pin gaps do not require new ACs. F-120 is a wording-tightening opportunity but deferred until user surfaces concrete confusion.

### Context carried forward for the next `/ck:make`

- T-150 HUMAN sign-off via tui-mcp still outstanding — the frontier per `context/impl/loop-log.md` iteration 34. Sign-off gates Tier 17 Codex tier review.
- Tier 18 (T-151..T-154) is polish-only — NOT a build blocker. Can proceed in any order, all disjoint files. No kit amendments required.
- F-120 kit-wording tightening is deferred; pick up with a broader "focus-loss semantics" pass if a user reports a surprising behavior on Header/StatusBar click.
- Pre-existing deferred items (F-110..F-114 from previous check) still apply.

---

## /ck:make run 2026-04-19 (Tier 19 Wave 2 — T-156 mouse-drag)

Source: in-flight T-156 implementation of cavekit-app-shell R15 mouse-drag resize.

| Finding | Severity | File:line | Status | Addressed by |
|---|---|---|---|---|
| F-121: cavekit-app-shell R15 AC 4 text is directionally inverted — "below-mode: dragging down grows the detail pane" contradicts the physical layout (detail is BELOW the divider; dragging divider down SHRINKS detail). Right-mode AC is physically correct and symmetric with the intended behaviour. Implementation follows physical correctness (down=shrink, up=grow) and the right-mode symmetry, NOT the literal kit text. | P3 (kit-wording) | `context/kits/cavekit-app-shell.md` R15 AC 4 | NEW | deferred — kit-text tightening (flip "down"↔"up" in AC 4); no code change |

### Context carried forward for the next `/ck:make`

- F-121 is a documentation inversion in the kit; the implementation and tests are physically correct. When Tier 19 HUMAN sign-off (T-159) runs via tui-mcp, the user will observe the physically-correct behaviour. If the user is confused by the kit text, rewrite R15 AC 4 to "dragging up grows the detail pane (divider moves up, list shrinks, detail area grows); dragging down shrinks it".

---

## /ck:make run 2026-04-19 (Tier 19 Wave 5 — T-159 HUMAN sign-off surface)

Source: T-159 HUMAN sign-off executed via tui-mcp on `logs/small.log` at 140x35 right-split and 80x24 below-mode (tokyo-night).

| Finding | Severity | File:line | Status | Addressed by |
|---|---|---|---|---|
| F-122: Router/renderer divider-col mismatch (right-split). `appshell.MouseRouter.Zone` computes `listEnd = ListContentWidth()+1` and `divider = listEnd+1` (treating ListContentWidth as CONTENT width), but `app/model.go:178` sends `WindowSizeMsg{Width: ListContentWidth()}` and `entrylist/list.go:308` subtracts 2 to derive its inner viewport (treating ListContentWidth as OUTER/allocated width). Net effect at 140x35 width_ratio=0.3: router thinks divider is at col 96, Lipgloss renders visible divider `│` at col 94. A user clicking the visible divider lands in ZoneEntryList and never triggers T-156 drag initiation. Pre-existing (predates Tier 19) — NOT a T-156/T-158 regression; surfaced during T-159 sign-off. | P1 (visible feature gap) | `internal/ui/appshell/mouse.go:66-67` vs `internal/ui/app/model.go:178` + `internal/ui/entrylist/list.go:308` | NEW | deferred — next tier should decide: either rename `ListContentWidth()` + audit all call sites, or fix the router's `listEnd` to subtract one. Existing unit tests use the router's coord convention so they pass, but real clicks at the visible divider don't. |
| F-123: `RatioFromDragY` off-by-one inverse of `HeightModel.PaneHeight`. T-156's helper uses `detail = termHeight - 2 - y` but the correct inverse of `DetailPaneHeight = int(termHeight * ratio)` given the below-mode layout (detail occupies rows `entryListEnd+1 .. termHeight-2`) is `detail = termHeight - 1 - y` (divider row is detail's top-border row and counts as part of the detail allocation). At termHeight=24 and the current divider row y=16 (which renders when ratio=0.30), the formula returns ratio=0.25 instead of 0.30 — so a Press directly on the visible divider Y snaps the ratio down by one preset step. T-156's unit tests use synthetic MouseMotion events with coords derived from the same off-by-one helper, so they pass trivially. | P2 (ratio step-snap on Press) | `internal/ui/appshell/ratiokeys.go` (`RatioFromDragY`) | NEW | deferred — one-line constant fix. Write a corresponding test that pins ratio-on-Press-at-current-divider is unchanged, not shifted. |
| F-124: tui-mcp emits only Press and Release mouse events — no intermediate Motion. Bubble Tea's drag loop depends on MouseMotion between Press and Release to progressively update the ratio. Consequence for HUMAN sign-off: drag behaviour can only be observed as "Press snaps to start-Y, Release saves once"; the proportional-motion AC (R15 AC 2) cannot be visually re-verified through tui-mcp. Real-world terminal emulators DO emit motion events, so production drag works and the unit tests pin it in-source. | P3 (tooling constraint, not product bug) | n/a (tui-mcp event model) | NEW | deferred — documented here so future T-159-equivalent HUMAN sign-offs don't treat the motion-phase absence as a regression. If stricter drag verification is needed, use `xdotool` or real X11-level event injection instead of tui-mcp. |

### Context carried forward for the next `/ck:make`

- F-122 is the more consequential of the three — it's a real UX regression surface (clicking the visible divider doesn't drag) that has been masked by unit tests agreeing with the router's (incorrect) divider column. Fixing it is a small-scope follow-up tier.
- F-123 fix is one constant + one test. Bundle with F-122 in the same follow-up.
- F-124 is tooling-only; carry forward as a note in the sign-off convention.

---

## /ck:check run 2026-04-19 (post-Tier 19 — divider-coord alignment + drag state-machine hardening)

Source: automated `/ck:check` after Tier 19 close (T-155..T-159 all DONE; 548/548 tests P). Surveyor verdict **REVISE**: 25/27 Tier-19 ACs MET, R15 AC 2 + AC 5 PARTIAL due to F-122 + F-123 which were deferred-during-loop but are real user-visible gaps. Inspector adds 5 new findings (F-125..F-130), re-confirms F-122 as P1 Tier-accept blocker, and withdraws F-131 after second audit.

| Finding | Severity | File:line | Status | Addressed by |
|---|---|---|---|---|
| F-122: Router/renderer divider-col mismatch (re-confirmed P1) | P1 | `internal/ui/appshell/mouse.go:66-67` vs `internal/ui/app/model.go:178` + `internal/ui/entrylist/list.go:308` | NEW | T-160 |
| F-123: `RatioFromDragY` off-by-one inverse of `PaneHeight` | P2 | `internal/ui/appshell/ratiokeys.go:146` | NEW | T-161 |
| F-125: Drag state survives pane auto-close on terminal resize | P2 | `internal/ui/app/model.go:157-188, 498-521` | NEW | T-162 |
| F-127: `rowForY` has no defensive lower bound when `contentTopY` unset | P3 | `internal/ui/entrylist/list.go:280-291` | NEW | T-163 |
| F-129: Bare Press+Release on divider rewrites config (no-motion write amplification) | P3 | `internal/ui/app/model.go:502-506` | NEW | T-164 |
| F-126: `RatioFromDragY/X` silently shadows persisted ratio when termHeight/Width ≤ 0 | P3 | `internal/ui/appshell/ratiokeys.go:117-119, 142-154` | NEW | T-165 |
| F-130: Unreachable `if detail > termHeight` branch in `RatioFromDragY` (dead code) | P3 | `internal/ui/appshell/ratiokeys.go:149-151` | NEW | T-166 |
| F-128: `|` at detail=0.50 flips across focus (both preset lists alias) | P3 | `internal/ui/appshell/ratiokeys.go:78-98` | DEFERRED | — (doc/kit tightening only; picked up later if user surfaces confusion) |
| F-131: R12 AC 6 clamp wording | P3 | n/a | WITHDRAWN | — (inspector audit: wording is letter-correct) |
| DESIGN VIOLATION: `DESIGN.md §5` shows 3-preset set | P3 | `DESIGN.md §5` | NEW | T-167 |
| DESIGN VIOLATION: `DESIGN.md §9` authoritative keymap shows 3-preset set + no R15 mouse-drag row | P3 | `DESIGN.md §9` | NEW | T-167 |
| F-121: R15 AC 4 below-mode text directionally inverted | P3 (kit-text) | `context/kits/cavekit-app-shell.md` R15 AC 4 | FIXED | T-168 (kit-text edit landed inline in this /ck:check — AC 4 now reads "dragging up grows the detail pane") |

### Kit revisions in this run

- **cavekit-app-shell.md R15** — 5 new ACs added:
  - **Renderer-truth divider-col assertion** (F-122): the router's divider cell MUST coincide with Lipgloss's rendered `│` glyph; a test must render the layout, locate the glyph programmatically, and assert `MouseRouter.Zone(glyphX, midY) == ZoneDivider` across presets `{0.10, 0.30, 0.50, 0.80}` in both orientations at 140x35 and 80x24.
  - **Inverse-math invariant** (F-123): when `RatioFromDragY/X` inverts `PaneHeight/Width = int(termDim * ratio)`, a Press on the currently rendered divider row/col MUST yield the current ratio unchanged.
  - **Mid-drag auto-close termination** (F-125): if R7 auto-closes the pane mid-drag, drag state must die — subsequent Motion is swallowed, no ratio mutation, no config write on eventual Release.
  - **No-motion-no-persist** (F-129): bare Press+Release with no intermediate Motion must not rewrite config.
  - **Degenerate-dim guard** (F-126): drag helpers preserve current ratio when termWidth/Height ≤ 0 rather than jumping to `RatioDefault`.
- **cavekit-app-shell.md R15 AC 4** — text flipped from "dragging down grows detail" to "dragging up grows" to match physical layout (closes F-121 kit-text inversion).
- **cavekit-entry-list.md R10** — 1 new AC: list MUST reject clicks when `contentTopY` has not been injected — defensive lower bound on the single-owner resolver contract (closes F-127 regression vector).

### Context carried forward for the next `/ck:make`

- Tier 20 tasks T-160..T-170 address F-122, F-123, F-125, F-126, F-127, F-129, F-130, DESIGN §5/§9 sync, and F-121 placeholder. Execute T-160 first — P1 Tier-accept blocker; F-122 fix also anchors the renderer-truth test that prevents the whole "unit-tests agree with router" failure mode.
- T-161 is a one-constant flip (+ Press-at-current-Y test) — parallelizable with T-160 (different files).
- T-162 (mid-drag auto-close) and T-164 (no-motion no-save) and T-165 (degenerate-dim guard) are all `app/model.go` + `appshell/ratiokeys.go` touches — inline-sequence to avoid merge friction.
- T-163 (unset-contentTopY safety) and T-166 (dead-code removal) are disjoint-file XS tasks; can parallelize.
- T-167 (DESIGN.md §5 + §9) waits on T-155 being landed (it already is) — touch DESIGN.md after kit amendments so docs match code.
- T-170 HUMAN gate: verify F-122 + F-123 + F-125 + F-129 behaviour via tui-mcp; F-124 (tui-mcp lacks Motion events) means motion-phase AC 2 continues to rely on unit-test evidence.
- F-128 (`|` flips across focus at 0.50) is deferred — debatable UX (user may find the aliasing natural); bring back if confusion surfaces.
- F-131 withdrawn after second audit; kept in the row above for traceability.

---

## /ck:check run 2026-04-19 (Tier 21 — F-132/F-133/F-134 review-followups branch)

Source: automated `/ck:check` on branch `tier-21-review-followups` (6 commits from `main`). Scope is narrow — test-fidelity hardening + one inverse-math bug fix — in response to a prior `/ck:review` Pass 2 over the Tier 20 branch. Three-agent review: **ck:verifier** (goal-backward on R15 ACs), **ck:surveyor** (gap analysis + backprop-log audit), **ck:inspector** (peer review). Build P. Tests 574/574 pass across 11 packages.

### Coverage of new/extended R15 ACs

| R15 AC | Mandate added/extended | Status | Evidence |
|---|---|---|---|
| L198 renderer-truth (ANSI-strip) | ANSI-stripping helper MUST handle full ECMA-48 CSI two-step form (`ESC [ params/intermediates 0x40..0x7e`) | COMPLETE | `mouse_test.go:237-266` 3-state machine; `mouse_test.go:279-302` 9 sub-case regression test; reverting to hardcoded subset → 9/9 fail |
| L199 inverse-math (parallel-axis) | X-axis Press-at-current pin MUST source canonical `dividerX` from `Layout.ListContentWidth()` (renderer-truth), not inverse formula | COMPLETE | `ratiokeys.go:121-134` `detail := usable - x` exact inverse; `ratiokeys_test.go:233-257` non-tautological X pin; T-104 mid pin 48/95 → 45/95; reverting formula → 5/5 drift > RatioStep/2 |
| L202-203 degenerate-dim | Test MUST drive `termW/termH<=0` guard with `pane.IsOpen()==true` simultaneously; removing guard MUST fail test | COMPLETE | `model_test.go:2424-2501` drives WindowSizeMsg→auto-close→re-open→Motion; removing guards shadows 0.55→0.30 (right) and 0.45→0.30 (below) |

### Backprop-log audit (`.cavekit/history/backprop-log.md`)

| Entry | Finding | SHAs resolve | Test refs | Verification | Verdict |
|---|---|---|---|---|---|
| #1 | F-132 (test-fidelity / validation-via-wrong-path) | `97c1b9b` ✓ | ✓ | mutation-consistent | CONSISTENT |
| #2 | F-133 (test-fidelity / parallel-axis-coverage) | `68d2548` + `fd5a26d` ✓ | ✓ | drift math checks out (100−50−2=48→0.505 vs need 0.55) | CONSISTENT |
| #3 | F-134 (test-fidelity / latent-fragility-in-test-helper) | `024d429` + `91df657` + backfill `3169c71` ✓ | ✓ | 3-state machine required because `[` is in 0x40..0x7e | CONSISTENT |

### Peer review findings

| Finding | Severity | File | Status |
|---|---|---|---|

No new P0/P1/P2 findings. Inspector logged 4 P3 observations explicitly classified as **non-defects, no action required**:

- O-001 (`mouse_test.go:228-266`) — stripAnsi handles CSI but not OSC/DCS/SOS/PM/APC. R15 line-198 mandate is scoped to CSI; Lipgloss does not emit OSC today. Open a new finding if OSC 8 hyperlinks land.
- O-002 (`mouse_test.go:253-256`) — literal ESC inside `stCsiBody` is consumed as a body byte rather than restarting state. Well-formed terminal output does not embed nested ESC mid-CSI.
- O-003 (`ratiokeys_test.go:185-191`) — T-104 mid pin `45/95` reads as a magic number but the derivation comment at `:180-184` explains it (`usable=95, detail=usable-x=45`).
- O-004 (`.cavekit/history/backprop-log.md:67-68`) — backprop-log entry #3 now correctly cites `024d429` + `91df657`. No dangling SHA.

### Verdict

**APPROVE.** Coverage 4/4 COMPLETE on scoped R15 ACs. Backprop-log 3/3 CONSISTENT. No over-builds, no DESIGN.md drift (off-preset inverse-math changes are not user-visible). No kit amendments needed beyond what shipped inline in the 6 fix commits.

### Context carried forward for the next `/ck:make`

- Tier 21 adds no new tasks to `context/plans/build-site.md`. F-132/F-133/F-134 are closed as follow-ups to the Tier 20 branch review.
- Loop-log does not yet have a Tier 21 entry. Cosmetic housekeeping — pick up on next consolidation pass.
- `context/impl/impl-app-shell.md:74` (T-165 row) text still names the deleted `TestModel_T165_Drag_Zero{Width,Height}_PreservesRatio` tests; they were superseded by the F-132 tests which validly cover the same AC. Cosmetic drift only; the AC IS met.
- Pattern for future kit amendments: when writing a regression test, source the canonical input from the **other side** of the contract (renderer output, forward math, physical spec, reality-based precondition) — not from the thing being tested. F-132/F-133/F-134 were all the same failure mode: tests that agreed with the code because they used the same model.

---

## /ck:check run 2026-04-19 (Tier 22 — housekeeping: M-001 dead clamp + impl-tracking drift)

Source: automated `/ck:check` on branch `tier-22-housekeeping` (1 commit `4aa8a64` from `main`). Scope is pure housekeeping — one MINOR Pass 1 finding from the Tier 21 `/ck:review` (M-001) plus two cosmetic drift items flagged during the Tier 21 `/ck:check`. Two-agent review: **ck:surveyor** (gap analysis + loop-log audit), **ck:inspector** (peer review). No verifier dispatched (branch introduces zero new ACs). Build P. Tests 574/574 pass across 11 packages.

### Coverage

| Status | R15 ACs | Delta from Tier 21 |
|---|---|---|
| COMPLETE | 12 | +0 (no new ACs introduced) |
| PARTIAL | 0 | 0 |
| MISSING | 0 | 0 |
| OVER-BUILT | 0 | 0 |

### M-001 safety derivation (surveyor + inspector converge)

- Formula `detail := usable - x`; caller `model.go:557` passes `msg.X ≥ 0` (Bubble Tea column index invariant)
- Therefore `detail ≤ usable` by construction — the removed `if detail > usable { detail = usable }` branch was unreachable under caller contract
- Inspector ran 816-case algebraic sweep (including negative-x inputs down to -20) — old and new formulas diverge on zero inputs because `ClampRatio` at the tail catches any ratio > RatioMax
- Replacement comment cites F-130/T-166 Y-axis precedent inline; shape now symmetric with `RatioFromDragY`

### Loop-log audit

| Claim | Verdict |
|---|---|
| Iter 43 SHAs (97c1b9b, 68d2548, fd5a26d, 024d429, 91df657, 3169c71) + merge e8bf635 | CONSISTENT |
| Pre/post test counts 552 → 574 (+22 regression tests) | CONSISTENT with +9 stripAnsi + 5 X pin + 2 F-132 + support adjustments |
| Iter 44 test count 574/574 | CONSISTENT with live run |
| `impl-app-shell.md:74` T-165 row drift | RESOLVED (no longer names deleted tests) |

### Peer review findings

**0 P0 · 0 P1 · 0 P2 · 0 P3.** Inspector: *"Me swing club at clamp. Clamp already dead. Me hit air."*

### Verdict

**APPROVE.** Pure dead-code elimination plus factually-accurate tracking updates. Ready to merge `tier-22-housekeeping` → `main`.

### Context carried forward for the next `/ck:make`

- No Tier 23 frontier exists unless a new user request surfaces. Build-site remains complete across all 20 feature tiers; Tier 21 and Tier 22 were post-build housekeeping tiers driven by review findings.
- Cross-tier pattern confirmed: "dead-code guards surface when the forward/inverse math is tightened". Both T-166 (Y-axis, F-130) and M-001 (X-axis, post-F-133) followed the same shape — after fixing the inverse formula to the exact algebraic inverse, the upper-clamp becomes provably unreachable under the caller contract. Watch for a third instance if any other inverse-math helper is added.

---

## /ck:check run 2026-04-19 (Tier 23 — DragHandle drag-seam token)

Source: post-loop inspection after Tier 23's 5 tasks (T-171..T-175) landed on `tier-23-drag-handle`. Scope: `DragHandle` theme token + right-mode `│` glyph recolor + below-mode top-border override + cross-orientation rendered-cell test + tui-mcp HUMAN sign-off across 3 bundled themes × 2 orientations. 9 commits (`ed91d17`..`e96bcba`); 17 files changed (+511/-34); 596 tests pass (+22 over Tier 22 baseline).

**Coverage:** 3/3 requirements COMPLETE, 6/6 acceptance criteria MET. No falsely-complete, no gaps, no over-builds. Verifier + surveyor both clean.

### Peer review findings

| Finding | Severity | File:line | Status | Addressed by |
|---|---|---|---|---|
| F-200: `WithBelowMode` not refreshed on WindowSizeMsg auto-orientation flip — stale pane flag yields wrong seam SGR when `position=auto` crosses threshold with pane open | P1 | `internal/ui/app/model.go:162-198` vs `:717-739` | NEW | pending /ck:revise --trace --from-finding F-200 (proposes new R7 AC 9: orientation-flip re-render contract) |
| F-201: Kit+DESIGN "shared border row" framing vs. two-row physical reality (list bottom border + detail top border render as two adjacent rows; only detail top is overridden to DragHandle) | P2 | `cavekit-app-shell.md` R10 AC 10 + R15 AC 14; `DESIGN.md` §6; `internal/ui/appshell/panestyle.go:57-59` | NEW | pending /ck:revise --trace --from-finding F-201 (kit language clarification OR new AC requiring both rows) |
| F-202: `TestDragSeam_RendersInDragHandle_AllThemes/right` uses ASCII stub panes — zero marginal coverage over T-172's `divider_test.go` | P3 | `internal/ui/appshell/mouse_test.go:310-336` | NEW | (defer) add integrated right-mode subtest via `m.View()` |
| F-203: `TestDragSeam.../below` bypasses `PaneModel.View()`; tests only style layer — wiring path (`WithBelowMode` setter + `relayout` wiring, the F-200 vector) is not exercised | P3 | `internal/ui/appshell/mouse_test.go:338-362` | NEW | (defer) direct test in `detailpane` package: `NewPaneModel + Open + WithBelowMode(true) + View()` asserts DragHandle SGR on top row |
| F-204: `theme_test.go` luminance-ordering doc-comment says tui-mcp was unavailable; harness was fixed + HUMAN sign-off landed in `e96bcba`. Comment drift | P3 | `internal/theme/theme_test.go:54-63` | NEW | (defer) reword as belt-and-braces regression guard for the perceptual invariant |
| F-205: Review-scope prompt listed catppuccin DividerColor as `#45475a` (Dim); actual `DividerColor=#313244`. Test is correct (reads struct), only prompt/notes drifted | P3 | notes-only; no code defect | CLOSED | no action — flagged for future review-scope generation |
| F-206: `WithDragSeamTop` applies unconditionally; no empty-DragHandle guard (symmetry-break with `PaneStyle`'s `UnfocusedBg != ""` guard) | P3 | `internal/ui/appshell/panestyle.go:57-59` | NEW | (defer) optional symmetry guard; zero impact today (all bundled themes non-empty; no user-input path to theme creation) |

### Pre-existing site/kit drift (not Tier 23 caused — surfaced during survey)

- `internal/ui/appshell/CLAUDE.md` "Implements" list stops at R13; package implements R14 (key-intercept ordering) and R15 (mouse drag).
- `context/plans/build-site.md:830` app-shell section header reads "(12 requirements, 80 criteria)"; `:948` Coverage Totals row reads "13 requirements, 101 criteria". Kit has R1..R15 (15 requirements). Undercount by 2.

Both are housekeeping — recommend folding into the next `/ck:check`-driven housekeeping tier (similar scope to Tier 22).

### Verdict

**REVISE.** 100% kit coverage but one P1 finding (F-200) is a silent visual regression that survives green tests. F-200 reveals a missing R7 AC (orientation-flip re-render contract) — route through `/ck:revise --trace --from-finding F-200` so the kit amendment + regression test are authored under the single-failure protocol with explicit user approval per `.cavekit/` convention.

### Context carried forward for the next `/ck:make`

- F-200 fix is one-line wiring in `WindowSizeMsg` branch + one regression test. Backprop-log entry will record the trace.
- F-201 is kit-language-only by default; may escalate to a code task if the tribe decides both border rows should paint DragHandle in below-mode.
- Defer F-202..F-204, F-206 until a future polish tier unless they block downstream work.
- Pre-existing drift (appshell CLAUDE.md + build-site counts) is independent of F-200/F-201 trace — fold into the next housekeeping tier.
