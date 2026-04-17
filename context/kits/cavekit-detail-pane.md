---
created: "2026-04-15T00:00:00Z"
last_edited: "2026-04-17T21:40:06+03:00"
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
- [ ] [auto] The detail pane's top border is visible in both below and right orientations (per DESIGN.md §4.4)
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

### R6: Pane Size and Resize
**Description:** In below-mode the pane's height is a ratio of terminal height (`height_ratio`, default 0.30). In right-mode the pane's width is a ratio of terminal width (`width_ratio`, default 0.30). Both ratios are preserved independently across orientation flips. The resize keymap itself (preset cycling, ±0.05 nudges, reset, clamping, live write-back) is defined in cavekit-app-shell R12 and is not duplicated here. Mouse drag on the visible divider resizes the appropriate dimension — height in below-mode, width in right-mode.
**Acceptance Criteria:**
- [ ] [auto] The detail pane opens at the configured height ratio
- [ ] [auto] Pressing `+` while the detail pane is focused increases its height
- [ ] [auto] Pressing `-` while the detail pane is focused decreases its height
- [ ] [auto] After a terminal resize event, the pane maintains its proportional height
- [ ] [auto] Mouse drag on the pane divider resizes the pane
- [ ] [auto] In right orientation the detail pane opens at the configured width ratio
- [ ] [auto] Pressing `+` while the detail pane is focused in right orientation increases its width ratio
- [ ] [auto] Pressing `-` while the detail pane is focused in right orientation decreases its width ratio
- [ ] [auto] Mouse drag on the vertical divider in right orientation resizes the pane width
- [ ] [auto] Flipping orientation from below to right preserves the previous height_ratio, and flipping back restores it
**Dependencies:** cavekit-config (detail pane height ratio, width ratio), cavekit-app-shell (R12 resize keymap, resize events, mouse routing)

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

### R9: Wrap Mode
**Description:** When a rendered content line is wider than the detail pane's content area, the `wrap_mode` config setting controls behavior. In this revision only `soft` is implemented — content lines wrap at the pane width so no content is silently hidden. `scroll` (horizontal scroll inside the pane) and `modal` (single-field expand view) are declared as future states for v1.5 and v2 respectively; they have no requirements here beyond their declaration as out-of-scope.
**Acceptance Criteria:**
- [ ] [auto] When wrap_mode = "soft" and a rendered line exceeds the pane's content width, the line wraps at the pane width
- [ ] [auto] No content is hard-truncated without a visible indicator
- [ ] [auto] When content wraps, total rendered height does not exceed the allocated pane height — overflow is navigated via the existing scroll model
**Dependencies:** cavekit-config (wrap_mode setting), R4 (scroll model)

### R10: Width Awareness and Safe Measurement
**Description:** The detail pane accepts a width (allocated by the layout) and measures all rendered content with an emoji-, CJK-, and ANSI-safe width measurement. Byte-length measurement is not acceptable — it miscounts multi-byte characters and ANSI escape sequences, producing column drift and pane overflow.
**Acceptance Criteria:**
- [ ] [auto] The detail pane renders correctly when given different widths; the rendered output's outer width equals the allocated width
- [ ] [auto] Rendering a line containing multi-byte characters (emoji or CJK) does not produce column drift or pane overflow
- [ ] [auto] Rendering a line containing ANSI escape sequences does not produce column drift or pane overflow
**Dependencies:** cavekit-app-shell (allocates pane width)

## Out of Scope

- Filter logic and filter panel (handled by filter-engine)
- Clipboard operations (handled by app-shell)
- Entry selection logic (handled by entry-list)
- Horizontal scroll of content (planned for v1.5 as `wrap_mode = "scroll"`)
- Modal single-field expand view (planned for v2 as `wrap_mode = "modal"`)
- Tree-fold rendering of nested JSON (not planned)

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

### 2026-04-17 — Revision (details-pane redesign)
- **Affected:** R1, R6 (renamed), new R9, new R10
- **Summary:** R1 gained an AC that the detail pane's top border is visible in both below and right orientations. R6 renamed from "Pane Height and Resize" to "Pane Size and Resize" and extended with width-ratio semantics for right orientation, mouse-drag on the vertical divider, and independent ratio preservation across orientation flips; the resize keymap itself is delegated to cavekit-app-shell R12. New R9 introduces `wrap_mode` with `soft` as the shipping default (preventing silent truncation). New R10 requires width awareness and safe width measurement for emoji, CJK, and ANSI content. Out of Scope extended with v1.5/v2 wrap modes and non-planned tree-fold.
- **Driven by:** DESIGN.md + research-brief-details-pane-redesign.md
