---
created: "2026-04-18T14:40:26+03:00"
last_edited: "2026-04-18T14:40:26+03:00"
---
# Implementation Tracking: scrolloff (Tier 14 cross-cutting)

Build site: context/plans/build-site.md

| Task | Status | Notes |
|------|--------|-------|
| T-130 | DONE | See impl-config.md. Shared top-level `scrolloff` config, default 5, negative→0 with warn. |
| T-131 | DONE | See impl-detail-pane.md. `ScrollModel.cursor` + `paintCursorRow` with CursorHighlight bg. |
| T-132 | DONE | See impl-detail-pane.md. Cursor-tracking keyboard nav + `followCursor` with scrolloff margin. |
| T-133 | DONE | See impl-detail-pane.md. Mouse wheel scrolloff drag. |
| T-134 | DONE | See impl-detail-pane.md. Search `n`/`N` moves cursor with scrolloff context. |
| T-135 | DONE | See impl-entry-list.md. `ScrollState.Scrolloff` + `followCursor` + `WheelDown`/`WheelUp` drag. |
| T-136 | DONE | See context/designs/design-changelog.md (2026-04-18 audit row). DESIGN.md §4/§4.3/§4.4/§6/§9 already match impl. |
| T-137 | PENDING | HUMAN sign-off via tui-mcp across 3 themes × 2 geometries × 2 orientations on logs/small.log + logs/tiny.log line 34. Attach theme + geometry + orientation + scrolloff observations here. |

## Cross-cutting notes

**Shared key**: one top-level TOML `scrolloff` key (cavekit-config R5) consumed by both `ListModel.WithScrolloff` and `PaneModel.WithScrolloff`. Wired from `cfg.Config.Scrolloff` at `openPane` + `WindowSizeMsg` + `relayout` in `internal/ui/app/model.go`.

**Clamp formula**: effective = `min(configured, floor(viewport/2))`. Ensures cursor movement remains possible even when configured scrolloff exceeds viewport. At document edges the margin yields — cursor can always reach line 0 / last line.

**Design priority on cursor row**: `paintCursorRow` is applied AFTER `overlayScrollIndicator` in `PaneModel.View()` so:
- the NN% indicator (theme.Dim fg on last content row) keeps rendering independently (cavekit R11 AC 8)
- the CursorHighlight bg composes over SearchHighlight fg when cursor lands on a match line (T-134)

**Mouse wheel semantics**:
- In the middle of the viewport (cursor > scrolloff rows from both edges): wheel moves the viewport *under* the cursor. Cursor's document line is unchanged.
- At the scrolloff margin: cursor is dragged along, held exactly at `offset + scrolloff` (wheelDown) or `offset + viewport - 1 - scrolloff` (wheelUp).
- At document edges: margin yields, cursor reaches line 0 or last line.

This mirrors nvim `scrolloff` behaviour and matches user expectation from the original bug report.
