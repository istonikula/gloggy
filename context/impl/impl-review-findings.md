---
created: "2026-04-18T09:40:17+03:00"
last_edited: "2026-04-18T14:20:33+03:00"
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
