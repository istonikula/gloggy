package entrylist

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// V41: visibleIndexForEntry ∘ entryIndexForVisible == identity across ALL pin
// states. `expected` is ground-truth (the involution property itself + an
// independently-derived visible layout), never the impl's own arithmetic
// re-run (V34).

// buildVisibleGroundTruth independently reconstructs the visible-list → full
// index mapping WITHOUT calling entryIndexForVisible/visibleIndexForEntry, so
// it is a true reference per V34(a). Returns one full-entries index per
// visible row, in visible order.
func buildVisibleGroundTruth(total int, filtered []int, pin int) []int {
	if filtered == nil {
		out := make([]int, total)
		for i := range out {
			out[i] = i
		}
		return out
	}
	// Pin shows only when it is set, in range, and NOT already in filtered.
	pinShows := pin >= 0 && pin < total
	if pinShows {
		for _, fi := range filtered {
			if fi == pin {
				pinShows = false
				break
			}
		}
	}
	if !pinShows {
		out := make([]int, len(filtered))
		copy(out, filtered)
		return out
	}
	// Splice pin at its sorted position (first filtered idx greater than pin).
	out := make([]int, 0, len(filtered)+1)
	inserted := false
	for _, fi := range filtered {
		if !inserted && fi > pin {
			out = append(out, pin)
			inserted = true
		}
		out = append(out, fi)
	}
	if !inserted {
		out = append(out, pin)
	}
	return out
}

func TestIdxRoundTrip_AllPinStates_V41(t *testing.T) {
	const total = 10
	cases := []struct {
		name     string
		filtered []int
		pin      int
	}{
		{"no-filter-no-pin", nil, -1},
		{"filter-no-pin", []int{1, 3, 5, 7}, -1},
		{"pin-in-filter", []int{1, 3, 5, 7}, 3},
		{"pin-out-of-filter-middle", []int{1, 3, 5, 7}, 4},
		{"pin-out-of-filter-before-all", []int{3, 5, 7}, 1},
		{"pin-out-of-filter-after-all", []int{1, 3, 5}, 8},
		{"pin-at-edge-zero", []int{2, 4, 6}, 0},
		{"pin-at-edge-last", []int{2, 4, 6}, 9},
		{"single-filtered-pin-out", []int{4}, 7},
		{"empty-filter-pin-out", []int{}, 5},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := defaultListModel(20).SetEntries(makeEntries(total))
			if tc.filtered != nil {
				m = m.SetFilter(tc.filtered)
			}
			m.pinnedFullIdx = tc.pin

			ground := buildVisibleGroundTruth(total, tc.filtered, tc.pin)
			vis, _ := m.visibleEntriesAndPin()
			require.Lenf(t, vis, len(ground),
				"visible length must match ground-truth layout (%v)", ground)

			for vIdx, wantFull := range ground {
				// (a) entryIndexForVisible matches the independent ground truth.
				gotFull := m.entryIndexForVisible(vIdx)
				assert.Equalf(t, wantFull, gotFull,
					"entryIndexForVisible(%d): ground-truth full idx", vIdx)

				// (b) round-trip identity: visible→full→visible == input.
				back := m.visibleIndexForEntry(gotFull)
				assert.Equalf(t, vIdx, back,
					"round-trip vis %d → full %d → vis %d (identity broken)",
					vIdx, gotFull, back)
			}
		})
	}
}

// V41 cursor contract: empty filtered list ⇒ CursorPosition() == 0, NOT 1.
func TestCursorPosition_EmptyFiltered_Zero_V41(t *testing.T) {
	m := defaultListModel(20).SetEntries(makeEntries(5))
	m = m.SetFilter([]int{}) // filter excludes everything
	require.Empty(t, m.visibleEntries(), "filtered set must be empty for this case")
	assert.Equal(t, 0, m.CursorPosition(), "empty filtered list ⇒ position 0")
}

// V41 cursor contract: empty entry list ⇒ idx helpers + selection report the
// absent ("present=false") form without panic.
func TestEmptyEntryList_PresentFalse_V41(t *testing.T) {
	m := defaultListModel(20) // no entries
	_, ok := m.SelectedEntry()
	assert.False(t, ok, "empty list ⇒ SelectedEntry present=false")
	assert.Equal(t, 0, m.CursorPosition(), "empty list ⇒ position 0")
	// idx helpers must not panic on a degenerate (no-filter) empty model.
	assert.NotPanics(t, func() {
		_ = m.entryIndexForVisible(0)
		_ = m.visibleIndexForEntry(0)
	})
}

func TestBuildVisibleGroundTruth_SelfCheck(t *testing.T) {
	// Sanity: the reference reconstructor itself is correct on a hand case.
	assert.Equal(t, []int{1, 3, 4, 5, 7},
		buildVisibleGroundTruth(10, []int{1, 3, 5, 7}, 4),
		fmt.Sprintf("pin 4 splices between 3 and 5"))
}
