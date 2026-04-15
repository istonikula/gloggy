---
created: "2026-04-15T00:00:00Z"
last_edited: "2026-04-15T00:00:00Z"
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
**Description:** When invoked with the tail flag on a file, newly appended lines are detected and emitted as entries in real time. Tail mode is not available for stdin.
**Acceptance Criteria:**
- [ ] [auto] With tail mode on a file, lines appended after initial load are emitted as new entries
- [ ] [auto] Tail-mode entries carry correct line numbers continuing from the last initially loaded line
- [ ] [auto] Tail mode is not activated when reading from stdin, regardless of flags
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
