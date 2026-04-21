package integration

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/istonikula/gloggy/internal/filter"
	"github.com/istonikula/gloggy/internal/logsource"
	"github.com/istonikula/gloggy/internal/theme"
	"github.com/istonikula/gloggy/internal/ui/detailpane"
	uifilter "github.com/istonikula/gloggy/internal/ui/filter"
)

func singleEntry() logsource.Entry {
	return logsource.Entry{
		IsJSON:     true,
		LineNumber: 1,
		Level:      "ERROR",
		Msg:        "fail",
		Time:       time.Now(),
		Raw:        []byte(`{"level":"ERROR","msg":"fail"}`),
	}
}

// T-056: detail-pane/R8 — click field → FieldClickMsg with pre-filled field/value.
// T-056: filter-engine/R4 — open prompt with pre-filled values, confirm → filter added.
func TestDetailPane_FieldClick_TriggersFilterPrompt(t *testing.T) {
	th := theme.GetTheme("tokyo-night")
	entry := singleEntry()

	// Detail pane is open.
	pane := detailpane.NewPaneModel(th, 10).Open(entry)
	require.True(t, pane.IsOpen(), "pane should be open")

	// Simulate FieldClickMsg (normally emitted by PaneModel mouse handler).
	fieldClick := detailpane.FieldClickMsg{Field: "level", Value: "ERROR"}

	// Wire to filter prompt.
	fs := filter.NewFilterSet()
	prompt := uifilter.NewPromptModel(fs).Open(fieldClick.Field, fieldClick.Value)
	assert.Equal(t, "level", prompt.Field(), "prompt field")
	assert.Equal(t, "ERROR", prompt.Pattern(), "prompt pattern")

	// Confirm the prompt (Enter).
	_, cmd := prompt.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, cmd, "expected FilterConfirmedMsg cmd")
	msg := cmd()
	confirmed, ok := msg.(uifilter.FilterConfirmedMsg)
	require.Truef(t, ok, "expected FilterConfirmedMsg, got %T", msg)

	// Filter should be in the set.
	filters := confirmed.FilterSet.GetAll()
	require.Len(t, filters, 1, "expected 1 filter")
	assert.Equal(t, "level", filters[0].Field, "filter field")
	assert.Equal(t, "ERROR", filters[0].Pattern, "filter pattern")

	// Apply filter to entry list → only ERROR entries pass.
	entries := makeFilterEntries()
	indices := filter.Apply(confirmed.FilterSet, entries)
	for _, idx := range indices {
		assert.Equalf(t, "ERROR", entries[idx].Level, "non-ERROR entry passed filter: %s", entries[idx].Level)
	}
}
