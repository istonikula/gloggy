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
