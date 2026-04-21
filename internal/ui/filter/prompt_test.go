package filter

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/istonikula/gloggy/internal/filter"
)

func defaultPrompt() PromptModel {
	return NewPromptModel(filter.NewFilterSet())
}

// T-044: R4.1 — field and pattern are pre-filled.
func TestPromptModel_Open_PreFilled(t *testing.T) {
	m := defaultPrompt().Open("level", "ERROR")
	assert.Equal(t, "level", m.Field())
	assert.Equal(t, "ERROR", m.Pattern())
	assert.True(t, m.IsActive(), "prompt should be active after Open()")
}

// T-044: R4.2 — user can choose include or exclude mode via Tab.
func TestPromptModel_Tab_TogglesMode(t *testing.T) {
	m := defaultPrompt().Open("level", "ERROR")
	assert.Equal(t, filter.Include, m.Mode(), "default mode")
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	assert.Equal(t, filter.Exclude, m2.Mode(), "after Tab")
	m3, _ := m2.Update(tea.KeyMsg{Type: tea.KeyTab})
	assert.Equal(t, filter.Include, m3.Mode(), "after second Tab")
}

// T-044: R4.3 — Enter confirms filter, adds to set, emits FilterConfirmedMsg.
func TestPromptModel_Enter_ConfirmsFilter(t *testing.T) {
	fs := filter.NewFilterSet()
	m := NewPromptModel(fs).Open("level", "ERROR")
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, cmd, "expected FilterConfirmedMsg cmd")
	msg := cmd()
	confirmed, ok := msg.(FilterConfirmedMsg)
	require.True(t, ok, "expected FilterConfirmedMsg, got %T", msg)
	assert.NotNil(t, confirmed.FilterSet, "FilterSet should not be nil")
	// Verify the filter was actually added.
	filters := fs.GetAll()
	require.Len(t, filters, 1, "expected 1 filter added")
	f := filters[0]
	assert.Equal(t, "level", f.Field)
	assert.Equal(t, "ERROR", f.Pattern)
	assert.Equal(t, filter.Include, f.Mode, "filter mode")
	assert.True(t, f.Enabled, "filter should be enabled after confirm")
}

// Esc cancels and emits FilterCancelledMsg.
func TestPromptModel_Esc_Cancels(t *testing.T) {
	m := defaultPrompt().Open("level", "ERROR")
	m2, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	assert.False(t, m2.IsActive(), "prompt should be closed after Esc")
	require.NotNil(t, cmd, "expected FilterCancelledMsg cmd")
	_, ok := cmd().(FilterCancelledMsg)
	assert.True(t, ok, "expected FilterCancelledMsg")
}

// When closed, Update is a no-op.
func TestPromptModel_Closed_Noop(t *testing.T) {
	m := defaultPrompt()
	m2, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.False(t, m2.IsActive(), "should remain closed")
	assert.Nil(t, cmd, "expected nil cmd when closed")
}
