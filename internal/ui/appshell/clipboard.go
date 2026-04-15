package appshell

import (
	"strings"

	"github.com/atotto/clipboard"

	"github.com/istonikula/gloggy/internal/logsource"
)

// ClipboardCopiedMsg is emitted after a successful clipboard write.
type ClipboardCopiedMsg struct{ Count int }

// ClipboardErrorMsg is emitted when the clipboard write fails.
type ClipboardErrorMsg struct{ Err error }

// CopyMarkedEntries writes marked entries to the system clipboard.
// JSONL format: one entry per line in original order.
// Non-JSON entries are written as raw text.
// If entries is empty or none are marked, this is a no-op and returns nil.
func CopyMarkedEntries(entries []logsource.Entry, markedIDs map[int]bool) (ClipboardCopiedMsg, error) {
	var lines []string
	for _, e := range entries {
		if !markedIDs[e.LineNumber] {
			continue
		}
		lines = append(lines, string(e.Raw))
	}
	if len(lines) == 0 {
		return ClipboardCopiedMsg{}, nil
	}
	content := strings.Join(lines, "\n")
	if err := clipboard.WriteAll(content); err != nil {
		return ClipboardCopiedMsg{}, err
	}
	return ClipboardCopiedMsg{Count: len(lines)}, nil
}
