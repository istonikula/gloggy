package filter

import (
	"encoding/json"
	"testing"

	"github.com/istonikula/gloggy/internal/logsource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func baseEntry() logsource.Entry {
	return logsource.Entry{
		Level:  "ERROR",
		Msg:    "connection timeout occurred",
		Logger: "com.example.service.HttpClient",
		Thread: "worker-5",
		Extra: map[string]json.RawMessage{
			"requestId":  json.RawMessage(`"abc-123"`),
			"retryCount": json.RawMessage(`3`),
		},
	}
}

// T-018: literal substring match on msg
func TestMatch_LiteralMsg(t *testing.T) {
	f := Filter{Field: "msg", Pattern: "timeout"}
	ok, err := Match(f, baseEntry())
	require.NoError(t, err)
	assert.True(t, ok, "expected match")
}

func TestMatch_LiteralMsg_NoMatch(t *testing.T) {
	f := Filter{Field: "msg", Pattern: "success"}
	ok, err := Match(f, baseEntry())
	require.NoError(t, err)
	assert.False(t, ok, "expected no match")
}

// T-018: regex match — Syntax must be explicit (V35: no auto-classify)
func TestMatch_Regex(t *testing.T) {
	f := Filter{Field: "msg", Pattern: `timeout.*occurred`, Syntax: Regex}
	ok, err := Match(f, baseEntry())
	require.NoError(t, err)
	assert.True(t, ok, "expected regex match")
}

func TestMatch_Regex_NoMatch(t *testing.T) {
	f := Filter{Field: "msg", Pattern: `^timeout$`, Syntax: Regex}
	ok, err := Match(f, baseEntry())
	require.NoError(t, err)
	assert.False(t, ok, "expected no regex match")
}

// T-018: invalid regex returns error (safety net; Validate catches this at Add/Edit time)
func TestMatch_InvalidRegex_Error(t *testing.T) {
	f := Filter{Field: "msg", Pattern: `[invalid`, Syntax: Regex}
	_, err := Match(f, baseEntry())
	assert.Error(t, err, "expected error for invalid regex")
}

// V35: Literal (default) uses substring match, never regex even with metacharacters.
func TestMatch_LiteralWithMetaChars_SubstringOnly(t *testing.T) {
	entry := logsource.Entry{Msg: "timeout.*occurred"}
	f := Filter{Field: "msg", Pattern: `timeout.*occurred`, Syntax: Literal}
	ok, err := Match(f, entry)
	require.NoError(t, err)
	assert.True(t, ok, "literal match must use substring, not regex")

	// Would match as regex on "connection timeout occurred" — must NOT with Literal.
	f2 := Filter{Field: "msg", Pattern: `timeout.*occurred`, Syntax: Literal}
	ok2, err2 := Match(f2, baseEntry())
	require.NoError(t, err2)
	assert.False(t, ok2, "literal pattern 'timeout.*occurred' is not a substring of the base entry msg")
}

// V35: Glob syntax — * and ? wildcards.
func TestMatch_Glob_Star(t *testing.T) {
	f := Filter{Field: "msg", Pattern: `timeout*occurred`, Syntax: Glob}
	ok, err := Match(f, baseEntry())
	require.NoError(t, err)
	assert.True(t, ok, "glob * should match any chars between timeout and occurred")
}

func TestMatch_Glob_Question(t *testing.T) {
	f := Filter{Field: "level", Pattern: "ERR?R", Syntax: Glob}
	ok, err := Match(f, baseEntry())
	require.NoError(t, err)
	assert.True(t, ok, "glob ? should match single char")
}

func TestMatch_Glob_NoMatch(t *testing.T) {
	f := Filter{Field: "level", Pattern: "WAR?", Syntax: Glob}
	ok, err := Match(f, baseEntry())
	require.NoError(t, err)
	assert.False(t, ok, "glob should not match ERROR with WAR?")
}

func TestMatch_Glob_LiteralDot_NotWildcard(t *testing.T) {
	// A dot in glob is escaped to \. in RE2 — matches literal dot, not any char.
	entry := logsource.Entry{Logger: "com.example.service"}
	f := Filter{Field: "logger", Pattern: "com.example", Syntax: Glob}
	ok, err := Match(f, entry)
	require.NoError(t, err)
	assert.True(t, ok, "dot in glob pattern should match literal dot")

	entryNoDot := logsource.Entry{Logger: "comXexample"}
	ok2, err2 := Match(f, entryNoDot)
	require.NoError(t, err2)
	assert.False(t, ok2, "dot in glob must not match non-dot char (RE2 . wildcard escaped)")
}

// V35: whole-line filter (Field == "") matches against entry.Raw.
func TestMatch_WholeLine_MatchesRaw(t *testing.T) {
	entry := logsource.Entry{
		Raw: []byte(`{"level":"ERROR","msg":"connection timeout occurred"}`),
	}
	f := Filter{Field: "", Pattern: "timeout", Syntax: Literal}
	ok, err := Match(f, entry)
	require.NoError(t, err)
	assert.True(t, ok, "whole-line filter should match substring in Raw")
}

func TestMatch_WholeLine_MatchesKey(t *testing.T) {
	entry := logsource.Entry{
		Raw: []byte(`{"requestId":"abc-123","level":"INFO"}`),
	}
	// Match on a key name — not possible via field-scoped filter.
	f := Filter{Field: "", Pattern: "requestId", Syntax: Literal}
	ok, err := Match(f, entry)
	require.NoError(t, err)
	assert.True(t, ok, "whole-line filter should match on key names too")
}

func TestMatch_WholeLine_NoMatch(t *testing.T) {
	entry := logsource.Entry{Raw: []byte(`{"level":"INFO","msg":"hello"}`)}
	f := Filter{Field: "", Pattern: "ERROR", Syntax: Literal}
	ok, err := Match(f, entry)
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestMatch_WholeLine_Glob(t *testing.T) {
	entry := logsource.Entry{Raw: []byte(`{"level":"ERROR","msg":"DrawStateChanged"}`)}
	f := Filter{Field: "", Pattern: "Draw*Changed", Syntax: Glob}
	ok, err := Match(f, entry)
	require.NoError(t, err)
	assert.True(t, ok, "whole-line glob should match")
}

func TestMatch_WholeLine_EmptyRaw_NoMatch(t *testing.T) {
	entry := logsource.Entry{Raw: nil}
	f := Filter{Field: "", Pattern: "anything", Syntax: Literal}
	ok, err := Match(f, entry)
	require.NoError(t, err)
	assert.False(t, ok, "nil Raw is empty string — no match for non-empty pattern")
}

// T-018: match against level, logger, thread
func TestMatch_Level(t *testing.T) {
	f := Filter{Field: "level", Pattern: "ERROR"}
	ok, err := Match(f, baseEntry())
	require.NoError(t, err)
	assert.True(t, ok, "expected match on level")
}

func TestMatch_Logger(t *testing.T) {
	f := Filter{Field: "logger", Pattern: "HttpClient"}
	ok, err := Match(f, baseEntry())
	require.NoError(t, err)
	assert.True(t, ok, "expected match on logger")
}

func TestMatch_Thread(t *testing.T) {
	f := Filter{Field: "thread", Pattern: "worker"}
	ok, err := Match(f, baseEntry())
	require.NoError(t, err)
	assert.True(t, ok, "expected match on thread")
}

// T-018: match against extra field (string value)
func TestMatch_ExtraStringField(t *testing.T) {
	f := Filter{Field: "requestId", Pattern: "abc"}
	ok, err := Match(f, baseEntry())
	require.NoError(t, err)
	assert.True(t, ok, "expected match on extra string field")
}

// T-018: match against extra field (numeric value)
func TestMatch_ExtraNumericField(t *testing.T) {
	f := Filter{Field: "retryCount", Pattern: "3"}
	ok, err := Match(f, baseEntry())
	require.NoError(t, err)
	assert.True(t, ok, "expected match on extra numeric field")
}

// T-018: missing field → no match (not error)
func TestMatch_MissingField(t *testing.T) {
	f := Filter{Field: "nonexistent", Pattern: "x"}
	ok, err := Match(f, baseEntry())
	require.NoError(t, err)
	assert.False(t, ok, "expected no match on missing field")
}

// T-026: ToggleAll disables all filters
func TestToggleAll_DisablesAll(t *testing.T) {
	fs := NewFilterSet()
	fs.Add(Filter{Field: "a", Enabled: true})
	fs.Add(Filter{Field: "b", Enabled: true})
	fs.Add(Filter{Field: "c", Enabled: false})
	fs.ToggleAll()
	assert.Empty(t, fs.GetEnabled(), "expected 0 enabled after ToggleAll")
}

// T-026: second ToggleAll re-enables previously-enabled filters
func TestToggleAll_ReEnablesEnabled(t *testing.T) {
	fs := NewFilterSet()
	fs.Add(Filter{Field: "a", Enabled: true})
	fs.Add(Filter{Field: "b", Enabled: true})
	fs.ToggleAll()
	fs.ToggleAll()
	assert.Len(t, fs.GetEnabled(), 2, "expected 2 re-enabled")
}

// T-026: individually-disabled filters stay disabled after re-enable
func TestToggleAll_IndividuallyDisabledStaysDisabled(t *testing.T) {
	fs := NewFilterSet()
	id0 := fs.Add(Filter{Field: "a", Enabled: true})
	fs.Add(Filter{Field: "b", Enabled: true})
	fs.Disable(id0) // disable "a" individually before global toggle
	fs.ToggleAll()
	fs.ToggleAll()
	enabled := fs.GetEnabled()
	require.Len(t, enabled, 1)
	assert.Equal(t, "b", enabled[0].Field, "expected 'b' to be enabled")
}

// T-069: Match must not panic when entry.Extra is nil.
func TestMatch_NilExtra(t *testing.T) {
	entry := logsource.Entry{
		Level: "INFO",
		Msg:   "hello",
		Extra: nil,
	}
	f := Filter{Field: "somekey", Pattern: "val"}
	ok, err := Match(f, entry)
	require.NoError(t, err)
	assert.False(t, ok, "expected no match for nil Extra")
}

// T-077: JSON string unquoting must handle escape sequences.
func TestMatch_JSONEscapedStrings(t *testing.T) {
	entry := logsource.Entry{
		Extra: map[string]json.RawMessage{
			"path":    json.RawMessage(`"C:\\Users\\test"`),
			"quoted":  json.RawMessage(`"say \"hello\""`),
			"newline": json.RawMessage(`"line1\nline2"`),
		},
	}

	tests := []struct {
		name, field, pattern string
		want                 bool
	}{
		{"backslash path substr", "path", "Users", true},
		{"escaped quotes", "quoted", `say "hello"`, true},
		{"contains newline", "newline", "line1\nline2", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := Filter{Field: tc.field, Pattern: tc.pattern}
			ok, err := Match(f, entry)
			require.NoError(t, err)
			assert.Equal(t, tc.want, ok, "Match()")
		})
	}

	// Verify entryFieldValue correctly unquotes JSON strings.
	unquoteTests := []struct {
		name, field, want string
	}{
		{"backslash unquote", "path", `C:\Users\test`},
		{"quotes unquote", "quoted", `say "hello"`},
		{"newline unquote", "newline", "line1\nline2"},
	}
	for _, tc := range unquoteTests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := entryFieldValue(tc.field, entry)
			require.True(t, ok, "expected field %q to be found", tc.field)
			assert.Equal(t, tc.want, got, "entryFieldValue(%q)", tc.field)
		})
	}
}

// T-071: Benchmark regex matching to verify caching helps.
func BenchmarkMatch_Regex(b *testing.B) {
	entry := logsource.Entry{Msg: "connection refused from 192.168.1.1"}
	f := Filter{Field: "msg", Pattern: `\d+\.\d+\.\d+\.\d+`, Syntax: Regex}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Match(f, entry)
	}
}

// T-074: Adding filter while globally disabled must not break re-enable.
func TestToggleAll_AddWhileDisabled(t *testing.T) {
	fs := NewFilterSet()
	fs.Add(Filter{Field: "a", Enabled: true})
	fs.Add(Filter{Field: "b", Enabled: false})

	fs.ToggleAll()

	// Add new filter while globally disabled.
	fs.Add(Filter{Field: "c", Enabled: true})

	// c should be disabled immediately.
	all := fs.GetAll()
	assert.False(t, all[2].Enabled, "filter c should be disabled while globally disabled")

	// Re-enable all.
	fs.ToggleAll()

	all = fs.GetAll()
	assert.True(t, all[0].Enabled, "a should be re-enabled")
	assert.False(t, all[1].Enabled, "b should still be disabled (was disabled before toggle)")
	assert.True(t, all[2].Enabled, "c should be re-enabled (was enabled when added)")
}

// T-074: Removing filter while globally disabled must not break re-enable.
func TestToggleAll_RemoveWhileDisabled(t *testing.T) {
	fs := NewFilterSet()
	id0 := fs.Add(Filter{Field: "a", Enabled: true})
	fs.Add(Filter{Field: "b", Enabled: true})

	fs.ToggleAll()
	fs.Remove(id0)
	fs.ToggleAll()

	all := fs.GetAll()
	require.Len(t, all, 1)
	assert.True(t, all[0].Enabled, "b should be re-enabled")
}
