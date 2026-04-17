package appshell

import (
	"math"
	"testing"
)

// T-098: + adds RatioStep, three presses from 0.30 → 0.45.
func TestNextRatio_PlusIncrements(t *testing.T) {
	r := 0.30
	for i := 0; i < 3; i++ {
		r, _ = NextRatio(r, "+")
	}
	if math.Abs(r-0.45) > 1e-9 {
		t.Errorf("0.30 + 3*step: got %.3f, want 0.450", r)
	}
}

// T-098: + at the max is a no-op (stays at 0.80).
func TestNextRatio_PlusAtMaxIsNoOp(t *testing.T) {
	r, _ := NextRatio(0.80, "+")
	if r != 0.80 {
		t.Errorf("+ at max: got %.3f, want 0.800", r)
	}
}

// T-098: - subtracts RatioStep.
func TestNextRatio_MinusDecrements(t *testing.T) {
	r, _ := NextRatio(0.30, "-")
	if math.Abs(r-0.25) > 1e-9 {
		t.Errorf("0.30 - step: got %.3f, want 0.250", r)
	}
}

// T-098: - at the min is a no-op (stays at 0.10).
func TestNextRatio_MinusAtMinIsNoOp(t *testing.T) {
	r, _ := NextRatio(0.10, "-")
	if r != 0.10 {
		t.Errorf("- at min: got %.3f, want 0.100", r)
	}
}

// T-098: = resets to default 0.30.
func TestNextRatio_EqualsResetsToDefault(t *testing.T) {
	r, _ := NextRatio(0.50, "=")
	if r != 0.30 {
		t.Errorf("= reset: got %.3f, want 0.300", r)
	}
}

// T-098: | cycles the [0.10, 0.30, 0.70] presets in order, wrapping back to 0.10.
func TestNextRatio_PipeCyclesPresets(t *testing.T) {
	cases := []struct {
		from float64
		want float64
	}{
		{0.10, 0.30},
		{0.30, 0.70},
		{0.70, 0.10},
	}
	for _, tc := range cases {
		got, ok := NextRatio(tc.from, "|")
		if !ok {
			t.Errorf("| not recognised as ratio key")
		}
		if math.Abs(got-tc.want) > 1e-9 {
			t.Errorf("| from %.2f: got %.2f, want %.2f", tc.from, got, tc.want)
		}
	}
}

// T-098: | from a non-preset jumps to the first preset.
func TestNextRatio_PipeFromNonPresetJumpsToFirst(t *testing.T) {
	got, _ := NextRatio(0.45, "|")
	if got != 0.10 {
		t.Errorf("| from off-preset 0.45: got %.2f, want 0.10", got)
	}
}

// T-098: unknown keys are not consumed.
func TestNextRatio_UnknownKey(t *testing.T) {
	got, ok := NextRatio(0.30, "j")
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
