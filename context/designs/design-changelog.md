# Design System Changelog

Append-only log of DESIGN.md changes.

| Date       | Section | Change                                      | Source       |
|------------|---------|---------------------------------------------|--------------|
| 2026-04-17 | All     | Initial focused design system — details-pane + layout-shell + theme-tokens. 9/9 sections. | `/ck:design` |
| 2026-04-18 | §4.4, §6, §9 | Added scroll-position feedback (`NN%` overlay semantics) to §4.4 Detail Pane; documented open-time focus policy (opening pane does not transfer focus) + Esc-from-list-closes-pane rule in §6 Focus model; extended §9 keymap matrix with vim nav keys (`g`/`G`/`Home`/`End`/`PgDn`/`PgUp`/`Ctrl+d`/`Ctrl+u`/`Space`/`b`) under Detail pane context. Closes F-021 (P2). | T-128 |
| 2026-04-18 | §4 matrix, §4.3, §4.4, §6, §9 | Redefined cursor-row scope in §4 matrix from "(list only)" to list AND detail pane. New §4.3 "Shared scrolloff" subsection documenting the top-level `scrolloff` config (default 5) consumed by both entry list and detail pane. New §4.4 "Cursor and scrolloff" subsection defining detail pane as cursor-tracking viewport with nvim-style scrolloff drag on mouse wheel; search `n`/`N` moves the cursor to the match line with scrolloff-respected context; cursor row bg priority over search highlight on the active row. §6 focus cue #4 rewritten to cover both panes. §9 keymap `j`/`k` row rewritten as "move cursor; viewport follows with scrolloff margin" + new Mouse wheel row. Closes F-026 (P1). Driven by user report "I still see no row highlight where cursor is when focused on details pane" + follow-ups on list scrolloff + nvim scrolloff drag semantics. | /ck:check 2026-04-18 (pre-T-131..T-137) |
