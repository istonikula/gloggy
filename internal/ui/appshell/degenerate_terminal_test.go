package appshell

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

// B39: RatioFromDragX must guard usable <= 0 so a zero/degenerate terminal
// width cannot produce a +Inf ratio (division by zero) that ClampRatio only
// masks after the fact. Returns the clamped default instead.
func TestRatioFromDragX_DegenerateWidth_NoInf_B39(t *testing.T) {
	for _, w := range []int{0, 1, 4, 5} { // all yield usable <= 0
		got := RatioFromDragX(50, w)
		assert.Falsef(t, math.IsInf(got, 0), "width %d must not yield ±Inf", w)
		assert.Falsef(t, math.IsNaN(got), "width %d must not yield NaN", w)
		assert.Equalf(t, ClampRatio(RatioDefault), got,
			"degenerate width %d falls back to the clamped default", w)
	}
}

// B40: in right-split, a degenerate terminal where ListContentWidth() == 0
// would make listRightBorder = -1 and collapse the `x < listRightBorder` list
// zone, routing every click into the divider/detail buffer. The width guard
// returns ZoneUnknown instead.
func TestMouseRouter_RightSplit_ZeroListWidth_ZoneUnknown_B40(t *testing.T) {
	l := Layout{
		Width: 4, Height: 24, HeaderHeight: 1, StatusBarHeight: 1,
		DetailPaneOpen: true, Orientation: OrientationRight, WidthRatio: 0.30,
	}
	require := l.ListContentWidth()
	assert.Equal(t, 0, require, "setup: ListContentWidth must be 0 for this layout")

	r := NewMouseRouter(l)
	for _, x := range []int{0, 1, 2, 3} {
		assert.Equalf(t, ZoneUnknown, r.Zone(x, 5),
			"degenerate zero-width list must route x=%d to ZoneUnknown", x)
	}
}
