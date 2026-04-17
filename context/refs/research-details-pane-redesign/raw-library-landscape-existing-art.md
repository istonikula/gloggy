## Agent: library-landscape-existing-art

> NOTE: briefing said ratatui. gloggy is actually Go + bubbletea/lipgloss. Much ratatui material below still relevant by analogy; bubbletea section added in synthesis.

### Library Landscape (ratatui — analogous patterns for bubbletea)

#### Layout primitives
- Ratatui: `Layout::default().direction()` with `Constraint::Percentage|Min|Ratio|Length|Max`. v0.30 adds `Spacing::Overlap` for border collapsing, `WidgetRef` trait. [https://ratatui.rs/concepts/layout/]
- No higher-level "split pane" crate exists; apps compose manually. [https://ratatui.rs/recipes/layout/grid/]
- Bubbletea analog: `lipgloss.JoinVertical`/`JoinHorizontal` + width/height arithmetic. Already in gloggy.
- Confidence: HIGH

#### Tree widgets
- `EdJoPaTo/tui-rs-tree-widget` — 123⭐, v0.24.0 Jan 2026, actively maintained, generic tree. [https://github.com/EdJoPaTo/tui-rs-tree-widget]
- `woxjro/ratatui-tree-widget` — 3⭐ fork. Insignificant adoption.
- No JSON-specific tree crate in ratatui ecosystem. App owns flatten logic.
- Bubbletea analog: **no dominant tree widget.** Community options: huh/form (input only), bubbles/list (flat only). For JSON tree with fold, implement custom or port tree-widget pattern.
- Confidence: HIGH

#### Focus indication
- Ratatui core: `Frame::set_cursor_position`, `Block::border_style`. No focus manager.
- `ratatui-interact` (Brainwires) — FocusManager<T>, Tab/Shift+Tab, interactive widgets.
- `rat-focus` (rat-salsa) — alternative focus abstraction.
- Pattern: Block with `border_style(Style)` flips color/bold/underline on focused pane.
- Bubbletea: no focus manager. Apps roll own (gloggy has `FocusTarget` enum already).
- Confidence: HIGH

#### Pane resize in mature apps
- **GitUI** (21.8k⭐, ratatui): master-detail (files | diff). No drag-resize; layout adapts to terminal. [https://github.com/gitui-org/gitui]
- **Yazi** (ratatui file manager): three-column (parent | current | preview), configurable `ratio = [1,4,3]`. No interactive drag. Plugin `toggle-pane.yazi` provides maximize. [https://github.com/sxyazi/yazi, https://github.com/yazi-rs/plugins/tree/main/toggle-pane.yazi]
- **Zellij** (multiplexer, non-ratatui): supports drag-resize, Ctrl+scroll resize, Alt+f floating maximize. Reference UX; complex to implement. [https://zellij.dev/documentation/options.html]
- **Lazygit** (Go/gocui): multi-pane status/files/branches/commits. Adaptive on terminal resize; granular drag not documented.
- Takeaway: **mature apps prefer maximize-one-pane toggle + preset ratios over true drag-resize.** Drag is nice-to-have.
- Confidence: HIGH

### Existing Art

#### jsonlogs.nvim (reference)
- **Layout:** configurable right-side vertical split, default 80 chars wide.
- **Keymaps:** Tab to switch panes; "f" to maximize/restore preview pane to full width.
- **Wide row handling:** table mode max column width 30 chars (configurable); cells truncated with '...'. Enter on cell opens full content view (new buffer/window).
- **Nested JSON:** flattens structures (`user.name` → `user.name` column, `tags[0]` → `tags[0]` column).
- **Focus jumps:** Tab cycles panes; preview auto-syncs with source panel selection.
- **Resize control:** "f" toggle maximize; `layout.width`/`layout.height` config for persistent sizing.
- Adoption: 2⭐ niche. [https://github.com/ll931217/jsonlogs.nvim]
- Confidence: HIGH

#### Log viewer TUIs
- **lnav** (ncurses C): mature, legacy, no modern split-pane detail. Less relevant.
- **gonzo** (Go TUI): live stream + color coding. Enter opens full-screen detail view (not split).
- **kl** (Rust K8s log TUI): multi-cluster filter; no split master-detail documented.
- Takeaway: log viewers traditionally use single-view or full-screen detail, not split. jsonlogs.nvim's split-pane approach is relatively novel.
- Confidence: MEDIUM

#### Rust master-detail references
- **GitUI**: left file list | right read-only diff. Arrow-nav in list, diff auto-updates. No drag-resize.
- **Yazi**: parent | current | preview. Toggle maximize via plugin.
- **Atuin**: full-screen, not master-detail. Less relevant.
- Confidence: HIGH

### Summary Table
| Component | Best option | Notes |
|-----------|-------------|-------|
| Layout primitives | Built-in (ratatui `Layout`+`Constraint` / lipgloss `JoinVertical`/`JoinHorizontal`) | No higher-level abstraction needed |
| Tree widget (nested JSON fold) | EdJoPaTo/tui-rs-tree-widget (ratatui only) | No bubbletea equivalent; build custom if needed |
| JSON flattening | Custom (jsonlogs.nvim pattern) | No off-the-shelf crate in either ecosystem |
| Focus mgmt | ratatui-interact or custom | gloggy already has `FocusTarget` enum |
| Focus indication | Border style change on focused block | Already implemented in gloggy |
| Master-detail pattern | GitUI / Yazi (static layout, terminal-adaptive) | Drag-resize optional; maximize toggle common |
| Maximize toggle | Custom state + `f` or Tab | jsonlogs.nvim pattern |

### Key Insights
1. Layout is constraint-based at draw time, not event-driven. Drag-resize requires state + event handling.
2. Tree widgets are generic; JSON flatten is application concern.
3. Focus mgmt not built-in; apps roll own with border-style feedback.
4. Mature apps avoid granular drag-resize; prefer maximize toggle + preset ratios.
5. Split-pane defaults vary: right-side vertical (jsonlogs), three-col horizontal (Yazi), multi-pane (GitUI). Choose based on UX need.

### Sources
- https://ratatui.rs/concepts/layout/
- https://docs.rs/ratatui/latest/ratatui/layout/enum.Constraint.html
- https://github.com/EdJoPaTo/tui-rs-tree-widget
- https://github.com/Brainwires/ratatui-interact
- https://github.com/gitui-org/gitui
- https://github.com/sxyazi/yazi
- https://github.com/ll931217/jsonlogs.nvim
- https://zellij.dev/documentation/options.html
- https://github.com/jesseduffield/lazygit
- https://ratatui.rs/highlights/v030/
