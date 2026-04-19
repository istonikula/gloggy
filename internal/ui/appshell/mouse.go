package appshell

import tea "github.com/charmbracelet/bubbletea"

// MouseZone identifies which layout area a mouse event belongs to.
type MouseZone int

const (
	ZoneEntryList  MouseZone = iota
	ZoneDetailPane
	ZoneDivider
	ZoneHeader
	ZoneStatusBar
	ZoneUnknown
)

// MouseRouter maps mouse events to layout zones based on current layout dimensions.
// It routes events so each component only receives events in its area.
type MouseRouter struct {
	layout Layout
}

// NewMouseRouter creates a MouseRouter for the given layout.
func NewMouseRouter(layout Layout) MouseRouter {
	return MouseRouter{layout: layout}
}

// Zone returns the layout zone for a mouse event at the given (x, y) terminal coordinate.
// y is 0-indexed from the top of the terminal; x is 0-indexed from the left.
//
// Below-mode row assignments:
//   0                  → ZoneHeader
//   1..entryListBottom → ZoneEntryList
//   entryListBottom+1  → ZoneDivider (only when detail pane is open)
//   divider+1..paneEnd → ZoneDetailPane (only when detail pane is open)
//   last row           → ZoneStatusBar
//
// Right-split mode (T-094): the main area is partitioned horizontally
// instead. The list pane occupies the left columns up to its right
// border, then a 1-cell buffer (the list pane's right border), then the
// 1-cell divider, then a 1-cell buffer (the detail pane's left border),
// then the detail pane fills the remaining columns. Clicks on either
// buffer are routed to ZoneUnknown so chrome cannot be mistakenly
// targeted.
func (r MouseRouter) Zone(x, y int) MouseZone {
	headerRow := 0
	if y == headerRow {
		return ZoneHeader
	}

	termH := r.layout.Height
	statusRow := termH - 1
	if y == statusRow {
		return ZoneStatusBar
	}

	entryListStart := 1

	// Right-split with detail open: horizontal partitioning.
	if r.layout.DetailPaneOpen && r.layout.Orientation == OrientationRight {
		// Right-split main area spans the full y between header and
		// status (no horizontal divider row).
		if y < entryListStart || y >= statusRow {
			return ZoneUnknown
		}
		// T-160 (F-122): the visible `│` glyph renders at column
		// ListContentWidth() — confirmed by rendering the layout and
		// locating the glyph programmatically (see
		// TestMouseRouter_T160_RendererTruth_DividerColMatchesGlyph).
		// The list pane is rendered as {leftBorder, content, rightBorder}
		// whose outer width equals ListContentWidth(), so the right
		// border sits at col ListContentWidth()-1; the divider glyph
		// follows at col ListContentWidth(); the detail pane's left
		// border sits at col ListContentWidth()+1. Prior arithmetic
		// (listEnd=LCW+1, divider=LCW+2) was 2 cols past the visible
		// divider, so drag-initiation Presses on the real divider were
		// misrouted to ZoneDetailPane.
		listRightBorder := r.layout.ListContentWidth() - 1
		divider := r.layout.ListContentWidth()
		detailLeftBorder := divider + 1
		switch {
		case x < listRightBorder:
			return ZoneEntryList
		case x == listRightBorder, x == detailLeftBorder:
			return ZoneUnknown // 1-cell buffer on each side of divider
		case x == divider:
			return ZoneDivider
		default:
			return ZoneDetailPane
		}
	}

	// Below-mode (default).
	entryListEnd := entryListStart + r.layout.EntryListHeight() - 1
	if !r.layout.DetailPaneOpen {
		if y >= entryListStart && y <= entryListEnd {
			return ZoneEntryList
		}
		return ZoneUnknown
	}
	if y >= entryListStart && y <= entryListEnd {
		return ZoneEntryList
	}
	dividerRow := entryListEnd + 1
	if y == dividerRow {
		return ZoneDivider
	}
	paneStart := dividerRow + 1
	paneEnd := paneStart + r.layout.DetailPaneHeight - 1
	if y >= paneStart && y <= paneEnd {
		return ZoneDetailPane
	}
	return ZoneUnknown
}

// RouteMouseMsg classifies a tea.MouseMsg and returns its zone.
// Returns ZoneUnknown if msg is not a MouseMsg.
func (r MouseRouter) RouteMouseMsg(msg tea.Msg) MouseZone {
	m, ok := msg.(tea.MouseMsg)
	if !ok {
		return ZoneUnknown
	}
	return r.Zone(m.X, m.Y)
}
