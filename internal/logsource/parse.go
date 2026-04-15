package logsource

import (
	"encoding/json"
	"time"
)

var knownKeys = map[string]bool{
	"time":   true,
	"level":  true,
	"msg":    true,
	"logger": true,
	"thread": true,
}

// ParseJSONL parses a JSON object line into an Entry with IsJSON=true.
// Known keys are extracted into structured fields; remaining keys go to Extra.
// Unparseable or missing time values leave Time at zero value.
func ParseJSONL(line []byte, lineNum int) Entry {
	e := Entry{
		LineNumber: lineNum,
		IsJSON:     true,
		Raw:        line,
	}

	var obj map[string]json.RawMessage
	if err := json.Unmarshal(line, &obj); err != nil {
		return e
	}

	if v, ok := obj["time"]; ok {
		var s string
		if json.Unmarshal(v, &s) == nil {
			if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
				e.Time = t
			}
		}
	}
	if v, ok := obj["level"]; ok {
		var s string
		if json.Unmarshal(v, &s) == nil {
			e.Level = s
		}
	}
	if v, ok := obj["msg"]; ok {
		var s string
		if json.Unmarshal(v, &s) == nil {
			e.Msg = s
		}
	}
	if v, ok := obj["logger"]; ok {
		var s string
		if json.Unmarshal(v, &s) == nil {
			e.Logger = s
		}
	}
	if v, ok := obj["thread"]; ok {
		var s string
		if json.Unmarshal(v, &s) == nil {
			e.Thread = s
		}
	}

	for k, v := range obj {
		if knownKeys[k] {
			continue
		}
		if e.Extra == nil {
			e.Extra = make(map[string]json.RawMessage)
		}
		e.Extra[k] = v
	}

	return e
}

// NewRawEntry creates an Entry for a non-JSON line.
func NewRawEntry(line []byte, lineNum int) Entry {
	return Entry{
		LineNumber: lineNum,
		IsJSON:     false,
		Raw:        line,
	}
}
