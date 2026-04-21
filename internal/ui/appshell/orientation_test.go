package appshell

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/istonikula/gloggy/internal/config"
)

func resizeWindowMsg(w, h int) tea.Msg {
	return tea.WindowSizeMsg{Width: w, Height: h}
}

// T-087: explicit position overrides.
func TestSelectOrientation_ExplicitBelow(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.DetailPane.Position = "below"
	assert.Equalf(t, OrientationBelow, SelectOrientation(1000, cfg), "explicit below")
}

func TestSelectOrientation_ExplicitRight(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.DetailPane.Position = "right"
	assert.Equalf(t, OrientationRight, SelectOrientation(50, cfg), "explicit right")
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
		assert.Equalf(t, tc.want, SelectOrientation(tc.width, cfg), "auto width=%d", tc.width)
	}
}

// T-087: custom threshold from config is honored.
func TestSelectOrientation_Auto_CustomThreshold(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.DetailPane.OrientationThresholdCols = 140
	assert.Equalf(t, OrientationBelow, SelectOrientation(120, cfg), "width=120 threshold=140")
	assert.Equalf(t, OrientationRight, SelectOrientation(140, cfg), "width=140 threshold=140")
}

// T-087: invalid Position treated as auto.
func TestSelectOrientation_InvalidPosition_FallsBackToAuto(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.DetailPane.Position = "diagonal" // junk
	assert.Equalf(t, OrientationRight, SelectOrientation(120, cfg), "invalid position at width 120")
	assert.Equalf(t, OrientationBelow, SelectOrientation(50, cfg), "invalid position at width 50")
}

// T-087: Orientation re-evaluated on every resize when ResizeModel hosts it.
func TestResizeModel_Orientation_ReevaluatedOnResize(t *testing.T) {
	cfg := config.DefaultConfig() // auto, threshold=100
	rm := NewResizeModel(120, 40).WithConfig(cfg)
	require.Equalf(t, OrientationRight, rm.Orientation(), "initial orientation at 120")
	rm, _ = rm.Update(resizeWindowMsg(90, 40))
	assert.Equalf(t, OrientationBelow, rm.Orientation(), "after shrink to 90")
	rm, _ = rm.Update(resizeWindowMsg(110, 40))
	assert.Equalf(t, OrientationRight, rm.Orientation(), "after grow to 110")
}
