package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T-049: R1.1 — gloggy <file> parses correctly.
func TestParseArgs_FileArg(t *testing.T) {
	args, err := ParseArgs([]string{"app.log"})
	require.NoError(t, err)
	assert.Equal(t, "app.log", args.FilePath)
	assert.False(t, args.FollowMode, "FollowMode should be false for plain file arg")
	assert.False(t, args.FromStdin, "FromStdin should be false for plain file arg")
}

// T-049: R1.2 — gloggy -f <file> sets follow mode.
func TestParseArgs_FollowMode(t *testing.T) {
	args, err := ParseArgs([]string{"-f", "app.log"})
	require.NoError(t, err)
	assert.Equal(t, "app.log", args.FilePath)
	assert.True(t, args.FollowMode, "FollowMode should be true with -f flag")
}

// T-049: R1.4 — invalid args return clear error.
func TestParseArgs_TooManyArgs_Error(t *testing.T) {
	_, err := ParseArgs([]string{"a.log", "b.log"})
	require.Error(t, err, "expected error for too many args")
	assert.Contains(t, err.Error(), "too many", "error message should mention 'too many'")
}

// swapStdinPipe replaces os.Stdin with a pipe reader (non-tty fd) so
// ParseArgs' `stdinStat.Mode() & os.ModeCharDevice == 0` branch fires.
// Returns a restore func for defer. Test-helper only.
func swapStdinPipe(t *testing.T) func() {
	t.Helper()
	pr, pw, err := os.Pipe()
	require.NoError(t, err)
	orig := os.Stdin
	os.Stdin = pr
	return func() {
		os.Stdin = orig
		_ = pr.Close()
		_ = pw.Close()
	}
}

// T25 / V23: piped stdin + no args → FromStdin=true, FollowMode=true
// (stdin auto-follows per V23/V31; no -f flag required).
func TestParseArgs_PipedStdin_NoFlag(t *testing.T) {
	restore := swapStdinPipe(t)
	defer restore()

	args, err := ParseArgs([]string{})
	require.NoError(t, err)
	assert.True(t, args.FromStdin, "FromStdin should be true when stdin is piped")
	assert.True(t, args.FollowMode, "FollowMode should auto-enable for piped stdin (V23)")
	assert.Empty(t, args.FilePath, "FilePath should be empty for piped stdin")
}

// T25 / V23: piped stdin + `-f` → same as no-flag; `-f` redundant-accepted
// (follow already on for stdin, flag is a no-op here).
func TestParseArgs_PipedStdin_WithFollowFlag_RedundantAccepted(t *testing.T) {
	restore := swapStdinPipe(t)
	defer restore()

	args, err := ParseArgs([]string{"-f"})
	require.NoError(t, err, "-f with piped stdin must not error (redundant-accepted)")
	assert.True(t, args.FromStdin, "FromStdin should be true when stdin is piped")
	assert.True(t, args.FollowMode, "FollowMode should be true (redundant -f is accepted)")
	assert.Empty(t, args.FilePath)
}
