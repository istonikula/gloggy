package logsource

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseJSONL_AllKnownFields(t *testing.T) {
	line := []byte(`{"time":"2024-01-15T10:30:00.123456789Z","level":"INFO","msg":"started","logger":"main","thread":"t-1"}`)
	e := ParseJSONL(line, 42)

	assert.Equal(t, 42, e.LineNumber)
	assert.True(t, e.IsJSON, "IsJSON should be true")
	expectedTime, _ := time.Parse(time.RFC3339Nano, "2024-01-15T10:30:00.123456789Z")
	assert.True(t, e.Time.Equal(expectedTime), "Time = %v, want %v", e.Time, expectedTime)
	assert.Equal(t, "INFO", e.Level)
	assert.Equal(t, "started", e.Msg)
	assert.Equal(t, "main", e.Logger)
	assert.Equal(t, "t-1", e.Thread)
	assert.True(t, bytes.Equal(e.Raw, line), "Raw should equal original line")
}

func TestParseJSONL_ExtraKeys(t *testing.T) {
	line := []byte(`{"level":"DEBUG","msg":"x","caller":"foo.go:10"}`)
	e := ParseJSONL(line, 1)
	assert.Len(t, e.Extra, 1, "Extra length")
}

func TestParseJSONL_MissingKnownKeys(t *testing.T) {
	e := ParseJSONL([]byte(`{"msg":"only msg"}`), 5)
	assert.Empty(t, e.Level, "Level should be zero")
	assert.True(t, e.Time.IsZero(), "Time should be zero")
}

func TestParseJSONL_UnparseableTime(t *testing.T) {
	e := ParseJSONL([]byte(`{"time":"not-a-timestamp","msg":"hi"}`), 3)
	assert.True(t, e.Time.IsZero(), "Time should be zero for unparseable timestamp")
}

func TestParseJSONL_PreservesRaw(t *testing.T) {
	line := []byte(`{"msg":"raw check"}`)
	e := ParseJSONL(line, 1)
	assert.True(t, bytes.Equal(e.Raw, line), "Raw bytes not preserved")
}

func TestNewRawEntry_Basic(t *testing.T) {
	line := []byte("plain log line")
	e := NewRawEntry(line, 99)
	assert.Equal(t, 99, e.LineNumber)
	assert.False(t, e.IsJSON, "IsJSON should be false")
	assert.True(t, bytes.Equal(e.Raw, line), "Raw should equal input line")
}

func TestNewRawEntry_StructuredFieldsZero(t *testing.T) {
	e := NewRawEntry([]byte("text"), 1)
	assert.True(t, e.Time.IsZero())
	assert.Empty(t, e.Level)
	assert.Empty(t, e.Msg)
	assert.Empty(t, e.Logger)
	assert.Empty(t, e.Thread)
	assert.Nil(t, e.Extra)
}
