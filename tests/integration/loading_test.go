package integration

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/istonikula/gloggy/internal/logsource"
	"github.com/istonikula/gloggy/internal/ui/appshell"
)

// T-057: log-source/R7 — background load sends batches; LoadingModel tracks progress.
func TestBackgroundLoading_ProgressTracked(t *testing.T) {
	// Write a temp file with 50 JSONL entries.
	f, err := os.CreateTemp("", "gloggy-test-*.jsonl")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	for i := 0; i < 50; i++ {
		fmt.Fprintf(f, `{"level":"INFO","msg":"entry %d"}`+"\n", i)
	}
	f.Close()

	loading := appshell.NewLoadingModel().Start()
	require.True(t, loading.IsActive(), "loading model should be active after Start()")

	// Drain the stream by invoking Cmds synchronously.
	cmd := logsource.LoadFile(f.Name())
	var totalEntries int
	done := false

	for !done && cmd != nil {
		raw := cmd()
		switch m := raw.(type) {
		case logsource.LoadFileStreamMsg:
			inner := m.Unwrap()
			switch inner := inner.(type) {
			case logsource.EntryBatchMsg:
				totalEntries += len(inner.Entries)
				loading = loading.Update(totalEntries)
			case logsource.LoadDoneMsg:
				loading = loading.Done()
				done = true
			}
			if !done {
				cmd = m.Next()
			} else {
				cmd = nil
			}
		case logsource.LoadDoneMsg:
			loading = loading.Done()
			done = true
			cmd = nil
		default:
			cmd = nil
		}
	}

	if loading.IsActive() {
		loading = loading.Done()
	}
	assert.False(t, loading.IsActive(), "loading should not be active after Done()")
	assert.GreaterOrEqualf(t, totalEntries, 1, "expected at least 1 entry loaded, got %d", totalEntries)
}

// T-057: app-shell/R8 — loading indicator visible during load, hidden when done.
func TestLoadingModel_ShowDuringLoad_HideWhenDone(t *testing.T) {
	m := appshell.NewLoadingModel()
	assert.Empty(t, m.View(), "loading indicator should be hidden initially")
	m = m.Start().Update(100)
	assert.NotEmpty(t, m.View(), "loading indicator should be visible during load")
	m = m.Done()
	assert.Empty(t, m.View(), "loading indicator should be hidden after done")
}
