---
created: "2026-04-18T09:40:17+03:00"
last_edited: "2026-04-19T00:26:42+03:00"
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
