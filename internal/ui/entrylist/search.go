package entrylist

import (
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"

	"github.com/istonikula/gloggy/internal/config"
	"github.com/istonikula/gloggy/internal/logsource"
	"github.com/istonikula/gloggy/internal/theme"
)

// SearchModel manages list-scope free-text search state (T-143, cavekit-
// entry-list R13). Mirrors `internal/ui/detailpane/search.go` but scans the
// compact row composition (time|level|logger|msg) rather than wrapped
// content lines. The model is value-semantic: callers (app.Update) own it
// and hand it back to `ListModel.WithSearch` before `View()`.
type SearchModel struct {
	active    bool
	query     string
	matches   []int // indices into the visible entry slice
	current   int   // index into matches
	inputMode bool  // true while typing the query; false after Enter
	notFound  bool  // query is non-empty but produced zero matches
	th        theme.Theme
}

// NewSearchModel creates a SearchModel bound to a theme.
func NewSearchModel(th theme.Theme) SearchModel { return SearchModel{th: th} }

// IsActive reports whether the search is open (activated, not yet dismissed).
func (m SearchModel) IsActive() bool { return m.active }

// InputMode reports whether the search is accepting query-typing input.
// After Enter the search transitions to navigate mode (InputMode=false) so
// the caller routes j/k/n/N through their normal handlers.
func (m SearchModel) InputMode() bool { return m.inputMode }

// Query returns the current query string.
func (m SearchModel) Query() string { return m.query }

// MatchCount returns the number of rows currently matching the query.
func (m SearchModel) MatchCount() int { return len(m.matches) }

// CurrentIndex returns the zero-based index of the current match within the
// match set. Returns 0 when the match set is empty.
func (m SearchModel) CurrentIndex() int { return m.current }

// CurrentMatchLine returns the visible-entry index of the current match, or
// -1 when the match set is empty.
func (m SearchModel) CurrentMatchLine() int {
	if len(m.matches) == 0 {
		return -1
	}
	return m.matches[m.current]
}

// NotFound reports whether the current query produced zero matches.
func (m SearchModel) NotFound() bool { return m.notFound }

// MatchLines returns the visible-entry indices for every match. Used by
// `ListModel.View()` to paint SearchHighlight bg on matched rows.
func (m SearchModel) MatchLines() []int {
	out := make([]int, len(m.matches))
	copy(out, m.matches)
	return out
}

// Activate opens the search, resetting any prior query + match state. Idempotent.
func (m SearchModel) Activate() SearchModel {
	m.active = true
	m.query = ""
	m.matches = nil
	m.current = 0
	m.inputMode = true
	m.notFound = false
	return m
}

// Deactivate closes the search and clears all state. Called on Esc, Tab
// (focus cycle), and filter changes.
func (m SearchModel) Deactivate() SearchModel {
	m.active = false
	m.query = ""
	m.matches = nil
	m.current = 0
	m.inputMode = false
	m.notFound = false
	return m
}

// CommitInput leaves input mode and enters navigate mode so n/N advance
// through the match set instead of extending the query. No-op when inactive.
func (m SearchModel) CommitInput() SearchModel {
	if !m.active {
		return m
	}
	m.inputMode = false
	return m
}

// AppendRune extends the query with `r` and recomputes matches against
// `entries`. Should be called only while `inputMode` is true.
func (m SearchModel) AppendRune(r rune, entries []logsource.Entry, width int, cfg config.Config) SearchModel {
	if !m.active || !m.inputMode {
		return m
	}
	m.query += string(r)
	return m.computeMatches(entries, width, cfg)
}

// BackspaceRune removes the last rune (UTF-8-safe) from the query and
// recomputes matches. No-op when the query is empty or search is inactive.
func (m SearchModel) BackspaceRune(entries []logsource.Entry, width int, cfg config.Config) SearchModel {
	if !m.active || !m.inputMode || m.query == "" {
		return m
	}
	_, size := utf8.DecodeLastRuneInString(m.query)
	m.query = m.query[:len(m.query)-size]
	return m.computeMatches(entries, width, cfg)
}

// Next advances to the next match, wrapping to the first on overflow. No-op
// when the match set is empty.
func (m SearchModel) Next() SearchModel {
	if len(m.matches) == 0 {
		return m
	}
	m.current = (m.current + 1) % len(m.matches)
	return m
}

// Prev moves to the previous match, wrapping to the last on underflow.
// No-op when the match set is empty.
func (m SearchModel) Prev() SearchModel {
	if len(m.matches) == 0 {
		return m
	}
	m.current = (m.current - 1 + len(m.matches)) % len(m.matches)
	return m
}

// computeMatches rebuilds the `matches` slice from `entries` using the
// compact-row composition that the user actually sees. Match is case-
// insensitive substring against the composed row text. Also resets
// `current` to 0 and sets `notFound` when query is non-empty but matched
// nothing.
func (m SearchModel) computeMatches(entries []logsource.Entry, width int, cfg config.Config) SearchModel {
	m.matches = nil
	m.current = 0
	if m.query == "" {
		m.notFound = false
		return m
	}
	needle := strings.ToLower(m.query)
	for i, e := range entries {
		hay := strings.ToLower(composeSearchRow(e, width, cfg))
		if strings.Contains(hay, needle) {
			m.matches = append(m.matches, i)
		}
	}
	m.notFound = len(m.matches) == 0
	return m
}

// composeSearchRow returns the plain-text compact row that the user sees,
// without ANSI styling. Mirrors `RenderCompactRow` field composition so
// matches align with the rendered view.
func composeSearchRow(entry logsource.Entry, width int, cfg config.Config) string {
	if !entry.IsJSON {
		raw := flattenNewlines(string(entry.Raw))
		if width > 0 && len(raw) > width {
			raw = raw[:width]
		}
		return raw
	}
	timeStr := timePlaceholder
	if !entry.Time.IsZero() {
		timeStr = entry.Time.Format(timeFormat)
	}
	lvl := padOrTrunc(strings.ToUpper(entry.Level), levelWidth)
	logger := AbbreviateLogger(entry.Logger, cfg.LoggerDepth)
	msg := flattenNewlines(entry.Msg)

	prefixLen := timeWidth + 1 + levelWidth + 1 + len(logger) + 1
	if width > 0 {
		remaining := width - prefixLen
		if remaining < 0 {
			remaining = 0
		}
		if len(msg) > remaining {
			msg = msg[:remaining]
		}
	}

	var b strings.Builder
	b.WriteString(timeStr)
	b.WriteByte(' ')
	b.WriteString(lvl)
	b.WriteByte(' ')
	b.WriteString(logger)
	b.WriteByte(' ')
	b.WriteString(msg)
	return b.String()
}

// searchHighlightStyle returns the lipgloss style used for match bg on the
// list (T-143). Cursor row keeps CursorHighlight — that takes priority in
// `ListModel.View()` — so this style is applied only to non-cursor matches.
func (m SearchModel) searchHighlightStyle() lipgloss.Style {
	return lipgloss.NewStyle().Background(m.th.SearchHighlight)
}
