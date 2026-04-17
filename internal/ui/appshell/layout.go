package appshell

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// FocusTarget identifies which component currently has keyboard focus.
type FocusTarget int

const (
	FocusEntryList   FocusTarget = iota
	FocusDetailPane
	FocusFilterPanel
)

// Border / divider constants for right-split composition (DESIGN.md §5).
const (
	paneBorders  = 2 // left + right border glyphs per pane
	dividerWidth = 1
)

// Layout computes pane heights (and, in right-split, widths) from terminal
// dimensions, the detail pane ratio, and the selected orientation. All panes
// together fill the full terminal with no gaps or overlap.
type Layout struct {
	Width            int
	Height           int
	HeaderHeight     int // always 1
	StatusBarHeight  int // always 1
	DetailPaneOpen   bool
	DetailPaneHeight int
	Orientation      Orientation
	WidthRatio       float64
}

// EntryListHeight returns the number of rows available for the entry list.
// In right-split mode the detail pane does not consume vertical space, so
// DetailPaneHeight is ignored.
func (l Layout) EntryListHeight() int {
	used := l.HeaderHeight + l.StatusBarHeight
	if l.DetailPaneOpen && l.Orientation != OrientationRight {
		used += l.DetailPaneHeight
	}
	h := l.Height - used
	if h < 1 {
		h = 1
	}
	return h
}

// usableSplitWidth returns the horizontal budget shared by the two pane
// content areas in right-split orientation, after subtracting both panes'
// borders and the 1-cell divider (DESIGN.md §5 Border accounting).
func (l Layout) usableSplitWidth() int {
	u := l.Width - 2*paneBorders - dividerWidth
	if u < 0 {
		u = 0
	}
	return u
}

// ListContentWidth returns the content-area width of the entry list pane.
// In below-mode or when the detail pane is closed, it is the full terminal
// width. In right-split, it is (usable * (1 - WidthRatio)).
func (l Layout) ListContentWidth() int {
	if l.Orientation != OrientationRight || !l.DetailPaneOpen {
		return l.Width
	}
	u := l.usableSplitWidth()
	return int(float64(u) * (1.0 - l.WidthRatio))
}

// DetailContentWidth returns the content-area width of the detail pane in
// right-split orientation. Returns 0 when the pane is closed or orientation
// is below.
func (l Layout) DetailContentWidth() int {
	if l.Orientation != OrientationRight || !l.DetailPaneOpen {
		return 0
	}
	u := l.usableSplitWidth()
	return u - int(float64(u)*(1.0-l.WidthRatio))
}

// NewLayout creates a below-mode Layout. detailPaneHeight is only used when
// detailPaneOpen is true.
func NewLayout(width, height int, detailPaneOpen bool, detailPaneHeight int) Layout {
	return Layout{
		Width:            width,
		Height:           height,
		HeaderHeight:     1,
		StatusBarHeight:  1,
		DetailPaneOpen:   detailPaneOpen,
		DetailPaneHeight: detailPaneHeight,
		Orientation:      OrientationBelow,
		WidthRatio:       0.30,
	}
}

// LayoutModel manages the TUI layout and assembles the final screen string via
// lipgloss.JoinVertical.
//
// The caller is responsible for providing rendered strings for each zone.
// LayoutModel only handles compositing.
type LayoutModel struct {
	layout Layout
	width  int
}

// NewLayoutModel creates a LayoutModel.
func NewLayoutModel(width, height int) LayoutModel {
	return LayoutModel{
		layout: NewLayout(width, height, false, 0),
		width:  width,
	}
}

// SetSize updates the terminal dimensions.
func (m LayoutModel) SetSize(width, height int) LayoutModel {
	m.layout.Width = width
	m.layout.Height = height
	m.width = width
	return m
}

// SetDetailPane updates the detail pane open/height state.
func (m LayoutModel) SetDetailPane(open bool, height int) LayoutModel {
	m.layout.DetailPaneOpen = open
	m.layout.DetailPaneHeight = height
	return m
}

// SetOrientation updates the pane orientation (below vs right-split).
func (m LayoutModel) SetOrientation(o Orientation) LayoutModel {
	m.layout.Orientation = o
	return m
}

// SetWidthRatio updates the width ratio used in right-split composition.
func (m LayoutModel) SetWidthRatio(r float64) LayoutModel {
	m.layout.WidthRatio = r
	return m
}

// Layout returns the current layout.
func (m LayoutModel) Layout() Layout { return m.layout }

// MinTerminalWidth and MinTerminalHeight define the minimum-viable terminal
// floor (T-090, DESIGN.md §8). Below either threshold the layout suppresses
// normal rendering and shows a centered "terminal too small" message.
const (
	MinTerminalWidth  = 60
	MinTerminalHeight = 15
)

// IsBelowMinFloor reports whether the current dimensions are below the
// minimum-viable terminal floor.
func (m LayoutModel) IsBelowMinFloor() bool {
	return m.layout.Width < MinTerminalWidth || m.layout.Height < MinTerminalHeight
}

// RenderTooSmall draws the terminal-too-small fallback message centered in
// the current dimensions (T-090).
func (m LayoutModel) RenderTooSmall() string {
	msg := fmt.Sprintf("terminal too small\nminimum %dx%d", MinTerminalWidth, MinTerminalHeight)
	return lipgloss.Place(m.layout.Width, m.layout.Height, lipgloss.Center, lipgloss.Center, msg)
}

// Render assembles the full screen from the provided zone strings.
// All zone strings must already be the correct width; Render only stacks them vertically.
//
// When the terminal falls below the minimum-viable floor (60x15 per
// DESIGN.md §8) the panel rendering is suppressed and a centered fallback
// message is shown instead.
//
// Parameters:
//   header    — rendered header bar (1 row)
//   entryList — rendered entry list (EntryListHeight rows)
//   detailPane — rendered detail pane (DetailPaneHeight rows; ignored when pane is closed)
//   statusBar — rendered status/key-hint bar (1 row)
func (m LayoutModel) Render(header, entryList, detailPane, statusBar string) string {
	if m.IsBelowMinFloor() {
		return m.RenderTooSmall()
	}
	// Right-split: compose the main area horizontally between the header
	// and status bar. The caller is expected to have rendered entryList
	// and detailPane at the widths returned by ListContentWidth()+borders
	// and DetailContentWidth()+borders respectively.
	if m.layout.Orientation == OrientationRight && m.layout.DetailPaneOpen && detailPane != "" {
		main := lipgloss.JoinHorizontal(lipgloss.Top, entryList, m.renderInlineDivider(), detailPane)
		return lipgloss.JoinVertical(lipgloss.Left, header, main, statusBar)
	}
	parts := []string{header, entryList}
	if m.layout.DetailPaneOpen && detailPane != "" {
		parts = append(parts, detailPane)
	}
	parts = append(parts, statusBar)
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// renderInlineDivider returns a minimal 1-cell vertical divider for the
// right-split main area. T-089 will replace this with a themed renderer
// (RenderDivider). The height matches the main area — terminal height minus
// the header and status rows.
func (m LayoutModel) renderInlineDivider() string {
	h := m.layout.Height - m.layout.HeaderHeight - m.layout.StatusBarHeight
	if h < 1 {
		h = 1
	}
	lines := make([]string, h)
	for i := range lines {
		lines[i] = "│"
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}
