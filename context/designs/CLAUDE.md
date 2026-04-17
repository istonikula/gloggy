# Design System

The project's visual design system in DESIGN.md format (9-section Google Stitch,
TUI-adapted).

## Conventions

- `DESIGN.md` at project root is the canonical source.
- All UI implementation must reference DESIGN.md tokens and patterns — never
  hardcode colors, border styles, focus logic, or keybindings in component
  files.
- Color token hex values live in `internal/theme/theme.go` (three bundled
  themes). DESIGN.md §2 references this file by design; do not duplicate hex
  values in DESIGN.md.
- Updated via `/ck:design` (full pass or `--section N`) or automatically
  surfaced during `/ck:check` and `/ck:revise`.
- Agents must read DESIGN.md before implementing any user-facing component.

## Scope note

This DESIGN.md is **focused** (per the /ck:design Step 3 scope decision):
details-pane + layout-shell + theme-tokens. It is the authoritative reference
for the right-side-split pane redesign and anything that touches pane focus,
borders, ratios, or the keymap. It intentionally does not attempt to spec every
UI surface.

## Related

- `context/refs/research-brief-details-pane-redesign.md` — research brief that
  drove this DESIGN.md's scope and recommendations.
- `context/kits/cavekit-detail-pane.md` — the domain kit that will be revised to
  reference DESIGN.md sections.
- `context/kits/cavekit-app-shell.md` — the layout-shell kit that will be
  revised for orientation logic.
