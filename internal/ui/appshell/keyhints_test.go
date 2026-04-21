package appshell

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/istonikula/gloggy/internal/theme"
)

func defaultKeyHints() KeyHintBarModel {
	// Use wide width so truncation doesn't hide hints in tests.
	return NewKeyHintBarModel(theme.GetTheme("tokyo-night"), 200)
}

// T-050: R4.1 — entry list focus shows entry-list bindings.
func TestKeyHintBar_EntryListFocus(t *testing.T) {
	m := defaultKeyHints().WithFocus(FocusEntryList)
	v := m.View()
	assert.Containsf(t, v, "j", "expected j/k hint for entry list focus: %q", v)
}

// T-050: R4.2 — detail pane focus shows detail pane bindings.
func TestKeyHintBar_DetailPaneFocus(t *testing.T) {
	m := defaultKeyHints().WithFocus(FocusDetailPane)
	v := m.View()
	assert.Containsf(t, v, "/", "expected search hint for detail pane focus: %q", v)
}

// T-050: R4.3 — filter panel focus shows filter bindings.
func TestKeyHintBar_FilterPanelFocus(t *testing.T) {
	m := defaultKeyHints().WithFocus(FocusFilterPanel)
	v := m.View()
	assert.Containsf(t, v, "Space", "expected Space hint for filter panel focus: %q", v)
}

// T-050: R4.4 — hints update immediately on focus change.
func TestKeyHintBar_FocusChangeUpdatesHints(t *testing.T) {
	m := defaultKeyHints().WithFocus(FocusEntryList)
	v1 := m.View()
	m2 := m.WithFocus(FocusDetailPane)
	v2 := m2.View()
	assert.NotEqualf(t, v1, v2, "hints should change between focus modes")
}

// T-092: focus label appears when the detail pane is visible.
func TestKeyHintBar_FocusLabel_ListFocus(t *testing.T) {
	m := defaultKeyHints().WithPaneOpen(true).WithFocus(FocusEntryList)
	v := m.View()
	assert.Containsf(t, v, "focus: list", "expected 'focus: list' label, got %q", v)
}

// T-144 (cavekit-app-shell R13 revised): the `/` hint in the entry-list
// focus advertises list-scope search — pane state no longer matters
// because list search is always available when the list is focused.
func TestKeyHintBar_Slash_EntryList_PaneOpen(t *testing.T) {
	m := defaultKeyHints().WithFocus(FocusEntryList).WithPaneOpen(true)
	v := m.View()
	assert.Containsf(t, v, "/: search list", "expected '/: search list' hint when pane is open: %q", v)
}

// T-144: `/` still advertises list-scope search when pane is closed —
// list is the only pane, and its own search is what activates.
func TestKeyHintBar_Slash_EntryList_PaneClosed(t *testing.T) {
	m := defaultKeyHints().WithFocus(FocusEntryList).WithPaneOpen(false)
	v := m.View()
	assert.Containsf(t, v, "/: search list", "expected '/: search list' hint when pane is closed: %q", v)
}

// T-144: when the detail pane is focused, `/` advertises pane-scope search.
func TestKeyHintBar_Slash_DetailPane(t *testing.T) {
	m := defaultKeyHints().WithFocus(FocusDetailPane).WithPaneOpen(true)
	v := m.View()
	assert.Containsf(t, v, "/: search pane", "expected '/: search pane' hint when detail pane is focused: %q", v)
}

// T-121: `/` is hidden while the filter panel is focused — it is a
// literal input character there, not a global activation.
func TestKeyHintBar_Slash_FilterPanel_Hidden(t *testing.T) {
	m := defaultKeyHints().WithFocus(FocusFilterPanel).WithPaneOpen(true)
	v := m.View()
	assert.NotContainsf(t, v, "/: search", "`/` hint should be hidden while filter panel is focused: %q", v)
}

// T-144: help overlay lists `/` under both the entry-list and detail
// pane domains with focus-based descriptions.
func TestHelpOverlay_SlashScopeDescribed(t *testing.T) {
	h := NewHelpOverlayModel().Open()
	v := h.View()
	// Entry-list scope: list-scope free-text search.
	assert.Containsf(t, v, "Search within list", "help overlay should describe list `/` scope: %q", v)
	// Detail pane scope: mentions "inside this pane" + commits note.
	assert.Containsf(t, v, "Search inside this pane", "help overlay should describe `/` scope for detail pane: %q", v)
}

// T-092: focus label reads 'details' when detail pane is focused.
func TestKeyHintBar_FocusLabel_DetailsFocus(t *testing.T) {
	m := defaultKeyHints().WithPaneOpen(true).WithFocus(FocusDetailPane)
	v := m.View()
	assert.Containsf(t, v, "focus: details", "expected 'focus: details' label, got %q", v)
}

// T-092: focus label reads 'filter' when filter panel is focused.
func TestKeyHintBar_FocusLabel_FilterFocus(t *testing.T) {
	m := defaultKeyHints().WithPaneOpen(true).WithFocus(FocusFilterPanel)
	v := m.View()
	assert.Containsf(t, v, "focus: filter", "expected 'focus: filter' label, got %q", v)
}

// T-092: label is omitted in single-pane state (pane closed).
func TestKeyHintBar_FocusLabel_OmittedWhenSinglePane(t *testing.T) {
	m := defaultKeyHints().WithPaneOpen(false).WithFocus(FocusEntryList)
	v := m.View()
	assert.NotContainsf(t, v, "focus:", "focus label must be omitted when only the list is visible, got %q", v)
}

// T-092: label is right-aligned — the hint text appears before it.
func TestKeyHintBar_FocusLabel_RightAligned(t *testing.T) {
	m := defaultKeyHints().WithPaneOpen(true).WithFocus(FocusEntryList)
	v := m.View()
	labelIdx := strings.Index(v, "focus: list")
	require.GreaterOrEqualf(t, labelIdx, 0, "expected label in output, got %q", v)
	// Some hint must precede the label (at minimum "j").
	jIdx := strings.Index(v, "j")
	assert.LessOrEqualf(t, jIdx, labelIdx,
		"label should be right of hints: label@%d, first 'j'@%d; view=%q",
		labelIdx, jIdx, v)
}
