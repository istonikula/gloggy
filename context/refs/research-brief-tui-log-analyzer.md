# Research Brief: TUI Log Analyzer

**Generated:** 2026-04-15
**Agents:** 0 codebase (greenfield), 3 web (training knowledge, cutoff Aug 2025)
**Sources consulted:** ~15 tools, libraries, and ecosystem references

## Summary

A modern JSONL-focused TUI log analyzer is a genuine gap in the tooling space — lnav is powerful but has a steep learning curve and clunky format definitions; no lightweight, keyboard-first alternative exists. Go + Bubble Tea is the pragmatic choice for fast development with a single-binary output; Rust + Ratatui is the choice if runtime performance and binary size are the priority. The log format (Logstash-style JSONL with `time`/`msg`/`logger`/`thread`/`level` + arbitrary extra fields) is well-understood and straightforward to parse. The biggest design constraint is handling large files and very long individual entries (multi-KB JSON values in `msg` or extra fields).

## Key Findings

### Library Landscape

- Recommended: **Go + Bubble Tea** — Elm-architecture TUI, Bubbles `list`+`viewport` components are purpose-built for list+detail-pane pattern, single binary via `go install`. [confidence: HIGH]
  - Companion: **Lip Gloss** for styling, **Bubbles** for list/viewport/textinput, **go-prettyjson** or **tidwall/pretty** for colorized JSON detail view
  - Tail mode: `fsnotify` or `hpcloud/tail` for live file following
- Alternative: **Rust + Ratatui** — fastest runtime, smallest binary, same immediate-mode rendering model. Higher development effort for UI-heavy code. Better choice if performance budget is tight or binary size matters for distribution. [confidence: HIGH]
- Avoid: Python Textual — no single binary, Python runtime dependency. Not appropriate for a fast local dev tool. [confidence: HIGH]
- Avoid: tview (Go) — less active ecosystem than Bubble Tea in 2025. [confidence: MEDIUM]

### Best Practices

- Separate parsing from rendering: parse JSON to `LogEntry` struct in a background goroutine, never in the render path
- Virtual list rendering: only render visible entries + buffer; lazy-load from file on scroll for large files
- Two modes: file mode (full history navigation) + tail mode (auto-scroll, live follow)
- All file I/O via `tea.Cmd` goroutines (Bubble Tea) or channels (Ratatui) — never block the main loop
- Non-JSON lines: parse as `RawEntry`, dim in list view, show raw text in detail pane
- Timestamp parsing: `time.RFC3339Nano` — the format uses nanosecond precision with timezone offset
- Level color palette: ERROR=red, WARN=yellow, INFO=white/default, DEBUG=grey/dim
- Keyboard-first: `j`/`k` navigate, `/` opens filter, `Enter` expands detail, `Esc` clears/backs out, `q` quits

### Existing Art

- **lnav** — gold standard; handles JSONL natively, SQL query interface, timeline. Reference for feature inspiration. Pain: complex format definitions, C++ codebase, steep learning curve. [confidence: HIGH]
- **jless** — excellent Rust JSON tree navigator for single documents. Reference for detail-pane tree UX. Not suitable for log streams. [confidence: HIGH]
- **logdy** — opens browser, not a TUI. Not a direct competitor. [confidence: HIGH]
- **k9s** — built with Bubble Tea, has a log pane. Good reference implementation for how Bubble Tea handles streaming log output at scale. [confidence: HIGH]

### Pitfalls to Avoid

- Parsing all lines synchronously on startup — causes multi-second delay for large files. Load in background with a spinner/progress indicator.
- Storing full pretty-printed JSON strings in memory for every entry — compute on demand only for the selected entry.
- Rendering full JSON in the list row — always truncate `msg` to terminal width; show ellipsis if overflowing.
- Blocking the update loop (Bubble Tea) or main thread (Ratatui) with file I/O or JSON parsing.
- Hardcoding known field names in detail pane — the format includes arbitrary extra fields (`arinaConfig`, etc.). Show all fields dynamically.
- Ignoring terminal resize events — recalculate layout on `tea.WindowSizeMsg` / terminal size change.
- Using `time.RFC3339` instead of `time.RFC3339Nano` — nanosecond timestamps won't parse.
- Over-engineering filter v1 — start with substring match on `msg`; add `logger:`, `level:`, `thread:` field filters in v2.

## Contradictions & Open Questions

- **Virtual list vs. full load**: For files under ~50k lines the distinction doesn't matter much. For the target use case (local dev logs), files are likely 10k–200k lines. Designing for virtual load from the start avoids a painful refactor later.
- **Filter UX**: Simple substring vs. structured query (like lnav's SQL). Open question for scope — research confirms simple substring is sufficient for v1.

## User Clarifications (post-research)

- **Language**: Go confirmed.
- **lnav**: User is familiar with lnav and likes it — specifically values its **interactive config and contextual help**. These are UX patterns worth emulating (e.g. `?` opens a help overlay, keyboard shortcuts discoverable in-app).
- **Scope**: Single file only. No multi-file support, no remote files. Keeps the tool simple.
- **Tail mode**: In scope for v1.

### Filter system requirements
- Filters are a primary feature, not an afterthought
- **Add filter from line**: with cursor on an entry, press a key to add a filter from a field value of that entry (e.g. filter to this logger, filter out this thread) — no manual typing required
- **Two filter modes**: INCLUDE (show only matching) and EXCLUDE/OUT (hide matching) — lnav calls these "out-filters"
- **Global filter toggle**: single key to temporarily disable all active filters and re-enable — essential for "what am I hiding?"
- **Quick level jumps**: `e` jump to next ERROR, `w` jump to next WARN (lnav-style); `E`/`W` for previous
- **Filter management panel**: visible list of active filters with their mode (in/out), ability to toggle individual filters, delete filters

### Entry display requirements
- **Compact list row by default**: show only the most useful fields — time, level badge, short logger, truncated msg. Not verbose.
- **Field visibility control**: user can configure which fields appear in the list row; some hidden by default (thread, extra fields)
- **Expand to full detail**: pressing Enter (or similar) expands the selected entry into a detail pane with all fields pretty-printed (syntax-highlighted JSON)
- **Field show/hide in detail pane**: ability to hide noisy fields even in the detail view (e.g. hide `arinaConfig` which is a 4KB config dump on every startup)
- The distinction between "compact list row" and "full detail" is core to the UX — list scans fast, detail gives full context

### Multi-row entry display (inline expansion)
- A single log event MAY occupy multiple rows in the list: a main row + optional sub-rows showing individual structured fields
- Sub-rows are useful for trace/correlation IDs and similar per-request context fields — not all fields, just configured ones
- Which fields appear as sub-rows is user-configurable (likely stored in config file or interactive config panel)
- **Navigation model (Option A — recommended, magit-style)**:
  - `j`/`k` always moves **event-to-event**, skipping over sub-rows — events are the primary navigation unit
  - `Tab` or `l`/`→` enters sub-rows of the current event (cursor moves into sub-row level)
  - `Esc`/`h`/`←` exits sub-rows back to event level
  - Sub-rows are visually indented and grouped under their parent event
  - Event boundary is always visually clear (e.g. different background, border, or separator) even when expanded
- This is orthogonal to the full detail pane (Enter) — inline sub-rows give quick field glance; detail pane gives full pretty-printed JSON

## Codebase Context

N/A — greenfield project.

## Implications for Design

- The core layout is: **filter bar** (top) + **entry list** (main, ~70% height) + **detail pane** (bottom, ~30% height, toggled with Enter)
- The `LogEntry` struct needs: `Time time.Time`, `Level string`, `Msg string`, `Logger string`, `Thread string`, `Extra map[string]json.RawMessage`, `Raw []byte`, `IsJSON bool`
- List row format: `[HH:MM:SS] LEVEL  short-logger  truncated-msg...` — all fields, no raw JSON
- Detail pane: full pretty-printed JSON with syntax highlighting, scrollable
- Non-JSON lines: show as-is in list with dim styling; detail pane shows raw text
- File reading strategy: read all lines into `[]LogEntry` in a goroutine on startup (for files up to ~100MB this is fast enough); show loading progress. For very large files, implement lazy seek-based loading as a stretch goal.
- Distribution: `go install` one-liner; single binary, no deps.
- **Scope constraint**: single file only — no multi-file tabs, no remote/SSH support.
- **UX inspiration from lnav**: in-app help overlay (`?`), discoverable keyboard shortcuts, interactive config where applicable.
- **UX style**: k9s, lazygit, lazyvim, emacs magit. Common patterns: persistent bottom bar showing context-sensitive key hints (no need to press `?` to discover basics), bordered panes, popup/modal overlays for actions like filtering, clean minimal chrome, section-based layout (magit influence).
- **Keybindings**: vi-style — `j`/`k` navigate, `g`/`G` top/bottom, `Ctrl-d`/`Ctrl-u` half-page, `/` search, `n`/`N` next/prev match, `Esc` cancel/back.

## Sources

- Bubble Tea — https://github.com/charmbracelet/bubbletea
- Lip Gloss — https://github.com/charmbracelet/lipgloss
- Bubbles — https://github.com/charmbracelet/bubbles
- go-prettyjson — https://github.com/hokaccha/go-prettyjson
- tidwall/pretty — https://github.com/tidwall/pretty
- fsnotify — https://github.com/fsnotify/fsnotify
- hpcloud/tail — https://github.com/hpcloud/tail
- lnav — https://lnav.org / https://github.com/tstack/lnav
- jless — https://github.com/PaulJuliusMartinez/jless
- logdy — https://github.com/logdy/logdy-core
- k9s — https://github.com/derailed/k9s
- Ratatui — https://github.com/ratatui-org/ratatui
- Python Textual — https://github.com/Textualize/textual
