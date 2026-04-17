package detailpane

import (
	tea "github.com/charmbracelet/bubbletea"

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
	open    bool
	entry   logsource.Entry
	scroll  ScrollModel
	th      theme.Theme
	height  int
	Focused bool // set by app before rendering for focus indicator
}

// NewPaneModel creates a PaneModel.
func NewPaneModel(th theme.Theme, height int) PaneModel {
	return PaneModel{
		th:     th,
		height: height,
		scroll: NewScrollModel("", height),
	}
}

// IsOpen returns true when the pane is visible.
func (m PaneModel) IsOpen() bool { return m.open }

// Open activates the detail pane with the given entry.
func (m PaneModel) Open(entry logsource.Entry) PaneModel {
	var content string
	if entry.IsJSON {
		content = RenderJSON(entry, m.th, nil)
	} else {
		content = RenderRaw(entry)
	}
	m.entry = entry
	m.open = true
	m.scroll = NewScrollModel(content, m.height)
	return m
}

// Close dismisses the detail pane.
func (m PaneModel) Close() PaneModel {
	m.open = false
	return m
}

// SetHeight updates the visible height of the pane.
func (m PaneModel) SetHeight(h int) PaneModel {
	m.height = h
	m.scroll.height = h
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

// borderRows returns how many rows the pane border consumes (top border).
func (m PaneModel) borderRows() int { return 1 }

// ContentHeight returns the height available for content after subtracting borders.
func (m PaneModel) ContentHeight() int {
	h := m.height - m.borderRows()
	if h < 1 {
		h = 1
	}
	return h
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
	m.scroll.height = m.ContentHeight()
	content := m.scroll.View()

	state := appshell.PaneStateUnfocused
	if m.Focused {
		state = appshell.PaneStateFocused
	}
	return appshell.PaneStyle(m.th, state).Render(content)
}
