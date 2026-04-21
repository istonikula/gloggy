package filter

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/istonikula/gloggy/internal/filter"
)

func makeFS(fields ...string) *filter.FilterSet {
	fs := filter.NewFilterSet()
	for i, field := range fields {
		fs.Add(filter.Filter{
			Field:   field,
			Pattern: "test" + field,
			Mode:    filter.Include,
			Enabled: i%2 == 0, // alternate enabled/disabled
		})
	}
	return fs
}

// T-039: R5.1 — View lists all filters with field, pattern, mode, enabled state
func TestPanel_ViewShowsAllFilters(t *testing.T) {
	fs := filter.NewFilterSet()
	fs.Add(filter.Filter{Field: "level", Pattern: "ERROR", Mode: filter.Include, Enabled: true})
	fs.Add(filter.Filter{Field: "msg", Pattern: "fail", Mode: filter.Exclude, Enabled: false})

	m := New(fs)
	view := m.View()

	assert.Contains(t, view, "level", "view missing 'level' field")
	assert.Contains(t, view, "ERROR", "view missing 'ERROR' pattern")
	assert.Contains(t, view, "include", "view missing 'include' mode")
	assert.Contains(t, view, "[x]", "view missing enabled indicator [x]")
	assert.Contains(t, view, "msg", "view missing 'msg' field")
	assert.Contains(t, view, "exclude", "view missing 'exclude' mode")
	assert.Contains(t, view, "[ ]", "view missing disabled indicator [ ]")
}

// T-039: R5.2 — j/k navigates between filters
func TestPanel_JKNavigation(t *testing.T) {
	m := New(makeFS("a", "b", "c"))
	require.Equal(t, 0, m.cursor, "initial cursor")

	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	assert.Equal(t, 1, m2.cursor, "after j")

	m3, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	assert.Equal(t, 0, m3.cursor, "after k")
}

// T-039: R5.3 — Space toggles enabled state
func TestPanel_SpaceTogglesEnabled(t *testing.T) {
	fs := filter.NewFilterSet()
	fs.Add(filter.Filter{Field: "level", Pattern: "INFO", Mode: filter.Include, Enabled: true})
	m := New(fs)

	m2, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})
	_ = m2

	require.NotNil(t, cmd, "expected a cmd after Space")
	msg := cmd()
	changed, ok := msg.(FilterChangedMsg)
	require.True(t, ok, "expected FilterChangedMsg, got %T", msg)
	filters := changed.FilterSet.GetAll()
	require.NotEmpty(t, filters, "FilterSet is empty")
	assert.False(t, filters[0].Enabled, "filter should be disabled after Space toggle")
}

// T-039: R5.4 — d deletes filter
func TestPanel_DDeletesFilter(t *testing.T) {
	fs := filter.NewFilterSet()
	fs.Add(filter.Filter{Field: "level", Pattern: "INFO", Mode: filter.Include, Enabled: true})
	fs.Add(filter.Filter{Field: "msg", Pattern: "hello", Mode: filter.Exclude, Enabled: true})
	m := New(fs)

	m2, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	_ = m2

	require.NotNil(t, cmd, "expected cmd after d")
	msg := cmd()
	changed, ok := msg.(FilterChangedMsg)
	require.True(t, ok, "expected FilterChangedMsg, got %T", msg)
	remaining := changed.FilterSet.GetAll()
	assert.Len(t, remaining, 1, "expected 1 filter remaining")
}

// T-039: R5.5 — FilterChangedMsg emitted on change (already tested above, explicit test)
func TestPanel_FilterChangedMsgOnChange(t *testing.T) {
	fs := filter.NewFilterSet()
	fs.Add(filter.Filter{Field: "level", Pattern: "DEBUG", Mode: filter.Include, Enabled: false})
	m := New(fs)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})
	require.NotNil(t, cmd, "expected FilterChangedMsg cmd")
	msg := cmd()
	_, ok := msg.(FilterChangedMsg)
	assert.True(t, ok, "expected FilterChangedMsg, got %T", msg)
}

// T-039: R5.6 — mouse click selects filter row
func TestPanel_MouseClickSelectsRow(t *testing.T) {
	m := New(makeFS("a", "b", "c"))
	m2, _ := m.Update(tea.MouseMsg{
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
		X:      0,
		Y:      2,
	})
	assert.Equal(t, 2, m2.cursor, "after click row 2")
}
