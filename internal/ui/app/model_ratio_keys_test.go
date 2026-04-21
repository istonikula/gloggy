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

// ---------- T-099: ratio live write-back to config ----------

// TestModel_RatioKey_PersistsToConfigFile verifies that pressing a ratio
// key (here `+` in right-split) writes the new width_ratio to disk.
func TestModel_RatioKey_PersistsToConfigFile(t *testing.T) {
	dir := t.TempDir()
	cfgPath := dir + "/config.toml"

	cfg := config.LoadResult{Config: config.DefaultConfig()}
	m := New("", false, cfgPath, cfg)
	m = resize(m, 200, 24)
	m = m.SetEntries(makeEntries(3))
	m = m.openPane(makeEntries(3)[0])
	// T-126: openPane no longer auto-focuses the pane, but the `+` ratio
	// key only fires when the pane is focused (handleKey's FocusDetailPane
	// branch). Simulate the Tab-to-pane step.
	m = setFocus(m, appshell.FocusDetailPane)

	require.Equalf(t, appshell.OrientationRight, m.resize.Orientation(),
		"precondition: want right orientation, got %v", m.resize.Orientation())
	beforeWidth := m.cfg.Config.DetailPane.WidthRatio
	beforeHeight := m.cfg.Config.DetailPane.HeightRatio

	m = key(m, "+") // increment width_ratio in right-split
	require.NotEqualf(t, beforeWidth, m.cfg.Config.DetailPane.WidthRatio,
		"'+' did not change in-memory width_ratio: %.3f", m.cfg.Config.DetailPane.WidthRatio)

	// Reload from disk and verify width_ratio persisted.
	reloaded := config.Load(cfgPath)
	assert.Equalf(t, m.cfg.Config.DetailPane.WidthRatio, reloaded.Config.DetailPane.WidthRatio,
		"disk width_ratio: got %.3f, want %.3f",
		reloaded.Config.DetailPane.WidthRatio, m.cfg.Config.DetailPane.WidthRatio)
	// height_ratio must NOT have been clobbered.
	assert.Equalf(t, beforeHeight, reloaded.Config.DetailPane.HeightRatio,
		"disk height_ratio mutated: got %.3f, want %.3f (untouched)",
		reloaded.Config.DetailPane.HeightRatio, beforeHeight)
	_ = m
}

// ---------- T-105: orientation flip preserves both ratios ----------

// TestModel_OrientationFlip_PreservesBothRatios verifies that flipping from
// right → below → right does not mutate height_ratio or width_ratio in the
// in-memory config, regardless of which ratio is "active" for the current
// orientation.
func TestModel_OrientationFlip_PreservesBothRatios(t *testing.T) {
	cfg := config.LoadResult{Config: config.DefaultConfig()}
	cfg.Config.DetailPane.HeightRatio = 0.60
	cfg.Config.DetailPane.WidthRatio = 0.20

	m := New("", false, "", cfg)
	require.Equalf(t, 0.60, m.cfg.Config.DetailPane.HeightRatio,
		"pre-resize height_ratio: got %.2f, want 0.60", m.cfg.Config.DetailPane.HeightRatio)

	assertRatios := func(step string) {
		t.Helper()
		assert.Equalf(t, 0.60, m.cfg.Config.DetailPane.HeightRatio,
			"%s: height_ratio = %.3f, want 0.600 (unmutated)", step, m.cfg.Config.DetailPane.HeightRatio)
		assert.Equalf(t, 0.20, m.cfg.Config.DetailPane.WidthRatio,
			"%s: width_ratio = %.3f, want 0.200 (unmutated)", step, m.cfg.Config.DetailPane.WidthRatio)
	}

	m = resize(m, 200, 24) // right
	require.Equalf(t, appshell.OrientationRight, m.resize.Orientation(),
		"precondition: 200 cols should be right, got %v", m.resize.Orientation())
	assertRatios("after initial right resize")

	m = resize(m, 80, 24) // below (under 100-col threshold)
	require.Equalf(t, appshell.OrientationBelow, m.resize.Orientation(),
		"precondition: 80 cols should be below, got %v", m.resize.Orientation())
	assertRatios("after flip to below")

	m = resize(m, 200, 24) // back to right
	require.Equalf(t, appshell.OrientationRight, m.resize.Orientation(),
		"precondition: 200 cols should flip back to right, got %v", m.resize.Orientation())
	assertRatios("after flip back to right")
}

// ---------- T-123: detail-pane vertical height in right orientation (P0 / F-013) ----------

// TestModel_PaneHeight_RightOrientation_UsesFullSlot verifies that in
// right-split the pane's ContentHeight fills the main-area slot
// (terminal_height - header - status - border_rows), not height_ratio.
func TestModel_PaneHeight_RightOrientation_UsesFullSlot(t *testing.T) {
	cfg := config.LoadResult{Config: config.DefaultConfig()}
	cfg.Config.DetailPane.Position = "right" // force right so auto-threshold is bypassed
	cfg.Config.DetailPane.HeightRatio = 0.30 // the below-mode value that used to clip

	m := New("", false, "", cfg)
	m = resize(m, 140, 24) // wide enough to keep right-split stable (detail ≥ 30 cells)
	entries := makeEntries(3)
	m = m.SetEntries(entries)
	m = m.openPane(entries[0])

	require.Equalf(t, appshell.OrientationRight, m.resize.Orientation(),
		"precondition: forced orientation=right, got %v", m.resize.Orientation())
	// 24 (height) - 1 (header) - 1 (status) = 22 outer rows → ContentHeight ≥ 20.
	assert.GreaterOrEqualf(t, m.pane.ContentHeight(), 20,
		"right-split ContentHeight: got %d, want ≥ 20 (F-013)", m.pane.ContentHeight())
	// Regression: height_ratio × terminalHeight = 7. If ContentHeight is ≤ 5
	// we are still applying below-mode math in right orientation.
	assert.Greaterf(t, m.pane.ContentHeight(), 5,
		"right-split ContentHeight = %d — still clipped to height_ratio", m.pane.ContentHeight())
}

// TestModel_PaneHeight_BelowOrientation_UsesRatio verifies the ratio-based
// vertical sizing still governs below-mode.
func TestModel_PaneHeight_BelowOrientation_UsesRatio(t *testing.T) {
	cfg := config.LoadResult{Config: config.DefaultConfig()}
	cfg.Config.DetailPane.Position = "below"
	cfg.Config.DetailPane.HeightRatio = 0.30

	m := New("", false, "", cfg)
	m = resize(m, 80, 24) // below-mode orientation
	entries := makeEntries(3)
	m = m.SetEntries(entries)
	m = m.openPane(entries[0])

	require.Equalf(t, appshell.OrientationBelow, m.resize.Orientation(),
		"precondition: forced orientation=below, got %v", m.resize.Orientation())
	// 24 × 0.30 = 7 outer rows → ContentHeight = 5.
	wantOuter := 7 // int(24 * 0.30) = 7
	assert.Equalf(t, wantOuter, m.paneHeight.PaneHeight(),
		"below-mode outer height: got %d, want %d", m.paneHeight.PaneHeight(), wantOuter)
	assert.Equalf(t, wantOuter-2, m.pane.ContentHeight(),
		"below-mode ContentHeight: got %d, want %d", m.pane.ContentHeight(), wantOuter-2)
}

// TestModel_RatioKey_StillWorksInBelowMode verifies that `+` in below-mode
// still increases height_ratio after the T-123 refactor (guards against
// accidental regression on the resize keymap path).
func TestModel_RatioKey_StillWorksInBelowMode(t *testing.T) {
	cfg := config.LoadResult{Config: config.DefaultConfig()}
	cfg.Config.DetailPane.Position = "below"
	cfg.Config.DetailPane.HeightRatio = 0.30

	m := New("", false, "", cfg)
	m = resize(m, 80, 30)
	entries := makeEntries(3)
	m = m.SetEntries(entries)
	m = m.openPane(entries[0])
	// Transfer focus to the pane so the ratio keymap fires.
	m = setFocus(m, appshell.FocusDetailPane)

	before := m.paneHeight.Ratio()
	m = key(m, "+")
	assert.Greaterf(t, m.paneHeight.Ratio(), before,
		"`+` in below-mode: ratio = %.2f, want > %.2f", m.paneHeight.Ratio(), before)
}

// TestModel_OrientationFlip_VerticalSizeTracks verifies that flipping from
// below to right uses the full main-area slot (T-123 F-013), and flipping
// back restores the below-mode ratio-derived height.
func TestModel_OrientationFlip_VerticalSizeTracks(t *testing.T) {
	cfg := config.LoadResult{Config: config.DefaultConfig()}
	cfg.Config.DetailPane.Position = "auto"
	cfg.Config.DetailPane.HeightRatio = 0.30
	cfg.Config.DetailPane.OrientationThresholdCols = 100

	m := New("", false, "", cfg)
	m = resize(m, 80, 24) // below (under threshold)
	entries := makeEntries(3)
	m = m.SetEntries(entries)
	m = m.openPane(entries[0])

	require.Equalf(t, appshell.OrientationBelow, m.resize.Orientation(),
		"precondition: 80 cols should be below, got %v", m.resize.Orientation())
	belowOuter := m.paneHeight.PaneHeight()
	assert.Equalf(t, 7, belowOuter, "below outer height: got %d, want %d", belowOuter, 7)

	// Flip to right — wide terminal, detailW ≥ 30 so pane stays open.
	m = resize(m, 140, 24)
	require.Equalf(t, appshell.OrientationRight, m.resize.Orientation(),
		"precondition: 140 cols should flip to right, got %v", m.resize.Orientation())
	require.True(t, m.pane.IsOpen(), "precondition: pane should still be open after flip to right")
	// Right-mode vertical = full main slot = 24 - 1 - 1 = 22.
	assert.GreaterOrEqualf(t, m.pane.ContentHeight(), 20,
		"right-mode ContentHeight after flip: got %d, want ≥ 20", m.pane.ContentHeight())

	// Flip back to below — height_ratio preserved, ContentHeight returns to
	// the ratio-derived value.
	m = resize(m, 80, 24)
	require.Equalf(t, appshell.OrientationBelow, m.resize.Orientation(),
		"precondition: 80 cols should flip back to below, got %v", m.resize.Orientation())
	assert.Equalf(t, belowOuter, m.paneHeight.PaneHeight(),
		"below outer height after round-trip: got %d, want %d (ratio not preserved)",
		m.paneHeight.PaneHeight(), belowOuter)
	assert.Equalf(t, 0.30, m.cfg.Config.DetailPane.HeightRatio,
		"height_ratio mutated by orientation flip: got %.3f, want 0.300",
		m.cfg.Config.DetailPane.HeightRatio)
}

// ---------- T-155: focus-aware keyboard resize (cavekit-app-shell R12 revised) ----------

// setupRatioModel opens the detail pane at the given terminal size and
// applies the requested focus. Common fixture for T-155/T-156/T-157/T-6.
func setupRatioModel(t *testing.T, w, h int, listFocus bool) Model {
	t.Helper()
	dir := t.TempDir()
	cfg := config.LoadResult{Config: config.DefaultConfig()}
	m := New("", false, dir+"/config.toml", cfg)
	m = resize(m, w, h)
	entries := makeEntries(3)
	m = m.SetEntries(entries)
	m = m.openPane(entries[0])
	var f appshell.FocusTarget = appshell.FocusDetailPane
	if listFocus {
		f = appshell.FocusEntryList
	}
	return setFocus(m, f)
}

// setupRatioModelRight returns a model in right-orientation with the
// detail pane open. Ratios are the defaults: detail width_ratio = 0.30,
// so list share = 0.70.
func setupRatioModelRight(t *testing.T, listFocus bool) Model {
	t.Helper()
	m := setupRatioModel(t, 200, 24, listFocus)
	require.Equalf(t, appshell.OrientationRight, m.resize.Orientation(),
		"precondition: want right orientation, got %v", m.resize.Orientation())
	return m
}

// setupRatioModelBelow returns a model in below-orientation with the
// detail pane open.
func setupRatioModelBelow(t *testing.T, listFocus bool) Model {
	t.Helper()
	m := setupRatioModel(t, 80, 24, listFocus)
	require.Equalf(t, appshell.OrientationBelow, m.resize.Orientation(),
		"precondition: want below orientation, got %v", m.resize.Orientation())
	return m
}

// TestModel_T155_RatioKeys_Right drives every +/-/|/= scenario in
// right-split through one table. Each step applies a key and runs its
// assertion with the before/after width_ratio values. Multi-step entries
// model the |-toggle round-trips.
func TestModel_T155_RatioKeys_Right(t *testing.T) {
	type step struct {
		key    string
		assert func(t *testing.T, before, after float64)
	}
	greater := func(t *testing.T, b, a float64) {
		t.Helper()
		assert.Greaterf(t, a, b, "ratio must grow: before=%.3f after=%.3f", b, a)
	}
	less := func(t *testing.T, b, a float64) {
		t.Helper()
		assert.Lessf(t, a, b, "ratio must shrink: before=%.3f after=%.3f", b, a)
	}
	eq := func(want float64) func(t *testing.T, before, after float64) {
		return func(t *testing.T, _, a float64) {
			t.Helper()
			assert.Equalf(t, want, a, "ratio: got %.3f, want %.3f", a, want)
		}
	}
	cases := []struct {
		name      string
		seed      float64 // 0 = leave default 0.30
		listFocus bool
		steps     []step
	}{
		{"plus_detail_grows", 0, false, []step{{"+", greater}}},
		{"plus_list_shrinks", 0, true, []step{{"+", less}}},
		{"minus_detail_shrinks", 0, false, []step{{"-", less}}},
		{"minus_list_grows", 0, true, []step{{"-", greater}}},
		{"pipe_detail_0.30_roundtrip", 0, false, []step{{"|", eq(0.50)}, {"|", eq(0.30)}}},
		{"pipe_list_from_0.30_cycles_detail_0.70_to_0.50", 0, true, []step{{"|", eq(0.70)}, {"|", eq(0.50)}}},
		{"equals_detail_resets", 0.50, false, []step{{"=", eq(appshell.RatioDefault)}}},
		{"equals_list_resets", 0.50, true, []step{{"=", eq(appshell.RatioDefault)}}},
		{"plus_detail_at_max_noop", appshell.RatioMax, false, []step{{"+", eq(appshell.RatioMax)}}},
		{"minus_detail_at_min_noop", appshell.RatioMin, false, []step{{"-", eq(appshell.RatioMin)}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := setupRatioModelRight(t, tc.listFocus)
			if tc.seed != 0 {
				m.cfg.Config.DetailPane.WidthRatio = tc.seed
				m.layout = m.layout.SetWidthRatio(tc.seed)
			}
			for _, s := range tc.steps {
				before := m.cfg.Config.DetailPane.WidthRatio
				m = key(m, s.key)
				s.assert(t, before, m.cfg.Config.DetailPane.WidthRatio)
			}
		})
	}
}

// TestModel_T155_Below_ActivatesHeightRatio: below-mode mutates
// height_ratio (not width_ratio).
func TestModel_T155_Below_ActivatesHeightRatio(t *testing.T) {
	m := setupRatioModelBelow(t, false)
	beforeH := m.cfg.Config.DetailPane.HeightRatio
	beforeW := m.cfg.Config.DetailPane.WidthRatio
	m = key(m, "+")
	assert.NotEqualf(t, beforeH, m.cfg.Config.DetailPane.HeightRatio,
		"+ in below-mode must change height_ratio: %.3f unchanged", beforeH)
	assert.Equalf(t, beforeW, m.cfg.Config.DetailPane.WidthRatio,
		"+ in below-mode must NOT change width_ratio: %.3f → %.3f",
		beforeW, m.cfg.Config.DetailPane.WidthRatio)
}

// TestModel_T155_PaneClosed_AllKeys_NoOp: all four ratio keys are silent
// no-ops when the detail pane is closed (no divider to move).
func TestModel_T155_PaneClosed_AllKeys_NoOp(t *testing.T) {
	dir := t.TempDir()
	cfgPath := dir + "/config.toml"
	cfg := config.LoadResult{Config: config.DefaultConfig()}
	m := New("", false, cfgPath, cfg)
	m = resize(m, 200, 24)
	m = m.SetEntries(makeEntries(3))
	require.False(t, m.pane.IsOpen(), "precondition: pane should be closed")
	beforeW := m.cfg.Config.DetailPane.WidthRatio
	beforeH := m.cfg.Config.DetailPane.HeightRatio

	for _, k := range []string{"+", "-", "=", "|"} {
		m = key(m, k)
		assert.Equalf(t, beforeW, m.cfg.Config.DetailPane.WidthRatio,
			"%q with pane closed must not change width_ratio: %.3f → %.3f",
			k, beforeW, m.cfg.Config.DetailPane.WidthRatio)
		assert.Equalf(t, beforeH, m.cfg.Config.DetailPane.HeightRatio,
			"%q with pane closed must not change height_ratio: %.3f → %.3f",
			k, beforeH, m.cfg.Config.DetailPane.HeightRatio)
	}
	// No disk write should have fired — config file must not exist.
	_, err := os.Stat(cfgPath)
	assert.Truef(t, os.IsNotExist(err),
		"pane-closed ratio keys must not trigger config save; file exists: %v", err)
}

// ---------- F-132: degenerate-dim guard fidelity (supersedes T-165) ----------
//
// The original T-165 tests drove a 0-dim WindowSizeMsg which auto-closed
// the pane, then re-set draggingDivider=true and sent Motion. With pane
// closed, the !m.pane.IsOpen() guard at handleMouse model.go:524
// short-circuits BEFORE the termW/termH<=0 guard at :554-556 / :565-567
// is ever reached. Removing the degenerate-dim guards left the T-165
// tests green — proving the tests asserted behaviour via the wrong code
// path. /ck:review Pass 2 (F-132). The replacement tests below force the
// pane open at Motion time so the IsOpen() guard passes and the only
// thing standing between Motion and ratio-shadowing is the degenerate-dim
// guard. cavekit-app-shell.md R15 degenerate-dim AC was sharpened in the
// same /ck:revise --trace cycle to mandate this test shape.

// TestModel_F132_DegenerateDim_Right_GuardFiresWith_PaneOpen pins the
// right-split termW<=0 caller-guard at model.go:554-556. Forces 0-dim
// resize → pane auto-closes → re-opens pane via openPane (relayout does
// NOT re-trigger auto-close, so pane stays open with termW=0) → activates
// drag flag → sends Motion. If the guard is removed,
// RatioFromDragX(10, 0) returns ClampRatio(RatioDefault)=0.30 and shadows
// the persisted 0.55 — failing this test.
func TestModel_F132_DegenerateDim_Right_GuardFiresWith_PaneOpen(t *testing.T) {
	dir := t.TempDir()
	cfgPath := dir + "/config.toml"
	cfg := config.LoadResult{Config: config.DefaultConfig()}
	cfg.Config.DetailPane.Position = "right"
	m := New("", false, cfgPath, cfg)
	m = resize(m, 200, 24)
	entries := makeEntries(3)
	m = m.SetEntries(entries)
	m = m.openPane(entries[0])
	m.cfg.Config.DetailPane.WidthRatio = 0.55
	m.layout = m.layout.SetWidthRatio(0.55)

	// Force termW=0 — auto-close fires, pane closes, draggingDivider clears.
	m = send(m, tea.WindowSizeMsg{Width: 0, Height: 24})
	require.False(t, m.pane.IsOpen(), "precondition: pane must auto-close at width=0")
	// Re-open the pane. relayout() does not re-evaluate ShouldAutoCloseDetail
	// (that lives in the WindowSizeMsg handler only), so the pane stays open
	// even though m.resize.Width() is still 0.
	m = m.openPane(entries[0])
	require.True(t, m.pane.IsOpen(), "precondition: pane must re-open via openPane")
	require.Equalf(t, 0, m.resize.Width(), "precondition: termW must remain 0, got %d", m.resize.Width())
	// All preconditions for the termW<=0 guard are now met: pane open,
	// drag flag set, Motion incoming with termW=0. The IsOpen() guard
	// passes — the only thing preventing ratio shadowing is :554-556.
	m.draggingDivider = true
	startRatio := m.cfg.Config.DetailPane.WidthRatio

	m = send(m, tea.MouseMsg{X: 10, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionMotion})

	assert.Equalf(t, startRatio, m.cfg.Config.DetailPane.WidthRatio,
		"termW<=0 guard absent or bypassed: ratio shadowed from %.3f to %.3f",
		startRatio, m.cfg.Config.DetailPane.WidthRatio)
}

// TestModel_F132_DegenerateDim_Below_GuardFiresWith_PaneOpen is the
// below-mode analog pinning model.go:565-567 (termH<=0 guard).
func TestModel_F132_DegenerateDim_Below_GuardFiresWith_PaneOpen(t *testing.T) {
	dir := t.TempDir()
	cfgPath := dir + "/config.toml"
	cfg := config.LoadResult{Config: config.DefaultConfig()}
	cfg.Config.DetailPane.Position = "below"
	m := New("", false, cfgPath, cfg)
	m = resize(m, 80, 24)
	entries := makeEntries(3)
	m = m.SetEntries(entries)
	m = m.openPane(entries[0])
	m.cfg.Config.DetailPane.HeightRatio = 0.45
	m.paneHeight = m.paneHeight.SetRatio(0.45)

	m = send(m, tea.WindowSizeMsg{Width: 80, Height: 0})
	require.False(t, m.pane.IsOpen(), "precondition: pane must auto-close at height=0")
	m = m.openPane(entries[0])
	require.True(t, m.pane.IsOpen(), "precondition: pane must re-open via openPane")
	require.Equalf(t, 0, m.resize.Height(), "precondition: termH must remain 0, got %d", m.resize.Height())
	m.draggingDivider = true
	startRatio := m.cfg.Config.DetailPane.HeightRatio

	m = send(m, tea.MouseMsg{X: 20, Y: 3, Button: tea.MouseButtonLeft, Action: tea.MouseActionMotion})

	assert.Equalf(t, startRatio, m.cfg.Config.DetailPane.HeightRatio,
		"termH<=0 guard absent or bypassed: ratio shadowed from %.3f to %.3f",
		startRatio, m.cfg.Config.DetailPane.HeightRatio)
}

// ---------- T6 (B3): handleRatioKey guards saveConfig on newR != current ----------

// TestModel_T6_RatioKey_NoOpAtBoundary_DoesNotWriteConfig covers V17 /
// B3: at the clamp-pin or preset no-op, the ratio value is unchanged and
// saveConfig must be skipped. Previously `handleRatioKey` wrote the
// config file on every keypress, inflating mtime + disk I/O at the
// boundaries where the ratio cannot move. Each subtest starts from an
// empty cfgPath; if the guard fires, the file must not exist after the
// no-op press.
func TestModel_T6_RatioKey_NoOpAtBoundary_DoesNotWriteConfig(t *testing.T) {
	cases := []struct {
		name      string
		key       string
		listFocus bool
		seedRatio float64
	}{
		{"plus_at_max_detail_focus", "+", false, appshell.RatioMax},
		{"minus_at_min_detail_focus", "-", false, appshell.RatioMin},
		{"plus_at_min_list_focus", "+", true, appshell.RatioMin},
		{"minus_at_max_list_focus", "-", true, appshell.RatioMax},
		{"equals_at_default_detail_focus", "=", false, appshell.RatioDefault},
		{"equals_at_default_list_focus", "=", true, appshell.RatioDefault},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			cfgPath := dir + "/config.toml"

			cfg := config.LoadResult{Config: config.DefaultConfig()}
			cfg.Config.DetailPane.WidthRatio = tc.seedRatio
			m := New("", false, cfgPath, cfg)
			m = resize(m, 200, 24)
			entries := makeEntries(3)
			m = m.SetEntries(entries)
			m = m.openPane(entries[0])
			require.Equalf(t, appshell.OrientationRight, m.resize.Orientation(),
				"precondition: want right orientation at 200 cols, got %v", m.resize.Orientation())
			var f appshell.FocusTarget = appshell.FocusDetailPane
			if tc.listFocus {
				f = appshell.FocusEntryList
			}
			m = setFocus(m, f)

			before := m.cfg.Config.DetailPane.WidthRatio
			m = key(m, tc.key)
			after := m.cfg.Config.DetailPane.WidthRatio

			require.Equalf(t, before, after,
				"precondition: %q at ratio=%.2f listFocus=%v should be a no-op; got %.3f → %.3f",
				tc.key, tc.seedRatio, tc.listFocus, before, after)
			_, err := os.Stat(cfgPath)
			assert.Truef(t, os.IsNotExist(err),
				"B3 regression: no-op %q press wrote config file; stat err: %v", tc.key, err)
		})
	}
}

// TestModel_T6_RatioKey_NoOpAtBoundary_Below_DoesNotWriteConfig mirrors
// the right-split test in below-orientation. Both branches of
// handleRatioKey share the guard; this pins the second branch so a
// future refactor can't drop the guard on just one axis.
func TestModel_T6_RatioKey_NoOpAtBoundary_Below_DoesNotWriteConfig(t *testing.T) {
	dir := t.TempDir()
	cfgPath := dir + "/config.toml"

	cfg := config.LoadResult{Config: config.DefaultConfig()}
	cfg.Config.DetailPane.Position = "below"
	cfg.Config.DetailPane.HeightRatio = appshell.RatioMax
	m := New("", false, cfgPath, cfg)
	m = resize(m, 80, 30)
	entries := makeEntries(3)
	m = m.SetEntries(entries)
	m = m.openPane(entries[0])
	require.Equalf(t, appshell.OrientationBelow, m.resize.Orientation(),
		"precondition: want below orientation, got %v", m.resize.Orientation())
	m = setFocus(m, appshell.FocusDetailPane)

	before := m.paneHeight.Ratio()
	m = key(m, "+") // detail-focus at RatioMax → no-op
	after := m.paneHeight.Ratio()

	require.Equalf(t, before, after,
		"precondition: `+` at height_ratio=%.2f detail-focus should be no-op; got %.3f → %.3f",
		appshell.RatioMax, before, after)
	_, err := os.Stat(cfgPath)
	assert.Truef(t, os.IsNotExist(err),
		"B3 regression (below-mode): no-op `+` wrote config file; stat err: %v", err)
}

// TestModel_T6_RatioKey_Change_WritesConfig is the positive-case anchor.
// The T6 guard only suppresses writes when newR==current; every key that
// actually moves the ratio must still persist. Without this, T6 would
// be trivially satisfied by nuking every saveConfig call. Covers `+`,
// `-`, `=`, and `|` change paths.
func TestModel_T6_RatioKey_Change_WritesConfig(t *testing.T) {
	cases := []struct {
		name      string
		key       string
		seedRatio float64
	}{
		{"plus_from_default", "+", appshell.RatioDefault},  // 0.30 → 0.35
		{"minus_from_default", "-", appshell.RatioDefault}, // 0.30 → 0.25
		{"equals_from_off_default", "=", 0.50},             // 0.50 → 0.30
		{"pipe_from_default", "|", appshell.RatioDefault},  // 0.30 → 0.50 (preset cycle)
		{"pipe_from_off_preset", "|", 0.45},                // off-preset → 0.30
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			cfgPath := dir + "/config.toml"

			cfg := config.LoadResult{Config: config.DefaultConfig()}
			cfg.Config.DetailPane.WidthRatio = tc.seedRatio
			m := New("", false, cfgPath, cfg)
			m = resize(m, 200, 24)
			entries := makeEntries(3)
			m = m.SetEntries(entries)
			m = m.openPane(entries[0])
			require.Equalf(t, appshell.OrientationRight, m.resize.Orientation(),
				"precondition: want right orientation, got %v", m.resize.Orientation())
			m = setFocus(m, appshell.FocusDetailPane)

			before := m.cfg.Config.DetailPane.WidthRatio
			m = key(m, tc.key)
			after := m.cfg.Config.DetailPane.WidthRatio

			require.NotEqualf(t, before, after,
				"precondition: %q at ratio=%.2f should change the value; stayed %.3f",
				tc.key, tc.seedRatio, before)
			_, err := os.Stat(cfgPath)
			assert.NoErrorf(t, err, "%q change path must write config: %v", tc.key, err)
			reloaded := config.Load(cfgPath)
			assert.Equalf(t, after, reloaded.Config.DetailPane.WidthRatio,
				"disk width_ratio: got %.3f, want %.3f", reloaded.Config.DetailPane.WidthRatio, after)
		})
	}
}
