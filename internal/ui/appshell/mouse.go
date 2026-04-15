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
// y is 0-indexed from the top of the terminal.
//
// Row assignments:
//   0                  → ZoneHeader
//   1..entryListBottom → ZoneEntryList
//   entryListBottom+1  → ZoneDivider (only when detail pane is open)
//   divider+1..paneEnd → ZoneDetailPane (only when detail pane is open)
//   last row           → ZoneStatusBar
func (r MouseRouter) Zone(x, y int) MouseZone {
	_ = x // x is unused; zones span full width

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
	entryListEnd := entryListStart + r.layout.EntryListHeight() - 1

	if !r.layout.DetailPaneOpen {
		if y >= entryListStart && y <= entryListEnd {
			return ZoneEntryList
		}
		return ZoneUnknown
	}

	// Detail pane is open.
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
