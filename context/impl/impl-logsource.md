---
created: "2026-04-15T00:00:00Z"
last_edited: "2026-04-15T00:00:00Z"
---
# Implementation Tracking: log-source

Build site: context/plans/build-site.md

| Task | Status | Notes |
|------|--------|-------|
| T-001 | DONE | go.mod, dirs, cmd/gloggy/main.go stub |
| T-002 | DONE | Entry struct in internal/logsource/entry.go |
| T-003 | DONE | Classify() in classify.go + 5 tests |
| T-004 | DONE | ParseJSONL() in parse.go + 5 tests |
| T-005 | DONE | NewRawEntry() in parse.go + 2 tests |
| T-015 | DONE | ReadFile() in reader.go — opens file, delegates to scanEntries |
| T-016 | DONE | ReadStdin(io.Reader) in reader.go — same scan path as ReadFile |
| T-017 | DONE | reader_test.go — ProducesEntries, NonexistentError, ReadStdin, LineNumbers, MixedContent |
| T-027 | DONE | logsource/loader.go — LoadFile() channel-polling background load; EntryBatchMsg/LoadProgressMsg/LoadDoneMsg |
| T-028 | DONE | logsource/tail.go — TailFile() with fsnotify; IsTailableFromStdin()=false; continuing line numbers |
