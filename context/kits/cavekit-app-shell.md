---
created: "2026-04-15T00:00:00Z"
last_edited: "2026-04-18T22:44:21+03:00"
---

# Cavekit: App Shell

## Scope

The top-level application entry point, layout management, domain wiring, mouse routing, clipboard, help overlay, and the context-sensitive key-hint bar. Owns the terminal UI lifecycle but delegates all domain-specific logic to the respective domain kits.

## Requirements

### R1: Entry Points
**Description:** The application supports three invocation modes: a file path argument for reading a file, a tail flag with a file path for follow mode, and no arguments for reading from stdin.
**Acceptance Criteria:**
- [ ] [auto] `gloggy <file>` starts the application reading from the specified file
- [ ] [auto] `gloggy -f <file>` starts the application in tail mode on the specified file
- [ ] [auto] `gloggy` with piped stdin starts the application reading from stdin
- [ ] [auto] Invalid arguments (e.g. both stdin pipe and file, or `-f` without file) produce a clear error message
**Dependencies:** cavekit-log-source (input handling)

### R2: Layout
**Description:** The terminal is divided into: a header bar (top), the main pane area (entry list plus optional detail pane), and a status/key-hint bar (bottom-most). The layout fills the full terminal width and height. The detail pane's placement relative to the entry list is governed by three orientation modes: `below` (detail pane stacked beneath the list), `right` (detail pane side-by-side with the list, separated by a 1-cell divider), and `auto` (flips between the two at the configured threshold from cavekit-config R5). In `right`-split composition the main area renders as `[entryList │ divider(1 cell) │ detailPane]` between the header and the status bar.
**Acceptance Criteria:**
- [ ] [auto] The header bar is rendered at the top of the terminal
- [ ] [auto] The entry list occupies the main area between the header and the bottom bars
- [ ] [auto] When the detail pane is open, it appears between the entry list and the status bar
- [ ] [auto] The status/key-hint bar is rendered at the bottom of the terminal
- [ ] [auto] All panes together fill the full terminal width and height with no gaps or overlap
- [ ] [auto] In right-split orientation, the main area composes as header / [entryList │ divider(1 cell) │ detailPane] / statusBar with the divider occupying exactly 1 terminal column
- [ ] [auto] In right-split orientation, pane widths are computed after subtracting both pane borders and the 1-cell divider from the usable terminal width (per DESIGN.md §5 border accounting)
- [ ] [auto] When terminal width is below 60 columns or terminal height is below 15 rows, normal rendering is suppressed and a centered "terminal too small" message is shown instead (per DESIGN.md §8)
**Dependencies:** cavekit-entry-list, cavekit-detail-pane, cavekit-config (orientation settings)

### R3: Header Bar
**Description:** The header bar displays: the file name (or "stdin" indicator), a tail/follow indicator when in tail mode, counts showing total entries and currently visible (filtered) entries, and the current cursor position (e.g. "42/110"). The header bar must be visually distinct from the entry list below it — rendered with a background color or inverse styling from the active theme so it does not blend into log lines.
**Acceptance Criteria:**
- [ ] [auto] The header bar shows the file name when reading from a file
- [ ] [auto] The header bar shows a stdin indicator when reading from stdin
- [ ] [auto] The header bar shows a `[FOLLOW]` badge when tail mode is active
- [ ] [auto] The header bar shows the total entry count
- [ ] [auto] The header bar shows the visible (filtered) entry count
- [ ] [auto] Counts update as new entries are loaded or filters change
- [ ] [auto] The header bar shows the current cursor position as a 1-based index within the visible set (e.g. "42/110")
- [ ] [auto] The header bar is rendered with a distinct background color from the theme (not plain unstyled text)
- [ ] [human] The header bar is clearly distinguishable from the entry list rows below it
- [ ] [auto] When the header's rendered width would exceed the terminal width, content is dropped in this order: focus label, then entry counts, then cursor position, then FOLLOW badge (per DESIGN.md §4.1 and §8)
- [ ] [auto] The source name is always visible in the header; when it alone would overflow it is truncated with an ellipsis rather than dropped
**Dependencies:** cavekit-log-source (file name, tail status, entry count), cavekit-filter-engine (filtered count), cavekit-entry-list (cursor position)

### R4: Context-Sensitive Key-Hint Bar
**Description:** The bottom status bar shows relevant keybindings for the currently focused component. The hints update as focus changes between components (entry list, detail pane, filter panel, help overlay). The key-hint bar must occupy exactly 1 row — content that exceeds the terminal width is truncated (never wrapped), since the layout reserves StatusBarHeight=1. When more than one pane is visible the bar's right side also shows an active-pane label (`focus: list | details | filter`) using Bold weight and the FocusBorder foreground (per DESIGN.md §3 type roles and §4.6).
**Acceptance Criteria:**
- [ ] [auto] When the entry list is focused, the key-hint bar shows entry-list keybindings (e.g. j/k, e/w, m, Enter)
- [ ] [auto] When the detail pane is focused, the key-hint bar shows detail-pane keybindings (e.g. j/k, /, +/-, Esc)
- [ ] [auto] When the filter panel is focused, the key-hint bar shows filter-panel keybindings (e.g. j/k, Space, d, Esc)
- [ ] [auto] Key hints update immediately when focus changes
- [ ] [auto] The key-hint bar renders as exactly 1 terminal row regardless of content length (truncated, not wrapped)
- [ ] [auto] When more than one pane is visible, the right side of the key-hint bar shows a focus label reading `focus: list`, `focus: details`, or `focus: filter` rendered in Bold with FocusBorder foreground
- [ ] [auto] When only one pane is visible, the focus label is omitted from the key-hint bar
**Dependencies:** cavekit-entry-list, cavekit-detail-pane, cavekit-filter-engine

### R5: Help Overlay
**Description:** Pressing `?` opens a full-screen overlay listing all keybindings across all domains. Esc closes the overlay.
**Acceptance Criteria:**
- [ ] [auto] Pressing `?` opens the help overlay
- [ ] [auto] The help overlay lists keybindings for all domains (entry list, detail pane, filter engine, app shell)
- [ ] [auto] Pressing Esc closes the help overlay and returns to the previous view
- [ ] [auto] While the help overlay is open, other keybindings are not processed
**Dependencies:** none

### R6: Mouse Mode and Routing
**Description:** Mouse mode is enabled globally. Mouse events are routed to the correct domain component based on the click/scroll position within the layout. In right-split orientation the main area is partitioned into horizontal zones (entry list, divider column, detail pane) rather than vertical zones, and mouse clicks inside a pane transfer focus to that pane. Mouse-button-1 press on the divider cell initiates a drag; drag semantics are defined in R15.
**Acceptance Criteria:**
- [ ] [auto] Mouse events in the entry list area are routed to the entry list
- [ ] [auto] Mouse events in the detail pane area are routed to the detail pane
- [ ] [auto] Mouse events do not cause crashes regardless of where in the terminal they occur
- [ ] [auto] In right-split orientation, mouse zones partition the main area horizontally into list / divider-column / detail, with the header and status bar remaining as separate horizontal zones
- [ ] [auto] A 1-cell buffer adjacent to the divider prevents clicks near the divider from being routed to the wrong pane
- [ ] [auto] Clicking inside a pane transfers focus to that pane
- [ ] [auto] Clicks landing on the divider cell itself are not routed to either pane as focus-transfer clicks — the divider cell is reserved for R15 drag initiation; focus is unchanged by a click on the divider
**Dependencies:** cavekit-entry-list (mouse handling), cavekit-detail-pane (mouse handling)

### R7: Terminal Resize Handling
**Description:** When the terminal is resized, all panes reflow to fill the new dimensions. No crashes, no layout corruption, and pane proportions are maintained. WindowSizeMsg must be processed by all child models even when they have no data yet — the initial resize arrives before async file loading completes. When the detail pane orientation is set to `auto`, every terminal-resize event re-evaluates orientation against the configured threshold. Both the below-mode height ratio and the right-mode width ratio are preserved independently across orientation flips.
**Acceptance Criteria:**
- [ ] [auto] After a terminal resize, the layout fills the new terminal dimensions
- [ ] [auto] Pane proportions (e.g. detail pane height ratio) are preserved after resize
- [ ] [auto] No content is clipped or overlapping after resize
- [ ] [auto] Resize does not cause a crash or panic
- [ ] [auto] Child models (e.g. entry list) process WindowSizeMsg even when their data set is empty
- [ ] [auto] When detail_pane.position is "auto", orientation is re-evaluated on every terminal-resize event against orientation_threshold_cols
- [ ] [auto] height_ratio and width_ratio are both preserved across orientation flips — flipping from below to right does not overwrite one with the other, and flipping back restores the previous values
- [ ] [auto] When the detail pane's computed dimension falls below the minimum (width < 30 cells in right orientation, height < 3 rows in below orientation) the pane auto-closes and the status bar emits a one-time notice
**Dependencies:** cavekit-detail-pane (pane proportions), cavekit-entry-list, cavekit-config (orientation settings)

### R8: Loading Indicator
**Description:** While the log source is reading the file on startup, a loading indicator is displayed. It disappears when loading completes.
**Acceptance Criteria:**
- [ ] [auto] While entries are being loaded, a loading indicator is visible
- [ ] [auto] When loading completes, the loading indicator is no longer visible
- [ ] [auto] The loading indicator shows progress (e.g. number of entries loaded so far)
**Dependencies:** cavekit-log-source (progress signals)

### R9: Clipboard
**Description:** Pressing `y` copies all marked entries to the system clipboard as JSONL (one JSON object per line). Non-JSON marked entries are included as raw text lines. The user must receive visible feedback on every `y` press: a success notice with the count, an error notice if the clipboard write fails (e.g. missing `xclip`/`wl-copy`), or a "no marked entries" notice when the mark set is empty. `y` must NEVER be a silent action — swallowing the error (e.g. `//nolint:errcheck`) is a kit violation.
**Acceptance Criteria:**
- [ ] [auto] Pressing `y` with marked entries copies them to the system clipboard
- [ ] [auto] The clipboard content is JSONL: one entry per line in original order
- [ ] [auto] Non-JSON marked entries are included as raw text lines
- [ ] [auto] Pressing `y` with no marked entries does not modify the clipboard
- [ ] [auto] Successful copy of N ≥ 1 marked entries emits a transient status-bar notice (e.g. `copied N entries`) via `keyhints.WithNotice`, auto-dismissed within ≤ 3 seconds
- [ ] [auto] Clipboard-write error (e.g. `atotto/clipboard.WriteAll` returns non-nil on a headless system with no clipboard binary) surfaces a visible transient error notice — the error is NEVER swallowed
- [ ] [auto] Pressing `y` with zero marked entries emits a `no marked entries` notice (visible feedback, not a silent no-op)
- [ ] [auto] The implementation does NOT use `//nolint:errcheck` or otherwise discard the `(ClipboardCopiedMsg, error)` return value from `CopyMarkedEntries`; both are routed back into the Bubble Tea update loop as a `tea.Cmd` so notices can be emitted
- [ ] [human] On `logs/small.log`, marking two entries with `m` then pressing `y` shows a visible "copied 2 entries" notice in the status bar; pressing `y` with no marks shows "no marked entries"; if clipboard cannot be reached the notice reads the system error
**Dependencies:** cavekit-entry-list (marks), `keyhints.WithNotice` (R4 status-bar notice contract)

### R10: Pane Visual-State Matrix
**Description:** Every focusable pane renders in one of three visual states — focused, unfocused-but-visible, or alone — per the matrix in DESIGN.md §4 (authoritative). The focused pane uses FocusBorder borders and full-contrast foreground; an unfocused visible pane uses DividerColor borders, an UnfocusedBg background tint, and a foreground blend toward Dim; a pane that is the only visible pane uses the focused treatment. The visual state must not alter the pane's rendered dimensions (no post-render border wrapping that adds rows or columns).
**Acceptance Criteria:**
- [ ] [auto] When the detail pane is open and focused, it has a visual indicator distinguishing it from the unfocused entry list (e.g. highlighted border or title)
- [ ] [auto] The focus indicator does not change the rendered width or height of any pane
- [ ] [auto] The focus indicator updates immediately when focus changes between panes
- [ ] [human] The focused pane is clearly identifiable at a glance
- [ ] [auto] Unfocused visible panes render with DividerColor borders
- [ ] [auto] Unfocused visible panes render with an UnfocusedBg background
- [ ] [auto] When a pane is the only visible pane, it uses the focused treatment (FocusBorder borders, base background)
- [ ] [auto][cross-kit] The cursor row in the entry list is always rendered, even when the entry list is unfocused; intensity and bold vary with focus (full enforcement depends on cavekit-entry-list revision)
- [ ] [auto] The detail pane top border is visible in both below and right orientations
**Dependencies:** cavekit-entry-list, cavekit-detail-pane, cavekit-config (DividerColor, UnfocusedBg, FocusBorder tokens)

### R11: Focus Cycle and Dismissal
**Description:** `Tab` cycles focus among the visible panes. Opening a pane does NOT itself transfer focus — focus transfers occur only on explicit actions: Tab (this requirement), mouse click on a pane (R6), or the cross-pane `/` activation (R13). This keeps the detail pane usable as a live preview while the user keeps navigating the entry list. When the filter panel or help overlay is open, Tab-cycling is paused and the overlay holds focus. `Esc` is context-sensitive: first close any open overlay; otherwise if the detail pane is open, close it and return focus to the entry list; otherwise clear transient state on the focused pane (e.g. an active search). Esc on entry-list focus with the detail pane open also closes the pane (the list doesn't need to be Tab'd to the pane first to dismiss it). A mouse click on a pane focuses that pane. Tab never closes a pane — closing is always explicit via Esc or a domain-specific dismissal key.
**Acceptance Criteria:**
- [ ] [auto] Pressing Tab with the entry list and detail pane both visible cycles focus between them
- [ ] [auto] Tab is inert (does not cycle focus) while the filter panel or help overlay is open
- [ ] [auto] Opening the detail pane (Enter, double-click) does NOT transfer focus to the pane — focus remains on the entry list
- [ ] [auto] Focus transfers to a newly opened pane only when the user takes an explicit action: Tab, mouse click on the pane, or `/` (R13)
- [ ] [auto] Pressing Esc with an overlay open closes the overlay only
- [ ] [auto] Pressing Esc with no overlay open and the detail pane focused closes the detail pane and returns focus to the entry list
- [ ] [auto] Pressing Esc with no overlay open, the entry list focused, and the detail pane open, closes the detail pane (focus stays on the list)
- [ ] [auto] Pressing Esc with no overlay open, the entry list focused, and no detail pane open, clears transient state (e.g. active level-jump wrap indicator) when present; otherwise it is a no-op
- [ ] [auto] Clicking inside a pane focuses that pane
- [ ] [auto] Tab never closes a pane
**Dependencies:** cavekit-entry-list, cavekit-detail-pane, cavekit-filter-engine

### R12: Layout Resize Controls (Keyboard, Focus-Aware)
**Description:** A single keyboard resize keymap operates uniformly across orientations and is **focus-aware** — all four keys act on the currently focused pane. The active dimension is `height_ratio` in below-mode and `width_ratio` in right-mode (stored on the detail pane per cavekit-detail-pane R6). When the detail pane is focused, keys change the detail pane's ratio directly; when the entry list is focused, keys change the list's share of the screen (the complement of the detail pane's ratio). `|` toggles the focused pane's share between two presets: 0.30 and 0.50. `+` grows the focused pane by 0.05; `-` shrinks it by 0.05. `=` resets to the layout default (detail ratio = 0.30, list share = 0.70) regardless of which pane is focused — "reset" is a global return-to-baseline, not focus-directional. All ratio values are clamped so the detail pane's ratio stays in `[0.10, 0.80]`; at the clamp boundary, further motion in the same direction is a no-op (not a wrap). Ratio changes are written back to the config file immediately so they persist across sessions. When only one pane is visible (detail pane closed), all four keys are silent no-ops — there is no divider to move.
**Acceptance Criteria:**
- [ ] [auto] Pressing `|` with the detail pane focused toggles the detail pane's ratio between 0.30 and 0.50
- [ ] [auto] Pressing `|` with the entry list focused toggles the list share between 0.30 and 0.50 (detail ratio 0.70 ↔ 0.50)
- [ ] [auto] Pressing `+` with the detail pane focused increases the detail pane's ratio by 0.05
- [ ] [auto] Pressing `+` with the entry list focused decreases the detail pane's ratio by 0.05 (the list grows)
- [ ] [auto] Pressing `-` is the symmetric inverse of `+` at each focus
- [ ] [auto] Pressing `=` sets the detail pane's ratio to 0.30 (list share 0.70) regardless of which pane is focused
- [ ] [auto] The detail pane's ratio stays in `[0.10, 0.80]`; keys at the clamp boundary are no-ops, not wrap
- [ ] [auto] In below orientation the active dimension is `height_ratio`; in right orientation it is `width_ratio`
- [ ] [auto] Ratio changes are persisted to the config file via live write-back (cavekit-config R6)
- [ ] [auto] When only the entry list is visible (detail pane closed), `|`/`+`/`-`/`=` are silent no-ops — no ratio change, no error, no config write
- [ ] [human, tui-mcp] Verify via tui-mcp first: launch the TUI on `logs/small.log`, open the detail pane, focus the list, press `+` and verify the list region grows by the expected 1 row/col in the current orientation via `snapshot`; repeat with detail focused; verify `|` toggles between 0.30 ↔ 0.50 share; verify `=` resets regardless of focus; verify clamp pin at 0.10 and 0.80. Only fall back to direct human sign-off if `send_keys` cannot emit the required key (per the user's HUMAN-sign-off-via-tui-mcp preference)
**Dependencies:** cavekit-detail-pane (applies new ratio, R6), cavekit-config (live write-back, ratio settings), R11 (focus state)

### R13: Cross-Pane Search Activation
**Description:** `/` activates an in-pane search. The target pane is determined by **current focus**, not by pane-open state. If the entry list is focused, `/` activates list search (cavekit-entry-list R13). If the detail pane is focused, `/` activates detail-pane search (cavekit-detail-pane R7). If the filter panel is focused, `/` is routed to the filter input as a literal character (not intercepted). The activation must never be a silent no-op at any focus. The key-hint bar and help overlay must advertise `/` accurately per focus — different label when list-focused vs pane-focused.
**Acceptance Criteria:**
- [ ] [auto] With entry list focused (detail pane closed OR open-but-not-focused), pressing `/` activates list search (entry-list R13) — NOT detail-pane search
- [ ] [auto] With detail pane focused, pressing `/` activates detail-pane search (detail-pane R7) in the detail pane
- [ ] [auto] With filter panel focused, pressing `/` is routed to the filter input as a literal character (not intercepted as a search activation)
- [ ] [auto] Pressing `/` is never a silent no-op at any focus; every focus state either activates a search or passes the literal character to the focused input
- [ ] [auto] The key-hint bar shows `/` with accurate scope per focus: `/ search list` when list focused, `/ search pane` when detail pane focused, hidden when filter panel focused
- [ ] [auto] The help overlay entry for `/` states its focus-sensitive scope explicitly ("Search in focused pane — list or detail")
- [ ] [human] On `logs/small.log` with list focused (no pane open), pressing `/` opens list search (not a "open an entry first" notice); pressing Tab to focus the pane (with pane open) then `/` opens pane search
**Dependencies:** cavekit-detail-pane R7 (detail-pane search activation target), cavekit-entry-list R13 (list search activation target), cavekit-config (theme for keyhint styling)

### R14: Global Key-Intercept Ordering under Active In-Pane Search
**Description:** Global single-key reservations at the top of the app-shell key handler (currently `q` for quit, `Tab` for focus cycle, `?` for help, `Esc` priority chain) MUST NOT pre-empt a pane's active in-pane search input handler. When any pane's search is in **input mode**, every printable rune — including reserved letters like `q` — MUST reach the search model before global reservations apply. `Tab` is exempt: it remains a valid search-dismissal trigger (R11 focus cycle clears search per entry-list R13 AC 7). `Esc` is exempt: it remains the canonical dismissal key and is routed through the pane's search handler first (entry-list R13, detail-pane R7), only falling through to the Esc priority chain (R11) when search is already dismissed. `?` opens the help overlay regardless — search state is preserved while the overlay is up and restored on Esc. The reservation split exists so a user typing a query containing any reserved letter never loses the query or the session.
**Acceptance Criteria:**
- [ ] [auto] With list-search active in input mode, pressing `q` extends the query by `q` — the app does NOT emit `tea.Quit`
- [ ] [auto] With detail-pane-search active in input mode, pressing `q` extends the query by `q`
- [ ] [auto] After list-search is dismissed (Esc / commit-to-navigate), `q` with list focus emits `tea.Quit` as usual
- [ ] [auto] With search active in input mode, `Tab` exits input mode via the focus-cycle branch (dismisses search per entry-list R13 AC 7) instead of being typed into the query
- [ ] [auto] With search active in input mode, `?` opens the help overlay; after `?`-Esc the search model still has its prior query and mode
**Dependencies:** cavekit-entry-list R13 (list search target), cavekit-detail-pane R7 (pane search target), R11 (focus cycle via Tab), R13 (`/` activation)

### R15: Layout Resize Controls (Mouse Drag)
**Description:** The pane divider is draggable with the mouse. In below-mode the divider is the horizontal border between the entry list and the detail pane; in right-mode it is the 1-cell vertical divider (cavekit-app-shell R2). Pressing mouse-button-1 within the divider cell initiates a drag; subsequent mouse-move events while the button remains held update the active pane ratio **live** (one visible move per event — no throttling that produces stutter on typical displays); releasing the button persists the final ratio to the config file via a **single** write-back (not one per mouse-move frame). Drag is **focus-neutral** — it does not change which pane is focused. Focus transfers are governed by app-shell R6 click-in-pane routing, which does NOT fire for clicks on the divider cell itself. Dragging respects the same clamp as R12: the detail pane's ratio stays in `[0.10, 0.80]`. Dragging past the clamp pins the ratio at the boundary; further motion in the same direction is a no-op until the cursor re-enters the valid range (drag does NOT auto-close the pane — that behaviour is owned by R7 and is triggered by terminal resize, not by drag).
**Acceptance Criteria:**
- [ ] [auto] Pressing and holding mouse-button-1 on the divider cell initiates a drag
- [ ] [auto] While dragging, mouse-move events update the active pane ratio live — `height_ratio` in below-mode, `width_ratio` in right-mode — proportional to cursor displacement along the drag axis
- [ ] [auto] Releasing mouse-button-1 writes the final ratio to the config file exactly once per drag (not once per mouse-move frame)
- [ ] [auto] Drag works in below-mode (horizontal divider, vertical drag axis): dragging down grows the detail pane, dragging up shrinks it
- [ ] [auto] Drag works in right-mode (vertical divider, horizontal drag axis): dragging left grows the detail pane, dragging right shrinks it
- [ ] [auto] Dragging does not change which pane has keyboard focus
- [ ] [auto] Dragging past the `[0.10, 0.80]` clamp pins the ratio at the boundary; further cursor motion in the same direction is a no-op until the cursor re-enters the valid range
- [ ] [auto] Starting a drag with the detail pane closed is a silent no-op — there is no divider cell to grab
- [ ] [human, tui-mcp] Verify via tui-mcp first: launch the TUI on `logs/small.log`, open the detail pane, use `send_mouse` to emit a press-hold on the divider cell, a series of move events to the target coordinate, and a release; `snapshot` before and after and verify the divider moved, pane dimensions changed proportionally to the drag distance, and the config file reflects the new ratio. Repeat for both below-mode and right-mode. If `send_mouse` cannot reliably emit the button-hold + move + release sequence on this platform, record the limitation in a comment and fall back to direct human sign-off (per the user's HUMAN-sign-off-via-tui-mcp preference)
**Dependencies:** cavekit-detail-pane (applies ratio change, R6), cavekit-config (live write-back, R6), R6 (mouse routing to the divider cell), R12 (shared clamp range and ratio storage)

## Out of Scope

- Domain-specific logic (parsing, filtering, rendering details -- all delegated)
- Configuration management (handled by config)
- Multi-window or split-pane layouts beyond the defined header/list/detail/status structure

## Cross-References

- See also: cavekit-log-source.md (invoked for loading, provides progress and tail status)
- See also: cavekit-entry-list.md (main area content, mouse routing target)
- See also: cavekit-detail-pane.md (bottom pane content, mouse routing target)
- See also: cavekit-filter-engine.md (filter panel overlay, filtered counts)
- See also: cavekit-config.md (theme for header/status bar styling)

## Changelog

### 2026-04-18 — Revision (R12 focus-aware keyboard resize + new R15 mouse drag + R6 drag delegation)
- **Affected:** R6, R12 (rewritten), new R15
- **Summary:** R12 rewritten to make keyboard resize **focus-aware**: `|`/`+`/`-`/`=` act on the currently focused pane (detail or list). Preset set reduced from three values (0.10/0.30/0.70) to two (0.30/0.50) — the 0.10 preset made the detail pane too small to read and 0.70 left the list too narrow for useful log-scanning. `=` still resets to the layout default (detail=0.30) regardless of focus, because "reset" is a global return-to-baseline, not focus-directional. New R15 separates mouse-drag resize from keyboard resize (they share the ratio storage and clamp but have different trigger models and deserve independent testability). The mouse-drag AC that was shipped under R6 passed a shallow test but the feature didn't actually work in practice — R15 tightens the contract: live ratio update on every mouse-move while button is held, single persist-on-release, focus-neutral, clamp-pin (not auto-close). R6 loses the drag AC (moved to R15) and gains a one-line delegation note. Both R12 and R15 add `[human, tui-mcp]` ACs so the next build verifies keyboard-resize and mouse-drag via the tui-mcp harness first (per the user's HUMAN-sign-off-via-tui-mcp preference) before direct human sign-off.
- **Driven by:** User feedback in /ck:sketch session: "details pane presets: too many, 0.3 and 0.5 is enough", "mouse drag (A5) does not work", "when list pane is focused, +/- targeting it would be nice". The latter generalised to focus-aware semantics across all four keys (B1 choice). Accompanied by a new entry-list R10 revision for the click-row offset bug.

### 2026-04-18 — Revision (new R14 global key-intercept ordering under active search)
- **Affected:** new R14
- **Summary:** Added R14 to legislate that global single-key reservations (`q`, `Tab`, `?`, `Esc`) must not pre-empt an active in-pane search in input mode. Discovered during `/ck:check` after Tier 15/16 build: T-144-fix rewired Enter/Esc into the search router but left `q` on the pre-intercept side — a user typing `query`/`queue`/`quit` into list-search would quit the app and lose their session. R14 codifies the required ordering so the fix has a kit anchor.
- **Driven by:** `/ck:check` 2026-04-18 (post-Tier 15/16), finding F-106 (P1 data-loss regression). Paired task T-146.

### 2026-04-18 — Revision (R9 clipboard feedback + R13 focus-based `/` routing)
- **Affected:** R9, R13
- **Summary:** R9 (Clipboard) expanded from 4 to 10 ACs to require **visible user feedback** on every `y` press: success notice with copied entry count, error notice surfacing clipboard-write failures, and a "no marked entries" notice for zero-mark case. The implementation MUST NOT use `//nolint:errcheck` to swallow the error — the silent `_ = CopyMarkedEntries(...)` pattern is now a kit violation. R13 rewritten from "pane-open-state based routing" to **focus-based routing**: `/` activates list search (entry-list R13) when list focused, detail-pane search (detail-pane R7) when pane focused, and is a literal character when filter panel focused. The earlier "open an entry first" notice is removed — list search works standalone.
- **Driven by:** `/ck:check` run 2026-04-18 after user notes "list: copy of marked message does not copy the entries" (F-102) and "list: no search" (F-101). R13 re-routing paired with new cavekit-entry-list.md R13 (list search).

### 2026-04-18 — Revision (R11 open-time focus policy)
- **Affected:** R11
- **Summary:** R11 now explicitly codifies that opening a pane does NOT transfer focus — focus stays wherever the user put it. Focus transfers happen only on Tab, mouse click (R6), or `/` (R13). Added ACs "opening detail pane does not transfer focus" and "Esc from list-focus with pane open closes the pane". Before this revision R11 described Tab/Esc/click flows but was silent on what happens when a pane opens, so the implementation grabbed focus on Enter — breaking the preview flow where `j`/`k` on the list should update the pane live without needing a Tab-round-trip each time.
- **Driven by:** `/ck:check` run on 2026-04-18, finding F-017 (auto-focus on open breaks preview flow). Paired with cavekit-detail-pane R1 revision.

### 2026-04-18 — Revision (R13 cross-pane search activation)
- **Affected:** new R13
- **Summary:** New R13 added to close a discoverability gap: with the entry list focused, pressing `/` fell through to `list.Update` which doesn't bind it, producing a silent no-op. Users with vim muscle memory try `/` first and believe search is broken. R13 requires `/` to either focus-transfer to the detail pane (if open) or surface a transient notice (if closed), and demands that the key-hint bar + help overlay advertise the key with accurate scope. Paired with cavekit-detail-pane R7 revisions that fix the rendering side.
- **Driven by:** `/ck:check` finding F-001 (silent no-op on list focus) and F-011 (keyhint discoverability). Related findings logged in `context/impl/impl-review-findings.md`.

### 2026-04-16 — Revision
- **Affected:** R3, new R10
- **Summary:** R3 updated to require visually distinct header bar (background color from theme), cursor position display (1-based index), and human sign-off criterion. New R10 added for focus indicator when multiple panes are visible, so user can identify which pane receives keyboard input. Driven by user observation that header blends into log lines and cursor location is unclear after opening detail pane.
- **Commits:** manual testing feedback (no commit)

### 2026-04-16 — Revision (layout fixes)
- **Affected:** R4, R7, R10
- **Summary:** R4: added requirement that key-hint bar must be exactly 1 row (truncated, never wrapped) — wrapping to 2 lines overflowed StatusBarHeight=1 and pushed the header off-screen. R7: added requirement that WindowSizeMsg must be processed by child models even when empty — the initial resize arrives before async loading finishes, causing width/height to remain at initialization defaults. R10: clarified that focus indicator must not alter pane dimensions — wrapping a pane's rendered output with a lipgloss border post-render adds rows/columns that corrupt the layout. Removed entry-list-side focus border requirement; only the detail pane shows a focus border (rendered within its own View).
- **Commits:** uncommitted (session fixes)

### 2026-04-17 — Revision (details-pane redesign)
- **Affected:** R2, R3, R4, R6, R7, R10 (replaced), new R11, new R12
- **Summary:** R2 extended with three orientation modes (below/right/auto), right-split composition, border accounting, and the 60x15 minimum-viable floor. R3 gained narrow-mode header degradation order. R4 gained the focus label on the right side of the key-hint bar. R6 gained horizontal mouse zones, a 1-cell buffer near the divider, and focus-on-click. R7 gained auto-orientation re-evaluation, independent preservation of height_ratio and width_ratio across flips, and auto-close on minimum-dimension underflow. R10 was replaced by the pane visual-state matrix (focused / unfocused-visible / alone) per DESIGN.md §4. New R11 codifies the Tab-cycles and Esc-context-sensitive focus model. New R12 codifies the uniform layout-resize keymap (| presets, +/- nudges, = reset, [0.10, 0.80] clamp) with live config write-back.
- **Driven by:** DESIGN.md + research-brief-details-pane-redesign.md
