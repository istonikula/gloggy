---
created: "2026-04-15T00:00:00Z"
last_edited: "2026-04-20T21:40:35+03:00"
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
| T-070 | DONE | scanner.Err() checked in scanEntries, streamEntries, TailFile skip+event loops |
| T-073 | DONE | TailFile accepts context.Context; goroutine exits on ctx.Done(); test verifies cleanup |
| T-backprop-R8 | DONE | Backprop 2026-04-20: tail.go rewritten with persistent *os.File + fresh bufio.Reader per drain + `pending` buffer; model.go Init() passes startLineNum=0. Locks R8 AC1 (continuous emission across Write events), AC4 (initial content emit), AC5 (UI-level e2e). See .cavekit/history/backprop-log.md #1 |
| T-backprop-R8-batching | DONE | Backprop 2026-04-20 #9: TailMsg shape changed from `{Entry Entry}` to `{Entries []Entry}`; `drain()` accumulates into a local batch and flushes once per drain event. `gloggy -f bigfile` startup + `G`-after-backlog now land in one frame instead of animating N cursor snaps. Consumers updated: `internal/ui/app/model.go` TailMsg case; `internal/logsource/tail_test.go` drain loops. Two new ACs enforced by unit tests; tui-mcp sign-off AC on `logs/big.log` pending. See .cavekit/history/backprop-log.md #9 |

## Revision Log
| Date | Commit | Issue | Cavekit Update | Plan Update |
|------|--------|-------|----------------|-------------|
| 2026-04-20 | 30f743b, f98c116 | Follow mode silent after first Write event; initial content not emitted; line 1 lost | R8 AC1 tightened, AC4 + AC5 added | T-backprop-R8 |
| 2026-04-20 | aff713a, b175dc8 | Per-line TailMsg emission caused N cursor snaps on `gloggy -f` startup (visible row-by-row scroll animation) | R8 Description + 3 new ACs (batched initial drain, batched live append, tui-mcp no-animation sign-off) | T-backprop-R8-batching |
