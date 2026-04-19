---
created: "2026-04-15T00:00:00Z"
last_edited: "2026-04-19T21:30:00+03:00"
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
**Description:** When the terminal is resized, all panes reflow to fill the new dimensions. No crashes, no layout corruption, and pane proportions are maintained. WindowSizeMsg must be processed by all child models even when they have no data yet — the initial resize arrives before async file loading completes. When the detail pane orientation is set to `auto`, every terminal-resize event re-evaluates orientation against the configured threshold. Both the below-mode height ratio and the right-mode width ratio are preserved independently across orientation flips. Every WindowSizeMsg-driven orientation change must also propagate into pane-local rendering flags that depend on orientation — the resize handler and the `relayout()` path must refresh the same flags, or the pane's rendered state drifts from the declared orientation.
**Acceptance Criteria:**
- [ ] [auto] After a terminal resize, the layout fills the new terminal dimensions
- [ ] [auto] Pane proportions (e.g. detail pane height ratio) are preserved after resize
- [ ] [auto] No content is clipped or overlapping after resize
- [ ] [auto] Resize does not cause a crash or panic
- [ ] [auto] Child models (e.g. entry list) process WindowSizeMsg even when their data set is empty
- [ ] [auto] When detail_pane.position is "auto", orientation is re-evaluated on every terminal-resize event against orientation_threshold_cols
- [ ] [auto] height_ratio and width_ratio are both preserved across orientation flips — flipping from below to right does not overwrite one with the other, and flipping back restores the previous values
- [ ] [auto] When the detail pane's computed dimension falls below the minimum (width < 30 cells in right orientation, height < 3 rows in below orientation) the pane auto-closes and the status bar emits a one-time notice
- [ ] [auto] When `detail_pane.position` is "auto" and the detail pane is open, a WindowSizeMsg that crosses `orientation_threshold_cols` must refresh every pane-local rendering flag that depends on orientation (at minimum the detail pane's below-mode flag that drives the R10 drag-seam top-border paint). Post-flip, the rendered drag-seam SGR at the correct seam location (right: `│` glyph column mid-Y; below: the detail pane's top-border row) must match the NEW orientation's seam contract per R10 AC 10, not the pre-flip state. The regression test must exercise both flip directions (right→below and below→right) with the pane open throughout and assert the rendered SGR at each step
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
- [ ] [auto] The pane-resize drag seam (R15) renders in `DragHandle` color independent of focus: in right-mode the 1-cell `│` divider glyph; in below-mode the detail pane's top border row (the row that physically sits at the detail pane's top edge; the list pane's own bottom border is an adjacent, separate row rendered in the list's focus-state color, NOT shared with or co-painted as part of the seam)
**Dependencies:** cavekit-entry-list, cavekit-detail-pane, cavekit-config (DividerColor, UnfocusedBg, FocusBorder, DragHandle tokens)

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
**Description:** The pane divider is draggable with the mouse. In below-mode the divider is the detail pane's top border row (which sits immediately below the list's own bottom border; only the detail pane's top border is painted in `DragHandle`, not the list's bottom — they are two adjacent rows, not a shared row); in right-mode it is the 1-cell vertical divider (cavekit-app-shell R2). Pressing mouse-button-1 within the divider cell initiates a drag; subsequent mouse-move events while the button remains held update the active pane ratio **live** (one visible move per event — no throttling that produces stutter on typical displays); releasing the button persists the final ratio to the config file via a **single** write-back (not one per mouse-move frame). Drag is **focus-neutral** — it does not change which pane is focused. Focus transfers are governed by app-shell R6 click-in-pane routing, which does NOT fire for clicks on the divider cell itself. Dragging respects the same clamp as R12: the detail pane's ratio stays in `[0.10, 0.80]`. Dragging past the clamp pins the ratio at the boundary; further motion in the same direction is a no-op until the cursor re-enters the valid range (drag does NOT auto-close the pane — that behaviour is owned by R7 and is triggered by terminal resize, not by drag).
**Acceptance Criteria:**
- [ ] [auto] Pressing and holding mouse-button-1 on the divider cell initiates a drag
- [ ] [auto] While dragging, mouse-move events update the active pane ratio live — `height_ratio` in below-mode, `width_ratio` in right-mode — proportional to cursor displacement along the drag axis
- [ ] [auto] Releasing mouse-button-1 writes the final ratio to the config file exactly once per drag (not once per mouse-move frame)
- [ ] [auto] Drag works in below-mode (horizontal divider, vertical drag axis): dragging **up** grows the detail pane (the divider moves up, list shrinks, detail area grows), dragging **down** shrinks it — the detail pane is physically below the divider, so divider motion and detail size are inversely related
- [ ] [auto] Drag works in right-mode (vertical divider, horizontal drag axis): dragging left grows the detail pane, dragging right shrinks it
- [ ] [auto] Dragging does not change which pane has keyboard focus
- [ ] [auto] Dragging past the `[0.10, 0.80]` clamp pins the ratio at the boundary; further cursor motion in the same direction is a no-op until the cursor re-enters the valid range
- [ ] [auto] Starting a drag with the detail pane closed is a silent no-op — there is no divider cell to grab
- [ ] [auto] The router's divider cell MUST coincide with the renderer's visible divider glyph — a test MUST render the layout, locate the visible divider glyph (`│` in right-split at mid-Y; the detail pane's top border row in below-mode), and assert `MouseRouter.Zone(glyphX, glyphY) == ZoneDivider` across all preset ratios {0.10, 0.30, 0.50, 0.80} in both orientations at 140x35 and 80x24. Synthetic unit tests that derive press coordinates from the router's own helpers are insufficient — they agree with the router by construction and cannot detect router/renderer drift. Any ANSI-stripping helper used to locate glyph columns in the rendered output MUST handle the full ECMA-48 CSI two-step form (`ESC [ <params/intermediates> <final-byte 0x40..0x7e>`) — a hardcoded terminator subset is insufficient because future styling layers may emit non-SGR CSI sequences (cursor positioning, mode setting, function-key codes) whose terminators would otherwise leak escape bytes into the stripped output and silently corrupt the glyph-column index
- [ ] [auto] When `RatioFromDragY` / `RatioFromDragX` is inverted against the layout's forward ratio→size math (e.g. `HeightModel.PaneHeight = int(termHeight * ratio)`), a Press directly on the currently rendered divider row/col MUST yield the current ratio unchanged — no step-snap on Press-without-motion. The regression test for this AC MUST exercise BOTH `RatioFromDragX` AND `RatioFromDragY` independently — covering one axis is insufficient because each has its own inverse formula tied to its own forward math (`Layout.DetailContentWidth` / `detailpane.HeightModel.PaneHeight`). The X-axis canonical Press column MUST be derived from `Layout.ListContentWidth()` (the renderer-truth divider X established by R15 line 198 / T-160), not from the inverse formula itself — deriving from the inverse would tautologically agree with whatever the inverse computes
- [ ] [auto] If the detail pane auto-closes (R7) mid-drag — e.g. a terminal resize takes the pane's computed dimension below the minimum — the drag session MUST terminate silently: subsequent Motion events are swallowed, no ratio mutation occurs, and no config write fires on the eventual Release
- [ ] [auto] A bare Press+Release on the divider cell with no intermediate Motion MUST NOT rewrite the config file — persistence fires only when the drag actually changed the ratio
- [ ] [auto] When the terminal's active dimension is degenerate (termHeight or termWidth ≤ 0; this window exists briefly at startup before the first WindowSizeMsg) the drag helpers MUST preserve the current ratio rather than jumping to `RatioDefault` — no silent shadowing of a persisted value. The regression test for this AC MUST drive the guard with `m.pane.IsOpen() == true` AND the active terminal dimension == 0 simultaneously — going through `WindowSizeMsg{Width:0}` is insufficient because the auto-close branch (R7) short-circuits the test flow before the degenerate-dim guard fires. Removing the caller-guard MUST make the test fail.
- [ ] [human, tui-mcp] Verify via tui-mcp first: launch the TUI on `logs/small.log`, open the detail pane, use `send_mouse` to emit a press-hold on the divider cell, a series of move events to the target coordinate, and a release; `snapshot` before and after and verify the divider moved, pane dimensions changed proportionally to the drag distance, and the config file reflects the new ratio. Repeat for both below-mode and right-mode. If `send_mouse` cannot reliably emit the button-hold + move + release sequence on this platform, record the limitation in a comment and fall back to direct human sign-off (per the user's HUMAN-sign-off-via-tui-mcp preference)
- [ ] [auto] The drag-initiating cell (divider in right-mode; the detail pane's top border row in below-mode — NOT the list's bottom border, which remains in its own focus-state color on an adjacent row) renders in `DragHandle`, which is distinct from BOTH `DividerColor` and `FocusBorder` in every bundled theme — a test MUST assert, per theme, that `DragHandle != DividerColor` AND `DragHandle != FocusBorder`, and that the rendered color at the drag-seam cell is `DragHandle` (not the adjacent pane's border colour). A pinning test MUST also assert the inverse: the list pane's bottom border row, rendered via `PaneStyle` in below-mode, does NOT carry `DragHandle` SGR — the seam is strictly scoped to the detail pane's top edge
- [ ] [human, tui-mcp] Verify via tui-mcp: launch the TUI on `logs/small.log`, snapshot right-split with the list focused, and confirm the 1-cell divider between list and detail is visibly distinct in hue/luminance from the detail pane's unfocused border. Repeat in below-split: the detail pane's top border row must be visibly distinct from the unfocused-pane top/bottom borders (and from the list pane's bottom border, which is an adjacent row above it). Verify for all three bundled themes (`tokyo-night`, `catppuccin-mocha`, `material-dark`)
**Dependencies:** cavekit-detail-pane (applies ratio change, R6), cavekit-config (live write-back, R6; DragHandle token, R4), R6 (mouse routing to the divider cell), R7 (auto-close on minimum-dimension underflow), R12 (shared clamp range and ratio storage)

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

### 2026-04-19 — Revision (R10/R15 drag-seam language: detail-top, not shared row — F-201)
- **Affected:** R10 AC 10; R15 description, ACs at lines 200/206/207
- **Summary:** /ck:check Tier 23 Pass 2 surfaced F-201: the kit described the below-mode drag seam as a "horizontal border row between list and detail" or "shared border row", implying a single row co-owned by both panes. In reality the render is two adjacent rows — the list pane's own bottom border (rendered by its PaneStyle in list-focus color) and the detail pane's own top border (rendered by PaneStyle + WithDragSeamTop in DragHandle) — and only the detail pane's top is overridden. `MouseRouter.Zone` and the paint both target the detail-pane top only. Classification: `wrong_criterion` — code is correct, spec language misdescribes the physical render. Kit language rewritten across 5 locations to say "detail pane's top border row"; R15 AC 11 (line 206) extended to mandate a pinning test asserting the list pane's bottom row does NOT carry DragHandle SGR. Companion edits in DESIGN.md §2 token table, §2 three-new-tokens paragraph, §4.5 drag-handle seam, §6 pane border matrix.
- **Driven by:** `/ck:revise --trace` cycle on `/ck:check` finding F-201. Pinning regression test `TestPaneStyle_DragSeamOnlyOverridesDetailTop_NotListBottom` in `internal/ui/appshell/panestyle_test.go`. No code change (the code was already correct; this is a pure spec-language fix).

### 2026-04-19 — Revision (R7 orientation-flip re-render contract — F-200)
- **Affected:** R7 (description extended; new AC 9 appended)
- **Summary:** /ck:check Tier 23 Pass 2 surfaced F-200: `WithBelowMode` was wired in `relayout()` (model.go:735) but omitted from the inline pane-wiring chain in the `WindowSizeMsg` handler (model.go:192-195). When `position=auto` + pane open, a terminal resize that crossed `orientation_threshold_cols` flipped `m.resize.Orientation()` but left `m.pane.belowMode` stale — so the detail pane's R10 drag-seam top-border rendered in the pre-flip orientation's colors. `TestModel_OrientationFlip_VerticalSizeTracks` at model_test.go:1154 exercised the exact flow but did not assert on seam SGR, so the bug shipped green. R7 description extended to mandate that the resize handler and `relayout()` refresh the same pane-local orientation-dependent flags. New AC 9 mandates that the post-flip rendered drag-seam SGR at the correct seam location match the NEW orientation's contract per R10 AC 10, with a regression test exercising both flip directions.
- **Driven by:** `/ck:revise --trace` cycle on `/ck:check` finding F-200. Regression test in `internal/ui/app/model_f200_test.go`. Code change in `internal/ui/app/model.go`.

### 2026-04-19 — Revision (R10/R15 DragHandle visual affordance)
- **Affected:** R10, R15
- **Summary:** Introduce a new theme token `DragHandle` to colour the pane-resize drag seam distinctly from the unfocused-pane-border colour (`DividerColor`). Before: in right-split, the 1-cell `│` divider glyph and the borders of any unfocused adjacent pane both rendered in `DividerColor`, yielding a 3-cell uniform band with no visual cue for the draggable seam; below-split had the symmetric problem on the shared horizontal border row. After: the drag seam renders in `DragHandle` — a mid-tone neutral brighter than `DividerColor` but dimmer than `FocusBorder` — independent of focus. New R10 AC enforces the seam colour; new R15 ACs (one auto, one human/tui-mcp) enforce distinctness from `DividerColor` across all three bundled themes. Companion edits in `cavekit-config.md` R4 (token declaration) and `DESIGN.md` §2 token table / §2 three-new-tokens paragraph / §4 matrix cross-cutting / §4.5 divider glyph fg.
- **Driven by:** User feedback in /ck:sketch session: "the border between the panes from which the dragging is done is the same color as the border of the unfocused pane. We should use distinct color so it would be easier to spot where to drag."

### 2026-04-19 — Revision (R15 renderer-truth AC: ANSI-strip CSI coverage — F-134)
- **Affected:** R15 (renderer-truth AC text extended at line 198)
- **Summary:** /ck:review Pass 2 surfaced F-134: the `stripAnsi` helper backing `locateGlyphCol` (which the R15 line-198 renderer-truth assertion depends on) used a hardcoded CSI terminator subset {m, K, H, A, B, C, D, J} instead of the full ECMA-48 final-byte range 0x40..0x7e. Today lipgloss only emits SGR (`m`), so the bug was latent — but any future styling layer emitting cursor positioning, mode setting, or function-key sequences would silently leak escape bytes into the stripped output, corrupting the glyph-column index and giving false-positive renderer-truth assertions. AC text extended to mandate that any ANSI-stripping helper handle the full ECMA-48 CSI two-step form (`ESC [ <params/intermediates> <final-byte 0x40..0x7e>`). `stripAnsi` rewritten as a proper three-state machine (`stPlain` / `stPostEsc` / `stCsiBody`) so the introducer `[` is consumed as introducer, not mistaken for a final byte. New regression test `TestStripAnsi_HandlesFullCSIFinalByteRange` covers 9 non-SGR CSI sequences (HVP, CHA, DECTCEM, function-key terminator `~`, DSR, DA, save/restore cursor) — all 9 fail with the hardcoded subset and pass after the state-machine rewrite.
- **Driven by:** `/ck:revise --trace` cycle on `/ck:review` finding F-134. Regression test in `internal/ui/appshell/mouse_test.go`. Code change in same file (test-helper).

### 2026-04-19 — Revision (R15 inverse-math AC: X-axis parity + formula fix — F-133)
- **Affected:** R15 (inverse-math AC text extended; `RatioFromDragX` formula corrected)
- **Summary:** /ck:review Pass 2 surfaced F-133: the R15 inverse-math AC mentioned BOTH `RatioFromDragY` AND `RatioFromDragX`, but only the Y-axis had a regression test. The X-axis math (`detail = termWidth - x - 2`) was off by 3 cells against the renderer-truth divider X established by T-160 — at termWidth=100, ratio=0.55, Press-at-current-X returned 0.589 (drift 0.039, exceeding the RatioStep/2=0.025 tolerance the Y-axis test uses). The author of T-161 explicitly punted on this in `ratiokeys.go` ("X-axis analogue of F-123 is present... left unchanged because the T-104 tests encode the current semantics"). The buggy T-104 pin (`x=50, termWidth=100 → 48/95`) encoded the broken formula. AC text extended to mandate parallel regression tests for BOTH axes, and that the X-axis canonical Press column MUST be sourced from `Layout.ListContentWidth()` (renderer-truth) rather than the inverse formula itself. `RatioFromDragX` rewritten as the exact inverse of `DetailContentWidth = usable - ListContentWidth`: `detail := usable - x`. T-104 mid pin updated from 48/95 → 45/95 to reflect the correct formula. New regression test `TestRatioFromDragX_PressAtCurrentDividerX_KeepsRatio` mirrors the Y-axis test, sweeping presets {0.30, 0.50, 0.55} × termWidth ∈ {80, 100}.
- **Driven by:** `/ck:revise --trace` cycle on `/ck:review` finding F-133. Regression test in `internal/ui/appshell/ratiokeys_test.go`. Code change in `internal/ui/appshell/ratiokeys.go`.

### 2026-04-19 — Revision (R15 degenerate-dim AC sharpened — F-132)
- **Affected:** R15 (degenerate-dim AC text extended)
- **Summary:** /ck:review Pass 2 surfaced F-132: the existing T-165 tests for the degenerate-dim guard drove a 0-dim `WindowSizeMsg` which auto-closed the pane, so the subsequent Motion was short-circuited by the prior `!m.pane.IsOpen()` guard at `model.go:524` rather than the `termW/termH<=0` guard at `:554-556`/`:565-567`. Removing the degenerate-dim guards left the T-165 tests green — proving the tests asserted the right behaviour via the wrong code path. AC text extended to mandate that the regression test must drive the guard with `pane.IsOpen()==true` AND `termDim==0` simultaneously, and removing the caller-guard must make the test fail. New regression tests `TestModel_F132_DegenerateDim_{Right,Below}_GuardFiresWith_PaneOpen` re-open the pane via `openPane` after the 0-dim resize (relayout does not re-trigger auto-close) so the IsOpen() guard passes and only the degenerate-dim guard prevents shadowing. Old T-165 tests deleted as superseded.
- **Driven by:** `/ck:revise --trace` cycle on `/ck:review` finding F-132. Regression tests in `internal/ui/app/model_test.go`.

### 2026-04-19 — Revision (R15 hardening from Tier 19 `/ck:check`)
- **Affected:** R15 (5 new ACs, AC 4 text inverted to match physical layout)
- **Summary:** Tier 19 HUMAN sign-off surfaced three real behavioural gaps that T-156 unit tests did not catch because the tests used the router's own coordinate helpers to synthesize press/motion events — agreeing with the router's model trivially. R15 expanded with: (a) **renderer-truth divider-col assertion** closing F-122 P1 (router's divider col and Lipgloss-rendered `│` col must match across all presets in both orientations), (b) **inverse-math invariant** closing F-123 P2 (`RatioFromDragY/X` must be an exact inverse of `PaneHeight/Width` forward math — Press on current divider yields current ratio, not a one-step snap), (c) **mid-drag auto-close termination** closing F-125 P2 (drag state must die when R7 auto-closes the pane mid-gesture — no ratio mutation, no config write), (d) **no-motion-no-persist** closing F-129 P3 (bare Press+Release on divider must not rewrite config), (e) **degenerate-dimension guard** closing F-126 P3 (startup termHeight/Width=0 window must not shadow persisted ratio with `RatioDefault`). AC 4 text flipped from "dragging down grows the detail pane" to "dragging up grows" — T-156's implementation is physically correct (detail is below the divider; divider-up = detail-grows), the prior kit text was a directional inversion logged as F-121.
- **Driven by:** `/ck:check` run 2026-04-19 on Tier 19, findings F-121, F-122, F-123, F-125, F-126, F-129. Paired tasks T-160..T-164, T-168 (below).

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
