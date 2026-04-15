package logsource

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// drainLoader fully consumes a LoadFile command and returns all entries,
// whether any progress was reported, and the final count.
func drainLoader(t *testing.T, cmd func() interface{}) (entries []Entry, progressSeen bool, lastCount int) {
	t.Helper()
	timeout := time.After(10 * time.Second)
	for {
		select {
		case <-timeout:
			t.Fatal("timed out draining loader")
		default:
		}
		raw := cmd()
		switch m := raw.(type) {
		case LoadFileStreamMsg:
			switch inner := m.Unwrap().(type) {
			case EntryBatchMsg:
				entries = append(entries, inner.Entries...)
			case LoadProgressMsg:
				progressSeen = true
				lastCount = inner.Count
			case LoadDoneMsg:
				return
			}
			// continue with next
			nextCmd := m.Next()
			cmd = func() interface{} { return nextCmd() }
		case LoadDoneMsg:
			return
		default:
			t.Fatalf("unexpected msg type %T", raw)
		}
	}
}

// T-027: R7.1 — progress signal emitted; R7.3 — done signal emitted
func TestLoadFile_ProgressAndDone(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.jsonl")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	for i := 1; i <= 250; i++ {
		fmt.Fprintf(f, `{"level":"INFO","msg":"line %d"}`+"\n", i)
	}
	f.Close()

	rawCmd := LoadFile(path)
	cmd := func() interface{} { return rawCmd() }
	entries, progressSeen, lastCount := drainLoader(t, cmd)

	if !progressSeen {
		t.Error("no LoadProgressMsg received")
	}
	if lastCount != 250 {
		t.Errorf("lastCount = %d, want 250", lastCount)
	}
	if len(entries) != 250 {
		t.Errorf("got %d entries, want 250", len(entries))
	}
}

// T-027: R7.2 — entries available before done (via batching)
func TestLoadFile_EntriesBeforeDone(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "big.jsonl")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 500; i++ {
		fmt.Fprintf(f, `{"msg":"entry %d"}`+"\n", i)
	}
	f.Close()

	rawCmd := LoadFile(path)
	// Read just the first message — it must be a batch, not done.
	first := rawCmd()
	switch m := first.(type) {
	case LoadFileStreamMsg:
		switch inner := m.Unwrap().(type) {
		case EntryBatchMsg:
			if len(inner.Entries) == 0 {
				t.Error("first batch is empty")
			}
			// Good — entries before done.
		case LoadDoneMsg:
			t.Error("got LoadDoneMsg as first message; expected entries first")
		}
	case LoadDoneMsg:
		t.Error("got LoadDoneMsg as first message; expected entries first")
	default:
		t.Fatalf("unexpected first msg type %T", first)
	}
}

// T-027: line numbers correct
func TestLoadFile_LineNumbers(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ln.jsonl")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Fprintln(f, `{"msg":"a"}`)
	fmt.Fprintln(f, `plain text`)
	fmt.Fprintln(f, `{"msg":"c"}`)
	f.Close()

	rawCmd := LoadFile(path)
	cmd := func() interface{} { return rawCmd() }
	entries, _, _ := drainLoader(t, cmd)

	if len(entries) != 3 {
		t.Fatalf("got %d entries, want 3", len(entries))
	}
	for i, e := range entries {
		if e.LineNumber != i+1 {
			t.Errorf("entry %d: LineNumber = %d, want %d", i, e.LineNumber, i+1)
		}
	}
}
