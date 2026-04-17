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

// RatioFromDragX converts a horizontal cursor column position into the new
// width_ratio for a right-split layout (T-104). The terminal x column is
// normalized to the usable split width (terminalWidth - 2*paneBorders -
// dividerWidth) so the divider follows the cursor. The returned ratio
// corresponds to the DETAIL pane slice (detailContent = usable * ratio)
// and is clamped to [RatioMin, RatioMax].
func RatioFromDragX(x, termWidth int) float64 {
	usable := termWidth - 2*paneBorders - dividerWidth
	if usable <= 0 {
		return ClampRatio(RatioDefault)
	}
	// The detail pane occupies the right-hand slice starting after the
	// divider + its left border. Its content width is therefore
	// (termWidth - x - 2): two cells reserved for the detail pane's
	// right border and the divider-to-right-border gap.
	detail := termWidth - x - 2
	if detail < 0 {
		detail = 0
	}
	if detail > usable {
		detail = usable
	}
	return ClampRatio(float64(detail) / float64(usable))
}
