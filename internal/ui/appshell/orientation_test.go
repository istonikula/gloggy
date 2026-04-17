package appshell

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/istonikula/gloggy/internal/config"
)

func resizeWindowMsg(w, h int) tea.Msg {
	return tea.WindowSizeMsg{Width: w, Height: h}
}

// T-087: explicit position overrides.
func TestSelectOrientation_ExplicitBelow(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.DetailPane.Position = "below"
	if got := SelectOrientation(1000, cfg); got != OrientationBelow {
		t.Errorf("explicit below: got %v, want OrientationBelow", got)
	}
}

func TestSelectOrientation_ExplicitRight(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.DetailPane.Position = "right"
	if got := SelectOrientation(50, cfg); got != OrientationRight {
		t.Errorf("explicit right: got %v, want OrientationRight", got)
	}
}

// T-087: auto mode threshold boundaries.
func TestSelectOrientation_Auto_Threshold(t *testing.T) {
	cfg := config.DefaultConfig() // Position=auto, threshold=100
	cases := []struct {
		width int
		want  Orientation
	}{
		{99, OrientationBelow},
		{100, OrientationRight},
		{120, OrientationRight},
		{80, OrientationBelow},
	}
	for _, tc := range cases {
		if got := SelectOrientation(tc.width, cfg); got != tc.want {
			t.Errorf("auto width=%d: got %v, want %v", tc.width, got, tc.want)
		}
	}
}

// T-087: custom threshold from config is honored.
func TestSelectOrientation_Auto_CustomThreshold(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.DetailPane.OrientationThresholdCols = 140
	if got := SelectOrientation(120, cfg); got != OrientationBelow {
		t.Errorf("width=120 threshold=140: got %v, want OrientationBelow", got)
	}
	if got := SelectOrientation(140, cfg); got != OrientationRight {
		t.Errorf("width=140 threshold=140: got %v, want OrientationRight", got)
	}
}

// T-087: invalid Position treated as auto.
func TestSelectOrientation_InvalidPosition_FallsBackToAuto(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.DetailPane.Position = "diagonal" // junk
	if got := SelectOrientation(120, cfg); got != OrientationRight {
		t.Errorf("invalid position at width 120: got %v, want OrientationRight", got)
	}
	if got := SelectOrientation(50, cfg); got != OrientationBelow {
		t.Errorf("invalid position at width 50: got %v, want OrientationBelow", got)
	}
}

// T-087: Orientation re-evaluated on every resize when ResizeModel hosts it.
func TestResizeModel_Orientation_ReevaluatedOnResize(t *testing.T) {
	cfg := config.DefaultConfig() // auto, threshold=100
	rm := NewResizeModel(120, 40).WithConfig(cfg)
	if rm.Orientation() != OrientationRight {
		t.Fatalf("initial orientation at 120: got %v, want OrientationRight", rm.Orientation())
	}
	rm, _ = rm.Update(resizeWindowMsg(90, 40))
	if rm.Orientation() != OrientationBelow {
		t.Errorf("after shrink to 90: got %v, want OrientationBelow", rm.Orientation())
	}
	rm, _ = rm.Update(resizeWindowMsg(110, 40))
	if rm.Orientation() != OrientationRight {
		t.Errorf("after grow to 110: got %v, want OrientationRight", rm.Orientation())
	}
}
