package detailpane

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/istonikula/gloggy/internal/logsource"
	"github.com/istonikula/gloggy/internal/theme"
)

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

// paneWithNLines opens a pane at (height, width) pre-populated with n copies
// of "ln" as rawContent. Collapses the hand-rolled setup that otherwise
// repeats across scroll/cursor-highlight/indicator tests.
func paneWithNLines(height, width, n int) PaneModel {
	m := defaultPane(height).SetWidth(width)
	m.open = true
	lines := make([]string, n)
	for i := range lines {
		lines[i] = "ln"
	}
	m.rawContent = strings.Join(lines, "\n")
	m.scroll = NewScrollModel(m.rawContent, m.ContentHeight())
	return m
}

// T-041: R1.1 — Enter on entry opens detail pane (caller opens via Open(); here we test Open sets state).
func TestPaneModel_Open_SetsOpen(t *testing.T) {
	m := defaultPane(10).Open(testEntry())
	assert.True(t, m.IsOpen(), "expected pane to be open after Open()")
}

// T-041: R1.2 — Double-click handled by ListModel; PaneModel.Open() is the activation path.
// Just verify Open() renders non-empty content.
func TestPaneModel_Open_RendersContent(t *testing.T) {
	m := defaultPane(10).Open(testEntry())
	assert.NotEmpty(t, m.View(), "expected non-empty view after Open()")
}

// T-041: R1.3/R1.4 — Esc and Enter both close the pane and emit BlurredMsg.
func TestPaneModel_DismissKey_ClosesAndEmitsBlurred(t *testing.T) {
	for _, tc := range []struct {
		name    string
		keyType tea.KeyType
	}{
		{"Esc", tea.KeyEsc},
		{"Enter", tea.KeyEnter},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := defaultPane(10).Open(testEntry())
			m2, cmd := m.Update(tea.KeyMsg{Type: tc.keyType})
			assert.False(t, m2.IsOpen(), "expected pane to be closed after %s", tc.name)
			require.NotNil(t, cmd, "expected BlurredMsg cmd")
			msg := cmd()
			_, ok := msg.(BlurredMsg)
			assert.Truef(t, ok, "expected BlurredMsg, got %T", msg)
		})
	}
}

// When pane is closed, Update is a no-op.
func TestPaneModel_Closed_UpdateNoop(t *testing.T) {
	m := defaultPane(10)
	m2, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	assert.False(t, m2.IsOpen(), "should remain closed")
	assert.Nil(t, cmd, "expected nil cmd when pane is closed")
}

// View returns empty string when closed.
func TestPaneModel_Closed_ViewEmpty(t *testing.T) {
	m := defaultPane(10)
	assert.Empty(t, m.View(), "expected empty view when pane is closed")
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
	assert.Greaterf(t, strings.Count(focused, "│"), 0, "focused pane should render vertical border: %q", focused)
	assert.Greaterf(t, strings.Count(unfocused, "│"), 0, "unfocused pane should render vertical border: %q", unfocused)
	assert.NotEqualf(t, unfocused, focused, "focused and unfocused outputs must differ (border color): %q", focused)
}

// T-107 / T-139 (F-103): pane outer width equals CONTENT width + 2 border
// cells — emoji/CJK content must not push the pane past its budget. With
// single-owner border accounting, `SetWidth(n)` receives CONTENT width and
// the outer rendered block is exactly n + 2 cells wide.
func TestPaneModel_View_OuterWidth_MatchesAllocation(t *testing.T) {
	const contentW = 24
	const outerW = contentW + 2
	entry := logsource.Entry{
		IsJSON: true,
		Time:   time.Now(),
		Level:  "INFO",
		Msg:    "🔥 fire — 日本語 — long enough to overflow naive budgets",
		Raw:    []byte(`{"msg":"🔥 fire 日本語"}`),
	}
	m := defaultPane(8).Open(entry).SetWidth(contentW)
	v := m.View()
	require.NotEmpty(t, v, "expected non-empty view")
	for i, line := range strings.Split(v, "\n") {
		w := lipglossWidth(line)
		if w > outerW {
			assert.Failf(t, "line exceeds outer width",
				"line %d width=%d exceeds outer=%d (content=%d + 2 borders): %q", i, w, outerW, contentW, line)
		}
	}
}

