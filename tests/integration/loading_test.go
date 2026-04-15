package integration

import (
	"fmt"
	"os"
	"testing"

	"github.com/istonikula/gloggy/internal/logsource"
	"github.com/istonikula/gloggy/internal/ui/appshell"
)

// T-057: log-source/R7 — background load sends batches; LoadingModel tracks progress.
func TestBackgroundLoading_ProgressTracked(t *testing.T) {
	// Write a temp file with 50 JSONL entries.
	f, err := os.CreateTemp("", "gloggy-test-*.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	for i := 0; i < 50; i++ {
		fmt.Fprintf(f, `{"level":"INFO","msg":"entry %d"}`+"\n", i)
	}
	f.Close()

	loading := appshell.NewLoadingModel().Start()
	if !loading.IsActive() {
		t.Fatal("loading model should be active after Start()")
	}

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
	if loading.IsActive() {
		t.Error("loading should not be active after Done()")
	}
	if totalEntries < 1 {
		t.Errorf("expected at least 1 entry loaded, got %d", totalEntries)
	}
}

// T-057: app-shell/R8 — loading indicator visible during load, hidden when done.
func TestLoadingModel_ShowDuringLoad_HideWhenDone(t *testing.T) {
	m := appshell.NewLoadingModel()
	if m.View() != "" {
		t.Error("loading indicator should be hidden initially")
	}
	m = m.Start().Update(100)
	if m.View() == "" {
		t.Error("loading indicator should be visible during load")
	}
	m = m.Done()
	if m.View() != "" {
		t.Error("loading indicator should be hidden after done")
	}
}
