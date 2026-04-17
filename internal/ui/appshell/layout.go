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

// Layout computes pane heights from terminal dimensions and the detail pane height ratio.
// All panes together fill the full terminal height with no gaps or overlap.
type Layout struct {
	Width          int
	Height         int
	HeaderHeight   int // always 1
	StatusBarHeight int // always 1
	DetailPaneOpen bool
	DetailPaneHeight int
}

// EntryListHeight returns the number of rows available for the entry list.
func (l Layout) EntryListHeight() int {
	used := l.HeaderHeight + l.StatusBarHeight
	if l.DetailPaneOpen {
		used += l.DetailPaneHeight
	}
	h := l.Height - used
	if h < 1 {
		h = 1
	}
	return h
}

// NewLayout creates a Layout. detailPaneHeight is only used when detailPaneOpen is true.
func NewLayout(width, height int, detailPaneOpen bool, detailPaneHeight int) Layout {
	return Layout{
		Width:            width,
		Height:           height,
		HeaderHeight:     1,
		StatusBarHeight:  1,
		DetailPaneOpen:   detailPaneOpen,
		DetailPaneHeight: detailPaneHeight,
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
	parts := []string{header, entryList}
	if m.layout.DetailPaneOpen && detailPane != "" {
		parts = append(parts, detailPane)
	}
	parts = append(parts, statusBar)
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}
