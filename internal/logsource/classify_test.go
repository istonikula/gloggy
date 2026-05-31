package logsource

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClassify_JSONObject(t *testing.T) {
	assert.Equal(t, LineTypeJSONL, Classify([]byte(`{"level":"info"}`)),
		"expected LineTypeJSONL for JSON object")
}

func TestClassify_PlainText(t *testing.T) {
	assert.Equal(t, LineTypeRaw, Classify([]byte("plain text")),
		"expected LineTypeRaw for plain text")
}

func TestClassify_EmptyLine(t *testing.T) {
	assert.Equal(t, LineTypeRaw, Classify([]byte("")),
		"expected LineTypeRaw for empty line")
}

// B43: Classify indexes line[0] after a len==0 guard. Callers strip the
// trailing "\n" before classifying, so an input that was exactly "\n" arrives
// here as a zero-length slice. Assert nil and a newline-stripped empty slice
// are handled defensively (LineTypeRaw, no panic).
func TestClassify_NilAndNewlineStripped_B43(t *testing.T) {
	assert.NotPanics(t, func() { Classify(nil) }, "nil slice must not panic")
	assert.Equal(t, LineTypeRaw, Classify(nil), "nil ⇒ raw")

	stripped := []byte("\n")[:0] // what a caller's strip("\n") yields: len 0
	assert.Equal(t, LineTypeRaw, Classify(stripped), "newline-stripped empty ⇒ raw")
}

func TestClassify_JSONArray(t *testing.T) {
	assert.Equal(t, LineTypeRaw, Classify([]byte(`[1,2,3]`)),
		"expected LineTypeRaw for JSON array")
}

func TestClassify_JSONScalar(t *testing.T) {
	assert.Equal(t, LineTypeRaw, Classify([]byte(`"hello"`)),
		"expected LineTypeRaw for JSON scalar")
}
