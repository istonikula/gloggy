package filter

import (
	"strings"
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

// T-039 R5.1: View lists all filters; V35 Direction A format.
func TestPanel_ViewShowsAllFilters(t *testing.T) {
	fs := filter.NewFilterSet()
	fs.Add(filter.Filter{Field: "level", Pattern: "ERROR", Mode: filter.Include, Enabled: true})
	fs.Add(filter.Filter{Field: "msg", Pattern: "fail", Mode: filter.Exclude, Enabled: false})

	m := New(fs)
	view := m.View()

	assert.Contains(t, view, "level", "view missing 'level' field")
	assert.Contains(t, view, "ERROR", "view missing 'ERROR' pattern")
	assert.Contains(t, view, "+", "view missing '+' for include")
	assert.Contains(t, view, "[x]", "view missing enabled indicator [x]")
	assert.Contains(t, view, "msg", "view missing 'msg' field")
	assert.Contains(t, view, "-", "view missing '-' for exclude")
	assert.Contains(t, view, "[ ]", "view missing disabled indicator [ ]")
}

// V35: whole-line filter shown as «line»:pattern.
func TestPanel_View_WholeLineFilter_ShowsPlaceholder(t *testing.T) {
	fs := filter.NewFilterSet()
	fs.Add(filter.Filter{Field: "", Pattern: "DrawState", Mode: filter.Include, Enabled: true})
	m := New(fs)
	view := m.View()
	assert.Contains(t, view, "«line»", "whole-line filter must show «line» placeholder")
	assert.Contains(t, view, "DrawState", "pattern must appear in view")
}

// V35: Glob filter shows ·glob suffix; Regex shows ·regex; Literal omits tag.
func TestPanel_View_SyntaxTags(t *testing.T) {
	fs := filter.NewFilterSet()
	fs.Add(filter.Filter{Field: "msg", Pattern: "foo*", Mode: filter.Include, Enabled: true, Syntax: filter.Glob})
	fs.Add(filter.Filter{Field: "msg", Pattern: `\d+`, Mode: filter.Include, Enabled: true, Syntax: filter.Regex})
	fs.Add(filter.Filter{Field: "msg", Pattern: "bar", Mode: filter.Include, Enabled: true, Syntax: filter.Literal})
	m := New(fs)
	view := m.View()
	assert.Contains(t, view, "·glob", "glob filter must have ·glob tag")
	assert.Contains(t, view, "·regex", "regex filter must have ·regex tag")
	assert.NotContains(t, view, "·literal", "literal filter must not have a syntax tag")
}

// V35: empty-state copy updated.
func TestPanel_View_EmptyState_UpdatedCopy(t *testing.T) {
	m := New(filter.NewFilterSet())
	view := m.View()
	assert.Contains(t, view, "press 'a' to add", "empty state must mention 'a' key")
}

// T-039 R5.2: j/k navigates between filters.
func TestPanel_JKNavigation(t *testing.T) {
	m := New(makeFS("a", "b", "c"))
	require.Equal(t, 0, m.cursor, "initial cursor")

	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	assert.Equal(t, 1, m2.cursor, "after j")

	m3, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	assert.Equal(t, 0, m3.cursor, "after k")
}

// T-039 R5.3: Space toggles enabled state.
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

// T-039 R5.4: d deletes filter.
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

// T-039 R5.5: FilterChangedMsg emitted on change.
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

// T-039 R5.6: mouse click selects filter row.
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

// V35: `a` emits OpenPromptMsg{IsEdit: false}.
func TestPanel_AKey_EmitsOpenPromptMsg_Add(t *testing.T) {
	m := New(makeFS("a"))
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	require.NotNil(t, cmd, "expected cmd from `a`")
	msg := cmd()
	op, ok := msg.(OpenPromptMsg)
	require.True(t, ok, "expected OpenPromptMsg, got %T", msg)
	assert.False(t, op.IsEdit, "`a` must be Add mode")
}

// V35: `e` emits OpenPromptMsg{IsEdit: true, Filter: highlighted, FilterID: id}.
func TestPanel_EKey_EmitsOpenPromptMsg_Edit(t *testing.T) {
	fs := filter.NewFilterSet()
	id := fs.Add(filter.Filter{Field: "level", Pattern: "INFO", Mode: filter.Include, Enabled: true, Syntax: filter.Glob})
	m := New(fs)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	require.NotNil(t, cmd, "expected cmd from `e`")
	msg := cmd()
	op, ok := msg.(OpenPromptMsg)
	require.True(t, ok, "expected OpenPromptMsg, got %T", msg)
	assert.True(t, op.IsEdit, "`e` must be Edit mode")
	assert.Equal(t, id, op.FilterID)
	assert.Equal(t, "INFO", op.Filter.Pattern)
	assert.Equal(t, filter.Glob, op.Filter.Syntax)
}

// V35: `e` on empty panel is a no-op (no cmd).
func TestPanel_EKey_EmptyPanel_Noop(t *testing.T) {
	m := New(filter.NewFilterSet())
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	assert.Nil(t, cmd, "`e` on empty panel must emit no cmd")
}

// V35: cursor row shows > prefix.
func TestPanel_View_CursorRow_HasPrefix(t *testing.T) {
	m := New(makeFS("a", "b"))
	view := m.View()
	// cursor = 0 by default; first line should have "> "
	lines := splitLines(view)
	require.True(t, len(lines) >= 1, "expected at least 1 line")
	assert.Contains(t, lines[0], "> ", "first row (cursor) must have '> ' prefix")
	// second line has no cursor prefix
	assert.NotContains(t, lines[1], "> ", "non-cursor row must not have '> ' prefix")
}

func splitLines(s string) []string {
	return strings.Split(s, "\n")
}

