package filter

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/istonikula/gloggy/internal/filter"
)

// V35 AMEND / B37: when the edited filter is removed between OpenEdit and the
// Enter commit, FilterSet.Update returns false. The prompt must surface
// FilterRejectedMsg{"filter not found"} and stay open — never close clean with
// no mutation and no signal.
func TestPromptModel_EditCommit_IdRemoved_KeepsOpen_B37(t *testing.T) {
	fs := filter.NewFilterSet()
	id := fs.Add(filter.Filter{
		Field: "msg", Pattern: "orig", Syntax: filter.Literal, Mode: filter.Include, Enabled: true,
	})

	m := NewPromptModel(fs).OpenEdit(filter.Filter{
		Field: "msg", Pattern: "edited", Syntax: filter.Literal, Mode: filter.Include,
	}, id)
	require.True(t, m.IsActive(), "precondition: prompt open in edit mode")

	// Filter deleted out from under the prompt (e.g. panel `d` in another frame).
	require.True(t, fs.Remove(id), "precondition: filter removed")

	m2, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	assert.True(t, m2.IsActive(), "prompt must stay open when the edited filter is gone")
	require.NotNil(t, cmd, "expected a FilterRejectedMsg cmd")
	msg, ok := cmd().(FilterRejectedMsg)
	require.True(t, ok, "expected FilterRejectedMsg, got %T", cmd())
	assert.Equal(t, "filter not found", msg.Reason)
}

// Control: a valid edit (id still present) commits and closes, emitting
// FilterConfirmedMsg{IsNew:false}. Guards against the B37 fix rejecting good
// edits.
func TestPromptModel_EditCommit_IdPresent_Commits(t *testing.T) {
	fs := filter.NewFilterSet()
	id := fs.Add(filter.Filter{
		Field: "msg", Pattern: "orig", Syntax: filter.Literal, Mode: filter.Include, Enabled: true,
	})

	m := NewPromptModel(fs).OpenEdit(filter.Filter{
		Field: "msg", Pattern: "edited", Syntax: filter.Literal, Mode: filter.Include,
	}, id)

	m2, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	assert.False(t, m2.IsActive(), "prompt closes on a successful edit commit")
	require.NotNil(t, cmd)
	msg, ok := cmd().(FilterConfirmedMsg)
	require.True(t, ok, "expected FilterConfirmedMsg, got %T", cmd())
	assert.False(t, msg.IsNew, "edit commit is not a new filter")
	assert.Equal(t, "edited", fs.GetAll()[0].Pattern, "Update applied the edited pattern")
}
