---
created: "2026-04-15T00:00:00Z"
last_edited: "2026-04-16T21:49:47+03:00"
---

# Cavekit: Detail Pane

## Scope

The bottom pane that displays a pretty-printed view of the currently selected log entry. Covers JSON syntax highlighting, scrolling within the pane, per-field visibility toggling, in-pane search, resize behavior, and mouse interactions within the pane.

## Requirements

### R1: Activation and Dismissal
**Description:** The detail pane opens when the user presses Enter or double-clicks an entry in the list. It closes and returns focus to the entry list on Esc or Enter. When open, the detail pane has a visible top border or separator line so the boundary between the entry list and detail pane is clear. The border row must be subtracted from the allocated pane height before rendering content, so the total rendered output (border + content) fits exactly within the layout slot.
**Acceptance Criteria:**
- [ ] [auto] Pressing Enter while an entry is selected in the list opens the detail pane showing that entry
- [ ] [auto] Double-clicking an entry in the list opens the detail pane showing that entry
- [ ] [auto] Pressing Esc while the detail pane is focused closes it and returns focus to the entry list
- [ ] [auto] Pressing Enter while the detail pane is focused closes it and returns focus to the entry list
- [ ] [auto] When open, the detail pane is rendered with a visible top border or separator line distinguishing it from the entry list above
- [ ] [auto] The total rendered height of the detail pane (border + content) equals the allocated pane height — border rows are subtracted from content height before rendering
- [ ] [human] The boundary between entry list and detail pane is clearly visible
**Dependencies:** cavekit-entry-list (selection signal), cavekit-app-shell (focus indicator)

### R2: JSON Pretty-Print with Syntax Highlighting
**Description:** JSONL entries are rendered as formatted JSON with syntax highlighting. Keys, strings, numbers, booleans, and null each use distinct colors from the active theme. All fields are rendered, including arbitrary extra fields not known at compile time. No color values are hardcoded — all resolved from the active theme's tokens.
**Acceptance Criteria:**
- [ ] [auto] A JSONL entry is rendered as indented, formatted JSON
- [ ] [auto] All fields from the entry are present in the rendered output, including extra fields
- [ ] [auto] Rendering a JSONL entry produces ANSI output where JSON keys contain the active theme's key color token value
- [ ] [auto] Rendering a JSONL entry produces ANSI output where string values contain the active theme's string color token value
- [ ] [auto] Rendering a JSONL entry produces ANSI output where numeric values contain the active theme's number color token value
- [ ] [auto] Rendering a JSONL entry produces ANSI output where boolean values contain the active theme's boolean color token value
- [ ] [auto] Rendering a JSONL entry produces ANSI output where null values contain the active theme's null color token value
- [ ] [auto] Switching the active theme changes the ANSI color codes in the rendered output to match the new theme's tokens
- [ ] [human] One-time visual sign-off per bundled theme: syntax highlighting is perceptually correct and readable
**Dependencies:** cavekit-config (theme)

### R3: Non-JSON Entry Display
**Description:** Non-JSON entries display their raw text without any pretty-printing or syntax highlighting.
**Acceptance Criteria:**
- [ ] [auto] A non-JSON entry is displayed as plain raw text in the detail pane
- [ ] [auto] No JSON formatting is applied to non-JSON entries
**Dependencies:** none

### R4: Pane Scrolling
**Description:** When the detail pane content exceeds the pane height, the content is scrollable using `j`/`k` while the pane is focused.
**Acceptance Criteria:**
- [ ] [auto] Pressing `j` while the detail pane is focused scrolls the content down
- [ ] [auto] Pressing `k` while the detail pane is focused scrolls the content up
- [ ] [auto] Mouse scroll wheel over the detail pane scrolls the content
- [ ] [auto] Scrolling stops at the top and bottom of the content
**Dependencies:** none

### R5: Per-Field Visibility
**Description:** Individual fields can be hidden in the detail pane. Hidden fields are not rendered at all (not collapsed). Toggling a field's visibility immediately writes the change to config. The set of hidden fields is persisted across sessions.
**Acceptance Criteria:**
- [ ] [auto] When a field is marked as hidden in config, it does not appear in the detail pane output
- [ ] [auto] Toggling a field's visibility causes the detail pane to immediately re-render without the hidden field
- [ ] [auto] The visibility change is written to the config file immediately
- [ ] [auto] After restarting the application, previously hidden fields remain hidden
**Dependencies:** cavekit-config (detail pane field visibility, write-back)

### R6: Pane Height and Resize
**Description:** The detail pane height is configurable as a ratio of the terminal height (default 30%). The pane can be resized with `+`/`-` keys. Proportions survive terminal resize events.
**Acceptance Criteria:**
- [ ] [auto] The detail pane opens at the configured height ratio
- [ ] [auto] Pressing `+` while the detail pane is focused increases its height
- [ ] [auto] Pressing `-` while the detail pane is focused decreases its height
- [ ] [auto] After a terminal resize event, the pane maintains its proportional height
- [ ] [auto] Mouse drag on the pane divider resizes the pane
**Dependencies:** cavekit-config (detail pane height ratio), cavekit-app-shell (resize events, mouse routing)

### R7: In-Pane Search
**Description:** `/` opens a search input scoped to the detail pane content (does not trigger the list filter). `n`/`N` cycle through matches with a wrap indicator. Matches are highlighted using the theme's highlight color. Esc dismisses the search.
**Acceptance Criteria:**
- [ ] [auto] Pressing `/` while the detail pane is focused opens a search input within the pane
- [ ] [auto] Typing a search term highlights matching text in the pane content
- [ ] [auto] Pressing `n` moves to the next match
- [ ] [auto] Pressing `N` moves to the previous match
- [ ] [auto] When matches wrap around, a wrap indicator is displayed
- [ ] [auto] Pressing Esc dismisses the search input and clears highlights
- [ ] [auto] The search does not affect the entry-list filter
**Dependencies:** cavekit-config (theme highlight color)

### R8: Mouse Filter Interaction
**Description:** Clicking on a field value in the detail pane initiates adding a filter for that field, prompting for include or exclude mode.
**Acceptance Criteria:**
- [ ] [auto] Clicking on a field value in the detail pane triggers a filter prompt with the field name and value pre-filled
- [ ] [auto] The prompt allows choosing include or exclude mode
- [ ] [auto] Confirming the prompt adds the filter to the filter engine
**Dependencies:** cavekit-filter-engine (filter creation)

## Out of Scope

- Filter logic and filter panel (handled by filter-engine)
- Clipboard operations (handled by app-shell)
- Entry selection logic (handled by entry-list)

## Cross-References

- See also: cavekit-entry-list.md (provides entry selection signal, double-click activation)
- See also: cavekit-filter-engine.md (receives filter-add requests from mouse clicks)
- See also: cavekit-config.md (theme, field visibility, pane height ratio)
- See also: cavekit-app-shell.md (layout, resize events, mouse routing)

## Changelog

### 2026-04-16 — Revision
- **Affected:** R1
- **Summary:** R1 updated to require a visible top border/separator when the detail pane is open, so the boundary between entry list and detail pane is clear. Added human sign-off criterion and dependency on app-shell focus indicator. Driven by user observation that it's unclear where the list ends and details begin.
- **Commits:** manual testing feedback (no commit)

### 2026-04-16 — Revision (layout fixes)
- **Affected:** R1
- **Summary:** R1: added requirement that border rows must be subtracted from content height before rendering. The lipgloss top border adds 1 row to the rendered output; without accounting for it, the pane's total height exceeds the layout allocation and pushes the header off-screen. This is the Bubble Tea golden rule: "Always account for borders — subtract border rows BEFORE rendering panels."
- **Commits:** uncommitted (session fixes)
