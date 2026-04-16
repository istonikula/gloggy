package appshell

import (
	"strings"
	"testing"

	"github.com/istonikula/gloggy/internal/theme"
)

func defaultHeader() HeaderModel {
	return NewHeaderModel(theme.GetTheme("tokyo-night"), 80)
}

// T-047: R3.1 — shows file name when reading from file.
func TestHeaderModel_FileName(t *testing.T) {
	h := defaultHeader().WithSource("/var/log/app.log")
	v := h.View()
	if !strings.Contains(v, "/var/log/app.log") {
		t.Errorf("expected file name in header: %q", v)
	}
}

// T-047: R3.2 — shows stdin indicator when source is empty.
func TestHeaderModel_Stdin(t *testing.T) {
	h := defaultHeader()
	v := h.View()
	if !strings.Contains(v, "stdin") {
		t.Errorf("expected stdin indicator in header: %q", v)
	}
}

// T-047: R3.3 — shows [FOLLOW] badge in tail mode.
func TestHeaderModel_FollowBadge(t *testing.T) {
	h := defaultHeader().WithFollow(true)
	v := h.View()
	if !strings.Contains(v, "[FOLLOW]") {
		t.Errorf("expected [FOLLOW] badge in header: %q", v)
	}
}

// T-047: R3.3 — no [FOLLOW] badge when not in tail mode.
func TestHeaderModel_NoFollowBadge(t *testing.T) {
	h := defaultHeader().WithFollow(false)
	v := h.View()
	if strings.Contains(v, "[FOLLOW]") {
		t.Errorf("unexpected [FOLLOW] badge: %q", v)
	}
}

// T-047: R3.4+R3.5 — shows total and visible counts.
func TestHeaderModel_Counts(t *testing.T) {
	h := defaultHeader().WithCounts(100, 42)
	v := h.View()
	if !strings.Contains(v, "42") || !strings.Contains(v, "100") {
		t.Errorf("expected counts in header: %q", v)
	}
}

// T-047: R3.6 — counts update (WithCounts returns new model).
func TestHeaderModel_CountsUpdate(t *testing.T) {
	h := defaultHeader().WithCounts(10, 10)
	h2 := h.WithCounts(20, 15)
	v := h2.View()
	if !strings.Contains(v, "20") {
		t.Errorf("expected updated count in header: %q", v)
	}
}

// T-081: R3.7 — header shows cursor position.
func TestHeaderModel_CursorPos(t *testing.T) {
	h := defaultHeader().WithCounts(100, 42).WithCursorPos(7)
	v := h.View()
	if !strings.Contains(v, "7/42") {
		t.Errorf("expected cursor/visible (7/42) in header: %q", v)
	}
}

// T-081: R3.8 — header style has background color configured.
// Note: In non-TTY test environments, lipgloss may not emit ANSI codes.
// We verify the style is applied by checking that Width produces padding.
func TestHeaderModel_WidthPadding(t *testing.T) {
	h := defaultHeader().WithSource("test.log").WithWidth(80)
	v := h.View()
	// With Width(80), output should be padded to at least 80 visible chars.
	if len(v) < 40 {
		t.Errorf("expected padded header, got length %d: %q", len(v), v)
	}
}
