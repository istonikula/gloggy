package detailpane

import tea "github.com/charmbracelet/bubbletea"

const (
	minHeightRatio = 0.10
	// maxHeightRatio matches DESIGN.md §5 ratio clamp [0.10, 0.80] (T-098).
	maxHeightRatio = 0.80
	heightStep     = 0.05
)

// HeightModel tracks the detail pane height as a ratio of terminal height.
// It handles +/- key adjustment and terminal resize.
type HeightModel struct {
	ratio         float64 // current ratio (0.0–1.0)
	terminalHeight int
}

// NewHeightModel creates a HeightModel with the given initial ratio.
// ratio is clamped to [minHeightRatio, maxHeightRatio].
func NewHeightModel(ratio float64, terminalHeight int) HeightModel {
	return HeightModel{
		ratio:         clampRatio(ratio),
		terminalHeight: terminalHeight,
	}
}

// PaneHeight returns the current pane height in rows.
func (m HeightModel) PaneHeight() int {
	h := int(float64(m.terminalHeight) * m.ratio)
	if h < 1 {
		h = 1
	}
	return h
}

// Ratio returns the current height ratio.
func (m HeightModel) Ratio() float64 { return m.ratio }

// Increase adds one step to the ratio.
func (m HeightModel) Increase() HeightModel {
	m.ratio = clampRatio(m.ratio + heightStep)
	return m
}

// Decrease subtracts one step from the ratio.
func (m HeightModel) Decrease() HeightModel {
	m.ratio = clampRatio(m.ratio - heightStep)
	return m
}

// SetTerminalHeight updates the terminal height (on window resize).
// The ratio is preserved, so the pane height scales proportionally.
func (m HeightModel) SetTerminalHeight(h int) HeightModel {
	m.terminalHeight = h
	return m
}

// SetRatio sets the ratio directly (used by the unified ratio keymap in
// appshell.NextRatio). Value is clamped to [minHeightRatio, maxHeightRatio].
func (m HeightModel) SetRatio(r float64) HeightModel {
	m.ratio = clampRatio(r)
	return m
}

// SetRatioFromDrag updates the ratio based on a drag event. newPaneHeight is
// the desired height in rows; it is converted back to a ratio.
func (m HeightModel) SetRatioFromDrag(newPaneHeight int) HeightModel {
	if m.terminalHeight <= 0 {
		return m
	}
	m.ratio = clampRatio(float64(newPaneHeight) / float64(m.terminalHeight))
	return m
}

// Update handles +/- keys and tea.WindowSizeMsg.
func (m HeightModel) Update(msg tea.Msg) (HeightModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "+":
			m = m.Increase()
		case "-":
			m = m.Decrease()
		}
	case tea.WindowSizeMsg:
		m = m.SetTerminalHeight(msg.Height)
	}
	return m, nil
}

func clampRatio(r float64) float64 {
	if r < minHeightRatio {
		return minHeightRatio
	}
	if r > maxHeightRatio {
		return maxHeightRatio
	}
	return r
}
