package appshell

import (
	"strings"
	"testing"

	"github.com/istonikula/gloggy/internal/theme"
)

// F-201 pinning test — below-mode drag-seam ownership.
//
// The below-mode drag seam is NOT a "shared row" between the entry list and
// the detail pane; they are adjacent rows (list's bottom border + detail
// pane's top border) and only the detail pane's top is overridden to
// DragHandle via WithDragSeamTop. This test pins that contract: if a future
// change tries to "fix" the paint by also applying DragHandle to the list's
// bottom border (making it a genuine shared row), this assertion fails and
// forces the change to be evaluated against R10 AC 10 + R15 language.
//
// cavekit-app-shell.md R10 AC 10, R15 ACs at lines 200/206/207; DESIGN.md
// §2 / §4.5 / §6.
func TestPaneStyle_DragSeamOnlyOverridesDetailTop_NotListBottom(t *testing.T) {
	for _, themeName := range theme.BuiltinNames() {
		th := theme.GetTheme(themeName)
		t.Run(themeName, func(t *testing.T) {
			dragSGR := colorANSI(th.DragHandle)
			if dragSGR == "" {
				t.Fatalf("empty DragHandle SGR — is the color profile TrueColor?")
			}

			// Detail pane in below-mode: focus-state base style with
			// WithDragSeamTop applied — the top row must carry DragHandle SGR
			// because the detail pane's top border IS the drag seam.
			detail := WithDragSeamTop(PaneStyle(th, PaneStateFocused).Width(20), th).Render("detail")
			detailLines := strings.Split(detail, "\n")
			if len(detailLines) < 3 {
				t.Fatalf("detail pane render too short: %d lines", len(detailLines))
			}
			if !strings.Contains(detailLines[0], dragSGR) {
				t.Errorf("detail pane top border must carry DragHandle SGR %q; got %q", dragSGR, detailLines[0])
			}

			// List pane in below-mode sits ABOVE the detail pane. Its bottom
			// border is a separate row from the detail's top border, rendered
			// by its own PaneStyle — no WithDragSeamTop wrapper. This bottom
			// row must NOT carry DragHandle SGR. If it does, the drag seam
			// has become a "shared row" contrary to the R10/R15 contract and
			// the mouse-hit-zone contract in R6 would need to widen too.
			list := PaneStyle(th, PaneStateUnfocused).Width(20).Render("list")
			listLines := strings.Split(list, "\n")
			if len(listLines) < 3 {
				t.Fatalf("list pane render too short: %d lines", len(listLines))
			}
			listBottom := listLines[len(listLines)-1]
			if strings.Contains(listBottom, dragSGR) {
				t.Errorf("list pane bottom border must NOT carry DragHandle SGR (drag seam is the detail pane's top border alone, not a shared row); got %q", listBottom)
			}
		})
	}
}

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
