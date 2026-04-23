package logsource

import (
	"bufio"
	"context"
	"io"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// stdinFlushInterval is V31's 50ms timer-flush cadence: coalesce parsed
// lines into one TailMsg per tick rather than per-line, which would storm
// cursor-snaps symmetric to V1's per-file-event failure mode (see
// TailFile drain batching).
const stdinFlushInterval = 50 * time.Millisecond

// TailStdin returns a tea.Cmd that follows stdin (or any io.Reader) and
// emits TailMsg batches via SPEC V31:
//
//   - A reader goroutine parses newline-terminated lines and pushes them
//     onto an unbounded in-memory batch.
//   - A 50ms time.Ticker flushes the current batch as one TailMsg
//     (skipping empty ticks — 0-entry ticks produce no emission).
//   - EOF → final flush + TailStopMsg{Err: nil} (pipe close is normal).
//     TUI stays interactive; V31 specifies the [FOLLOW] badge drops on
//     EOF (handled by app.Model on TailStopMsg receipt).
//   - ctx cancellation closes the channel cleanly without emitting
//     TailStopMsg, mirroring TailFile.
//
// Per-line emission is a kit violation: it would cause the same row-by-row
// scroll animation that V1 guards against for files.
func TailStdin(ctx context.Context, r io.Reader) tea.Cmd {
	ch := make(chan tea.Msg, 64)

	// parsed carries entries from the reader goroutine to the flusher.
	// Buffered to absorb bursts; the flusher drains on each tick.
	parsed := make(chan Entry, 1024)
	readerDone := make(chan error, 1)

	// Reader goroutine: bufio.Scanner with a large buffer for long JSON
	// lines; emit one Entry per line to parsed.
	go func() {
		defer close(parsed)
		scanner := bufio.NewScanner(r)
		scanner.Buffer(make([]byte, 64*1024), 1024*1024)
		lineNum := 0
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				readerDone <- ctx.Err()
				return
			default:
			}
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
			select {
			case parsed <- e:
			case <-ctx.Done():
				readerDone <- ctx.Err()
				return
			}
		}
		readerDone <- scanner.Err()
	}()

	// Flusher goroutine: drains parsed on 50ms tick; emits one TailMsg
	// per non-empty batch; on reader EOF, flushes remainder + stop msg.
	go func() {
		defer close(ch)
		ticker := time.NewTicker(stdinFlushInterval)
		defer ticker.Stop()

		var batch []Entry
		flush := func() {
			if len(batch) == 0 {
				return
			}
			select {
			case ch <- TailMsg{Entries: batch}:
			case <-ctx.Done():
			}
			batch = nil
		}

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Drain all currently-available parsed entries into batch.
				for {
					select {
					case e, ok := <-parsed:
						if !ok {
							flush()
							err := <-readerDone
							if err == context.Canceled {
								return
							}
							select {
							case ch <- TailStopMsg{Err: err}:
							case <-ctx.Done():
							}
							return
						}
						batch = append(batch, e)
					default:
						goto tickDone
					}
				}
			tickDone:
				flush()
			}
		}
	}()

	return drainTail(ch)
}
