# gloggy

> **100% vibe coded.** This project was designed, architected, and implemented entirely by [Claude Sonnet 4.6](https://anthropic.com/claude) via [Claude Code](https://claude.ai/code) — zero hand-written lines of Go. No prompts like "write me a log viewer"; instead, a structured methodology called [Cavekit](https://github.com/JuliusBrussee/cavekit) was used to translate product ideas into kits, kits into a dependency-ordered build plan, and the build plan into working code — autonomously, wave by wave. Development will continue the same way.

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
- **Scroll navigation** — `g`/`G` go to top/bottom, `Ctrl-d`/`Ctrl-u` half-page.
- **Marks** — `m` toggles a bookmark; `u`/`U` jump between marks. Marked entries are visually indicated.
- **Mouse support** — click to select, scroll wheel to scroll, click a selected entry to open the detail pane.

### Detail Pane
- **Syntax-highlighted JSON** — keys, strings, numbers, booleans, and nulls each use a distinct theme color.
- **Scrollable** — `j`/`k` or mouse wheel.
- **In-pane search** — `/` opens a search scoped to the pane; `n`/`N` cycle matches with a wrap indicator; `Esc` clears.
- **Field visibility** — hide/show individual fields; persisted to config and survives restart.
- **Resizable** — `+`/`-` adjust the pane height ratio; proportions survive terminal resize.
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
- Settings: theme, logger abbreviation depth, sub-row fields, detail pane height ratio, hidden fields
- Unknown keys are preserved on write-back (no data loss on upgrade)

## Keybindings

| Key | Action |
|-----|--------|
| `j`/`k` | Move cursor down/up |
| `g`/`G` | Go to top/bottom |
| `Ctrl-d`/`Ctrl-u` | Half page down/up |
| `e`/`E` | Next/prev ERROR |
| `w`/`W` | Next/prev WARN |
| `m` | Toggle mark |
| `u`/`U` | Next/prev mark |
| `Enter` | Open detail pane |
| `Esc` | Close detail pane / exit sub-level |
| `/` | In-pane search |
| `n`/`N` | Next/prev search match |
| `+`/`-` | Resize detail pane |
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

gloggy was built entirely with **[Claude Code](https://claude.ai/code)** running **Claude Sonnet 4.6**, using the **[Cavekit](https://github.com/JuliusBrussee/cavekit)** methodology in quality mode:

1. **Design phase** — product requirements were translated into implementation-agnostic *kits* covering 6 domains: log source, entry list, detail pane, filter engine, config, and app shell. 49 requirements, 210 acceptance criteria.
2. **Architecture phase** — Cavekit generated a concrete implementation plan from the kits: a 68-task dependency graph, tier-ordered for safe parallel execution.
3. **Build phase** — Claude Code worked through the build plan autonomously in waves, implementing each task, running tests, validating against acceptance criteria, and committing — without human intervention between tasks.

The full context (kits, build site, impl tracking, loop log) lives in [`context/`](context/) and serves as the living spec for future development. When new features are needed, the cycle repeats: update the kits, regenerate the plan, build.
