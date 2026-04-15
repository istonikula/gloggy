package integration

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

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
	if !pane.IsOpen() {
		t.Fatal("pane should be open")
	}

	// Simulate FieldClickMsg (normally emitted by PaneModel mouse handler).
	fieldClick := detailpane.FieldClickMsg{Field: "level", Value: "ERROR"}

	// Wire to filter prompt.
	fs := filter.NewFilterSet()
	prompt := uifilter.NewPromptModel(fs).Open(fieldClick.Field, fieldClick.Value)
	if prompt.Field() != "level" {
		t.Errorf("prompt field: got %q, want %q", prompt.Field(), "level")
	}
	if prompt.Pattern() != "ERROR" {
		t.Errorf("prompt pattern: got %q, want %q", prompt.Pattern(), "ERROR")
	}

	// Confirm the prompt (Enter).
	_, cmd := prompt.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected FilterConfirmedMsg cmd")
	}
	msg := cmd()
	confirmed, ok := msg.(uifilter.FilterConfirmedMsg)
	if !ok {
		t.Fatalf("expected FilterConfirmedMsg, got %T", msg)
	}

	// Filter should be in the set.
	filters := confirmed.FilterSet.GetAll()
	if len(filters) != 1 {
		t.Fatalf("expected 1 filter, got %d", len(filters))
	}
	if filters[0].Field != "level" || filters[0].Pattern != "ERROR" {
		t.Errorf("filter mismatch: %+v", filters[0])
	}

	// Apply filter to entry list → only ERROR entries pass.
	entries := makeFilterEntries()
	indices := filter.Apply(confirmed.FilterSet, entries)
	for _, idx := range indices {
		if entries[idx].Level != "ERROR" {
			t.Errorf("non-ERROR entry passed filter: %s", entries[idx].Level)
		}
	}
}
