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

// ParseArgs parses raw argument slices into CLIArgs, resolving the stdin
// source from the real os.Stdin. Returns an error for invalid/missing or
// conflicting arguments.
func ParseArgs(args []string) (CLIArgs, error) {
	return parseArgs(args, stdinIsPiped(os.Stdin))
}

// stdinIsPiped reports whether f (stdin) is a pipe/redirect rather than a tty.
// V39a: a Stat error is treated as "not piped" — never dereferenced — so a
// revoked/unusable stdin fd cannot panic ParseArgs.
func stdinIsPiped(f *os.File) bool {
	st, err := f.Stat()
	if err != nil {
		return false
	}
	return (st.Mode() & os.ModeCharDevice) == 0
}

// parseArgs is the pure core of ParseArgs: input-source resolution given an
// explicit fromStdin flag (V39). Valid modes: {tty + file}, {pipe-stdin (+
// redundant -f)}, {tty + -f + file}.
func parseArgs(args []string, fromStdin bool) (CLIArgs, error) {
	fs := flag.NewFlagSet("gloggy", flag.ContinueOnError)
	follow := fs.Bool("f", false, "tail/follow mode")

	if err := fs.Parse(args); err != nil {
		return CLIArgs{}, err
	}

	switch {
	case fromStdin && fs.NArg() == 0:
		// V23/V31: stdin auto-follows; `-f` is redundant-accepted.
		return CLIArgs{FromStdin: true, FollowMode: true}, nil
	case fromStdin && fs.NArg() >= 1:
		// V39b: piped stdin AND a positional file is ambiguous — reject
		// rather than silently letting the file win and dropping stdin.
		return CLIArgs{}, fmt.Errorf("cannot read both piped stdin and file %q; choose one", fs.Arg(0))
	case fs.NArg() == 1:
		return CLIArgs{FilePath: fs.Arg(0), FollowMode: *follow}, nil
	case fs.NArg() == 0:
		// V39c: no file and no pipe — invalid even with -f.
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
