## Agent: library-landscape
*Source: training knowledge, cutoff August 2025. No live star counts.*

### Q1: TUI frameworks comparison

- Found: **Bubble Tea (Go, charmbracelet/bubbletea)** — Elm-architecture TUI framework for Go. Event-driven model (Msg/Cmd/Model), composable components. ~29k GitHub stars as of mid-2025. Paired with Lip Gloss (styling) and Bubbles (common components: list, viewport, textinput, spinner). Very active development, widely adopted. [source: https://github.com/charmbracelet/bubbletea]
- Found: **Lip Gloss** — CSS-like styling for terminal output in Go. Handles borders, padding, colors, alignment. Essential companion to Bubble Tea. [source: https://github.com/charmbracelet/lipgloss]
- Found: **Bubbles** — Component library for Bubble Tea: `list`, `viewport` (scrollable text pane), `textinput`, `textarea`, `table`, `spinner`, `paginator`. The `viewport` component is specifically designed for large scrollable content. [source: https://github.com/charmbracelet/bubbles]
- Found: **Python Textual** — CSS-driven TUI framework for Python. Rich widget set, reactive model, supports async. ~28k stars. Ships with a demo app. Good for Python devs but requires Python runtime; no single binary. [source: https://github.com/Textualize/textual]
- Found: **Ratatui (Rust)** — Rust TUI library, successor to tui-rs. ~12k stars. Immediate-mode rendering, composable widgets. Excellent performance. Requires more boilerplate than Bubble Tea. [source: https://github.com/ratatui-org/ratatui]
- Found: **tview (Go)** — Alternative Go TUI using tcell directly. More traditional widget model (not Elm). Good for forms/tables. Less active than Bubble Tea in 2025. [source: https://github.com/rivo/tview]
- Found: **Charm ecosystem** (Bubble Tea + Lip Gloss + Bubbles + Glamour + Wish) is the most cohesive TUI stack in 2025 with the best documentation and examples. [confidence: HIGH]

### Q2: Best fit for a log viewer (large files, scrollable list, filtering, JSON detail pane)

- Found: **Bubble Tea + Bubbles viewport** is purpose-built for this pattern: list component for entry selection, viewport component for expanded detail view, textinput for filter bar. Multiple production log-viewer tools use this stack. [confidence: HIGH]
- Found: The Bubbles `list` component supports filtering natively (type to filter), custom item rendering, and keyboard navigation. [source: https://github.com/charmbracelet/bubbles/tree/master/list]
- Found: The Bubbles `viewport` component handles large content with efficient scrolling — suitable for displaying a pretty-printed JSON blob for the selected entry. [confidence: HIGH]
- Found: For very large files (100MB+), the critical constraint is that Bubble Tea is in-memory by default. A virtual/lazy-loading approach is needed for the list — load lines on demand rather than all at once. This is achievable with a custom model but requires deliberate design. [confidence: HIGH]
- Found: Python Textual has a `DataTable` widget that can handle large datasets with virtual rendering, but Python startup time and no-single-binary are drawbacks for a local dev tool. [confidence: MEDIUM]
- Found: Ratatui would require more code for the same functionality but would be faster at runtime and produce the smallest binary. Tradeoff: slower development velocity. [confidence: HIGH]

### Q3: JSON pretty-printing with syntax highlighting in terminal

- Found: **charmbracelet/glamour** — Markdown renderer for terminal. Not directly for JSON, but Charm ecosystem. [source: https://github.com/charmbracelet/glamour]
- Found: **tidwall/pretty** (Go) — JSON pretty-printer with optional color output for terminals. Lightweight, widely used. [source: https://github.com/tidwall/pretty]
- Found: **nwidger/jsoncolor** / **hokaccha/go-prettyjson** — Go libraries for colorized JSON output in terminals. go-prettyjson is commonly paired with Bubble Tea apps. [confidence: HIGH]
- Found: Custom rendering via Lip Gloss is straightforward: walk the JSON AST (encoding/json or go-json) and apply colors by token type (key, string, number, bool, null). Many Bubble Tea log viewers use this approach for full control over color scheme. [confidence: HIGH]
- Found: In Python Textual, `rich.syntax` with `json` language provides syntax-highlighted JSON out of the box. [confidence: HIGH]

### Q4: Streaming/tailing log files efficiently

- Found: For live tailing, `fsnotify` (Go) provides cross-platform file-change notifications. Pattern: tail goroutine sends `tea.Msg` to Bubble Tea when new lines arrive. [source: https://github.com/fsnotify/fsnotify]
- Found: `hpcloud/tail` (Go) — library that wraps the tail behavior with seek, follow, and re-open-on-rotate semantics. Often used in Go log viewers. [source: https://github.com/hpcloud/tail]
- Found: For large files, seek to end for tail mode vs. read full file for historical analysis — these are two distinct operational modes worth supporting. [confidence: HIGH]

### Q5: Language choice for local dev TUI tool in 2026

- Found: **Go is the pragmatic winner** for this use case: single binary distribution (no runtime deps), fast startup, excellent TUI ecosystem (Charm), strong JSON stdlib, goroutines for concurrent file reading. User already has Go projects (hetu-go, hrs-go, realworld-go). [confidence: HIGH]
- Found: Go `go install` makes it trivial to install: `go install github.com/user/gloggy@latest` → one binary in $GOPATH/bin. No Docker, no pip, no brew needed. [confidence: HIGH]
- Found: Python Textual is compelling if the user prefers Python, but packaging as a single executable requires PyInstaller or similar — more friction than Go. [confidence: HIGH]
- Found: Rust + Ratatui produces the fastest, smallest binary but development velocity is lower for a UI-heavy tool. Better suited for a tool that needs extreme performance or will be distributed widely. [confidence: HIGH]
