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

	// Command-chain driver: a dedicated goroutine runs whichever tea.Cmd is
	// currently "active", forwards its produced msg through msgChan, then
	// parks until the test hands it the next cmd via nextCmdChan. This
	// mirrors how tea.Program serialises cmd → msg → update without
	// needing the real scheduler. Critically, the goroutine holds a single
	// in-flight cmd call — we never abandon a blocking cmd, which is what
	// loses messages in a naive harness.
	initialCmd := m.Init()
	if initialCmd == nil {
		t.Fatal("Init() returned nil cmd in follow mode")
	}

	msgChan := make(chan tea.Msg, 128)
	nextCmdChan := make(chan tea.Cmd, 16)
	stop := make(chan struct{})
	defer close(stop)

	go func() {
		current := initialCmd
		for {
			if current == nil {
				select {
				case c := <-nextCmdChan:
					current = c
				case <-stop:
					return
				}
				continue
			}
			msg := current()
			select {
			case msgChan <- msg:
			case <-stop:
				return
			}
			current = nil
		}
	}()

	// Give the tail goroutine a moment to do its initial drain + arm the watcher.
	time.Sleep(150 * time.Millisecond)

	pump := func(want int, label string) {
		deadline := time.Now().Add(8 * time.Second)
		for len(m.entries) < want {
			if time.Now().After(deadline) {
				t.Fatalf("%s: timed out (entries=%d, want=%d)", label, len(m.entries), want)
			}
			select {
			case msg := <-msgChan:
				updated, nextCmd := m.Update(msg)
				m = updated.(Model)
				if nextCmd != nil {
					nextCmdChan <- nextCmd
				}
			case <-time.After(100 * time.Millisecond):
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

	// Phase 0: initial content must be visible without any appends.
	pump(initialLines, "initial content")

	// Phase 1: first append.
	appendBatch(3, 1)
	pump(initialLines+3, "batch 1")

	// Phase 2: second append — the regression case.
	appendBatch(3, 2)
	pump(initialLines+6, "batch 2")

	// Line numbers must be monotonic from 1 and match append ordering.
	for i, e := range m.entries {
		if e.LineNumber != i+1 {
			t.Errorf("entry %d: LineNumber=%d, want %d", i, e.LineNumber, i+1)
		}
	}

	if m.tailCancel != nil {
		m.tailCancel()
	}
}
