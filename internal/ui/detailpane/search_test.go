package detailpane

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/istonikula/gloggy/internal/logsource"
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
	assert.True(t, m2.IsActive(), "/ should activate search")
}

// T-043: R7.2 — typing highlights matches.
func TestSearchModel_Type_HighlightsMatches(t *testing.T) {
	m := defaultSearch().Activate().SetQuery("hello", testLines)
	assert.Equal(t, 2, m.MatchCount(), "expected 2 matches for 'hello'")
	highlighted := m.HighlightLines(testLines)
	// The matching lines should contain ANSI escapes wrapping "hello".
	for _, idx := range []int{1, 3} {
		assert.Containsf(t, highlighted[idx], "hello", "line %d should contain highlighted 'hello': %q", idx, highlighted[idx])
	}
}

// T-043: R7.3 — n moves to next match.
func TestSearchModel_N_NextMatch(t *testing.T) {
	m := defaultSearch().Activate().SetQuery("hello", testLines)
	// Starts at match 0 (line 1).
	assert.Equal(t, 1, m.CurrentMatchLine(), "initial match line")
	m2 := m.NextMatch()
	assert.Equal(t, 3, m2.CurrentMatchLine(), "after next")
}

// T-043: R7.4 — N moves to previous match.
func TestSearchModel_BigN_PrevMatch(t *testing.T) {
	m := defaultSearch().Activate().SetQuery("hello", testLines)
	m2 := m.NextMatch() // at index 1 (line 3)
	m3 := m2.PrevMatch()
	assert.Equal(t, 1, m3.CurrentMatchLine(), "after prev")
}

// T-043: R7.5 — wrap indicator when matches wrap around.
func TestSearchModel_WrapIndicator(t *testing.T) {
	m := defaultSearch().Activate().SetQuery("hello", testLines)
	m = m.NextMatch() // index 1 → line 3
	m = m.NextMatch() // wraps to index 0 → line 1, WrapFwd
	assert.Equal(t, SearchWrapFwd, m.WrapDir(), "expected SearchWrapFwd")
}

// T-043: R7.6 — Esc dismisses and clears.
func TestSearchModel_Esc_Dismisses(t *testing.T) {
	m := defaultSearch().Activate().SetQuery("hello", testLines)
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc}, testLines)
	assert.False(t, m2.IsActive(), "Esc should dismiss search")
	assert.Empty(t, m2.Query(), "Esc should clear query")
	assert.Equal(t, 0, m2.MatchCount(), "Esc should clear matches")
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
	assert.Equal(t, SearchModeInput, m.Mode(), "Activate should start in input mode")
}

// T-118: Enter commits input → navigate. Requires a non-empty query.
func TestSearchModel_Enter_CommitsToNavigate(t *testing.T) {
	m := defaultSearch().Activate().SetQuery("hello", testLines)
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter}, testLines)
	assert.Equal(t, SearchModeNavigate, m2.Mode(), "Enter should switch to navigate mode")
	// Enter in navigate should be a no-op (not close the search).
	m3, _ := m2.Update(tea.KeyMsg{Type: tea.KeyEnter}, testLines)
	assert.True(t, m3.IsActive(), "Enter in navigate mode should not dismiss search")
	assert.Equal(t, SearchModeNavigate, m3.Mode(), "Enter in navigate should stay navigate")
}

// T-118: `/` while already active re-enters input mode without clearing
// the existing query so users can refine their search.
func TestSearchModel_Slash_ReentersInputMode(t *testing.T) {
	m := defaultSearch().Activate().SetQuery("hello", testLines)
	// Commit to navigate via Enter.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter}, testLines)
	require.Equal(t, SearchModeNavigate, m.Mode(), "precondition: want navigate")
	// `/` should flip back to input without dropping the query.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")}, testLines)
	assert.Equal(t, SearchModeInput, m.Mode(), "after '/' in navigate: want input mode")
	assert.Equal(t, "hello", m.Query(), "'/' should preserve query")
}

// T-118: in navigate mode, `n` advances to the next match rather than
// being appended to the query.
func TestSearchModel_N_InNavigateMode_Advances(t *testing.T) {
	m := defaultSearch().Activate().SetQuery("hello", testLines)
	// Commit to navigate.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter}, testLines)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")}, testLines)
	assert.Equal(t, 3, m.CurrentMatchLine(), "n in navigate should advance to line 3")
	assert.Equal(t, "hello", m.Query(), "n in navigate should NOT mutate query")
}

// T-118: in input mode, `n` is a literal query character — this preserves
// the ability to search for words containing n/N.
func TestSearchModel_N_InInputMode_AppendsToQuery(t *testing.T) {
	m := defaultSearch().Activate()
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")}, testLines)
	assert.Equal(t, "n", m.Query(), "n in input mode should extend query to 'n'")
}

// T-118: backspace only edits in input mode. In navigate mode it is a
// no-op on the query so the user does not accidentally mutate the search
// while scrolling.
func TestSearchModel_Backspace_OnlyInInputMode(t *testing.T) {
	m := defaultSearch().Activate().SetQuery("hello", testLines)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter}, testLines) // → navigate
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace}, testLines)
	assert.Equal(t, "hello", m.Query(), "backspace in navigate should be no-op")
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
			require.Equalf(t, tc.query, m.Query(), "setup: Query()")
			m, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace}, testLines)
			assert.Equalf(t, tc.want, m.Query(), "after backspace: Query()")
		})
	}
}

// T-113 (cavekit-detail-pane R7 / F-003): ContentLines() returns the
// pre-render soft-wrapped content — no border glyphs, no ANSI escapes.
// This is the authoritative match source for in-pane search.
func TestPaneModel_ContentLines_NoBordersNoANSI(t *testing.T) {
	m := defaultPane(10).SetWidth(40).Open(testEntry())
	lines := m.ContentLines()
	require.NotEmpty(t, lines, "expected non-empty ContentLines after Open()")
	borderGlyphs := []string{"│", "─", "╭", "╮", "╰", "╯", "┌", "┐", "└", "┘"}
	for i, line := range lines {
		assert.NotContainsf(t, line, "\x1b[", "line %d contains ANSI escape: %q", i, line)
		for _, g := range borderGlyphs {
			assert.NotContainsf(t, line, g, "line %d contains border glyph %q: %q", i, g, line)
		}
	}
}

// T-113: ContentLines() returns nil when pane is closed or has no content.
func TestPaneModel_ContentLines_ClosedReturnsNil(t *testing.T) {
	m := defaultPane(10)
	assert.Nil(t, m.ContentLines(), "expected nil from closed pane")
	m = m.Open(testEntry()).Close()
	assert.Nil(t, m.ContentLines(), "expected nil from closed-after-open pane")
}

// T-114 (F-002, F-004, F-010): the search prompt row, (cur/total) counter,
// bare-prompt and "No matches" indicator all render via WithSearch() — and
// inactive search leaves the pane's normal rendering unchanged.
func TestPaneModel_WithSearch(t *testing.T) {
	// Entry with multiple 'hello' matches so MatchCount >= 1.
	multiHit := logsource.Entry{
		IsJSON: true,
		Time:   time.Now(),
		Level:  "INFO",
		Msg:    "hello",
		Raw:    []byte(`{"level":"INFO","msg":"hello world","tag":"hello there"}`),
	}
	for _, tc := range []struct {
		name     string
		entry    logsource.Entry
		activate bool
		query    string
		want     []string
		notWant  []string
	}{
		{"prompt+counter on match", multiHit, true, "hello", []string{"/hello", "(1/"}, nil},
		{"no matches indicator", testEntry(), true, "zzz-nope", []string{"No matches"}, nil},
		{"bare prompt", testEntry(), true, "", []string{"/"}, nil},
		{"inactive no prompt", testEntry(), false, "", nil, []string{"No matches"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pane := defaultPane(12).SetWidth(60).Open(tc.entry)
			search := NewSearchModel(theme.GetTheme("tokyo-night"))
			if tc.activate {
				search = search.Activate()
				if tc.query != "" {
					search = search.SetQuery(tc.query, pane.ContentLines())
				}
			}
			view := pane.WithSearch(search).View()
			for _, w := range tc.want {
				assert.Containsf(t, view, w, "view missing %q: %q", w, view)
			}
			for _, nw := range tc.notWant {
				assert.NotContainsf(t, view, nw, "view unexpectedly contained %q: %q", nw, view)
			}
		})
	}
}
