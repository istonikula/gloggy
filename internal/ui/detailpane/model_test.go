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
