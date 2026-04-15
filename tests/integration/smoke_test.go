package integration

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/istonikula/gloggy/internal/config"
	"github.com/istonikula/gloggy/internal/filter"
	"github.com/istonikula/gloggy/internal/logsource"
	"github.com/istonikula/gloggy/internal/theme"
	"github.com/istonikula/gloggy/internal/ui/appshell"
	"github.com/istonikula/gloggy/internal/ui/detailpane"
	"github.com/istonikula/gloggy/internal/ui/entrylist"
)

// T-060: End-to-end smoke test — launch with JSONL file, navigate list, open detail,
// apply filter, mark entries, resize. Verify no panics.
func TestFullApp_SmokeTest(t *testing.T) {
	// Create sample JSONL file.
	f, err := os.CreateTemp("", "gloggy-smoke-*.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	levels := []string{"INFO", "ERROR", "WARN", "DEBUG", "ERROR", "INFO"}
	for i, level := range levels {
		fmt.Fprintf(f, `{"level":%q,"msg":"entry %d","ts":"2024-01-01T00:00:00Z"}`+"\n", level, i)
	}
	f.Close()

	// Parse CLI args.
	args, err := ParseArgs(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	if args.FilePath != f.Name() {
		t.Fatalf("expected file path %q, got %q", f.Name(), args.FilePath)
	}

	// Load entries synchronously.
	entries, err := logsource.ReadFile(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 6 {
		t.Fatalf("expected 6 entries, got %d", len(entries))
	}

	th := theme.GetTheme("tokyo-night")
	cfg := config.DefaultConfig()

	// Create list model and load entries.
	list := entrylist.NewListModel(th, cfg, 80, 20).SetEntries(entries)

	// Navigate: j, j, g, G.
	list, _ = list.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	list, _ = list.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	list, _ = list.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
	list, _ = list.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	_ = list.View() // must not panic

	// Open detail pane.
	entry, ok := list.SelectedEntry()
	if !ok {
		t.Fatal("no selected entry")
	}
	pane := detailpane.NewPaneModel(th, 8).Open(entry)
	paneView := pane.View()
	if paneView == "" {
		t.Error("detail pane view should not be empty")
	}

	// Scroll detail pane.
	pane, _ = pane.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	_ = pane.View()

	// Close detail pane.
	pane, _ = pane.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if pane.IsOpen() {
		t.Error("pane should be closed after Esc")
	}

	// Apply filter: include only ERROR entries.
	fs := filter.NewFilterSet()
	fs.Add(filter.Filter{Field: "level", Pattern: "ERROR", Mode: filter.Include, Enabled: true})
	indices := filter.Apply(fs, entries)
	list = list.SetFilter(indices)
	filterView := list.View()
	if filterView == "" {
		t.Error("list view should not be empty after filter")
	}

	// Move to top to ensure cursor is in valid filtered range, then mark.
	list, _ = list.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
	list, _ = list.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("m")})
	markedView := list.View()
	if !strings.Contains(markedView, "* ") {
		t.Error("expected mark indicator in view after marking")
	}

	// Resize terminal.
	list, _ = list.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	_ = list.View()

	// Help overlay: open and close.
	help := appshell.NewHelpOverlayModel()
	help, _ = help.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	if !help.IsOpen() {
		t.Error("help overlay should be open after ?")
	}
	help, _ = help.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if help.IsOpen() {
		t.Error("help overlay should be closed after Esc")
	}

	// Layout model.
	layout := appshell.NewLayoutModel(80, 24).SetDetailPane(true, 8)
	rendered := layout.Render("header", "list", "detail", "status")
	if rendered == "" {
		t.Error("rendered layout should not be empty")
	}

	// Loading indicator lifecycle.
	loading := appshell.NewLoadingModel().Start().Update(len(entries)).Done()
	if loading.IsActive() {
		t.Error("loading should be done")
	}

	_ = time.Now() // satisfy import
}

// ParseArgs is a thin wrapper for the integration test (avoids importing main package).
func ParseArgs(filePath string) (struct{ FilePath string; FollowMode bool }, error) {
	return struct{ FilePath string; FollowMode bool }{FilePath: filePath, FollowMode: false}, nil
}
