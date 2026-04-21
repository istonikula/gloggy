package appshell

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/istonikula/gloggy/internal/theme"
)

// Force TrueColor so SGR codes are embedded in rendered output —
// required for T-172 + T-174 drag-seam SGR assertions.
func init() {
	lipgloss.SetColorProfile(termenv.TrueColor)
}

// colorANSI renders a probe with color c and returns the SGR prefix
// lipgloss emits for it. Mirrors the helper in entrylist/detailpane
// tests — keeps assertions independent of termenv rounding.
func colorANSI(c lipgloss.Color) string {
	rendered := lipgloss.NewStyle().Foreground(c).Render("x")
	end := strings.Index(rendered, "x")
	if end <= 0 {
		return ""
	}
	return rendered[:end]
}

// T-089: each row of the divider is exactly 1 cell wide (lipgloss.Width).
func TestRenderDivider_RowWidthIsOneCell(t *testing.T) {
	th := theme.GetTheme("tokyo-night")
	out := RenderDivider(5, th)
	for i, line := range strings.Split(out, "\n") {
		assert.Equalf(t, 1, lipgloss.Width(line), "row %d width: got %d, want 1; line=%q", i, lipgloss.Width(line), line)
	}
}

// T-089: row count matches the requested height.
func TestRenderDivider_RowCountMatchesHeight(t *testing.T) {
	th := theme.GetTheme("tokyo-night")
	cases := []int{1, 3, 8, 24}
	for _, h := range cases {
		out := RenderDivider(h, th)
		got := len(strings.Split(out, "\n"))
		assert.Equalf(t, h, got, "RenderDivider(%d): got %d rows, want %d", h, got, h)
	}
}

// T-089: zero/negative heights produce an empty string.
func TestRenderDivider_ZeroHeightEmpty(t *testing.T) {
	th := theme.GetTheme("tokyo-night")
	assert.Emptyf(t, RenderDivider(0, th), "RenderDivider(0): want empty")
	assert.Emptyf(t, RenderDivider(-3, th), "RenderDivider(-3): want empty")
}

// T-089: glyph is the documented vertical-bar character.
func TestRenderDivider_GlyphIsVerticalBar(t *testing.T) {
	th := theme.GetTheme("tokyo-night")
	out := RenderDivider(1, th)
	assert.Containsf(t, out, dividerGlyph, "divider must include %q glyph, got %q", dividerGlyph, out)
}

// T-172: glyph foreground is DragHandle (not DividerColor). Focus-neutral
// drag-seam colouring per Tier 23 kit revision (config R4 AC 9, app-shell
// R10 AC 10, R15 AC 15). Asserts across all bundled themes.
func TestRenderDivider_GlyphUsesDragHandle_AllThemes(t *testing.T) {
	for _, name := range theme.BuiltinNames() {
		t.Run(name, func(t *testing.T) {
			th := theme.GetTheme(name)
			out := RenderDivider(1, th)
			wantSGR := colorANSI(th.DragHandle)
			divSGR := colorANSI(th.DividerColor)
			require.NotEmptyf(t, wantSGR, "empty DragHandle probe SGR (profile not TrueColor?)")
			require.NotEmptyf(t, divSGR, "empty DividerColor probe SGR (profile not TrueColor?)")
			assert.Containsf(t, out, wantSGR, "divider missing DragHandle SGR %q; got %q", wantSGR, out)
			assert.NotContainsf(t, out, divSGR, "divider still paints DividerColor SGR %q; got %q", divSGR, out)
		})
	}
}
