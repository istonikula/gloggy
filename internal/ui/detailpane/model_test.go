package detailpane

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/istonikula/gloggy/internal/logsource"
	"github.com/istonikula/gloggy/internal/theme"
)

// test helpers (T-107).
func lipglossWidth(s string) int { return lipgloss.Width(s) }
func lipglossStyle(s string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000")).Render(s)
}

func testEntry() logsource.Entry {
	return logsource.Entry{
		IsJSON:     true,
		LineNumber: 1,
		Time:       time.Now(),
		Level:      "INFO",
		Msg:        "hello",
		Raw:        []byte(`{"level":"INFO","msg":"hello"}`),
	}
}

func defaultPane(height int) PaneModel {
	return NewPaneModel(theme.GetTheme("tokyo-night"), height)
}

// T-041: R1.1 — Enter on entry opens detail pane (caller opens via Open(); here we test Open sets state).
func TestPaneModel_Open_SetsOpen(t *testing.T) {
	m := defaultPane(10).Open(testEntry())
	if !m.IsOpen() {
		t.Error("expected pane to be open after Open()")
	}
}

// T-041: R1.2 — Double-click handled by ListModel; PaneModel.Open() is the activation path.
// Just verify Open() renders non-empty content.
func TestPaneModel_Open_RendersContent(t *testing.T) {
	m := defaultPane(10).Open(testEntry())
	v := m.View()
	if v == "" {
		t.Error("expected non-empty view after Open()")
	}
}

// T-041: R1.3 — Esc closes pane and emits BlurredMsg.
func TestPaneModel_Esc_ClosesAndEmitsBlurred(t *testing.T) {
	m := defaultPane(10).Open(testEntry())
	m2, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m2.IsOpen() {
		t.Error("expected pane to be closed after Esc")
	}
	if cmd == nil {
		t.Fatal("expected BlurredMsg cmd")
	}
	msg := cmd()
	if _, ok := msg.(BlurredMsg); !ok {
		t.Errorf("expected BlurredMsg, got %T", msg)
	}
}

// T-041: R1.4 — Enter in pane closes and emits BlurredMsg.
func TestPaneModel_Enter_ClosesAndEmitsBlurred(t *testing.T) {
	m := defaultPane(10).Open(testEntry())
	m2, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m2.IsOpen() {
		t.Error("expected pane to be closed after Enter")
	}
	if cmd == nil {
		t.Fatal("expected BlurredMsg cmd")
	}
	msg := cmd()
	if _, ok := msg.(BlurredMsg); !ok {
		t.Errorf("expected BlurredMsg, got %T", msg)
	}
}

// When pane is closed, Update is a no-op.
func TestPaneModel_Closed_UpdateNoop(t *testing.T) {
	m := defaultPane(10)
	m2, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if m2.IsOpen() {
		t.Error("should remain closed")
	}
	if cmd != nil {
		t.Error("expected nil cmd when pane is closed")
	}
}

// View returns empty string when closed.
func TestPaneModel_Closed_ViewEmpty(t *testing.T) {
	m := defaultPane(10)
	if m.View() != "" {
		t.Error("expected empty view when pane is closed")
	}
}

// T-082: R1.5 — open pane View starts with a top border character.
func TestPaneModel_TopBorder(t *testing.T) {
	m := defaultPane(10).Open(testEntry())
	v := m.View()
	if len(v) == 0 {
		t.Fatal("expected non-empty view")
	}
	// NormalBorder top uses "─" characters.
	if !strings.Contains(v, "─") {
		t.Errorf("expected top border character '─' in view: %q", v)
	}
}

// T-100: focused vs unfocused panes use the DESIGN.md §4 matrix —
// borders render in BOTH states, only the color differs (FocusBorder vs
// DividerColor). Vertical bar count therefore matches; the discriminator
// is the rendered ANSI color of the border foreground.
func TestPaneModel_Focused_VsUnfocused_DifferentBorderColor(t *testing.T) {
	m := defaultPane(10).Open(testEntry())
	m.Focused = true
	focused := m.View()
	m.Focused = false
	unfocused := m.View()
	if strings.Count(focused, "│") == 0 {
		t.Errorf("focused pane should render vertical border: %q", focused)
	}
	if strings.Count(unfocused, "│") == 0 {
		t.Errorf("unfocused pane should render vertical border: %q", unfocused)
	}
	if focused == unfocused {
		t.Errorf("focused and unfocused outputs must differ (border color): %q", focused)
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
			if got := lipglossWidth(tc.s); got != tc.want {
				t.Errorf("lipgloss.Width(%q) = %d, want %d", tc.s, got, tc.want)
			}
		})
	}
}

// T-107: pane outer width equals the allocated width — emoji/CJK content
// must not push the pane past its budget.
func TestPaneModel_View_OuterWidth_MatchesAllocation(t *testing.T) {
	const allocated = 24
	entry := logsource.Entry{
		IsJSON: true,
		Time:   time.Now(),
		Level:  "INFO",
		Msg:    "🔥 fire — 日本語 — long enough to overflow naive budgets",
		Raw:    []byte(`{"msg":"🔥 fire 日本語"}`),
	}
	m := defaultPane(8).Open(entry).SetWidth(allocated)
	v := m.View()
	if v == "" {
		t.Fatal("expected non-empty view")
	}
	for i, line := range strings.Split(v, "\n") {
		w := lipglossWidth(line)
		if w > allocated {
			t.Errorf("line %d width=%d exceeds allocated=%d: %q", i, w, allocated, line)
		}
	}
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
		if len(lines) < 2 {
			t.Fatalf("focused=%v: expected at least 2 lines (top border + content), got %d", focused, len(lines))
		}
		// The first line is the top border. Strip ANSI escapes by
		// scanning for the box-drawing horizontal glyph; lipgloss.Width
		// returns cell width regardless of escape sequences, so a top
		// border line cell-width must equal the rendered output width.
		if !strings.ContainsRune(lines[0], '─') {
			t.Errorf("focused=%v: first line missing top border glyph '─': %q", focused, lines[0])
		}
	}
}

// T-113 (cavekit-detail-pane R7 / F-003): ContentLines() returns the
// pre-render soft-wrapped content — no border glyphs, no ANSI escapes.
// This is the authoritative match source for in-pane search.
func TestPaneModel_ContentLines_NoBordersNoANSI(t *testing.T) {
	m := defaultPane(10).SetWidth(40).Open(testEntry())
	lines := m.ContentLines()
	if len(lines) == 0 {
		t.Fatal("expected non-empty ContentLines after Open()")
	}
	borderGlyphs := []string{"│", "─", "╭", "╮", "╰", "╯", "┌", "┐", "└", "┘"}
	for i, line := range lines {
		if strings.Contains(line, "\x1b[") {
			t.Errorf("line %d contains ANSI escape: %q", i, line)
		}
		for _, g := range borderGlyphs {
			if strings.Contains(line, g) {
				t.Errorf("line %d contains border glyph %q: %q", i, g, line)
			}
		}
	}
}

// T-113: ContentLines() returns nil when pane is closed or has no content.
func TestPaneModel_ContentLines_ClosedReturnsNil(t *testing.T) {
	m := defaultPane(10)
	if got := m.ContentLines(); got != nil {
		t.Errorf("expected nil from closed pane, got %v", got)
	}
	m = m.Open(testEntry()).Close()
	if got := m.ContentLines(); got != nil {
		t.Errorf("expected nil from closed-after-open pane, got %v", got)
	}
}

// T-114 (F-002, F-004, F-010): an active SearchModel attached via
// WithSearch() renders a visible prompt row and (cur/total) counter in
// the pane's View output.
func TestPaneModel_WithSearch_RendersPromptAndCounter(t *testing.T) {
	pane := defaultPane(12).SetWidth(60).Open(logsource.Entry{
		IsJSON: true,
		Time:   time.Now(),
		Level:  "INFO",
		Msg:    "hello",
		Raw:    []byte(`{"level":"INFO","msg":"hello world","tag":"hello there"}`),
	})
	search := NewSearchModel(theme.GetTheme("tokyo-night")).Activate().SetQuery("hello", pane.ContentLines())
	if search.MatchCount() < 1 {
		t.Fatalf("setup: expected >=1 match, got %d", search.MatchCount())
	}
	view := pane.WithSearch(search).View()
	if !strings.Contains(view, "/hello") {
		t.Errorf("view missing active query prompt '/hello': %q", view)
	}
	// Counter "(1/N)" — N depends on match count but we check the format.
	if !strings.Contains(view, "(1/") {
		t.Errorf("view missing (cur/total) counter: %q", view)
	}
}

// T-114: non-empty query with zero matches renders "No matches".
func TestPaneModel_WithSearch_NoMatchesIndicator(t *testing.T) {
	pane := defaultPane(12).SetWidth(60).Open(testEntry())
	search := NewSearchModel(theme.GetTheme("tokyo-night")).Activate().SetQuery("zzz-nope", pane.ContentLines())
	if search.MatchCount() != 0 {
		t.Fatalf("setup: expected 0 matches, got %d", search.MatchCount())
	}
	view := pane.WithSearch(search).View()
	if !strings.Contains(view, "No matches") {
		t.Errorf("view should show 'No matches' for query with zero hits: %q", view)
	}
}

// T-114: bare `/` prompt (no query yet) renders just the slash so the
// user sees that search is running.
func TestPaneModel_WithSearch_BarePrompt(t *testing.T) {
	pane := defaultPane(12).SetWidth(60).Open(testEntry())
	search := NewSearchModel(theme.GetTheme("tokyo-night")).Activate()
	view := pane.WithSearch(search).View()
	if !strings.Contains(view, "/") {
		t.Errorf("view should contain '/' prompt while search is active: %q", view)
	}
}

// T-114: when search is inactive, View() does not render a prompt row —
// the pane reverts to its normal rendering.
func TestPaneModel_WithSearch_InactiveNoPrompt(t *testing.T) {
	pane := defaultPane(12).SetWidth(60).Open(testEntry())
	search := NewSearchModel(theme.GetTheme("tokyo-night")) // not activated
	view := pane.WithSearch(search).View()
	if strings.Contains(view, "No matches") {
		t.Errorf("inactive search should not render 'No matches': %q", view)
	}
}

// T-115 (F-005): ScrollToLine brings an out-of-window line into view.
// Uses many-line content so the viewport can actually be smaller than
// the line count.
func TestPaneModel_ScrollToLine_BringsMatchIntoView(t *testing.T) {
	// Build an entry whose rendered JSON will have plenty of fields so
	// the wrapped content exceeds the pane height.
	raw := `{"a":"1","b":"2","c":"3","d":"4","e":"5","f":"6","g":"7","h":"8","i":"9","j":"10","target":"find-me"}`
	entry := logsource.Entry{IsJSON: true, Time: time.Now(), Level: "INFO", Msg: "x", Raw: []byte(raw)}
	pane := defaultPane(6).SetWidth(40).Open(entry) // content height ~4
	lines := pane.ContentLines()
	// Locate the line containing "target".
	targetIdx := -1
	for i, l := range lines {
		if strings.Contains(l, "target") {
			targetIdx = i
			break
		}
	}
	if targetIdx < 0 {
		t.Fatalf("setup: no 'target' in content: %v", lines)
	}
	if targetIdx <= pane.ContentHeight()-1 {
		t.Skipf("target at idx=%d already in initial viewport (content height %d); test not applicable", targetIdx, pane.ContentHeight())
	}
	scrolled := pane.ScrollToLine(targetIdx)
	view := scrolled.View()
	if !strings.Contains(view, "target") {
		t.Errorf("after ScrollToLine(%d), view should contain 'target': %q", targetIdx, view)
	}
}

// T-134 (F-026, cavekit R11): ScrollToLine moves the cursor AND scrolls so
// the cursor has scrolloff context. 100-line doc, viewport=10, scrolloff=5,
// target line 40 → cursor=40; offset slides to put cursor in viewport with
// 5 rows of margin. Since followCursor uses offset = cursor - viewport + 1
// + scrolloff when jumping down, expect offset = 40 - 10 + 1 + 5 = 36. The
// visible cursor-row is then 40 - 36 = row 4 (0-indexed) — that is
// (height - 1 - scrolloff) = 10 - 1 - 5 = 4, exactly at the bottom-margin
// boundary so subsequent `j` presses would shift the viewport.
func TestPaneModel_ScrollToLine_MovesCursorWithScrolloffContext(t *testing.T) {
	m := defaultPane(12).SetWidth(40) // ContentHeight = 10
	m = m.WithScrolloff(5)
	m.open = true
	lines := make([]string, 100)
	for i := range lines {
		lines[i] = "ln"
	}
	m.rawContent = strings.Join(lines, "\n")
	m.scroll = NewScrollModel(m.rawContent, m.ContentHeight()).WithScrolloff(5)

	m = m.ScrollToLine(40)
	if m.scroll.Cursor() != 40 {
		t.Errorf("cursor = %d, want 40", m.scroll.Cursor())
	}
	if m.scroll.Offset() != 36 {
		t.Errorf("offset = %d, want 36 (cursor at bottom-margin)", m.scroll.Offset())
	}
}

// T-134: cursor-row render still has CursorHighlight bg when search is
// active — the bg is the last paint in View() so it composes on top of
// SearchHighlight fg.
func TestPaneModel_View_CursorBgOverSearchActive(t *testing.T) {
	th := theme.GetTheme("tokyo-night")
	m := defaultPane(12).SetWidth(40)
	m = m.WithScrolloff(5)
	m.open = true
	lines := make([]string, 100)
	for i := range lines {
		lines[i] = "ln"
	}
	m.rawContent = strings.Join(lines, "\n")
	m.scroll = NewScrollModel(m.rawContent, m.ContentHeight()).WithScrolloff(5)
	m.Focused = true

	m = m.ScrollToLine(40)
	view := m.View()
	bgSGR := bgSGRFor(th.CursorHighlight)
	if !strings.Contains(view, bgSGR) {
		t.Errorf("cursor-row bg missing after ScrollToLine; view=%q", view)
	}
}

// T-125 (F-016): scroll indicator reports a percentage when content
// exceeds the viewport.
func TestPaneModel_ScrollPercent_ExceedingViewport(t *testing.T) {
	m := defaultPane(22) // ContentHeight = 20
	// Seed scroll with 200 lines of known content.
	lines := make([]string, 200)
	for i := range lines {
		lines[i] = "line"
	}
	m.open = true
	m.rawContent = strings.Join(lines, "\n")
	m.scroll = NewScrollModel(m.rawContent, m.ContentHeight())
	// offset 0, height 20, total 200 → (0+20)/200 = 10%
	if got := m.ScrollPercent(); got != 10 {
		t.Errorf("ScrollPercent at offset=0: got %d, want 10", got)
	}
}

// T-125: at the bottom of a long document the indicator reads 100%.
func TestPaneModel_ScrollPercent_AtBottom(t *testing.T) {
	m := defaultPane(22)
	lines := make([]string, 200)
	for i := range lines {
		lines[i] = "line"
	}
	m.open = true
	m.rawContent = strings.Join(lines, "\n")
	m.scroll = NewScrollModel(m.rawContent, m.ContentHeight())
	m.scroll.offset = 180 // max offset = 200-20
	if got := m.ScrollPercent(); got != 100 {
		t.Errorf("ScrollPercent at bottom: got %d, want 100", got)
	}
}

// T-125: content that fits entirely → indicator omitted (sentinel -1).
func TestPaneModel_ScrollPercent_FitsViewport(t *testing.T) {
	m := defaultPane(22)
	lines := []string{"only", "a", "few", "lines"}
	m.open = true
	m.rawContent = strings.Join(lines, "\n")
	m.scroll = NewScrollModel(m.rawContent, m.ContentHeight())
	if got := m.ScrollPercent(); got != -1 {
		t.Errorf("ScrollPercent with short content: got %d, want -1", got)
	}
}

// T-125: View overlays the indicator on the last body line (dim text).
func TestPaneModel_View_IncludesScrollIndicator(t *testing.T) {
	m := defaultPane(22).SetWidth(30)
	lines := make([]string, 200)
	for i := range lines {
		lines[i] = "ln"
	}
	m.open = true
	m.rawContent = strings.Join(lines, "\n")
	m.scroll = NewScrollModel(m.rawContent, m.ContentHeight())
	view := m.View()
	if !strings.Contains(view, "10%") {
		t.Errorf("view should contain \"10%%\" indicator, got: %q", view)
	}
}

// T-125: View omits the indicator when content fits (no "0%" noise).
func TestPaneModel_View_OmitsIndicatorOnShortContent(t *testing.T) {
	m := defaultPane(22).SetWidth(30).Open(testEntry())
	view := m.View()
	if strings.Contains(view, "%") {
		t.Errorf("short-content view should not render a percentage indicator, got: %q", view)
	}
}

// T-127 (F-020): hidden fields set via WithHiddenFields reach the JSON
// renderer through Open — the suppressed key must not appear in rawContent.
func TestPaneModel_Open_HonorsHiddenFields(t *testing.T) {
	entry := logsource.Entry{
		IsJSON:     true,
		LineNumber: 1,
		Time:       time.Now(),
		Level:      "INFO",
		Msg:        "hello",
		Raw:        []byte(`{"level":"INFO","msg":"hello","secret":"hunter2"}`),
	}
	pane := defaultPane(20).WithHiddenFields([]string{"secret"}).Open(entry)
	if strings.Contains(pane.rawContent, "secret") {
		t.Errorf("raw content should not include suppressed field `secret`, got: %q", pane.rawContent)
	}
	if strings.Contains(pane.rawContent, "hunter2") {
		t.Errorf("raw content should not include suppressed value, got: %q", pane.rawContent)
	}
	if !strings.Contains(pane.rawContent, "hello") {
		t.Errorf("raw content should still include non-suppressed fields, got: %q", pane.rawContent)
	}
}

// T-127 (F-020): Rerender with an updated hiddenFields set re-renders
// the current entry without the newly suppressed field.
func TestPaneModel_Rerender_RemovesNewlyHiddenField(t *testing.T) {
	entry := logsource.Entry{
		IsJSON:     true,
		LineNumber: 1,
		Time:       time.Now(),
		Level:      "INFO",
		Msg:        "hello",
		Raw:        []byte(`{"level":"INFO","msg":"hello","secret":"hunter2"}`),
	}
	pane := defaultPane(20).Open(entry)
	if !strings.Contains(pane.rawContent, "secret") {
		t.Fatalf("precondition: raw content should include `secret` before hide, got: %q", pane.rawContent)
	}
	pane = pane.WithHiddenFields([]string{"secret"}).Rerender()
	if strings.Contains(pane.rawContent, "secret") {
		t.Errorf("raw content should not include suppressed field after Rerender, got: %q", pane.rawContent)
	}
}

// T-127 (F-020): Rerender preserves scroll offset so toggling a field's
// visibility does not jump the viewport back to the top.
func TestPaneModel_Rerender_PreservesScrollOffset(t *testing.T) {
	// Build an entry with enough fields to guarantee scrolling room.
	raw := `{"a":"1","b":"2","c":"3","d":"4","e":"5","f":"6","g":"7","h":"8","i":"9","j":"10","k":"11","l":"12"}`
	entry := logsource.Entry{IsJSON: true, Time: time.Now(), Level: "INFO", Msg: "x", Raw: []byte(raw)}
	pane := defaultPane(6).SetWidth(40).Open(entry)
	// Scroll down a few lines so we can detect offset preservation.
	pane.scroll.offset = 3
	pane.scroll = pane.scroll.Clamp()
	offBefore := pane.scroll.offset
	if offBefore == 0 {
		t.Skipf("content is too short to scroll; test not applicable")
	}

	pane = pane.WithHiddenFields([]string{"a"}).Rerender()
	offAfter := pane.scroll.offset
	if offAfter == 0 {
		t.Errorf("Rerender jumped to top; expected to preserve offset ~%d, got 0", offBefore)
	}
}

// T-127: Rerender on a closed pane is a safe no-op.
func TestPaneModel_Rerender_ClosedPaneNoOp(t *testing.T) {
	pane := defaultPane(20)
	out := pane.Rerender()
	if out.IsOpen() {
		t.Error("Rerender on closed pane must not open it")
	}
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

// T-131: cursor row renders with CursorHighlight bg + Bold when focused.
func TestPaneModel_View_CursorHighlight_FocusedBold(t *testing.T) {
	th := theme.GetTheme("tokyo-night")
	m := defaultPane(12).SetWidth(30)
	m.open = true
	lines := make([]string, 40)
	for i := range lines {
		lines[i] = "ln"
	}
	m.rawContent = strings.Join(lines, "\n")
	m.scroll = NewScrollModel(m.rawContent, m.ContentHeight())
	m.Focused = true
	view := m.View()
	bgSGR := bgSGRFor(th.CursorHighlight)
	if bgSGR == "" {
		t.Fatal("could not synthesize CursorHighlight bg SGR")
	}
	if !strings.Contains(view, bgSGR) {
		t.Errorf("expected CursorHighlight bg SGR %q in focused view; view=%q", bgSGR, view)
	}
	// Focused cursor row should be bold — lipgloss emits SGR 1 combined
	// with the bg attribute, e.g. `\x1b[1;48;2;...m`.
	if !strings.Contains(view, "\x1b[1;48;") && !strings.Contains(view, ";1;48;") {
		t.Errorf("expected Bold+bg combined SGR on focused cursor row; view=%q", view)
	}
}

// T-131: cursor row keeps CursorHighlight bg when pane is unfocused but
// uses Faint instead of Bold.
func TestPaneModel_View_CursorHighlight_UnfocusedNoBold(t *testing.T) {
	th := theme.GetTheme("tokyo-night")
	m := defaultPane(12).SetWidth(30)
	m.open = true
	lines := make([]string, 40)
	for i := range lines {
		lines[i] = "ln"
	}
	m.rawContent = strings.Join(lines, "\n")
	m.scroll = NewScrollModel(m.rawContent, m.ContentHeight())
	m.Focused = false
	view := m.View()
	bgSGR := bgSGRFor(th.CursorHighlight)
	if !strings.Contains(view, bgSGR) {
		t.Errorf("expected CursorHighlight bg SGR %q in unfocused view; view=%q", bgSGR, view)
	}
	// Unfocused cursor row combines Faint (SGR 2) with bg.
	if !strings.Contains(view, "\x1b[2;48;") && !strings.Contains(view, ";2;48;") {
		t.Errorf("expected Faint+bg combined SGR on unfocused cursor row; view=%q", view)
	}
	// And no Bold on the cursor row.
	if strings.Contains(view, "\x1b[1;48;") || strings.Contains(view, ";1;48;") {
		t.Errorf("expected no Bold SGR on unfocused cursor row; view=%q", view)
	}
}

// T-131: cursor row paints at the correct visible position when offset > 0.
func TestPaneModel_View_CursorHighlight_HonorsOffset(t *testing.T) {
	th := theme.GetTheme("tokyo-night")
	m := defaultPane(10).SetWidth(20)
	m.open = true
	lines := make([]string, 50)
	for i := range lines {
		lines[i] = "ln"
	}
	m.rawContent = strings.Join(lines, "\n")
	m.scroll = NewScrollModel(m.rawContent, m.ContentHeight())
	m.scroll.cursor = 7
	m.scroll.offset = 3
	m.Focused = true
	view := m.View()
	bgSGR := bgSGRFor(th.CursorHighlight)
	if !strings.Contains(view, bgSGR) {
		t.Errorf("expected CursorHighlight bg in view with cursor=7, offset=3; view=%q", view)
	}
	// visible row = cursor - offset = 4. Rendered content rows start at
	// index 1 (after top border row), so cursor visually lands at row 5.
	rows := strings.Split(view, "\n")
	if len(rows) < 6 {
		t.Fatalf("expected >=6 rows, got %d", len(rows))
	}
	// The cursor row must contain the CursorHighlight bg; other content rows
	// (index 1..4 and 6..contentH) must not.
	for i := 1; i <= 4; i++ {
		if strings.Contains(rows[i], bgSGR) {
			t.Errorf("row %d should not have CursorHighlight bg; row=%q", i, rows[i])
		}
	}
	if !strings.Contains(rows[5], bgSGR) {
		t.Errorf("row 5 (visible cursor row) should have CursorHighlight bg; row=%q", rows[5])
	}
}

// T-131: closed pane renders nothing at all — no cursor paint possible.
func TestPaneModel_View_Closed_NoCursor(t *testing.T) {
	m := defaultPane(10)
	if m.View() != "" {
		t.Error("closed pane should render empty string")
	}
}

// T-131: SetContent resets cursor to 0.
func TestScrollModel_SetContent_ResetsCursor(t *testing.T) {
	m := NewScrollModel("a\nb\nc\nd\ne\nf", 3)
	m.cursor = 4
	m.offset = 2
	m = m.SetContent("x\ny\nz", 3)
	if m.Cursor() != 0 {
		t.Errorf("SetContent should reset cursor to 0, got %d", m.Cursor())
	}
	if m.Offset() != 0 {
		t.Errorf("SetContent should reset offset to 0, got %d", m.Offset())
	}
}

// T-125: Overlay preserves overall pane width — indicator does NOT add columns.
func TestPaneModel_View_IndicatorDoesNotExpandWidth(t *testing.T) {
	m := defaultPane(22).SetWidth(30)
	lines := make([]string, 200)
	for i := range lines {
		lines[i] = "ln"
	}
	m.open = true
	m.rawContent = strings.Join(lines, "\n")
	m.scroll = NewScrollModel(m.rawContent, m.ContentHeight())
	view := m.View()
	// Inspect each row's cell width; all should equal outer pane width.
	for i, row := range strings.Split(view, "\n") {
		if w := lipgloss.Width(row); w != 30 {
			t.Errorf("row %d cell width: got %d, want 30; row=%q", i, w, row)
		}
	}
}
