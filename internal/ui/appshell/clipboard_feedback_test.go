package appshell

import (
	"bytes"
	"errors"
	"os"
	"testing"

	"github.com/istonikula/gloggy/internal/logsource"
)

// withStubClipboard swaps clipboardWriteFn for the duration of fn. The stub
// captures the written content and may return err. Returns the captured text
// and whether the stub was invoked.
func withStubClipboard(t *testing.T, err error, fn func()) (captured string, called bool) {
	t.Helper()
	prev := clipboardWriteFn
	clipboardWriteFn = func(s string) error {
		captured = s
		called = true
		return err
	}
	defer func() { clipboardWriteFn = prev }()
	fn()
	return captured, called
}

// T-138: success path routes Count into ClipboardCopiedMsg via the tea.Cmd.
func TestCopyMarkedEntriesCmd_Success_EmitsCopiedMsg(t *testing.T) {
	entries := makeClipEntries(3)
	marked := map[int]bool{1: true, 3: true}

	var resultMsg interface{}
	_, called := withStubClipboard(t, nil, func() {
		cmd := CopyMarkedEntriesCmd(entries, marked)
		if cmd == nil {
			t.Fatal("CopyMarkedEntriesCmd returned nil tea.Cmd")
		}
		resultMsg = cmd()
	})

	if !called {
		t.Fatal("clipboardWriteFn was never invoked")
	}
	msg, ok := resultMsg.(ClipboardCopiedMsg)
	if !ok {
		t.Fatalf("expected ClipboardCopiedMsg, got %T", resultMsg)
	}
	if msg.Count != 2 {
		t.Errorf("expected Count=2, got %d", msg.Count)
	}
}

// T-138: error path routes the write failure into ClipboardErrorMsg.
func TestCopyMarkedEntriesCmd_WriteError_EmitsErrorMsg(t *testing.T) {
	entries := makeClipEntries(2)
	marked := map[int]bool{1: true, 2: true}
	wantErr := errors.New("xclip: not found")

	var resultMsg interface{}
	withStubClipboard(t, wantErr, func() {
		resultMsg = CopyMarkedEntriesCmd(entries, marked)()
	})

	errMsg, ok := resultMsg.(ClipboardErrorMsg)
	if !ok {
		t.Fatalf("expected ClipboardErrorMsg, got %T", resultMsg)
	}
	if !errors.Is(errMsg.Err, wantErr) {
		t.Errorf("expected err=%v, got %v", wantErr, errMsg.Err)
	}
}

// T-138: single marked entry uses the singular noun (sanity for the notice).
// Kit R9 AC phrases the notice as "copied N entries" — count=1 is the edge
// that the app layer formats as singular.
func TestCopyMarkedEntries_SingleMarked_ReturnsCountOne(t *testing.T) {
	entries := makeClipEntries(4)
	marked := map[int]bool{2: true}

	var msg ClipboardCopiedMsg
	withStubClipboard(t, nil, func() {
		got, err := CopyMarkedEntries(entries, marked)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		msg = got
	})
	if msg.Count != 1 {
		t.Errorf("expected Count=1, got %d", msg.Count)
	}
}

// T-138: zero marks → no write, zero count, no error.
func TestCopyMarkedEntries_ZeroMarks_DoesNotWrite(t *testing.T) {
	entries := makeClipEntries(3)

	var called bool
	_, got := withStubClipboard(t, nil, func() {
		_, err := CopyMarkedEntries(entries, map[int]bool{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	called = got
	if called {
		t.Error("clipboardWriteFn should not be called with zero marks")
	}
}

// T-138 (regression): the y-handler path must NOT use `//nolint:errcheck`.
// Checked via the source file's bytes — this test is cheap and the constraint
// is explicit in cavekit-app-shell R9.
func TestYHandler_NoNolintErrcheck(t *testing.T) {
	data, err := os.ReadFile("../app/model.go")
	if err != nil {
		t.Fatalf("read model.go: %v", err)
	}
	if bytes.Contains(data, []byte("nolint:errcheck")) {
		t.Error("internal/ui/app/model.go contains forbidden `//nolint:errcheck` — cavekit-app-shell R9 violation")
	}
}

// makeClipEntries is defined in clipboard_test.go.
var _ = logsource.Entry{}
