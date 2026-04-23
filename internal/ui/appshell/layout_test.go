package appshell

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T-046: R2.1 — header at top, entrylist in middle, status at bottom.
func TestLayout_Render_WithoutDetailPane(t *testing.T) {
	m := NewLayoutModel(80, 24)
	out := m.Render("HEADER", "ENTRYLIST", "DETAIL", "STATUS")
	lines := strings.Split(out, "\n")
	assert.Equalf(t, "HEADER", strings.TrimSpace(lines[0]), "first line should be HEADER")
	assert.Equalf(t, "ENTRYLIST", strings.TrimSpace(lines[1]), "second line should be ENTRYLIST")
	assert.Equalf(t, "STATUS", strings.TrimSpace(lines[len(lines)-1]), "last line should be STATUS")
	// Detail pane should not appear when closed.
	assert.NotContainsf(t, out, "DETAIL", "DETAIL should not appear when pane is closed")
}

// T-046: R2.3 — detail pane appears between entry list and status bar when open.
func TestLayout_Render_WithDetailPane(t *testing.T) {
	m := NewLayoutModel(80, 24).SetDetailPane(true, 5)
	out := m.Render("HEADER", "ENTRYLIST", "DETAIL", "STATUS")
	lines := strings.Split(out, "\n")
	// Should be: HEADER, ENTRYLIST, DETAIL, STATUS
	assert.Equalf(t, "HEADER", strings.TrimSpace(lines[0]), "line 0: want HEADER")
	assert.Equalf(t, "ENTRYLIST", strings.TrimSpace(lines[1]), "line 1: want ENTRYLIST")
	assert.Equalf(t, "DETAIL", strings.TrimSpace(lines[2]), "line 2: want DETAIL")
	assert.Equalf(t, "STATUS", strings.TrimSpace(lines[3]), "line 3: want STATUS")
}

// T-046: R2.5 — entry list height fills available space.
func TestLayout_EntryListHeight_FillsSpace(t *testing.T) {
	l := NewLayout(80, 24, false, 0)
	// 24 rows - 1 header - 1 status = 22
	assert.Equalf(t, 22, l.EntryListHeight(), "entry list height")
}

// T-046: R2.3 — entry list height reduced by detail pane when open.
func TestLayout_EntryListHeight_ReducedByDetailPane(t *testing.T) {
	l := NewLayout(80, 24, true, 8)
	// 24 - 1 - 1 - 8 = 14
	assert.Equalf(t, 14, l.EntryListHeight(), "entry list height with detail pane")
}

// SetSize updates dimensions.
func TestLayoutModel_SetSize(t *testing.T) {
	m := NewLayoutModel(80, 24).SetSize(120, 40)
	l := m.Layout()
	assert.Equalf(t, 120, l.Width, "SetSize width")
	assert.Equalf(t, 40, l.Height, "SetSize height")
}

// T-090: terminal-too-small fallback at 59x15.
func TestLayout_FallbackBelowMinWidth(t *testing.T) {
	m := NewLayoutModel(59, 15)
	out := m.Render("HEADER", "ENTRYLIST", "DETAIL", "STATUS")
	assert.Containsf(t, out, "terminal too small", "expected fallback message at 59x15, got %q", out)
	assert.NotContainsf(t, out, "HEADER", "panels should be suppressed at 59x15, got %q", out)
	assert.NotContainsf(t, out, "ENTRYLIST", "panels should be suppressed at 59x15, got %q", out)
}

// T-090: terminal-too-small fallback at 60x14.
func TestLayout_FallbackBelowMinHeight(t *testing.T) {
	m := NewLayoutModel(60, 14)
	out := m.Render("HEADER", "ENTRYLIST", "DETAIL", "STATUS")
	assert.Containsf(t, out, "terminal too small", "expected fallback message at 60x14, got %q", out)
	assert.NotContainsf(t, out, "HEADER", "panels should be suppressed at 60x14, got %q", out)
}

// T-090: normal render resumes at 60x15.
func TestLayout_NormalRenderAtMinFloor(t *testing.T) {
	m := NewLayoutModel(60, 15)
	out := m.Render("HEADER", "ENTRYLIST", "DETAIL", "STATUS")
	assert.NotContainsf(t, out, "terminal too small", "normal render should resume at 60x15, got %q", out)
	assert.Containsf(t, out, "HEADER", "HEADER should appear at 60x15, got %q", out)
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
		assert.Equalf(t, tc.termW, got,
			"widths+chrome (termW=%d, ratio=%.2f): list=%d, detail=%d",
			tc.termW, tc.ratio, listW, detailW)
	}
}

// T-088: Width-ratio formula matches DESIGN.md §5 example exactly
// (termWidth=120, widthRatio=0.30 ⇒ usable=115, listW=80, detailW=35).
func TestLayout_RightSplit_DesignExampleNumbers(t *testing.T) {
	l := NewLayout(120, 24, true, 0)
	l.Orientation = OrientationRight
	l.WidthRatio = 0.30
	assert.Equalf(t, 80, l.ListContentWidth(), "ListContentWidth(120, 0.30)")
	assert.Equalf(t, 35, l.DetailContentWidth(), "DetailContentWidth(120, 0.30)")
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
	assert.Equalf(t, 120, lipgloss.Width(mainLine), "main line width; line=%q", mainLine)
}

// T-088: Below-mode keeps the existing vertical composition unchanged.
func TestLayoutModel_Render_BelowMode_Unchanged(t *testing.T) {
	m := NewLayoutModel(80, 24).SetDetailPane(true, 5) // orientation defaults to below
	out := m.Render("HEADER", "ENTRYLIST", "DETAIL", "STATUS")
	assert.Containsf(t, out, "DETAIL", "below-mode render must include DETAIL line")
	// Should not contain the divider glyph.
	assert.NotContainsf(t, out, "│", "below-mode render must not include the right-split divider, got %q", out)
}

// T-088: EntryListHeight ignores DetailPaneHeight in right-split.
func TestLayout_RightSplit_EntryListHeightFull(t *testing.T) {
	l := NewLayout(120, 24, true, 8)
	l.Orientation = OrientationRight
	// In right-split the detail pane is alongside, not below — list height
	// should be height - header - status = 22, NOT reduced by 8.
	assert.Equalf(t, 22, l.EntryListHeight(), "right-split EntryListHeight")
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
		assert.Equalf(t, tc.want, m.IsBelowMinFloor(), "IsBelowMinFloor(%dx%d)", tc.w, tc.h)
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
	assert.Equalf(t, 22, DetailPaneVerticalRows(l), "right-split vertical rows")
}

// T-123: In below-mode the function preserves DetailPaneHeight (height_ratio).
func TestDetailPaneVerticalRows_BelowUsesRatio(t *testing.T) {
	l := NewLayout(80, 24, true, 7)
	l.Orientation = OrientationBelow
	assert.Equalf(t, 7, DetailPaneVerticalRows(l), "below-mode vertical rows")
}

// T-123: closed pane returns 0 in both orientations.
func TestDetailPaneVerticalRows_ClosedReturnsZero(t *testing.T) {
	l := NewLayout(80, 24, false, 0)
	assert.Equalf(t, 0, DetailPaneVerticalRows(l), "closed below")
	l.Orientation = OrientationRight
	assert.Equalf(t, 0, DetailPaneVerticalRows(l), "closed right")
}

// T-123: degenerate dimensions (header+status exceed height) still return ≥ 1.
func TestDetailPaneVerticalRows_FloorAtOne(t *testing.T) {
	l := NewLayout(80, 2, true, 0)
	l.Orientation = OrientationRight
	assert.GreaterOrEqualf(t, DetailPaneVerticalRows(l), 1, "degenerate right")
}

// ---------- T-158: single-owner click-row resolver (cavekit-entry-list R10) ----------

// TestLayout_ListContentTopY_BelowMode: header(1) + list top border(1) = 2.
func TestLayout_ListContentTopY_BelowMode(t *testing.T) {
	l := NewLayout(80, 24, true, 7)
	assert.Equalf(t, 2, l.ListContentTopY(), "below-mode ListContentTopY")
}

// TestLayout_ListContentTopY_RightMode: right-split also has header+border,
// so the offset is still 2.
func TestLayout_ListContentTopY_RightMode(t *testing.T) {
	l := NewLayout(200, 24, true, 0)
	l.Orientation = OrientationRight
	assert.Equalf(t, 2, l.ListContentTopY(), "right-mode ListContentTopY")
}

// TestLayout_ListContentTopY_PaneClosed: closed-pane layouts still have a
// header and a list top border — offset stays 2.
func TestLayout_ListContentTopY_PaneClosed(t *testing.T) {
	l := NewLayout(80, 24, false, 0)
	assert.Equalf(t, 2, l.ListContentTopY(), "pane-closed ListContentTopY")
}

// TestLayout_ClickToListRow_FirstContentRow: y = ListContentTopY → row 0.
func TestLayout_ClickToListRow_FirstContentRow(t *testing.T) {
	l := NewLayout(80, 24, true, 7) // EntryListHeight = 15, viewport = 13
	row, ok := l.ClickToListRow(2)
	require.Truef(t, ok, "y=2 should be in content, got ok=false")
	assert.Equalf(t, 0, row, "y=2 row")
}

// TestLayout_ClickToListRow_SequentialRows: each +1 terminal Y advances the
// row by one.
func TestLayout_ClickToListRow_SequentialRows(t *testing.T) {
	l := NewLayout(80, 24, true, 7)
	for y, wantRow := range map[int]int{2: 0, 3: 1, 5: 3, 14: 12} {
		row, ok := l.ClickToListRow(y)
		if !assert.Truef(t, ok, "y=%d: expected in-content, got ok=false", y) {
			continue
		}
		assert.Equalf(t, wantRow, row, "y=%d row", y)
	}
}

// TestLayout_ClickToListRow_HeaderIsOutsideContent: y=0 (header) returns
// ok=false.
func TestLayout_ClickToListRow_HeaderIsOutsideContent(t *testing.T) {
	l := NewLayout(80, 24, true, 7)
	_, ok := l.ClickToListRow(0)
	assert.Falsef(t, ok, "y=0 (header) should return ok=false")
}

// TestLayout_ClickToListRow_TopBorderIsOutsideContent: y=1 (list top border)
// returns ok=false.
func TestLayout_ClickToListRow_TopBorderIsOutsideContent(t *testing.T) {
	l := NewLayout(80, 24, true, 7)
	_, ok := l.ClickToListRow(1)
	assert.Falsef(t, ok, "y=1 (list top border) should return ok=false")
}

// TestLayout_ClickToListRow_BottomBorderIsOutsideContent: the bottom border
// row of the list pane is out of content.
func TestLayout_ClickToListRow_BottomBorderIsOutsideContent(t *testing.T) {
	// EntryListHeight = 24 - 2 - 7 = 15 → content rows = 13 → last content y = 2+12 = 14
	// Bottom border sits at entryListEnd = 15.
	l := NewLayout(80, 24, true, 7)
	_, ok := l.ClickToListRow(15)
	assert.Falsef(t, ok, "y=15 (list bottom border) should return ok=false")
}

// TestLayout_ClickToListRow_DividerIsOutsideContent: below-mode divider row
// returns ok=false.
func TestLayout_ClickToListRow_DividerIsOutsideContent(t *testing.T) {
	l := NewLayout(80, 24, true, 7)
	// Divider row in below-mode = entryListEnd+1 = 16.
	_, ok := l.ClickToListRow(16)
	assert.Falsef(t, ok, "y=16 (divider) should return ok=false")
}

// TestLayout_ClickToListRow_DegenerateHeight: a layout with no room for
// content returns ok=false for any Y.
func TestLayout_ClickToListRow_DegenerateHeight(t *testing.T) {
	l := NewLayout(80, 4, true, 0)
	for y := 0; y < 4; y++ {
		_, ok := l.ClickToListRow(y)
		assert.Falsef(t, ok, "degenerate height: y=%d should be ok=false", y)
	}
}

// ---------- T28 / V8: single-owner click-to-pane-row resolver ----------

// TestLayout_DetailPaneContentTopY_Below: header(1) + list content rows +
// list borders + divider(1) + pane top border(1). For 80x30, ratio=0.30 →
// DetailPaneHeight=9, EntryListHeight=30-1-1-9=19, DetailPaneContentTopY
// = ListContentTopY(2) + 19 + 1 = 22.
func TestLayout_DetailPaneContentTopY_Below(t *testing.T) {
	l := NewLayout(80, 30, true, 9)
	assert.Equalf(t, 22, l.DetailPaneContentTopY(), "below DetailPaneContentTopY")
}

// TestLayout_DetailPaneContentTopY_Right: right-split places the pane
// directly under the header — only HeaderHeight(1) + pane top border(1).
func TestLayout_DetailPaneContentTopY_Right(t *testing.T) {
	l := NewLayout(200, 30, true, 0)
	l.Orientation = OrientationRight
	assert.Equalf(t, 2, l.DetailPaneContentTopY(), "right DetailPaneContentTopY")
}

// TestLayout_DetailPaneContentTopY_ClosedReturnsZero: closed pane has no
// content row, so the helper returns 0 as a sentinel (callers should
// never call ClickToPaneRow when the pane is closed).
func TestLayout_DetailPaneContentTopY_ClosedReturnsZero(t *testing.T) {
	l := NewLayout(80, 30, false, 0)
	assert.Equalf(t, 0, l.DetailPaneContentTopY(), "closed DetailPaneContentTopY")
}

// TestLayout_ClickToPaneRow_FirstContentRow: Y at DetailPaneContentTopY
// maps to pane-local row 0.
func TestLayout_ClickToPaneRow_FirstContentRow(t *testing.T) {
	l := NewLayout(80, 30, true, 9)
	y, ok := l.ClickToPaneRow(22)
	require.Truef(t, ok, "y=22 (first pane content row) should be in-content")
	assert.Equalf(t, 0, y, "first content row → pane-local 0")
}

// TestLayout_ClickToPaneRow_HeaderOutsideContent: the header row y=0
// maps to out-of-content.
func TestLayout_ClickToPaneRow_HeaderOutsideContent(t *testing.T) {
	l := NewLayout(80, 30, true, 9)
	_, ok := l.ClickToPaneRow(0)
	assert.Falsef(t, ok, "y=0 (header) should return ok=false")
}

// TestLayout_ClickToPaneRow_DividerOutsideContent: the between-pane
// divider row is not inside pane content.
func TestLayout_ClickToPaneRow_DividerOutsideContent(t *testing.T) {
	l := NewLayout(80, 30, true, 9)
	// Divider row = ListContentTopY(2) + EntryListHeight(19) = 21.
	_, ok := l.ClickToPaneRow(21)
	assert.Falsef(t, ok, "y=21 (divider / pane top border boundary) should be out")
}

// TestLayout_ClickToPaneRow_BottomBorderOutsideContent: last content row
// is at start + viewportRows - 1; anything ≥ start + viewportRows is out.
// Pane height 9 → 7 content rows → last content row Y = 22 + 6 = 28.
// Y=29 (status bar) is out.
func TestLayout_ClickToPaneRow_BottomBorderOutsideContent(t *testing.T) {
	l := NewLayout(80, 30, true, 9)
	_, ok := l.ClickToPaneRow(29)
	assert.Falsef(t, ok, "y=29 (pane bottom border / status bar) should be out")
}

// TestLayout_ClickToPaneRow_ClosedPaneReturnsFalse: calling on a closed
// pane returns ok=false for any Y rather than a silent 0-row default.
func TestLayout_ClickToPaneRow_ClosedPaneReturnsFalse(t *testing.T) {
	l := NewLayout(80, 30, false, 0)
	for y := 0; y < 30; y++ {
		_, ok := l.ClickToPaneRow(y)
		assert.Falsef(t, ok, "closed pane: y=%d should be ok=false", y)
	}
}
