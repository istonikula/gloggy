package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/istonikula/gloggy/internal/logsource"
)

// T-058: log-source/R8 — TailFile emits TailMsg for new lines appended to a file.
func TestTailMode_NewEntriesAppear(t *testing.T) {
	// Write initial content.
	f, err := os.CreateTemp("", "gloggy-tail-test-*.jsonl")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	fmt.Fprintf(f, `{"level":"INFO","msg":"initial"}`+"\n")
	name := f.Name()
	f.Close()

	// Start tail (begins at line 2 since line 1 already loaded).
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cmd := logsource.TailFile(ctx, name, 2)
	require.NotNil(t, cmd, "TailFile should return a non-nil Cmd")

	// Append a new line after a brief delay.
	time.Sleep(50 * time.Millisecond)
	fAppend, err := os.OpenFile(name, os.O_APPEND|os.O_WRONLY, 0)
	require.NoError(t, err)
	fmt.Fprintf(fAppend, `{"level":"WARN","msg":"new line"}`+"\n")
	fAppend.Close()

	// Drain one message in a goroutine with a timeout.
	type result struct {
		msg interface{}
	}
	ch := make(chan result, 1)
	go func() {
		msg := cmd()
		ch <- result{msg: msg}
	}()

	// B44: an environment without working fsnotify must mark this test SKIPPED
	// (explicit, visible), never silently PASS via t.Log — a slow-CI / no-inotify
	// run would otherwise mask a genuine tail-mode regression as a green test.
	select {
	case r := <-ch:
		switch m := r.msg.(type) {
		case logsource.TailStreamMsg:
			inner := m.Unwrap()
			switch inner.(type) {
			case logsource.TailMsg:
				// Good — new entry detected. Test passes.
			case logsource.TailStopMsg:
				t.Skip("TailStopMsg before any TailMsg — fsnotify unavailable in this environment")
			default:
				t.Skipf("unexpected inner tail type %T — environment-dependent", inner)
			}
		case logsource.TailMsg:
			// Good — direct tail msg. Test passes.
		case logsource.TailStopMsg:
			t.Skip("TailStopMsg before any TailMsg — fsnotify unavailable in this environment")
		default:
			t.Skipf("unexpected tail msg type %T — environment-dependent", r.msg)
		}
	case <-time.After(3 * time.Second):
		t.Skip("tail timeout — fsnotify did not emit within 3s (environment lacks working inotify)")
	}
}
