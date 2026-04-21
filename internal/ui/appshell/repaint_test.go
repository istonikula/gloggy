package appshell

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.Containsf(t, open, "48;2;26;27;38", "BgSGROpen returned %q, want 48;2;26;27;38 fragment", open)
	require.Truef(t, strings.HasPrefix(open, "\x1b["), "BgSGROpen %q: missing CSI prefix", open)
	require.Truef(t, strings.HasSuffix(open, "m"), "BgSGROpen %q: missing SGR terminator", open)
	require.NotContainsf(t, open, "\x1b[0m", "BgSGROpen %q: must not include reset", open)
}

func TestBgSGROpen_EmptyColorReturnsEmpty(t *testing.T) {
	forceTrueColor(t)
	got := BgSGROpen(lipgloss.Color(""))
	require.Emptyf(t, got, "BgSGROpen(\"\") = %q, want empty", got)
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
	c := strings.Count(out, "\x1b[0m"+open)
	require.GreaterOrEqualf(t, c, 3, "RepaintBg: expected >= 3 reset+reassert pairs, got %d\nout=%q", c, out)
	// No bare reset must be left un-repainted.
	scanned := out
	for {
		i := strings.Index(scanned, "\x1b[0m")
		if i < 0 {
			break
		}
		tail := scanned[i+len("\x1b[0m"):]
		require.Truef(t, strings.HasPrefix(tail, open),
			"RepaintBg: bare reset at byte %d not followed by bg open\nout=%q", i, out)
		scanned = tail
	}
}

func TestRepaintBg_NoOpWhenNoBg(t *testing.T) {
	body := "\x1b[38;2;255;0;0mred\x1b[0m rest"
	got := RepaintBg(body, lipgloss.Color(""))
	assert.Equalf(t, body, got, "RepaintBg with empty bg must be a no-op; got %q", got)
}
