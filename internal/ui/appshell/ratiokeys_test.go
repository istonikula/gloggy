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

// T-104: dragging the divider to x=50 on a 100-wide terminal halves the
// usable space → ratio ≈ 48/95 (clamped to RatioMax=0.80).
// At termWidth=100 and x=50: detail = 100-50-2 = 48, usable = 95, ratio = 48/95 ≈ 0.505.
func TestRatioFromDragX_Mid(t *testing.T) {
	r := RatioFromDragX(50, 100)
	want := 48.0 / 95.0
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
