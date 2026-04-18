package entrylist

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/istonikula/gloggy/internal/config"
	"github.com/istonikula/gloggy/internal/logsource"
	"github.com/istonikula/gloggy/internal/theme"
)

// bgEscape returns the truecolor ANSI bg prefix lipgloss actually emits
// for a given color — e.g. "48;2;54;73;130" or "48;5;17" depending on
// the active termenv profile. Derived by rendering a probe string
// and extracting the bg sequence, so it matches the renderer byte-for-
// byte regardless of sRGB rounding inside lipgloss.
func bgEscape(c lipgloss.Color) string {
	probe := lipgloss.NewStyle().Background(c).Render("X")
	re := regexp.MustCompile(`\x1b\[(48;[0-9;]+)m`)
	m := re.FindStringSubmatch(probe)
	if m == nil {
		return ""
	}
	return m[1]
}

// entriesWithMessages produces N JSON entries with the supplied messages.
// All entries are INFO level so search queries only match the message
// column, not the level column (cavekit-entry-list R13 AC 2 matches
// against the full compact row — level+logger included).
func entriesWithMessages(msgs ...string) []logsource.Entry {
	entries := make([]logsource.Entry, len(msgs))
	for i, msg := range msgs {
		entries[i] = logsource.Entry{
			IsJSON:     true,
			LineNumber: i + 1,
			Time:       time.Date(2026, 4, 18, 10, 0, i, 0, time.UTC),
			Level:      "INFO",
			Logger:     "com.example.svc",
			Msg:        msg,
			Raw:        []byte(fmt.Sprintf(`{"msg":%q}`, msg)),
		}
	}
	return entries
}

// T-143 (cavekit-entry-list R13 AC 2): typed query matches case-
// insensitively against compact-row text of visible entries.
func TestSearchModel_ActivateAndType_MatchesSubstrings(t *testing.T) {
	cfg := config.DefaultConfig()
	th := theme.GetTheme("tokyo-night")
	entries := entriesWithMessages(
		"benign info",
		"error connecting to db",
		"warning timeout",
		"another ERROR occurred",
		"clean",
	)

	s := NewSearchModel(th).Activate()
	for _, r := range "err" {
		s = s.AppendRune(r, entries, 80, cfg)
	}

	if !s.IsActive() {
		t.Fatal("search should be active after Activate()")
	}
	if s.Query() != "err" {
		t.Errorf("query: got %q, want %q", s.Query(), "err")
	}
	got := s.MatchLines()
	want := []int{1, 3} // "error ..." and "another ERROR ..." (message column only)
	if len(got) != len(want) {
		t.Fatalf("matches: got %v, want %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("match[%d]: got %d, want %d", i, got[i], want[i])
		}
	}
	if s.NotFound() {
		t.Error("NotFound should be false when matches exist")
	}
}

// T-143 (cavekit-entry-list R13 AC 3): query yielding zero matches sets
// NotFound=true so the caller can render a "No matches" indicator.
func TestSearchModel_NoMatches_SetsNotFound(t *testing.T) {
	cfg := config.DefaultConfig()
	th := theme.GetTheme("tokyo-night")
	entries := entriesWithMessages("alpha", "beta", "gamma")

	s := NewSearchModel(th).Activate()
	for _, r := range "zzz" {
		s = s.AppendRune(r, entries, 80, cfg)
	}

	if s.MatchCount() != 0 {
		t.Errorf("matches: got %d, want 0", s.MatchCount())
	}
	if !s.NotFound() {
		t.Error("NotFound should be true when non-empty query matches nothing")
	}
}

// T-143 (cavekit-entry-list R13 AC 4): Next/Prev cycle through matches
// with wrap semantics.
func TestSearchModel_NextPrev_CyclesWithWrap(t *testing.T) {
	cfg := config.DefaultConfig()
	th := theme.GetTheme("tokyo-night")
	entries := entriesWithMessages(
		"error a",
		"benign",
		"error b",
		"error c",
	)

	s := NewSearchModel(th).Activate()
	for _, r := range "error" {
		s = s.AppendRune(r, entries, 80, cfg)
	}
	if s.MatchCount() != 3 {
		t.Fatalf("precondition: want 3 matches, got %d", s.MatchCount())
	}

	// Initially current=0 → first match line is 0.
	if got := s.CurrentMatchLine(); got != 0 {
		t.Errorf("initial CurrentMatchLine: got %d, want 0", got)
	}
	s = s.Next()
	if got := s.CurrentMatchLine(); got != 2 {
		t.Errorf("after 1 Next: got %d, want 2", got)
	}
	s = s.Next()
	if got := s.CurrentMatchLine(); got != 3 {
		t.Errorf("after 2 Next: got %d, want 3", got)
	}
	// Wrap forward.
	s = s.Next()
	if got := s.CurrentMatchLine(); got != 0 {
		t.Errorf("after wrap-forward: got %d, want 0", got)
	}
	// Wrap back.
	s = s.Prev()
	if got := s.CurrentMatchLine(); got != 3 {
		t.Errorf("after wrap-back: got %d, want 3", got)
	}
}

// T-143 (cavekit-entry-list R13 AC 9): Backspace on a query containing a
// multi-byte rune (emoji) removes exactly one rune without corrupting
// UTF-8.
func TestSearchModel_BackspaceRune_UTF8Safe(t *testing.T) {
	cfg := config.DefaultConfig()
	th := theme.GetTheme("tokyo-night")
	entries := entriesWithMessages("alpha")

	s := NewSearchModel(th).Activate()
	s = s.AppendRune('a', entries, 80, cfg)
	s = s.AppendRune('🔥', entries, 80, cfg)
	s = s.AppendRune('b', entries, 80, cfg)
	if s.Query() != "a🔥b" {
		t.Fatalf("precondition: query = %q", s.Query())
	}

	s = s.BackspaceRune(entries, 80, cfg)
	if s.Query() != "a🔥" {
		t.Errorf("after 1 backspace: got %q, want %q", s.Query(), "a🔥")
	}
	s = s.BackspaceRune(entries, 80, cfg)
	if s.Query() != "a" {
		t.Errorf("after 2 backspace (multibyte rune): got %q, want %q", s.Query(), "a")
	}
	s = s.BackspaceRune(entries, 80, cfg)
	if s.Query() != "" {
		t.Errorf("after 3 backspace: got %q, want empty", s.Query())
	}
}

// T-143: Deactivate clears all state so a later Activate starts fresh.
func TestSearchModel_Deactivate_ClearsState(t *testing.T) {
	cfg := config.DefaultConfig()
	th := theme.GetTheme("tokyo-night")
	entries := entriesWithMessages("error a", "error b")

	s := NewSearchModel(th).Activate()
	for _, r := range "error" {
		s = s.AppendRune(r, entries, 80, cfg)
	}
	s = s.Next()

	s = s.Deactivate()

	if s.IsActive() {
		t.Error("after Deactivate: IsActive should be false")
	}
	if s.Query() != "" {
		t.Errorf("after Deactivate: query should be empty, got %q", s.Query())
	}
	if s.MatchCount() != 0 {
		t.Errorf("after Deactivate: matches should be empty, got %d", s.MatchCount())
	}
	if s.CurrentIndex() != 0 {
		t.Errorf("after Deactivate: current should be 0, got %d", s.CurrentIndex())
	}
	if s.NotFound() {
		t.Error("after Deactivate: NotFound should be false")
	}
}

// T-143 (list integration): ActivateSearch + typed query produces a
// visible SearchHighlight band on non-cursor match rows. The cursor
// row keeps its CursorHighlight bg (R1, R12, R13 AC 10 priority).
func TestListModel_View_SearchHighlightsNonCursorMatches(t *testing.T) {
	cfg := config.DefaultConfig()
	th := theme.GetTheme("tokyo-night")
	entries := entriesWithMessages(
		"quiet line",       // index 0 (cursor starts here)
		"error one",        // index 1 — non-cursor match
		"benign again",     // index 2
		"error two",        // index 3 — non-cursor match
	)
	m := NewListModel(th, cfg, 80, 10).SetEntries(entries)
	m.Focused = true
	m = m.ActivateSearch()
	// Type "error" using Update so we exercise the full key-routing path.
	for _, r := range []rune("error") {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	if !m.HasActiveSearch() {
		t.Fatal("precondition: list search should be active")
	}
	if got := m.Search().MatchCount(); got != 2 {
		t.Fatalf("want 2 matches, got %d (matches=%v)", got, m.Search().MatchLines())
	}

	view := m.View()
	// Cursor row is row 0 ("quiet line") — not a match, no highlight
	// expected there. Rows 1 and 3 are non-cursor matches — they must
	// render with SearchHighlight bg. Match the ANSI truecolor bg
	// escape (48;2;R;G;B) that lipgloss emits for the color.
	bgCode := bgEscape(th.SearchHighlight)
	if bgCode == "" {
		t.Fatalf("could not derive bg escape for %q", th.SearchHighlight)
	}
	occurrences := strings.Count(view, bgCode)
	if occurrences < 2 {
		t.Errorf("expected SearchHighlight bg escape %q in view at least 2x, got %d occurrences\nview=\n%s",
			bgCode, occurrences, view)
	}
}

// T-143 (cavekit-entry-list R13 AC 10): on the active match row (which
// is also the cursor row after Enter+n), the cursor highlight must take
// visual priority — SearchHighlight must NOT appear on that row.
func TestListModel_View_CursorRowPriorityOverSearchHighlight(t *testing.T) {
	cfg := config.DefaultConfig()
	th := theme.GetTheme("tokyo-night")
	entries := entriesWithMessages(
		"error one",
		"quiet",
		"error two",
		"quiet",
	)
	m := NewListModel(th, cfg, 80, 10).SetEntries(entries)
	m.Focused = true
	m = m.ActivateSearch()
	for _, r := range []rune("error") {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	// Commit to navigate mode, then Next — jumps cursor to first match.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.Cursor() != 0 {
		t.Fatalf("after Enter cursor should be at first match (0), got %d", m.Cursor())
	}

	// The view wraps the list in a lipgloss border, so the first
	// rendered data row is lines[1]. Verify that the cursor row — the
	// one containing "error one" — carries CursorHighlight (not
	// SearchHighlight). The second match row ("error two") must carry
	// SearchHighlight instead.
	view := m.View()
	lines := strings.Split(view, "\n")
	var cursorRow, secondMatchRow string
	for _, ln := range lines {
		if strings.Contains(ln, "error one") && cursorRow == "" {
			cursorRow = ln
		}
		if strings.Contains(ln, "error two") && secondMatchRow == "" {
			secondMatchRow = ln
		}
	}
	if cursorRow == "" {
		t.Fatalf("cursor row with 'error one' not found in view:\n%s", view)
	}
	if secondMatchRow == "" {
		t.Fatalf("second match row 'error two' not found in view:\n%s", view)
	}
	cursorBg := bgEscape(th.CursorHighlight)
	searchBg := bgEscape(th.SearchHighlight)
	if !strings.Contains(cursorRow, cursorBg) {
		t.Errorf("cursor row should contain CursorHighlight bg %q\ncursorRow=%q",
			cursorBg, cursorRow)
	}
	if strings.Contains(cursorRow, searchBg) {
		t.Errorf("cursor row must NOT contain SearchHighlight bg (priority violation)\ncursorRow=%q",
			cursorRow)
	}
	if !strings.Contains(secondMatchRow, searchBg) {
		t.Errorf("non-cursor match row should contain SearchHighlight bg %q\nrow=%q",
			searchBg, secondMatchRow)
	}
}

// T-143 (cavekit-entry-list R13 AC): SetFilter clears active search
// because the filtered set invalidates any pre-computed match indices.
func TestListModel_SetFilter_ClearsActiveSearch(t *testing.T) {
	cfg := config.DefaultConfig()
	th := theme.GetTheme("tokyo-night")
	entries := entriesWithMessages("error a", "info b", "error c")
	m := NewListModel(th, cfg, 80, 10).SetEntries(entries)
	m = m.ActivateSearch()
	for _, r := range []rune("error") {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	if !m.HasActiveSearch() {
		t.Fatal("precondition: search should be active")
	}

	m = m.SetFilter([]int{0, 2})

	if m.HasActiveSearch() {
		t.Error("SetFilter should clear active search")
	}
}

// T-143 (cavekit-entry-list R13 AC 5): Next/Prev honour scrolloff so
// the cursor lands with R12 context rows above/below where the filtered
// entry set permits.
func TestListModel_SearchNext_HonoursScrolloff(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Scrolloff = 3
	th := theme.GetTheme("tokyo-night")
	// 30 entries. The word "target" only appears at index 20 so Next
	// jumps the cursor there directly.
	msgs := make([]string, 30)
	for i := range msgs {
		msgs[i] = "filler " + fmt.Sprintf("%d", i)
	}
	msgs[20] = "target hit"
	entries := entriesWithMessages(msgs...)
	// Viewport height 10, scrolloff 3.
	m := NewListModel(th, cfg, 80, 10).SetEntries(entries)
	m = m.ActivateSearch()
	for _, r := range []rune("target") {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	if got := m.Search().MatchCount(); got != 1 {
		t.Fatalf("want 1 match, got %d", got)
	}
	// Commit to navigate, then Next lands cursor on index 20.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.Cursor() != 20 {
		t.Errorf("Enter should land on first match (20), got %d", m.Cursor())
	}
	// Viewport should respect scrolloff=3 from the top edge.
	if m.scroll.Offset > 20-3 {
		t.Errorf("scrolloff not honoured: cursor=20 but offset=%d (want ≤ 17)", m.scroll.Offset)
	}
}

// T-143 (cavekit-entry-list R13 AC 8): list search must NOT modify the
// filter engine — entries that do not match remain visible (unhighlighted).
func TestListModel_Search_DoesNotChangeVisibleSet(t *testing.T) {
	cfg := config.DefaultConfig()
	th := theme.GetTheme("tokyo-night")
	entries := entriesWithMessages("alpha", "error", "beta", "gamma")
	m := NewListModel(th, cfg, 80, 10).SetEntries(entries)
	before := m.RenderedRowCount()

	m = m.ActivateSearch()
	for _, r := range []rune("error") {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	after := m.RenderedRowCount()

	if before != after {
		t.Errorf("list search changed visible-row count: before=%d after=%d", before, after)
	}
}
