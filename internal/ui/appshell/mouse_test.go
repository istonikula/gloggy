package appshell

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/istonikula/gloggy/internal/theme"
)

// Layout: 80 wide, 24 tall, no detail pane.
// Row 0 = header, rows 1-22 = entry list (22 rows), row 23 = status bar.
func testRouter(detailOpen bool, detailHeight int) MouseRouter {
	l := NewLayout(80, 24, detailOpen, detailHeight)
	return NewMouseRouter(l)
}

// T-052: R6.1 — header row is ZoneHeader.
func TestMouseRouter_HeaderRow(t *testing.T) {
	r := testRouter(false, 0)
	assert.Equalf(t, ZoneHeader, r.Zone(0, 0), "row 0 should be ZoneHeader")
}

// T-052: R6.1 — status bar row is ZoneStatusBar.
func TestMouseRouter_StatusBarRow(t *testing.T) {
	r := testRouter(false, 0)
	assert.Equalf(t, ZoneStatusBar, r.Zone(0, 23), "row 23 should be ZoneStatusBar")
}

// T-052: R6.1 — entry list rows without detail pane.
func TestMouseRouter_EntryListRows_NoDetailPane(t *testing.T) {
	r := testRouter(false, 0)
	for y := 1; y <= 22; y++ {
		assert.Equalf(t, ZoneEntryList, r.Zone(0, y), "row %d should be ZoneEntryList", y)
	}
}

// T-052: R6.1 — detail pane and divider zones when pane is open.
// Layout with 8-row detail pane: header=0, entries=1-13 (13 rows), divider=14,
// detail=15-22, status=23.
func TestMouseRouter_DetailPaneOpen(t *testing.T) {
	// height=24, header=1, status=1, detailPane=8 → entryList=14 rows
	l := NewLayout(80, 24, true, 8)
	r := NewMouseRouter(l)

	entryListHeight := l.EntryListHeight() // 24-1-1-8=14
	// Entry list: rows 1..14
	for y := 1; y <= entryListHeight; y++ {
		assert.Equalf(t, ZoneEntryList, r.Zone(0, y), "row %d should be ZoneEntryList", y)
	}
	// Divider: row 15
	dividerRow := 1 + entryListHeight
	assert.Equalf(t, ZoneDivider, r.Zone(0, dividerRow), "row %d should be ZoneDivider", dividerRow)
	// Detail pane: rows 16..23-1 = rows 16..22
	for y := dividerRow + 1; y < 23; y++ {
		assert.Equalf(t, ZoneDetailPane, r.Zone(0, y), "row %d should be ZoneDetailPane", y)
	}
}

// T-052: R6.2 — no crash on any mouse position (just call Zone for all rows).
func TestMouseRouter_NoCrashAnyPosition(t *testing.T) {
	r := testRouter(true, 8)
	for y := -1; y <= 30; y++ {
		_ = r.Zone(0, y) // must not panic
	}
}

// T-052: RouteMouseMsg classifies tea.MouseMsg.
func TestMouseRouter_RouteMouseMsg(t *testing.T) {
	r := testRouter(false, 0)
	msg := tea.MouseMsg{X: 0, Y: 0}
	assert.Equalf(t, ZoneHeader, r.RouteMouseMsg(msg), "expected ZoneHeader for Y=0")
}

// T-094 (revised T-160, F-122): right-split horizontal zoning with 1-cell
// buffer on each side of the divider column. The visible `│` glyph renders
// at col ListContentWidth(); list right border sits at LCW-1; detail left
// border at LCW+1. Confirmed by
// TestMouseRouter_T160_RendererTruth_DividerColMatchesGlyph.
//
// Layout: width=100, height=24, widthRatio=0.30 →
//   usable         = 100 - 4 - 1 = 95
//   listContent    = int(95*0.7)  = 66 (outer width of list pane)
//   detailContent  = 95 - 66      = 29
//   listRightBrdr  = 66 - 1 = 65  (list pane right border, buffer)
//   divider        = 66           (ZoneDivider — the visible `│`)
//   detailLeftBrdr = 67           (detail pane left border, buffer)
//   detail content = 68..          (29 cols + right border)
func TestMouseRouter_RightSplit_HorizontalZones(t *testing.T) {
	l := NewLayout(100, 24, true, 0)
	l.Orientation = OrientationRight
	l.WidthRatio = 0.30
	r := NewMouseRouter(l)

	const row = 5

	// Click strictly inside list content → list.
	assert.Equalf(t, ZoneEntryList, r.Zone(l.ListContentWidth()-2, row),
		"LCW-2 (%d): want ZoneEntryList", l.ListContentWidth()-2)
	// Click on list right border → unknown.
	listRightBorder := l.ListContentWidth() - 1
	assert.Equalf(t, ZoneUnknown, r.Zone(listRightBorder, row),
		"list right border (%d): want ZoneUnknown", listRightBorder)
	// Click on divider → divider.
	divider := l.ListContentWidth()
	assert.Equalf(t, ZoneDivider, r.Zone(divider, row),
		"divider (%d): want ZoneDivider", divider)
	// Click on detail left border → unknown.
	detailLeftBorder := divider + 1
	assert.Equalf(t, ZoneUnknown, r.Zone(detailLeftBorder, row),
		"detail left border (%d): want ZoneUnknown", detailLeftBorder)
	// Click immediately after the left-border buffer → detail content.
	assert.Equalf(t, ZoneDetailPane, r.Zone(detailLeftBorder+1, row),
		"detailLeftBorder+1 (%d): want ZoneDetailPane", detailLeftBorder+1)
}

// T-094: header + status bar still take precedence over horizontal zones.
func TestMouseRouter_RightSplit_HeaderAndStatus(t *testing.T) {
	l := NewLayout(100, 24, true, 0)
	l.Orientation = OrientationRight
	l.WidthRatio = 0.30
	r := NewMouseRouter(l)
	assert.Equalf(t, ZoneHeader, r.Zone(50, 0), "y=0: want ZoneHeader")
	assert.Equalf(t, ZoneStatusBar, r.Zone(50, 23), "y=23: want ZoneStatusBar")
}

// TestMouseRouter_T160_RendererTruth_DividerColMatchesGlyph is the
// renderer-truth anchor mandated by app-shell/R15 post-T-160. It renders
// the layout with stub pane strings at the widths reported by
// Layout.ListContentWidth / DetailContentWidth, locates the visible `│`
// glyph column programmatically on a mid-row, and asserts
// MouseRouter.Zone at that column returns ZoneDivider. Sweeps both
// terminal sizes from the HUMAN sign-off matrix (140x35, 80x24) across
// presets {0.10, 0.30, 0.50, 0.80}. Regression guard against the
// two-column offset that caused F-122 (T-156 tests passed because they
// used the router's own coordinate helpers to locate the divider; that
// tautology is now broken by this test, which forces agreement with the
// actually-rendered output).
func TestMouseRouter_T160_RendererTruth_DividerColMatchesGlyph(t *testing.T) {
	cases := []struct {
		name  string
		w, h  int
		ratio float64
	}{
		{"140x35 r=0.10", 140, 35, 0.10},
		{"140x35 r=0.30", 140, 35, 0.30},
		{"140x35 r=0.50", 140, 35, 0.50},
		{"140x35 r=0.80", 140, 35, 0.80},
		{"80x24 r=0.30", 80, 24, 0.30},
		{"80x24 r=0.50", 80, 24, 0.50},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			lm := NewLayoutModel(tc.w, tc.h).
				WithTheme(theme.GetTheme("tokyo-night")).
				SetDetailPane(true, 10).
				SetOrientation(OrientationRight).
				SetWidthRatio(tc.ratio)
			lay := lm.Layout()
			listW := lay.ListContentWidth()
			detailW := lay.DetailContentWidth()
			// Stub pane strings sized to their declared slot widths.
			// Use ASCII fillers (not `│`) so the only vertical-bar
			// glyph in the rendered output is the inline divider.
			stubList := strings.Repeat("L", listW)
			stubDetail := strings.Repeat("D", detailW+2)
			out := lm.Render("H", stubList, stubDetail, "S")
			lines := strings.Split(out, "\n")
			require.GreaterOrEqualf(t, len(lines), 3, "render produced too few lines: %d", len(lines))
			midRow := len(lines) / 2
			glyphCol := locateGlyphCol(t, lines[midRow], '│')
			require.GreaterOrEqualf(t, glyphCol, 0, "no `│` glyph on mid-row %d", midRow)
			r := NewMouseRouter(lay)
			assert.Equalf(t, ZoneDivider, r.Zone(glyphCol, midRow),
				"renderer-truth: glyph at col %d; listW=%d detailW=%d",
				glyphCol, listW, detailW)
			// Also verify the router's computed divider column equals
			// the actual glyph col — this is the invariant T-160 fixes.
			assert.Equalf(t, lay.ListContentWidth(), glyphCol,
				"glyph col mismatch: glyph=%d, ListContentWidth=%d", glyphCol, lay.ListContentWidth())
		})
	}
}

// locateGlyphCol returns the rune-index of the first occurrence of glyph
// in line after stripping ANSI escape sequences. Returns -1 if not found.
func locateGlyphCol(t *testing.T, line string, glyph rune) int {
	t.Helper()
	plain := stripAnsi(line)
	for i, r := range []rune(plain) {
		if r == glyph {
			return i
		}
	}
	return -1
}

// stripAnsi removes ANSI escape sequences from s. The state machine
// recognises the two-step CSI form `ESC [ <params/intermediates>
// <final>` where the final byte is in the ECMA-48 range 0x40..0x7e
// (`@`..`~`) — see F-134. A hardcoded terminator subset is insufficient
// because future styling layers may emit non-SGR CSI sequences (cursor
// positioning, mode setting, function-key codes) that would otherwise
// leak escape bytes into the stripped output and corrupt
// locateGlyphCol's column index. Non-CSI escape forms (`ESC <byte>`)
// are treated as two-byte sequences and discarded as a unit.
func stripAnsi(s string) string {
	b := strings.Builder{}
	const (
		stPlain   = 0
		stPostEsc = 1 // saw ESC; next byte is either `[` (→ CSI) or a single-byte escape final
		stCsiBody = 2 // saw ESC [; consume params/intermediates until final 0x40..0x7e
	)
	state := stPlain
	for _, r := range s {
		switch state {
		case stPostEsc:
			if r == '[' {
				state = stCsiBody
			} else {
				state = stPlain
			}
		case stCsiBody:
			if r >= 0x40 && r <= 0x7e {
				state = stPlain
			}
		default:
			if r == 0x1b {
				state = stPostEsc
				continue
			}
			b.WriteRune(r)
		}
	}
	return b.String()
}

// T-174: drag-seam distinctness + rendered-cell SGR test (Tier 23
// kit revision ed91d17). Per theme × orientation: render the seam
// cell and assert its SGR carries theme.DragHandle's code — never
// theme.DividerColor's code (the unfocused-border colour) and never
// theme.FocusBorder's (the focused-border / focus-cue colour).
// Also includes the data-only distinctness check from config R4
// AC 10 (DragHandle differs from both neighbours in every theme).
//
// Right-split reuses the T-160 renderer-truth locator: stub pane
// strings so the inline `│` is the only vertical-bar glyph in the
// rendered output, then find the glyph column on a mid-row and scan
// backwards for the opening CSI introducer.
//
// Below-mode exercises the PaneStyle + WithDragSeamTop composition
// directly (avoids a circular detailpane → appshell import) — the
// PaneModel wraps exactly this style in its View(), so asserting on
// the raw style output equals asserting on the rendered pane's top
// row.
func TestDragSeam_RendersInDragHandle_AllThemes(t *testing.T) {
	for _, themeName := range theme.BuiltinNames() {
		th := theme.GetTheme(themeName)

		// Data-only: config R4 AC 10 — DragHandle distinct from both
		// neighbours so no theme collapses into an ambiguous cue.
		require.NotEmptyf(t, string(th.DragHandle), "[%s] DragHandle empty", themeName)
		assert.NotEqualf(t, string(th.DividerColor), string(th.DragHandle),
			"[%s] DragHandle == DividerColor (%s)", themeName, th.DragHandle)
		assert.NotEqualf(t, string(th.FocusBorder), string(th.DragHandle),
			"[%s] DragHandle == FocusBorder (%s)", themeName, th.DragHandle)

		drag := colorANSI(th.DragHandle)
		div := colorANSI(th.DividerColor)
		focus := colorANSI(th.FocusBorder)
		require.NotEmptyf(t, drag, "[%s] empty DragHandle SGR probe — check TrueColor profile", themeName)
		require.NotEmptyf(t, div, "[%s] empty DividerColor SGR probe — check TrueColor profile", themeName)
		require.NotEmptyf(t, focus, "[%s] empty FocusBorder SGR probe — check TrueColor profile", themeName)

		t.Run(themeName+"/right", func(t *testing.T) {
			lm := NewLayoutModel(120, 30).
				WithTheme(th).
				SetDetailPane(true, 10).
				SetOrientation(OrientationRight).
				SetWidthRatio(0.30)
			lay := lm.Layout()
			stubList := strings.Repeat("L", lay.ListContentWidth())
			stubDetail := strings.Repeat("D", lay.DetailContentWidth()+2)
			out := lm.Render("H", stubList, stubDetail, "S")
			lines := strings.Split(out, "\n")
			midRow := len(lines) / 2
			glyphCol := locateGlyphCol(t, lines[midRow], '│')
			require.GreaterOrEqualf(t, glyphCol, 0, "no `│` on mid-row %d: %q", midRow, lines[midRow])
			sgr := sgrBeforeGlyph(t, lines[midRow], glyphCol, '│')
			assert.Containsf(t, sgr, drag, "right divider SGR missing DragHandle %q; got %q", drag, sgr)
			assert.NotContainsf(t, sgr, div, "right divider SGR still carries DividerColor %q; got %q", div, sgr)
			assert.NotContainsf(t, sgr, focus,
				"right divider SGR carries FocusBorder %q (drag-seam must be focus-neutral); got %q", focus, sgr)
		})

		t.Run(themeName+"/below", func(t *testing.T) {
			// Below-mode detail pane seam = the top border row of the
			// pane style with WithDragSeamTop applied. Exercise both
			// focus states since WithDragSeamTop must override either
			// base colour without leaking off the top edge.
			for _, state := range []PaneVisualState{PaneStateFocused, PaneStateUnfocused} {
				base := PaneStyle(th, state).Width(10)
				seam := WithDragSeamTop(base, th).Render("body")
				lines := strings.Split(seam, "\n")
				require.GreaterOrEqualf(t, len(lines), 3, "rendered pane has <3 rows: %q", seam)
				top := lines[0]
				assert.Containsf(t, top, drag, "below/%v top border missing DragHandle %q; got %q", state, drag, top)
				// Bottom border must still carry the focus-state
				// colour (bottom ≠ seam). Strictly check absence of
				// DragHandle there.
				bottom := lines[len(lines)-1]
				assert.NotContainsf(t, bottom, drag,
					"below/%v bottom border carries DragHandle SGR %q; override leaked off the top edge: %q", state, drag, bottom)
			}
		})
	}
}

// sgrBeforeGlyph returns the CSI SGR sequence immediately preceding the
// first occurrence of `glyph` on (ANSI-stripped) column `col` in `line`.
// Uses the same three-state ANSI parse as stripAnsi so future styling
// layers that emit non-SGR CSI sequences (cursor positioning, etc.) do
// not skew the locator. Returns the last SGR encountered (`\x1b[...m`)
// before the rune at `col`, or the empty string if none was seen.
func sgrBeforeGlyph(t *testing.T, line string, col int, glyph rune) string {
	t.Helper()
	const (
		stPlain   = 0
		stPostEsc = 1
		stCsiBody = 2
	)
	state := stPlain
	plainCol := 0
	var current strings.Builder // accumulates the currently-open CSI sequence
	lastSGR := ""
	for _, r := range line {
		switch state {
		case stPostEsc:
			current.WriteRune(r)
			if r == '[' {
				state = stCsiBody
			} else {
				state = stPlain
				current.Reset()
			}
		case stCsiBody:
			current.WriteRune(r)
			if r >= 0x40 && r <= 0x7e {
				if r == 'm' {
					lastSGR = "\x1b" + current.String()
				}
				state = stPlain
				current.Reset()
			}
		default:
			if r == 0x1b {
				state = stPostEsc
				current.Reset()
				current.WriteRune(r)
				continue
			}
			if plainCol == col {
				if r == glyph {
					return lastSGR
				}
				// Column matched but not the glyph we expected —
				// advance to keep scanning.
			}
			plainCol++
		}
	}
	return lastSGR
}

// F-134: stripAnsi must handle the full ECMA-48 CSI final-byte range
// (0x40..0x7e), not a hardcoded subset. This guards locateGlyphCol —
// which the R15 line-198 renderer-truth assertion depends on — against
// silent corruption when styling layers emit non-SGR CSI sequences
// (cursor positioning, mode setting, function-key codes). Today
// lipgloss only emits SGR (`m`), so the bug is latent — but the next
// styling change could quietly skew glyph-column detection.
//
// Each case has the form `\x1b[<params><terminator>X` — if the
// terminator is unrecognised, the literal `X` gets swallowed and the
// stripped string is empty.
func TestStripAnsi_HandlesFullCSIFinalByteRange(t *testing.T) {
	cases := []struct {
		name string
		in   string
	}{
		{"HVP_f", "\x1b[10;5fX"},
		{"CHA_G", "\x1b[5GX"},
		{"DECTCEM_show_h", "\x1b[?25hX"},
		{"DECTCEM_hide_l", "\x1b[?25lX"},
		{"function_key_tilde", "\x1b[2~X"},
		{"DSR_n", "\x1b[6nX"},
		{"DA_c", "\x1b[0cX"},
		{"save_cursor_s", "\x1b[sX"},
		{"restore_cursor_u", "\x1b[uX"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := stripAnsi(tc.in)
			assert.Equalf(t, "X", got, "stripAnsi(%q) = %q; want %q (escape sequence leaked through)", tc.in, got, "X")
		})
	}
}
