package appshell

import (
	"strings"
	"testing"
)

// T-048: R8.1 — loading indicator shown during background load.
func TestLoadingModel_Active_ShowsIndicator(t *testing.T) {
	m := NewLoadingModel().Start().Update(250)
	v := m.View()
	if !strings.Contains(v, "250") {
		t.Errorf("expected count in loading view: %q", v)
	}
	if !m.IsActive() {
		t.Error("model should be active")
	}
}

// T-048: R8.2 — loading indicator hidden when done.
func TestLoadingModel_Done_HidesIndicator(t *testing.T) {
	m := NewLoadingModel().Start().Update(1000).Done()
	if m.IsActive() {
		t.Error("model should not be active after Done()")
	}
	if m.View() != "" {
		t.Errorf("view should be empty when done: %q", m.View())
	}
}

// Not started — view is empty.
func TestLoadingModel_NotStarted_Empty(t *testing.T) {
	m := NewLoadingModel()
	if m.View() != "" {
		t.Errorf("expected empty view before start: %q", m.View())
	}
}
