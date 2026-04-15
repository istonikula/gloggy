package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/istonikula/gloggy/internal/config"
	"github.com/istonikula/gloggy/internal/logsource"
	"github.com/istonikula/gloggy/internal/ui/app"
)

func main() {
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
		return CLIArgs{FromStdin: true, FollowMode: false}, nil
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

	// For stdin: read synchronously before starting the TUI so the full entry
	// list is available immediately (stdin can't be re-read inside the program).
	if parsed.FromStdin {
		entries := logsource.ReadStdin(os.Stdin)
		model = model.SetEntries(entries)
	}

	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	_, err = p.Run()
	return err
}
