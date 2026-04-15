package filter

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

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

	if !strings.Contains(view, "level") {
		t.Error("view missing 'level' field")
	}
	if !strings.Contains(view, "ERROR") {
		t.Error("view missing 'ERROR' pattern")
	}
	if !strings.Contains(view, "include") {
		t.Error("view missing 'include' mode")
	}
	if !strings.Contains(view, "[x]") {
		t.Error("view missing enabled indicator [x]")
	}
	if !strings.Contains(view, "msg") {
		t.Error("view missing 'msg' field")
	}
	if !strings.Contains(view, "exclude") {
		t.Error("view missing 'exclude' mode")
	}
	if !strings.Contains(view, "[ ]") {
		t.Error("view missing disabled indicator [ ]")
	}
}

// T-039: R5.2 — j/k navigates between filters
func TestPanel_JKNavigation(t *testing.T) {
	m := New(makeFS("a", "b", "c"))
	if m.cursor != 0 {
		t.Fatalf("initial cursor = %d, want 0", m.cursor)
	}

	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if m2.cursor != 1 {
		t.Errorf("after j: cursor = %d, want 1", m2.cursor)
	}

	m3, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	if m3.cursor != 0 {
		t.Errorf("after k: cursor = %d, want 0", m3.cursor)
	}
}

// T-039: R5.3 — Space toggles enabled state
func TestPanel_SpaceTogglesEnabled(t *testing.T) {
	fs := filter.NewFilterSet()
	fs.Add(filter.Filter{Field: "level", Pattern: "INFO", Mode: filter.Include, Enabled: true})
	m := New(fs)

	m2, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})
	_ = m2

	if cmd == nil {
		t.Fatal("expected a cmd after Space, got nil")
	}
	msg := cmd()
	changed, ok := msg.(FilterChangedMsg)
	if !ok {
		t.Fatalf("expected FilterChangedMsg, got %T", msg)
	}
	filters := changed.FilterSet.GetAll()
	if len(filters) == 0 {
		t.Fatal("FilterSet is empty")
	}
	if filters[0].Enabled {
		t.Error("filter should be disabled after Space toggle")
	}
}

// T-039: R5.4 — d deletes filter
func TestPanel_DDeletesFilter(t *testing.T) {
	fs := filter.NewFilterSet()
	fs.Add(filter.Filter{Field: "level", Pattern: "INFO", Mode: filter.Include, Enabled: true})
	fs.Add(filter.Filter{Field: "msg", Pattern: "hello", Mode: filter.Exclude, Enabled: true})
	m := New(fs)

	m2, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	_ = m2

	if cmd == nil {
		t.Fatal("expected cmd after d")
	}
	msg := cmd()
	changed, ok := msg.(FilterChangedMsg)
	if !ok {
		t.Fatalf("expected FilterChangedMsg, got %T", msg)
	}
	remaining := changed.FilterSet.GetAll()
	if len(remaining) != 1 {
		t.Errorf("expected 1 filter remaining, got %d", len(remaining))
	}
}

// T-039: R5.5 — FilterChangedMsg emitted on change (already tested above, explicit test)
func TestPanel_FilterChangedMsgOnChange(t *testing.T) {
	fs := filter.NewFilterSet()
	fs.Add(filter.Filter{Field: "level", Pattern: "DEBUG", Mode: filter.Include, Enabled: false})
	m := New(fs)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})
	if cmd == nil {
		t.Fatal("expected FilterChangedMsg cmd, got nil")
	}
	msg := cmd()
	if _, ok := msg.(FilterChangedMsg); !ok {
		t.Errorf("expected FilterChangedMsg, got %T", msg)
	}
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
	if m2.cursor != 2 {
		t.Errorf("after click row 2: cursor = %d, want 2", m2.cursor)
	}
}
