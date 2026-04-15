# Findings Board: Build a TUI to analyze logs interactively when developing on local machine

> Shared coordination state for research agents.
> Source: training knowledge (cutoff Aug 2025). No live web data.

## Agent: library-landscape

- **Go + Bubble Tea** is the leading TUI stack (29k stars, Elm architecture, Lip Gloss + Bubbles ecosystem). Best fit for this use case.
- Bubbles `list` + `viewport` components are purpose-built for the list-with-detail-pane pattern.
- For large files, virtual rendering is needed — not load-everything-at-once.
- Single binary Go distribution (`go install`) is ideal for a local dev tool.
- Python Textual is strong but no single binary; Ratatui (Rust) is fast but slower to develop.
- `tidwall/pretty` or `hokaccha/go-prettyjson` for colorized JSON in terminal.
- `fsnotify` + `hpcloud/tail` for file-tail mode.

## Agent: existing-art

- **lnav** is the state-of-the-art for local log analysis — SQL queries, format autodetection, mixed-format support. But: steep learning curve, obscure format definitions, complex setup for custom JSON formats.
- **logdy** opens a web browser — not a TUI.
- **jless** is great for single JSON documents, not log streams.
- **No purpose-built modern JSONL TUI viewer exists** — genuine gap in the tooling space.
- UX patterns: split-pane (list + detail), level color coding, inline filter bar, logger abbreviation, keyboard-first (vim bindings).
- Developer pain points: lnav friction for custom formats, jq not interactive, no "click to expand all fields" TUI tool.

## Agent: best-practices + pitfalls

- Separate parsing from rendering; parse to `LogEntry` struct in background goroutine.
- Virtual list for large files; two modes: file mode and tail mode.
- All I/O in `tea.Cmd`, never block the update loop.
- Handle non-JSON lines as "raw text" entries (dim in list, show raw in detail).
- Use `time.RFC3339Nano` for parsing timestamps.
- Implement `list.Item.FilterValue()` with `msg` field for built-in Bubbles filtering.
- Start simple: substring filter on msg. Field-specific filtering is v2.
- Handle `tea.WindowSizeMsg` for terminal resize.
