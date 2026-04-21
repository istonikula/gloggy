package app

import (
	"bytes"
	"errors"
	"regexp"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/istonikula/gloggy/internal/ui/appshell"
)

// sgrOpenRe captures the parameter string inside `\x1b[...m` SGR sequences.
var sgrOpenRe = regexp.MustCompile(`\x1b\[([0-9;]*)m`)

// assertNoticeIsStyledDistinctly locates `notice` in `captured` and verifies
// the last non-reset SGR sequence before it carries the Bold parameter (1).
// After T7 the notice branch renders with `Bold(true).Foreground(FocusBorder)`
// so the presence of Bold is the most robust signal that the notice is
// visually distinct from the Dim keyhints row — V25 class-(b).
func assertNoticeIsStyledDistinctly(t *testing.T, captured, notice string) {
	t.Helper()
	idx := strings.Index(captured, notice)
	if idx < 0 {
		t.Fatalf("notice %q not in captured bytes — live-buffer test preconditions failed", notice)
	}
	prefix := captured[:idx]
	matches := sgrOpenRe.FindAllStringSubmatchIndex(prefix, -1)
	for i := len(matches) - 1; i >= 0; i-- {
		m := matches[i]
		params := prefix[m[2]:m[3]]
		if params == "" || params == "0" {
			continue // reset, skip — style comes from an earlier SGR
		}
		if !sgrContainsParam(params, "1") {
			t.Errorf("notice %q rendered without Bold SGR — V25 class-(b) perceptual distinctness lost.\n  preceding SGR params: %q",
				notice, params)
		}
		return
	}
	t.Errorf("no non-reset SGR sequence precedes notice %q in captured bytes — V25 class-(b) violation",
		notice)
}

// sgrContainsParam reports whether `target` appears as a standalone
// parameter in the semicolon-delimited SGR parameter list `params`.
func sgrContainsParam(params, target string) bool {
	for _, p := range strings.Split(params, ";") {
		if p == target {
			return true
		}
	}
	return false
}

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
	assertNoticeIsStyledDistinctly(t, got, clipboardNoMarksNotice)
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
	assertNoticeIsStyledDistinctly(t, got, want)
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
	// model.go formats this path as "clipboard error: <err>"; assert on
	// the full prefix so the class-(b) distinctness check has a stable
	// anchor to locate in the captured byte stream.
	const want = "clipboard error: pbcopy: broken"
	if !strings.Contains(got, want) {
		t.Errorf("clipboard-error notice %q missing from live-buffer output; captured bytes (first 500):\n%s",
			want, truncateForLog(got, 500))
		return
	}
	assertNoticeIsStyledDistinctly(t, got, want)
}

func truncateForLog(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "...(truncated)"
}
