package appshell

import (
	"testing"

	"github.com/istonikula/gloggy/internal/logsource"
)

func makeClipEntries(n int) []logsource.Entry {
	entries := make([]logsource.Entry, n)
	for i := range entries {
		entries[i] = logsource.Entry{
			LineNumber: i + 1,
			IsJSON:     true,
			Raw:        []byte(`{"msg":"entry"}`),
		}
	}
	return entries
}

// T-054: R9.4 — no marks → no-op (count=0, no error).
func TestCopyMarkedEntries_NoMarks_Noop(t *testing.T) {
	entries := makeClipEntries(3)
	msg, err := CopyMarkedEntries(entries, map[int]bool{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.Count != 0 {
		t.Errorf("expected count=0 with no marks, got %d", msg.Count)
	}
}

// T-054: R9.1+R9.2 — marked entries written in original order.
// We can't assert clipboard contents directly without a real clipboard,
// but we can verify CopyMarkedEntries returns the right count and no error
// (clipboard may fail in CI without a display; skip if so).
func TestCopyMarkedEntries_MarkedEntries_Count(t *testing.T) {
	entries := makeClipEntries(5)
	marked := map[int]bool{2: true, 4: true} // LineNumbers 2 and 4

	msg, err := CopyMarkedEntries(entries, marked)
	if err != nil {
		// In CI without a display/clipboard, writing may fail — that's acceptable.
		t.Skipf("clipboard write failed (likely no display): %v", err)
	}
	if msg.Count != 2 {
		t.Errorf("expected count=2, got %d", msg.Count)
	}
}
