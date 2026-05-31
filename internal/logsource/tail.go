package logsource

import (
	"bufio"
	"context"
	"errors"
	"io"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"

	tea "github.com/charmbracelet/bubbletea"
)

// TailMsg carries a batch of entries made available by a single filesystem
// Write event (or by the initial pre-watcher drain of existing content).
// Emission is batched per event: if a Write delivers K newline-terminated
// lines, they are grouped into one TailMsg with Entries of length K, rather
// than K separate TailMsgs of one entry each (cavekit-log-source.md R8).
// Batched emission keeps cavekit-entry-list.md R14 tail-follow to a single
// cursor/viewport snap per event — without it, opening `gloggy -f` on a
// large file shows a visible row-by-row scroll animation as each per-line
// message triggers its own snap.
type TailMsg struct{ Entries []Entry }

// TailStopMsg signals that the tail watcher stopped. Emitted only on a
// non-recoverable condition (cavekit-log-source.md R8, V37c): a transient
// drain IO error is reported via TailErrMsg and retried with backoff before
// any TailStopMsg is sent.
type TailStopMsg struct{ Err error }

// TailErrMsg reports a drain IO error that did not (yet) stop the tail.
// Retryable=true means the watcher is backing off and reopening by path and
// the [FOLLOW] badge should stay live; Retryable=false means retries were
// exhausted and a TailStopMsg follows (V37c, V43).
type TailErrMsg struct {
	Err       error
	Retryable bool
}

const (
	maxTailRetries = 3
	tailBackoffCap = 5 * time.Second
)

// tailReader owns the persistent file handle, partial-line buffer, and the
// byte offset consumed so far. It is the testable core of TailFile: rotation
// and truncate-in-place survival (V37a/b) live here, decoupled from fsnotify
// so they can be exercised deterministically.
type tailReader struct {
	path         string
	f            *os.File
	pending      []byte
	lineNum      int
	startLineNum int
	offset       int64 // bytes consumed from the current file handle
}

func newTailReader(path string, startLineNum int) (*tailReader, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return &tailReader{path: path, f: f, startLineNum: startLineNum}, nil
}

func (tr *tailReader) close() {
	if tr.f != nil {
		_ = tr.f.Close()
	}
}

// reopen closes the current handle and reopens path from the start, resetting
// the partial-line buffer and offset. Used when a size-shrink (truncate or
// rotation) is detected, or as a recovery step after a transient IO error.
func (tr *tailReader) reopen() error {
	if tr.f != nil {
		_ = tr.f.Close()
		tr.f = nil
	}
	f, err := os.Open(tr.path)
	if err != nil {
		return err
	}
	tr.f = f
	tr.pending = tr.pending[:0]
	tr.offset = 0
	return nil
}

// drain reads all currently-available newline-terminated lines and returns
// them as one batch (cavekit-log-source.md R8 batched emission — one batch per
// filesystem event keeps cavekit-entry-list.md R14 tail-follow to a single
// cursor snap).
//
// Before reading, drain stats the path: if the file is now smaller than the
// bytes already consumed, the file was truncated in place or rotated, so it
// reopens by path and drops the stale partial-line buffer (V37b). A fresh
// bufio.Reader is created per call to sidestep bufio's sticky io.EOF state.
func (tr *tailReader) drain() ([]Entry, error) {
	if fi, err := os.Stat(tr.path); err == nil && fi.Size() < tr.offset {
		if rerr := tr.reopen(); rerr != nil {
			return nil, rerr
		}
	}

	reader := bufio.NewReaderSize(tr.f, 512*1024)
	var batch []Entry
	for {
		chunk, err := reader.ReadBytes('\n')
		tr.offset += int64(len(chunk))
		tr.pending = append(tr.pending, chunk...)
		if len(tr.pending) > 0 && tr.pending[len(tr.pending)-1] == '\n' {
			tr.lineNum++
			line := tr.pending[:len(tr.pending)-1]
			lineCopy := make([]byte, len(line))
			copy(lineCopy, line)
			tr.pending = tr.pending[:0]
			if tr.lineNum > tr.startLineNum {
				batch = append(batch, parseTailLine(lineCopy, tr.lineNum))
			}
		}
		if errors.Is(err, io.EOF) {
			return batch, nil
		}
		if err != nil {
			return batch, err
		}
	}
}

func parseTailLine(line []byte, lineNum int) Entry {
	switch Classify(line) {
	case LineTypeJSONL:
		return ParseJSONL(line, lineNum)
	default:
		return NewRawEntry(line, lineNum)
	}
}

// TailFile returns a tea.Cmd that watches path and emits TailMsg for every
// newline-terminated line, across an unbounded number of filesystem Write
// events. startLineNum controls initial emission: pass 0 to emit every line
// in the file (initial content + subsequent appends), or pass N to skip the
// first N lines and emit only lines N+1, N+2, … (used by callers that have
// already rendered the initial content via a separate loader).
//
// The ctx parameter allows cancellation; when cancelled, the goroutine closes
// the watcher and file and returns.
//
// Implementation notes (cavekit-log-source.md R8 AC1/AC4):
//   - Uses a persistent *os.File across Write events; file position is
//     preserved between drains so appended bytes are read exactly once.
//   - A fresh bufio.Reader is created per drain to sidestep bufio.Reader's
//     sticky io.EOF state, which otherwise goes deaf after the first drain.
//   - A `pending` buffer carries any trailing bytes that arrived without a
//     newline so partial writes (logger flushed mid-line) are completed on
//     the next Write event rather than emitted as a truncated line.
func TailFile(ctx context.Context, path string, startLineNum int) tea.Cmd {
	ch := make(chan tea.Msg, 64)
	go func() {
		defer close(ch)
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			ch <- TailStopMsg{Err: err}
			return
		}
		defer watcher.Close()

		tr, err := newTailReader(path, startLineNum)
		if err != nil {
			ch <- TailStopMsg{Err: err}
			return
		}
		defer tr.close()

		// emit sends a batch as a single TailMsg, honouring cancellation.
		// Returns false when ctx is cancelled mid-send.
		emit := func(batch []Entry) bool {
			if len(batch) == 0 {
				return true
			}
			select {
			case <-ctx.Done():
				return false
			case ch <- TailMsg{Entries: batch}:
				return true
			}
		}

		// drainAndRetry drains once and, on a drain IO error, reports
		// TailErrMsg + reopens by path with exponential backoff (cap 5s,
		// max 3 attempts) before giving up (V37c). Returns alive=false when
		// ctx is cancelled and a non-nil fatal error when retries are
		// exhausted (caller then sends TailStopMsg).
		drainAndRetry := func() (alive bool, fatal error) {
			backoff := 100 * time.Millisecond
			for attempt := 0; ; attempt++ {
				batch, derr := tr.drain()
				if !emit(batch) {
					return false, nil
				}
				if derr == nil {
					return true, nil
				}
				if attempt >= maxTailRetries {
					ch <- TailErrMsg{Err: derr, Retryable: false}
					return true, derr
				}
				ch <- TailErrMsg{Err: derr, Retryable: true}
				select {
				case <-ctx.Done():
					return false, nil
				case <-time.After(backoff):
				}
				if backoff *= 2; backoff > tailBackoffCap {
					backoff = tailBackoffCap
				}
				_ = tr.reopen() // best-effort; persistent failure surfaces on next drain
			}
		}

		// Emit (or skip) whatever is currently in the file before arming the
		// watcher. Arming must happen after this initial drain — otherwise
		// a Write that lands between open() and Add() would be missed, and
		// its lines would stay invisible until a second append arrived.
		if alive, ferr := drainAndRetry(); ferr != nil {
			ch <- TailStopMsg{Err: ferr}
			return
		} else if !alive {
			return
		}

		if err := watcher.Add(path); err != nil {
			ch <- TailStopMsg{Err: err}
			return
		}

		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				// Rotation: on Rename/Remove the path may now point to a new
				// inode (logrotate) — re-arm the watcher so subsequent writes
				// to the replacement file are seen. drain's size-shrink check
				// reopens the handle (V37a/b).
				if event.Op&(fsnotify.Rename|fsnotify.Remove) != 0 {
					_ = watcher.Remove(path)
					_ = watcher.Add(path)
				} else if event.Op&fsnotify.Write == 0 {
					continue
				}
				if alive, ferr := drainAndRetry(); ferr != nil {
					ch <- TailStopMsg{Err: ferr}
					return
				} else if !alive {
					return
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				ch <- TailStopMsg{Err: err}
				return
			}
		}
	}()
	return drainTail(ch)
}

// TailStreamMsg wraps a TailMsg or TailStopMsg and carries the continuation.
type TailStreamMsg struct {
	inner tea.Msg
	ch    <-chan tea.Msg
}

// Unwrap returns the inner message (TailMsg or TailStopMsg).
func (m TailStreamMsg) Unwrap() tea.Msg { return m.inner }

// Next returns a tea.Cmd to wait for the next tail event.
func (m TailStreamMsg) Next() tea.Cmd { return drainTail(m.ch) }

func drainTail(ch <-chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return TailStopMsg{}
		}
		return TailStreamMsg{inner: msg, ch: ch}
	}
}

// NewTailStreamMsgForTest constructs a TailStreamMsg carrying the given
// inner message (typically a TailMsg or TailStopMsg) for tests that drive
// the tail-stream code path without a real fsnotify watcher. The returned
// message's Next() cmd reads from a nil channel and will block forever —
// callers should ignore the tea.Cmd returned from Update.
func NewTailStreamMsgForTest(inner tea.Msg) TailStreamMsg {
	return TailStreamMsg{inner: inner}
}
