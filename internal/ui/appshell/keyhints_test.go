package appshell

import (
	"strings"
	"testing"

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
	if !strings.Contains(v, "j") {
		t.Errorf("expected j/k hint for entry list focus: %q", v)
	}
}

// T-050: R4.2 — detail pane focus shows detail pane bindings.
func TestKeyHintBar_DetailPaneFocus(t *testing.T) {
	m := defaultKeyHints().WithFocus(FocusDetailPane)
	v := m.View()
	if !strings.Contains(v, "/") {
		t.Errorf("expected search hint for detail pane focus: %q", v)
	}
}

// T-050: R4.3 — filter panel focus shows filter bindings.
func TestKeyHintBar_FilterPanelFocus(t *testing.T) {
	m := defaultKeyHints().WithFocus(FocusFilterPanel)
	v := m.View()
	if !strings.Contains(v, "Space") {
		t.Errorf("expected Space hint for filter panel focus: %q", v)
	}
}

// T-050: R4.4 — hints update immediately on focus change.
func TestKeyHintBar_FocusChangeUpdatesHints(t *testing.T) {
	m := defaultKeyHints().WithFocus(FocusEntryList)
	v1 := m.View()
	m2 := m.WithFocus(FocusDetailPane)
	v2 := m2.View()
	if v1 == v2 {
		t.Error("hints should change between focus modes")
	}
}

// T-092: focus label appears when the detail pane is visible.
func TestKeyHintBar_FocusLabel_ListFocus(t *testing.T) {
	m := defaultKeyHints().WithPaneOpen(true).WithFocus(FocusEntryList)
	v := m.View()
	if !strings.Contains(v, "focus: list") {
		t.Errorf("expected 'focus: list' label, got %q", v)
	}
}

// T-092: focus label reads 'details' when detail pane is focused.
func TestKeyHintBar_FocusLabel_DetailsFocus(t *testing.T) {
	m := defaultKeyHints().WithPaneOpen(true).WithFocus(FocusDetailPane)
	v := m.View()
	if !strings.Contains(v, "focus: details") {
		t.Errorf("expected 'focus: details' label, got %q", v)
	}
}

// T-092: focus label reads 'filter' when filter panel is focused.
func TestKeyHintBar_FocusLabel_FilterFocus(t *testing.T) {
	m := defaultKeyHints().WithPaneOpen(true).WithFocus(FocusFilterPanel)
	v := m.View()
	if !strings.Contains(v, "focus: filter") {
		t.Errorf("expected 'focus: filter' label, got %q", v)
	}
}

// T-092: label is omitted in single-pane state (pane closed).
func TestKeyHintBar_FocusLabel_OmittedWhenSinglePane(t *testing.T) {
	m := defaultKeyHints().WithPaneOpen(false).WithFocus(FocusEntryList)
	v := m.View()
	if strings.Contains(v, "focus:") {
		t.Errorf("focus label must be omitted when only the list is visible, got %q", v)
	}
}

// T-092: label is right-aligned — the hint text appears before it.
func TestKeyHintBar_FocusLabel_RightAligned(t *testing.T) {
	m := defaultKeyHints().WithPaneOpen(true).WithFocus(FocusEntryList)
	v := m.View()
	labelIdx := strings.Index(v, "focus: list")
	if labelIdx < 0 {
		t.Fatalf("expected label in output, got %q", v)
	}
	// Some hint must precede the label (at minimum "j").
	if strings.Index(v, "j") > labelIdx {
		t.Errorf("label should be right of hints: label@%d, first 'j'@%d; view=%q",
			labelIdx, strings.Index(v, "j"), v)
	}
}
