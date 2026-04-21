package app

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/istonikula/gloggy/internal/ui/appshell"
)

// runYProgramAndCapture drives a live tea.Program with the given seed Model,
// optionally pre-marks N rows with `m`, then sends `y`, waits for the
// renderer to emit a frame, quits, and returns the bytes the renderer
// wrote. V25 requires that every y-feedback path's notice be verified
// against the renderer output (not just m.View()).
func runYProgramAndCapture(t *testing.T, seed Model, preMarkCount int) string {
	t.Helper()

	out := &bytes.Buffer{}
	p := tea.NewProgram(seed,
		tea.WithInput(&bytes.Buffer{}),
		tea.WithOutput(out),
		tea.WithoutSignalHandler(),
	)

	go func() {
		// Let the program start and render the initial frame.
		time.Sleep(150 * time.Millisecond)
		for i := 0; i < preMarkCount; i++ {
			p.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
			p.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		}
		// Let the marks render before pressing y.
		time.Sleep(100 * time.Millisecond)
		p.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
		// Give the notice time to pass through Update + renderer frame
		// but finish before the 2s auto-clear would hide it.
		time.Sleep(400 * time.Millisecond)
		p.Quit()
	}()

	if _, err := p.Run(); err != nil {
		t.Fatalf("tea.Program.Run: %v", err)
	}
	return out.String()
}

// TestLiveBuffer_YNoMarks_NoticeReachesRenderer covers V25 for the no-marks
// path (V15, B1). A user pressed y with no marks and saw no feedback; unit
// tests of m.View() passed because the notice string was present in the
// View output, but the bytes never reached the terminal (fix: render the
// notice in a high-contrast style distinct from the Dim keyhints row).
func TestLiveBuffer_YNoMarks_NoticeReachesRenderer(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	m = m.SetEntries(makeEntries(3))

	got := runYProgramAndCapture(t, m, 0)
	if !strings.Contains(got, clipboardNoMarksNotice) {
		t.Errorf("no-marks notice %q missing from live-buffer output", clipboardNoMarksNotice)
	}
}

// TestLiveBuffer_YCopied_NoticeReachesRenderer covers V25 for the success
// path: with marks present, y produces a "copied N entries" notice. The
// clipboard write is stubbed so the test does not touch the system
// clipboard.
func TestLiveBuffer_YCopied_NoticeReachesRenderer(t *testing.T) {
	restore := appshell.SetClipboardWriteFnForTesting(func(string) error { return nil })
	defer restore()

	m := newModel()
	m = resize(m, 80, 24)
	m = m.SetEntries(makeEntries(3))

	got := runYProgramAndCapture(t, m, 2)
	want := formatClipboardCopiedNotice(2)
	if !strings.Contains(got, want) {
		t.Errorf("copied notice %q missing from live-buffer output", want)
	}
}

// TestLiveBuffer_YClipboardErr_NoticeReachesRenderer covers V25 for the
// error path: the clipboard write fails, the y-handler surfaces a visible
// error via the notice bar.
func TestLiveBuffer_YClipboardErr_NoticeReachesRenderer(t *testing.T) {
	restore := appshell.SetClipboardWriteFnForTesting(func(string) error {
		return errors.New("pbcopy: broken")
	})
	defer restore()

	m := newModel()
	m = resize(m, 80, 24)
	m = m.SetEntries(makeEntries(3))

	got := runYProgramAndCapture(t, m, 2)
	// The app phrases clipboard errors with the literal "clipboard"
	// substring or the underlying error message — assert on the error
	// string we injected, which is the most specific signal.
	if !strings.Contains(got, "pbcopy: broken") && !strings.Contains(got, "clipboard") {
		t.Errorf("clipboard-error notice missing from live-buffer output; captured bytes (first 500):\n%s",
			truncateForLog(got, 500))
	}
}

func truncateForLog(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "...(truncated)"
}
