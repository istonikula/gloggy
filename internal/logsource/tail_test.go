package logsource

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// T-028: R8.3 — stdin not tailablefunction
func TestIsTailableFromStdin(t *testing.T) {
	if IsTailableFromStdin() {
		t.Error("IsTailableFromStdin() must return false")
	}
}

// T-028: R8.1 + R8.2 — new lines detected with correct line numbers
func TestTailFile_DetectsNewLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tail.jsonl")

	// Write initial lines.
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	const initialLines = 5
	for i := 1; i <= initialLines; i++ {
		fmt.Fprintf(f, `{"msg":"initial %d"}`+"\n", i)
	}
	f.Close()

	// Start tail from end of initial content.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tailCmd := TailFile(ctx, path, initialLines)
	cmd := func() interface{} { return tailCmd() }

	// Give the watcher time to set up before writing.
	time.Sleep(150 * time.Millisecond)

	// Append new lines.
	f, err = os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatal(err)
	}
	const appendCount = 3
	for i := 1; i <= appendCount; i++ {
		fmt.Fprintf(f, `{"msg":"new %d"}`+"\n", i)
	}
	f.Close()

	// Collect tail entries.
	var tailEntries []Entry
	timeout := time.After(5 * time.Second)

	for len(tailEntries) < appendCount {
		select {
		case <-timeout:
			t.Fatalf("timed out waiting for tail entries (got %d, want %d)", len(tailEntries), appendCount)
		default:
		}

		raw := cmd()
		switch m := raw.(type) {
		case TailStreamMsg:
			if tm, ok := m.Unwrap().(TailMsg); ok {
				tailEntries = append(tailEntries, tm.Entry)
			}
			nextCmd := m.Next()
			cmd = func() interface{} { return nextCmd() }
		case TailStopMsg:
			t.Fatalf("tail stopped unexpectedly: %v", m.Err)
		}
	}

	// Verify line numbers continue from initialLines.
	for i, e := range tailEntries {
		want := initialLines + i + 1
		if e.LineNumber != want {
			t.Errorf("tailEntries[%d].LineNumber = %d, want %d", i, e.LineNumber, want)
		}
	}
}

// T-073: cancelling context cleans up goroutine/watcher/file.
func TestTailFile_CancelCleansUp(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cancel.jsonl")
	if err := os.WriteFile(path, []byte("line\n"), 0644); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	tailCmd := TailFile(ctx, path, 1)

	// Give watcher time to set up.
	time.Sleep(100 * time.Millisecond)

	// Cancel context.
	cancel()

	// The tail cmd should eventually yield TailStopMsg (channel closes).
	done := make(chan struct{})
	go func() {
		defer close(done)
		raw := tailCmd()
		// Drain until we get stop or channel-closed.
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
		// OK — goroutine cleaned up.
	case <-time.After(3 * time.Second):
		t.Fatal("timeout: TailFile goroutine did not clean up after cancel")
	}
}
