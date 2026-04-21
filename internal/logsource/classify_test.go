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

func TestClassify_JSONArray(t *testing.T) {
	assert.Equal(t, LineTypeRaw, Classify([]byte(`[1,2,3]`)),
		"expected LineTypeRaw for JSON array")
}

func TestClassify_JSONScalar(t *testing.T) {
	assert.Equal(t, LineTypeRaw, Classify([]byte(`"hello"`)),
		"expected LineTypeRaw for JSON scalar")
}
