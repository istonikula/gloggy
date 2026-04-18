package appshell

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

// T-046: R2.1 — header at top, entrylist in middle, status at bottom.
func TestLayout_Render_WithoutDetailPane(t *testing.T) {
	m := NewLayoutModel(80, 24)
	out := m.Render("HEADER", "ENTRYLIST", "DETAIL", "STATUS")
	lines := strings.Split(out, "\n")
	if strings.TrimSpace(lines[0]) != "HEADER" {
		t.Errorf("first line should be HEADER, got %q", lines[0])
	}
	if strings.TrimSpace(lines[1]) != "ENTRYLIST" {
		t.Errorf("second line should be ENTRYLIST, got %q", lines[1])
	}
	if strings.TrimSpace(lines[len(lines)-1]) != "STATUS" {
		t.Errorf("last line should be STATUS, got %q", lines[len(lines)-1])
	}
	// Detail pane should not appear when closed.
	if strings.Contains(out, "DETAIL") {
		t.Error("DETAIL should not appear when pane is closed")
	}
}

// T-046: R2.3 — detail pane appears between entry list and status bar when open.
func TestLayout_Render_WithDetailPane(t *testing.T) {
	m := NewLayoutModel(80, 24).SetDetailPane(true, 5)
	out := m.Render("HEADER", "ENTRYLIST", "DETAIL", "STATUS")
	lines := strings.Split(out, "\n")
	// Should be: HEADER, ENTRYLIST, DETAIL, STATUS
	if strings.TrimSpace(lines[0]) != "HEADER" {
		t.Errorf("line 0: want HEADER, got %q", lines[0])
	}
	if strings.TrimSpace(lines[1]) != "ENTRYLIST" {
		t.Errorf("line 1: want ENTRYLIST, got %q", lines[1])
	}
	if strings.TrimSpace(lines[2]) != "DETAIL" {
		t.Errorf("line 2: want DETAIL, got %q", lines[2])
	}
	if strings.TrimSpace(lines[3]) != "STATUS" {
		t.Errorf("line 3: want STATUS, got %q", lines[3])
	}
}

// T-046: R2.5 — entry list height fills available space.
func TestLayout_EntryListHeight_FillsSpace(t *testing.T) {
	l := NewLayout(80, 24, false, 0)
	// 24 rows - 1 header - 1 status = 22
	if l.EntryListHeight() != 22 {
		t.Errorf("entry list height: got %d, want 22", l.EntryListHeight())
	}
}

// T-046: R2.3 — entry list height reduced by detail pane when open.
func TestLayout_EntryListHeight_ReducedByDetailPane(t *testing.T) {
	l := NewLayout(80, 24, true, 8)
	// 24 - 1 - 1 - 8 = 14
	if l.EntryListHeight() != 14 {
		t.Errorf("entry list height with detail pane: got %d, want 14", l.EntryListHeight())
	}
}

// SetSize updates dimensions.
func TestLayoutModel_SetSize(t *testing.T) {
	m := NewLayoutModel(80, 24).SetSize(120, 40)
	l := m.Layout()
	if l.Width != 120 || l.Height != 40 {
		t.Errorf("SetSize: got %dx%d, want 120x40", l.Width, l.Height)
	}
}

// T-090: terminal-too-small fallback at 59x15.
func TestLayout_FallbackBelowMinWidth(t *testing.T) {
	m := NewLayoutModel(59, 15)
	out := m.Render("HEADER", "ENTRYLIST", "DETAIL", "STATUS")
	if !strings.Contains(out, "terminal too small") {
		t.Errorf("expected fallback message at 59x15, got %q", out)
	}
	if strings.Contains(out, "HEADER") || strings.Contains(out, "ENTRYLIST") {
		t.Errorf("panels should be suppressed at 59x15, got %q", out)
	}
}

// T-090: terminal-too-small fallback at 60x14.
func TestLayout_FallbackBelowMinHeight(t *testing.T) {
	m := NewLayoutModel(60, 14)
	out := m.Render("HEADER", "ENTRYLIST", "DETAIL", "STATUS")
	if !strings.Contains(out, "terminal too small") {
		t.Errorf("expected fallback message at 60x14, got %q", out)
	}
	if strings.Contains(out, "HEADER") {
		t.Errorf("panels should be suppressed at 60x14, got %q", out)
	}
}

// T-090: normal render resumes at 60x15.
func TestLayout_NormalRenderAtMinFloor(t *testing.T) {
	m := NewLayoutModel(60, 15)
	out := m.Render("HEADER", "ENTRYLIST", "DETAIL", "STATUS")
	if strings.Contains(out, "terminal too small") {
		t.Errorf("normal render should resume at 60x15, got %q", out)
	}
	if !strings.Contains(out, "HEADER") {
		t.Errorf("HEADER should appear at 60x15, got %q", out)
	}
}

// T-088: ListContentWidth + DetailContentWidth in right-split sum with chrome
// (4 borders + 1 divider = 5 cells) to exactly the terminal width.
func TestLayout_RightSplit_ContentWidthsSumWithChrome(t *testing.T) {
	cases := []struct {
		termW int
		ratio float64
	}{
		{120, 0.30},
		{100, 0.20},
		{80, 0.50},
		{60, 0.30},
	}
	for _, tc := range cases {
		l := NewLayout(tc.termW, 24, true, 0)
		l.Orientation = OrientationRight
		l.WidthRatio = tc.ratio
		listW := l.ListContentWidth()
		detailW := l.DetailContentWidth()
		// Sum: list + 2 borders + 1 divider + 2 borders + detail.
		got := listW + 2 + 1 + 2 + detailW
		if got != tc.termW {
			t.Errorf("widths+chrome (termW=%d, ratio=%.2f): list=%d, detail=%d, sum=%d, want %d",
				tc.termW, tc.ratio, listW, detailW, got, tc.termW)
		}
	}
}

// T-088: Width-ratio formula matches DESIGN.md §5 example exactly
// (termWidth=120, widthRatio=0.30 ⇒ usable=115, listW=80, detailW=35).
func TestLayout_RightSplit_DesignExampleNumbers(t *testing.T) {
	l := NewLayout(120, 24, true, 0)
	l.Orientation = OrientationRight
	l.WidthRatio = 0.30
	if got := l.ListContentWidth(); got != 80 {
		t.Errorf("ListContentWidth(120, 0.30): got %d, want 80", got)
	}
	if got := l.DetailContentWidth(); got != 35 {
		t.Errorf("DetailContentWidth(120, 0.30): got %d, want 35", got)
	}
}

// T-088: Render in right-split composes JoinHorizontal(list, divider, detail)
// between header and status; total rendered width equals termWidth.
func TestLayoutModel_Render_RightSplit_TotalWidth(t *testing.T) {
	m := NewLayoutModel(120, 24).
		SetDetailPane(true, 0).
		SetOrientation(OrientationRight).
		SetWidthRatio(0.30)

	l := m.Layout()
	listView := strings.Repeat("L", l.ListContentWidth()+2)     // pretend pane includes 2 borders
	detailView := strings.Repeat("D", l.DetailContentWidth()+2) // ditto
	header := strings.Repeat("H", 120)
	status := strings.Repeat("S", 120)

	out := m.Render(header, listView, detailView, status)
	lines := strings.Split(out, "\n")
	// Header is line 0; main starts at line 1; status is the last line.
	mainLine := lines[1]
	if w := lipgloss.Width(mainLine); w != 120 {
		t.Errorf("main line width: got %d, want 120; line=%q", w, mainLine)
	}
}

// T-088: Below-mode keeps the existing vertical composition unchanged.
func TestLayoutModel_Render_BelowMode_Unchanged(t *testing.T) {
	m := NewLayoutModel(80, 24).SetDetailPane(true, 5) // orientation defaults to below
	out := m.Render("HEADER", "ENTRYLIST", "DETAIL", "STATUS")
	if !strings.Contains(out, "DETAIL") {
		t.Error("below-mode render must include DETAIL line")
	}
	// Should not contain the divider glyph.
	if strings.Contains(out, "│") {
		t.Errorf("below-mode render must not include the right-split divider, got %q", out)
	}
}

// T-088: EntryListHeight ignores DetailPaneHeight in right-split.
func TestLayout_RightSplit_EntryListHeightFull(t *testing.T) {
	l := NewLayout(120, 24, true, 8)
	l.Orientation = OrientationRight
	// In right-split the detail pane is alongside, not below — list height
	// should be height - header - status = 22, NOT reduced by 8.
	if l.EntryListHeight() != 22 {
		t.Errorf("right-split EntryListHeight: got %d, want 22", l.EntryListHeight())
	}
}

// T-090: IsBelowMinFloor predicate.
func TestLayout_IsBelowMinFloor(t *testing.T) {
	cases := []struct {
		w, h int
		want bool
	}{
		{60, 15, false},
		{60, 14, true},
		{59, 15, true},
		{120, 40, false},
		{0, 0, true},
	}
	for _, tc := range cases {
		m := NewLayoutModel(tc.w, tc.h)
		if got := m.IsBelowMinFloor(); got != tc.want {
			t.Errorf("IsBelowMinFloor(%dx%d) = %v, want %v", tc.w, tc.h, got, tc.want)
		}
	}
}

// T-123 (F-013): In right-split the detail pane gets the full main-area
// slot (terminal_height - header - status). height_ratio must NOT be
// applied to the vertical dimension.
func TestDetailPaneVerticalRows_RightUsesFullSlot(t *testing.T) {
	// 80x24 terminal, detail pane open at height_ratio 0.30 → below-mode
	// PaneHeight == 7. Right-mode must override to 22 (24 - 1 - 1).
	l := NewLayout(80, 24, true, 7)
	l.Orientation = OrientationRight
	if got := DetailPaneVerticalRows(l); got != 22 {
		t.Errorf("right-split vertical rows: got %d, want 22", got)
	}
}

// T-123: In below-mode the function preserves DetailPaneHeight (height_ratio).
func TestDetailPaneVerticalRows_BelowUsesRatio(t *testing.T) {
	l := NewLayout(80, 24, true, 7)
	l.Orientation = OrientationBelow
	if got := DetailPaneVerticalRows(l); got != 7 {
		t.Errorf("below-mode vertical rows: got %d, want 7", got)
	}
}

// T-123: closed pane returns 0 in both orientations.
func TestDetailPaneVerticalRows_ClosedReturnsZero(t *testing.T) {
	l := NewLayout(80, 24, false, 0)
	if got := DetailPaneVerticalRows(l); got != 0 {
		t.Errorf("closed below: got %d, want 0", got)
	}
	l.Orientation = OrientationRight
	if got := DetailPaneVerticalRows(l); got != 0 {
		t.Errorf("closed right: got %d, want 0", got)
	}
}

// T-123: degenerate dimensions (header+status exceed height) still return ≥ 1.
func TestDetailPaneVerticalRows_FloorAtOne(t *testing.T) {
	l := NewLayout(80, 2, true, 0)
	l.Orientation = OrientationRight
	if got := DetailPaneVerticalRows(l); got < 1 {
		t.Errorf("degenerate right: got %d, want ≥ 1", got)
	}
}

// ---------- T-158: single-owner click-row resolver (cavekit-entry-list R10) ----------

// TestLayout_ListContentTopY_BelowMode: header(1) + list top border(1) = 2.
func TestLayout_ListContentTopY_BelowMode(t *testing.T) {
	l := NewLayout(80, 24, true, 7)
	if got := l.ListContentTopY(); got != 2 {
		t.Errorf("below-mode ListContentTopY: got %d, want 2", got)
	}
}

// TestLayout_ListContentTopY_RightMode: right-split also has header+border,
// so the offset is still 2.
func TestLayout_ListContentTopY_RightMode(t *testing.T) {
	l := NewLayout(200, 24, true, 0)
	l.Orientation = OrientationRight
	if got := l.ListContentTopY(); got != 2 {
		t.Errorf("right-mode ListContentTopY: got %d, want 2", got)
	}
}

// TestLayout_ListContentTopY_PaneClosed: closed-pane layouts still have a
// header and a list top border — offset stays 2.
func TestLayout_ListContentTopY_PaneClosed(t *testing.T) {
	l := NewLayout(80, 24, false, 0)
	if got := l.ListContentTopY(); got != 2 {
		t.Errorf("pane-closed ListContentTopY: got %d, want 2", got)
	}
}

// TestLayout_ClickToListRow_FirstContentRow: y = ListContentTopY → row 0.
func TestLayout_ClickToListRow_FirstContentRow(t *testing.T) {
	l := NewLayout(80, 24, true, 7) // EntryListHeight = 15, viewport = 13
	row, ok := l.ClickToListRow(2)
	if !ok {
		t.Fatalf("y=2 should be in content, got ok=false")
	}
	if row != 0 {
		t.Errorf("y=2: got row %d, want 0", row)
	}
}

// TestLayout_ClickToListRow_SequentialRows: each +1 terminal Y advances the
// row by one.
func TestLayout_ClickToListRow_SequentialRows(t *testing.T) {
	l := NewLayout(80, 24, true, 7)
	for y, wantRow := range map[int]int{2: 0, 3: 1, 5: 3, 14: 12} {
		row, ok := l.ClickToListRow(y)
		if !ok {
			t.Errorf("y=%d: expected in-content, got ok=false", y)
			continue
		}
		if row != wantRow {
			t.Errorf("y=%d: got row %d, want %d", y, row, wantRow)
		}
	}
}

// TestLayout_ClickToListRow_HeaderIsOutsideContent: y=0 (header) returns
// ok=false.
func TestLayout_ClickToListRow_HeaderIsOutsideContent(t *testing.T) {
	l := NewLayout(80, 24, true, 7)
	if _, ok := l.ClickToListRow(0); ok {
		t.Error("y=0 (header) should return ok=false")
	}
}

// TestLayout_ClickToListRow_TopBorderIsOutsideContent: y=1 (list top border)
// returns ok=false.
func TestLayout_ClickToListRow_TopBorderIsOutsideContent(t *testing.T) {
	l := NewLayout(80, 24, true, 7)
	if _, ok := l.ClickToListRow(1); ok {
		t.Error("y=1 (list top border) should return ok=false")
	}
}

// TestLayout_ClickToListRow_BottomBorderIsOutsideContent: the bottom border
// row of the list pane is out of content.
func TestLayout_ClickToListRow_BottomBorderIsOutsideContent(t *testing.T) {
	// EntryListHeight = 24 - 2 - 7 = 15 → content rows = 13 → last content y = 2+12 = 14
	// Bottom border sits at entryListEnd = 15.
	l := NewLayout(80, 24, true, 7)
	if _, ok := l.ClickToListRow(15); ok {
		t.Error("y=15 (list bottom border) should return ok=false")
	}
}

// TestLayout_ClickToListRow_DividerIsOutsideContent: below-mode divider row
// returns ok=false.
func TestLayout_ClickToListRow_DividerIsOutsideContent(t *testing.T) {
	l := NewLayout(80, 24, true, 7)
	// Divider row in below-mode = entryListEnd+1 = 16.
	if _, ok := l.ClickToListRow(16); ok {
		t.Error("y=16 (divider) should return ok=false")
	}
}

// TestLayout_ClickToListRow_DegenerateHeight: a layout with no room for
// content returns ok=false for any Y.
func TestLayout_ClickToListRow_DegenerateHeight(t *testing.T) {
	l := NewLayout(80, 4, true, 0)
	for y := 0; y < 4; y++ {
		if _, ok := l.ClickToListRow(y); ok {
			t.Errorf("degenerate height: y=%d should be ok=false", y)
		}
	}
}
