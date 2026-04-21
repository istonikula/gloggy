package logsource

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T-015: file path → entries; nonexistent file → error
func TestReadFile_ProducesEntries(t *testing.T) {
	path := writeTempLog(t, "line one\nline two\nline three\n")
	entries, err := ReadFile(path)
	require.NoError(t, err)
	require.Len(t, entries, 3, "expected 3 entries")
}

func TestReadFile_NonexistentReturnsError(t *testing.T) {
	_, err := ReadFile("/tmp/gloggy-nonexistent-reader-test-xyzzy.log")
	require.Error(t, err, "expected error for nonexistent file")
}

// T-016: stdin (io.Reader) → entries
func TestReadStdin_ProducesEntries(t *testing.T) {
	r := strings.NewReader("hello\nworld\n")
	entries, err := ReadStdin(r)
	require.NoError(t, err)
	require.Len(t, entries, 2, "expected 2 entries")
	assert.Equal(t, 1, entries[0].LineNumber, "entry 0 LineNumber")
	assert.Equal(t, "hello", string(entries[0].Raw), "entry 0 Raw")
}

// T-017: line numbers sequential 1..N
func TestReadFile_LineNumbersSequential(t *testing.T) {
	path := writeTempLog(t, "a\nb\nc\nd\ne\n")
	entries, err := ReadFile(path)
	require.NoError(t, err)
	for i, e := range entries {
		assert.Equal(t, i+1, e.LineNumber, "entry %d LineNumber", i)
	}
}

// T-017: interleaved JSON and raw preserve order and line numbers
func TestReadFile_MixedContentOrder(t *testing.T) {
	lines := []string{
		`plain text line`,
		`{"level":"INFO","msg":"hello"}`,
		`another raw line`,
		`{"level":"ERROR","msg":"boom"}`,
		`final plain`,
	}
	path := writeTempLog(t, strings.Join(lines, "\n")+"\n")

	entries, err := ReadFile(path)
	require.NoError(t, err)
	require.Len(t, entries, 5, "expected 5 entries")

	wantJSON := []bool{false, true, false, true, false}
	for i, e := range entries {
		assert.Equal(t, wantJSON[i], e.IsJSON, "entry %d IsJSON", i)
		assert.Equal(t, i+1, e.LineNumber, "entry %d LineNumber", i)
	}

	assert.Equal(t, "plain text line", string(entries[0].Raw), "entry 0 Raw")
	assert.Equal(t, "hello", entries[1].Msg, "entry 1 Msg")
}

func writeTempLog(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.log")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644), "write temp log")
	return path
}
