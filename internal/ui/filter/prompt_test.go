package filter

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/istonikula/gloggy/internal/filter"
)

func defaultPrompt() PromptModel {
	return NewPromptModel(filter.NewFilterSet())
}

// T-044: R4.1 — field and pattern are pre-filled.
func TestPromptModel_Open_PreFilled(t *testing.T) {
	m := defaultPrompt().Open("level", "ERROR")
	if m.Field() != "level" {
		t.Errorf("field: got %q, want %q", m.Field(), "level")
	}
	if m.Pattern() != "ERROR" {
		t.Errorf("pattern: got %q, want %q", m.Pattern(), "ERROR")
	}
	if !m.IsActive() {
		t.Error("prompt should be active after Open()")
	}
}

// T-044: R4.2 — user can choose include or exclude mode via Tab.
func TestPromptModel_Tab_TogglesMode(t *testing.T) {
	m := defaultPrompt().Open("level", "ERROR")
	if m.Mode() != filter.Include {
		t.Errorf("default mode: got %v, want Include", m.Mode())
	}
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m2.Mode() != filter.Exclude {
		t.Errorf("after Tab: got %v, want Exclude", m2.Mode())
	}
	m3, _ := m2.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m3.Mode() != filter.Include {
		t.Errorf("after second Tab: got %v, want Include", m3.Mode())
	}
}

// T-044: R4.3 — Enter confirms filter, adds to set, emits FilterConfirmedMsg.
func TestPromptModel_Enter_ConfirmsFilter(t *testing.T) {
	fs := filter.NewFilterSet()
	m := NewPromptModel(fs).Open("level", "ERROR")
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected FilterConfirmedMsg cmd")
	}
	msg := cmd()
	confirmed, ok := msg.(FilterConfirmedMsg)
	if !ok {
		t.Fatalf("expected FilterConfirmedMsg, got %T", msg)
	}
	if confirmed.FilterSet == nil {
		t.Error("FilterSet should not be nil")
	}
	// Verify the filter was actually added.
	filters := fs.GetAll()
	if len(filters) != 1 {
		t.Fatalf("expected 1 filter added, got %d", len(filters))
	}
	f := filters[0]
	if f.Field != "level" || f.Pattern != "ERROR" {
		t.Errorf("filter mismatch: %+v", f)
	}
	if f.Mode != filter.Include {
		t.Errorf("filter mode: got %v, want Include", f.Mode)
	}
	if !f.Enabled {
		t.Error("filter should be enabled after confirm")
	}
}

// Esc cancels and emits FilterCancelledMsg.
func TestPromptModel_Esc_Cancels(t *testing.T) {
	m := defaultPrompt().Open("level", "ERROR")
	m2, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m2.IsActive() {
		t.Error("prompt should be closed after Esc")
	}
	if cmd == nil {
		t.Fatal("expected FilterCancelledMsg cmd")
	}
	if _, ok := cmd().(FilterCancelledMsg); !ok {
		t.Error("expected FilterCancelledMsg")
	}
}

// When closed, Update is a no-op.
func TestPromptModel_Closed_Noop(t *testing.T) {
	m := defaultPrompt()
	m2, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m2.IsActive() {
		t.Error("should remain closed")
	}
	if cmd != nil {
		t.Error("expected nil cmd when closed")
	}
}
