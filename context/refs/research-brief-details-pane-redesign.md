# Research Brief: Details Pane Redesign — Right-Side Split with jsonlogs.nvim UX

**Generated:** 2026-04-17
**Agents:** 2 codebase, 2 web, 1 direct TUI inspection
**Sources consulted:** 22 unique URLs/repos (see Sources)

## Summary

Moving gloggy's details pane from a below-list vertical stack to a right-side split is well-supported by the existing architecture: `lipgloss` already ships `JoinHorizontal`, the `FocusTarget` enum and mouse-zone router are centralized, and the theme has `FocusBorder`/`HeaderBg`/`CursorHighlight` tokens. The main work is structural (add width tracking + orientation branch + horizontal mouse zones) plus six UX decisions that research now makes concrete: breakpoint at ~100 cols, soft-wrap as default for wide rows, flat dot-path rendering for nested JSON (defer drill-down modal), Tab to cycle focus (not close — jsonlogs.nvim's Tab-closes is a footgun), three-preset maximize cycle, and a combined cursor-line-highlight + focus-border cue when details is closed. Key risks are small but real: `lipgloss` v1.1.0 has an emoji/CJK width bug (#562), bubbletea has no partial render so log-add bursts need Ticker batching, and naive 50/50 `JoinHorizontal` drops border width on the floor (gloggy already knows this category — see commit daa9fca). No new dependencies are needed; adopting `teatest` for golden-file layout tests is recommended.

## Key Findings

### Current Gloggy Architecture (codebase)
- **Framework:** Go + `charmbracelet/bubbletea v1.3.10` + `lipgloss v1.1.0` + termenv; `fsnotify v1.9.0` for tail; `pelletier/go-toml` for config; `atotto/clipboard`. Entry: `cmd/gloggy/main.go` → `internal/ui/app/model.go` (root Model).
- **Details pane package:** `internal/ui/detailpane/` — `PaneModel` (model.go:17-26), `ScrollModel` (scroll.go), `HeightModel` (0.30 default ratio, `+`/`-` adjust; height.go), `SearchModel`, `VisibilityModel` (global hidden-fields persisted to config), `fieldclick.go`.
- **Layout:** `appshell/layout.go:96-102` composes `header(1) | entryList | detailPane(if open) | statusBar(1)` via `lipgloss.JoinVertical` exclusively. `EntryListHeight = Height - Header - StatusBar - DetailPaneHeight` (layout.go:28-38).
- **Focus:** `FocusTarget` enum = `FocusEntryList | FocusDetailPane | FocusFilterPanel` (appshell/layout.go:7-14). Routing in `app/model.go:211-280`. Transitions: `openPane()` → `FocusDetailPane` (334); `BlurredMsg` → `FocusEntryList` (192).
- **Mouse:** `MouseRouter` maps (x,y) to `ZoneEntryList | ZoneDetailPane | ZoneDivider | ZoneHeader | ZoneStatusBar` (appshell/mouse.go:19-85) — **vertical zones only**.
- **Theme:** centralized `theme.Theme` (theme/theme.go:12-37). Tokens include `FocusBorder`, `HeaderBg`, `CursorHighlight`, `SyntaxKey/String/Number/Boolean/Null`, level colors. Built-ins: tokyo-night (default), catppuccin-mocha, material-dark.
- **Render style:** Value-receiver immutable chainable models. `Update(tea.Msg) (tea.Model, tea.Cmd)`. Root type-switch dispatch (app/model.go:108-209).
- **Existing focus indicator:** colored left border (`BorderLeft(m.Focused)` + `FocusBorder` color) at detailpane/model.go:122-124; T-082 top border at line 119.
- **Gaps:** no horizontal split, no width tracking on panes, no horizontal scroll, no tree-fold, no snapshot/golden-file tests, no mouse-zone tests for splits.
- **Tests:** 44 `_test.go` (unit + `tests/integration/smoke_test.go:23-135` full lifecycle). Stdlib `testing` only; substring + ANSI assertions, no full-screen snapshots.
- **Context artifacts:** `context/kits/cavekit-detail-pane.md` (8 reqs), `context/impl/impl-detail-pane.md`, build-site 83 tasks / 9 tiers (Tier 8 visual polish just landed — T-081 focus, T-082 top border, T-083 left border).
- Confidence: HIGH (both codebase agents concur line-for-line).

### jsonlogs.nvim reference (observed via tui-mcp)
- Right-side vertical split; on activation left ~14-20 cols / right ~60 cols (preview-dominant) at 80x40; at 200x50 still ~50/50 (preserves ratio proportionally).
- Preview is YAML-ish flat `key: value` one per line, ending `}`, with `↴` continuation marker.
- **Live-sync:** j/k in source instantly re-renders preview for new selection. No flicker.
- **Ratio toggle:** `f` cycles between **2 presets** (source-wide ↔ preview-wide). No third "full-maximize" state observed. README says "maximize/restore" but behavior is toggle.
- **Tab closes the pane entirely** — not a focus-cycle. Re-open with `<leader>jl`.
- **Wide row handling: hard truncation at border.** No soft-wrap, no h-scroll, no `…` indicator, no "press Enter to expand." Content past width is invisibly cut (observed `SecretsManagerPropertySour…` losing the rest).
- **Nested JSON:** flattened to dot-paths (`user.name`, `tags[0]`).
- **No visible focus indicator** on either pane.
- Adoption: 2⭐ niche; treat as inspiration, not gospel.
- Confidence: HIGH (directly inspected).

### Bubbletea / lipgloss landscape
- `lipgloss.JoinHorizontal` is built-in and sufficient for side-by-side layout; no higher-level split crate exists in either ratatui or bubbletea. Manual width arithmetic is the norm.
- **No dominant tree widget in bubbletea.** ratatui has `EdJoPaTo/tui-rs-tree-widget` (123⭐); bubbletea community has nothing equivalent. Custom build if fold desired.
- **Focus management: no built-in.** Apps roll own (gloggy already has `FocusTarget` enum). Options: `pgavlin/bubbletea-nav` (Tab/Shift+Tab FocusManager, auto-focus on mouse click), or continue rolling custom.
- **Mouse routing:** `lrstanley/bubblezone` is the de-facto pattern — zero-width zone IDs embedded in output, `zone.Get(id).InBounds(ev)` without coord math. gloggy's custom router works but bubblezone simplifies splits.
- **Snapshot testing:** `charmbracelet/x/exp/teatest` is the official experimental lib; golden files in testdata, `-update` flag. glow/gh-dash/soft-serve don't use it, but the tooling is blessed.
- **Responsive layout:** `winder/bubblelayout` translates `WindowSizeMsg` into weighted tile layouts — likely overkill for gloggy's two-pane need, but useful reference for constraint semantics.
- **Pitfall P8 — lipgloss v1.1.0 emoji/CJK width bug (#562):** `ansi.StringWidth()` miscounts emoji, CJK, ZWJ → layout overflow. PR #563 falls back to go-runewidth. Upgrade to v1.2+ when available. Use `lipgloss.Width()` not `len()`.
- **Pitfall P9 — no partial render:** bubbletea re-renders whole screen every Update→View cycle (issue #32 unfixed). Batch log-add updates via a 50–100ms Ticker; only re-render details on selection change.
- **Pitfall P7 — border off-by-one in JoinHorizontal:** naive 50/50 split drops border width → right pane gets negative space or wraps. Subtract border width explicitly. gloggy commit daa9fca already fixed this class of bug in the vertical direction.
- Confidence: HIGH.

### Comparable master-detail TUIs
- **GitUI** (21.8k⭐, ratatui): static master-detail (files | diff). Terminal-adaptive; no drag-resize.
- **Yazi** (ratatui file manager): three columns with configurable `ratio = [1,4,3]`. Plugin `toggle-pane.yazi` for maximize toggle.
- **Zellij:** true drag-resize + Ctrl+scroll + Alt+f floating maximize — complex; reference UX only.
- **Lazygit** (Go/gocui): multi-pane status/files/branches; terminal-adaptive; no granular drag-resize documented.
- **fzf preview:** `:wrap` flag for soft-wrap; no native h-scroll (issues #1339, #2182).
- **jless:** dot-path notation; `yp` copies current path.
- **gron:** JSONL → greppable flat `json.key = value` lines; `--ungron` reverses.
- **Convergent pattern across mature TUIs:** preset ratios + maximize toggle beats true drag-resize. Drag is rare and complex to ship correctly.
- Confidence: HIGH.

## Design Recommendations

### 1. Orientation & breakpoints
- **Recommended:** right-side split by default when width ≥ **100 cols**. Below that, fall back to current below-list stacking. Rationale: list needs ~60 cols to render time+level+logger+msg usefully; details pane needs ~40 cols minimum; <100 cols side-by-side cramps both.
- **Alternative (preferred in config):** expose `detail_pane.position = "right" | "below" | "auto"` with default `"auto"` and threshold `detail_pane.orientation_threshold_cols = 100`. "auto" flips on every `WindowSizeMsg`.
- **Tradeoff:** "auto" is forgiving but can be startling mid-session if user resizes terminal across the boundary. A status-bar hint ("layout: right-split") helps.

### 2. Very wide rows in the details pane
Three options — recommend (a), ship (b) later, defer (c):
- **(a) Soft-wrap at pane width** — lipgloss `Width()` + manual wrap, or bubbles `viewport` with `SoftWrap=true`. Safe default; matches fzf `:wrap`. **Recommended initial shipping behavior.**
- **(b) Horizontal scroll with h/l** — viewport `SetHorizontalStep()` + `ScrollLeft`/`ScrollRight` when `SoftWrap=false`. Richer but needs a visible indicator (`◀ …content… ▶`) and conflict-checks with vim-focus-switch keys.
- **(c) Modal expand** — Enter on a field → full-screen single-field view with its own scroll. Best for truly huge values (stack traces, base64 blobs); ship as v2.
- **Avoid:** jsonlogs.nvim's silent truncation — user-hostile; content vanishes with no cue.

### 3. Many properties / tall content
- Reuse current `ScrollModel` (j/k/mouse wheel). Already works.
- Keep the "known fields first, rest alphabetical" ordering in `RenderJSON()` (detailpane/render.go:64-84).
- Extend existing `/` search (`SearchModel`) to be the "jump to field" feature. No new keymap needed.
- If lists grow very long, consider an overview gutter like minimap dots later — defer.

### 4. Properties with embedded JSON
Brainstorm, ranked by build cost / ceiling:
- **Flatten inline to dot-paths** (`request.body.user.id = 123`). Cheap; matches gron/jless/jsonlogs. Risk: lossy for heterogeneous arrays; noisy for deeply nested objects. Good v1.
- **Detect valid-JSON string fields and pretty-print inline** under the parent field (indented block). Medium cost. Preserves structure and reads well at a glance.
- **Enter on field → modal drill-in** treating the embedded JSON as a new detail view. Highest ceiling; pairs naturally with 2(c). Defer to v2.
- **Tree-fold with expand/collapse per path.** High cost (bubbletea has no tree widget — build from scratch). Defer unless users ask for it.
- **Recommendation:** ship flat dot-paths first; add "Enter to drill in" modal once the flat view is proven. Revisit fold only if users request it.

### 5. Jumping between list and details when both open
- **Recommended:** **Tab cycles focus** between list and details (standard TUI convention; matches `bubbletea-nav`, tmux, most IDE splits).
- **Secondary:** `h` focuses list, `l` focuses details (vim convention) — only when not in search or filter input. Watch for conflict with horizontal-scroll `h`/`l` if you enable 2(b); if so, use Alt+h / Alt+l or omit the vim shortcut.
- **Esc from details** returns focus to list (already implemented via `BlurredMsg`). Keep.
- **Mouse:** click on a pane → focus it (adopt `bubblezone` or extend existing `MouseRouter`).
- **Do NOT overload Tab with "close pane"** (jsonlogs.nvim does this; it's surprising and unrecoverable for users who expect focus-cycle). Use Esc or a dedicated close key (`d` for details, Ctrl+w then `q` if you want vim-flavored).

### 6. Resize (maximize toggle + proportional)
- **Keymap collision check:** `m` is already taken (mark row in list pane). `f` is already bound to "open filter panel" — **but the filter feature currently does not work**, so `f` is effectively free from a user-facing standpoint. Using it here risks a future collision when filter is fixed; better to pick a key with no current binding.
- **Recommended primary:** **`z`** (vim fold-cycle metaphor — "cycle layout state" reads naturally) OR **`|`** (tmux/vim "vertical max" convention). Either cycles **three states**: `compact-details` (list-wide) ↔ `balanced` (default ratio) ↔ `maximize-details` (details-wide). Three states beats jsonlogs's two because it gives both "focus the list" and "focus the details" explicitly.
- **Fine-grained:** extend existing `+` / `-` convention — currently adjusts height ratio; in right-side mode adjust width ratio by 5% per press.
- **Reset:** `=` returns to balanced (vim `<C-w>=` convention).
- **Focus-switch:** **Tab** cycles focus list↔details. Avoid `h` / `l` for focus-switch — reserve for horizontal scroll inside details pane if 2(b) ships.
- **Persist:** save `detail_pane.width_ratio` to TOML config (gloggy already persists height ratio).
- **Avoid:**
  - `m` — collides with mark-row in list.
  - `f` — currently bound to filter (though broken); re-using invites collision once filter is fixed.
  - Ctrl+hjkl — collides with vim nav expectations.
  - True mouse-drag resize — complex, buggy on terminal-emulator boundaries (daa9fca class).

**Final suggested split-mode keymap:**
```
Tab        cycle focus list↔details
Esc        close details / blur
z (or |)   cycle 3 layout presets
+ / -      resize ±5%
=          reset to balanced
```

### 7. Focus indicator when details closed (list-only)
Four ideas — recommend **(a) + (d)** combined:
- **(a) Cursor-line highlight on selected row** — reverse-video or bg-tint using `CursorHighlight` theme token (already exists). Make it more prominent when list has focus; dimmer when another pane has focus. **Recommended.**
- **(b) Gutter glyph `▶`** in left column of cursor row. Cheap, but competes with the left border and can look noisy.
- **(c) Status-bar text** `[focus: list]` / `[focus: details]` when multiple panes exist. Good disambiguator but low-glance-value.
- **(d) Keep `FocusBorder`-colored left edge on the list frame even when it's the only pane** — at minimum a left-edge colored column. Consistent with the focused-details treatment. **Recommended.**
- **Rationale:** combining (a) and (d) is visually lightweight, theme-coherent, and answers both "what row am I on" and "which pane listens to keys". Add (c) as a free bonus when the status bar has space.

## Contradictions & Open Questions

### Contradictions resolved
- **jsonlogs.nvim `f` behavior:** README says "maximize/restore"; observed behavior is 2-preset toggle. Adopt a **3-preset cycle** for gloggy to cover both "focus list" and "focus details" intents.
- **Tab semantics:** jsonlogs.nvim uses Tab to close the pane; mainstream TUI convention uses Tab to cycle focus. Resolve in favor of the mainstream convention; add an explicit close key.
- **Ratatui vs bubbletea material:** Wave 1 (library landscape) initially framed in ratatui terms; corrected via findings-board. All "constraint" talk maps to `JoinHorizontal` + manual width arithmetic in bubbletea. No library swap needed.

### Current state quirks worth noting
- **`f` is bound to "open filter panel" but the filter feature currently does not work** end-to-end. That means `f` is effectively a dead key for users today. Attractive for re-binding to maximize-cycle, but inviting a future collision once filter is fixed — pick `z` or `|` instead.
- `m` is the mark-row key in the list pane (used with `y` to copy marks). Do not re-bind for resize.

### Open questions (flagged for prototype/decision)
- **bubbletea performance with real-time split re-renders at 10k+ lines/sec:** unknown. Mitigation is Ticker batching (50–100ms), but needs a prototype measurement once the right-split lands. Existing `entrylist` perf is <16ms for 100k entries, which is encouraging.
- **lipgloss v1.2 availability:** emoji/CJK bug #562 is fixed in PR #563 but not yet released at time of research. Check `go.mod` at implementation time; upgrade if possible. If not available, audit any width math for emoji/CJK test cases.
- **Mouse-zone buffer near border:** recommended 1–2 char buffer to prevent accidental cross-pane click, but exact buffer size is platform/terminal-dependent. Test on iTerm2, WezTerm, Kitty, Windows Terminal, xterm; provide ASCII border fallback for Alacritty.
- **Should horizontal-scroll `h`/`l` coexist with vim-focus-switch `h`/`l`?** If yes, one must win contextually (e.g. `h` scrolls only when pane is already focused and has truncated content). Consider skipping vim-focus-switch entirely and relying on Tab.
- **Config migration:** existing `detail_pane.height_ratio` stays relevant only in below-mode. Decide whether `width_ratio` is separate or both collapse into a single `detail_pane.size_ratio` keyed by current orientation.

## Implications for Design

### Should user run /ck:design — even for a TUI?
**Yes, but abbreviated and scoped.** DESIGN.md's 9-section Stitch format is optimized for web/mobile UIs, but the principles translate cleanly to TUI: component inventory, color/theme tokens (gloggy already has), layout system, focus states, keymap matrix, breakpoints. A focused DESIGN pass on the details-pane + layout-shell pays off because:
- It forces explicit decisions on all seven design questions above *before* kit/map/make — avoiding ratcheting during build.
- It produces a shared artifact that the `cavekit-detail-pane` revision and `impl-right-pane` tracking can reference.
- It formalizes keymap ownership (Tab, `f`/`m`, `+`/`-`, `=`, `h`/`l`, Esc) — which is the TUI equivalent of component naming conventions.

**Recommendation:** run `/ck:design` scoped to **details-pane + layout-shell + theme-token** domains (not the whole app). Expected output: mini-DESIGN with sections on TUI layout modes (below/right/auto + breakpoint), pane resize presets + states, focus semantics + cues, wide-row handling (soft-wrap + horizontal-scroll + modal-expand states), and the full keymap matrix. Skip the color-palette section — theme tokens already exist.

### Concrete build implications
- **Refactor `appshell/layout.go`** `Layout` struct: add `Orientation` (`below | right | auto`) + `DetailPaneWidth` alongside existing `DetailPaneHeight`. `Render()` branches between `JoinVertical` and `JoinHorizontal`. Explicitly subtract border width in horizontal case (daa9fca-class bug).
- **Update `appshell/mouse.go`** `MouseRouter.Zone()` to compute vertical slices in right-split mode; optionally adopt `lrstanley/bubblezone`. Add 1–2 char buffer near the divider.
- **Extend `config/config.go`:** `detail_pane.position`, `detail_pane.width_ratio`, `detail_pane.orientation_threshold_cols`, `detail_pane.wrap_mode` (`soft | scroll | modal`).
- **Extend `detailpane/model.go`:** `SetWidth(int)` alongside `SetHeight(int)`. Wrap logic per `wrap_mode`. Respect `lipgloss.Width()` for measurement.
- **Keymap additions:** Tab (cycle focus), `z` or `|` (3-preset maximize cycle — avoid `m` [mark-row collision] and `f` [bound to filter, though filter currently broken]), `=` (reset ratio), Esc (close details). Skip `h`/`l` as focus-switch to keep them free for in-pane horizontal scroll.
- **Tests:** introduce `charmbracelet/x/exp/teatest` for golden-file layout snapshots. Extend `tests/integration/smoke_test.go` to cover right-split + focus cycle + Tab cycle + maximize cycle. Force ASCII color profile in CI; `.gitattributes` for line endings.
- **Dependencies:** no required new deps. Optional additions: `lrstanley/bubblezone` (mouse), `charmbracelet/x/exp/teatest` (test-only). Upgrade `lipgloss` to v1.2+ once published.
- **Context artifacts:** revise `context/kits/cavekit-detail-pane.md` (split into orientation-aware requirements) OR add `cavekit-right-pane.md` + update `cavekit-app-shell.md` for orientation logic. Track tasks in new `context/impl/impl-right-pane.md`. Follow cavekit sketch → map → make.

## Codebase Context
- Go + bubbletea v1.3.10, lipgloss v1.1.0, fsnotify v1.9.0, pelletier/go-toml, atotto/clipboard, termenv.
- Entry: `cmd/gloggy/main.go` → `internal/ui/app/model.go` (root Model holding 12 subsystems at lines 20-51).
- Details pane: `internal/ui/detailpane/` — `PaneModel`, `ScrollModel`, `HeightModel`, `SearchModel`, `VisibilityModel`, `fieldclick.go`.
- Layout: `internal/ui/appshell/layout.go` — currently vertical-only; `Render()` uses `JoinVertical` exclusively (lines 96-103).
- Mouse router: `internal/ui/appshell/mouse.go` — vertical zones only (lines 19-85).
- Focus enum: `FocusTarget` = {`FocusEntryList`, `FocusDetailPane`, `FocusFilterPanel`} at `appshell/layout.go:7-14`.
- Theme: `theme/theme.go:12-37` — includes `FocusBorder`, `HeaderBg`, `CursorHighlight`, `SyntaxKey/String/Number/Boolean/Null`, level colors.
- Tests: 44 `_test.go` + `tests/integration/smoke_test.go:23-135`. Stdlib only; no snapshots.
- Kits: `context/kits/cavekit-detail-pane.md` (8 reqs). Redesign = revise existing kit or add `cavekit-right-pane.md`.
- Recent commit daa9fca ("Fix 6 layout bugs: viewport overflow, newline flattening, border accounting, WindowSizeMsg race") proves the team already knows this bug class.

## Sources

### Codebase refs (gloggy)
- `go.mod`, `cmd/gloggy/main.go`, `internal/ui/app/model.go`, `internal/ui/appshell/layout.go`, `internal/ui/appshell/mouse.go`, `internal/ui/detailpane/{model,render,scroll,height,search,visibility,fieldclick}.go`, `internal/ui/entrylist/{list,row,scroll}.go`, `internal/logsource/loader.go`, `theme/theme.go`, `config/config.go`, `tests/integration/smoke_test.go`, `context/kits/cavekit-detail-pane.md`, `context/plans/build-site.md`

### External URLs
- https://github.com/ll931217/jsonlogs.nvim
- https://github.com/charmbracelet/bubbletea (issue #32, discussion #307)
- https://github.com/charmbracelet/lipgloss (issue #562, PR #563)
- https://pkg.go.dev/github.com/charmbracelet/bubbles/viewport
- https://charm.land/blog/teatest/
- https://github.com/lrstanley/bubblezone
- https://github.com/pgavlin/bubbletea-nav
- https://pkg.go.dev/github.com/winder/bubblelayout
- https://leg100.github.io/en/posts/building-bubbletea-programs/
- https://ratatui.rs/concepts/layout/
- https://docs.rs/ratatui/latest/ratatui/layout/enum.Constraint.html
- https://github.com/EdJoPaTo/tui-rs-tree-widget
- https://github.com/Brainwires/ratatui-interact
- https://github.com/gitui-org/gitui
- https://github.com/sxyazi/yazi
- https://github.com/yazi-rs/plugins/tree/main/toggle-pane.yazi
- https://github.com/jesseduffield/lazygit
- https://zellij.dev/documentation/options.html
- https://github.com/junegunn/fzf (issues #1339, #2182)
- https://peterfaulconbridge.com/posts/jless/
- https://github.com/tomnomnom/gron
- https://github.com/mrjones2014/smart-splits.nvim
- https://vim.fandom.com/wiki/Fast_window_resizing_with_plus/minus_keys
- https://unicode.org/reports/tr11/
- https://github.com/ghostty-org/ghostty/discussions/11405
- https://github.com/alacritty/alacritty/issues/6144
- https://wezterm.org/features.html
- https://github.com/warpdotdev/Warp/issues/304
