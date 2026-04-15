## Agent: best-practices + pitfalls
*Source: training knowledge, cutoff August 2025. Builds on library-landscape and existing-art findings.*

### Best Practices

#### Architecture
- Found: **Separate parsing from rendering.** Parse JSONL lines into a structured type (`LogEntry`) eagerly on load or in a background goroutine. Never parse JSON in the render path — it causes jank. [confidence: HIGH]
- Found: **Virtual list rendering** for large files. Don't hold all rendered lines in memory — only render what's visible plus a small buffer. Bubbles `list` component handles this. For files >10k lines, lazy-load from file on scroll. [confidence: HIGH]
- Found: **Two operational modes**: (1) file mode — load a file, navigate history; (2) tail mode — follow appended lines like `tail -f`. Distinguish with a flag (`--tail` / `-f`). Tail mode should auto-scroll to bottom. [confidence: HIGH]
- Found: **Background goroutine for file I/O**. In Bubble Tea, all I/O should happen in `tea.Cmd` (which runs in a goroutine), sending `tea.Msg` back to the main model. Never block the update loop. [confidence: HIGH]
- Found: **Filter as a separate state** from the full entry list. Keep the full slice of `LogEntry`, maintain a filtered view index. Recalculate filtered view on filter change. This avoids re-parsing the file on every keystroke. [confidence: HIGH]

#### UX
- Found: **Progressive disclosure**: list view shows minimal info (timestamp, level, truncated msg, logger short form); detail pane shows full entry with all fields pretty-printed. Never show raw JSON in list view. [confidence: HIGH]
- Found: **Keyboard-first design**: `j`/`k` or arrow keys to navigate, `/` to focus filter input, `Enter` to open detail, `Esc` to close detail / clear filter, `q` to quit. Vim-like bindings are expected for TUI tools. [confidence: HIGH]
- Found: **Status bar**: show total lines, visible lines (after filter), current position, file name, tail mode indicator. Essential for orientation in large files. [confidence: HIGH]
- Found: **Non-JSON line handling**: dim these lines (grey) in the list, show raw text in detail pane. Don't skip them — they often contain context (JVM startup, Logback init). [confidence: HIGH]

#### Go/Bubble Tea specifics
- Found: Use `encoding/json` or `go-json` (bytedance/sonic) for parsing. For this use case (one entry at a time), stdlib `encoding/json` is sufficient. [confidence: HIGH]
- Found: Use `lipgloss.NewStyle()` with `.Foreground(lipgloss.Color("..."))` for level-based coloring. Define a color palette once and reuse across components. [confidence: HIGH]
- Found: The Bubbles `list.Model` uses `list.Item` interface — implement `FilterValue() string` to enable built-in filtering. Return `msg` field value for text search. [confidence: HIGH]

### Pitfalls to Avoid

- Found: **Parsing all lines on startup for large files.** A 100MB log file with 500k lines will cause multi-second startup delay if parsed synchronously. Load in chunks or background goroutine with a loading indicator. [confidence: HIGH]
- Found: **Storing full raw JSON in memory for every entry.** Store only parsed fields + raw string for detail view. A `[]byte` raw copy per entry is fine but avoid redundant string allocations. [confidence: MEDIUM]
- Found: **Rendering full JSON in the list row.** Long JSON (like the arinaConfig dump) will overflow the terminal width and slow rendering. Always truncate in list view. [confidence: HIGH]
- Found: **Blocking the Bubble Tea update loop.** Any slow operation in `Update()` freezes the UI. File reading, JSON parsing, filtering must all happen in `tea.Cmd` goroutines. [confidence: HIGH]
- Found: **Hardcoding field names.** The JSONL format has standard fields (time/msg/logger/thread/level) but also arbitrary extra fields. The detail pane must handle unknown fields gracefully — show all key-value pairs, not just known ones. [confidence: HIGH]
- Found: **Ignoring terminal resize events.** Bubble Tea sends `tea.WindowSizeMsg` — handle it to recalculate layout. Split-pane layouts are especially prone to breaking on resize. [confidence: HIGH]
- Found: **Not handling invalid JSON lines gracefully.** Some lines will be non-JSON (Logback init, JVM warnings). A strict JSON parser will error — catch parse errors and treat as "raw text" entries. [confidence: HIGH]
- Found: **Assuming all timestamps are the same format.** The `time` field in this format is RFC 3339 with nanoseconds and timezone. Use `time.Parse(time.RFC3339Nano, ...)` — not `time.RFC3339` which won't parse nanoseconds. [confidence: HIGH]
- Found: **Over-engineering filter UX on first pass.** Start with simple substring match on `msg` field. Add field-specific filtering (logger:, level:, thread:) as a v2 feature. Premature filter complexity blocks shipping a useful tool. [confidence: HIGH]
