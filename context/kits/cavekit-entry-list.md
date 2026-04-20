---
created: "2026-04-15T00:00:00Z"
last_edited: "2026-04-20T21:07:43+03:00"
---

# Cavekit: Entry List

## Scope

The primary scrollable list of log entries displayed in a compact format. Covers row rendering, two-level cursor navigation, virtual scrolling, level-jump navigation, marks/bookmarks, entry selection signaling, and mouse interactions within the list. All visual styling defers to the active theme.

## Requirements

### R1: Compact Row Format
**Description:** Each JSONL entry row shows time (HH:MM:SS), level badge, abbreviated logger, and truncated message. Non-JSON entries show dimmed raw text. Which fields appear in the compact row is determined by config. The currently selected (cursor) row is visually highlighted with a distinct background color from the active theme, so the user can always see which row is selected. Each compact row must occupy exactly one terminal line — embedded newlines in messages or raw text must be flattened to spaces before rendering.
**Acceptance Criteria:**
- [ ] [auto] A JSONL entry row contains the time formatted as HH:MM:SS
- [ ] [auto] A JSONL entry row contains the level value
- [ ] [auto] A JSONL entry row contains the logger abbreviated to the configured depth
- [ ] [auto] A JSONL entry row contains the message, truncated to fit the available width
- [ ] [auto] A non-JSON entry row shows the raw text
- [ ] [human] Non-JSON entry rows are visually dimmed compared to JSONL rows
- [ ] [auto] An entry with zero time displays a placeholder (e.g. blank or dashes) in the time column
- [ ] [auto] An entry whose message contains embedded newlines renders as exactly one terminal line (newlines flattened to spaces)
- [ ] [auto] The cursor row is rendered with the theme's cursor highlight background color applied to the full row width
- [ ] [human] The cursor row is clearly distinguishable from non-selected rows
**Dependencies:** cavekit-log-source (entry model), cavekit-config (field visibility, logger depth, theme)

### R2: Logger Abbreviation
**Description:** Logger names are abbreviated by showing the last N segments at full length and abbreviating earlier segments to their first character. The segment depth is configurable.
**Acceptance Criteria:**
- [ ] [auto] With depth 2, `org.springframework.data.repository.RepositoryDelegate` abbreviates to `o.s.d.repository.RepositoryDelegate` (last 2 segments kept full, all prior segments abbreviated to their first character)
- [ ] [auto] With depth 2, `com.example.server.AppServerKt` abbreviates to `s.AppServerKt`
- [ ] [auto] With depth 1, `com.example.server.AppServerKt` abbreviates to `c.e.s.AppServerKt`
- [ ] [auto] A logger with fewer segments than the configured depth is shown unabbreviated
**Dependencies:** cavekit-config (logger abbreviation depth)

### R3: Level Badge Colors
**Description:** Level badges are styled using colors from the active theme. ERROR uses the theme's error color, WARN uses warning color, INFO uses default color, DEBUG uses dim color. No color values are hardcoded — all resolved from the active theme's tokens.
**Acceptance Criteria:**
- [ ] [auto] Rendering an ERROR entry with the default theme produces ANSI output containing the default theme's error color token value
- [ ] [auto] Rendering a WARN entry with the default theme produces ANSI output containing the default theme's warning color token value
- [ ] [auto] Rendering an INFO entry with the default theme produces ANSI output containing the default theme's info color token value
- [ ] [auto] Rendering a DEBUG entry with the default theme produces ANSI output containing the default theme's dim color token value
- [ ] [auto] Switching the active theme changes the ANSI color codes in the rendered output to match the new theme's tokens
- [ ] [human] One-time visual sign-off per bundled theme: level badge colors are perceptually correct and readable
**Dependencies:** cavekit-config (theme)

### R4: Two-Level Cursor Navigation
**Description:** Navigation uses a magit-style two-level model. At event level, `j`/`k` move between entries. Entering sub-level with `l`/right/Tab reveals per-field sub-rows for the current entry. `h`/left/Esc returns to event level. Sub-rows show one configured field per row, indented.
**Acceptance Criteria:**
- [ ] [auto] Pressing `j` moves the cursor to the next entry
- [ ] [auto] Pressing `k` moves the cursor to the previous entry
- [ ] [auto] `j`/`k` never land the cursor on a sub-row
- [ ] [auto] Pressing `l`, right arrow, or Tab on an entry enters sub-row level, displaying sub-rows for the configured fields
- [ ] [auto] Sub-rows are displayed indented beneath their parent entry
- [ ] [auto] Each sub-row shows one field name and its value
- [ ] [auto] Pressing `h`, left arrow, or Esc while in sub-row level returns to event level
- [ ] [auto] At event level, entries with sub-row fields show a visual boundary whether expanded or collapsed
- [ ] [human] Event boundaries are visually clear and readable
**Dependencies:** cavekit-config (sub-row fields)

### R5: Scroll Navigation
**Description:** `g` jumps to the first entry, `G` jumps to the last entry, `Ctrl-d` scrolls half a page down, `Ctrl-u` scrolls half a page up.
**Acceptance Criteria:**
- [ ] [auto] Pressing `g` moves the cursor to the first entry and scrolls to top
- [ ] [auto] Pressing `G` moves the cursor to the last entry and scrolls to bottom
- [ ] [auto] Pressing `Ctrl-d` scrolls approximately half the visible height downward
- [ ] [auto] Pressing `Ctrl-u` scrolls approximately half the visible height upward
**Dependencies:** none

### R6: Virtual Rendering
**Description:** The list View output contains exactly ViewportHeight rows — no more, no less — regardless of total entry count. Shortfalls are padded with empty lines. This ensures the list occupies exactly its allocated layout slot and remains responsive with large files.
**Acceptance Criteria:**
- [ ] [auto] With 100,000 entries loaded, the number of rendered rows equals the viewport height
- [ ] [auto] Scrolling through a large dataset does not degrade in responsiveness (render time per frame stays below 16ms)
**Dependencies:** none

### R7: Filtered View
**Description:** The list displays only entries that pass the active filter set. When filters change, the list updates to reflect the new filtered set while preserving cursor position where possible.
**Acceptance Criteria:**
- [ ] [auto] When a filter excludes an entry, that entry does not appear in the list
- [ ] [auto] When filters change, the list updates to show only passing entries
- [ ] [auto] If the previously selected entry still passes filters, it remains selected after filter change
- [ ] [auto] If the previously selected entry is filtered out, the cursor moves to the nearest passing entry
**Dependencies:** cavekit-filter-engine (filtered entry index)

### R8: Level-Jump Navigation
**Description:** `e`/`E` navigate to the next/previous ERROR entry. `w`/`W` navigate to the next/previous WARN entry. Search spans the full entry set regardless of active filters, and wraps around with an indicator. If the target entry is currently excluded by filters it is still navigated to and shown, with a visual indicator that it would be hidden under the current filter set.
**Acceptance Criteria:**
- [ ] [auto] Pressing `e` moves the cursor to the next entry with level ERROR in the full entry set
- [ ] [auto] Pressing `E` moves the cursor to the previous entry with level ERROR in the full entry set
- [ ] [auto] Pressing `w` moves the cursor to the next entry with level WARN in the full entry set
- [ ] [auto] Pressing `W` moves the cursor to the previous entry with level WARN in the full entry set
- [ ] [auto] When no more matching entries exist in the search direction, the search wraps to the other end
- [ ] [auto] When a wrap occurs, an indicator is shown
- [ ] [auto] When level-jump lands on an entry that is excluded by active filters, the entry is shown and a visual indicator communicates it is outside the current filter
- [ ] [human] The "filtered-out but visible" indicator is clearly distinguishable from normal entries
**Dependencies:** cavekit-log-source (entry level field), cavekit-filter-engine (filtered entry index)

### R9: Marks and Bookmarks
**Description:** `m` toggles a mark on the current entry. Marked entries show a visual indicator in the list. `u`/`U` navigate to the next/previous marked entry.
**Acceptance Criteria:**
- [ ] [auto] Pressing `m` on an unmarked entry marks it; pressing `m` again unmarks it
- [ ] [auto] Marked entries display a visual indicator in their row
- [ ] [auto] Pressing `u` moves the cursor to the next marked entry
- [ ] [auto] Pressing `U` moves the cursor to the previous marked entry
- [ ] [auto] Mark navigation wraps with an indicator when reaching the end/beginning
**Dependencies:** none

### R10: Entry Selection and Mouse
**Description:** The currently selected entry emits a signal consumed by the detail pane. Mouse click selects the entry at the clicked row — the selection MUST land on the row the user visually clicked, not an offset row. Mouse scroll wheel scrolls the list per the R12 scrolloff-drag semantics. Double-click on a row opens the detail pane for **that** row. Click-row resolution must account for the header bar height, the list's own top border (when rendered), and any active pane-layout dividers: the mapping from terminal Y coordinate to visible list row index has a **single owner** (the layout math, cavekit-app-shell R2) and must NOT be duplicated or re-derived inside the list's mouse handler. A failure to subtract the header height and list top border from the incoming mouse Y coordinate produces an offset-by-N bug where clicking row N highlights row N+2 — specifically disallowed by the ACs below.
**Acceptance Criteria:**
- [ ] [auto] Clicking on an entry row with the mouse selects **that** entry (the visually clicked row, not an offset row)
- [ ] [auto] For a given terminal Y coordinate, the list's click-to-row resolver returns a row index derived by subtracting exactly the header height (1) plus the list's top border (0 or 1; the exact value is owned by cavekit-app-shell R2 layout math and must be read from there — the list's mouse handler must NOT re-derive it) from the terminal Y — verified by unit test across all combinations of orientation × focus state (no double-count, no missing subtract)
- [ ] [auto] When the terminal Y coordinate falls outside the list's content area (on the header bar, status bar, the pane divider, or inside the detail pane), the click is NOT routed to the list at all — partitioning is owned by cavekit-app-shell R6
- [ ] [auto] Double-click uses the same click-to-row resolver as single-click — double-clicking the Nth visible row opens the detail pane for the Nth visible row, not an offset row
- [ ] [auto] Mouse scroll wheel scrolls the list (unchanged)
- [ ] [auto] The list's click-to-row resolver MUST reject every terminal Y with no row when the layout-owned content-top-Y offset has not been injected (e.g. on a freshly constructed list model that has not received a `WindowSizeMsg`/`relayout` cycle). Defaulting to zero MUST NOT silently reintroduce the 2-row-offset bug — the list either requires the offset at construction (panic on unset) or tracks an explicit "wired" flag and returns "no row" until it is set. Verified by a unit test that constructs a list WITHOUT wiring the offset and asserts a y=0 click does NOT select row 0
- [ ] [human, tui-mcp] Verify via tui-mcp first: launch the TUI on `logs/small.log`, take a `snapshot` to record the coordinates of visible rows 1, 2, 5, and the last row; use `send_mouse` to click each coordinate; after each click take another `snapshot` and confirm the cursor highlight lies on the same row that was clicked (no 1- or 2-row offset). Repeat with the detail pane open in below-mode and again in right-mode (per the user's HUMAN-sign-off-via-tui-mcp preference)
**Dependencies:** cavekit-detail-pane (receives selection signal), cavekit-app-shell (R2 layout math — single owner of terminal-Y-to-row mapping, R6 mouse routing partitioning)

### R11: Cursor Position Indicator
**Description:** The current cursor position within the entry list is communicated to the app shell so it can be displayed in the header or status bar. The position includes the 1-based index of the cursor entry within the visible (filtered) set.
**Acceptance Criteria:**
- [ ] [auto] The list model exposes the current cursor position as a 1-based index within the visible entry set
- [ ] [auto] The cursor position updates when the cursor moves (j/k, g/G, level-jump, mark-nav)
- [ ] [auto] When filters change, the cursor position reflects the new filtered set
**Dependencies:** cavekit-app-shell (displays the position)

### R12: Scrolloff Context Rows
**Description:** The entry list is a cursor-tracking viewport and honours the shared top-level `scrolloff` config (default 5) — the same key consumed by the detail pane (see cavekit-detail-pane.md R11 and DESIGN.md §4.3 "Shared scrolloff"). All vertical navigation (`j`/`k`, level jumps `e`/`w`, mark jumps, `g`/`G`, PgDn/PgUp/Ctrl+d/Ctrl+u) scrolls the viewport to keep the cursor at least `scrolloff` rows away from the nearest edge wherever the filtered-entry count permits. Mouse wheel scrolls the viewport; if the cursor would leave the visible window minus `scrolloff` rows from the nearest edge, the cursor is **dragged along** so it stays on the `scrolloff`-th row from that edge (nvim-style scrolloff drag). When the visible set is shorter than `2 × scrolloff + 1` rows, the margin is proportionally reduced (clamped to `floor(VisibleRows / 2)`).
**Acceptance Criteria:**
- [ ] [auto] When the cursor is far from both edges and the user presses `j`/`k`, the viewport does not scroll until the cursor comes within `scrolloff` rows of the top or bottom edge
- [ ] [auto] Once the cursor is within `scrolloff` rows of the bottom edge, pressing `j` scrolls the viewport by one and keeps the cursor exactly `scrolloff` rows from the bottom
- [ ] [auto] Once the cursor is within `scrolloff` rows of the top edge, pressing `k` scrolls the viewport by one and keeps the cursor exactly `scrolloff` rows from the top
- [ ] [auto] At document boundaries (cursor at first or last row) scrolloff yields — the cursor is allowed to sit on the very first / very last visible row so no rows are ever hidden below `scrolloff`-rows of blank space
- [ ] [auto] Mouse wheel down scrolls the viewport; when the cursor is more than `scrolloff` rows from both edges it stays on its current document line; when the cursor would leave the `scrolloff` margin it is dragged along to stay on the `scrolloff`-th row from the nearest edge
- [ ] [auto] `g`/`G`/PgDn/PgUp/Ctrl+d/Ctrl+u + level-jump + mark-jump all honour scrolloff — after they move the cursor, the viewport is adjusted so the cursor sits with `scrolloff` rows of context above and below where possible
- [ ] [auto] `scrolloff` value is read from the shared top-level config key `scrolloff` (NOT a list-specific key); at use time it is clamped to `[0, floor(VisibleContentRows / 2)]`
- [ ] [human] On `logs/small.log`, pressing `j` repeatedly with a moderate-sized filtered set shows the cursor staying 5 rows above the bottom border before the viewport starts scrolling; mouse-wheel scrolling in the middle of the list keeps the cursor on its document line while rows move under it; wheel near an edge drags the cursor to stay in the scrolloff margin
**Dependencies:** cavekit-config (shared top-level `scrolloff`), cavekit-detail-pane (shares the same config key — R11)

### R13: Free-Text List Search
**Description:** The entry list supports an in-list free-text search, scoped to the list pane, distinct from the filter engine and distinct from detail-pane search (R7). Activated with `/` when the list is focused and the detail pane is NOT open-and-focused (routing legislated by cavekit-app-shell R13). Typed query matches case-insensitively against the rendered compact-row text (time, level, logger, message) of each **visible (filtered)** entry. While search is active, matching entries are highlighted with the theme's `SearchHighlight` bg on their row; the search prompt + query + `(current/total)` match counter renders in the key-hint bar or an in-list status line. `n` / `N` cycles the cursor to the next / previous match and honours R12 scrolloff. Esc dismisses search and clears highlights. Search does NOT change the filter set — it is a navigation aid, not a filter; entries that don't match remain visible (just not highlighted). Search state is local to the list and must not leak into the detail pane or filter engine.
**Acceptance Criteria:**
- [ ] [auto] Pressing `/` with the entry list focused and no detail pane open (or detail pane open-but-not-focused) opens a search input scoped to the list
- [ ] [auto] Typing a query matches case-insensitively against the compact-row text (time, level, logger, message) of visible entries — the visible set is unchanged, matching entries simply get `SearchHighlight` bg applied
- [ ] [auto] The active query and `(current/total)` match counter are rendered visibly while search is active; when `query != "" && matches == 0` a "No matches" indicator is shown
- [ ] [auto] Pressing `n` moves the cursor to the next matching entry (wrapping with an indicator at the end); `N` moves to the previous (wrapping at the start)
- [ ] [auto] Cursor movement via `n`/`N` honours R12 scrolloff — the cursor lands with `scrolloff` rows of context above and below where the filtered-entry set permits
- [ ] [auto] Pressing Esc dismisses the search input, clears the `SearchHighlight` bg on all rows, and leaves the cursor at its current position
- [ ] [auto] Search state is cleared automatically on every focus-loss trigger from the list — Tab cycle, `f` (transfer to filter panel), mouse click in any non-list zone — AND when filters change (a filter change invalidates the match set)
- [ ] [auto] When entries arrive during active search (via `AppendEntries` on batch load or tail), newly arriving entries that match the query join the match set and render with `SearchHighlight` bg; `(current/total)` reflects the live total without requiring the user to re-type the query
- [ ] [auto] List search does NOT modify the filter engine — entries that do not match the query remain visible (non-highlighted) in the list
- [ ] [auto] Backspace on a query containing multi-byte runes (emoji / CJK) removes exactly one rune without corrupting UTF-8
- [ ] [auto] The cursor-row bg (R1, R12) takes visual priority over the `SearchHighlight` bg on the same row — other matching rows keep their `SearchHighlight` styling normally
- [ ] [human] On `logs/small.log`, pressing `/` then typing a substring of a known log message visibly highlights the matching rows, shows `(cur/total)`, and `n`/`N` scrolls the cursor onto each match with scrolloff context preserved
**Dependencies:** cavekit-config (`SearchHighlight` theme token), cavekit-app-shell (R13 cross-pane `/` routing — routes `/` to list when list focused and detail not focused), R12 (scrolloff-respecting cursor move)

### R14: Tail-Follow on Append
**Description:** When `AppendEntries` delivers new entries, tail-follow semantics apply: if the cursor was on the last entry before the append, the cursor advances to the new last entry and the viewport scrolls so the new last entry is visible with R12 scrolloff at the bottom edge. If the cursor was NOT at the last entry, neither cursor nor viewport moves — the user's navigation context is preserved. Applies to both background load (`EntryBatchMsg`) and tail mode (`TailMsg`). Mirrors `less +F`. While tail mode is active AND the cursor is on the last entry, a `FOLLOW` indicator is rendered in the header; any upward cursor move (`k`, wheel up, `Ctrl-u`, `g`, level-prev, mark-prev, search-prev) clears it, and the user re-engages follow by pressing `G` to land back on the last entry.
**Acceptance Criteria:**
- [ ] [auto] When AppendEntries is called and Cursor == TotalEntries-1 before the call, afterwards Cursor == new TotalEntries-1
- [ ] [auto] When AppendEntries is called and Cursor == TotalEntries-1 before the call, afterwards the viewport offset is adjusted so the new last entry is visible with R12 scrolloff at the bottom edge
- [ ] [auto] When AppendEntries is called and Cursor < TotalEntries-1 before the call, Cursor is unchanged afterwards
- [ ] [auto] When AppendEntries is called and Cursor < TotalEntries-1 before the call, the viewport Offset is unchanged afterwards
- [ ] [auto] On an empty list (TotalEntries == 0), the first append leaves Cursor at 0 and Offset at 0
- [ ] [auto] `IsAtTail()` reports true iff TotalEntries > 0 and Cursor == TotalEntries-1; the app wires `header.WithFollow(followMode && IsAtTail())` at every render so the `[FOLLOW]` badge tracks cursor state in real time
- [ ] [auto] When tail-follow snaps the cursor on AppendEntries (pre-append Cursor == TotalEntries-1), the entry-selection signal (R10) is delivered for the new last entry — i.e. any consumer that re-renders on selection (detail pane R1 live-preview) shows the newly appended entry without requiring a subsequent keypress. The signal fires exactly once per snap, and only when the cursor actually moved (no spurious re-renders when Cursor < TotalEntries-1 and the cursor is unchanged). Applies symmetrically to tail mode (`TailMsg`) and background load (`EntryBatchMsg`).
- [ ] [human, tui-mcp] On `logs/small.log` with `gloggy -f`: open the detail pane on the last entry (Enter), append a line externally — verify via `snapshot` that the detail pane content updates to the appended entry (not the previous one) in the same frame that the cursor advances. Repeat in both below- and right-orientations.
- [ ] [human, tui-mcp] On a file under `logs/` with `gloggy -f`: initial tail-mode startup lands cursor on the last entry with `[FOLLOW]` visible in the header. Append lines externally and verify via `snapshot` that the viewport advances and new entries are visible without any keypress. Press `k` to leave the tail — `[FOLLOW]` disappears. Append again and verify the viewport does NOT advance (but header counts update). Press `G` — `[FOLLOW]` returns and subsequent appends follow again
**Dependencies:** R12 (scrolloff — viewport adjustment uses the same `followCursor` path), cavekit-log-source R8 (tail emission), cavekit-app-shell R3 (header hosts the `[FOLLOW]` badge)

## Out of Scope

- Detail pane rendering (handled by detail-pane)
- Filter logic and filter panel (handled by filter-engine)
- Deciding which fields are sub-rows (handled by config)
- Clipboard operations (handled by app-shell)

## Cross-References

- See also: cavekit-log-source.md (provides entry data)
- See also: cavekit-detail-pane.md (receives entry selection)
- See also: cavekit-filter-engine.md (provides filtered entry index)
- See also: cavekit-config.md (field visibility, sub-row fields, logger depth, theme)
- See also: cavekit-app-shell.md (layout, mouse routing)

## Changelog

### 2026-04-20 — Revision (R14 selection signal on tail-follow snap)
- **Affected:** R14 (two new ACs)
- **Summary:** R14 specified cursor + viewport snap on AppendEntries but omitted the selection signal. Consequence: with the detail pane open and cursor on the last entry, an incoming tail/batch append advanced the cursor but the pane kept rendering the previous entry until the user pressed a key — silently violating detail-pane R1's live-preview invariant ("With the pane open and the list focused, pressing j/k moves the list cursor and the detail pane re-renders with the new selection"). New ACs tie the R10 selection signal to every tail-follow cursor snap (fire exactly once, only when cursor moved, symmetric across `TailMsg` and `EntryBatchMsg`), plus a tui-mcp sign-off that the pane content tracks the newly appended entry in the same frame.
- **Driven by:** User report after branch `fix-r14-tail-follow-viewport` landed the cursor+viewport snap: "when on last line (and thus in follow mode), if the details pane is also open, when new line comes in, the cursor is moved onto that, but details pane content is not updated". Backprop-log entry #8.

### 2026-04-20 — Revision (R14 tail-follow on append)
- **Affected:** new R14
- **Summary:** Added R14 to legislate `less +F`-style tail-follow behavior in the entry list. `AppendEntries` was previously stateless with respect to cursor/viewport — TotalEntries bumped, but Cursor and Offset never moved. With the cursor at the last entry in tail mode, incoming entries were invisible (the viewport stayed put) until the user nudged the cursor. R14 rules: follow iff pre-append Cursor == TotalEntries-1; any upward nav disengages follow; `G` re-engages it. Reuses R12 `followCursor` for bottom-edge scrolloff. Header `[FOLLOW]` badge wired to `followMode && IsAtTail()` at View time so it tracks cursor state live, giving users a visible indicator of when the list is auto-advancing and when it has paused. Verified end-to-end via tui-mcp: initial startup, paused-by-`k`, and resumed-by-`G` all exercised with external appends.
- **Driven by:** User report after the Tier 25 tail-emission fix: "with big enough log file that fills the screen and I navigate to the last line, when new lines are appended it is not shown any other way than the line counter on the top bar, then when I move the cursor the newly appended lines start appearing". Backprop-log entry #7. Paired with test + fix commits on branch `fix-r14-tail-follow-viewport`.

### 2026-04-19 — Revision (R10 unset-contentTopY safety)
- **Affected:** R10 (one new AC)
- **Summary:** Tier 19 `/ck:check` flagged F-127: the T-158 single-owner click-row resolver stores its content-top-Y offset in `ListModel.contentTopY int` with a zero-value default. If a future refactor drops the `WithContentTopY(l.ListContentTopY())` wire in `app/model.go:177,666`, a y=0 header click silently maps back to row 0 — the exact 2-row-offset bug T-158 closed reappears with no test to catch it. New AC requires the list to reject clicks when the offset has not been injected (panic on unset OR explicit "wired" flag). Adds defensive lower bound to the single-owner contract.
- **Driven by:** `/ck:check` run 2026-04-19 on Tier 19, finding F-127. Paired task T-163.

### 2026-04-18 — Revision (R10 click-row alignment bug)
- **Affected:** R10
- **Summary:** R10 revised to tighten the mouse-click-to-row mapping after a user-observed 2-row offset (clicking row 1 highlights row 3, clicking row 2 highlights row 4). Root cause: header height (1) plus list top border (1) not subtracted from the incoming mouse Y coordinate. The prior AC "Clicking on an entry row with the mouse selects that entry" passed a shallow test because any selection occurred, not because the clicked row was the selected row. Added a unit-level AC that pins the resolver math to a single-owner (cavekit-app-shell R2 layout), an AC forbidding duplicate resolution inside the list handler, and a `[human, tui-mcp]` AC that drives `send_mouse` clicks at known row coordinates and verifies via `snapshot` that the cursor highlight lands on the same row (no offset) in both orientations.
- **Driven by:** User feedback in /ck:sketch session: "there's something off with mouse row position calculation, when I click row 1, row 3 is highlighted, when on row 2, row 4 is highlighted, etc. it seems two rows off". Paired with the cavekit-app-shell R12/R15 resize revision in the same session.

### 2026-04-16 — Revision
- **Affected:** R1, new R11
- **Summary:** R1 updated to require cursor row highlighting with theme background color. New R11 added for cursor position indicator (1-based index in visible set) to be displayed in header/status bar. Both driven by user observation that cursor row has no visual indication and no position info is shown.
- **Commits:** manual testing feedback (no commit)

### 2026-04-18 — Revision (R12 scrolloff)
- **Affected:** new R12
- **Summary:** Added R12 to legislate nvim-style `scrolloff` on the entry list — the viewport keeps a configurable number of context rows (default 5, shared top-level `scrolloff` config) between cursor and edges. All cursor-moving navigation (j/k, g/G, PgDn/PgUp, level-jump, mark-jump) adjusts the viewport to honour the margin; mouse wheel drags the cursor along when the cursor would leave the margin. Same config key used by cavekit-detail-pane.md R11 — users tune one number for both panes. Companion to DESIGN.md §4.3 "Shared scrolloff" subsection.
- **Driven by:** `/ck:check` run 2026-04-18 after user report "I still see no row highlight where cursor is when focused on details pane" + follow-up "scrolloff should be implemented on the list as well". Addresses finding F-026 (list portion).

### 2026-04-18 — Revision (R13 focus-loss triggers + streaming matches)
- **Affected:** R13 (AC expansion)
- **Summary:** R13 AC 7 broadened from "Tab cycle OR filter change" to enumerate every focus-loss trigger that must clear search: Tab cycle, `f` (transfer to filter panel), mouse click in non-list zone. New AC added for streaming entries: when `AppendEntries` delivers new entries while search is active, matching new entries join the match set and `(current/total)` updates live. Discovered during `/ck:check` after Tier 15/16 — `f` transfer left search rendered while filter panel focused (F-108); tail/batch entries silently skipped by `n`/`N` (F-109).
- **Driven by:** `/ck:check` 2026-04-18 (post-Tier 15/16), findings F-108 (P2) + F-109 (P2). Paired tasks T-147 + T-148.

### 2026-04-18 — Revision (R13 free-text list search)
- **Affected:** new R13
- **Summary:** Added R13 for free-text search scoped to the entry list. `/` when list is focused and detail pane is not open-and-focused opens a query input; matching rows get `SearchHighlight` bg; `n`/`N` cycles the cursor across matches honouring R12 scrolloff. Distinct from filter engine — search does NOT change which entries are visible, only highlights matches and moves the cursor. Distinct from detail-pane search (cavekit-detail-pane.md R7) — cross-pane `/` routing disambiguated by cavekit-app-shell.md R13.
- **Driven by:** `/ck:check` run 2026-04-18 after user note "list: no search". Addresses finding F-101.

### 2026-04-16 — Revision (layout fixes)
- **Affected:** R1, R6
- **Summary:** R1: added requirement that each compact row must be exactly one terminal line — messages with embedded newlines must be flattened to spaces. Without this, multi-line messages overflow the viewport and corrupt the layout. R6: changed from "visible rows plus a small buffer" to "exactly ViewportHeight rows". The buffer concept is incompatible with Bubble Tea's compositing model where View() output is stacked via JoinVertical — extra rows overflow into adjacent layout zones.
- **Commits:** uncommitted (session fixes)
