package main

import (
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
