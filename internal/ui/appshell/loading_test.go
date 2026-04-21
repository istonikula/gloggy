package appshell

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// T-048: R8.1 — loading indicator shown during background load.
func TestLoadingModel_Active_ShowsIndicator(t *testing.T) {
	m := NewLoadingModel().Start().Update(250)
	v := m.View()
	assert.Containsf(t, v, "250", "expected count in loading view: %q", v)
	assert.Truef(t, m.IsActive(), "model should be active")
}

// T-048: R8.2 — loading indicator hidden when done.
func TestLoadingModel_Done_HidesIndicator(t *testing.T) {
	m := NewLoadingModel().Start().Update(1000).Done()
	assert.Falsef(t, m.IsActive(), "model should not be active after Done()")
	assert.Emptyf(t, m.View(), "view should be empty when done: %q", m.View())
}

// Not started — view is empty.
func TestLoadingModel_NotStarted_Empty(t *testing.T) {
	m := NewLoadingModel()
	assert.Emptyf(t, m.View(), "expected empty view before start: %q", m.View())
}
