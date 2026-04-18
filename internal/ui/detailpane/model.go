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
	open         bool
	entry        logsource.Entry
	scroll       ScrollModel
	th           theme.Theme
	height       int
	width        int         // outer width allocation; 0 means content-driven (T-107)
	rawContent   string      // unwrapped pre-rendered content; re-wrapped on width change (T-106)
	search       SearchModel // T-114: attached by caller via WithSearch() to drive render
	hiddenFields []string    // T-127 (F-020): field names suppressed from JSON render
	Focused      bool        // set by app before rendering for focus indicator
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

// WithScrolloff sets the scrolloff margin (context rows kept around the
// cursor during keyboard nav and mouse-wheel drag) on the inner ScrollModel.
// Wired from the app layer by reading `cfg.Scrolloff` (T-130). Shared key
// means both the list and this pane honour the same value (cavekit-config R5).
// T-132 (F-026).
func (m PaneModel) WithScrolloff(n int) PaneModel {
	m.scroll = m.scroll.WithScrolloff(n)
	return m
}

// WithHiddenFields stores the set of field names that should be suppressed
// from JSON rendering. Pass the caller's current visibility list (typically
// `VisibilityModel.HiddenFields()`) before calling Open so the set reaches
// the JSON renderer. Nil clears the suppression list. T-127 (F-020).
func (m PaneModel) WithHiddenFields(hidden []string) PaneModel {
	if hidden == nil {
		m.hiddenFields = nil
		return m
	}
	cp := make([]string, len(hidden))
	copy(cp, hidden)
	m.hiddenFields = cp
	return m
}

// Open activates the detail pane with the given entry.
// T-106: stores the raw rendered content so SetWidth can re-wrap it on
// width changes.
// T-123 (F-018): seeds the scroll viewport with ContentHeight (border-
// subtracted), so internal offset clamping stays inside the visible window.
// T-127 (F-020): passes the stored hiddenFields slice (set via
// WithHiddenFields) to RenderJSON so config-driven field hiding reaches
// the pane.
func (m PaneModel) Open(entry logsource.Entry) PaneModel {
	var content string
	if entry.IsJSON {
		content = RenderJSON(entry, m.th, m.hiddenFields)
	} else {
		content = RenderRaw(entry)
	}
	m.entry = entry
	m.open = true
	m.rawContent = content
	prevScrolloff := m.scroll.Scrolloff()
	m.scroll = NewScrollModel(SoftWrap(content, m.contentWidth()), m.ContentHeight()).
		WithScrolloff(prevScrolloff)
	return m
}

// Rerender re-renders the currently open entry with the current
// hiddenFields + width + theme, preserving the scroll offset so toggling
// visibility does not jump the viewport. No-op when the pane is closed.
// T-127 (F-020).
func (m PaneModel) Rerender() PaneModel {
	if !m.open {
		return m
	}
	prevOffset := m.scroll.offset
	var content string
	if m.entry.IsJSON {
		content = RenderJSON(m.entry, m.th, m.hiddenFields)
	} else {
		content = RenderRaw(m.entry)
	}
	m.rawContent = content
	prevScrolloff := m.scroll.Scrolloff()
	m.scroll = NewScrollModel(SoftWrap(content, m.contentWidth()), m.ContentHeight()).
		WithScrolloff(prevScrolloff)
	m.scroll.offset = prevOffset
	m.scroll = m.scroll.Clamp()
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

// ScrollToLine moves the cursor to `idx` and scrolls the viewport so the
// cursor has scrolloff context around it (T-115 / T-134 / F-026, cavekit
// R7 extended + R11). Search n/N calls this after every match update so
// the active match lands on the cursor row — paintCursorRow then renders
// it with CursorHighlight bg over the SearchHighlight fg. Negative indices
// and out-of-range values are clamped by the underlying scroll model.
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
	// Use a local model whose height reflects the search-adjusted viewport
	// so followCursor computes margins against the visible content area,
	// not the full ContentHeight (which would include the prompt row).
	s := m.scroll
	s.height = viewport
	s = s.SetCursor(idx)
	m.scroll = s
	m.scroll.height = m.ContentHeight()
	m.scroll = m.scroll.Clamp()
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

// ScrollPercent returns the scroll-position percentage (0..100) for the
// currently visible viewport, or -1 when an indicator should not be
// rendered (content fits entirely within the viewport, or pane closed).
// Exported so tests can assert without parsing the rendered view.
// T-125 (F-016).
func (m PaneModel) ScrollPercent() int {
	if !m.open {
		return -1
	}
	total := len(m.scroll.lines)
	h := m.ContentHeight()
	if m.search.IsActive() && h > 1 {
		h--
	}
	if h < 1 {
		h = 1
	}
	if total <= h {
		return -1
	}
	lastVisible := m.scroll.offset + h
	if lastVisible > total {
		lastVisible = total
	}
	pct := lastVisible * 100 / total
	if pct > 100 {
		pct = 100
	}
	return pct
}

// paintCursorRow applies CursorHighlight bg to the body line under the
// cursor (T-131, F-026, cavekit R11). Focused → Bold + full CursorHighlight;
// unfocused-but-visible → no Bold + Faint (reduced intensity per DESIGN.md
// §4 matrix). Returns body unchanged when cursor row is outside the
// visible window or pane is closed. The bg spans the pane's full content
// width so the highlight is unambiguous even on short lines.
//
// Cursor bg is applied AFTER overlayScrollIndicator so, when the cursor
// sits on the last content row, the NN% indicator keeps rendering (its Dim
// fg is preserved) and the CursorHighlight bg visually takes priority —
// per R11 AC 8 "the NN% scroll-position indicator continues to render
// independently of the cursor row — the cursor does not replace or
// displace it".
func (m PaneModel) paintCursorRow(body string, contentH int) string {
	if !m.open || contentH <= 0 {
		return body
	}
	visible := m.scroll.cursor - m.scroll.offset
	if visible < 0 || visible >= contentH {
		return body
	}
	lines := strings.Split(body, "\n")
	if visible >= len(lines) {
		return body
	}
	style := lipgloss.NewStyle().Background(m.th.CursorHighlight)
	if m.Focused {
		style = style.Bold(true)
	} else {
		style = style.Faint(true)
	}
	if cellW := m.contentWidth(); cellW > 0 {
		style = style.Inline(true).Width(cellW).MaxWidth(cellW)
	}
	lines[visible] = style.Render(lines[visible])
	return strings.Join(lines, "\n")
}

// overlayScrollIndicator right-aligns the percentage label on the last
// line of body. It does NOT add rows or columns — the indicator composes
// within the allocated content width. When the indicator is omitted
// (content fits viewport), body is returned unchanged.
// T-125 (F-016).
func (m PaneModel) overlayScrollIndicator(body string, contentH int) string {
	_ = contentH // contentH captured in ScrollPercent via ContentHeight
	pct := m.ScrollPercent()
	if pct < 0 {
		return body
	}
	label := fmt.Sprintf(" %d%%", pct)
	styled := lipgloss.NewStyle().Foreground(m.th.Dim).Render(label)
	labelW := lipgloss.Width(label)

	lines := strings.Split(body, "\n")
	if len(lines) == 0 {
		return body
	}
	lastIdx := len(lines) - 1
	lastLine := lines[lastIdx]
	lastW := lipgloss.Width(lastLine)

	cellW := m.contentWidth()
	if cellW == 0 {
		lines[lastIdx] = lastLine + styled
		return strings.Join(lines, "\n")
	}
	if lastW+labelW <= cellW {
		padN := cellW - lastW - labelW
		lines[lastIdx] = lastLine + strings.Repeat(" ", padN) + styled
	} else {
		keepW := cellW - labelW
		if keepW < 0 {
			keepW = 0
		}
		lines[lastIdx] = ansi.Truncate(lastLine, keepW, "") + styled
	}
	return strings.Join(lines, "\n")
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
	// T-125 (F-016): overlay scroll-position indicator on the last content
	// row when content exceeds the viewport. Omitted when content fits.
	body = m.overlayScrollIndicator(body, contentH)
	// T-131 (F-026): paint cursor row with CursorHighlight bg. Applied
	// AFTER indicator so the indicator is not displaced when cursor sits
	// on the last row (cavekit R11 AC 8).
	body = m.paintCursorRow(body, contentH)
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
