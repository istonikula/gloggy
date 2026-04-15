package logsource

import (
	"bufio"
	"log/slog"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// LoadProgressMsg reports how many entries have been loaded so far.
type LoadProgressMsg struct{ Count int }

// LoadDoneMsg signals that loading is complete.
type LoadDoneMsg struct{}

// EntryBatchMsg carries a batch of newly parsed entries.
type EntryBatchMsg struct{ Entries []Entry }

// LoadFile returns a tea.Cmd that reads a file in the background,
// emitting EntryBatchMsg + LoadProgressMsg batches and a final LoadDoneMsg.
//
// The caller's Update should handle LoadFileStreamMsg and call its Next()
// to keep draining the stream.
func LoadFile(path string) tea.Cmd {
	ch := make(chan tea.Msg, 64)
	go func() {
		defer close(ch)
		f, err := os.Open(path)
		if err != nil {
			ch <- LoadDoneMsg{}
			return
		}
		defer f.Close()
		streamEntries(bufio.NewScanner(f), ch)
	}()
	return drainOne(ch)
}

// streamEntries reads from scanner and sends batches + progress onto ch.
func streamEntries(scanner *bufio.Scanner, ch chan<- tea.Msg) {
	scanner.Buffer(make([]byte, 0, 512*1024), 512*1024)

	const batchSize = 100
	const flushInterval = 10 * time.Millisecond

	var (
		batch     []Entry
		lineNum   int
		total     int
		lastFlush = time.Now()
	)

	flush := func() {
		if len(batch) == 0 {
			return
		}
		b := make([]Entry, len(batch))
		copy(b, batch)
		ch <- EntryBatchMsg{Entries: b}
		total += len(b)
		ch <- LoadProgressMsg{Count: total}
		batch = batch[:0]
		lastFlush = time.Now()
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
		batch = append(batch, e)
		if len(batch) >= batchSize || time.Since(lastFlush) >= flushInterval {
			flush()
		}
	}
	flush()
	if err := scanner.Err(); err != nil {
		slog.Error("stream entries scan error", "error", err)
	}
	ch <- LoadDoneMsg{}
}

// LoadFileStreamMsg wraps an inner message from the load stream and carries
// the channel needed to continue polling.
type LoadFileStreamMsg struct {
	inner tea.Msg
	ch    <-chan tea.Msg
}

// Unwrap returns the inner message (EntryBatchMsg, LoadProgressMsg, or LoadDoneMsg).
func (m LoadFileStreamMsg) Unwrap() tea.Msg { return m.inner }

// Next returns a tea.Cmd to fetch the next message from the stream.
// Call this from Update until you receive a LoadDoneMsg.
func (m LoadFileStreamMsg) Next() tea.Cmd { return drainOne(m.ch) }

// drainOne returns a tea.Cmd that reads one message from ch.
func drainOne(ch <-chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return LoadDoneMsg{}
		}
		return LoadFileStreamMsg{inner: msg, ch: ch}
	}
}
