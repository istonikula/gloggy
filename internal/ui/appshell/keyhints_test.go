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
