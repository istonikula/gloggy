package appshell

import (
	"math"
	"testing"
)

// T-155: + with detail focused grows the detail ratio by one step.
func TestNextDetailRatio_PlusDetailFocus(t *testing.T) {
	r, ok := NextDetailRatio(0.30, "+", false)
	if !ok {
		t.Fatalf("+ not recognised")
	}
	if math.Abs(r-0.35) > 1e-9 {
		t.Errorf("+ detail 0.30: got %.3f, want 0.350", r)
	}
}

// T-155: + with list focused shrinks the detail ratio (grows list share).
func TestNextDetailRatio_PlusListFocus(t *testing.T) {
	r, _ := NextDetailRatio(0.30, "+", true)
	if math.Abs(r-0.25) > 1e-9 {
		t.Errorf("+ list 0.30: got %.3f, want 0.250", r)
	}
}

// T-155: - with detail focused shrinks the detail ratio.
func TestNextDetailRatio_MinusDetailFocus(t *testing.T) {
	r, _ := NextDetailRatio(0.30, "-", false)
	if math.Abs(r-0.25) > 1e-9 {
		t.Errorf("- detail 0.30: got %.3f, want 0.250", r)
	}
}

// T-155: - with list focused grows the detail ratio (shrinks list share).
func TestNextDetailRatio_MinusListFocus(t *testing.T) {
	r, _ := NextDetailRatio(0.30, "-", true)
	if math.Abs(r-0.35) > 1e-9 {
		t.Errorf("- list 0.30: got %.3f, want 0.350", r)
	}
}

// T-155: + at RatioMax with detail focused is a no-op (clamp-pin).
func TestNextDetailRatio_PlusAtMaxDetail_NoOp(t *testing.T) {
	r, _ := NextDetailRatio(RatioMax, "+", false)
	if r != RatioMax {
		t.Errorf("+ at max detail: got %.3f, want %.3f", r, RatioMax)
	}
}

// T-155: + at RatioMin with list focused is a no-op (clamp-pin — list
// cannot grow beyond its maximum share, which means detail cannot drop
// below RatioMin).
func TestNextDetailRatio_PlusAtMinList_NoOp(t *testing.T) {
	r, _ := NextDetailRatio(RatioMin, "+", true)
	if r != RatioMin {
		t.Errorf("+ at min list: got %.3f, want %.3f", r, RatioMin)
	}
}

// T-155: - at RatioMin with detail focused is a no-op.
func TestNextDetailRatio_MinusAtMinDetail_NoOp(t *testing.T) {
	r, _ := NextDetailRatio(RatioMin, "-", false)
	if r != RatioMin {
		t.Errorf("- at min detail: got %.3f, want %.3f", r, RatioMin)
	}
}

// T-155: - at RatioMax with list focused is a no-op.
func TestNextDetailRatio_MinusAtMaxList_NoOp(t *testing.T) {
	r, _ := NextDetailRatio(RatioMax, "-", true)
	if r != RatioMax {
		t.Errorf("- at max list: got %.3f, want %.3f", r, RatioMax)
	}
}

// T-155: = resets detail ratio to RatioDefault regardless of focus.
func TestNextDetailRatio_EqualsResetsDefault_BothFocus(t *testing.T) {
	for _, listFocus := range []bool{false, true} {
		r, _ := NextDetailRatio(0.50, "=", listFocus)
		if r != RatioDefault {
			t.Errorf("= listFocused=%v: got %.3f, want %.3f", listFocus, r, RatioDefault)
		}
	}
}

// T-155: | with detail focused toggles detail ratio between 0.30 ↔ 0.50.
func TestNextDetailRatio_PipeDetailFocus_TogglesPresets(t *testing.T) {
	cases := []struct {
		from float64
		want float64
	}{
		{RatioDefault, 0.50},
		{0.50, RatioDefault},
	}
	for _, tc := range cases {
		got, _ := NextDetailRatio(tc.from, "|", false)
		if math.Abs(got-tc.want) > 1e-9 {
			t.Errorf("| detail from %.2f: got %.2f, want %.2f", tc.from, got, tc.want)
		}
	}
}

// T-155: | with list focused toggles list share between 0.30 ↔ 0.50
// (detail ratio 0.70 ↔ 0.50).
func TestNextDetailRatio_PipeListFocus_TogglesShare(t *testing.T) {
	cases := []struct {
		from float64
		want float64
	}{
		{0.70, 0.50}, // list share 0.30 → 0.50
		{0.50, 0.70}, // list share 0.50 → 0.30
	}
	for _, tc := range cases {
		got, _ := NextDetailRatio(tc.from, "|", true)
		if math.Abs(got-tc.want) > 1e-9 {
			t.Errorf("| list from detail=%.2f: got %.2f, want %.2f", tc.from, got, tc.want)
		}
	}
}

// T-155: | from an off-preset detail ratio jumps to the first preset.
func TestNextDetailRatio_PipeOffPreset_JumpsToFirst(t *testing.T) {
	// Detail focus: off-preset 0.45 → first preset 0.30.
	got, _ := NextDetailRatio(0.45, "|", false)
	if math.Abs(got-RatioDefault) > 1e-9 {
		t.Errorf("| detail off-preset 0.45: got %.3f, want %.3f", got, RatioDefault)
	}
	// List focus: off-preset (detail 0.45 → share 0.55) → first preset
	// share 0.30 → detail 0.70.
	got, _ = NextDetailRatio(0.45, "|", true)
	if math.Abs(got-0.70) > 1e-9 {
		t.Errorf("| list off-preset detail=0.45: got %.3f, want 0.70", got)
	}
}

// T-155: unknown keys are not consumed.
func TestNextDetailRatio_UnknownKey(t *testing.T) {
	got, ok := NextDetailRatio(0.30, "j", false)
	if ok {
		t.Errorf("'j' must not be a ratio key")
	}
	if got != 0.30 {
		t.Errorf("unknown key must not change ratio: got %.2f", got)
	}
}

// T-098: ClampRatio enforces [0.10, 0.80].
func TestClampRatio(t *testing.T) {
	cases := []struct {
		in, want float64
	}{
		{0.05, 0.10},
		{0.10, 0.10},
		{0.50, 0.50},
		{0.80, 0.80},
		{0.95, 0.80},
	}
	for _, tc := range cases {
		if got := ClampRatio(tc.in); got != tc.want {
			t.Errorf("ClampRatio(%.2f): got %.2f, want %.2f", tc.in, got, tc.want)
		}
	}
}

// T-098: IsRatioKey identifies the ratio-keymap keys.
func TestIsRatioKey(t *testing.T) {
	for _, k := range []string{"+", "-", "=", "|"} {
		if !IsRatioKey(k) {
			t.Errorf("IsRatioKey(%q): want true", k)
		}
	}
	for _, k := range []string{"j", "k", "esc", "tab", " "} {
		if IsRatioKey(k) {
			t.Errorf("IsRatioKey(%q): want false", k)
		}
	}
}

// T-104 / F-133: dragging the divider to x=50 on a 100-wide terminal.
// Forward math: usable = 95, detail content = usable - x = 45, ratio = 45/95.
// (Pre-F-133 formula was `termWidth-x-2 = 48`, which encoded the off-by-3
// inverse-math bug — the pin was updated when the formula was made the
// exact inverse of `DetailContentWidth = usable - ListContentWidth`.)
func TestRatioFromDragX_Mid(t *testing.T) {
	r := RatioFromDragX(50, 100)
	want := 45.0 / 95.0
	if math.Abs(r-want) > 1e-9 {
		t.Errorf("RatioFromDragX(50,100) = %.4f, want %.4f", r, want)
	}
}

// T-104: dragging far left clamps to RatioMax (detail takes most of usable).
func TestRatioFromDragX_ClampMax(t *testing.T) {
	r := RatioFromDragX(2, 100)
	if math.Abs(r-RatioMax) > 1e-9 {
		t.Errorf("RatioFromDragX(2,100) = %.4f, want RatioMax=%.4f", r, RatioMax)
	}
}

// T-104: dragging far right clamps to RatioMin (detail shrinks).
func TestRatioFromDragX_ClampMin(t *testing.T) {
	r := RatioFromDragX(99, 100)
	if math.Abs(r-RatioMin) > 1e-9 {
		t.Errorf("RatioFromDragX(99,100) = %.4f, want RatioMin=%.4f", r, RatioMin)
	}
}

// T-104: dragging is a normalized function of terminal width — at
// proportional positions the ratio matches regardless of absolute width.
func TestRatioFromDragX_NormalizedByWidth(t *testing.T) {
	// x / termWidth ≈ 0.5 at both sizes; ratios should be close.
	r1 := RatioFromDragX(60, 120)
	r2 := RatioFromDragX(100, 200)
	if math.Abs(r1-r2) > 0.02 {
		t.Errorf("expected similar ratios at proportional positions, got %.3f vs %.3f", r1, r2)
	}
}

// F-133: pressing on the divider at its current physical X column must
// re-return the current ratio (no step-snap). Mirrors the Y-axis pin.
// The canonical divider X is sourced from the renderer-truth invariant
// established by R15/T-160 (`Layout.ListContentWidth()`), NOT from the
// inverse formula itself — deriving from the inverse would tautologically
// agree with whatever the inverse computes.
//
// Tolerance RatioStep/2 = 0.025 matches the Y-axis test. Old formula
// (`termWidth - x - 2`) drifts ~0.04 at termWidth=100, ratio=0.55 — well
// outside tolerance.
//
// RatioMin (0.10) and RatioMax (0.80) are excluded: at the boundary,
// ClampRatio rescues both formulas and the test cannot distinguish them.
func TestRatioFromDragX_PressAtCurrentDividerX_KeepsRatio(t *testing.T) {
	cases := []struct {
		termWidth int
		ratio     float64
	}{
		{100, 0.30},
		{100, 0.50},
		{100, 0.55},
		{80, 0.30},
		{80, 0.50},
	}
	for _, tc := range cases {
		// Canonical divider X from renderer-truth: dividerX equals the
		// list pane's content width (R15 AC line 198 / T-160).
		l := NewLayout(tc.termWidth, 24, true, 0)
		l.Orientation = OrientationRight
		l.WidthRatio = tc.ratio
		dividerX := l.ListContentWidth()
		got := RatioFromDragX(dividerX, tc.termWidth)
		if math.Abs(got-tc.ratio) > RatioStep/2+1e-9 {
			t.Errorf("RatioFromDragX(%d, %d) = %.4f; want ≈%.4f (±%.3f) for ratio %.2f",
				dividerX, tc.termWidth, got, tc.ratio, RatioStep/2, tc.ratio)
		}
	}
}

// T-161 (F-123): pressing on the divider at its current physical Y row
// must re-return the current ratio (no step-snap). Uses the forward
// math from `detailpane.HeightModel.PaneHeight = int(termHeight*ratio)`
// to compute the canonical divider Y for each preset, then asserts the
// inverse brings us back within int-truncation tolerance (one-row drift
// per termHeight). The old formula (`termHeight - 2 - y`) was off by a
// full row, so press-at-current-Y would snap the ratio one RatioStep
// down. Tolerance RatioStep/2 = 0.025 is tight enough to catch the bug
// (old drift ≥ 0.042 at termHeight=24) while tolerating the floor-div
// drift of the fixed formula (≤ 0.009 at the same termHeight).
//
// RatioMin (0.10) is excluded: at termHeight=24 both old and new
// formulas clamp to RatioMin, so the test cannot distinguish them there.
func TestRatioFromDragY_PressAtCurrentDividerY_KeepsRatio(t *testing.T) {
	const termHeight = 24
	presets := []float64{0.30, 0.50, 0.80}
	for _, r := range presets {
		paneHeight := int(float64(termHeight) * r)
		if paneHeight < 1 {
			paneHeight = 1
		}
		// Canonical divider Y: detail occupies rows dividerY..termHeight-2
		// (divider row = detail's top border), so paneHeight = termHeight
		// - 1 - dividerY, i.e. dividerY = termHeight - 1 - paneHeight.
		dividerY := termHeight - 1 - paneHeight
		got := RatioFromDragY(dividerY, termHeight)
		if math.Abs(got-r) > RatioStep/2+1e-9 {
			t.Errorf("RatioFromDragY(%d, %d) = %.4f; want ≈%.4f (±%.3f) for ratio %.2f",
				dividerY, termHeight, got, r, RatioStep/2, r)
		}
	}
}
