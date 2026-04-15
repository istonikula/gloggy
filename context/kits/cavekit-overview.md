---
created: "2026-04-15T00:00:00Z"
last_edited: "2026-04-15T00:00:00Z"
---

# Cavekit Overview: gloggy

A terminal UI tool for interactively analyzing JSONL log files during local development. Single binary, single file input only.

## Domain Index

| Domain         | File                        | Reqs | Status | Description                                              |
|----------------|-----------------------------|------|--------|----------------------------------------------------------|
| log-source     | cavekit-log-source.md       | 8    | DRAFT  | File/stdin reading, line classification, JSONL parsing   |
| entry-list     | cavekit-entry-list.md       | 10   | DRAFT  | Compact scrollable list with two-level cursor and marks  |
| detail-pane    | cavekit-detail-pane.md      | 8    | DRAFT  | Pretty-printed JSON detail view with in-pane search      |
| filter-engine  | cavekit-filter-engine.md    | 7    | DRAFT  | Include/exclude filter model and filter panel overlay    |
| config         | cavekit-config.md           | 7    | DRAFT  | TOML config with themes, field visibility, live writes   |
| app-shell      | cavekit-app-shell.md        | 9    | DRAFT  | Top-level layout, wiring, clipboard, help overlay        |

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

## Coverage Summary

- **Total domains:** 6
- **Total requirements:** 49
- **Total acceptance criteria:** 210
