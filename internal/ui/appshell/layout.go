package appshell

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/istonikula/gloggy/internal/theme"
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

// ListContentTopY returns the terminal Y coordinate of the first entry-list
// content row — i.e. the header height + the list pane's top border (1).
// This is the single owner of the click-to-row Y offset per
// cavekit-entry-list R10 / cavekit-app-shell R2 (T-158). Consumers MUST
// read this value rather than re-derive the header or border arithmetic
// locally; drift between consumers is what caused the historical
// 2-row click-offset bug.
func (l Layout) ListContentTopY() int {
	return l.HeaderHeight + 1
}

// ClickToListRow converts a terminal Y coordinate into a list-viewport
// relative row (0-indexed from the first visible row). Returns ok=false
// when the coordinate lies outside the list's content rows (on the
// header, list top/bottom border, divider, detail pane, or status bar).
// Adding the list's current scroll offset to the returned viewportY
// yields the clicked visible-entry index.
//
// Partitioning between panes is owned by cavekit-app-shell R6 (the
// MouseRouter zone resolver); this helper only resolves Y within the
// list pane. Callers that skip the zone check must treat ok=false as
// "not inside list content" — never map it to row 0.
func (l Layout) ClickToListRow(terminalY int) (int, bool) {
	start := l.ListContentTopY()
	// The list pane is wrapped in a 1-row top border and a 1-row bottom
	// border (entrylist.NewListModel subtracts 2 from EntryListHeight
	// for its inner viewport). The bottom border sits at entryListEnd,
	// so content rows span [start, start+viewportRows).
	viewportRows := l.EntryListHeight() - 2
	if viewportRows < 1 {
		return 0, false
	}
	if terminalY < start || terminalY >= start+viewportRows {
		return 0, false
	}
	return terminalY - start, true
}

// DetailPaneContentTopY returns the terminal Y coordinate of the first
// detail-pane content row. V8 single-owner of click-to-pane-row math (T28).
// In below-mode the pane follows the list outer block (HeaderHeight rows
// above, EntryListHeight rows in the list including its top + bottom
// borders) and then its own top border (+1). Right-split places the pane
// directly under the header, so only HeaderHeight + pane-top-border.
// Returns 0 when the pane is closed.
func (l Layout) DetailPaneContentTopY() int {
	if !l.DetailPaneOpen {
		return 0
	}
	if l.Orientation == OrientationRight {
		return l.HeaderHeight + 1
	}
	return l.HeaderHeight + l.EntryListHeight() + 1
}

// ClickToPaneRow converts a terminal Y coordinate into a detail-pane
// content-local row (0-indexed from the first content row). Returns
// ok=false when the coordinate lies outside the pane's content rows
// (on borders, divider, other panes, header, or status bar). Partitioning
// between panes is owned by cavekit-app-shell R6 (MouseRouter zone
// resolver); this helper only resolves Y within the detail pane's content
// rows. Single owner per V8 — callers must NOT re-derive borders.
func (l Layout) ClickToPaneRow(terminalY int) (int, bool) {
	if !l.DetailPaneOpen {
		return 0, false
	}
	start := l.DetailPaneContentTopY()
	viewportRows := DetailPaneVerticalRows(l) - 2 // subtract top + bottom borders
	if viewportRows < 1 {
		return 0, false
	}
	if terminalY < start || terminalY >= start+viewportRows {
		return 0, false
	}
	return terminalY - start, true
}

// DetailPaneVerticalRows returns the outer vertical allocation (rows,
// border-inclusive) for the detail pane. In below-mode this is
// DetailPaneHeight (height_ratio * terminalHeight). In right-mode the pane
// occupies the full main-area slot between the header and status bar and
// height_ratio must NOT be applied to the vertical dimension (T-123, F-013).
// Returns 0 when the pane is closed.
func DetailPaneVerticalRows(l Layout) int {
	if !l.DetailPaneOpen {
		return 0
	}
	if l.Orientation == OrientationRight {
		h := l.Height - l.HeaderHeight - l.StatusBarHeight
		if h < 1 {
			h = 1
		}
		return h
	}
	return l.DetailPaneHeight
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
// lipgloss.JoinVertical (or JoinHorizontal in right-split mode).
//
// The caller is responsible for providing rendered strings for each zone.
// LayoutModel only handles compositing.
type LayoutModel struct {
	layout Layout
	width  int
	theme  theme.Theme
}

// NewLayoutModel creates a LayoutModel.
func NewLayoutModel(width, height int) LayoutModel {
	return LayoutModel{
		layout: NewLayout(width, height, false, 0),
		width:  width,
	}
}

// WithTheme attaches a theme so the right-split divider can be rendered with
// the canonical DividerColor (T-089). Without a theme, the divider falls
// back to an unstyled glyph.
func (m LayoutModel) WithTheme(th theme.Theme) LayoutModel {
	m.theme = th
	return m
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

// renderInlineDivider returns the right-split divider for the main area.
// Height matches terminal height minus the header and status rows. When a
// theme is attached the divider uses theme.DividerColor (T-089); otherwise
// the glyph is unstyled.
func (m LayoutModel) renderInlineDivider() string {
	h := m.layout.Height - m.layout.HeaderHeight - m.layout.StatusBarHeight
	if h < 1 {
		h = 1
	}
	if m.theme.DividerColor != "" {
		return RenderDivider(h, m.theme)
	}
	lines := make([]string, h)
	for i := range lines {
		lines[i] = dividerGlyph
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}
