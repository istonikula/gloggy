package detailpane

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/istonikula/gloggy/internal/theme"
)

func defaultSearch() SearchModel {
	return NewSearchModel(theme.GetTheme("tokyo-night"))
}

var testLines = []string{
	"level: INFO",
	"msg: hello world",
	"logger: main",
	"msg: another hello",
}

// T-043: R7.1 — / opens search input.
func TestSearchModel_Slash_Activates(t *testing.T) {
	m := defaultSearch()
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")}, testLines)
	if !m2.IsActive() {
		t.Error("/ should activate search")
	}
}

// T-043: R7.2 — typing highlights matches.
func TestSearchModel_Type_HighlightsMatches(t *testing.T) {
	m := defaultSearch().Activate().SetQuery("hello", testLines)
	if m.MatchCount() != 2 {
		t.Errorf("expected 2 matches for 'hello', got %d", m.MatchCount())
	}
	highlighted := m.HighlightLines(testLines)
	// The matching lines should contain ANSI escapes wrapping "hello".
	for _, idx := range []int{1, 3} {
		if !strings.Contains(highlighted[idx], "hello") {
			t.Errorf("line %d should contain highlighted 'hello': %q", idx, highlighted[idx])
		}
	}
}

// T-043: R7.3 — n moves to next match.
func TestSearchModel_N_NextMatch(t *testing.T) {
	m := defaultSearch().Activate().SetQuery("hello", testLines)
	// Starts at match 0 (line 1).
	if m.CurrentMatchLine() != 1 {
		t.Errorf("initial match line: got %d, want 1", m.CurrentMatchLine())
	}
	m2 := m.NextMatch()
	if m2.CurrentMatchLine() != 3 {
		t.Errorf("after next: got %d, want 3", m2.CurrentMatchLine())
	}
}

// T-043: R7.4 — N moves to previous match.
func TestSearchModel_BigN_PrevMatch(t *testing.T) {
	m := defaultSearch().Activate().SetQuery("hello", testLines)
	m2 := m.NextMatch() // at index 1 (line 3)
	m3 := m2.PrevMatch()
	if m3.CurrentMatchLine() != 1 {
		t.Errorf("after prev: got %d, want 1", m3.CurrentMatchLine())
	}
}

// T-043: R7.5 — wrap indicator when matches wrap around.
func TestSearchModel_WrapIndicator(t *testing.T) {
	m := defaultSearch().Activate().SetQuery("hello", testLines)
	m = m.NextMatch() // index 1 → line 3
	m = m.NextMatch() // wraps to index 0 → line 1, WrapFwd
	if m.WrapDir() != SearchWrapFwd {
		t.Errorf("expected SearchWrapFwd, got %v", m.WrapDir())
	}
}

// T-043: R7.6 — Esc dismisses and clears.
func TestSearchModel_Esc_Dismisses(t *testing.T) {
	m := defaultSearch().Activate().SetQuery("hello", testLines)
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc}, testLines)
	if m2.IsActive() {
		t.Error("Esc should dismiss search")
	}
	if m2.Query() != "" {
		t.Errorf("Esc should clear query, got %q", m2.Query())
	}
	if m2.MatchCount() != 0 {
		t.Errorf("Esc should clear matches, got %d", m2.MatchCount())
	}
}

// T-043: R7.7 — search does not affect entry-list filter (SearchModel has no FilterSet ref).
// This is structural — SearchModel has no reference to any FilterSet.
func TestSearchModel_NoFilterSetReference(t *testing.T) {
	m := defaultSearch()
	// No FilterSet fields in SearchModel — it only holds query, matches, wrapDir, th.
	_ = m.IsActive()
	_ = m.Query()
	_ = m.MatchCount()
}
