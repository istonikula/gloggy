package filter

import (
	"encoding/json"
	"testing"

	"github.com/istonikula/gloggy/internal/logsource"
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
	if err != nil || !ok {
		t.Errorf("expected match; err=%v ok=%v", err, ok)
	}
}

func TestMatch_LiteralMsg_NoMatch(t *testing.T) {
	f := Filter{Field: "msg", Pattern: "success"}
	ok, err := Match(f, baseEntry())
	if err != nil || ok {
		t.Errorf("expected no match; err=%v ok=%v", err, ok)
	}
}

// T-018: regex match
func TestMatch_Regex(t *testing.T) {
	f := Filter{Field: "msg", Pattern: `timeout.*occurred`}
	ok, err := Match(f, baseEntry())
	if err != nil || !ok {
		t.Errorf("expected regex match; err=%v ok=%v", err, ok)
	}
}

func TestMatch_Regex_NoMatch(t *testing.T) {
	f := Filter{Field: "msg", Pattern: `^timeout$`}
	ok, err := Match(f, baseEntry())
	if err != nil || ok {
		t.Errorf("expected no regex match; err=%v ok=%v", err, ok)
	}
}

// T-018: invalid regex returns error, not applied
func TestMatch_InvalidRegex_Error(t *testing.T) {
	f := Filter{Field: "msg", Pattern: `[invalid`}
	_, err := Match(f, baseEntry())
	if err == nil {
		t.Error("expected error for invalid regex")
	}
}

// T-018: match against level, logger, thread
func TestMatch_Level(t *testing.T) {
	f := Filter{Field: "level", Pattern: "ERROR"}
	ok, err := Match(f, baseEntry())
	if err != nil || !ok {
		t.Errorf("expected match on level; err=%v ok=%v", err, ok)
	}
}

func TestMatch_Logger(t *testing.T) {
	f := Filter{Field: "logger", Pattern: "HttpClient"}
	ok, err := Match(f, baseEntry())
	if err != nil || !ok {
		t.Errorf("expected match on logger; err=%v ok=%v", err, ok)
	}
}

func TestMatch_Thread(t *testing.T) {
	f := Filter{Field: "thread", Pattern: "worker"}
	ok, err := Match(f, baseEntry())
	if err != nil || !ok {
		t.Errorf("expected match on thread; err=%v ok=%v", err, ok)
	}
}

// T-018: match against extra field (string value)
func TestMatch_ExtraStringField(t *testing.T) {
	f := Filter{Field: "requestId", Pattern: "abc"}
	ok, err := Match(f, baseEntry())
	if err != nil || !ok {
		t.Errorf("expected match on extra string field; err=%v ok=%v", err, ok)
	}
}

// T-018: match against extra field (numeric value)
func TestMatch_ExtraNumericField(t *testing.T) {
	f := Filter{Field: "retryCount", Pattern: "3"}
	ok, err := Match(f, baseEntry())
	if err != nil || !ok {
		t.Errorf("expected match on extra numeric field; err=%v ok=%v", err, ok)
	}
}

// T-018: missing field → no match (not error)
func TestMatch_MissingField(t *testing.T) {
	f := Filter{Field: "nonexistent", Pattern: "x"}
	ok, err := Match(f, baseEntry())
	if err != nil || ok {
		t.Errorf("expected no match on missing field; err=%v ok=%v", err, ok)
	}
}

// T-026: ToggleAll disables all filters
func TestToggleAll_DisablesAll(t *testing.T) {
	fs := NewFilterSet()
	fs.Add(Filter{Field: "a", Enabled: true})
	fs.Add(Filter{Field: "b", Enabled: true})
	fs.Add(Filter{Field: "c", Enabled: false})
	fs.ToggleAll()
	if got := fs.GetEnabled(); len(got) != 0 {
		t.Errorf("expected 0 enabled after ToggleAll, got %d", len(got))
	}
}

// T-026: second ToggleAll re-enables previously-enabled filters
func TestToggleAll_ReEnablesEnabled(t *testing.T) {
	fs := NewFilterSet()
	fs.Add(Filter{Field: "a", Enabled: true})
	fs.Add(Filter{Field: "b", Enabled: true})
	fs.ToggleAll()
	fs.ToggleAll()
	if got := fs.GetEnabled(); len(got) != 2 {
		t.Errorf("expected 2 re-enabled, got %d", len(got))
	}
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
	if len(enabled) != 1 {
		t.Fatalf("expected 1 enabled, got %d", len(enabled))
	}
	if enabled[0].Field != "b" {
		t.Errorf("expected 'b' to be enabled, got %q", enabled[0].Field)
	}
}
