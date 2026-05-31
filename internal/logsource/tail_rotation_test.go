package logsource

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeFile (over)writes path with the given content.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
}

func msgs(entries []Entry) []string {
	out := make([]string, len(entries))
	for i, e := range entries {
		out[i] = e.Msg
	}
	return out
}

// V37a: a persistent tailReader reads appended lines exactly once across
// successive drains — no re-read, no skip.
func TestTailReader_AppendAcrossDrains_V37(t *testing.T) {
	path := filepath.Join(t.TempDir(), "append.jsonl")
	writeFile(t, path, `{"msg":"a"}`+"\n"+`{"msg":"b"}`+"\n")

	tr, err := newTailReader(path, 0)
	require.NoError(t, err)
	defer tr.close()

	first, err := tr.drain()
	require.NoError(t, err)
	assert.Equal(t, []string{"a", "b"}, msgs(first))

	// No new bytes → empty drain, no duplicate re-read.
	empty, err := tr.drain()
	require.NoError(t, err)
	assert.Empty(t, empty)

	// Append → only the new line is returned.
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	require.NoError(t, err)
	fmt.Fprintf(f, `{"msg":"c"}`+"\n")
	f.Close()

	third, err := tr.drain()
	require.NoError(t, err)
	require.Len(t, third, 1)
	assert.Equal(t, "c", third[0].Msg)
	assert.Equal(t, 3, third[0].LineNumber, "line numbering continues")
}

// V37b: truncate-in-place (file shrinks below the consumed offset) is detected,
// the handle is reopened by path, and post-truncation content is read with zero
// stale lines.
func TestTailReader_TruncateInPlace_V37(t *testing.T) {
	path := filepath.Join(t.TempDir(), "trunc.jsonl")
	writeFile(t, path, `{"msg":"old1"}`+"\n"+`{"msg":"old2"}`+"\n"+`{"msg":"old3"}`+"\n")

	tr, err := newTailReader(path, 0)
	require.NoError(t, err)
	defer tr.close()

	first, err := tr.drain()
	require.NoError(t, err)
	assert.Equal(t, []string{"old1", "old2", "old3"}, msgs(first))

	// Truncate in place + write fresh, shorter content.
	writeFile(t, path, `{"msg":"fresh"}`+"\n")

	after, err := tr.drain()
	require.NoError(t, err)
	require.Len(t, after, 1, "exactly the post-truncation line, no stale")
	assert.Equal(t, "fresh", after[0].Msg)
}

// V37a/b: rotation — the original file is renamed away and a new (smaller) file
// takes its path. The size-shrink check reopens by path so the replacement
// file's content is read with no loss.
func TestTailReader_Rotation_V37(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rotate.jsonl")
	writeFile(t, path, `{"msg":"pre1"}`+"\n"+`{"msg":"pre2"}`+"\n")

	tr, err := newTailReader(path, 0)
	require.NoError(t, err)
	defer tr.close()

	first, err := tr.drain()
	require.NoError(t, err)
	assert.Equal(t, []string{"pre1", "pre2"}, msgs(first))

	// logrotate-style: move the current file aside, create a new one at path.
	require.NoError(t, os.Rename(path, filepath.Join(dir, "rotate.jsonl.1")))
	writeFile(t, path, `{"msg":"post"}`+"\n")

	after, err := tr.drain()
	require.NoError(t, err)
	require.Len(t, after, 1)
	assert.Equal(t, "post", after[0].Msg, "reads the rotated-in replacement file")
}

// V37b: a partial (newline-less) line buffered before a truncation must be
// dropped on reopen — never concatenated with post-truncation bytes.
func TestTailReader_PartialLineDroppedOnTruncate_V37(t *testing.T) {
	path := filepath.Join(t.TempDir(), "partial.jsonl")
	writeFile(t, path, `{"msg":"complete"}`+"\n"+`{"msg":"part`) // trailing partial, no newline

	tr, err := newTailReader(path, 0)
	require.NoError(t, err)
	defer tr.close()

	first, err := tr.drain()
	require.NoError(t, err)
	require.Len(t, first, 1, "only the newline-terminated line emits")
	assert.Equal(t, "complete", first[0].Msg)
	assert.NotEmpty(t, tr.pending, "partial line is buffered")

	// Truncate + write fresh content shorter than the consumed offset.
	writeFile(t, path, `{"msg":"new"}`+"\n")

	after, err := tr.drain()
	require.NoError(t, err)
	require.Len(t, after, 1)
	assert.Equal(t, "new", after[0].Msg, "stale partial discarded, not prepended")
}

// V37/V3 end-to-end: TailFile survives a truncate-in-place mid-stream and keeps
// streaming (no TailStopMsg) — [FOLLOW] would stay live. Uses real fsnotify; on
// environments where the Write event never arrives this hard-fails rather than
// silently passing (B44 lesson).
func TestTailFile_SurvivesTruncateInPlace_V37(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "live.jsonl")
	writeFile(t, path, `{"msg":"a"}`+"\n"+`{"msg":"b"}`+"\n")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tailCmd := TailFile(ctx, path, 2) // skip the 2 initial lines
	cmd := func() interface{} { return tailCmd() }

	time.Sleep(150 * time.Millisecond) // watcher warmup

	// Truncate in place, then write 3 fresh lines.
	writeFile(t, path, "")
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	require.NoError(t, err)
	for i := 1; i <= 3; i++ {
		fmt.Fprintf(f, `{"msg":"post %d"}`+"\n", i)
	}
	f.Close()

	var got []Entry
	timeout := time.After(5 * time.Second)
	for len(got) < 3 {
		select {
		case <-timeout:
			require.Failf(t, "timed out", "post-truncation lines not received (got %d, want 3)", len(got))
		default:
		}
		switch m := cmd().(type) {
		case TailStreamMsg:
			if tm, ok := m.Unwrap().(TailMsg); ok {
				got = append(got, tm.Entries...)
			}
			next := m.Next()
			cmd = func() interface{} { return next() }
		case TailStopMsg:
			require.Failf(t, "tail stopped during truncation", "%v", m.Err)
		}
	}
	assert.Equal(t, []string{"post 1", "post 2", "post 3"}, msgs(got))
}
