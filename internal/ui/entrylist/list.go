package entrylist

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/istonikula/gloggy/internal/config"
	"github.com/istonikula/gloggy/internal/logsource"
	"github.com/istonikula/gloggy/internal/theme"
)

// SelectionMsg is emitted when the cursor moves to a new entry.
type SelectionMsg struct {
	Entry logsource.Entry
}

// WrapIndicator is emitted when a level-jump or mark navigation wraps around.
type WrapIndicator struct {
	Direction WrapDirection
}

// OpenDetailPaneMsg is emitted when the user double-clicks an entry to open the detail pane.
type OpenDetailPaneMsg struct {
	Entry logsource.Entry
}

// ListModel is the virtual-rendering entry list Bubble Tea model.
// It only renders visible rows plus a small buffer, regardless of total entry count.
type ListModel struct {
	entries     []logsource.Entry
	filtered    []int // filtered indices; nil means show all entries
	scroll      ScrollState
	th          theme.Theme
	cfg         config.Config
	width       int
	marks         *MarkSet
	wrapDir       WrapDirection // last wrap direction, reset on next navigation
	lastClickRow  int
	lastClickTime time.Time
}

// NewListModel creates a ListModel.
func NewListModel(th theme.Theme, cfg config.Config, width, height int) ListModel {
	return ListModel{
		th:     th,
		cfg:    cfg,
		width:  width,
		marks:  NewMarkSet(),
		scroll: ScrollState{ViewportHeight: height},
	}
}

// SetEntries replaces the entry list and resets scroll state.
func (m ListModel) SetEntries(entries []logsource.Entry) ListModel {
	m.entries = entries
	m.scroll.TotalEntries = len(entries)
	m.scroll.Cursor = 0
	m.scroll.Offset = 0
	return m
}

// AppendEntries adds entries (used during background loading).
func (m ListModel) AppendEntries(entries []logsource.Entry) ListModel {
	m.entries = append(m.entries, entries...)
	m.scroll.TotalEntries = len(m.entries)
	return m
}

// Cursor returns the current cursor index into entries.
func (m ListModel) Cursor() int { return m.scroll.Cursor }

// CursorPosition returns 1-based cursor position within visible set, or 0 when empty.
func (m ListModel) CursorPosition() int {
	vis := m.visibleEntries()
	if len(vis) == 0 {
		return 0
	}
	return m.scroll.Cursor + 1
}

// SelectedEntry returns the entry at the cursor, or zero value if empty.
func (m ListModel) SelectedEntry() (logsource.Entry, bool) {
	if len(m.entries) == 0 || m.scroll.Cursor >= len(m.entries) {
		return logsource.Entry{}, false
	}
	return m.entries[m.scroll.Cursor], true
}

// SetFilter applies a filtered index (slice of entry indices that pass filters).
// On filter change, keeps selection if still passing, else moves to nearest.
func (m ListModel) SetFilter(indices []int) ListModel {
	// Check if current cursor entry still passes.
	current := m.scroll.Cursor
	found := false
	nearest := 0
	for i, idx := range indices {
		if idx == current {
			found = true
			nearest = i
			break
		}
		if idx <= current {
			nearest = i
		}
	}
	m.filtered = indices
	if found {
		m.scroll.Cursor = current
	} else if len(indices) > 0 {
		m.scroll.Cursor = indices[nearest]
	}
	m.scroll.TotalEntries = len(indices)
	m.scroll.Offset = clampOffset(m.scroll.Offset, len(indices), m.scroll.ViewportHeight)
	return m
}

// ClearFilter removes the filter, showing all entries.
func (m ListModel) ClearFilter() ListModel {
	m.filtered = nil
	m.scroll.TotalEntries = len(m.entries)
	return m
}

// ClearTransient clears transient in-list state — currently the mark-nav /
// level-jump wrap indicator. Invoked when the list receives Esc with no
// higher-priority dismissal pending (T-097).
func (m ListModel) ClearTransient() ListModel {
	m.wrapDir = NoWrap
	return m
}

// HasTransient reports whether transient state is set (wrap indicator active).
func (m ListModel) HasTransient() bool {
	return m.wrapDir != NoWrap
}

// visibleEntries returns the entries to display (filtered or all).
func (m ListModel) visibleEntries() []logsource.Entry {
	if m.filtered == nil {
		return m.entries
	}
	out := make([]logsource.Entry, len(m.filtered))
	for i, idx := range m.filtered {
		out[i] = m.entries[idx]
	}
	return out
}

// Marks returns the mark set.
func (m ListModel) Marks() *MarkSet { return m.marks }

// rowForY converts a mouse Y coordinate (relative to the list top) to a
// visible-list cursor index, or -1 if out of bounds.
func (m ListModel) rowForY(y int) int {
	vis := m.visibleEntries()
	idx := m.scroll.Offset + y
	if idx < 0 || idx >= len(vis) {
		return -1
	}
	return idx
}

// Init satisfies tea.Model.
func (m ListModel) Init() tea.Cmd { return nil }

// Update handles keyboard navigation.
func (m ListModel) Update(msg tea.Msg) (ListModel, tea.Cmd) {
	// Always process WindowSizeMsg, even when entries are empty.
	// The initial resize arrives before async loading finishes.
	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		m.scroll.ViewportHeight = ws.Height
		m.width = ws.Width
		return m, nil
	}

	vis := m.visibleEntries()
	n := len(vis)
	if n == 0 {
		return m, nil
	}

	var cmd tea.Cmd
	prev := m.scroll.Cursor
	m.wrapDir = NoWrap

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			m.scroll.Cursor = clampCursor(m.scroll.Cursor+1, n)
		case "k", "up":
			m.scroll.Cursor = clampCursor(m.scroll.Cursor-1, n)
		case "g":
			m.scroll = GoTop(m.scroll)
		case "G":
			m.scroll = GoBottom(m.scroll)
		case "ctrl+d":
			m.scroll = HalfPageDown(m.scroll)
		case "ctrl+u":
			m.scroll = HalfPageUp(m.scroll)
		case "m":
			// Toggle mark on current visible entry.
			if m.scroll.Cursor < len(vis) {
				id := vis[m.scroll.Cursor].LineNumber
				m.marks.Toggle(id)
			}
		case "u":
			// Next mark with wrap.
			visIDs := make([]int, len(vis))
			for i, e := range vis {
				visIDs[i] = e.LineNumber
			}
			if next := m.marks.NextMark(m.scroll.Cursor, visIDs); next >= 0 {
				if next <= m.scroll.Cursor {
					m.wrapDir = WrapForward
				}
				m.scroll.Cursor = next
			}
		case "U":
			// Previous mark with wrap.
			visIDs := make([]int, len(vis))
			for i, e := range vis {
				visIDs[i] = e.LineNumber
			}
			if prev := m.marks.PrevMark(m.scroll.Cursor, visIDs); prev >= 0 {
				if prev >= m.scroll.Cursor {
					m.wrapDir = WrapBack
				}
				m.scroll.Cursor = prev
			}
		case "e":
			// Next ERROR in full entries set.
			newIdx, dir := NextLevel(m.entries, m.entryIndexForVisible(m.scroll.Cursor), "ERROR")
			m.scroll.Cursor = m.visibleIndexForEntry(newIdx)
			m.wrapDir = dir
		case "E":
			// Previous ERROR in full entries set.
			newIdx, dir := PrevLevel(m.entries, m.entryIndexForVisible(m.scroll.Cursor), "ERROR")
			m.scroll.Cursor = m.visibleIndexForEntry(newIdx)
			m.wrapDir = dir
		case "w":
			// Next WARN in full entries set.
			newIdx, dir := NextLevel(m.entries, m.entryIndexForVisible(m.scroll.Cursor), "WARN")
			m.scroll.Cursor = m.visibleIndexForEntry(newIdx)
			m.wrapDir = dir
		case "W":
			// Previous WARN in full entries set.
			newIdx, dir := PrevLevel(m.entries, m.entryIndexForVisible(m.scroll.Cursor), "WARN")
			m.scroll.Cursor = m.visibleIndexForEntry(newIdx)
			m.wrapDir = dir
		}
		// Keep cursor visible in viewport.
		m.scroll.Offset = ensureVisible(m.scroll.Cursor, m.scroll.Offset, m.scroll.ViewportHeight)
		m.scroll.Offset = clampOffset(m.scroll.Offset, n, m.scroll.ViewportHeight)

		if m.scroll.Cursor != prev {
			if m.scroll.Cursor < len(vis) {
				entry := vis[m.scroll.Cursor]
				cmd = func() tea.Msg { return SelectionMsg{Entry: entry} }
			}
		}

	case tea.MouseMsg:
		vis := m.visibleEntries()
		n := len(vis)
		switch msg.Button {
		case tea.MouseButtonLeft:
			if msg.Action == tea.MouseActionPress {
				row := m.rowForY(msg.Y)
				if row >= 0 && row < n {
					m.scroll.Cursor = row
					m.scroll.Offset = ensureVisible(m.scroll.Cursor, m.scroll.Offset, m.scroll.ViewportHeight)
					if m.scroll.Cursor != prev {
						entry := vis[m.scroll.Cursor]
						cmd = func() tea.Msg { return SelectionMsg{Entry: entry} }
					}
				}
			} else if msg.Action == tea.MouseActionRelease {
				// Double-click detection: check if this is a second click on same row.
				// Bubble Tea exposes double-click as Action == tea.MouseActionMotion or
				// explicit double-click event. For simplicity, treat consecutive quick
				// presses as double-click — use DoubleClick action if available.
			}
		case tea.MouseButtonWheelDown:
			m.scroll.Cursor = clampCursor(m.scroll.Cursor+1, n)
			m.scroll.Offset = ensureVisible(m.scroll.Cursor, m.scroll.Offset, m.scroll.ViewportHeight)
			m.scroll.Offset = clampOffset(m.scroll.Offset, n, m.scroll.ViewportHeight)
			if m.scroll.Cursor != prev {
				entry := vis[m.scroll.Cursor]
				cmd = func() tea.Msg { return SelectionMsg{Entry: entry} }
			}
		case tea.MouseButtonWheelUp:
			m.scroll.Cursor = clampCursor(m.scroll.Cursor-1, n)
			m.scroll.Offset = ensureVisible(m.scroll.Cursor, m.scroll.Offset, m.scroll.ViewportHeight)
			m.scroll.Offset = clampOffset(m.scroll.Offset, n, m.scroll.ViewportHeight)
			if m.scroll.Cursor != prev {
				entry := vis[m.scroll.Cursor]
				cmd = func() tea.Msg { return SelectionMsg{Entry: entry} }
			}
		}
		// Timestamp-based double-click detection.
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			row := m.rowForY(msg.Y)
			now := time.Now()
			if row >= 0 && row == m.lastClickRow &&
				now.Sub(m.lastClickTime) < 500*time.Millisecond && n > 0 {
				// Double-click on same row within 500ms — open detail pane.
				entry := vis[m.scroll.Cursor]
				existingCmd := cmd
				cmd = func() tea.Msg {
					if existingCmd != nil {
						existingCmd() // drain selection msg
					}
					return OpenDetailPaneMsg{Entry: entry}
				}
				m.lastClickTime = time.Time{} // reset to prevent triple-click
			} else if row >= 0 {
				m.lastClickRow = row
				m.lastClickTime = now
			}
		}

	}

	return m, cmd
}

// entryIndexForVisible converts a visible-list cursor position to an index
// into the full m.entries slice.
func (m ListModel) entryIndexForVisible(visIdx int) int {
	if m.filtered == nil {
		return visIdx
	}
	if visIdx >= 0 && visIdx < len(m.filtered) {
		return m.filtered[visIdx]
	}
	return visIdx
}

// visibleIndexForEntry converts a full-entries index to the best visible index.
// If the entry is not in the filtered set, returns the current cursor.
func (m ListModel) visibleIndexForEntry(entryIdx int) int {
	if m.filtered == nil {
		return entryIdx
	}
	// Find the entry in the filtered list.
	for vi, fi := range m.filtered {
		if fi == entryIdx {
			return vi
		}
	}
	// Entry is filtered out — show its position indicator by keeping current cursor.
	// The view can check wrapDir != NoWrap or similar for indicator.
	return m.scroll.Cursor
}

// WrapDir returns the last wrap direction from a level-jump or mark nav.
func (m ListModel) WrapDir() WrapDirection { return m.wrapDir }

// View renders exactly ViewportHeight rows — no more, no less.
// Rows are taken from offset..offset+ViewportHeight; shortfalls are padded with
// empty lines so the layout never overflows or underflows.
func (m ListModel) View() string {
	vis := m.visibleEntries()
	n := len(vis)
	vh := m.scroll.ViewportHeight
	if vh <= 0 {
		return ""
	}

	start := m.scroll.Offset
	end := start + vh
	if end > n {
		end = n
	}

	var sb strings.Builder
	rendered := 0
	for i := start; i < end; i++ {
		if rendered > 0 {
			sb.WriteByte('\n')
		}
		isCursor := i == m.scroll.Cursor
		mark := ""
		if m.marks.IsMarked(vis[i].LineNumber) {
			mark = lipgloss.NewStyle().Foreground(m.th.Mark).Render("* ")
		}
		if isCursor {
			sb.WriteString(mark + RenderCompactRowWithBg(vis[i], m.width, m.th, m.cfg, m.th.CursorHighlight))
		} else {
			sb.WriteString(mark + RenderCompactRow(vis[i], m.width, m.th, m.cfg))
		}
		rendered++
	}
	// Pad remaining lines so the list always occupies exactly ViewportHeight rows.
	for rendered < vh {
		if rendered > 0 {
			sb.WriteByte('\n')
		}
		rendered++
	}
	return sb.String()
}

// RenderedRowCount returns how many entry rows were rendered in the last View() call.
// Used for benchmark validation.
func (m ListModel) RenderedRowCount() int {
	vis := m.visibleEntries()
	n := len(vis)
	if n == 0 {
		return 0
	}
	end := m.scroll.Offset + m.scroll.ViewportHeight
	if end > n {
		end = n
	}
	count := end - m.scroll.Offset
	if count < 0 {
		return 0
	}
	return count
}
