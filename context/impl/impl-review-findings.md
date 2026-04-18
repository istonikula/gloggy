---
created: "2026-04-18T09:40:17+03:00"
last_edited: "2026-04-18T09:40:17+03:00"
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
