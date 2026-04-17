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
