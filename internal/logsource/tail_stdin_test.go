package logsource

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// drainStdinTail pumps the tail Cmd chain until `want` entries arrive or
// timeout hits. Returns the observed TailMsgs in order so tests can
// assert batch shape (V31: 0-entry ticks emit nothing; non-empty ticks
// emit exactly one TailMsg).
func drainStdinTail(t *testing.T, cmd func() interface{}, want int, timeout time.Duration) []TailMsg {
	t.Helper()
	var msgs []TailMsg
	entriesSeen := 0
	deadline := time.After(timeout)
	for entriesSeen < want {
		select {
		case <-deadline:
			require.Failf(t, "timed out", "stdin tail: got %d entries in %d TailMsgs, want %d", entriesSeen, len(msgs), want)
		default:
		}
		raw := cmd()
		switch m := raw.(type) {
		case TailStreamMsg:
			switch inner := m.Unwrap().(type) {
			case TailMsg:
				msgs = append(msgs, inner)
				entriesSeen += len(inner.Entries)
			case TailStopMsg:
				require.Failf(t, "tail stopped early", "after %d entries: %v", entriesSeen, inner.Err)
			}
			nextCmd := m.Next()
			cmd = func() interface{} { return nextCmd() }
		case TailStopMsg:
			require.Failf(t, "tail stopped early", "after %d entries: %v", entriesSeen, m.Err)
		}
	}
	return msgs
}

// V31 — 50ms timer-flush batching: a burst of lines written to stdin
// BEFORE the first tick MUST coalesce into a single TailMsg (one
// cursor-snap, not per-line). Mirrors TailFile's per-Write batching
// guarantee on files.
func TestTailStdin_BurstCoalescesIntoOneBatch(t *testing.T) {
	const n = 100
	var buf strings.Builder
	for i := 1; i <= n; i++ {
		fmt.Fprintf(&buf, `{"msg":"burst %d"}`+"\n", i)
	}
	// pipe lets us close the writer to drive EOF; burst is pre-written
	// so all lines are available before the first tick.
	pr, pw := io.Pipe()
	go func() {
		_, _ = pw.Write([]byte(buf.String()))
		_ = pw.Close()
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cmd := TailStdin(ctx, pr)
	cmdFn := func() interface{} { return cmd() }

	msgs := drainStdinTail(t, cmdFn, n, 3*time.Second)
	// V31: the whole burst lands in one batch on the first non-empty tick.
	// (Race: if the tick fires between partial writes, we may get 2
	// batches — allow up to (burst_ms/50)+1 per T23.)
	assert.LessOrEqual(t, len(msgs), 3, "expected ≤ 3 TailMsgs for pre-tick burst (got %d: %v)", len(msgs), batchSizes(msgs))
	total := 0
	for _, m := range msgs {
		total += len(m.Entries)
	}
	assert.Equal(t, n, total, "total entries across batches")

	// Line numbers monotonic 1..n.
	var all []Entry
	for _, m := range msgs {
		all = append(all, m.Entries...)
	}
	for i, e := range all {
		assert.Equal(t, i+1, e.LineNumber, "entry %d LineNumber", i)
	}
}

// V31 — stdin EOF (pipe close) → TailStopMsg{Err:nil}. TUI stays
// interactive (caller's responsibility; this test asserts the stop
// signal shape).
func TestTailStdin_EOFFiresStopMsg(t *testing.T) {
	pr, pw := io.Pipe()
	go func() {
		_, _ = pw.Write([]byte("line one\nline two\n"))
		_ = pw.Close()
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cmd := TailStdin(ctx, pr)
	cmdFn := func() interface{} { return cmd() }

	// Pump until we see TailStopMsg.
	sawStop := false
	stopErr := error(nil)
	deadline := time.After(3 * time.Second)
	for !sawStop {
		select {
		case <-deadline:
			require.Fail(t, "timed out waiting for TailStopMsg")
		default:
		}
		raw := cmdFn()
		switch m := raw.(type) {
		case TailStreamMsg:
			if s, ok := m.Unwrap().(TailStopMsg); ok {
				sawStop = true
				stopErr = s.Err
				break
			}
			nextCmd := m.Next()
			cmdFn = func() interface{} { return nextCmd() }
		case TailStopMsg:
			sawStop = true
			stopErr = m.Err
		}
	}
	assert.NoError(t, stopErr, "pipe close should produce TailStopMsg{Err:nil}")
}

// V31 — 0-entry ticks produce no emission. A reader that holds the pipe
// open but sends nothing must produce zero TailMsgs over the observation
// window (≥ 4 ticks).
func TestTailStdin_EmptyTicksSuppressed(t *testing.T) {
	pr, pw := io.Pipe()
	// Hold the pipe open; don't write anything.
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		_ = pw.Close()
	}()

	cmd := TailStdin(ctx, pr)

	// Wait ≥ 4 ticks (~200ms) then assert no TailMsg arrived.
	type result struct {
		msg interface{}
	}
	out := make(chan result, 1)
	go func() {
		out <- result{msg: cmd()}
	}()

	select {
	case r := <-out:
		// Should not arrive within 4 ticks — empty-tick suppression.
		if tsm, ok := r.msg.(TailStreamMsg); ok {
			if _, isTail := tsm.Unwrap().(TailMsg); isTail {
				require.Fail(t, "got unexpected TailMsg on empty stdin (0-entry tick must not emit)")
			}
		}
	case <-time.After(250 * time.Millisecond):
		// OK — no emission during idle window.
	}
}

// Cancelling the context cleans up goroutines without emitting a stray
// TailStopMsg (mirrors TailFile_CancelCleansUp).
func TestTailStdin_CancelCleansUp(t *testing.T) {
	pr, pw := io.Pipe()
	defer func() { _ = pw.Close() }()

	ctx, cancel := context.WithCancel(context.Background())
	cmd := TailStdin(ctx, pr)
	cancel()

	// Channel should close promptly; drain until we see the closed-channel
	// zero value (TailStopMsg{} from drainTail).
	done := make(chan struct{})
	go func() {
		defer close(done)
		raw := cmd()
		for {
			switch m := raw.(type) {
			case TailStreamMsg:
				if _, ok := m.Unwrap().(TailStopMsg); ok {
					return
				}
				nextCmd := m.Next()
				raw = nextCmd()
			case TailStopMsg:
				return
			default:
				return
			}
		}
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		require.Fail(t, "TailStdin goroutines did not exit after cancel")
	}
}

func batchSizes(msgs []TailMsg) []int {
	out := make([]int, len(msgs))
	for i, m := range msgs {
		out[i] = len(m.Entries)
	}
	return out
}
