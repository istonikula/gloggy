package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/istonikula/gloggy/internal/config"
	"github.com/istonikula/gloggy/internal/ui/app"
)

// forceTrueColorIfSupported honours COLORTERM=truecolor / 24bit by forcing
// lipgloss onto the TrueColor profile. Without this, termenv's default
// detection can return a downsampled profile (256-color or Ascii) when
// stdout is wrapped by a non-canonical PTY (e.g. a test harness or an MCP
// TUI driver). In that state every theme's `BaseBg` hex collapses to the
// same xterm-256 palette slot and the three bundled themes render as
// visually identical — which is what defeated the first T-179 verification.
// Real-world terminals that support TrueColor advertise it via COLORTERM,
// so this is a no-op on terminals that don't, and correct on the ones that
// do. Closes the observability half of F-202.
func forceTrueColorIfSupported() {
	switch os.Getenv("COLORTERM") {
	case "truecolor", "24bit":
		lipgloss.SetColorProfile(termenv.TrueColor)
	}
}

func main() {
	forceTrueColorIfSupported()
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// CLIArgs holds the parsed command-line arguments.
type CLIArgs struct {
	FilePath   string
	FollowMode bool
	FromStdin  bool
}

// ParseArgs parses raw argument slices into CLIArgs.
// Returns an error for invalid/missing arguments.
func ParseArgs(args []string) (CLIArgs, error) {
	fs := flag.NewFlagSet("gloggy", flag.ContinueOnError)
	follow := fs.Bool("f", false, "tail/follow mode")

	if err := fs.Parse(args); err != nil {
		return CLIArgs{}, err
	}

	// Check if stdin is piped (not a tty).
	stdinStat, _ := os.Stdin.Stat()
	fromStdin := (stdinStat.Mode() & os.ModeCharDevice) == 0

	switch {
	case fs.NArg() == 0 && fromStdin:
		// V23/V31: stdin auto-follows; `-f` is redundant-accepted.
		return CLIArgs{FromStdin: true, FollowMode: true}, nil
	case fs.NArg() == 1:
		return CLIArgs{FilePath: fs.Arg(0), FollowMode: *follow}, nil
	case fs.NArg() == 0:
		return CLIArgs{}, fmt.Errorf("usage: gloggy [-f] <file>  or  gloggy (with piped stdin)")
	default:
		return CLIArgs{}, fmt.Errorf("too many arguments: expected 1 file path")
	}
}

func run(args []string) error {
	parsed, err := ParseArgs(args)
	if err != nil {
		return err
	}

	// Load config.
	cfgPath, err := config.DefaultConfigPath()
	if err != nil {
		cfgPath = ""
	}
	cfgResult := config.Load(cfgPath)
	for _, w := range cfgResult.Warnings {
		fmt.Fprintln(os.Stderr, "config warning:", w)
	}

	// Determine source name for the header.
	sourceName := parsed.FilePath
	if parsed.FromStdin {
		sourceName = ""
	}

	model := app.New(sourceName, parsed.FollowMode, cfgPath, cfgResult)
	if parsed.FromStdin {
		model = model.WithStdinReader(os.Stdin)
	}

	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	_, err = p.Run()
	return err
}
