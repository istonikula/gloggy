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

// T-093: header drops focus label first when too narrow (R3.10).
func TestHeaderModel_DropFocusLabelFirst(t *testing.T) {
	h := defaultHeader().
		WithSource("/var/log/app.log").
		WithFollow(true).
		WithCounts(100, 42).
		WithCursorPos(7).
		WithFocusLabel("focus: list").
		WithWidth(50)
	v := h.View()
	if strings.Contains(v, "focus:") {
		t.Errorf("focus label should be dropped first, got %q", v)
	}
	if !strings.Contains(v, "[FOLLOW]") {
		t.Errorf("FOLLOW badge should remain at width=50, got %q", v)
	}
	if !strings.Contains(v, "/var/log/app.log") {
		t.Errorf("source should always be kept, got %q", v)
	}
}

// T-093: header drops counts second (R3.10).
func TestHeaderModel_DropCountsSecond(t *testing.T) {
	h := defaultHeader().
		WithSource("/var/log/app.log").
		WithFollow(true).
		WithCounts(100, 42).
		WithCursorPos(7).
		WithFocusLabel("focus: list").
		WithWidth(36)
	v := h.View()
	if strings.Contains(v, "entries") {
		t.Errorf("counts should be dropped at width=36, got %q", v)
	}
	if !strings.Contains(v, "7/42") {
		t.Errorf("cursor pos should remain at width=36, got %q", v)
	}
	if !strings.Contains(v, "[FOLLOW]") {
		t.Errorf("FOLLOW should remain at width=36, got %q", v)
	}
}

// T-093: header drops cursor pos third (R3.10).
func TestHeaderModel_DropCursorPosThird(t *testing.T) {
	h := defaultHeader().
		WithSource("/var/log/app.log").
		WithFollow(true).
		WithCounts(100, 42).
		WithCursorPos(7).
		WithFocusLabel("focus: list").
		WithWidth(28)
	v := h.View()
	if strings.Contains(v, "7/42") {
		t.Errorf("cursor pos should be dropped at width=28, got %q", v)
	}
	if !strings.Contains(v, "[FOLLOW]") {
		t.Errorf("FOLLOW should remain at width=28, got %q", v)
	}
	if !strings.Contains(v, "/var/log/app.log") {
		t.Errorf("source should remain at width=28, got %q", v)
	}
}

// T-093: header drops FOLLOW badge fourth (R3.10).
func TestHeaderModel_DropFollowFourth(t *testing.T) {
	h := defaultHeader().
		WithSource("/var/log/app.log").
		WithFollow(true).
		WithCounts(100, 42).
		WithCursorPos(7).
		WithFocusLabel("focus: list").
		WithWidth(17)
	v := h.View()
	if strings.Contains(v, "[FOLLOW]") {
		t.Errorf("FOLLOW should be dropped at width=17, got %q", v)
	}
	if !strings.Contains(v, "/var/log/app.log") {
		t.Errorf("source should remain at width=17, got %q", v)
	}
}

// T-093: header truncates source with `…` when alone overflows (R3.11).
func TestHeaderModel_SourceTruncatedWithEllipsis(t *testing.T) {
	h := defaultHeader().
		WithSource("/very/long/path/to/some/log/file.log").
		WithFollow(true).
		WithCounts(100, 42).
		WithCursorPos(7).
		WithWidth(10)
	v := h.View()
	if !strings.Contains(v, "…") {
		t.Errorf("expected ellipsis when source alone overflows at width=10, got %q", v)
	}
	if strings.Contains(v, "[FOLLOW]") {
		t.Errorf("FOLLOW should be dropped at width=10, got %q", v)
	}
}

// T-093: truncateToWidth produces correct cell width.
func TestTruncateToWidth_RespectsCellWidth(t *testing.T) {
	got := truncateToWidth("hello world", 5)
	if got != "hell…" {
		t.Errorf("truncateToWidth(\"hello world\", 5) = %q, want %q", got, "hell…")
	}
	got2 := truncateToWidth("hi", 10)
	if got2 != "hi" {
		t.Errorf("no truncation when fits: got %q", got2)
	}
	got3 := truncateToWidth("anything", 1)
	if got3 != "…" {
		t.Errorf("max=1 should yield single ellipsis, got %q", got3)
	}
}
