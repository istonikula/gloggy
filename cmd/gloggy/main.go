package main

import (
	"flag"
	"fmt"
	"os"
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
	// Placeholder: real app would launch the Bubble Tea program here.
	if parsed.FromStdin {
		fmt.Println("reading from stdin")
	} else {
		mode := ""
		if parsed.FollowMode {
			mode = " (follow)"
		}
		fmt.Printf("reading file: %s%s\n", parsed.FilePath, mode)
	}
	return nil
}
