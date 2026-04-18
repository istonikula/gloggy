package detailpane

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/istonikula/gloggy/internal/theme"
)

// SearchModel manages in-pane search state: input, matches, and navigation.
// It is scoped to the detail pane content and does not affect the entry-list filter.
type SearchModel struct {
	active  bool
	query   string
	matches []int  // line indices of matching lines
	current int    // index into matches
	wrapDir WrapSearchDir
	th      theme.Theme
}

// WrapSearchDir indicates the wrap direction for search navigation.
type WrapSearchDir int

const (
	SearchNoWrap   WrapSearchDir = iota
	SearchWrapFwd
	SearchWrapBack
)

// NewSearchModel creates a SearchModel.
func NewSearchModel(th theme.Theme) SearchModel {
	return SearchModel{th: th}
}

// IsActive returns true when search mode is open.
func (m SearchModel) IsActive() bool { return m.active }

// Query returns the current search term.
func (m SearchModel) Query() string { return m.query }

// MatchCount returns the number of matching lines.
func (m SearchModel) MatchCount() int { return len(m.matches) }

// WrapDir returns the last wrap direction.
func (m SearchModel) WrapDir() WrapSearchDir { return m.wrapDir }

// CurrentMatchLine returns the line index of the current match, or -1 if none.
func (m SearchModel) CurrentMatchLine() int {
	if len(m.matches) == 0 {
		return -1
	}
	return m.matches[m.current]
}

// Activate opens the search input.
func (m SearchModel) Activate() SearchModel {
	m.active = true
	m.query = ""
	m.matches = nil
	m.current = 0
	m.wrapDir = SearchNoWrap
	return m
}

// Dismiss clears the search and closes the input.
func (m SearchModel) Dismiss() SearchModel {
	m.active = false
	m.query = ""
	m.matches = nil
	m.current = 0
	m.wrapDir = SearchNoWrap
	return m
}

// SetQuery updates the query and recomputes matches against lines.
func (m SearchModel) SetQuery(q string, lines []string) SearchModel {
	m.query = q
	m.matches = nil
	m.current = 0
	m.wrapDir = SearchNoWrap
	if q == "" {
		return m
	}
	lq := strings.ToLower(q)
	for i, line := range lines {
		if strings.Contains(strings.ToLower(line), lq) {
			m.matches = append(m.matches, i)
		}
	}
	return m
}

// NextMatch advances to the next match, wrapping with a WrapFwd indicator.
func (m SearchModel) NextMatch() SearchModel {
	if len(m.matches) == 0 {
		return m
	}
	m.wrapDir = SearchNoWrap
	next := m.current + 1
	if next >= len(m.matches) {
		next = 0
		m.wrapDir = SearchWrapFwd
	}
	m.current = next
	return m
}

// PrevMatch moves to the previous match, wrapping with a WrapBack indicator.
func (m SearchModel) PrevMatch() SearchModel {
	if len(m.matches) == 0 {
		return m
	}
	m.wrapDir = SearchNoWrap
	prev := m.current - 1
	if prev < 0 {
		prev = len(m.matches) - 1
		m.wrapDir = SearchWrapBack
	}
	m.current = prev
	return m
}

// HighlightLines applies search highlight styling to lines that match the query.
// Lines without matches are returned unchanged.
func (m SearchModel) HighlightLines(lines []string) []string {
	if m.query == "" || len(m.matches) == 0 {
		return lines
	}
	style := lipgloss.NewStyle().Foreground(m.th.SearchHighlight).Bold(true)
	lq := strings.ToLower(m.query)
	out := make([]string, len(lines))
	copy(out, lines)
	for _, idx := range m.matches {
		if idx < len(out) {
			out[idx] = highlightSubstring(out[idx], lq, style)
		}
	}
	return out
}

// highlightSubstring wraps all case-insensitive occurrences of sub in s with style.
func highlightSubstring(s, sub string, style lipgloss.Style) string {
	if sub == "" {
		return s
	}
	var sb strings.Builder
	ls := strings.ToLower(s)
	for {
		idx := strings.Index(ls, sub)
		if idx == -1 {
			sb.WriteString(s)
			break
		}
		sb.WriteString(s[:idx])
		sb.WriteString(style.Render(s[idx : idx+len(sub)]))
		s = s[idx+len(sub):]
		ls = ls[idx+len(sub):]
	}
	return sb.String()
}

// Update processes key events when the search is active.
// '/' activates, Esc dismisses, Backspace removes last char, n/N navigate.
// Printable runes are appended to the query.
// lines is the current content lines, used to recompute matches on input.
func (m SearchModel) Update(msg tea.Msg, lines []string) (SearchModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "/":
			if !m.active {
				m = m.Activate()
			}
		case "esc":
			if m.active {
				m = m.Dismiss()
			}
		case "backspace", "ctrl+h":
			if m.active && len(m.query) > 0 {
				// T-119 (F-009): rune-slice, not byte-slice. The old
				// `m.query[:len([]rune(m.query))-1]` used a rune count to
				// index a byte slice, corrupting multi-byte queries.
				runes := []rune(m.query)
				m.query = string(runes[:len(runes)-1])
				m = m.SetQuery(m.query, lines)
			}
		case "n":
			if m.active {
				m = m.NextMatch()
			}
		case "N":
			if m.active {
				m = m.PrevMatch()
			}
		default:
			if m.active && len(msg.Runes) > 0 {
				m.query += string(msg.Runes)
				m = m.SetQuery(m.query, lines)
			}
		}
	}
	return m, nil
}
