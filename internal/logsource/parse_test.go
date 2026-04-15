package logsource

import (
	"bytes"
	"testing"
	"time"
)

func TestParseJSONL_AllKnownFields(t *testing.T) {
	line := []byte(`{"time":"2024-01-15T10:30:00.123456789Z","level":"INFO","msg":"started","logger":"main","thread":"t-1"}`)
	e := ParseJSONL(line, 42)

	if e.LineNumber != 42 {
		t.Errorf("LineNumber = %d, want 42", e.LineNumber)
	}
	if !e.IsJSON {
		t.Error("IsJSON should be true")
	}
	expectedTime, _ := time.Parse(time.RFC3339Nano, "2024-01-15T10:30:00.123456789Z")
	if !e.Time.Equal(expectedTime) {
		t.Errorf("Time = %v, want %v", e.Time, expectedTime)
	}
	if e.Level != "INFO" {
		t.Errorf("Level = %q, want INFO", e.Level)
	}
	if e.Msg != "started" {
		t.Errorf("Msg = %q, want started", e.Msg)
	}
	if e.Logger != "main" {
		t.Errorf("Logger = %q, want main", e.Logger)
	}
	if e.Thread != "t-1" {
		t.Errorf("Thread = %q, want t-1", e.Thread)
	}
	if !bytes.Equal(e.Raw, line) {
		t.Error("Raw should equal original line")
	}
}

func TestParseJSONL_ExtraKeys(t *testing.T) {
	line := []byte(`{"level":"DEBUG","msg":"x","caller":"foo.go:10"}`)
	e := ParseJSONL(line, 1)
	if len(e.Extra) != 1 {
		t.Fatalf("Extra length = %d, want 1", len(e.Extra))
	}
}

func TestParseJSONL_MissingKnownKeys(t *testing.T) {
	e := ParseJSONL([]byte(`{"msg":"only msg"}`), 5)
	if e.Level != "" {
		t.Errorf("Level should be zero, got %q", e.Level)
	}
	if !e.Time.IsZero() {
		t.Error("Time should be zero")
	}
}

func TestParseJSONL_UnparseableTime(t *testing.T) {
	e := ParseJSONL([]byte(`{"time":"not-a-timestamp","msg":"hi"}`), 3)
	if !e.Time.IsZero() {
		t.Error("Time should be zero for unparseable timestamp")
	}
}

func TestParseJSONL_PreservesRaw(t *testing.T) {
	line := []byte(`{"msg":"raw check"}`)
	e := ParseJSONL(line, 1)
	if !bytes.Equal(e.Raw, line) {
		t.Error("Raw bytes not preserved")
	}
}

func TestNewRawEntry_Basic(t *testing.T) {
	line := []byte("plain log line")
	e := NewRawEntry(line, 99)
	if e.LineNumber != 99 {
		t.Errorf("LineNumber = %d, want 99", e.LineNumber)
	}
	if e.IsJSON {
		t.Error("IsJSON should be false")
	}
	if !bytes.Equal(e.Raw, line) {
		t.Error("Raw should equal input line")
	}
}

func TestNewRawEntry_StructuredFieldsZero(t *testing.T) {
	e := NewRawEntry([]byte("text"), 1)
	if !e.Time.IsZero() || e.Level != "" || e.Msg != "" || e.Logger != "" || e.Thread != "" || e.Extra != nil {
		t.Error("all structured fields should be zero")
	}
}
