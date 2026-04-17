package appshell

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Layout: 80 wide, 24 tall, no detail pane.
// Row 0 = header, rows 1-22 = entry list (22 rows), row 23 = status bar.
func testRouter(detailOpen bool, detailHeight int) MouseRouter {
	l := NewLayout(80, 24, detailOpen, detailHeight)
	return NewMouseRouter(l)
}

// T-052: R6.1 — header row is ZoneHeader.
func TestMouseRouter_HeaderRow(t *testing.T) {
	r := testRouter(false, 0)
	if r.Zone(0, 0) != ZoneHeader {
		t.Errorf("row 0 should be ZoneHeader")
	}
}

// T-052: R6.1 — status bar row is ZoneStatusBar.
func TestMouseRouter_StatusBarRow(t *testing.T) {
	r := testRouter(false, 0)
	if r.Zone(0, 23) != ZoneStatusBar {
		t.Errorf("row 23 should be ZoneStatusBar")
	}
}

// T-052: R6.1 — entry list rows without detail pane.
func TestMouseRouter_EntryListRows_NoDetailPane(t *testing.T) {
	r := testRouter(false, 0)
	for y := 1; y <= 22; y++ {
		if r.Zone(0, y) != ZoneEntryList {
			t.Errorf("row %d should be ZoneEntryList, got %v", y, r.Zone(0, y))
		}
	}
}

// T-052: R6.1 — detail pane and divider zones when pane is open.
// Layout with 8-row detail pane: header=0, entries=1-13 (13 rows), divider=14,
// detail=15-22, status=23.
func TestMouseRouter_DetailPaneOpen(t *testing.T) {
	// height=24, header=1, status=1, detailPane=8 → entryList=14 rows
	l := NewLayout(80, 24, true, 8)
	r := NewMouseRouter(l)

	entryListHeight := l.EntryListHeight() // 24-1-1-8=14
	// Entry list: rows 1..14
	for y := 1; y <= entryListHeight; y++ {
		if r.Zone(0, y) != ZoneEntryList {
			t.Errorf("row %d should be ZoneEntryList, got %v", y, r.Zone(0, y))
		}
	}
	// Divider: row 15
	dividerRow := 1 + entryListHeight
	if r.Zone(0, dividerRow) != ZoneDivider {
		t.Errorf("row %d should be ZoneDivider, got %v", dividerRow, r.Zone(0, dividerRow))
	}
	// Detail pane: rows 16..23-1 = rows 16..22
	for y := dividerRow + 1; y < 23; y++ {
		if r.Zone(0, y) != ZoneDetailPane {
			t.Errorf("row %d should be ZoneDetailPane, got %v", y, r.Zone(0, y))
		}
	}
}

// T-052: R6.2 — no crash on any mouse position (just call Zone for all rows).
func TestMouseRouter_NoCrashAnyPosition(t *testing.T) {
	r := testRouter(true, 8)
	for y := -1; y <= 30; y++ {
		_ = r.Zone(0, y) // must not panic
	}
}

// T-052: RouteMouseMsg classifies tea.MouseMsg.
func TestMouseRouter_RouteMouseMsg(t *testing.T) {
	r := testRouter(false, 0)
	msg := tea.MouseMsg{X: 0, Y: 0}
	if r.RouteMouseMsg(msg) != ZoneHeader {
		t.Error("expected ZoneHeader for Y=0")
	}
}

// T-094: right-split horizontal zoning with 1-cell buffer on each side
// of the divider column.
//
// Layout: width=100, height=24, widthRatio=0.30 →
//   usable      = 100 - 4 - 1 = 95
//   listContent = int(95*0.7)  = 66
//   detailContent = 95 - 66    = 29
//   listEnd     = 66 + 1 = 67  (list pane right border, buffer)
//   divider     = 68           (ZoneDivider)
//   detailStart = 69           (detail pane left border, buffer)
//   detail data = 70..98       (29 cols of content + right border at 99)
func TestMouseRouter_RightSplit_HorizontalZones(t *testing.T) {
	l := NewLayout(100, 24, true, 0)
	l.Orientation = OrientationRight
	l.WidthRatio = 0.30
	r := NewMouseRouter(l)

	const row = 5

	// Click within list content → list.
	if z := r.Zone(l.ListContentWidth()-1, row); z != ZoneEntryList {
		t.Errorf("listEnd-1 (%d): want ZoneEntryList, got %v", l.ListContentWidth()-1, z)
	}
	// Click on listEnd (buffer column) → unknown.
	listEnd := l.ListContentWidth() + 1
	if z := r.Zone(listEnd, row); z != ZoneUnknown {
		t.Errorf("listEnd buffer (%d): want ZoneUnknown, got %v", listEnd, z)
	}
	// Click on divider → divider.
	divider := listEnd + 1
	if z := r.Zone(divider, row); z != ZoneDivider {
		t.Errorf("divider (%d): want ZoneDivider, got %v", divider, z)
	}
	// Click on detailStart (buffer column) → unknown.
	detailStart := divider + 1
	if z := r.Zone(detailStart, row); z != ZoneUnknown {
		t.Errorf("detailStart buffer (%d): want ZoneUnknown, got %v", detailStart, z)
	}
	// Click immediately after the right buffer → detail.
	if z := r.Zone(detailStart+1, row); z != ZoneDetailPane {
		t.Errorf("detailStart+1 (%d): want ZoneDetailPane, got %v", detailStart+1, z)
	}
}

// T-094: header + status bar still take precedence over horizontal zones.
func TestMouseRouter_RightSplit_HeaderAndStatus(t *testing.T) {
	l := NewLayout(100, 24, true, 0)
	l.Orientation = OrientationRight
	l.WidthRatio = 0.30
	r := NewMouseRouter(l)
	if z := r.Zone(50, 0); z != ZoneHeader {
		t.Errorf("y=0: want ZoneHeader, got %v", z)
	}
	if z := r.Zone(50, 23); z != ZoneStatusBar {
		t.Errorf("y=23: want ZoneStatusBar, got %v", z)
	}
}
