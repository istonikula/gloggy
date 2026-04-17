package appshell

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/istonikula/gloggy/internal/theme"
)

// T-089: each row of the divider is exactly 1 cell wide (lipgloss.Width).
func TestRenderDivider_RowWidthIsOneCell(t *testing.T) {
	th := theme.GetTheme("tokyo-night")
	out := RenderDivider(5, th)
	for i, line := range strings.Split(out, "\n") {
		if w := lipgloss.Width(line); w != 1 {
			t.Errorf("row %d width: got %d, want 1; line=%q", i, w, line)
		}
	}
}

// T-089: row count matches the requested height.
func TestRenderDivider_RowCountMatchesHeight(t *testing.T) {
	th := theme.GetTheme("tokyo-night")
	cases := []int{1, 3, 8, 24}
	for _, h := range cases {
		out := RenderDivider(h, th)
		got := len(strings.Split(out, "\n"))
		if got != h {
			t.Errorf("RenderDivider(%d): got %d rows, want %d", h, got, h)
		}
	}
}

// T-089: zero/negative heights produce an empty string.
func TestRenderDivider_ZeroHeightEmpty(t *testing.T) {
	th := theme.GetTheme("tokyo-night")
	if out := RenderDivider(0, th); out != "" {
		t.Errorf("RenderDivider(0): got %q, want empty", out)
	}
	if out := RenderDivider(-3, th); out != "" {
		t.Errorf("RenderDivider(-3): got %q, want empty", out)
	}
}

// T-089: glyph is the documented vertical-bar character.
func TestRenderDivider_GlyphIsVerticalBar(t *testing.T) {
	th := theme.GetTheme("tokyo-night")
	out := RenderDivider(1, th)
	if !strings.Contains(out, dividerGlyph) {
		t.Errorf("divider must include %q glyph, got %q", dividerGlyph, out)
	}
}
