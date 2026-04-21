package detailpane

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/istonikula/gloggy/internal/logsource"
	"github.com/istonikula/gloggy/internal/theme"
)

// test helpers (T-107).
func lipglossWidth(s string) int { return lipgloss.Width(s) }
func lipglossStyle(s string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000")).Render(s)
}

// bgColorANSI is the detailpane-local mirror of the entrylist helper:
// renders a bg probe and returns the opening SGR sequence. Package-local
// so the test doesn't depend on another test package's helpers.
func bgColorANSI(c lipgloss.Color) string {
	rendered := lipgloss.NewStyle().Background(c).Render("x")
	end := strings.Index(rendered, "x")
	if end <= 0 {
		return ""
	}
	return rendered[:end]
}

// bgSGRFor extracts the background SGR prefix that lipgloss emits for a
// given color under the current render profile. Keeps the tests robust
// against termenv's color-profile rounding (truecolor RGB can drift by 1
// when lipgloss routes through 256-color intermediates).
func bgSGRFor(c lipgloss.Color) string {
	rendered := lipgloss.NewStyle().Background(c).Render("x")
	// The SGR is everything from "48;" up to the closing "m" before the
	// character payload. Extract the "48;2;r;g;b" substring — the caller
	// just needs a reliable sentinel.
	i := strings.Index(rendered, "48;")
	if i < 0 {
		return ""
	}
	end := strings.Index(rendered[i:], "m")
	if end < 0 {
		return ""
	}
	return rendered[i : i+end]
}

// innerResetAfterBgReopen matches `\x1b[0m` IMMEDIATELY followed by a bg-
// opening SGR (`\x1b[48…m`) — the legitimate lipgloss text→padding boundary
// that does NOT visually break the bg. Used by the T-141 test to scrub out
// these benign transitions before checking for bare resets.
var innerResetAfterBgReopen = regexp.MustCompile(`\x1b\[0m\x1b\[48[;\d]*m`)

func init() {
	// Force TrueColor in tests so ANSI color codes are embedded in output.
	lipgloss.SetColorProfile(termenv.TrueColor)
}

// colorANSI renders a test string with the given color and extracts the
// foreground ANSI escape code (e.g. "38;2;115;218;202") from the output.
// This mirrors exactly what lipgloss/termenv will produce, avoiding any
// manual hex→int conversion that might differ by rounding.
func colorANSI(c lipgloss.Color) string {
	rendered := lipgloss.NewStyle().Foreground(c).Render("x")
	start := strings.Index(rendered, "\x1b[")
	if start == -1 {
		return string(c)
	}
	end := strings.Index(rendered[start:], "m")
	if end == -1 {
		return string(c)
	}
	return rendered[start+2 : start+end]
}

func jsonEntry() logsource.Entry {
	raw := []byte(`{"time":"2024-01-01T12:00:00Z","level":"INFO","msg":"hello world","count":42,"active":true,"data":null}`)
	return logsource.Entry{
		IsJSON:     true,
		Raw:        raw,
		LineNumber: 1,
		Time:       time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		Level:      "INFO",
		Msg:        "hello world",
		Extra: map[string]json.RawMessage{
			"count":  json.RawMessage(`42`),
			"active": json.RawMessage(`true`),
			"data":   json.RawMessage(`null`),
		},
	}
}

// T-035: R2.1 — JSONL entry renders as indented JSON
func TestRenderJSON_IsIndented(t *testing.T) {
	result := RenderJSON(jsonEntry(), theme.GetTheme("tokyo-night"), nil)
	assert.Contains(t, result, "{\n", "expected indented JSON with newlines")
	assert.Contains(t, result, "  ", "expected indentation spaces")
}

// T-035: R2.2 — All fields present including extra
func TestRenderJSON_AllFieldsPresent(t *testing.T) {
	result := RenderJSON(jsonEntry(), theme.GetTheme("tokyo-night"), nil)
	for _, field := range []string{"time", "level", "msg", "count", "active", "data"} {
		assert.Containsf(t, result, `"`+field+`"`, "expected field %q in output", field)
	}
}

// T-035: R2.3 — Key color in ANSI output
func TestRenderJSON_KeyColor(t *testing.T) {
	th := theme.GetTheme("tokyo-night")
	result := RenderJSON(jsonEntry(), th, nil)
	want := colorANSI(th.SyntaxKey)
	assert.Containsf(t, result, want, "expected SyntaxKey ANSI code %s in output", want)
}

// T-035: R2.4 — String value color
func TestRenderJSON_StringColor(t *testing.T) {
	th := theme.GetTheme("tokyo-night")
	result := RenderJSON(jsonEntry(), th, nil)
	want := colorANSI(th.SyntaxString)
	assert.Containsf(t, result, want, "expected SyntaxString ANSI code %s in output", want)
}

// T-035: R2.5 — Number value color
func TestRenderJSON_NumberColor(t *testing.T) {
	th := theme.GetTheme("tokyo-night")
	result := RenderJSON(jsonEntry(), th, nil)
	want := colorANSI(th.SyntaxNumber)
	assert.Containsf(t, result, want, "expected SyntaxNumber ANSI code %s in output", want)
}

// T-035: R2.6 — Boolean value color
func TestRenderJSON_BoolColor(t *testing.T) {
	th := theme.GetTheme("tokyo-night")
	result := RenderJSON(jsonEntry(), th, nil)
	want := colorANSI(th.SyntaxBoolean)
	assert.Containsf(t, result, want, "expected SyntaxBoolean ANSI code %s in output", want)
}

// T-035: R2.7 — Null value color
func TestRenderJSON_NullColor(t *testing.T) {
	th := theme.GetTheme("tokyo-night")
	result := RenderJSON(jsonEntry(), th, nil)
	want := colorANSI(th.SyntaxNull)
	assert.Containsf(t, result, want, "expected SyntaxNull ANSI code %s in output", want)
}

// T-035: R2.8 — Theme switch changes ANSI codes
func TestRenderJSON_ThemeSwitch(t *testing.T) {
	entry := jsonEntry()
	th1 := theme.GetTheme("tokyo-night")
	th2 := theme.GetTheme("catppuccin-mocha")
	out1 := RenderJSON(entry, th1, nil)
	out2 := RenderJSON(entry, th2, nil)

	if string(th1.SyntaxKey) == string(th2.SyntaxKey) {
		t.Skip("themes have identical key color, skipping switch test")
	}
	assert.NotEqual(t, out1, out2, "expected different output when switching themes")
	want := colorANSI(th2.SyntaxKey)
	assert.Containsf(t, out2, want, "catppuccin-mocha output missing its key ANSI code %s", want)
}

// T-035: hidden fields are omitted
func TestRenderJSON_HiddenFields(t *testing.T) {
	result := RenderJSON(jsonEntry(), theme.GetTheme("tokyo-night"), []string{"level", "count"})
	assert.NotContains(t, result, `"level"`, "hidden field 'level' should not appear in output")
	assert.NotContains(t, result, `"count"`, "hidden field 'count' should not appear in output")
	assert.Contains(t, result, `"msg"`, "non-hidden field 'msg' should appear in output")
}

// T-036: R3.1 — Non-JSON entry displays as plain raw text
func TestRenderRaw_PlainText(t *testing.T) {
	entry := logsource.Entry{
		IsJSON: false,
		Raw:    []byte("2024-01-01 ERROR could not connect to database"),
	}
	result := RenderRaw(entry)
	assert.Equal(t, "2024-01-01 ERROR could not connect to database", result)
}

// T-036: R3.2 — No JSON formatting applied to non-JSON entries
func TestRenderRaw_NoANSI(t *testing.T) {
	entry := logsource.Entry{
		IsJSON: false,
		Raw:    []byte("plain text {not json}"),
	}
	result := RenderRaw(entry)
	// Result must not contain ANSI escape sequences.
	assert.NotContains(t, result, "\x1b[", "expected no ANSI escape codes in raw output")
	assert.Equal(t, "plain text {not json}", result)
}

// Regression for F-203 (two-tone detail-pane bug, Tier 24 follow-up):
// `render.go` styles each JSON token (key, string, number, boolean, null)
// with `lipgloss.NewStyle().Foreground(...)` only — no Background. Every
// token ends in `\x1b[0m` which, when the outer `PaneStyle` paints
// `BaseBg`, punches a hole through the pane bg: the screenshot showed
// BaseBg up through the quoted key and terminal-default bg starting at
// the `:` separator.
//
// `PaneModel.View` now calls `appshell.RepaintBg` to re-assert BaseBg
// after every inner reset. This test pins the user-observed invariant
// directly: the `\x1b[0m` emitted after the `"time"` key must be
// immediately followed by the BaseBg reassert sequence so the subsequent
// `": "` separator and value inherit BaseBg.
func TestPaneModel_View_KeyTokenReset_FollowedByBaseBg(t *testing.T) {
	th := theme.GetTheme("tokyo-night")
	m := NewPaneModel(th, 20)
	m.Focused = true
	m = m.SetWidth(60).Open(logsource.Entry{
		IsJSON:     true,
		LineNumber: 1,
		Time:       time.Now(),
		Level:      "INFO",
		Msg:        "hello",
		Raw: []byte(`{"time":"2026-04-14T23:39:10.868Z",` +
			`"level":"INFO","msg":"hello","count":42,"active":true,"data":null}`),
	})
	out := m.View()

	baseBgOpen := bgColorANSI(th.BaseBg)
	require.NotEmpty(t, baseBgOpen, "empty BaseBg SGR probe — is TrueColor forced in init()?")

	// Find the first `"time"` key token — its trailing `\x1b[0m` is the one
	// that historically "punched out" the pane bg before the `": "` separator.
	idx := strings.Index(out, "\"time\"")
	require.GreaterOrEqual(t, idx, 0, "`\"time\"` key missing from view output — test premise invalid")
	tail := out[idx:]
	reset := strings.Index(tail, "\x1b[0m")
	require.GreaterOrEqual(t, reset, 0, "no reset after `\"time\"` — test premise invalid")
	after := tail[reset+len("\x1b[0m"):]
	if !strings.HasPrefix(after, baseBgOpen) {
		previewEnd := 60
		if previewEnd > len(after) {
			previewEnd = len(after)
		}
		require.Failf(t, "reset after `\"time\"` not followed by BaseBg reassert",
			"want %q\nsaw: %q", baseBgOpen, after[:previewEnd])
	}
}

// T-082: R1.5 — open pane View starts with a top border character.
func TestPaneModel_TopBorder(t *testing.T) {
	m := defaultPane(10).Open(testEntry())
	v := m.View()
	require.NotEmpty(t, v, "expected non-empty view")
	// NormalBorder top uses "─" characters.
	assert.Containsf(t, v, "─", "expected top border character '─' in view: %q", v)
}

// T-103: the detail pane top border renders in both orientations. The pane
// itself is orientation-agnostic — the layout composes it via either
// JoinVertical (below) or JoinHorizontal (right). The pane's first View()
// line must always be the top border row.
func TestPaneModel_TopBorder_InBothOrientationContexts(t *testing.T) {
	for _, focused := range []bool{true, false} {
		m := defaultPane(10).Open(testEntry())
		m.Focused = focused
		v := m.View()
		lines := strings.Split(v, "\n")
		require.GreaterOrEqualf(t, len(lines), 2, "focused=%v: expected at least 2 lines (top border + content), got %d", focused, len(lines))
		// The first line is the top border. Strip ANSI escapes by
		// scanning for the box-drawing horizontal glyph; lipgloss.Width
		// returns cell width regardless of escape sequences, so a top
		// border line cell-width must equal the rendered output width.
		assert.Truef(t, strings.ContainsRune(lines[0], '─'),
			"focused=%v: first line missing top border glyph '─': %q", focused, lines[0])
	}
}

// T-107: lipgloss.Width measures cell width correctly across emoji, CJK,
// and ANSI-styled text — verifying our chosen primitive is safe for the
// pane's width-aware code paths.
func TestPaneModel_LipglossWidth_HandlesEmojiCJKAnsi(t *testing.T) {
	for _, tc := range []struct {
		name string
		s    string
		want int
	}{
		{"emoji", "🔥", 2},
		{"cjk", "日本語", 6},
		{"ansi-wrapped ascii", lipglossStyle("X"), 1},
		{"mixed", "a🔥b", 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equalf(t, tc.want, lipglossWidth(tc.s), "lipgloss.Width(%q)", tc.s)
		})
	}
}

// T-131: cursor row keeps CursorHighlight bg in both focus states; focused
// combines Bold (SGR 1) with bg, unfocused combines Faint (SGR 2) with bg —
// never both.
func TestPaneModel_View_CursorHighlight_FocusAttr(t *testing.T) {
	th := theme.GetTheme("tokyo-night")
	bgSGR := bgSGRFor(th.CursorHighlight)
	require.NotEmpty(t, bgSGR, "could not synthesize CursorHighlight bg SGR")

	hasAttrBg := func(view string, attr int) bool {
		prefix := "\x1b[" + string(rune('0'+attr)) + ";48;"
		infix := ";" + string(rune('0'+attr)) + ";48;"
		return strings.Contains(view, prefix) || strings.Contains(view, infix)
	}

	for _, tc := range []struct {
		name    string
		focused bool
		wantBg  bool // always true; kept for clarity
		// Exactly one of wantBold/wantFaint is true; the other must be absent.
		wantBold, wantFaint bool
	}{
		{"focused bold", true, true, true, false},
		{"unfocused faint", false, true, false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := paneWithNLines(12, 30, 40)
			m.Focused = tc.focused
			view := m.View()
			assert.Containsf(t, view, bgSGR, "expected CursorHighlight bg SGR %q: %q", bgSGR, view)
			assert.Equalf(t, tc.wantBold, hasAttrBg(view, 1), "bold+bg presence: %q", view)
			assert.Equalf(t, tc.wantFaint, hasAttrBg(view, 2), "faint+bg presence: %q", view)
		})
	}
}

// T-131: cursor row paints at the correct visible position when offset > 0.
func TestPaneModel_View_CursorHighlight_HonorsOffset(t *testing.T) {
	th := theme.GetTheme("tokyo-night")
	m := paneWithNLines(10, 20, 50)
	m.scroll.cursor = 7
	m.scroll.offset = 3
	m.Focused = true
	view := m.View()
	bgSGR := bgSGRFor(th.CursorHighlight)
	assert.Containsf(t, view, bgSGR, "expected CursorHighlight bg in view with cursor=7, offset=3; view=%q", view)
	// visible row = cursor - offset = 4. Rendered content rows start at
	// index 1 (after top border row), so cursor visually lands at row 5.
	rows := strings.Split(view, "\n")
	require.GreaterOrEqualf(t, len(rows), 6, "expected >=6 rows, got %d", len(rows))
	// The cursor row must contain the CursorHighlight bg; other content rows
	// (index 1..4 and 6..contentH) must not.
	for i := 1; i <= 4; i++ {
		assert.NotContainsf(t, rows[i], bgSGR, "row %d should not have CursorHighlight bg; row=%q", i, rows[i])
	}
	assert.Containsf(t, rows[5], bgSGR, "row 5 (visible cursor row) should have CursorHighlight bg; row=%q", rows[5])
}

// T-131: closed pane renders nothing at all — no cursor paint possible.
func TestPaneModel_View_Closed_NoCursor(t *testing.T) {
	m := defaultPane(10)
	assert.Empty(t, m.View(), "closed pane should render empty string")
}

// T-141 (F-105): paintCursorRow strips inner `\x1b[0m` resets from the
// cursor row before applying the outer CursorHighlight bg.
func TestPaneModel_PaintCursorRow_T141_StripsInnerResets(t *testing.T) {
	m := defaultPane(10).SetWidth(40)
	m.open = true
	m.Focused = true
	m.scroll = NewScrollModel("", 4)
	m.scroll.cursor = 0
	m.scroll.offset = 0

	body := "\x1b[32m\"key\"\x1b[0m: \x1b[36m\"value\"\x1b[0m,"
	got := m.paintCursorRow(body, 4)

	require.Containsf(t, got, "\x1b[", "painted row missing any SGR: %q", got)
	// Drop trailing terminator.
	trimmed := strings.TrimSuffix(got, "\x1b[0m")
	// Remove legitimate "reset + bg-reopen" transitions from inside the
	// output; any leftover `\x1b[0m` is a bare reset that would visually
	// break the bg run (F-105).
	scrubbed := innerResetAfterBgReopen.ReplaceAllString(trimmed, "")
	assert.NotContainsf(t, scrubbed, "\x1b[0m", "painted cursor row contains a bare inner `\\x1b[0m` (F-105): %q", got)
	// Also assert the outer bold+bg opens exactly once — regression guard
	// for the strip logic. lipgloss's bold-bg opener starts with `\x1b[1;`.
	openers := strings.Count(got, "\x1b[1;")
	assert.Equalf(t, 1, openers, "expected 1 outer bold+bg opener, got %d: %q", openers, got)
}
