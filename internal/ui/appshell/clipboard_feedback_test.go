package appshell

import (
	"bytes"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
		require.NotNil(t, cmd, "CopyMarkedEntriesCmd returned nil tea.Cmd")
		resultMsg = cmd()
	})

	require.True(t, called, "clipboardWriteFn was never invoked")
	msg, ok := resultMsg.(ClipboardCopiedMsg)
	require.Truef(t, ok, "expected ClipboardCopiedMsg, got %T", resultMsg)
	assert.Equalf(t, 2, msg.Count, "expected Count=2, got %d", msg.Count)
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
	require.Truef(t, ok, "expected ClipboardErrorMsg, got %T", resultMsg)
	assert.ErrorIsf(t, errMsg.Err, wantErr, "expected err=%v, got %v", wantErr, errMsg.Err)
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
		require.NoError(t, err)
		msg = got
	})
	assert.Equalf(t, 1, msg.Count, "expected Count=1, got %d", msg.Count)
}

// T-138: zero marks → no write, zero count, no error.
func TestCopyMarkedEntries_ZeroMarks_DoesNotWrite(t *testing.T) {
	entries := makeClipEntries(3)

	_, called := withStubClipboard(t, nil, func() {
		_, err := CopyMarkedEntries(entries, map[int]bool{})
		require.NoError(t, err)
	})
	assert.Falsef(t, called, "clipboardWriteFn should not be called with zero marks")
}

// T-138 (regression): the y-handler path must NOT use `//nolint:errcheck`.
// Checked via the source file's bytes — this test is cheap and the constraint
// is explicit in cavekit-app-shell R9.
func TestYHandler_NoNolintErrcheck(t *testing.T) {
	data, err := os.ReadFile("../app/model.go")
	require.NoErrorf(t, err, "read model.go")
	assert.Falsef(t, bytes.Contains(data, []byte("nolint:errcheck")),
		"internal/ui/app/model.go contains forbidden `//nolint:errcheck` — cavekit-app-shell R9 violation")
}

// makeClipEntries is defined in clipboard_test.go.
var _ = logsource.Entry{}
