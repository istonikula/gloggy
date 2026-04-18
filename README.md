# gloggy

> **100% vibe coded.** This project was designed, architected, and implemented entirely by [Claude](https://anthropic.com/claude) via [Claude Code](https://claude.ai/code) — zero hand-written lines of Go. No prompts like "write me a log viewer"; instead, a structured methodology called [Cavekit](https://github.com/JuliusBrussee/cavekit) was used to translate product ideas into kits, kits into a dependency-ordered build plan, and the build plan into working code — autonomously, wave by wave. Development will continue the same way.

> ⚠️ **Not yet human tested.** The codebase has 100% automated test coverage against its acceptance criteria, but has not been run as an actual TUI by a human yet. Screenshots and real-world validation are coming soon.

A terminal UI for interactively analyzing JSONL log files during local development. Single binary, reads from a file or stdin.

## Comparison

| Feature | **gloggy** | [logtui](https://github.com/jnatten/logtui) | [lnav](https://lnav.org) | [jless](https://jless.io) |
|---|:---:|:---:|:---:|:---:|
| JSONL-first design | ✓ | ✓ | ~ | ✓ |
| Tail / follow mode | ✓ | ~ ¹ | ✓ | ✗ |
| Include / exclude filters | ✓ | ~ ² | ~ ³ | ✗ |
| Level-jump navigation (e/w) | ✓ | ✗ | ✗ | ✗ |
| Two-level cursor (entry → fields) | ✓ | ✗ | ✗ | ✓ |
| In-pane search | ✓ | ✓ | ✓ | ✓ |
| Field visibility toggle (persisted) | ✓ | ~ | ✗ | ✗ |
| Bookmarks / marks | ✓ | ✗ | ✗ | ✗ |
| Clipboard copy of entries | ✓ | ✗ | ✗ | ~ ⁴ |
| Multiple themes | ✓ | ✗ | ✓ | ✗ |
| Virtual rendering (100k+ entries) | ✓ | ✗ | ✓ | ✗ |
| Mouse support | ✓ | ✗ | ✓ | ✗ |
| Multi-format log support | ✗ | ✗ | ✓ | ✗ |
| SQL queries | ✗ | ✗ | ✓ | ✗ |
| Histogram / timeline view | ✗ | ✗ | ✓ | ✗ |
| Open entry in $EDITOR | ✗ | ✓ | ✗ | ✗ |
| Pause/resume ingestion | ✗ | ✓ | ✗ | ✗ |

¹ Autoscroll mode for stdin; no inotify file watching  
² Regex filter across all fields; no include/exclude field-scoped rules  
³ SQL `WHERE` clauses rather than a filter panel  
⁴ Copies jq-style paths and values, not whole entries  

**Choose gloggy if** you spend your day tailing JSONL logs from a single service and want fast level-jumping, bookmarks, and field-scoped filters without leaving the terminal.  
**Choose lnav if** you need to correlate multiple heterogeneous log files, run SQL queries, or work with non-JSON formats.  
**Choose logtui if** you want live stdin streaming with pause/resume and quick editor integration.  
**Choose jless if** you need a `less` replacement for exploring JSON/JSONL files and don't need filtering or follow mode.

## Features

### Entry List
- **Compact row format** — each entry shows `HH:MM:SS`, a colored level badge, an abbreviated logger name, and a truncated message. Non-JSON lines are shown dimmed.
- **Virtual rendering** — only visible rows plus a small buffer are rendered, keeping the list responsive with 100k+ entry files.
- **Level-jump navigation** — `e`/`E` jump to the next/previous ERROR, `w`/`W` for WARN. Wraps with an indicator.
- **Two-level cursor** — magit-style: `j`/`k` move between entries; `l`/`Tab`/`→` expand into per-field sub-rows; `h`/`←`/`Esc` collapse back.
- **Scroll navigation** — `g`/`G` go to top/bottom, `Ctrl-d`/`Ctrl-u` half-page. Viewport follows the cursor with a configurable `scrolloff` margin of context rows around the cursor (default 5).
- **Marks** — `m` toggles a bookmark; `u`/`U` jump between marks. Marked entries are visually indicated.
- **Mouse support** — click to select, scroll wheel to scroll (cursor is dragged along when it would leave the scrolloff margin, nvim-style), click a selected entry to open the detail pane.

### Detail Pane
- **Syntax-highlighted JSON** — keys, strings, numbers, booleans, and nulls each use a distinct theme color.
- **Cursor-tracking viewport** — one highlighted "active line" at all times. `j`/`k` move the cursor; `g`/`G`/`Home`/`End` jump to top/bottom; `PgDn`/`PgUp`/`Ctrl-d`/`Ctrl-u`/`Space`/`b` page. Mouse wheel scrolls the viewport and drags the cursor at the edges so a configurable `scrolloff` margin (default 5) of context rows is always preserved — the same model as `nvim`.
- **Scroll-position indicator** — an `NN%` overlay on the last content row shows position within the entry; omitted when content fits.
- **In-pane search** — `/` opens a search scoped to the pane; `n`/`N` move the cursor to the next/prev match with scrolloff context; wrap indicator and `(current/total)` counter shown; `Esc` clears.
- **Field visibility** — hide/show individual fields; persisted to config and survives restart.
- **Right-split or below layouts** — auto-flips based on terminal width; `|` cycles layout presets, `+`/`-` adjust the pane ratio, `=` resets; independent height/width ratios are preserved across flips. Pane is resizable by dragging the divider too.
- **Filter from field** — click a field value to pre-fill a filter prompt.

### Filter Engine
- **Include/exclude filters** — field:pattern rules; choose include or exclude mode per filter.
- **Filter panel** — `f` opens an overlay listing active filters; `j`/`k` navigate, `Space` toggles enabled, `d` deletes.
- **Global toggle** — disable all filters at once and restore them.

### App Shell
- **Three themes** — `tokyo-night`, `catppuccin-mocha`, `material-dark`; set in TOML config.
- **Header bar** — file name (or `stdin`), `[FOLLOW]` badge in tail mode, total/visible entry counts.
- **Tail mode** — `gloggy -f <file>` follows a live file via inotify.
- **Background loading** — large files load incrementally with a progress indicator.
- **Clipboard** — `y` copies all marked entries to the system clipboard in JSONL format.
- **Help overlay** — `?` shows all keybindings by domain; `Esc` closes.
- **Context-sensitive key hints** — the bottom bar shows relevant keybindings for the focused component.

### Config
- TOML config at `~/.config/gloggy/config.toml`
- Settings: theme, logger abbreviation depth, sub-row fields, hidden fields, detail pane height/width ratios, orientation (`below` / `right` / `auto`) + auto-flip threshold, wrap mode, and a shared top-level **`scrolloff`** (default 5) honoured by both the entry list and the detail pane.
- Unknown keys are preserved on write-back (no data loss on upgrade)

## Keybindings

### Movement (entry list OR focused detail pane)

| Key | Action |
|-----|--------|
| `j`/`k` (or `↓`/`↑`) | Move cursor down/up by one row; viewport follows with `scrolloff` context |
| `g`/`G` (or `Home`/`End`) | Jump cursor to top/bottom |
| `PgDn`/`PgUp` | Page down/up |
| `Ctrl-d`/`Ctrl-u` | Half page down/up |
| `Space` / `b` | Page down / up (detail pane) |
| Mouse wheel | Scroll viewport; cursor dragged at edges to respect `scrolloff` |

### Entry list only

| Key | Action |
|-----|--------|
| `e`/`E` | Next/prev ERROR |
| `w`/`W` | Next/prev WARN |
| `m` | Toggle mark |
| `u`/`U` | Next/prev mark |
| `Tab` / `l` / `→` | Expand entry into per-field sub-rows |
| `h` / `←` | Collapse sub-rows |
| `Enter` | Open detail pane for current entry |

### Detail pane + search

| Key | Action |
|-----|--------|
| `/` | Open in-pane search (if list focused, focus transfers to detail pane with search active) |
| `n`/`N` | Move cursor to next/prev match (viewport adjusts to keep `scrolloff` context) |
| `Esc` | Dismiss search, then close pane (two-step) |

### Layout + focus

| Key | Action |
|-----|--------|
| `Tab` | Cycle focus between visible panes |
| `Esc` | Close overlay → close detail pane → clear transient state |
| `\|` | Cycle layout ratio presets (10% / 30% / 70%) |
| `+` / `-` | Adjust the active pane ratio by ±5% |
| `=` | Reset ratio to 0.30 |

### Global

| Key | Action |
|-----|--------|
| `f` | Open filter panel |
| `y` | Copy marked entries to clipboard |
| `?` | Help overlay |
| `q` | Quit |

## Installation

```sh
go install github.com/istonikula/gloggy/cmd/gloggy@latest
```

### Build from source

```sh
git clone https://github.com/istonikula/gloggy.git
cd gloggy
make install        # installs to $GOBIN (or $GOPATH/bin)
```

Or build without installing:

```sh
make build          # produces dist/gloggy
```

Run `make help` to see all available targets.

## Usage

```sh
gloggy app.log          # open a file
gloggy -f app.log       # tail/follow mode
cat app.log | gloggy    # read from stdin
```

## Built With

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — Elm-architecture TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) — layout and styling
- [fsnotify](https://github.com/fsnotify/fsnotify) — file watching for tail mode
- [go-toml](https://github.com/pelletier/go-toml) — TOML config parsing
- [atotto/clipboard](https://github.com/atotto/clipboard) — system clipboard

## How It Was Built

gloggy was built entirely with **[Claude Code](https://claude.ai/code)** driving **[Claude](https://anthropic.com/claude)** (multiple models across phases — reasoning, execution, and exploration), using the **[Cavekit](https://github.com/JuliusBrussee/cavekit)** methodology in quality mode:

1. **Design phase** — product requirements were translated into implementation-agnostic *kits* covering 6 domains: log source, entry list, detail pane, filter engine, config, and app shell. Currently 58 requirements across ~330 acceptance criteria.
2. **Architecture phase** — Cavekit generated a concrete implementation plan from the kits: a tiered dependency graph (currently 137 tasks across 15 tiers) for safe parallel execution.
3. **Build phase** — Claude Code worked through the build plan autonomously in waves, implementing each task, running tests, validating against acceptance criteria, and committing — without human intervention between tasks.

The full context (kits, build site, impl tracking, loop log) lives in [`context/`](context/) and serves as the living spec for future development. When new features are needed or bugs surface, the cycle repeats: `/ck:check` surfaces gaps, kits and DESIGN.md are revised, the site gets new tasks, and `/ck:make` builds them.
