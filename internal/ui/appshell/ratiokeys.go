package appshell

import "math"

// Ratio constants shared by below-mode (height_ratio) and right-mode
// (width_ratio) keymaps. Clamp [0.10, 0.80] inclusive (DESIGN.md §5).
const (
	RatioMin     = 0.10
	RatioMax     = 0.80
	RatioStep    = 0.05
	RatioDefault = 0.30
)

// ratioPresets is the cycle order for the `|` key (DESIGN.md §5).
var ratioPresets = []float64{0.10, 0.30, 0.70}

// NextRatio computes the new ratio for a single key press. Returns
// (newRatio, true) if the key is a ratio key (`+`, `-`, `=`, `|`), or
// (current, false) otherwise. All adjustments clamp to [RatioMin, RatioMax]
// (T-098).
func NextRatio(current float64, key string) (float64, bool) {
	switch key {
	case "+":
		return ClampRatio(current + RatioStep), true
	case "-":
		return ClampRatio(current - RatioStep), true
	case "=":
		return RatioDefault, true
	case "|":
		return cycleRatioPreset(current), true
	}
	return current, false
}

// ClampRatio clamps a ratio into the inclusive [RatioMin, RatioMax] range.
func ClampRatio(r float64) float64 {
	if r < RatioMin {
		return RatioMin
	}
	if r > RatioMax {
		return RatioMax
	}
	return r
}

// cycleRatioPreset advances the ratio to the next preset in
// ratioPresets. Matches the current ratio against presets within ±step/2;
// if no preset matches, jumps to the first preset.
func cycleRatioPreset(current float64) float64 {
	for i, p := range ratioPresets {
		if math.Abs(current-p) < RatioStep/2 {
			return ratioPresets[(i+1)%len(ratioPresets)]
		}
	}
	return ratioPresets[0]
}

// IsRatioKey reports whether the key string is one handled by NextRatio.
func IsRatioKey(key string) bool {
	switch key {
	case "+", "-", "=", "|":
		return true
	}
	return false
}
