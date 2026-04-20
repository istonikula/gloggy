package appshell

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// forceTrueColor pins lipgloss's renderer to TrueColor for this test file's
// duration so SGR assertions do not depend on the host terminal's detected
// color profile (e.g. the `go test` CI harness often reports Ascii).
func forceTrueColor(t *testing.T) {
	t.Helper()
	prev := lipgloss.DefaultRenderer().ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(prev) })
}

func TestBgSGROpen_TrueColor(t *testing.T) {
	forceTrueColor(t)
	open := BgSGROpen(lipgloss.Color("#1a1b26"))
	// #1a1b26 = (26, 27, 38). TrueColor bg SGR is `48;2;R;G;B`.
	if !strings.Contains(open, "48;2;26;27;38") {
		t.Fatalf("BgSGROpen returned %q, want 48;2;26;27;38 fragment", open)
	}
	if !strings.HasPrefix(open, "\x1b[") {
		t.Fatalf("BgSGROpen %q: missing CSI prefix", open)
	}
	if !strings.HasSuffix(open, "m") {
		t.Fatalf("BgSGROpen %q: missing SGR terminator", open)
	}
	if strings.Contains(open, "\x1b[0m") {
		t.Fatalf("BgSGROpen %q: must not include reset", open)
	}
}

func TestBgSGROpen_EmptyColorReturnsEmpty(t *testing.T) {
	forceTrueColor(t)
	if got := BgSGROpen(lipgloss.Color("")); got != "" {
		t.Fatalf("BgSGROpen(\"\") = %q, want empty", got)
	}
}

func TestRepaintBg_RewritesEveryReset(t *testing.T) {
	forceTrueColor(t)
	bg := lipgloss.Color("#1a1b26")
	// Simulate a body with three foreground-only styled tokens separated by
	// plain text (the exact shape `RenderCompactRow` and `detailpane.renderValue`
	// produce today).
	fg := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))
	body := fg.Render("time") + ": " + fg.Render("value") + "," + fg.Render("x")

	out := RepaintBg(body, bg)

	open := BgSGROpen(bg)
	// Every `\x1b[0m` in the original body must be followed by the bg-set SGR.
	if c := strings.Count(out, "\x1b[0m"+open); c < 3 {
		t.Fatalf("RepaintBg: expected >= 3 reset+reassert pairs, got %d\nout=%q", c, out)
	}
	// No bare reset must be left un-repainted.
	scanned := out
	for {
		i := strings.Index(scanned, "\x1b[0m")
		if i < 0 {
			break
		}
		tail := scanned[i+len("\x1b[0m"):]
		if !strings.HasPrefix(tail, open) {
			t.Fatalf("RepaintBg: bare reset at byte %d not followed by bg open\nout=%q", i, out)
		}
		scanned = tail
	}
}

func TestRepaintBg_NoOpWhenNoBg(t *testing.T) {
	body := "\x1b[38;2;255;0;0mred\x1b[0m rest"
	if got := RepaintBg(body, lipgloss.Color("")); got != body {
		t.Fatalf("RepaintBg with empty bg must be a no-op; got %q", got)
	}
}
