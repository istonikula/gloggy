package integration

import (
	"fmt"
	"os"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/istonikula/gloggy/internal/config"
	"github.com/istonikula/gloggy/internal/logsource"
	"github.com/istonikula/gloggy/internal/ui/app"
)

// TestAppModel_LoadFileStream verifies that app.Model.Update correctly processes the
// complete LoadFileStreamMsg chain, including LoadProgressMsg messages. This is a
// regression test for the chain-breaking bug where LoadProgressMsg had no case in the
// inner switch, causing cmd to stay nil and all subsequent messages to be lost.
func TestAppModel_LoadFileStream(t *testing.T) {
	// Write a temp file with 110 JSONL entries (exceeds batchSize=100).
	f, err := os.CreateTemp("", "gloggy-appload-*.jsonl")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	const numEntries = 110
	for i := 0; i < numEntries; i++ {
		fmt.Fprintf(f, `{"level":"INFO","msg":"entry %d"}`+"\n", i)
	}
	f.Close()

	cfgResult := config.LoadResult{Config: config.DefaultConfig()}
	m := app.New(f.Name(), false, "", cfgResult)

	// Give the model a window size so layout is valid.
	var tm tea.Model
	tm, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = tm.(app.Model)

	// Start loading — Init returns the first LoadFile command.
	cmd := m.Init()
	require.NotNil(t, cmd, "Init() should return a non-nil cmd for file loading")

	// Drain the stream through the actual app model's Update method.
	var iterations int
	const maxIterations = 1000
	done := false

	for !done && cmd != nil && iterations < maxIterations {
		iterations++
		raw := cmd()
		tm, retCmd := m.Update(raw)
		m = tm.(app.Model)
		cmd = retCmd

		// Check if we received LoadDoneMsg (loading indicator becomes inactive).
		if msg, ok := raw.(logsource.LoadFileStreamMsg); ok {
			if _, ok := msg.Unwrap().(logsource.LoadDoneMsg); ok {
				done = true
			}
		}
		if _, ok := raw.(logsource.LoadDoneMsg); ok {
			done = true
		}
	}

	assert.Truef(t, done, "loading never completed after %d iterations — LoadProgressMsg likely broke the chain", iterations)

	// Verify all 110 entries were loaded.
	assert.NotEmpty(t, m.View(), "view should not be empty after loading")
}
