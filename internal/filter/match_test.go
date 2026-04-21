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

// T-018: regex match
func TestMatch_Regex(t *testing.T) {
	f := Filter{Field: "msg", Pattern: `timeout.*occurred`}
	ok, err := Match(f, baseEntry())
	require.NoError(t, err)
	assert.True(t, ok, "expected regex match")
}

func TestMatch_Regex_NoMatch(t *testing.T) {
	f := Filter{Field: "msg", Pattern: `^timeout$`}
	ok, err := Match(f, baseEntry())
	require.NoError(t, err)
	assert.False(t, ok, "expected no regex match")
}

// T-018: invalid regex returns error, not applied
func TestMatch_InvalidRegex_Error(t *testing.T) {
	f := Filter{Field: "msg", Pattern: `[invalid`}
	_, err := Match(f, baseEntry())
	assert.Error(t, err, "expected error for invalid regex")
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
	f := Filter{Field: "msg", Pattern: `\d+\.\d+\.\d+\.\d+`}
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
