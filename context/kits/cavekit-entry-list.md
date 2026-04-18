---
created: "2026-04-15T00:00:00Z"
last_edited: "2026-04-18T14:20:33+03:00"
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
**Description:** The currently selected entry emits a signal consumed by the detail pane. Mouse click selects an entry, scroll wheel scrolls the list, double-click opens the detail pane.
**Acceptance Criteria:**
- [ ] [auto] When the cursor moves to a new entry, a selection signal is emitted with that entry's data
- [ ] [auto] Clicking on an entry row with the mouse selects that entry
- [ ] [auto] Mouse scroll wheel scrolls the list
- [ ] [auto] Double-clicking an entry opens the detail pane for that entry
**Dependencies:** cavekit-detail-pane (receives selection signal), cavekit-app-shell (mouse routing)

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

### 2026-04-16 — Revision
- **Affected:** R1, new R11
- **Summary:** R1 updated to require cursor row highlighting with theme background color. New R11 added for cursor position indicator (1-based index in visible set) to be displayed in header/status bar. Both driven by user observation that cursor row has no visual indication and no position info is shown.
- **Commits:** manual testing feedback (no commit)

### 2026-04-18 — Revision (R12 scrolloff)
- **Affected:** new R12
- **Summary:** Added R12 to legislate nvim-style `scrolloff` on the entry list — the viewport keeps a configurable number of context rows (default 5, shared top-level `scrolloff` config) between cursor and edges. All cursor-moving navigation (j/k, g/G, PgDn/PgUp, level-jump, mark-jump) adjusts the viewport to honour the margin; mouse wheel drags the cursor along when the cursor would leave the margin. Same config key used by cavekit-detail-pane.md R11 — users tune one number for both panes. Companion to DESIGN.md §4.3 "Shared scrolloff" subsection.
- **Driven by:** `/ck:check` run 2026-04-18 after user report "I still see no row highlight where cursor is when focused on details pane" + follow-up "scrolloff should be implemented on the list as well". Addresses finding F-026 (list portion).

### 2026-04-16 — Revision (layout fixes)
- **Affected:** R1, R6
- **Summary:** R1: added requirement that each compact row must be exactly one terminal line — messages with embedded newlines must be flattened to spaces. Without this, multi-line messages overflow the viewport and corrupt the layout. R6: changed from "visible rows plus a small buffer" to "exactly ViewportHeight rows". The buffer concept is incompatible with Bubble Tea's compositing model where View() output is stacked via JoinVertical — extra rows overflow into adjacent layout zones.
- **Commits:** uncommitted (session fixes)
