package logsource

import (
	"bufio"
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

// TailFile returns a tea.Cmd that watches path for new lines appended after
// initial load. startLineNum is the count of lines already read; new entries
// are numbered startLineNum+1, startLineNum+2, etc.
//
// The returned cmd yields TailStreamMsg values; call TailStreamMsg.Next()
// to keep watching. When a TailStopMsg is received, tailing has ended.
func TailFile(path string, startLineNum int) tea.Cmd {
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

		// Skip past already-loaded lines.
		scanner := bufio.NewScanner(f)
		scanner.Buffer(make([]byte, 0, 512*1024), 512*1024)
		for i := 0; i < startLineNum && scanner.Scan(); i++ {
		}

		lineNum := startLineNum

		if err := watcher.Add(path); err != nil {
			ch <- TailStopMsg{Err: err}
			return
		}

		for event := range watcher.Events {
			if event.Op&fsnotify.Write == 0 {
				continue
			}
			for scanner.Scan() {
				lineNum++
				line := scanner.Bytes()
				lineCopy := make([]byte, len(line))
				copy(lineCopy, line)
				var e Entry
				switch Classify(lineCopy) {
				case LineTypeJSONL:
					e = ParseJSONL(lineCopy, lineNum)
				default:
					e = NewRawEntry(lineCopy, lineNum)
				}
				ch <- TailMsg{Entry: e}
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
