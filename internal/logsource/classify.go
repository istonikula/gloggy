package logsource

import "encoding/json"

// LineType indicates whether a log line is structured JSON or raw text.
type LineType int

const (
	LineTypeRaw   LineType = iota
	LineTypeJSONL
)

// Classify determines whether a line is a JSON object (JSONL) or raw text.
// A line is JSONL only if it starts with '{' and is a valid JSON object.
func Classify(line []byte) LineType {
	if len(line) == 0 {
		return LineTypeRaw
	}
	if line[0] != '{' {
		return LineTypeRaw
	}
	var obj map[string]json.RawMessage
	if json.Unmarshal(line, &obj) != nil {
		return LineTypeRaw
	}
	return LineTypeJSONL
}
