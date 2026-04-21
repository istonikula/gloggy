package appshell

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// T-096: Tab cycles list → details when both visible.
func TestNextFocus_ListToDetails(t *testing.T) {
	visible := []FocusTarget{FocusEntryList, FocusDetailPane}
	got := NextFocus(FocusEntryList, visible, false)
	assert.Equalf(t, FocusDetailPane, got, "NextFocus(list, [list,details], false) = %v, want details", got)
}

// T-096: Tab cycles details → list (wraps).
func TestNextFocus_DetailsToList(t *testing.T) {
	visible := []FocusTarget{FocusEntryList, FocusDetailPane}
	got := NextFocus(FocusDetailPane, visible, false)
	assert.Equalf(t, FocusEntryList, got, "NextFocus(details, [list,details], false) = %v, want list", got)
}

// T-096: Tab is inert when overlay (filter/help) is open.
func TestNextFocus_OverlayOpenIsInert(t *testing.T) {
	visible := []FocusTarget{FocusEntryList, FocusDetailPane}
	got := NextFocus(FocusEntryList, visible, true)
	assert.Equalf(t, FocusEntryList, got, "NextFocus with overlayOpen=true must be inert, got %v", got)
	got2 := NextFocus(FocusDetailPane, visible, true)
	assert.Equalf(t, FocusDetailPane, got2, "NextFocus with overlayOpen=true must be inert, got %v", got2)
}

// T-096: Tab is a no-op when only one pane is visible.
func TestNextFocus_SinglePaneIsNoOp(t *testing.T) {
	visible := []FocusTarget{FocusEntryList}
	got := NextFocus(FocusEntryList, visible, false)
	assert.Equalf(t, FocusEntryList, got, "NextFocus on single pane = %v, want list", got)
}

// T-096: Tab on empty visible set is a no-op.
func TestNextFocus_EmptyVisibleIsNoOp(t *testing.T) {
	got := NextFocus(FocusEntryList, nil, false)
	assert.Equalf(t, FocusEntryList, got, "NextFocus on empty visible = %v, want list (no change)", got)
}

// T-096: focus current not in visible jumps to first visible (e.g. pane just closed).
func TestNextFocus_CurrentNotInVisible(t *testing.T) {
	visible := []FocusTarget{FocusEntryList}
	got := NextFocus(FocusDetailPane, visible, false)
	assert.Equalf(t, FocusEntryList, got, "NextFocus when current not in visible = %v, want list", got)
}
