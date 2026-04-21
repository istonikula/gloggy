package app

import (
	"os"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/istonikula/gloggy/internal/config"
	"github.com/istonikula/gloggy/internal/ui/appshell"
)

// ---------- T-104: mouse drag on divider resizes detail pane width ----------

// TestModel_DividerDrag_UpdatesWidthRatio verifies a Press on the divider
// followed by a Motion to a new column translates to a width_ratio update.
func TestModel_DividerDrag_UpdatesWidthRatio(t *testing.T) {
	m := newModel()
	m = resize(m, 200, 24)
	entries := makeEntries(5)
	m = m.SetEntries(entries)
	m = m.openPane(entries[0])
	require.Equalf(t, appshell.OrientationRight, m.resize.Orientation(),
		"precondition: want right orientation, got %v", m.resize.Orientation())
	before := m.cfg.Config.DetailPane.WidthRatio

	// Press on the divider column to start the drag session.
	// Post-T-160 the visible `│` column equals ListContentWidth().
	l := m.layout.Layout()
	divider := l.ListContentWidth()
	m = send(m, tea.MouseMsg{X: divider, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	require.Truef(t, m.draggingDivider,
		"precondition: expected draggingDivider=true after Press on divider column %d", divider)

	// Motion to a column 20 cells to the left → detail grows → ratio rises.
	m = send(m, tea.MouseMsg{X: divider - 20, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionMotion})
	after := m.cfg.Config.DetailPane.WidthRatio
	assert.Greaterf(t, after, before,
		"drag-left should increase width_ratio: before=%.3f after=%.3f", before, after)

	// Release ends the drag session.
	m = send(m, tea.MouseMsg{X: divider - 20, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionRelease})
	assert.False(t, m.draggingDivider, "expected draggingDivider=false after Release")
}

// ---------- T-099: ratio live write-back to config ----------

// TestModel_DividerDragRelease_PersistsWidthRatio verifies that releasing
// a divider drag flushes the new width_ratio to disk.
func TestModel_DividerDragRelease_PersistsWidthRatio(t *testing.T) {
	dir := t.TempDir()
	cfgPath := dir + "/config.toml"

	cfg := config.LoadResult{Config: config.DefaultConfig()}
	m := New("", false, cfgPath, cfg)
	m = resize(m, 200, 24)
	m = m.SetEntries(makeEntries(3))
	m = m.openPane(makeEntries(3)[0])
	require.Equalf(t, appshell.OrientationRight, m.resize.Orientation(),
		"precondition: want right orientation, got %v", m.resize.Orientation())

	l := m.layout.Layout()
	divider := l.ListContentWidth() // post-T-160 visible-glyph column

	m = send(m, tea.MouseMsg{X: divider, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	require.Truef(t, m.draggingDivider, "precondition: drag did not start at divider x=%d", divider)
	m = send(m, tea.MouseMsg{X: divider - 30, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionMotion})
	m = send(m, tea.MouseMsg{X: divider - 30, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionRelease})

	reloaded := config.Load(cfgPath)
	assert.Equalf(t, m.cfg.Config.DetailPane.WidthRatio, reloaded.Config.DetailPane.WidthRatio,
		"disk width_ratio after drag release: got %.3f, want %.3f",
		reloaded.Config.DetailPane.WidthRatio, m.cfg.Config.DetailPane.WidthRatio)
}

// ---------- T-156: mouse drag resize (cavekit-app-shell R15) ----------

// belowDividerY computes the divider row for a below-mode model —
// entryListEnd+1, where entryListEnd = layout.EntryListHeight().
func belowDividerY(m Model) int {
	return m.layout.Layout().EntryListHeight() + 1
}

// rightDividerX computes the divider column for a right-split model.
// Mirrors MouseRouter.Zone post-T-160: divider = ListContentWidth() (the
// visible `│` glyph column, verified by renderer-truth test in appshell/).
func rightDividerX(m Model) int {
	return m.layout.Layout().ListContentWidth()
}

// TestModel_T156_BelowDrag_PressOnDivider_StartsDrag confirms that a left
// Press on the divider row in below-orientation flips `draggingDivider`.
func TestModel_T156_BelowDrag_PressOnDivider_StartsDrag(t *testing.T) {
	m := setupRatioModelBelow(t, false)
	dy := belowDividerY(m)
	m = send(m, tea.MouseMsg{X: 20, Y: dy, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	assert.Truef(t, m.draggingDivider, "Press on below-mode divider row y=%d must start drag", dy)
}

// TestModel_T156_BelowDrag_Motion covers the below-mode drag motion
// matrix: grow on up, shrink on down, pin at RatioMax/RatioMin for
// out-of-bounds motion. `targetY(dy, termH)` returns the absolute Y the
// Motion message should aim at — dy is the initial divider row.
func TestModel_T156_BelowDrag_Motion(t *testing.T) {
	greater := func(t *testing.T, b, a float64) {
		t.Helper()
		assert.Greaterf(t, a, b, "height_ratio must grow: before=%.3f after=%.3f", b, a)
	}
	less := func(t *testing.T, b, a float64) {
		t.Helper()
		assert.Lessf(t, a, b, "height_ratio must shrink: before=%.3f after=%.3f", b, a)
	}
	eq := func(want float64) func(t *testing.T, _, after float64) {
		return func(t *testing.T, _, a float64) {
			t.Helper()
			assert.Equalf(t, want, a, "height_ratio: got %.3f, want %.3f", a, want)
		}
	}
	cases := []struct {
		name      string
		seedRatio float64 // 0 = leave default
		targetY   func(dy, termH int) int
		assertion func(t *testing.T, before, after float64)
	}{
		{"motionUp_growsDetail", 0, func(dy, _ int) int { return dy - 4 }, greater},
		{"motionDown_from_0.50_shrinksDetail", 0.50, func(dy, _ int) int { return dy + 3 }, less},
		{"extremeUp_clampsAtMax", 0, func(_, _ int) int { return -100 }, eq(appshell.RatioMax)},
		{"extremeDown_clampsAtMin", 0, func(_, termH int) int { return termH + 100 }, eq(appshell.RatioMin)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := setupRatioModelBelow(t, false)
			if tc.seedRatio != 0 {
				m.cfg.Config.DetailPane.HeightRatio = tc.seedRatio
				m.paneHeight = m.paneHeight.SetRatio(tc.seedRatio)
				m = m.relayout()
			}
			dy := belowDividerY(m)
			termH := m.resize.Height()
			before := m.cfg.Config.DetailPane.HeightRatio

			m = send(m, tea.MouseMsg{X: 20, Y: dy, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
			m = send(m, tea.MouseMsg{X: 20, Y: tc.targetY(dy, termH), Button: tea.MouseButtonLeft, Action: tea.MouseActionMotion})

			tc.assertion(t, before, m.cfg.Config.DetailPane.HeightRatio)
		})
	}
}

// TestModel_T156_BelowDrag_Release_PersistsHeightRatio ensures the final
// ratio is flushed to disk exactly once on mouse release, and that the
// in-memory value matches.
func TestModel_T156_BelowDrag_Release_PersistsHeightRatio(t *testing.T) {
	dir := t.TempDir()
	cfgPath := dir + "/config.toml"
	cfg := config.LoadResult{Config: config.DefaultConfig()}
	m := New("", false, cfgPath, cfg)
	m = resize(m, 80, 24)
	entries := makeEntries(3)
	m = m.SetEntries(entries)
	m = m.openPane(entries[0])
	require.Equalf(t, appshell.OrientationBelow, m.resize.Orientation(),
		"precondition: want below, got %v", m.resize.Orientation())
	dy := belowDividerY(m)

	m = send(m, tea.MouseMsg{X: 20, Y: dy, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	// Ensure config file does NOT exist yet — motion must not save.
	m = send(m, tea.MouseMsg{X: 20, Y: dy - 3, Button: tea.MouseButtonLeft, Action: tea.MouseActionMotion})
	_, err := os.Stat(cfgPath)
	assert.Error(t, err, "motion during drag must NOT write config; file exists already")
	m = send(m, tea.MouseMsg{X: 20, Y: dy - 3, Button: tea.MouseButtonLeft, Action: tea.MouseActionRelease})
	assert.False(t, m.draggingDivider, "Release must end drag session")
	reloaded := config.Load(cfgPath)
	assert.Equalf(t, m.cfg.Config.DetailPane.HeightRatio, reloaded.Config.DetailPane.HeightRatio,
		"disk height_ratio after release: got %.3f, want %.3f",
		reloaded.Config.DetailPane.HeightRatio, m.cfg.Config.DetailPane.HeightRatio)
}

// TestModel_T156_Drag_IsFocusNeutral ensures dragging the divider never
// mutates m.focus — in either orientation, starting from either focus.
func TestModel_T156_Drag_IsFocusNeutral(t *testing.T) {
	cases := []struct {
		name      string
		setup     func(t *testing.T, listFocus bool) Model
		dividerFn func(Model) (x, y int)
		motionDX  int
		motionDY  int
	}{
		{
			name:      "right/detail-focus",
			setup:     setupRatioModelRight,
			dividerFn: func(m Model) (int, int) { return rightDividerX(m), 5 },
			motionDX:  -10,
		},
		{
			name:      "right/list-focus",
			setup:     func(t *testing.T, _ bool) Model { return setupRatioModelRight(t, true) },
			dividerFn: func(m Model) (int, int) { return rightDividerX(m), 5 },
			motionDX:  -10,
		},
		{
			name:      "below/detail-focus",
			setup:     setupRatioModelBelow,
			dividerFn: func(m Model) (int, int) { return 20, belowDividerY(m) },
			motionDY:  -3,
		},
		{
			name:      "below/list-focus",
			setup:     func(t *testing.T, _ bool) Model { return setupRatioModelBelow(t, true) },
			dividerFn: func(m Model) (int, int) { return 20, belowDividerY(m) },
			motionDY:  -3,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := tc.setup(t, false)
			startFocus := m.focus
			x, y := tc.dividerFn(m)
			m = send(m, tea.MouseMsg{X: x, Y: y, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
			m = send(m, tea.MouseMsg{X: x + tc.motionDX, Y: y + tc.motionDY, Button: tea.MouseButtonLeft, Action: tea.MouseActionMotion})
			m = send(m, tea.MouseMsg{X: x + tc.motionDX, Y: y + tc.motionDY, Button: tea.MouseButtonLeft, Action: tea.MouseActionRelease})
			assert.Equalf(t, startFocus, m.focus,
				"drag changed focus: start=%v end=%v", startFocus, m.focus)
		})
	}
}

// TestModel_T156_PaneClosed_PressIsNoOp: pressing anywhere with the pane
// closed never starts a drag session (there is no divider).
func TestModel_T156_PaneClosed_PressIsNoOp(t *testing.T) {
	dir := t.TempDir()
	cfgPath := dir + "/config.toml"
	cfg := config.LoadResult{Config: config.DefaultConfig()}

	for _, size := range []struct{ w, h int }{{80, 24}, {200, 24}} {
		m := New("", false, cfgPath, cfg)
		m = resize(m, size.w, size.h)
		m = m.SetEntries(makeEntries(3))
		require.Falsef(t, m.pane.IsOpen(), "precondition: pane must be closed at %dx%d", size.w, size.h)
		m = send(m, tea.MouseMsg{X: 10, Y: 10, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
		assert.Falsef(t, m.draggingDivider,
			"%dx%d press with pane closed must not start drag", size.w, size.h)
	}
	_, err := os.Stat(cfgPath)
	assert.Truef(t, os.IsNotExist(err),
		"pane-closed press must not save config; file exists: %v", err)
}

// ---------- T-157: divider-cell click is focus-neutral (cavekit-app-shell R6 AC 7) ----------

// TestModel_T157_DividerClick_DoesNotTransferFocus verifies the R6 contract:
// a Press + Release on the divider cell itself never transfers focus to
// either pane. The divider is reserved for R15 drag initiation; focus
// stays wherever it was before the click.
func TestModel_T157_DividerClick_DoesNotTransferFocus(t *testing.T) {
	cases := []struct {
		name  string
		setup func(t *testing.T, listFocus bool) Model
		xy    func(m Model) (int, int)
	}{
		{
			name:  "right/start-list-focus",
			setup: func(t *testing.T, _ bool) Model { return setupRatioModelRight(t, true) },
			xy:    func(m Model) (int, int) { return rightDividerX(m), 5 },
		},
		{
			name:  "right/start-detail-focus",
			setup: func(t *testing.T, _ bool) Model { return setupRatioModelRight(t, false) },
			xy:    func(m Model) (int, int) { return rightDividerX(m), 5 },
		},
		{
			name:  "below/start-list-focus",
			setup: func(t *testing.T, _ bool) Model { return setupRatioModelBelow(t, true) },
			xy:    func(m Model) (int, int) { return 20, belowDividerY(m) },
		},
		{
			name:  "below/start-detail-focus",
			setup: func(t *testing.T, _ bool) Model { return setupRatioModelBelow(t, false) },
			xy:    func(m Model) (int, int) { return 20, belowDividerY(m) },
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := tc.setup(t, false)
			startFocus := m.focus
			x, y := tc.xy(m)
			// Bare click: Press then Release with no motion between.
			m = send(m, tea.MouseMsg{X: x, Y: y, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
			m = send(m, tea.MouseMsg{X: x, Y: y, Button: tea.MouseButtonLeft, Action: tea.MouseActionRelease})
			assert.Equalf(t, startFocus, m.focus,
				"divider click changed focus: start=%v end=%v", startFocus, m.focus)
		})
	}
}

// TestModel_T156_RightDrag_UpdatesWidthRatio mirrors the T-104 coverage
// after the dual-orientation refactor — confirms right-split drag still
// updates width_ratio (not height_ratio) and saves once on release.
func TestModel_T156_RightDrag_UpdatesWidthRatio(t *testing.T) {
	dir := t.TempDir()
	cfgPath := dir + "/config.toml"
	cfg := config.LoadResult{Config: config.DefaultConfig()}
	m := New("", false, cfgPath, cfg)
	m = resize(m, 200, 24)
	entries := makeEntries(3)
	m = m.SetEntries(entries)
	m = m.openPane(entries[0])
	require.Equalf(t, appshell.OrientationRight, m.resize.Orientation(),
		"precondition: want right, got %v", m.resize.Orientation())
	beforeH := m.cfg.Config.DetailPane.HeightRatio
	dx := rightDividerX(m)

	m = send(m, tea.MouseMsg{X: dx, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	m = send(m, tea.MouseMsg{X: dx - 20, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionMotion})
	m = send(m, tea.MouseMsg{X: dx - 20, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionRelease})

	assert.Equalf(t, beforeH, m.cfg.Config.DetailPane.HeightRatio,
		"right-mode drag must not mutate height_ratio: %.3f → %.3f",
		beforeH, m.cfg.Config.DetailPane.HeightRatio)
	reloaded := config.Load(cfgPath)
	assert.Equalf(t, m.cfg.Config.DetailPane.WidthRatio, reloaded.Config.DetailPane.WidthRatio,
		"disk width_ratio after right-drag release: got %.3f, want %.3f",
		reloaded.Config.DetailPane.WidthRatio, m.cfg.Config.DetailPane.WidthRatio)
}

// ---------- T-162: mid-drag auto-close terminates the drag (F-125) ----------

// TestModel_T162_Drag_AutoClose_TerminatesSession verifies that if the
// pane auto-closes mid-drag (e.g. terminal shrinks below the right-split
// min-width threshold), the subsequent Motion + Release stream does NOT
// mutate the persisted ratio and does NOT write to the config file. The
// belt-and-braces clear in the WindowSizeMsg auto-close branch guarantees
// `draggingDivider` is false by the time the next Motion arrives.
func TestModel_T162_Drag_AutoClose_TerminatesSession(t *testing.T) {
	dir := t.TempDir()
	cfgPath := dir + "/config.toml"
	cfg := config.LoadResult{Config: config.DefaultConfig()}
	cfg.Config.DetailPane.Position = "right" // pin orientation across resizes
	m := New("", false, cfgPath, cfg)
	m = resize(m, 200, 24)
	entries := makeEntries(3)
	m = m.SetEntries(entries)
	m = m.openPane(entries[0])
	require.True(t, m.pane.IsOpen(), "precondition: pane must be open")
	require.Equalf(t, appshell.OrientationRight, m.resize.Orientation(),
		"precondition: want right orientation, got %v", m.resize.Orientation())
	dividerX := rightDividerX(m)
	startW := m.cfg.Config.DetailPane.WidthRatio

	// Begin a drag with a Press on the divider.
	m = send(m, tea.MouseMsg{X: dividerX, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	require.Truef(t, m.draggingDivider,
		"precondition: Press at divider x=%d must start drag", dividerX)

	// Shrink the terminal so the right-split detail width falls below
	// MinDetailWidth (30). Usable = w-5, detailW = usable * 0.30.
	// Using w=70 → usable=65 → detailW=19 < 30 → auto-close fires.
	m = resize(m, 70, 24)
	require.False(t, m.pane.IsOpen(), "precondition: pane must have auto-closed at w=70")
	assert.False(t, m.draggingDivider, "auto-close must clear draggingDivider (belt-and-braces)")

	// Any subsequent Motion + Release must be a no-op for ratio + config.
	m = send(m, tea.MouseMsg{X: 20, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionMotion})
	m = send(m, tea.MouseMsg{X: 20, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionRelease})

	assert.Equalf(t, startW, m.cfg.Config.DetailPane.WidthRatio,
		"width_ratio mutated after mid-drag auto-close: start=%.3f end=%.3f",
		startW, m.cfg.Config.DetailPane.WidthRatio)
	_, err := os.Stat(cfgPath)
	assert.Truef(t, os.IsNotExist(err),
		"mid-drag auto-close must NOT write config; file exists: %v", err)
}

// TestModel_T162_DragBranch_GuardsOnClosedPane covers the inner guard —
// if something somehow leaves draggingDivider=true with pane closed (e.g.
// future code path we haven't anticipated), the drag branch must still
// short-circuit before touching the ratio or saving config.
func TestModel_T162_DragBranch_GuardsOnClosedPane(t *testing.T) {
	dir := t.TempDir()
	cfgPath := dir + "/config.toml"
	cfg := config.LoadResult{Config: config.DefaultConfig()}
	cfg.Config.DetailPane.Position = "right"
	m := New("", false, cfgPath, cfg)
	m = resize(m, 200, 24)
	m = m.SetEntries(makeEntries(3))
	// Artificial stale state: pane closed but drag flag set.
	m.draggingDivider = true
	startW := m.cfg.Config.DetailPane.WidthRatio

	m = send(m, tea.MouseMsg{X: 50, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionMotion})
	assert.False(t, m.draggingDivider, "closed-pane drag-branch guard must clear draggingDivider")
	assert.Equalf(t, startW, m.cfg.Config.DetailPane.WidthRatio,
		"closed-pane drag-branch must not mutate ratio: %.3f → %.3f",
		startW, m.cfg.Config.DetailPane.WidthRatio)
	_, err := os.Stat(cfgPath)
	assert.Truef(t, os.IsNotExist(err),
		"closed-pane drag-branch must not write config; file exists: %v", err)
}

// ---------- T-164: bare Press+Release skips config write (F-129) ----------

// TestModel_T164_BareClick_OnDivider_DoesNotWriteConfig verifies that a
// Press immediately followed by a Release on the divider (no intervening
// Motion) does NOT write the config file. Previously the Release branch
// unconditionally called saveConfig, so bare clicks on the divider
// rewrote `config.toml` every time a user clicked near the border.
func TestModel_T164_BareClick_OnDivider_DoesNotWriteConfig(t *testing.T) {
	dir := t.TempDir()
	cfgPath := dir + "/config.toml"
	cfg := config.LoadResult{Config: config.DefaultConfig()}
	cfg.Config.DetailPane.Position = "right"
	m := New("", false, cfgPath, cfg)
	m = resize(m, 200, 24)
	entries := makeEntries(3)
	m = m.SetEntries(entries)
	m = m.openPane(entries[0])
	dividerX := rightDividerX(m)
	startW := m.cfg.Config.DetailPane.WidthRatio

	m = send(m, tea.MouseMsg{X: dividerX, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	require.Truef(t, m.draggingDivider,
		"precondition: Press at divider x=%d must start drag", dividerX)
	assert.False(t, m.dragDirty, "Press must reset dragDirty to false")
	m = send(m, tea.MouseMsg{X: dividerX, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionRelease})

	assert.False(t, m.draggingDivider, "Release must end drag session")
	assert.Equalf(t, startW, m.cfg.Config.DetailPane.WidthRatio,
		"bare click mutated width_ratio: start=%.3f end=%.3f",
		startW, m.cfg.Config.DetailPane.WidthRatio)
	_, err := os.Stat(cfgPath)
	assert.Truef(t, os.IsNotExist(err),
		"bare Press+Release on divider must NOT write config; file exists: %v", err)
}

// TestModel_T164_Drag_WithMotion_DoesWriteConfig is the positive-case
// anchor — a drag that actually moves the divider still flushes the new
// ratio to disk on Release. Without this, T-164 would be trivially
// satisfied by disabling all writes.
func TestModel_T164_Drag_WithMotion_DoesWriteConfig(t *testing.T) {
	dir := t.TempDir()
	cfgPath := dir + "/config.toml"
	cfg := config.LoadResult{Config: config.DefaultConfig()}
	cfg.Config.DetailPane.Position = "right"
	m := New("", false, cfgPath, cfg)
	m = resize(m, 200, 24)
	entries := makeEntries(3)
	m = m.SetEntries(entries)
	m = m.openPane(entries[0])
	dividerX := rightDividerX(m)

	m = send(m, tea.MouseMsg{X: dividerX, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	m = send(m, tea.MouseMsg{X: dividerX - 30, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionMotion})
	assert.True(t, m.dragDirty, "Motion that changes ratio must set dragDirty=true")
	m = send(m, tea.MouseMsg{X: dividerX - 30, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionRelease})

	_, err := os.Stat(cfgPath)
	assert.NoErrorf(t, err, "drag with motion must write config: %v", err)
}
