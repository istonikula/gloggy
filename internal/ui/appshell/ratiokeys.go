package appshell

import "math"

// Ratio constants shared by below-mode (height_ratio) and right-mode
// (width_ratio) keymaps. Clamp [0.10, 0.80] inclusive (DESIGN.md §5,
// cavekit-app-shell R12).
const (
	RatioMin     = 0.10
	RatioMax     = 0.80
	RatioStep    = 0.05
	RatioDefault = 0.30
)

// detailPresets is the toggle set for `|` when the detail pane is focused
// (cavekit-app-shell R12 rev 2026-04-18). The former 0.10 and 0.70
// presets are deleted: 0.10 left the detail pane unreadably small and
// 0.70 left the list too narrow for log-scanning.
var detailPresets = []float64{RatioDefault, 0.50}

// listSharePresets is the toggle set for `|` when the entry list is
// focused — expressed as a list-share value (detail ratio = 1 - listShare).
// Share {0.30, 0.50} ⇔ detail {0.70, 0.50}.
var listSharePresets = []float64{RatioDefault, 0.50}

// NextDetailRatio computes the new detail-pane ratio for a resize key
// press based on which pane is focused (cavekit-app-shell R12, revised).
//
//   - `+` grows the focused pane's share by RatioStep; with list focus
//     this means the detail ratio shrinks by RatioStep.
//   - `-` shrinks the focused pane's share by RatioStep.
//   - `|` toggles the focused pane's share between 0.30 and 0.50.
//   - `=` sets the detail ratio to RatioDefault (0.30) regardless of
//     focus — "reset" is a global return-to-baseline.
//
// All results clamp to [RatioMin, RatioMax]; at the boundary a further
// press in the same direction is a no-op (the returned ratio equals
// `current`). Returns (current, false) for unknown keys.
func NextDetailRatio(current float64, key string, listFocused bool) (float64, bool) {
	switch key {
	case "=":
		return RatioDefault, true
	case "|":
		if listFocused {
			return cycleListSharePreset(current), true
		}
		return cycleDetailPreset(current), true
	case "+":
		delta := RatioStep
		if listFocused {
			delta = -delta
		}
		return ClampRatio(current + delta), true
	case "-":
		delta := -RatioStep
		if listFocused {
			delta = -delta
		}
		return ClampRatio(current + delta), true
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

// cycleDetailPreset advances the detail ratio to the next preset in
// detailPresets. Uses ±RatioStep/2 tolerance; an off-preset ratio jumps
// to the first preset.
func cycleDetailPreset(current float64) float64 {
	for i, p := range detailPresets {
		if math.Abs(current-p) < RatioStep/2 {
			return detailPresets[(i+1)%len(detailPresets)]
		}
	}
	return detailPresets[0]
}

// cycleListSharePreset advances the list share (= 1 - detail ratio) to
// the next preset in listSharePresets, then returns the corresponding
// detail ratio. An off-preset share jumps to the first preset.
func cycleListSharePreset(currentDetail float64) float64 {
	share := 1 - currentDetail
	for i, p := range listSharePresets {
		if math.Abs(share-p) < RatioStep/2 {
			return 1 - listSharePresets[(i+1)%len(listSharePresets)]
		}
	}
	return 1 - listSharePresets[0]
}

// IsRatioKey reports whether the key string is one handled by NextDetailRatio.
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

// RatioFromDragY converts a vertical cursor row position into the new
// height_ratio for a below-split layout (T-156, cavekit-app-shell R15).
// The divider is the 1-row horizontal border between the entry list and
// the detail pane; rows `y+1..termHeight-2` belong to the detail pane
// (status bar occupies `termHeight-1`). Height ratio storage mirrors
// `detailpane.HeightModel.PaneHeight = int(termHeight * ratio)`, so the
// inverse maps divider-row back to ratio via `(termHeight - y - 2) /
// termHeight`. Result is clamped to [RatioMin, RatioMax].
func RatioFromDragY(y, termHeight int) float64 {
	if termHeight <= 0 {
		return ClampRatio(RatioDefault)
	}
	detail := termHeight - 2 - y
	if detail < 0 {
		detail = 0
	}
	if detail > termHeight {
		detail = termHeight
	}
	return ClampRatio(float64(detail) / float64(termHeight))
}
