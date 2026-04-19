package appshell

import (
	"strings"
	"testing"

	"github.com/istonikula/gloggy/internal/theme"
)

// T-173: WithDragSeamTop overrides the top border SGR with DragHandle while
// leaving the other three borders in the focus-state colour. Asserts across
// all bundled themes × both focus states.
func TestPaneStyle_WithDragSeamTop_TopRowUsesDragHandle(t *testing.T) {
	cases := []struct {
		name  string
		state PaneVisualState
	}{
		{"focused", PaneStateFocused},
		{"unfocused", PaneStateUnfocused},
	}
	for _, themeName := range theme.BuiltinNames() {
		th := theme.GetTheme(themeName)
		for _, tc := range cases {
			t.Run(themeName+"/"+tc.name, func(t *testing.T) {
				base := PaneStyle(th, tc.state).Width(6)
				seam := WithDragSeamTop(base, th).Render("hello")
				ref := base.Render("hello")

				lines := strings.Split(seam, "\n")
				refLines := strings.Split(ref, "\n")
				if len(lines) < 3 || len(refLines) < 3 {
					t.Fatalf("expected >=3 lines (top border, body, bottom border); got seam=%d ref=%d", len(lines), len(refLines))
				}
				topSGR := colorANSI(th.DragHandle)
				if topSGR == "" {
					t.Fatalf("empty DragHandle SGR — is the color profile TrueColor?")
				}
				if !strings.Contains(lines[0], topSGR) {
					t.Errorf("top border missing DragHandle SGR %q; got %q", topSGR, lines[0])
				}
				// Bottom row must NOT use DragHandle — the override is
				// strictly scoped to the top edge.
				bottom := lines[len(lines)-1]
				if strings.Contains(bottom, topSGR) {
					t.Errorf("bottom border must NOT use DragHandle; got %q", bottom)
				}
				// Middle + bottom rows must be byte-identical to the base
				// (un-overridden) style so left/right/bottom borders keep
				// their focus-state colour.
				for i := 1; i < len(lines) && i < len(refLines); i++ {
					if lines[i] != refLines[i] {
						t.Errorf("row %d differs between base and drag-seam styles; override leaked off the top edge\nbase: %q\nseam: %q", i, refLines[i], lines[i])
					}
				}
			})
		}
	}
}
