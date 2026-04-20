# appshell

Implements:
- cavekit-app-shell.md R1 (Entry Points and CLI)
- cavekit-app-shell.md R2 (Layout)
- cavekit-app-shell.md R3 (Header Bar)
- cavekit-app-shell.md R4 (Context-Sensitive Key-Hint Bar)
- cavekit-app-shell.md R5 (Help Overlay)
- cavekit-app-shell.md R6 (Mouse Mode and Routing)
- cavekit-app-shell.md R7 (Terminal Resize)
- cavekit-app-shell.md R8 (Loading Indicator)
- cavekit-app-shell.md R9 (Clipboard)
- cavekit-app-shell.md R10 (Focus Indicator)
- cavekit-app-shell.md R11 (Focus Cycle)
- cavekit-app-shell.md R12 (Pane Resize Controls)
- cavekit-app-shell.md R13 (Cross-Pane Search Activation)

Build tasks: T-014, T-046, T-047, T-048, T-050, T-051, T-052, T-053, T-054, T-072, T-081, T-083, T-087, T-088, T-089, T-090, T-091, T-092, T-093, T-094, T-095, T-096, T-097, T-098, T-099, T-100, T-104, T-105, T-108, T-116, T-121, T-121-fix, T-123, T-126, T-138, T-142, T-144, T-144-fix, T-145, T-172, T-173, T-174, T-179 (build-site.md)

Pane background paint (T-179): `PaneStyle(th, state)` applies `theme.BaseBg` on focused panes and `theme.UnfocusedBg` on unfocused panes (with `BaseBg` as defensive fallback). No pane falls through to the terminal's default background — this closes cavekit-config.md R4 AC 13. Drag-seam token (T-172/T-173/T-174): `RenderDivider` + `WithDragSeamTop` paint `theme.DragHandle` on the 1-cell seam shared between the list and detail pane; focus-neutral per app-shell R15.
