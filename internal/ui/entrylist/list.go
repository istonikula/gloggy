package entrylist

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/istonikula/gloggy/internal/config"
	"github.com/istonikula/gloggy/internal/logsource"
	"github.com/istonikula/gloggy/internal/theme"
)

const renderBuffer = 5 // extra rows rendered above/below the visible window

// SelectionMsg is emitted when the cursor moves to a new entry.
type SelectionMsg struct {
	Entry logsource.Entry
}

// ListModel is the virtual-rendering entry list Bubble Tea model.
// It only renders visible rows plus a small buffer, regardless of total entry count.
type ListModel struct {
	entries  []logsource.Entry
	scroll   ScrollState
	th       theme.Theme
	cfg      config.Config
	width    int
}

// NewListModel creates a ListModel.
func NewListModel(th theme.Theme, cfg config.Config, width, height int) ListModel {
	return ListModel{
		th:    th,
		cfg:   cfg,
		width: width,
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

// SelectedEntry returns the entry at the cursor, or zero value if empty.
func (m ListModel) SelectedEntry() (logsource.Entry, bool) {
	if len(m.entries) == 0 || m.scroll.Cursor >= len(m.entries) {
		return logsource.Entry{}, false
	}
	return m.entries[m.scroll.Cursor], true
}

// Init satisfies tea.Model.
func (m ListModel) Init() tea.Cmd { return nil }

// Update handles keyboard navigation.
func (m ListModel) Update(msg tea.Msg) (ListModel, tea.Cmd) {
	n := len(m.entries)
	if n == 0 {
		return m, nil
	}

	var cmd tea.Cmd
	prev := m.scroll.Cursor

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
		}
		// Keep cursor visible in viewport.
		m.scroll.Offset = ensureVisible(m.scroll.Cursor, m.scroll.Offset, m.scroll.ViewportHeight)
		m.scroll.Offset = clampOffset(m.scroll.Offset, n, m.scroll.ViewportHeight)

		if m.scroll.Cursor != prev {
			entry := m.entries[m.scroll.Cursor]
			cmd = func() tea.Msg { return SelectionMsg{Entry: entry} }
		}

	case tea.WindowSizeMsg:
		m.scroll.ViewportHeight = msg.Height
		m.width = msg.Width
	}

	return m, cmd
}

// View renders only the visible rows plus a small buffer.
func (m ListModel) View() string {
	n := len(m.entries)
	if n == 0 {
		return ""
	}

	start := m.scroll.Offset - renderBuffer
	if start < 0 {
		start = 0
	}
	end := m.scroll.Offset + m.scroll.ViewportHeight + renderBuffer
	if end > n {
		end = n
	}

	var sb strings.Builder
	for i := start; i < end; i++ {
		row := RenderCompactRow(m.entries[i], m.width, m.th, m.cfg)
		sb.WriteString(row)
		if i < end-1 {
			sb.WriteByte('\n')
		}
	}
	return sb.String()
}

// RenderedRowCount returns how many rows were rendered in the last View() call.
// Used for benchmark validation.
func (m ListModel) RenderedRowCount() int {
	n := len(m.entries)
	if n == 0 {
		return 0
	}
	start := m.scroll.Offset - renderBuffer
	if start < 0 {
		start = 0
	}
	end := m.scroll.Offset + m.scroll.ViewportHeight + renderBuffer
	if end > n {
		end = n
	}
	if end < start {
		return 0
	}
	return end - start
}
