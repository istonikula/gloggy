package appshell

import tea "github.com/charmbracelet/bubbletea"

// ResizeModel handles terminal resize events and propagates updated dimensions
// to the layout and pane models. It maintains the detail pane's proportional height.
type ResizeModel struct {
	width  int
	height int
}

// NewResizeModel creates a ResizeModel with initial dimensions.
func NewResizeModel(width, height int) ResizeModel {
	return ResizeModel{width: width, height: height}
}

// Width returns the current terminal width.
func (m ResizeModel) Width() int { return m.width }

// Height returns the current terminal height.
func (m ResizeModel) Height() int { return m.height }

// Update handles tea.WindowSizeMsg and updates the stored dimensions.
func (m ResizeModel) Update(msg tea.Msg) (ResizeModel, tea.Cmd) {
	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = ws.Width
		m.height = ws.Height
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
