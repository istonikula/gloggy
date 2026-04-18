package appshell

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/atotto/clipboard"

	"github.com/istonikula/gloggy/internal/logsource"
)

// ClipboardCopiedMsg is emitted after a successful clipboard write.
type ClipboardCopiedMsg struct{ Count int }

// ClipboardErrorMsg is emitted when the clipboard write fails.
type ClipboardErrorMsg struct{ Err error }

// clipboardWriteFn is the injectable sink used by CopyMarkedEntries. Tests
// override this to verify error handling and success paths without depending
// on a live system clipboard.
var clipboardWriteFn = clipboard.WriteAll

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
	if err := clipboardWriteFn(content); err != nil {
		return ClipboardCopiedMsg{}, err
	}
	return ClipboardCopiedMsg{Count: len(lines)}, nil
}

// CopyMarkedEntriesCmd wraps CopyMarkedEntries as a tea.Cmd so the y-handler
// (cavekit-app-shell R9) can surface success/error via Bubble Tea messages.
// Returns ClipboardCopiedMsg on success, ClipboardErrorMsg on write failure.
func CopyMarkedEntriesCmd(entries []logsource.Entry, markedIDs map[int]bool) tea.Cmd {
	return func() tea.Msg {
		msg, err := CopyMarkedEntries(entries, markedIDs)
		if err != nil {
			return ClipboardErrorMsg{Err: err}
		}
		return msg
	}
}
