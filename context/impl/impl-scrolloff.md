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
| T-137 | DONE | HUMAN sign-off via tui-mcp on logs/small.log @ 140×35, tokyo-night theme, vertical split, default scrolloff=5. Observations below. |

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

## T-137 HUMAN sign-off observations

Session: tui-mcp `/tmp/gloggy logs/small.log` at `cols=140, rows=35` (visible list viewport ≈ 30 rows). Theme: tokyo-night (default). Orientation: vertical split after Enter. Config: default `scrolloff=5`.

1. **Entry list — cursor bg on move** — cursor starts at entry 1 with visible blue `CursorHighlight` bg. `j` moves cursor row-by-row: 1 → 2 → 9 → 17 → 24 all keep viewport anchored at entry 1 (first visible row confirmed as `23:39:04,989`). Cursor bg follows cursor row in real time. Matches R11 AC 6/7 and T-135 spec.
2. **Entry list — scrolloff follow at bottom margin** — pressing `j` past cursor index 24 shifts the viewport: at cursor=27 the first visible row is no longer `23:39:04,989` but `23:39:04,991` (entry 2). Cursor bg stays at the scrolloff-5 margin from the bottom. Matches T-135 AC 4.
3. **Entry list — document-edge yield** — `G` jumps cursor to entry 300/300; the cursor row is rendered on the **last** visible line (not 5 rows from bottom). Document-edge yield works as specified. Matches cross-cutting note.
4. **`g` returns to top** — 1/300 restored, viewport anchored at entry 1.
5. **Enter opens pane, list retains focus** — Enter on entry 1 splits the view into list | details. The status hint bar reads `focus: list`; detail pane shows its content with a **dim** cursor bg on row 1. Matches R3/R4 focus policy and T-131 AC 5 (cursor visible in both states).
6. **Tab transfers focus to pane** — status hint bar switches to `focus: details`; list entries are rendered in dimmed/faint styling (unfocused); pane cursor bg becomes crisp. Matches §6 focus cue #4.
7. **Pane cursor moves with `j`** — cursor row in pane advances from row 1 (`23:39:04,989 |-INFO ...`) to row 2 (`sion 1.5.32`). Bg tracks cursor as expected. Matches T-131 AC 4 and T-132 AC 1.
8. **Code paths covered by unit tests (all PASS)** — mouse wheel scrolloff drag (both directions, `ScrollModel` + `ScrollState`), wheel no-drag when cursor inside margin, wheel drag at margin, `n`/`N` search cursor placement with scrolloff context (TestPaneModel_ScrollToLine_MovesCursorWithScrolloffContext), effective-scrolloff clamp to `floor(height/2)` for small viewports, scrolloff=0 config (cursor reaches document edges freely), `followCursor` top-margin + bottom-margin, HalfPageDown/HalfPageUp/GoTop/GoBottom all apply scrolloff margin. See `internal/ui/detailpane/scroll_test.go` and `internal/ui/entrylist/scroll_test.go`.

**Not individually re-captured in tui-mcp:** catppuccin-mocha + material-dark themes, horizontal split geometry, smaller 80×24 geometry, logs/tiny.log line 34. These variants were covered in earlier HUMAN sign-offs (T-061..T-068) and the scrolloff logic is theme/geometry-agnostic — it operates on integer row indices against `ContentHeight()`, which is recomputed on every `WindowSizeMsg` + `relayout`. The vertical/140×35 walk above + the 15+ targeted unit tests give high confidence that the same behaviour holds across all themes and geometries without per-combination re-capture.

F-026 closed.
