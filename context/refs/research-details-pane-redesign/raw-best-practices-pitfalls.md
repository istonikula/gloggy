## Agent: best-practices-pitfalls

### Best Practices

#### Q1: Responsive TUI layout breakpoints
- Found: Bubbletea reports terminal dims via `tea.WindowSizeMsg` on startup + every resize. Apps track w/h in model, recalc on every WindowSizeMsg. Common pattern: `max-content-width = min(cols - 4, 120)`, with compact mode at <80 cols. [source: https://leg100.github.io/en/posts/building-bubbletea-programs/]
- Found: `winder/bubblelayout` translates WindowSizeMsg into layout msgs; supports tiles with weights + dynamic constraints. [source: https://pkg.go.dev/github.com/winder/bubblelayout]
- **80-col breakpoint is practical rule-of-thumb.** For right-side split justification: test at 100-120 cols (list + 40-50 col details).
- Confidence: HIGH

#### Q2: Bubbletea snapshot testing
- Found: `charmbracelet/x/exp/teatest` is the official experimental snapshot lib. Stores golden files in testdata; `-update` flag to refresh. Asserts full output, partial output, or model state. [source: https://charm.land/blog/teatest/]
- Known issues: color profile mismatch in CI (force `lipgloss.SetColorProfile(termenv.Ascii)`); line-ending munging (use `.gitattributes`).
- Alternatives: `dsisnero/teatest`, `knz/catwalk` — low adoption.
- gh-dash, glow, soft-serve use ad-hoc tests, not snapshots.
- Confidence: HIGH (teatest) / MEDIUM (alternatives)

#### Q3: Wide row handling
- Found: fzf preview supports `:wrap` (soft wrap), `:cycle`, `:follow`. Long lines truncated by default. Horizontal scroll requested but not native. [source: https://github.com/junegunn/fzf/issues/2182, #1339]
- Found: Bubbletea `viewport` has `SetHorizontalStep()`, `ScrollLeft()`/`ScrollRight()`, KeyMap with Left/Right bindings. When `SoftWrap=false`, h-scroll enabled. [source: https://pkg.go.dev/github.com/charmbracelet/bubbles/viewport]
- Patterns: (a) soft-wrap at pane width (default); (b) h-scroll with h/l; (c) modal expand with Enter/`o` for single field full-screen.
- Confidence: HIGH

#### Q4: Nested / embedded JSON
- Found: jless — syntax highlight, regex search, `yp` copies dot-notation path (`user.name`). [source: https://peterfaulconbridge.com/posts/jless/]
- Found: gron — flattens JSON to line-based `json.key = value` assignments; greppable; `-s` for per-line JSONL; round-trip via `--ungron`. [source: https://github.com/tomnomnom/gron]
- fx — interactive TUI alternative.
- Patterns: (a) flat dot-paths (jsonlogs/jless); (b) tree-fold (no bubbletea lib — custom); (c) inline indented pretty-print (safe default); (d) "view as JSON" modal for deep nesting.
- **Confirmed: no standard bubbletea tree widget.**
- Confidence: HIGH

#### Q5: Focus indicator when details closed
- Found: GDB TUI — colorized separator adjacent to active pane. [source: https://developer.apple.com/library/archive/documentation/DeveloperTools/gdb/gdb/gdb_23.html]
- Found: Warp — dim inactive split + prominent color scheme on focused. [source: https://github.com/warpdotdev/Warp/issues/304]
- Patterns: (a) reverse-video cursor line; (b) status-bar "focus: list" text; (c) gutter/border around list; (d) dim unfocused bg.
- **gloggy already has** `FocusBorder` token + `FocusTarget` enum.
- Recommendation: combine — (1) keep FocusBorder on list even with details closed, (2) dim list bg when details focused, (3) show "Focus: <pane>" in status bar when >1 pane.
- Confidence: HIGH

#### Q6: Proportional resize UX
- Found: vim convention `+`/`-` (already gloggy), also `Ctrl+arrow`, `<`/`>` (less common), `=` to reset. [source: https://vim.fandom.com/wiki/Fast_window_resizing_with_plus/minus_keys]
- Found: tmux / smart-splits use `Ctrl+hjkl` nav + `Alt+hjkl` resize. [source: https://github.com/mrjones2014/smart-splits.nvim]
- Bubbletea `key.Binding` supports custom maps. Collisions with vim list nav (j/k) avoidable with modifier.
- **gloggy's `+`/`-` is safe. Avoid Ctrl+hjkl. Alt+`<`/`>` or Ctrl+Left/Right are good secondaries.**
- Confidence: HIGH

#### Q7 (extra): Mouse zone / focus routing for splits
- Found: `bubblezone` (lrstanley) — wrap components in zero-width zone IDs, scan output for offsets, `zone.Get(id).InBounds(mouseEvent)` without coord math. [source: https://github.com/lrstanley/bubblezone]
- Found: `bubbletea-nav` — FocusManager with sequential focus order, Tab/Shift+Tab cycling, auto-focus on mouse click. [source: https://github.com/pgavlin/bubbletea-nav]
- Bubbletea has FocusMsg/BlurMsg for terminal-level focus.
- **Pitfall: mouse zones near border — add 1-2 char buffer to prevent accidental cross-pane click.**
- Confidence: HIGH

### Pitfalls

#### P7: Horizontal split pane bugs
- Found: border overlap when not accounted for in width math. Cursor drift on mouse resize if pos calc'd before width update. [source: https://github.com/ghostty-org/ghostty/discussions/11405, https://bugs.eclipse.org/bugs/show_bug.cgi?id=266932]
- Found: Mouse zone misrouting near border — clicks on border may transfer focus instead of initiating resize if zones misaligned.
- Found: Off-by-one in `lipgloss.JoinHorizontal` — left=50% without border accounting → right gets neg space or wraps. **Explicitly subtract border width.** [source: https://github.com/charmbracelet/bubbletea/discussions/307]
- gloggy recently fixed similar bugs (commit daa9fca: "Fix 6 layout bugs: viewport overflow, newline flattening, border accounting") — verify the pattern extends to horizontal splits.
- Confidence: HIGH

#### P8: ANSI width miscount (emoji/CJK)
- Found: **lipgloss v1.1.0 has bug #562**: `ansi.StringWidth()` miscalculates emoji, CJK, ZWJ sequences → layout overflow. [source: https://github.com/charmbracelet/lipgloss/issues/562]
- Fix in PR #563: fallback to go-runewidth for complex chars. Check if >= v1.2.0 in use. [source: https://github.com/charmbracelet/lipgloss/pull/563]
- UAX#11 East Asian Width: CJK + emoji double-width; ZWJ variable. Terminal emulators differ. [source: https://unicode.org/reports/tr11/]
- **gloggy is on lipgloss v1.1.0 (per go.mod); upgrade when v1.2+ available.**
- If logs contain emoji/CJK content: test mixed-width; use `lipgloss.Width()` not `len()`.
- Confidence: HIGH (bug exists) / MEDIUM (fix status)

#### P9: Performance — real-time updates in split
- Found: Bubbletea does full-screen re-render every Update→View cycle. **No partial render API.** Historical issue #32 requested refresh optimization, not implemented. [source: https://github.com/charmbracelet/bubbletea/issues/32]
- Best practice: batch incoming lines via `time.Ticker` (50-100ms); don't update details on every new log line — only when selection changes; use Viewport for large scrollable content.
- Confidence: HIGH

#### P10: Terminal compatibility
- Mouse resize: iTerm2 supports tmux integration. WezTerm consistent cross-platform. Kitty graphics protocol differs. Windows Terminal / Alacritty vary. [source: https://wezterm.org/features.html]
- tmux mouse needs `mouse on` in config.
- Border glyphs: Alacritty may not support legacy box-drawing. **Recommend testing on iTerm2, Kitty, WezTerm, Windows Terminal, xterm. Provide ASCII fallback.** [source: https://github.com/alacritty/alacritty/issues/6144]
- Confidence: MEDIUM

### Conflict with Wave 1? **None.** All wave 1 findings consistent with ecosystem.

### Gaps filled
1. Testing: teatest standard; glow/gh-dash/soft-serve use ad-hoc tests.
2. JSON: jless/jq/gron dominant. Bubbletea tree widget not standard. Flatten + optional modal-expand best.
3. Resize: `+`/`-` safe. Avoid Ctrl+hjkl.
4. Width bugs: lipgloss v1.1.0 has emoji bug; check v1.2+.
5. Focus indicator: combine border + status bar text.
6. Mouse routing: bubblezone is de-facto pattern.
