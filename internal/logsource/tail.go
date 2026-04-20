package logsource

import (
	"bufio"
	"context"
	"errors"
	"io"
	"os"

	"github.com/fsnotify/fsnotify"

	tea "github.com/charmbracelet/bubbletea"
)

// TailMsg carries a single entry appended to a tailed file.
type TailMsg struct{ Entry Entry }

// TailStopMsg signals that the tail watcher stopped.
type TailStopMsg struct{ Err error }

// IsTailableFromStdin returns false — stdin cannot be tailed.
func IsTailableFromStdin() bool { return false }

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

		f, err := os.Open(path)
		if err != nil {
			ch <- TailStopMsg{Err: err}
			return
		}
		defer f.Close()

		var pending []byte
		lineNum := 0

		drain := func() error {
			reader := bufio.NewReaderSize(f, 512*1024)
			for {
				chunk, err := reader.ReadBytes('\n')
				pending = append(pending, chunk...)
				if len(pending) > 0 && pending[len(pending)-1] == '\n' {
					lineNum++
					line := pending[:len(pending)-1]
					lineCopy := make([]byte, len(line))
					copy(lineCopy, line)
					pending = pending[:0]
					if lineNum > startLineNum {
						var e Entry
						switch Classify(lineCopy) {
						case LineTypeJSONL:
							e = ParseJSONL(lineCopy, lineNum)
						default:
							e = NewRawEntry(lineCopy, lineNum)
						}
						select {
						case <-ctx.Done():
							return ctx.Err()
						case ch <- TailMsg{Entry: e}:
						}
					}
				}
				if errors.Is(err, io.EOF) {
					return nil
				}
				if err != nil {
					return err
				}
			}
		}

		// Emit (or skip) whatever is currently in the file before arming the
		// watcher. Arming must happen after this initial drain — otherwise
		// a Write that lands between open() and Add() would be missed, and
		// its lines would stay invisible until a second append arrived.
		if err := drain(); err != nil {
			ch <- TailStopMsg{Err: err}
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
				if event.Op&fsnotify.Write == 0 {
					continue
				}
				if err := drain(); err != nil {
					ch <- TailStopMsg{Err: err}
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
