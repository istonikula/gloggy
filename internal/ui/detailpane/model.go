package detailpane

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/istonikula/gloggy/internal/logsource"
	"github.com/istonikula/gloggy/internal/theme"
	"github.com/istonikula/gloggy/internal/ui/appshell"
)

// FocusedMsg is emitted when the detail pane gains focus.
type FocusedMsg struct{}

// BlurredMsg is emitted when the detail pane closes and returns focus to entry list.
type BlurredMsg struct{}

// PaneModel is the Bubble Tea model for the detail pane.
// It manages open/close state and delegates to ScrollModel for scrolling.
type PaneModel struct {
	open       bool
	entry      logsource.Entry
	scroll     ScrollModel
	th         theme.Theme
	height     int
	width      int         // outer width allocation; 0 means content-driven (T-107)
	rawContent string      // unwrapped pre-rendered content; re-wrapped on width change (T-106)
	search     SearchModel // T-114: attached by caller via WithSearch() to drive render
	Focused    bool        // set by app before rendering for focus indicator
}

// NewPaneModel creates a PaneModel.
func NewPaneModel(th theme.Theme, height int) PaneModel {
	return PaneModel{
		th:     th,
		height: height,
		scroll: NewScrollModel("", height),
	}
}

// SetWidth updates the outer width allocation for the pane (T-107). When
// non-zero, the pane's View is constrained so its outer width equals w —
// using lipgloss cell-width measurement so emoji and CJK do not overflow.
// T-106: re-wraps the stored content at the new content width so the
// scroll viewport reflects the wrapped layout.
// T-123 (F-018): re-wrap passes ContentHeight (border-subtracted) to the
// scroll model, not the outer pane height. Passing the outer height caused
// SetContent to size the viewport to outer height which extends the visible
// window past the last renderable content row and masks clipping bugs.
func (m PaneModel) SetWidth(w int) PaneModel {
	m.width = w
	if m.open && m.rawContent != "" {
		m.scroll = m.scroll.SetContent(SoftWrap(m.rawContent, m.contentWidth()), m.ContentHeight())
	}
	return m
}

// contentWidth returns the inner width available for content after the
// border has been subtracted. Returns 0 when no width allocation is set.
func (m PaneModel) contentWidth() int {
	if m.width <= 0 {
		return 0
	}
	w := m.width - 2 // left + right border
	if w < 1 {
		return 0
	}
	return w
}

// IsOpen returns true when the pane is visible.
func (m PaneModel) IsOpen() bool { return m.open }

// Open activates the detail pane with the given entry.
// T-106: stores the raw rendered content so SetWidth can re-wrap it on
// width changes.
// T-123 (F-018): seeds the scroll viewport with ContentHeight (border-
// subtracted), so internal offset clamping stays inside the visible window.
func (m PaneModel) Open(entry logsource.Entry) PaneModel {
	var content string
	if entry.IsJSON {
		content = RenderJSON(entry, m.th, nil)
	} else {
		content = RenderRaw(entry)
	}
	m.entry = entry
	m.open = true
	m.rawContent = content
	m.scroll = NewScrollModel(SoftWrap(content, m.contentWidth()), m.ContentHeight())
	return m
}

// Close dismisses the detail pane.
func (m PaneModel) Close() PaneModel {
	m.open = false
	return m
}

// SetHeight updates the outer visible height (border-inclusive) of the pane.
// T-123 (F-014, F-018): keeps scroll.height in sync with ContentHeight so
// offset clamping bounds the viewport to the renderable content rows — not
// the outer pane including borders.
func (m PaneModel) SetHeight(h int) PaneModel {
	m.height = h
	m.scroll.height = m.ContentHeight()
	m.scroll = m.scroll.Clamp()
	return m
}

// Init satisfies tea.Model.
func (m PaneModel) Init() tea.Cmd { return nil }

// Update handles key events when the pane is open.
// Esc or Enter closes the pane and emits BlurredMsg.
// All other key/mouse events are forwarded to ScrollModel.
func (m PaneModel) Update(msg tea.Msg) (PaneModel, tea.Cmd) {
	if !m.open {
		return m, nil
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "enter":
			m = m.Close()
			return m, func() tea.Msg { return BlurredMsg{} }
		default:
			var cmd tea.Cmd
			m.scroll, cmd = m.scroll.Update(msg)
			return m, cmd
		}
	default:
		var cmd tea.Cmd
		m.scroll, cmd = m.scroll.Update(msg)
		return m, cmd
	}
}

// borderRows returns how many rows the pane border consumes. T-100 added a
// full lipgloss border in PaneStyle, so both the top and bottom borders eat
// one row each.
func (m PaneModel) borderRows() int { return 2 }

// ContentHeight returns the height available for content after subtracting borders.
func (m PaneModel) ContentHeight() int {
	h := m.height - m.borderRows()
	if h < 1 {
		h = 1
	}
	return h
}

// ScrollToLine adjusts the scroll offset so line index `idx` is visible
// in the current viewport (T-115, cavekit R7). If idx is already inside
// the window, the offset is unchanged; otherwise it scrolls the minimum
// amount needed — aligning to the top when idx is above, to the bottom
// when idx is below. Negative indices and out-of-range values are
// clamped by the underlying scroll model.
func (m PaneModel) ScrollToLine(idx int) PaneModel {
	if len(m.scroll.lines) == 0 {
		return m
	}
	viewport := m.ContentHeight()
	if m.search.IsActive() && viewport > 1 {
		viewport-- // prompt row reserves the same row as View() does.
	}
	if viewport < 1 {
		viewport = 1
	}
	if idx < m.scroll.offset {
		m.scroll.offset = idx
	} else if idx >= m.scroll.offset+viewport {
		m.scroll.offset = idx - viewport + 1
	}
	m.scroll.clamp()
	return m
}

// ContentLines returns the soft-wrapped, unstyled content lines that align
// with the pane's visual line positions — ANSI escapes from syntax
// highlighting are stripped, and the raw content is re-run through SoftWrap
// at the current contentWidth so line indices match what the user actually
// sees after the pane wraps its content. T-113 (closes F-003): splitting
// View() output would include border glyphs AND styled ANSI; splitting
// rawContent alone drops soft-wrap rows. Cavekit R7 requires matching
// against pre-syntax-highlight, post-soft-wrap text. Returns nil when the
// pane is closed or has no content.
func (m PaneModel) ContentLines() []string {
	if !m.open || m.rawContent == "" {
		return nil
	}
	wrapped := SoftWrap(m.rawContent, m.contentWidth())
	lines := strings.Split(wrapped, "\n")
	out := make([]string, len(lines))
	for i, line := range lines {
		out[i] = ansi.Strip(line)
	}
	return out
}

// WithSearch attaches a SearchModel to the pane for rendering (T-114).
// Call this from the app before View() so the pane can render the prompt
// row, match counter, and highlight matches. The SearchModel is not
// mutated by the pane — navigation (n/N) and activation happen in the
// caller's Update path.
func (m PaneModel) WithSearch(s SearchModel) PaneModel {
	m.search = s
	return m
}

// renderSearchPrompt builds the bottom prompt row shown while search is
// active. Cavekit detail-pane R7 requires the active query and (cur/total)
// to be visibly rendered; shows "No matches" when query is non-empty but
// no matches exist; appends a wrap arrow after (cur/total) when n/N wraps.
func (m PaneModel) renderSearchPrompt() string {
	q := m.search.Query()
	total := m.search.MatchCount()
	line := "/" + q
	switch {
	case q == "":
		// bare "/" — show just the prompt while the user starts typing.
	case total == 0:
		line += "  No matches"
	default:
		cur := m.search.CurrentIndex() + 1
		line += fmt.Sprintf("  (%d/%d)", cur, total)
		switch m.search.WrapDir() {
		case SearchWrapFwd:
			line += " ↓"
		case SearchWrapBack:
			line += " ↑"
		}
	}
	return lipgloss.NewStyle().Foreground(m.th.Dim).Render(line)
}

// View renders the detail pane content, or empty string when closed.
// T-082: Renders a top border separator line.
// T-100: Uses the DESIGN.md §4 pane style matrix — focused panes get
// FocusBorder + base bg, unfocused get DividerColor + UnfocusedBg + Faint.
// All four borders render in both states (T-103 verifies the top border).
func (m PaneModel) View() string {
	if !m.open {
		return ""
	}
	// Use content height (minus border rows) so total output fits allocation.
	contentH := m.ContentHeight()
	searchActive := m.search.IsActive()
	// T-114: reserve one row at the bottom of the content area for the
	// search prompt so the total pane height still matches the allocation.
	if searchActive && contentH > 1 {
		contentH--
	}

	var body string
	if searchActive && m.search.MatchCount() > 0 {
		// T-114: when there are live matches, render from the unstyled
		// soft-wrapped lines with highlight styling applied. This keeps
		// match positions in lockstep with ContentLines() (the source of
		// truth for match indices) and gives the user visible feedback.
		lines := m.search.HighlightLines(m.ContentLines())
		scroll := m.scroll
		scroll.lines = lines
		scroll.height = contentH
		scroll = scroll.Clamp()
		body = scroll.View()
	} else {
		m.scroll.height = contentH
		m.scroll = m.scroll.Clamp()
		body = m.scroll.View()
	}
	if searchActive {
		body += "\n" + m.renderSearchPrompt()
	}

	state := appshell.PaneStateUnfocused
	if m.Focused {
		state = appshell.PaneStateFocused
	}
	style := appshell.PaneStyle(m.th, state)
	if m.width > 0 {
		// Reserve 2 cells for left/right borders so the OUTER width
		// equals the allocation. lipgloss measures content in cells
		// (emoji/CJK = 2 cells, ANSI = 0) so this is width-safe.
		inner := m.width - 2
		if inner < 0 {
			inner = 0
		}
		style = style.Width(inner).MaxWidth(m.width)
	}
	return style.Render(body)
}
