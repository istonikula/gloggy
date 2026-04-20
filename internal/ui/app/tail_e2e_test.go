package app

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Backprop R8 AC5 (new): tail entries reach the entry-list render path across
// multiple Write events, not just the logsource emission channel.
//
// Drives the real app.Model through Init + Update in a loop that mimics what
// tea.Program would do, then asserts m.entries grows by the expected count
// after each append batch. Before the fix this test fails at the *second*
// batch because the underlying bufio.Scanner has gone EOF-deaf.
func TestTailE2E_EntryListReceivesMultipleAppends(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "e2e.jsonl")

	// Seed file with 5 initial lines.
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	const initialLines = 5
	for i := 1; i <= initialLines; i++ {
		fmt.Fprintf(f, `{"level":"INFO","msg":"initial %d"}`+"\n", i)
	}
	f.Close()

	m := New(path, true, "", testCfg())
	m = send(m, tea.WindowSizeMsg{Width: 120, Height: 40})

	cmd := m.Init()
	if cmd == nil {
		t.Fatal("Init() returned nil cmd in follow mode")
	}

	// Watcher warmup.
	time.Sleep(150 * time.Millisecond)

	// pump runs the Bubble Tea command → msg → update loop until the model's
	// m.entries count reaches `want`, or until timeout. It mirrors what
	// tea.Program does internally, minus the scheduler.
	pump := func(initialCmd tea.Cmd, want int, label string) {
		deadline := time.Now().Add(5 * time.Second)
		currentCmd := initialCmd
		for len(m.entries) < want {
			if time.Now().After(deadline) {
				t.Fatalf("%s: timed out (entries=%d, want=%d)", label, len(m.entries), want)
			}
			if currentCmd == nil {
				time.Sleep(20 * time.Millisecond)
				continue
			}
			// Run the command in a goroutine with a short timeout so a command
			// that blocks on a channel doesn't stall the test.
			ch := make(chan tea.Msg, 1)
			go func(c tea.Cmd) { ch <- c() }(currentCmd)
			select {
			case msg := <-ch:
				var next tea.Cmd
				updated, nextCmd := m.Update(msg)
				m = updated.(Model)
				next = nextCmd
				currentCmd = next
			case <-time.After(500 * time.Millisecond):
				// No message yet — keep polling; don't re-queue the blocked cmd
				// because another goroutine is already waiting on it. Spin.
				// To avoid spinning forever, we bail via the outer deadline.
				// Drop currentCmd to nil so we re-enter the sleep branch above.
				currentCmd = nil
			}
		}
	}

	appendBatch := func(count, offset int) {
		f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			t.Fatal(err)
		}
		for i := 1; i <= count; i++ {
			fmt.Fprintf(f, `{"level":"WARN","msg":"batch %d line %d"}`+"\n", offset, i)
		}
		f.Close()
	}

	// Batch 0: initial content must be visible without any appends.
	pump(cmd, initialLines, "initial content")

	// Batch 1.
	appendBatch(3, 1)
	pump(nil, initialLines+3, "batch 1")

	// Batch 2 — the regression case.
	appendBatch(3, 2)
	pump(nil, initialLines+6, "batch 2")

	// Line numbers must be monotonic from 1.
	for i, e := range m.entries {
		if e.LineNumber != i+1 {
			t.Errorf("entry %d: LineNumber=%d, want %d", i, e.LineNumber, i+1)
		}
	}

	if m.tailCancel != nil {
		m.tailCancel()
	}
}
