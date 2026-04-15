---
created: "2026-04-15T00:00:00Z"
last_edited: "2026-04-15T00:00:00Z"
---

# Cavekit: Filter Engine

## Scope

The filter model, matching logic, filter panel overlay for managing active filters, and the mechanism for adding filters from field values. Emits a filtered entry index consumed by the entry list.

## Requirements

### R1: Filter Model
**Description:** Each filter has a field name, a pattern, a mode (include or exclude), and an enabled flag. Multiple filters can coexist.
**Acceptance Criteria:**
- [ ] [auto] A filter can be created with a field name, pattern, mode (include/exclude), and enabled state
- [ ] [auto] Multiple filters can be active simultaneously
- [ ] [auto] Each filter can be individually enabled or disabled without being deleted
**Dependencies:** none

### R2: Pattern Matching
**Description:** Patterns support both literal substring matching and regex matching (RE2 syntax). Matching is applied against the value of the specified field in each entry.
**Acceptance Criteria:**
- [ ] [auto] A literal pattern matches entries where the field value contains the pattern as a substring
- [ ] [auto] A regex pattern matches entries where the field value matches the RE2 expression
- [ ] [auto] An invalid regex is detected and reported as an inline error; the filter is not applied
- [ ] [auto] Matching is performed against `msg`, `level`, `logger`, `thread`, and any extra field key
**Dependencies:** cavekit-log-source (entry data model)

### R3: Include/Exclude Logic
**Description:** If any enabled include filters exist, only entries matching at least one include filter are shown. Entries matching any enabled exclude filter are hidden regardless of include filters. If no include filters are active, all entries are candidates (subject to excludes).
**Acceptance Criteria:**
- [ ] [auto] With one include filter for `level=ERROR`, only ERROR entries are shown
- [ ] [auto] With two include filters (`level=ERROR`, `level=WARN`), both ERROR and WARN entries are shown
- [ ] [auto] With no include filters and one exclude filter for `logger=noisy`, entries from `noisy` logger are hidden and all others are shown
- [ ] [auto] With one include filter for `level=ERROR` and one exclude filter for `msg=heartbeat`, only ERROR entries that do not contain "heartbeat" in msg are shown
- [ ] [auto] Disabled filters do not affect the result
**Dependencies:** R1, R2

### R4: Add Filter from Field Value
**Description:** A filter can be created from a specific field name and value (initiated from keyboard or mouse click on a field in the detail pane). The user is prompted to choose include or exclude mode. The value is pre-filled.
**Acceptance Criteria:**
- [ ] [auto] When a filter is added from a field value, the field name and pattern are pre-filled with the clicked field's name and value
- [ ] [auto] The user can choose include or exclude mode before the filter is committed
- [ ] [auto] After confirmation, the filter appears in the active filter set and the filtered index is updated
**Dependencies:** cavekit-detail-pane (field click interaction)

### R5: Filter Panel Overlay
**Description:** A panel overlay displays all active filters with their mode and enabled state. Within the panel, `j`/`k` navigate between filters, Space toggles the selected filter's enabled state, and `d` deletes the selected filter.
**Acceptance Criteria:**
- [ ] [auto] The filter panel lists all filters showing field, pattern, mode, and enabled state
- [ ] [auto] Pressing `j`/`k` in the panel navigates between filters
- [ ] [auto] Pressing Space toggles the enabled state of the selected filter
- [ ] [auto] Pressing `d` deletes the selected filter from the set
- [ ] [auto] Changes in the panel immediately update the filtered entry index
- [ ] [auto] The filter panel is navigable by mouse (click to select, click toggle/delete controls)
**Dependencies:** R1

### R6: Global Filter Toggle
**Description:** A single key enables or disables all filters simultaneously without deleting them.
**Acceptance Criteria:**
- [ ] [auto] Pressing the global toggle key disables all filters; the entry list shows all entries
- [ ] [auto] Pressing the global toggle key again re-enables all previously enabled filters
- [ ] [auto] Filters that were individually disabled before the global toggle remain disabled after re-enabling
**Dependencies:** R1

### R7: Filtered Entry Index
**Description:** The filter engine emits an ordered index of entries that pass all active filters. This index is consumed by the entry list.
**Acceptance Criteria:**
- [ ] [auto] The emitted index contains exactly the entries that pass all active include/exclude logic
- [ ] [auto] The index preserves the original entry order
- [ ] [auto] When filters change, the index is recomputed and re-emitted
**Dependencies:** R3, cavekit-log-source (entry data)

## Out of Scope

- Time range filtering (v2)
- Filter history/undo (v2)
- Named filter presets (v2)
- Persisting filters across sessions
- Rendering filtered entries (handled by entry-list)

## Cross-References

- See also: cavekit-log-source.md (provides entry data model)
- See also: cavekit-entry-list.md (consumes filtered entry index)
- See also: cavekit-detail-pane.md (initiates add-filter from field click)
- See also: cavekit-app-shell.md (routes keyboard/mouse to filter panel)

## Changelog
