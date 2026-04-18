---
created: "2026-04-15T00:00:00Z"
last_edited: "2026-04-18T09:40:17+03:00"
---

# Cavekit Overview: gloggy

A terminal UI tool for interactively analyzing JSONL log files during local development. Single binary, single file input only.

## Domain Index

| Domain         | File                        | Reqs | Status | Description                                              |
|----------------|-----------------------------|------|--------|----------------------------------------------------------|
| log-source     | cavekit-log-source.md       | 8    | DRAFT  | File/stdin reading, line classification, JSONL parsing   |
| entry-list     | cavekit-entry-list.md       | 11   | DRAFT  | Compact scrollable list with two-level cursor and marks  |
| detail-pane    | cavekit-detail-pane.md      | 10   | DRAFT  | Pretty-printed JSON detail view with in-pane search      |
| filter-engine  | cavekit-filter-engine.md    | 7    | DRAFT  | Include/exclude filter model and filter panel overlay    |
| config         | cavekit-config.md           | 7    | DRAFT  | TOML config with themes, field visibility, live writes   |
| app-shell      | cavekit-app-shell.md        | 13   | DRAFT  | Top-level layout, wiring, clipboard, help overlay        |

## Cross-Reference Map

| Domain A       | Interacts With   | Interaction Type                                         |
|----------------|------------------|----------------------------------------------------------|
| log-source     | entry-list       | Emits parsed entries consumed by entry-list              |
| log-source     | app-shell        | Reports loading progress and tail status                 |
| entry-list     | detail-pane      | Emits "entry selected" signal consumed by detail-pane    |
| entry-list     | filter-engine    | Receives filtered entry index, displays only passing entries |
| entry-list     | config           | Reads field visibility, sub-row fields, logger depth, theme |
| detail-pane    | filter-engine    | Mouse click on field value triggers filter creation       |
| detail-pane    | config           | Reads/writes field visibility, reads theme and pane height |
| detail-pane    | app-shell        | Focus cycle (Tab), shared resize controls (R12)           |
| filter-engine  | entry-list       | Emits filtered entry index                               |
| filter-engine  | detail-pane      | Receives filter-add requests from field clicks           |
| config         | entry-list       | Supplies theme, field visibility, sub-row fields, logger depth |
| config         | detail-pane      | Supplies theme, field visibility, pane height             |
| config         | app-shell        | Supplies theme                                            |
| app-shell      | all domains      | Initializes, wires, routes input, manages layout         |

## Dependency Graph

Implementation order (domains that must be available before dependents):

```
config          (no dependencies — needed by all UI domains)
  |
log-source      (depends on: none, but config supplies theme indirectly)
  |
filter-engine   (depends on: log-source entry model)
  |
entry-list      (depends on: log-source, filter-engine, config)
  |
detail-pane     (depends on: entry-list selection, filter-engine, config)
  |
app-shell       (depends on: all above)
```

Parallelizable: `config` and `log-source` can be built concurrently. `filter-engine` needs only the entry data model from `log-source`. `entry-list` and `detail-pane` can be built concurrently once their dependencies exist.

Note: right-split orientation introduces a vertical divider and horizontal mouse zones; see cavekit-app-shell R2, R6, R10, R11, R12.

## Coverage Summary

- **Total domains:** 6
- **Total requirements:** 56 (2026-04-18: +1 new app-shell R13)
- **Total acceptance criteria:** 290 (2026-04-18: +7 R7 + +7 R13; was 276)

## Verification Conventions

Applies to every kit in this project. Per-domain kits may tighten, never loosen.

### HUMAN sign-off (any AC tagged `[human]` or task prefixed `[HUMAN]`)

1. Verification is performed via **tui-mcp** — live TUI inspection using the `mcp__tui-mcp__*` tools (`launch`, `screenshot`, `snapshot`, `send_keys`, `wait_for_text`). Visual judgement from the raw terminal without capture artifacts is not sufficient evidence.
2. Fixture: use one of the bundled logs under `logs/` (`tiny.log` / `small.log` / `medium.log` / `big.log`), picking the smallest fixture that exercises the criterion. Pathological sizes (`big.log`) only when the AC itself concerns scale.
3. Terminal geometry: verify at **both** a representative small (e.g. `80x24`) and a representative large (e.g. `140x35` or `198x48`) size, and — for any AC that mentions orientation or layout — at both `right` and `below` pane orientations.
4. Theme coverage: any AC that mentions color, readability, or perceptual clarity must be verified against **all three bundled themes** (`tokyo-night`, `catppuccin-mocha`, `material-dark`).
5. Evidence in impl-tracking: the task's row in `context/impl/impl-*.md` must record, in the `Notes` column:
   - the capture method (`HUMAN sign-off via tui-mcp`),
   - the fixture (`on small.log`, etc.),
   - the geometry / orientations / themes exercised,
   - the specific visual observation that satisfied the criterion (color codes, glyph positions, border behaviour — whatever the AC asks for).
6. If tui-mcp cannot reproduce the condition (e.g. hardware-only concerns such as real clipboard integration), record the AC as `[HUMAN] — out-of-band` and document the alternative method inline. This is an escape hatch, not the norm.

Automated ACs (tagged `[auto]`) are verified by Go unit and integration tests and do not require tui-mcp.

