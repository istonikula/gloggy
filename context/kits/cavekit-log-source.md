---
created: "2026-04-15T00:00:00Z"
last_edited: "2026-04-20T21:40:35+03:00"
---

# Cavekit: Log Source

## Scope

Reading log data from a single file or stdin, classifying each line as structured JSONL or raw text, parsing structured lines into a canonical entry model, and emitting entries in order. Includes background loading and tail mode for file appends.

## Requirements

### R1: Input Sources
**Description:** The tool accepts a single file path as an argument, or reads from stdin when no argument is given. These are the only two input modes.
**Acceptance Criteria:**
- [ ] [auto] Given a file path argument, entries are produced from that file's contents
- [ ] [auto] Given no file argument and piped input on stdin, entries are produced from stdin
- [ ] [auto] Given a nonexistent file path, an error is reported before any UI renders
**Dependencies:** none

### R2: Line Classification
**Description:** Each line is classified as either JSONL (valid JSON object) or raw text. Both types are emitted as entries preserving the original line number.
**Acceptance Criteria:**
- [ ] [auto] A line containing a valid JSON object is classified as JSONL
- [ ] [auto] A line containing plain text (e.g. Logback init output, JVM warnings) is classified as raw text
- [ ] [auto] An empty line is classified as raw text
- [ ] [auto] A line containing a JSON array or scalar is classified as raw text
**Dependencies:** none

### R3: JSONL Parsing
**Description:** JSONL lines are parsed into a structured entry containing: time (RFC3339Nano), level, msg, logger, thread, a map of all remaining extra fields, the raw bytes of the original line, and a flag indicating the entry is JSON.
**Acceptance Criteria:**
- [ ] [auto] The `time` field is parsed as RFC3339Nano; the parsed value matches the original timestamp
- [ ] [auto] The `level`, `msg`, `logger`, and `thread` fields are extracted as strings
- [ ] [auto] Any JSON keys beyond `time`, `level`, `msg`, `logger`, `thread` are captured in the extra fields map with their original values
- [ ] [auto] The raw bytes of the original line are preserved verbatim
- [ ] [auto] The entry is flagged as JSON
**Dependencies:** none

### R4: Raw Text Entries
**Description:** Non-JSON lines produce a raw entry containing only the original text, flagged as non-JSON, with no structured fields.
**Acceptance Criteria:**
- [ ] [auto] A raw entry contains the original line text
- [ ] [auto] A raw entry is flagged as non-JSON
- [ ] [auto] A raw entry has no structured fields (time, level, msg, logger, thread, extra fields are all absent/zero-valued)
**Dependencies:** none

### R5: Unparseable Timestamps
**Description:** If a JSONL line has a `time` field that cannot be parsed as RFC3339Nano, the entry loads with a zero-valued time rather than failing.
**Acceptance Criteria:**
- [ ] [auto] A JSONL line with `"time": "not-a-timestamp"` produces an entry with zero time and all other fields parsed normally
- [ ] [auto] A JSONL line with no `time` key produces an entry with zero time
**Dependencies:** R3

### R6: Order and Line Numbers
**Description:** Entries are emitted in the order they appear in the source, and each entry carries its original 1-based line number.
**Acceptance Criteria:**
- [ ] [auto] Given a file with N lines, entries are emitted in order from line 1 to line N
- [ ] [auto] Each entry's line number matches its position in the original source
- [ ] [auto] Interleaved JSON and raw-text lines preserve their relative order
**Dependencies:** R2, R3, R4

### R7: Background Loading
**Description:** Reading and parsing never blocks the UI. Progress is reported as entries become available.
**Acceptance Criteria:**
- [ ] [auto] The UI receives a progress signal indicating how many entries have been loaded so far
- [ ] [auto] The UI is able to begin displaying entries before the entire file has been read
- [ ] [auto] Loading completes and a "done" signal is emitted when the source is exhausted
**Dependencies:** none

### R8: Tail Mode
**Description:** When invoked with the tail flag on a file, tail mode is a combined "initial emit + live append" mode: existing file content is emitted as entries on startup, and newly appended lines are detected and emitted as entries in real time for the duration of the session. Tail mode is not available for stdin. Emission is **batched per filesystem event**: each drain (including the initial pre-watcher drain of existing content) delivers all lines made available by that event as a single cohesive batch of entries — NOT one message per line. Batching preserves `less +F`-style instant jump-to-tail on startup and on `G` re-engage; per-line emission would cause cavekit-entry-list.md R14 tail-follow to snap the cursor N times for an N-line drain, visible as a row-by-row scroll animation that makes `gloggy -f bigfile` unusable for large files.
**Acceptance Criteria:**
- [ ] [auto] With tail mode on a file, lines appended by each filesystem write event after the watcher starts are emitted as new entries. Emission continues for the entire session — the 1st, 2nd, and Nth append batches must all produce entries (the watcher must not go deaf after any intermediate EOF)
- [ ] [auto] Tail-mode entries carry correct line numbers continuing monotonically from line 1 (for initial content) through every subsequent append
- [ ] [auto] Tail mode is not activated when reading from stdin, regardless of flags
- [ ] [auto] When tail mode starts on a non-empty file, all existing lines are emitted as entries (starting from line 1) before any subsequent append events are processed. No line is skipped
- [ ] [auto] Tail-mode entries reach the entry-list render path (`app.Model.entries`), not just the logsource emission channel. At least one test drives the model via `Init`/`Update` and asserts entry-list state grows after multiple append events
- [ ] [auto] Initial drain emits exactly one batch: for a seed file of N lines with startLineNum=0, the first TailMsg carries `len(Entries) == N`, not N separate TailMsgs of one entry each. Verified by reading the first message from `TailFile` and asserting `len(tm.Entries) == N` for N ≥ 50
- [ ] [auto] Live append emits exactly one batch per Write event: a single `os.File.Write` that delivers K newline-terminated lines produces ONE TailMsg with `len(Entries) == K`, not K TailMsgs of one entry each
- [ ] [human, tui-mcp] On `logs/big.log` with `gloggy -f`: startup lands the viewport at the tail with NO visible scroll animation — the first rendered frame already shows the last entries. After `k` to leave follow and `G` to re-engage under append pressure, `G` jumps directly to the current tail in a single frame (no intermediate animation). Verify at both small (80x24) and large (140x35) geometry
**Dependencies:** R1, R6

## Out of Scope

- Multiple file inputs
- Remote file sources (HTTP, S3, etc.)
- Log format configuration or custom parsers
- Filtering or searching (handled by filter-engine)
- Compressed files
- Character encoding detection (UTF-8 assumed)

## Cross-References

- See also: cavekit-entry-list.md (consumes emitted entries)
- See also: cavekit-filter-engine.md (filters over entry data model)
- See also: cavekit-app-shell.md (invokes loading, displays progress/tail status)

## Changelog

### 2026-04-20 — Revision (R8 batched emission)
- **Affected:** R8 (Description expanded; 3 new ACs — batched initial drain, batched live append, tui-mcp no-animation sign-off)
- **Summary:** R8 previously specified that tail mode emits every line with correct numbering across every Write event, but said nothing about emission **granularity**. The implementation emitted one `TailMsg` per line — satisfying every AC while causing a user-visible scroll animation on `gloggy -f bigfile` (N cursor snaps during the initial drain of an N-line file). Revised R8 requires batched emission: one `TailMsg{Entries: []Entry}` per drain event (initial pre-watcher drain AND each subsequent Write event), grouping all lines made available by that event into a single message. This locks cavekit-entry-list.md R14 tail-follow to one cursor/viewport snap per event — startup is an instant jump, not a scroll animation.
- **Driven by:** User report via `/ck:revise --trace`: "is it necessary that when file is opened with -f it scrolls until the end is reached, if scroll is stopped by moving cursor and then G is pressed is starts scrolling again until end instead of just going there directly, the scrolling can be problematic with big files". Backprop-log entry #9.

### 2026-04-20 — Backprop trace (single-failure)
- **Affected:** R8 (AC1 tightened; AC4 + AC5 added; Description expanded)
- **Summary:** Tail mode was user-visibly broken — follow mode on a non-empty file showed nothing until the first append, then dumped all initial content (minus line 1) in a single burst, then went silent for every subsequent append. Root cause: long-lived `bufio.Scanner` goes EOF-deaf, plus `app.Init()` passed `startLineNum=1` so line 1 was always lost and initial content was never emitted until a Write event. R8 AC1 only asserted "appended lines emitted" in a single-batch test; it did not require emission to survive multiple Write events or that initial content be emitted on startup. AC5 adds an end-to-end assertion that tail entries reach `app.Model.entries`, not just the logsource channel.
- **Commits:** 30f743b (failing regression tests), f98c116 (fix + kit-referenced harness rewrite)
