---
created: "2026-04-17T20:25:44+03:00"
last_edited: "2026-04-17T20:32:00+03:00"
---

# gloggy — Visual Design System

Scope: this DESIGN.md is **focused**. It specifies the visual language for the
details-pane, the layout-shell (header + panes + status bar), and the
theme-token surface. It does **not** attempt to spec every UI surface. Filter
panel internals, help overlay content, and future modal wrap are mentioned
only where they touch pane composition, focus, or theming.

Canonical theme tokens live in `internal/theme/theme.go`. **This document
never re-declares hex values** — it names tokens and the role each plays.

---

## 1. Visual Theme & Atmosphere

**Modern DevEx, dark, opinionated.** gloggy feels like a tuned instrument for
reading production JSONL during local development — closer to the Bubble Tea
showcase apps (glow, soft-serve, lazygit) than to a raw vim split. Chrome is
present and confident, not whispering. Borders are stronger than a Neovim
window separator; the accent color (`FocusBorder`) carries hierarchy and
tells the eye where keystrokes will land.

**Key attributes:** opinionated, dense-but-breathable, accent-led, dark-first.

**Density:** information-dense, not anemic. Entry rows are single-line and
packed; the detail pane is full-fidelity pretty-printed JSON. Whitespace is
earned by the borders, not by empty rows.

**Personality:** a disciplined SRE's workbench — nothing ornamental, but the
workbench itself is crafted. The app knows which pane you're looking at and
is unafraid to say so.

Positioning axis: `vim/less` ← `lazygit/glow` (gloggy lives here) → modern
web UI. Left has no chrome; right drowns in it; gloggy chooses confident
chrome over invisible chrome.

---

## 2. Color Palette & Roles

Canonical source: `internal/theme/theme.go`. Three bundled themes ship:
`tokyo-night` (default), `catppuccin-mocha`, `material-dark`. Every token
below exists in every bundled theme; a theme switch must preserve the same
*roles* with theme-appropriate values.

**Rule:** component code must reference tokens by semantic name. Hex literals
inside `internal/ui/**` are a DESIGN.md violation.

### Token table (semantic groups)

| Group | Token | Role |
|---|---|---|
| Level | `LevelError` | `error`/`fatal` badge fg |
| Level | `LevelWarn` | `warn`/`warning` badge fg |
| Level | `LevelInfo` | `info` badge fg |
| Level | `LevelDebug` | `debug`/`trace` badge fg |
| Syntax | `SyntaxKey` | JSON key names in detail pane |
| Syntax | `SyntaxString` | JSON string literals |
| Syntax | `SyntaxNumber` | JSON numeric literals |
| Syntax | `SyntaxBoolean` | `true` / `false` |
| Syntax | `SyntaxNull` | `null` (same hue as `Dim` by convention) |
| UI | `Mark` | Marked-row gutter indicator |
| UI | `Dim` | Timestamps, abbreviated logger prefixes, null-ish text |
| UI | `SearchHighlight` | `/` search match background |
| Polish | `CursorHighlight` | Selected-row background in entry list |
| Polish | `HeaderBg` | Header bar background |
| Polish | `FocusBorder` | Focused pane border color (accent hue) |
| **NEW** | `DividerColor` | Right-split divider; border of unfocused panes |
| **NEW** | `UnfocusedBg` | Background fill of unfocused pane (dim tint) |

### The two new tokens

`DividerColor` is a **quiet** neutral — closer to `Dim` than to `FocusBorder`.
Used in two places: (1) the 1-column divider between panes in `right`-split,
and (2) the border of any pane that is currently unfocused but still visible.
The divider itself **does not recolor** on focus change; focus is
communicated by the pane borders, not by the divider.

`UnfocusedBg` is a **dim tint** laid down as the background of an unfocused
pane. It is *not* the same as `Dim`: `Dim` is a foreground color used for
timestamps, nulls, and abbreviated logger prefixes (e.g., `c.e.s.MyClass`).
`UnfocusedBg` is a background. An unfocused pane gets both — `UnfocusedBg`
behind everything plus a foreground blend toward `Dim`.

### Focus-state role summary

| Pane state | Border color | Background | Foreground |
|---|---|---|---|
| Focused | `FocusBorder` | base | full contrast |
| Unfocused | `DividerColor` | `UnfocusedBg` | blend toward `Dim` |
| Alone | `FocusBorder` | base | full contrast |

See §4 for the full component matrix and §6 for the focus model.

---

## 3. Typography — TUI-Adapted

gloggy runs in a terminal. Typography reduces to **one monospace font at one
size on a character-cell grid.** There is no type scale, no font-family
choice, no letter-spacing, no pixel metrics. What remains is *text
attributes* and the *roles* they play.

### Cell-grid assumptions

- One cell is one character. Width is measured with `lipgloss.Width()` —
  handles emoji and CJK. **Never** use `len()` on rendered strings (§7).
- Line height = one row. Vertical rhythm is counted in rows.
- Column rhythm is counted in cells.

### Text-attribute conventions

| Attribute | Use for | Don't use for |
|---|---|---|
| **Bold** | Header bar; level badge; active-pane label in status bar; current search hit | Decorative emphasis |
| Faint / `Dim` fg | Timestamps; null values; abbreviated loggers; unfocused-pane blend | Primary content |
| *Italic* | **Avoid.** Inconsistent across terminals | Anything |
| Underline | Clickable field names in detail pane | Plain emphasis |
| Reverse | **Avoid.** Use `CursorHighlight` bg instead | Selection |
| Strikethrough | Not used | — |

### Type roles (the only table that matters)

| Role | Attribute | Fg token | Example |
|---|---|---|---|
| Header source name | Bold | default | `app.log` |
| Header FOLLOW badge | Bold | `LevelWarn` | `[FOLLOW]` |
| Header counts | Bold | `Dim` | `42/200  200/4821 entries` |
| Level badge | Bold | `Level{Error\|Warn\|Info\|Debug}` | `ERROR` |
| Entry row — message | Default | default | `user login succeeded` |
| Entry row — timestamp | Faint | `Dim` | `10:42:01.233` |
| Entry row — logger | Faint | `Dim` | `c.e.s.MyClass` |
| Cursor row | Bold (if focused) | default | — |
| JSON key | Default | `SyntaxKey` | `"request_id"` |
| JSON string | Default | `SyntaxString` | `"abc-123"` |
| JSON number | Default | `SyntaxNumber` | `42` |
| JSON boolean | Default | `SyntaxBoolean` | `true` |
| JSON null | Faint | `SyntaxNull` (≈ `Dim`) | `null` |
| Clickable field | Underline | `SyntaxKey` | `request_id` |
| Search hit | Bold | default on `SearchHighlight` bg | — |
| Status bar hint | Default | `Dim` | `tab focus · \| layout · ? help` |
| Status bar focus label | Bold | `FocusBorder` | `focus: list` |

Single size across the entire app. If the user runs their terminal at 12pt
or 18pt, the cell grid stays self-consistent.

---

## 4. Components

Each entry lists purpose, states, styling notes, and minimum dimensions.

**Inventory:** Header Bar · Entry List Row · Entry List Pane · Detail Pane ·
Divider (NEW) · Status/Key-Hint Bar · Filter Panel (overlay) · Help Overlay
(overlay) · Loading/Empty indicator.

### Pane visual-state matrix (authoritative)

This governs every focusable pane — entry list, detail pane, filter overlay.
Any divergence is a bug.

| State | Left border | Top/bottom border | Background | Foreground | Cursor row |
|---|---|---|---|---|---|
| **Focused** | `FocusBorder`, bold | `FocusBorder` | base | full contrast | `CursorHighlight` bg, bold |
| **Unfocused (visible)** | `DividerColor` | `DividerColor` | `UnfocusedBg` | `Dim` blend | `CursorHighlight` bg at reduced intensity, non-bold |
| **Alone** | `FocusBorder` | n/a | base | full contrast | `CursorHighlight` bg, bold |
| **Overlay** (filter/help) | overlay border | overlay border | overlay bg | full contrast | underlying list frozen + dimmed |

**Cross-cutting:**

- The cursor row is **always rendered**, even when its pane is unfocused — so
  the user can predict where the cursor will be on re-focus. Only intensity
  and bold change. Applies to **entry list AND detail pane**; see §4.4 for
  detail-pane cursor semantics (scrolloff + search integration).
- The `CursorHighlight` bg MUST render as a **contiguous** cell span across
  the full pane content width. No gap at structural separators (`:`, `,`,
  whitespace) or between syntax-highlighted token boundaries. Per-token
  `lipgloss.Style.Render()` emits `SGR … \x1b[0m` sequences; the outer
  cursor-row bg paint does NOT re-inject the outer SGR across those inner
  resets. Implementations must either strip inner `\x1b[0m` resets on the
  cursor row OR paint bg cell-by-cell after width-aware reflow. Byte-level
  concatenation of per-token `Style.Render()` output with embedded resets
  on the cursor row is a design violation.
- Panes use `lipgloss.NormalBorder()`. Overlays use `RoundedBorder()` or
  `DoubleBorder()` so they are visually distinct from panes.
- Closing a pane (e.g., details on `Esc`) is an atomic redraw — no fade.

### 4.1 Header Bar

Persistent 1-row header showing source, tail state, and cursor position.

Format: `<source>  [FOLLOW?]  <cursorPos>/<visible>  <visible>/<total> entries`

- `source`: bold, default fg.
- `[FOLLOW]`: bold, `LevelWarn` fg, only present while tailing.
- Counts: bold, `Dim` fg.
- Background: `HeaderBg` across full width.

Degradation order when narrow (§8): drop focus label → counts → cursor-pos →
FOLLOW. Source always visible; truncate with `…` rather than drop it.

**Minimum:** 1 row × ~20 cells.

### 4.2 Entry List — Row

One log entry, single-line density.

**States:** `default`, `cursor` (bg `CursorHighlight`, bold if pane focused),
`marked` (gutter `*` in `Mark`), `search-hit` (matched span on
`SearchHighlight` bg). `cursor` and `marked` combine additively.

**Layout:** 1-cell mark gutter · timestamp (`Dim`) · level badge (`Level*`
bold) · abbreviated logger ~12 cells (`Dim`) · message (default, truncate
with `…`).

Never hard-truncate silently.

**Minimum:** 1 row × 20 cells.

### 4.3 Entry List — Pane

Container for rows. All styling follows the pane visual-state matrix.
`lipgloss.NormalBorder()`. Focused border = `FocusBorder`; unfocused =
`DividerColor`; bg swaps to `UnfocusedBg` when unfocused.

**Cursor and scrolloff:** the list is a cursor-tracking viewport.
Navigation (`j`/`k`/`g`/`G`/PgDn/PgUp/Ctrl+d/Ctrl+u + level jumps `e`/`w`
and mark jumps) move the cursor; the viewport scrolls when the cursor
would approach the top or bottom edge, keeping a **shared `scrolloff`
margin** (see below) of context rows around the cursor. Mouse wheel
scrolls the viewport; if the cursor's row would leave the visible window
minus `scrolloff` rows from the nearest edge, the cursor is dragged
along to stay in the scrolloff margin. This mirrors nvim `scrolloff`.

**Minimum:** ~30 cells × 5 rows. Below that the "terminal too small"
fallback (§8) kicks in.

### Shared scrolloff

Both the entry list and the detail pane honour a single top-level config
key `scrolloff` (TOML int, default `5`). It sets the minimum number of
context rows between the cursor and the top/bottom edge of its pane
during vertical navigation. At use time it is clamped to
`[0, floor(PaneContentHeight / 2)]` so it can never reach the midpoint.
Keep one config for both panes — users expect a single "context rows"
setting, not two. The filter panel and overlays do not honour scrolloff
(they are not scrolling viewports).

### 4.4 Detail Pane

Pretty-printed syntax-highlighted JSON for the cursored entry, with
clickable (underlined) field names as a filter-add hint.

**States:** follows the matrix. Additional substates: `closed` (not
rendered), `open, soft-wrap` (default), `open, scroll` (v1.5), `open, modal`
(v2).

**Styling:**

- Borders on **all four sides** per the matrix. The **top border is visible
  in both `below` and `right` orientations** (this was a regression fixed in
  T-082; do not drop it in either layout).
- Body uses the `Syntax*` tokens.
- Field names clickable via mouse; underline attribute signals this.
- Unfocused: `UnfocusedBg` bg + fg blend toward `Dim` + `DividerColor`
  borders. Syntax colors remain but are perceptually muted by the bg.

**Scroll position feedback:**

When the rendered JSON exceeds the visible viewport, the pane shows an
`NN%` indicator right-aligned on the **last content row** (above any
search prompt row), using `theme.Dim` foreground. The indicator must be
strictly an **overlay** — it does not add a row, does not widen the pane,
and does not displace body text. Implementation overlays by composing
onto the existing last line using cell-width truncation (`lipgloss.Width`
+ `ansi.Truncate`), so total rendered rows/cols match the pane's
allocated size.

Omit the indicator when content fits the viewport (0 scroll range). At
the top show `0%`; at the bottom show `100%`; intermediate positions are
`round(offset / (total - viewport) * 100)`. The indicator reflects the
current scroll state only — it does not replace the search-prompt row,
which always wins the last reserved row when search is active.

**Cursor and scrolloff (focused pane):**

The detail pane is a **cursor-tracking viewport** — there is always exactly
one "active content line" inside the wrapped content, marked with
`CursorHighlight` bg per the §4 matrix (reduced intensity when the pane is
unfocused-but-visible; restored when re-focused). Navigation semantics:

- `j`/`k`/`Down`/`Up`: move **cursor** by one line. Viewport follows with a
  `scrolloff` margin (cursor never closer than `scrolloff` rows to top or
  bottom edge, unless the document itself is shorter).
- `g`/`Home`: cursor → first line; viewport offset = 0.
- `G`/`End`: cursor → last line; viewport scrolled so cursor is visible at
  bottom, respecting scrolloff from the bottom edge.
- `PgDn` / `Ctrl+d` / `Space`: cursor + viewport move `max(1, viewport−1)`
  lines down, clamped.
- `PgUp` / `Ctrl+u` / `b`: symmetric upward move.
- **Mouse wheel scroll:** scrolls the viewport. If the cursor's document line
  would leave the visible window minus `scrolloff` rows from the nearest
  edge, the cursor is **dragged along** so it stays exactly on the
  `scrolloff`-th row from that edge. While the cursor is more than
  `scrolloff` rows from both edges, wheel scrolling moves the viewport
  **under** the cursor without changing the cursor's document position.
  This mirrors nvim `scrolloff` behaviour.
- **Search `n`/`N`:** move cursor to the match line; viewport adjusts so the
  cursor lands with `scrolloff` rows of context above/below where possible.
  The search highlight (`SearchHighlight` bg/fg) composes with the cursor
  row — cursor takes priority on the active line, search highlight on
  other matching lines.
- **Entry change / pane re-open:** cursor resets to line 0; viewport offset
  resets to 0.

`scrolloff` is the **shared top-level** config key defined in §4.3 — the
same value governs both the entry list and the detail pane so users can
tune one "context rows" setting. Clamped to
`[0, floor(ContentHeight / 2)]` at use time so it can never reach or
exceed the midpoint.

**Minimum:** ~30 cells × 3 rows. Below this the pane auto-closes and the
status bar emits a notice (§8).

### 4.5 Divider (NEW, right-split only)

A single column separating list from details in `right` orientation.

- **Width:** exactly 1 cell.
- **Glyph:** `│`, filled top-to-bottom, fg `DividerColor`.
- **Static:** does not recolor on focus change. Pane borders carry focus.

Both panes also draw their own borders, so the chrome stack between list
body and details body in right-split is:

```
[list R-border][divider][details L-border]  → 3 cells of chrome
```

Layout math must account for all 3 (see §5 Border accounting).

### 4.6 Status / Key-Hint Bar

One-row footer. Left side: context-specific key hints. Right side: active-
pane label (when >1 pane visible).

- Hints: `Dim`, default weight.
- Focus label: Bold, `FocusBorder`, rendered **only when >1 pane is
  visible**. In single-pane states it is omitted.
- One-shot notices (e.g., "terminal too narrow, forcing below layout")
  replace the hints for ~3 seconds, then restore.

**Minimum:** 1 row × terminal width.

### 4.7 Filter Panel (overlay — composition only)

Internals out of scope. Composition rules:

- Rendered as overlay with `lipgloss.RoundedBorder()`.
- Backdrop dimmed via fg blend (see §6).
- Takes focus while open. Tab-cycle paused. `Esc` closes.
- The entry list cursor is preserved beneath — drawn, just dimmed.

### 4.8 Help Overlay

Same composition rules as filter. `DoubleBorder()` to distinguish visually.
Content is out of scope; §9 keymap matrix is the canonical list.

### 4.9 Loading / Empty

When the list is empty: centered text inside the pane — e.g., `No entries
match the current filter` or `Waiting for input…`. Foreground `Dim`, no
bold, no border change.

---

## 5. Layout

### Orientation modes

| Mode | Stack | When |
|---|---|---|
| `below` | `header(1) / entryList / detailPane(optional) / statusBar(1)` | Default for narrow terminals; always available. |
| `right` | `header(1) / [entryList │ divider(1) │ detailPane] / statusBar(1)` | Wide terminals with details open. |
| `auto` | Flips to `right` when `terminal_width >= orientation_threshold_cols` (default 100). Re-evaluated on every `tea.WindowSizeMsg`. | Default. |

### Ratios

Split ratio is the fraction of inner area that the **detail pane** occupies.

- `below`: `detail_pane.height_ratio` — default **0.30**.
- `right`: `detail_pane.width_ratio` — default **0.30** (NEW).
- Both preserved **independently** across orientation flips. A flip from
  `below` to `right` does not overwrite one with the other.

### Ratio keymap

Resize the detail pane's share of the split. `+`/`-`/`|` act on the
**focused pane's** share (detail directly when detail-focused; list share
= 1 − detail ratio when list-focused). Presets are `{0.30, 0.50}`. Reset
`=` is focus-independent. With the detail pane closed, all four keys are
silent no-ops (nothing to resize).

| Key | Focused pane | Action |
|---|---|---|
| `\|` | detail | Toggle detail ratio between `0.30` and `0.50` |
| `\|` | list | Toggle list share between `0.30` and `0.50` (detail = 1 − share) |
| `+` | detail | Grow detail share by `0.05` (detail ratio +0.05) |
| `+` | list | Grow list share by `0.05` (detail ratio −0.05) |
| `-` | detail | Shrink detail share by `0.05` (detail ratio −0.05) |
| `-` | list | Shrink list share by `0.05` (detail ratio +0.05) |
| `=` | any | Reset detail ratio to `0.30` (list share `0.70`) |

Clamp `[0.10, 0.80]`. At a clamp boundary, further motion in the same
direction is a silent no-op — not a wrap. See cavekit-app-shell R12.

### Pane resize by mouse drag (R15)

Press-and-hold mouse-button-1 on the divider cell (the 1-column vertical
`│` in right-split, the 1-row horizontal border between panes in
below-mode). Motion updates the active ratio live (`width_ratio` in
right-split, `height_ratio` in below-mode), one visible step per event,
no throttling. Release persists the final ratio to config **exactly
once** per drag (not once per motion frame). The drag is **focus-
neutral** — `m.focus` never changes as a result of a drag. Dragging past
`[0.10, 0.80]` pins the ratio at the boundary; further motion in the
same direction is a no-op until the cursor re-enters the valid range.
Starting a drag with the detail pane closed is a silent no-op (no
divider cell to grab). See cavekit-app-shell R15.

### Border accounting (critical)

When composing right-split panes with `lipgloss.JoinHorizontal`, **explicitly
subtract the divider width (1) plus the border widths of both panes (2 each
= 4 cells total) before allocating widths**:

```go
const DividerWidth = 1

const (
    listBorders   = 2 // left + right
    detailBorders = 2
)

usable := termWidth - listBorders - detailBorders - DividerWidth
listW := int(float64(usable) * (1.0 - widthRatio))
detailW := usable - listW
```

This is the class of bug that commit `daa9fca` fixed. Re-introducing it
causes pane overflow and frame stutter.

### Config schema additions

```toml
[detail_pane]
position = "auto"                  # "below" | "right" | "auto"
orientation_threshold_cols = 100   # auto flip threshold
height_ratio = 0.30                # below mode
width_ratio = 0.30                 # right mode, NEW
wrap_mode = "soft"                 # "soft" | "scroll" (v1.5) | "modal" (v2)
```

Missing keys fall back to these defaults.

---

## 6. Elevation & Focus — TUI "Depth"

Terminals have no shadows. gloggy's elevation equivalent is **border
intensity + background tint + content contrast**.

### Focus model

- There is always **exactly one focused pane** among visible panes.
- **Opening the detail pane does NOT transfer focus.** Pressing `Enter` on
  an entry opens the pane with that entry's content, but focus stays on
  the list so `j`/`k` continue to move the list cursor and re-render the
  pane as a live preview. Users explicitly request focus with `Tab`
  (or mouse click on the pane). This preserves browse-without-commit as
  the primary interaction — see `cavekit-app-shell.md` R11 for the
  open-time focus policy and rationale.
- `Tab` cycles focus: `list → details → list`. When the filter overlay opens
  it takes focus and cycling is paused until it closes.
- `Esc` is context-sensitive:
  1. Overlay open → close the overlay.
  2. Else details open and focused → close details, return focus to list.
  3. Else **list focused with details open** → close details (matches the
     Esc-closes-pane rule so the user can dismiss the pane without first
     refocusing it).
  4. Else list focused → clear transient state (active search, mark
     selection). Otherwise no-op.
- Mouse click on a pane focuses it.
- `h` / `l` are **reserved** for horizontal scroll inside the detail pane
  when `wrap_mode = "scroll"` lands in v1.5. Do **not** rebind them to
  focus-switch now; that closes the door.
- Do **not** adopt jsonlogs.nvim's Tab-closes-pane semantics. Surprising and
  unrecoverable. Tab always cycles; closing is always explicit (`Esc`).

### Focus cues (layered, in priority order)

1. **Primary — pane border color.** Focused: `FocusBorder`. Unfocused:
   `DividerColor`.
2. **Secondary — pane background.** Focused: base. Unfocused: `UnfocusedBg`.
3. **Tertiary — status bar label.** `focus: list | details | filter`.
   Rendered **only when >1 pane is visible.** Bold, fg `FocusBorder`.
4. **Cursor row (list AND detail pane).** Focused: `CursorHighlight` bg +
   Bold. Unfocused (but visible): `CursorHighlight` bg reduced + non-Bold.
   In the detail pane, cursor semantics include `scrolloff` and search
   integration — see §4.4 "Cursor and scrolloff".

### Border conventions

| Element | Border | Focused | Unfocused |
|---|---|---|---|
| Entry list pane | `NormalBorder()` | `FocusBorder` | `DividerColor` |
| Detail pane | `NormalBorder()` | `FocusBorder` | `DividerColor` |
| Right-split divider | 1-cell `│` glyph | `DividerColor` | `DividerColor` (static) |
| Filter panel | `RoundedBorder()` | `FocusBorder` | — (always focused) |
| Help overlay | `DoubleBorder()` | `FocusBorder` | — |

**Detail pane top border** is visible in both `below` and `right`
orientations (T-082). The top border encodes focus state and must not be
dropped in either layout.

### Overlay backdrop

Overlays dim the underlying panes via a foreground blend (not a full
repaint), so users can still orient themselves:

```go
underlying := joined // the rendered panes+header+statusbar
dimmed := lipgloss.NewStyle().Faint(true).Render(underlying)
// overlay drawn on top with its own RoundedBorder/DoubleBorder
```

The overlay's own border is `FocusBorder` so it's clear where keys go.

---

## 7. Do's and Don'ts

### DO: measure rendered strings with `lipgloss.Width()`

```go
// Good — handles emoji + CJK correctly under lipgloss v1.1.0.
w := lipgloss.Width(renderedRow)

if lipgloss.Width(msg) > budget {
    msg = truncate(msg, budget-1) + "…"
}
```

### DON'T: use `len()` on rendered strings

```go
// Bad — byte count, not cell count. Emoji (4 bytes), CJK (3 bytes),
// ANSI escapes all break this. Produces overflow + column drift.
w := len(renderedRow)
```

### DO: subtract divider + border widths before splitting

```go
// Good — matches §5 Border accounting.
const DividerWidth = 1
usable := termWidth - 4 /* list + details borders */ - DividerWidth
listW := int(float64(usable) * (1.0 - widthRatio))
detailW := usable - listW
```

### DON'T: allocate pane widths from raw terminal width

```go
// Bad — pane overflow, frame stutter. The daa9fca-class bug.
listW := int(float64(termWidth) * (1.0 - widthRatio))
detailW := termWidth - listW
```

### DO: batch rapid entry-add updates via a Ticker

```go
// Good — Bubble Tea re-renders the whole screen per Update→View cycle.
// Batch new-entry inserts on a 50–100ms tick.
type flushTickMsg struct{}

func flushTickCmd() tea.Cmd {
    return tea.Tick(80*time.Millisecond, func(time.Time) tea.Msg {
        return flushTickMsg{}
    })
}
```

### DON'T: re-render on every fsnotify event

```go
// Bad — a log burst can fire hundreds of events per second.
// Each causes a full-screen re-render; the UI stutters.
case entryAddedMsg:
    return m, renderFullScreen() // no batching
```

### DO: reference theme tokens — never hex literals

```go
// Good
import "github.com/istonikula/gloggy/internal/theme"
style := lipgloss.NewStyle().
    Foreground(th.FocusBorder).
    BorderForeground(th.FocusBorder)
```

### DON'T: hardcode colors in component files

```go
// Bad — defeats theme switching, fragments the palette.
style := lipgloss.NewStyle().Foreground(lipgloss.Color("#7aa2f7"))
```

### DO: reference DESIGN.md by section in kits/impl docs

```markdown
<!-- In cavekit-detail-pane.md -->
- [ ] Detail pane renders a visible top border in both orientations
      (DESIGN.md §4.4 + §6 border conventions).
```

### Other don'ts (brief)

- **Don't hard-truncate silently.** Always show `…`.
- **Don't use `Reverse` for cursor.** Use `CursorHighlight` bg.
- **Don't use italic.** Terminal support is inconsistent.
- **Don't bind Tab to "close pane".** Tab always cycles focus.
- **Don't assume 80×24.** Always read from `tea.WindowSizeMsg`.
- **Don't re-declare theme hex values.** `internal/theme/theme.go` is canonical.

---

## 8. Responsive — Terminal Size

### Size breakpoints

| Width × Height | Behavior |
|---|---|
| `>= 100 cols` | `right`-split allowed. `auto` flips to `right`. |
| `< 100 cols` | `auto` uses `below`. If `position = "right"` is explicitly set, force `below` and emit a one-time status-bar notice: `terminal too narrow for right-split, using below layout`. |
| `>= 60 cols × >= 15 rows` | Minimum viable UI. |
| `< 60 cols` or `< 15 rows` | Render centered `terminal too small` message; suppress normal rendering. |

`orientation_threshold_cols` is configurable (default 100). The minimum-
viable floor (60×15) is hardcoded.

### Adaptive rules

- **Re-evaluate orientation on every `tea.WindowSizeMsg`** when `position =
  "auto"`. Don't debounce; `WindowSizeMsg` is already rate-limited.
- **Preserve ratios across flips.** `height_ratio` and `width_ratio` live
  independently in config and in-memory model.
- **Header bar degradation.** When narrow, drop in this order: focus label →
  entry counts (`<visible>/<total> entries`) → cursor position
  (`<cursorPos>/<visible>`) → FOLLOW badge. Source is **always** visible;
  truncate it with `…` rather than drop it.
- **Detail pane auto-close.** If computed width < 30 cells in `right` or
  height < 3 rows in `below`, auto-close the pane and emit a status-bar
  notice.

### What this is **not**

No "mobile" or "tablet". No touch-target sizing. No srcset. Do not import
web-responsive thinking.

---

## 9. Agent Prompt Guide

Quick reference for agents working on gloggy UI.

### Token quick-reference

Definitions in `internal/theme/theme.go`. Never redeclare a hex value.

- **Focus / state:** `FocusBorder`, `DividerColor`, `UnfocusedBg`,
  `CursorHighlight`, `HeaderBg`, `Dim`.
- **Semantic:** `Level{Error,Warn,Info,Debug}`,
  `Syntax{Key,String,Number,Boolean,Null}`.
- **UI:** `Mark`, `SearchHighlight`.

### Keymap matrix (authoritative)

| Key | Action | Context |
|---|---|---|
| `Tab` | Cycle focus between visible panes | Always (paused while filter/help overlay open) |
| `Esc` | Close overlay → close details → clear transient | Context-sensitive (§6) |
| `\|` | Toggle focused pane's share between presets `{0.30, 0.50}` | Any (pane-closed no-op) |
| `+` / `-` | Grow / shrink focused pane's share by ± 0.05 | Any (pane-closed no-op) |
| `=` | Reset detail ratio to 0.30 (focus-independent) | Any (pane-closed no-op) |
| `/` | Search in focused pane — routes to list search when list focused, detail-pane search when detail focused; literal character when filter panel focused | Entry list or detail pane |
| `n` / `N` | Next / previous match (search active) — cursor moves to match, viewport respects `scrolloff` | Search-active pane |
| `?` | Toggle help overlay | Any |
| `q` | Quit | Any |
| `j` / `k` | Move **cursor** down / up; viewport follows with `scrolloff` margin | Focused pane (list or detail — §4.3, §4.4) |
| `g` | Cursor → top | Detail pane (list: `gg` optional — see entry-list kit) |
| `G` | Cursor → bottom; viewport respects `scrolloff` from bottom | Detail pane (list: same) |
| `Home` | Cursor → top | Detail pane |
| `End` | Cursor → bottom | Detail pane |
| `PgDn` / `Ctrl+d` / `Space` | Cursor + viewport down (viewport − 1), clamped | Detail pane |
| `PgUp` / `Ctrl+u` / `b` | Cursor + viewport up (viewport − 1), clamped | Detail pane |
| Mouse wheel | Scroll viewport; cursor dragged when it would enter the `scrolloff` margin | List + detail pane |
| Mouse drag on divider | Resize panes live; single config write on release; focus-neutral | Any (pane-closed no-op) |
| `m` | Mark row | Entry list |
| `y` | Copy marked rows as JSONL | Any |
| `f` | Open filter panel (in progress) | Any |

### Orientation quick-reference

| Setting | Value |
|---|---|
| `position` | `"auto"` \| `"below"` \| `"right"` |
| Auto flip threshold | `orientation_threshold_cols` (default 100) |
| Below ratio | `height_ratio` (default 0.30) |
| Right ratio | `width_ratio` (default 0.30) — NEW |
| Ratio clamp | `[0.10, 0.80]` |
| Divider width | 1 cell (right mode only) — NEW |
| Min viable terminal | 60 cols × 15 rows |

### How to use this document

1. Before writing UI code, read the sections touching your surface. Pane
   work = §4 + §5 + §6 at minimum.
2. Reference by section in kits and impl-tracking docs: "Per DESIGN.md §4
   matrix, unfocused panes use `DividerColor` borders."
3. Use theme tokens by name in Go. Never hex.
4. Check §7 before opening a PR that touches layout, borders, or pane
   composition.
5. If you need a pattern this doc doesn't cover, propose an addition in the
   same PR — don't invent it in the component file.

### Example agent prompts

**Example 1 — empty list state:**

> "Add a new pane state: when the entry list is empty, render centered
> `No entries match the current filter` in `Dim` foreground, inside the
> pane's current focus border (DESIGN.md §4 matrix + §4.9). Do not change
> the pane border — follow whatever focus state the pane is already in. Use
> `Waiting for input…` if initial input hasn't arrived yet."

**Example 2 — right-split layout:**

> "Adding right-split in `internal/ui/appshell/layout.go`. Per DESIGN.md §5
> Border accounting and §7 Do's #2, subtract `DividerWidth` (= 1) plus both
> panes' border widths (4 cells total) before computing widths. Use
> `lipgloss.Width()` on any rendered row to verify the final composition.
> The divider is a 1-cell vertical `│` glyph in `DividerColor`; it does not
> recolor on focus change (§4.5)."

**Example 3 — status bar focus label:**

> "Implement the focus label using Bold with `FocusBorder` fg (§3 type roles
> + §6 tertiary cue). Render it **only when >1 pane is visible** (§4.6).
> When the filter overlay is open the label reads `focus: filter`; when
> details is focused and no overlay is open, `focus: details`; when the
> list alone is focused, suppress it entirely."

### Iteration guide

- Change one component or layout concern at a time.
- Reference token names and section numbers explicitly.
- Describe the pane's focus state (focused, unfocused, alone, overlay).
- Name the orientation (`below` / `right` / `auto`) when layout math is
  involved.
- If a proposed change breaks a rule in this document, change this document
  first — in the same PR.

---

*End of DESIGN.md.*
