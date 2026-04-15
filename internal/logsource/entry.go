package logsource

import (
	"encoding/json"
	"time"
)

// Entry represents a single log line, either structured JSON or raw text.
type Entry struct {
	LineNumber int
	IsJSON     bool
	Time       time.Time
	Level      string
	Msg        string
	Logger     string
	Thread     string
	Extra      map[string]json.RawMessage
	Raw        []byte
}
