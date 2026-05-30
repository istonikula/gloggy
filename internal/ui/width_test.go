package ui

import (
	"testing"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	lipgloss.SetColorProfile(termenv.TrueColor)
}

// V36: SafeTruncate keeps output cell-width <= cells and never cuts a rune
// mid-byte, for ASCII, CJK (2-cell) and emoji (2-cell) input.
func TestSafeTruncate_CellWidthAndRuneSafety(t *testing.T) {
	cases := []struct {
		name  string
		in    string
		cells int
		want  string
	}{
		{"ascii fits", "hello", 10, "hello"},
		{"ascii exact", "hello", 5, "hello"},
		{"ascii cut", "hello world", 5, "hello"},
		{"zero cells", "hello", 0, ""},
		{"negative cells", "hello", -3, ""},
		// CJK: each ideograph is 2 cells.
		{"cjk fits", "日本語", 6, "日本語"},
		{"cjk cut to 2", "日本語", 3, "日"},   // 日=2, 本 would push to 4 > 3
		{"cjk cut at boundary", "日本語", 4, "日本"}, // 4 cells exactly two ideographs
		{"cjk one cell budget", "日本", 1, ""},    // can't fit a 2-cell rune in 1
		// Emoji (2 cells) mixed with ascii.
		{"emoji mixed", "ab🎉cd", 4, "ab🎉"}, // a+b+🎉 = 4
		{"emoji over budget", "🎉🎉", 3, "🎉"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := SafeTruncate(tc.in, tc.cells)
			assert.Equal(t, tc.want, got)
			require.True(t, utf8.ValidString(got), "output must be valid UTF-8")
			if tc.cells > 0 {
				assert.LessOrEqual(t, lipgloss.Width(got), tc.cells,
					"rendered cell width must not exceed budget")
			}
		})
	}
}
