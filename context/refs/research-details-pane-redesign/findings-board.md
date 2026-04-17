# Findings Board: Details pane redesign — right-side layout, jsonlogs.nvim-inspired UX

> IMPORTANT for downstream agents:
> gloggy is **Go + charmbracelet/bubbletea + lipgloss**, NOT Rust+ratatui.
> Ratatui findings below are analogous patterns only — use equivalent bubbletea concepts (lipgloss.JoinVertical/JoinHorizontal, no built-in focus manager, tea.Msg dispatch).

## Wave 1 — Codebase + Libraries + Existing Art

### Architecture (codebase)
- Go + bubbletea v1.3.10. Entry: `cmd/gloggy/main.go`. Root: `internal/ui/app/model.go` `Model` struct holds 12 subsystems.
- Current detail pane is vertical-stacked below list (`lipgloss.JoinVertical`). Implemented in `internal/ui/detailpane/`: PaneModel (model.go), ScrollModel (scroll.go), HeightModel (height.go 0.30 default ratio, +/- adjust), SearchModel, VisibilityModel, fieldclick.go.
- Layout: `header(1) | entryList | detailPane(if open) | statusBar(1)` via `appshell/layout.go:96-103`. `EntryListHeight() = Height - (Header + StatusBar + DetailPaneHeight)`.
- **No horizontal split exists.** Moving to right-side requires `JoinHorizontal` branch + Width tracking alongside existing Height.

### Current list rendering
- Virtual rendering (only visible rows), one-line-per-entry, **truncation not wrap**. `entrylist/row.go:36-90` — `msg[:remaining]`. Newlines flattened to space.
- JSON/non-JSON mixed. Non-JSON in dim color.
- Perf: <16ms for 100k entries.

### State mgmt
- Elm-architecture. `Update(tea.Msg) (tea.Model, tea.Cmd)`. Root type-switch dispatch (app/model.go:108-209).
- Focus: `FocusTarget` enum = `FocusEntryList | FocusDetailPane | FocusFilterPanel`. Centralized key-routing on focus (lines 211-280).
- Mouse: `MouseRouter` maps (x,y) to zones (`appshell/mouse.go:19-85`). **Zones assume vertical stack — needs rewrite for horizontal split.**

### Theme/focus indicators (already implemented)
- Theme tokens: `theme/theme.go:12-37` — includes `FocusBorder`, `HeaderBg`, `CursorHighlight`.
- Current focus indicator: colored left border on focused pane (`detailpane/model.go:122-124` `BorderLeft(m.Focused)` + `FocusBorder` color). Flag `Focused` set pre-render at app/model.go:311.

### Tests
- Stdlib `testing`. **No snapshot framework.** 44 `_test.go`. Unit + integration (`tests/integration/smoke_test.go:23-135`).
- Gap: no golden-file layout tests, no mouse zone tests, no visual regression.

### Context dir
- Kits: `context/kits/cavekit-detail-pane.md` (8 reqs, lines 14-94), overview with 6 domains.
- Impl tracking: `context/impl/impl-detail-pane.md` etc.
- Build site: 83 tasks in 9 tiers.
- Implication: redesign = revise `cavekit-detail-pane.md` (or new `cavekit-right-pane.md`) + new `impl-right-pane.md`; follow cavekit methodology.

### jsonlogs.nvim — reference behavior (observed + docs)
- Layout: right-side vertical split, default 80 cols wide. At 80x40 portrait it uses ~50/50. At 200x50 landscape still ~50/50 (fixed ratio, not responsive width).
- Left pane (source): raw JSONL with line numbers, truncated at border (no wrap).
- Right pane (preview): YAML-style flat `key: value` per property, one per line.
- Nested JSON: flattened to `user.name`, `tags[0]` column format.
- Keymaps (from docs): Tab cycle panes; `f` maximize/restore preview; Enter on cell opens full-content view (new buffer).
- Wide row: table mode max column width 30, cells truncated with `...`; Enter opens full content.
- Live sync: moving cursor in source pane updates preview automatically.
- Adoption: 2⭐ niche.

### Library landscape (analogy — bubbletea has limited equivalents)
- Ratatui: constraint-based `Layout::default()` with `Percentage|Min|Ratio|Length|Max`. v0.30 adds `Spacing::Overlap`.
- Bubbletea/lipgloss: `JoinVertical`/`JoinHorizontal` + manual width/height arithmetic. Already in gloggy.
- Tree widgets: ratatui has `EdJoPaTo/tui-rs-tree-widget` (123⭐). **Bubbletea has no dominant tree widget** — would need custom for folded JSON.
- Focus mgmt: ratatui-interact (Brainwires) provides FocusManager. Bubbletea apps roll own — gloggy already has `FocusTarget` enum.

### Master-detail references
- GitUI (21.8k⭐ ratatui): files | diff. No drag-resize.
- Yazi (ratatui file mgr): parent | current | preview, configurable `ratio=[1,4,3]`. Plugin `toggle-pane.yazi` for maximize.
- Lazygit (Go/gocui): multi-pane, no granular drag-resize.
- Zellij: drag-resize + Ctrl+scroll + Alt+f maximize. Complex; reference UX only.
- **Convergent pattern: maximize-one-pane toggle + preset ratios > true drag-resize.** Drag-resize rare in TUIs.

## Wave 2 — Best Practices + Pitfalls

### Breakpoint rule-of-thumb
- Practical breakpoint **80 cols** for compact mode; **100-120 cols** minimum for right-side split (list + 40-50 col details).
- Track `tea.WindowSizeMsg`, recalc on every event.
- `winder/bubblelayout` available for tile-based layouts; likely overkill for gloggy's current needs.

### Bubbletea testing
- **`charmbracelet/x/exp/teatest` is the official snapshot lib.** Golden files in testdata; `-update` refresh flag.
- CI gotchas: force ASCII color profile; `.gitattributes` for line endings.
- Major bubbletea apps (glow, gh-dash, soft-serve) don't use snapshots — we'd be breaking new ground but with blessed tooling.

### Wide row handling
- Three patterns: (a) soft-wrap at pane width (default, fzf), (b) h-scroll with h/l (viewport `SetHorizontalStep`), (c) modal-expand with Enter/`o` for full-screen single-field view.
- Bubbletea `viewport` supports h-scroll natively when `SoftWrap=false`.

### Nested JSON
- Patterns: (a) flat dot-paths (jsonlogs/jless/gron), (b) tree-fold (no bubbletea lib — custom build), (c) indented pretty-print (current gloggy default), (d) modal for deep nesting.
- **No standard bubbletea tree widget confirmed.** Custom fold state if desired.
- gron is useful prior-art for JSONL flattening pattern.

### Focus indicator (closed details)
- Combine: (1) keep `FocusBorder` on list even when alone; (2) dim unfocused bg when details open; (3) status bar text "Focus: list".
- Warp, GDB, tmux all use dim-inactive + highlight-focused combinations.

### Resize keymap
- `+`/`-` (gloggy current) safe, vim convention.
- Avoid Ctrl+hjkl — conflicts with vim nav.
- Good secondaries: Alt+`<`/`>`, Ctrl+Left/Right, `=` to reset.

### Mouse routing for splits
- `lrstanley/bubblezone` is the de-facto pattern — zero-width zone IDs, output scan for offsets, `zone.Get(id).InBounds(ev)`.
- gloggy has custom mouse router already; consider adopting bubblezone for split layout — or add 1-2 char buffer near border to prevent accidental cross-pane click.

### Pitfalls to defend against
1. **Border off-by-one** — `lipgloss.JoinHorizontal` with naive 50/50 leaves no room for border. Subtract border width explicitly. gloggy recently fixed similar bugs (commit daa9fca).
2. **lipgloss v1.1.0 emoji/CJK width bug** (issue #562). Upgrade to v1.2+ once released. Use `lipgloss.Width()` not `len()`.
3. **Full re-render per Update** — no partial render in bubbletea. Batch log-add updates via 50-100ms Ticker; don't re-render details on every log line.
4. **Terminal compatibility** — test on iTerm2, Kitty, WezTerm, Windows Terminal, xterm. ASCII border fallback for Alacritty edge cases.
5. **Mouse zone near border** — 1-2 char buffer prevents focus transfer when aiming at resize divider.

### Gloggy-specific observations
- gloggy's `+`/`-` + FocusTarget enum + colored FocusBorder = already aligned with recommended patterns.
- Missing: snapshot tests (teatest), bubblezone, upgraded lipgloss.
- Recent commit daa9fca suggests team already knows layout-pitfall territory.
