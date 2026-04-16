---
created: "2026-04-15T00:00:00Z"
last_edited: "2026-04-16T21:49:47+03:00"
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
**Description:** The terminal is divided into: a header bar (top), the entry list (main area), an optional detail pane (bottom, toggleable), and a status/key-hint bar (bottom-most). The layout fills the full terminal width and height.
**Acceptance Criteria:**
- [ ] [auto] The header bar is rendered at the top of the terminal
- [ ] [auto] The entry list occupies the main area between the header and the bottom bars
- [ ] [auto] When the detail pane is open, it appears between the entry list and the status bar
- [ ] [auto] The status/key-hint bar is rendered at the bottom of the terminal
- [ ] [auto] All panes together fill the full terminal width and height with no gaps or overlap
**Dependencies:** cavekit-entry-list, cavekit-detail-pane

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
**Dependencies:** cavekit-log-source (file name, tail status, entry count), cavekit-filter-engine (filtered count), cavekit-entry-list (cursor position)

### R4: Context-Sensitive Key-Hint Bar
**Description:** The bottom status bar shows relevant keybindings for the currently focused component. The hints update as focus changes between components (entry list, detail pane, filter panel, help overlay). The key-hint bar must occupy exactly 1 row — content that exceeds the terminal width is truncated (never wrapped), since the layout reserves StatusBarHeight=1.
**Acceptance Criteria:**
- [ ] [auto] When the entry list is focused, the key-hint bar shows entry-list keybindings (e.g. j/k, e/w, m, Enter)
- [ ] [auto] When the detail pane is focused, the key-hint bar shows detail-pane keybindings (e.g. j/k, /, +/-, Esc)
- [ ] [auto] When the filter panel is focused, the key-hint bar shows filter-panel keybindings (e.g. j/k, Space, d, Esc)
- [ ] [auto] Key hints update immediately when focus changes
- [ ] [auto] The key-hint bar renders as exactly 1 terminal row regardless of content length (truncated, not wrapped)
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
**Description:** Mouse mode is enabled globally. Mouse events are routed to the correct domain component based on the click/scroll position within the layout.
**Acceptance Criteria:**
- [ ] [auto] Mouse events in the entry list area are routed to the entry list
- [ ] [auto] Mouse events in the detail pane area are routed to the detail pane
- [ ] [auto] Mouse drag on the pane divider between entry list and detail pane triggers pane resize
- [ ] [auto] Mouse events do not cause crashes regardless of where in the terminal they occur
**Dependencies:** cavekit-entry-list (mouse handling), cavekit-detail-pane (mouse handling)

### R7: Terminal Resize Handling
**Description:** When the terminal is resized, all panes reflow to fill the new dimensions. No crashes, no layout corruption, and pane proportions are maintained. WindowSizeMsg must be processed by all child models even when they have no data yet — the initial resize arrives before async file loading completes.
**Acceptance Criteria:**
- [ ] [auto] After a terminal resize, the layout fills the new terminal dimensions
- [ ] [auto] Pane proportions (e.g. detail pane height ratio) are preserved after resize
- [ ] [auto] No content is clipped or overlapping after resize
- [ ] [auto] Resize does not cause a crash or panic
- [ ] [auto] Child models (e.g. entry list) process WindowSizeMsg even when their data set is empty
**Dependencies:** cavekit-detail-pane (pane proportions), cavekit-entry-list

### R8: Loading Indicator
**Description:** While the log source is reading the file on startup, a loading indicator is displayed. It disappears when loading completes.
**Acceptance Criteria:**
- [ ] [auto] While entries are being loaded, a loading indicator is visible
- [ ] [auto] When loading completes, the loading indicator is no longer visible
- [ ] [auto] The loading indicator shows progress (e.g. number of entries loaded so far)
**Dependencies:** cavekit-log-source (progress signals)

### R9: Clipboard
**Description:** Pressing `y` copies all marked entries to the system clipboard as JSONL (one JSON object per line). Non-JSON marked entries are included as raw text lines.
**Acceptance Criteria:**
- [ ] [auto] Pressing `y` with marked entries copies them to the system clipboard
- [ ] [auto] The clipboard content is JSONL: one entry per line in original order
- [ ] [auto] Non-JSON marked entries are included as raw text lines
- [ ] [auto] Pressing `y` with no marked entries does not modify the clipboard
**Dependencies:** cavekit-entry-list (marks)

### R10: Focus Indicator
**Description:** When multiple panes are visible (entry list + detail pane), the currently focused pane is visually indicated so the user knows which pane will receive keyboard input. The indicator must not alter the pane's rendered dimensions (no post-render border wrapping that adds rows or columns). Acceptable approaches: a colored border rendered as part of the pane's own View(), or a title bar highlight.
**Acceptance Criteria:**
- [ ] [auto] When the detail pane is open and focused, it has a visual indicator distinguishing it from the unfocused entry list (e.g. highlighted border or title)
- [ ] [auto] The focus indicator does not change the rendered width or height of any pane
- [ ] [auto] The focus indicator updates immediately when focus changes between panes
- [ ] [human] The focused pane is clearly identifiable at a glance
**Dependencies:** cavekit-entry-list, cavekit-detail-pane

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

### 2026-04-16 — Revision
- **Affected:** R3, new R10
- **Summary:** R3 updated to require visually distinct header bar (background color from theme), cursor position display (1-based index), and human sign-off criterion. New R10 added for focus indicator when multiple panes are visible, so user can identify which pane receives keyboard input. Driven by user observation that header blends into log lines and cursor location is unclear after opening detail pane.
- **Commits:** manual testing feedback (no commit)

### 2026-04-16 — Revision (layout fixes)
- **Affected:** R4, R7, R10
- **Summary:** R4: added requirement that key-hint bar must be exactly 1 row (truncated, never wrapped) — wrapping to 2 lines overflowed StatusBarHeight=1 and pushed the header off-screen. R7: added requirement that WindowSizeMsg must be processed by child models even when empty — the initial resize arrives before async loading finishes, causing width/height to remain at initialization defaults. R10: clarified that focus indicator must not alter pane dimensions — wrapping a pane's rendered output with a lipgloss border post-render adds rows/columns that corrupt the layout. Removed entry-list-side focus border requirement; only the detail pane shows a focus border (rendered within its own View).
- **Commits:** uncommitted (session fixes)
