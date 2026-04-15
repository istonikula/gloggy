package main

import (
	"strings"
	"testing"
)

// T-049: R1.1 — gloggy <file> parses correctly.
func TestParseArgs_FileArg(t *testing.T) {
	args, err := ParseArgs([]string{"app.log"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if args.FilePath != "app.log" {
		t.Errorf("FilePath: got %q, want %q", args.FilePath, "app.log")
	}
	if args.FollowMode {
		t.Error("FollowMode should be false for plain file arg")
	}
	if args.FromStdin {
		t.Error("FromStdin should be false for plain file arg")
	}
}

// T-049: R1.2 — gloggy -f <file> sets follow mode.
func TestParseArgs_FollowMode(t *testing.T) {
	args, err := ParseArgs([]string{"-f", "app.log"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if args.FilePath != "app.log" {
		t.Errorf("FilePath: got %q, want %q", args.FilePath, "app.log")
	}
	if !args.FollowMode {
		t.Error("FollowMode should be true with -f flag")
	}
}

// T-049: R1.4 — invalid args return clear error.
func TestParseArgs_TooManyArgs_Error(t *testing.T) {
	_, err := ParseArgs([]string{"a.log", "b.log"})
	if err == nil {
		t.Error("expected error for too many args")
	}
	if !strings.Contains(err.Error(), "too many") {
		t.Errorf("error message should mention 'too many': %v", err)
	}
}
