package logsource

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// T-015: file path → entries; nonexistent file → error
func TestReadFile_ProducesEntries(t *testing.T) {
	path := writeTempLog(t, "line one\nline two\nline three\n")
	entries, err := ReadFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
}

func TestReadFile_NonexistentReturnsError(t *testing.T) {
	_, err := ReadFile("/tmp/gloggy-nonexistent-reader-test-xyzzy.log")
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}

// T-016: stdin (io.Reader) → entries
func TestReadStdin_ProducesEntries(t *testing.T) {
	r := strings.NewReader("hello\nworld\n")
	entries, err := ReadStdin(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].LineNumber != 1 {
		t.Errorf("entry 0 LineNumber = %d, want 1", entries[0].LineNumber)
	}
	if string(entries[0].Raw) != "hello" {
		t.Errorf("entry 0 Raw = %q, want %q", string(entries[0].Raw), "hello")
	}
}

// T-017: line numbers sequential 1..N
func TestReadFile_LineNumbersSequential(t *testing.T) {
	path := writeTempLog(t, "a\nb\nc\nd\ne\n")
	entries, err := ReadFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i, e := range entries {
		if e.LineNumber != i+1 {
			t.Errorf("entry %d: LineNumber = %d, want %d", i, e.LineNumber, i+1)
		}
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
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 5 {
		t.Fatalf("expected 5 entries, got %d", len(entries))
	}

	wantJSON := []bool{false, true, false, true, false}
	for i, e := range entries {
		if e.IsJSON != wantJSON[i] {
			t.Errorf("entry %d: IsJSON = %v, want %v", i, e.IsJSON, wantJSON[i])
		}
		if e.LineNumber != i+1 {
			t.Errorf("entry %d: LineNumber = %d, want %d", i, e.LineNumber, i+1)
		}
	}

	if string(entries[0].Raw) != "plain text line" {
		t.Errorf("entry 0 Raw = %q", string(entries[0].Raw))
	}
	if entries[1].Msg != "hello" {
		t.Errorf("entry 1 Msg = %q, want %q", entries[1].Msg, "hello")
	}
}

func writeTempLog(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.log")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write temp log: %v", err)
	}
	return path
}
