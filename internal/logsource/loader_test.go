package logsource

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// drainLoader fully consumes a LoadFile command and returns all entries,
// whether any progress was reported, and the final count.
func drainLoader(t *testing.T, cmd func() interface{}) (entries []Entry, progressSeen bool, lastCount int) {
	t.Helper()
	timeout := time.After(10 * time.Second)
	for {
		select {
		case <-timeout:
			require.Fail(t, "timed out draining loader")
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
			require.Failf(t, "unexpected msg type", "%T", raw)
		}
	}
}

// T-027: R7.1 — progress signal emitted; R7.3 — done signal emitted
func TestLoadFile_ProgressAndDone(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.jsonl")
	f, err := os.Create(path)
	require.NoError(t, err)
	for i := 1; i <= 250; i++ {
		fmt.Fprintf(f, `{"level":"INFO","msg":"line %d"}`+"\n", i)
	}
	f.Close()

	rawCmd := LoadFile(path)
	cmd := func() interface{} { return rawCmd() }
	entries, progressSeen, lastCount := drainLoader(t, cmd)

	assert.True(t, progressSeen, "no LoadProgressMsg received")
	assert.Equal(t, 250, lastCount, "lastCount")
	assert.Len(t, entries, 250, "entries count")
}

// T-027: R7.2 — entries available before done (via batching)
func TestLoadFile_EntriesBeforeDone(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "big.jsonl")
	f, err := os.Create(path)
	require.NoError(t, err)
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
			assert.NotEmpty(t, inner.Entries, "first batch is empty")
			// Good — entries before done.
		case LoadDoneMsg:
			assert.Fail(t, "got LoadDoneMsg as first message; expected entries first")
		}
	case LoadDoneMsg:
		assert.Fail(t, "got LoadDoneMsg as first message; expected entries first")
	default:
		require.Failf(t, "unexpected first msg type", "%T", first)
	}
}

// T-027: line numbers correct
func TestLoadFile_LineNumbers(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ln.jsonl")
	f, err := os.Create(path)
	require.NoError(t, err)
	fmt.Fprintln(f, `{"msg":"a"}`)
	fmt.Fprintln(f, `plain text`)
	fmt.Fprintln(f, `{"msg":"c"}`)
	f.Close()

	rawCmd := LoadFile(path)
	cmd := func() interface{} { return rawCmd() }
	entries, _, _ := drainLoader(t, cmd)

	require.Len(t, entries, 3, "entries count")
	for i, e := range entries {
		assert.Equal(t, i+1, e.LineNumber, "entry %d LineNumber", i)
	}
}
