package appshell

import (
	"strings"
	"testing"
)

// T-046: R2.1 — header at top, entrylist in middle, status at bottom.
func TestLayout_Render_WithoutDetailPane(t *testing.T) {
	m := NewLayoutModel(80, 24)
	out := m.Render("HEADER", "ENTRYLIST", "DETAIL", "STATUS")
	lines := strings.Split(out, "\n")
	if strings.TrimSpace(lines[0]) != "HEADER" {
		t.Errorf("first line should be HEADER, got %q", lines[0])
	}
	if strings.TrimSpace(lines[1]) != "ENTRYLIST" {
		t.Errorf("second line should be ENTRYLIST, got %q", lines[1])
	}
	if strings.TrimSpace(lines[len(lines)-1]) != "STATUS" {
		t.Errorf("last line should be STATUS, got %q", lines[len(lines)-1])
	}
	// Detail pane should not appear when closed.
	if strings.Contains(out, "DETAIL") {
		t.Error("DETAIL should not appear when pane is closed")
	}
}

// T-046: R2.3 — detail pane appears between entry list and status bar when open.
func TestLayout_Render_WithDetailPane(t *testing.T) {
	m := NewLayoutModel(80, 24).SetDetailPane(true, 5)
	out := m.Render("HEADER", "ENTRYLIST", "DETAIL", "STATUS")
	lines := strings.Split(out, "\n")
	// Should be: HEADER, ENTRYLIST, DETAIL, STATUS
	if strings.TrimSpace(lines[0]) != "HEADER" {
		t.Errorf("line 0: want HEADER, got %q", lines[0])
	}
	if strings.TrimSpace(lines[1]) != "ENTRYLIST" {
		t.Errorf("line 1: want ENTRYLIST, got %q", lines[1])
	}
	if strings.TrimSpace(lines[2]) != "DETAIL" {
		t.Errorf("line 2: want DETAIL, got %q", lines[2])
	}
	if strings.TrimSpace(lines[3]) != "STATUS" {
		t.Errorf("line 3: want STATUS, got %q", lines[3])
	}
}

// T-046: R2.5 — entry list height fills available space.
func TestLayout_EntryListHeight_FillsSpace(t *testing.T) {
	l := NewLayout(80, 24, false, 0)
	// 24 rows - 1 header - 1 status = 22
	if l.EntryListHeight() != 22 {
		t.Errorf("entry list height: got %d, want 22", l.EntryListHeight())
	}
}

// T-046: R2.3 — entry list height reduced by detail pane when open.
func TestLayout_EntryListHeight_ReducedByDetailPane(t *testing.T) {
	l := NewLayout(80, 24, true, 8)
	// 24 - 1 - 1 - 8 = 14
	if l.EntryListHeight() != 14 {
		t.Errorf("entry list height with detail pane: got %d, want 14", l.EntryListHeight())
	}
}

// SetSize updates dimensions.
func TestLayoutModel_SetSize(t *testing.T) {
	m := NewLayoutModel(80, 24).SetSize(120, 40)
	l := m.Layout()
	if l.Width != 120 || l.Height != 40 {
		t.Errorf("SetSize: got %dx%d, want 120x40", l.Width, l.Height)
	}
}

// T-090: terminal-too-small fallback at 59x15.
func TestLayout_FallbackBelowMinWidth(t *testing.T) {
	m := NewLayoutModel(59, 15)
	out := m.Render("HEADER", "ENTRYLIST", "DETAIL", "STATUS")
	if !strings.Contains(out, "terminal too small") {
		t.Errorf("expected fallback message at 59x15, got %q", out)
	}
	if strings.Contains(out, "HEADER") || strings.Contains(out, "ENTRYLIST") {
		t.Errorf("panels should be suppressed at 59x15, got %q", out)
	}
}

// T-090: terminal-too-small fallback at 60x14.
func TestLayout_FallbackBelowMinHeight(t *testing.T) {
	m := NewLayoutModel(60, 14)
	out := m.Render("HEADER", "ENTRYLIST", "DETAIL", "STATUS")
	if !strings.Contains(out, "terminal too small") {
		t.Errorf("expected fallback message at 60x14, got %q", out)
	}
	if strings.Contains(out, "HEADER") {
		t.Errorf("panels should be suppressed at 60x14, got %q", out)
	}
}

// T-090: normal render resumes at 60x15.
func TestLayout_NormalRenderAtMinFloor(t *testing.T) {
	m := NewLayoutModel(60, 15)
	out := m.Render("HEADER", "ENTRYLIST", "DETAIL", "STATUS")
	if strings.Contains(out, "terminal too small") {
		t.Errorf("normal render should resume at 60x15, got %q", out)
	}
	if !strings.Contains(out, "HEADER") {
		t.Errorf("HEADER should appear at 60x15, got %q", out)
	}
}

// T-090: IsBelowMinFloor predicate.
func TestLayout_IsBelowMinFloor(t *testing.T) {
	cases := []struct {
		w, h int
		want bool
	}{
		{60, 15, false},
		{60, 14, true},
		{59, 15, true},
		{120, 40, false},
		{0, 0, true},
	}
	for _, tc := range cases {
		m := NewLayoutModel(tc.w, tc.h)
		if got := m.IsBelowMinFloor(); got != tc.want {
			t.Errorf("IsBelowMinFloor(%dx%d) = %v, want %v", tc.w, tc.h, got, tc.want)
		}
	}
}
