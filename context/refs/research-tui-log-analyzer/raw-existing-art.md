## Agent: existing-art
*Source: training knowledge, cutoff August 2025.*

### Q1: Existing TUI log analysis tools

- Found: **lnav** (Log File Navigator) — C++, ~8k stars. Full-featured TUI log viewer. Understands many log formats including JSON. Has a SQL query interface for logs, filtering, search, log format autodetection, timeline. Active development. The gold standard for local log analysis. [source: https://lnav.org / https://github.com/tstack/lnav]
- Found: lnav supports JSONL natively via format definitions. For Logstash-format JSON (which Spring/Logstash JSON encoder resembles), lnav has built-in format support. [confidence: HIGH]
- Found: **logdy** — Go-based log viewer with a **web UI** (opens in browser). Accepts stdin, files, or socket. Good for structured JSON logs. Not a true TUI — requires a browser. ~7k stars mid-2025. [source: https://github.com/logdy/logdy-core]
- Found: **angle-grinder (ag)** — Rust CLI for log aggregation, filtering, and analytics. Pipeline syntax for transforming log streams. Not a TUI per se — more of a command-line analytics tool. [source: https://github.com/rcoh/angle-grinder]
- Found: **stern** — Kubernetes multi-pod log tailing. Not a viewer/TUI, just a multi-source tail. Not relevant for local file analysis. [source: https://github.com/stern/stern]
- Found: **k9s** — Kubernetes TUI, has a log view as a feature. Built with Bubble Tea. Good reference for how Bubble Tea handles log streams. Not a standalone log analyzer. [source: https://github.com/derailed/k9s]
- Found: **fzf** — Not a log viewer, but commonly piped with log tools for interactive filtering. Shows the demand for interactive filtering of log output. [source: https://github.com/junegunn/fzf]
- Found: **jq + less/bat** — Common developer workflow: `jq -C . log.jsonl | less -R`. Works but not interactive or stateful. [confidence: HIGH]

### Q2: lnav's JSONL handling

- Found: lnav auto-detects JSON log format. For custom formats (like the Logstash format used here with `time`/`msg`/`logger`/`thread`/`level` fields), a format definition file (~/.lnav/formats/) maps fields to lnav's internal schema. Once defined, lnav provides: log level coloring, time-based navigation, field-based filtering, SQL queries like `SELECT * FROM logs WHERE log_level = 'error'`. [confidence: HIGH]
- Found: lnav's main pain points: complex format definition syntax, C++ codebase hard to extend, SQL syntax not intuitive for quick filtering during development, steep learning curve. [confidence: HIGH]
- Found: lnav handles mixed format files (JSON lines + text lines) gracefully — text lines are shown in a different color. [confidence: HIGH]

### Q3: Purpose-built JSONL TUI viewers

- Found: **jless** — Rust TUI JSON viewer. Designed for navigating a single large JSON object (not JSONL). Excellent key-based tree navigation. Not suited for log streams. ~4k stars. [source: https://github.com/PaulJuliusMartinez/jless]
- Found: No widely-adopted purpose-built JSONL-specific TUI viewer exists (as of Aug 2025) that targets the "developer analyzing app logs" use case with a modern UX. This is a genuine gap. [confidence: HIGH]
- Found: Most developers working with JSONL logs either use lnav (powerful but complex) or ad-hoc shell pipelines (jq, grep, less). [confidence: HIGH]

### Q4: UX patterns for log viewer TUIs

- Found: **Split-pane layout** is the dominant pattern: top/main pane = list of log entries (one line each), bottom/side pane = detail view of selected entry (pretty-printed JSON, all fields). k9s and lnav both use this pattern. [confidence: HIGH]
- Found: **Level-based color coding**: ERROR=red, WARN=yellow, INFO=default/white, DEBUG=grey/dim. This is universal and expected by developers. [confidence: HIGH]
- Found: **Inline filter bar**: persistent text input at the top or bottom of the list view. Filter applies in real-time as you type. fzf-style fuzzy matching is popular. [confidence: HIGH]
- Found: **Logger hierarchy** navigation: Java logger names like `org.springframework.data.repository.config.RepositoryConfigurationDelegate` are long. Showing only the last 2-3 segments (package abbreviation) saves horizontal space. Allow toggling full/short logger. [confidence: HIGH]
- Found: **Timestamp display**: Show relative time ("+0.5s", "3m ago") vs. absolute. Relative is more useful for tracking timing during development. Toggle between formats. [confidence: MEDIUM]
- Found: **Wrap vs. truncate for msg field**: Long messages (like Kafka config dump) should be truncated in list view with expansion in detail pane. [confidence: HIGH]
- Found: **Level filter toggle**: Quick keyboard shortcuts to show/hide DEBUG, show only WARN+ERROR, etc. (e.g., `d` toggles DEBUG visibility). [confidence: HIGH]
- Found: **Thread filter**: When debugging concurrent issues, filtering by thread is essential. [confidence: MEDIUM]

### Q5: Developer frustrations with existing log tools

- Found: Common complaint: lnav's format definition syntax is obscure and requires JSON config files — friction when adopting for a new log format. [confidence: HIGH]
- Found: Common complaint: `jq` pipelines are powerful but not interactive — have to re-run command to change filter. [confidence: HIGH]
- Found: Common complaint: No good tool handles "I want to click on a log entry and see all its structured fields expanded" in a terminal-native way. [confidence: HIGH]
- Found: Common want: Filter by field value (e.g., "show only lines where logger contains 'kafka'") without writing SQL or shell one-liners. [confidence: HIGH]
- Found: Common want: Copy a selected log entry's JSON to clipboard. [confidence: MEDIUM]
