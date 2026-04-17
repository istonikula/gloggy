package appshell

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/istonikula/gloggy/internal/config"
)

// ResizeModel handles terminal resize events and propagates updated dimensions
// to the layout and pane models. It maintains the detail pane's proportional
// height and re-evaluates the pane orientation (below vs right-split) on every
// tea.WindowSizeMsg using the config-driven rules in SelectOrientation (T-087).
type ResizeModel struct {
	width       int
	height      int
	cfg         config.Config
	hasCfg      bool
	orientation Orientation
}

// NewResizeModel creates a ResizeModel with initial dimensions. Orientation
// defaults to below until a config is attached via WithConfig.
func NewResizeModel(width, height int) ResizeModel {
	return ResizeModel{width: width, height: height}
}

// WithConfig attaches a config and computes the initial orientation. The
// config is retained so every subsequent WindowSizeMsg re-evaluates the
// orientation against the updated width.
func (m ResizeModel) WithConfig(cfg config.Config) ResizeModel {
	m.cfg = cfg
	m.hasCfg = true
	m.orientation = SelectOrientation(m.width, cfg)
	return m
}

// Width returns the current terminal width.
func (m ResizeModel) Width() int { return m.width }

// Height returns the current terminal height.
func (m ResizeModel) Height() int { return m.height }

// Orientation returns the currently resolved orientation. When no config is
// attached, returns OrientationBelow.
func (m ResizeModel) Orientation() Orientation { return m.orientation }

// Update handles tea.WindowSizeMsg and updates the stored dimensions. When a
// config is attached, the orientation is re-evaluated on every resize.
func (m ResizeModel) Update(msg tea.Msg) (ResizeModel, tea.Cmd) {
	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = ws.Width
		m.height = ws.Height
		if m.hasCfg {
			m.orientation = SelectOrientation(m.width, m.cfg)
		}
	}
	return m, nil
}

// ApplyToLayout returns a new Layout that reflects the current terminal dimensions
// while preserving the detail pane height ratio.
// ratio is the fractional height of the detail pane (e.g. 0.30).
// detailOpen controls whether the pane is counted in the layout.
func ApplyToLayout(rm ResizeModel, ratio float64, detailOpen bool) Layout {
	paneH := 0
	if detailOpen {
		paneH = int(float64(rm.Height()) * ratio)
		if paneH < 1 {
			paneH = 1
		}
	}
	return NewLayout(rm.Width(), rm.Height(), detailOpen, paneH)
}
