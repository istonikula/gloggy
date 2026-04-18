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

// T-118 (F-008): Activate() leaves SearchModel in input mode.
func TestSearchModel_Activate_StartsInputMode(t *testing.T) {
	m := defaultSearch().Activate()
	if m.Mode() != SearchModeInput {
		t.Errorf("Activate should start in input mode, got %v", m.Mode())
	}
}

// T-118: Enter commits input → navigate. Requires a non-empty query.
func TestSearchModel_Enter_CommitsToNavigate(t *testing.T) {
	m := defaultSearch().Activate().SetQuery("hello", testLines)
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter}, testLines)
	if m2.Mode() != SearchModeNavigate {
		t.Errorf("Enter should switch to navigate mode, got %v", m2.Mode())
	}
	// Enter in navigate should be a no-op (not close the search).
	m3, _ := m2.Update(tea.KeyMsg{Type: tea.KeyEnter}, testLines)
	if !m3.IsActive() {
		t.Error("Enter in navigate mode should not dismiss search")
	}
	if m3.Mode() != SearchModeNavigate {
		t.Errorf("Enter in navigate should stay navigate, got %v", m3.Mode())
	}
}

// T-118: `/` while already active re-enters input mode without clearing
// the existing query so users can refine their search.
func TestSearchModel_Slash_ReentersInputMode(t *testing.T) {
	m := defaultSearch().Activate().SetQuery("hello", testLines)
	// Commit to navigate via Enter.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter}, testLines)
	if m.Mode() != SearchModeNavigate {
		t.Fatalf("precondition: want navigate, got %v", m.Mode())
	}
	// `/` should flip back to input without dropping the query.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")}, testLines)
	if m.Mode() != SearchModeInput {
		t.Errorf("after '/' in navigate: want input mode, got %v", m.Mode())
	}
	if m.Query() != "hello" {
		t.Errorf("'/' should preserve query, got %q", m.Query())
	}
}

// T-118: in navigate mode, `n` advances to the next match rather than
// being appended to the query.
func TestSearchModel_N_InNavigateMode_Advances(t *testing.T) {
	m := defaultSearch().Activate().SetQuery("hello", testLines)
	// Commit to navigate.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter}, testLines)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")}, testLines)
	if m.CurrentMatchLine() != 3 {
		t.Errorf("n in navigate should advance to line 3, got %d", m.CurrentMatchLine())
	}
	if m.Query() != "hello" {
		t.Errorf("n in navigate should NOT mutate query, got %q", m.Query())
	}
}

// T-118: in input mode, `n` is a literal query character — this preserves
// the ability to search for words containing n/N.
func TestSearchModel_N_InInputMode_AppendsToQuery(t *testing.T) {
	m := defaultSearch().Activate()
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")}, testLines)
	if m.Query() != "n" {
		t.Errorf("n in input mode should extend query to 'n', got %q", m.Query())
	}
}

// T-118: backspace only edits in input mode. In navigate mode it is a
// no-op on the query so the user does not accidentally mutate the search
// while scrolling.
func TestSearchModel_Backspace_OnlyInInputMode(t *testing.T) {
	m := defaultSearch().Activate().SetQuery("hello", testLines)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter}, testLines) // → navigate
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace}, testLines)
	if m.Query() != "hello" {
		t.Errorf("backspace in navigate should be no-op, got %q", m.Query())
	}
}

// T-119 (F-009): backspace on a multi-byte rune query must trim exactly one
// rune without corrupting UTF-8. The pre-fix code sliced a byte string with
// a rune-count index, producing invalid UTF-8 for café/日本語/emoji.
func TestSearchModel_UTF8Backspace(t *testing.T) {
	cases := []struct {
		name  string
		query string
		want  string
	}{
		{"ascii", "hello", "hell"},
		{"latin1 accent", "café", "caf"},
		{"cjk", "日本語", "日本"},
		{"emoji + ascii", "🚀x", "🚀"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := defaultSearch().Activate()
			// Type the query rune-by-rune through Update so the code path
			// matches real input handling.
			for _, r := range tc.query {
				m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}, testLines)
			}
			if m.Query() != tc.query {
				t.Fatalf("setup: Query()=%q, want %q", m.Query(), tc.query)
			}
			m, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace}, testLines)
			if m.Query() != tc.want {
				t.Errorf("after backspace: Query()=%q, want %q", m.Query(), tc.want)
			}
		})
	}
}
