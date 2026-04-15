package logsource

import "testing"

func TestClassify_JSONObject(t *testing.T) {
	if Classify([]byte(`{"level":"info"}`)) != LineTypeJSONL {
		t.Error("expected LineTypeJSONL for JSON object")
	}
}

func TestClassify_PlainText(t *testing.T) {
	if Classify([]byte("plain text")) != LineTypeRaw {
		t.Error("expected LineTypeRaw for plain text")
	}
}

func TestClassify_EmptyLine(t *testing.T) {
	if Classify([]byte("")) != LineTypeRaw {
		t.Error("expected LineTypeRaw for empty line")
	}
}

func TestClassify_JSONArray(t *testing.T) {
	if Classify([]byte(`[1,2,3]`)) != LineTypeRaw {
		t.Error("expected LineTypeRaw for JSON array")
	}
}

func TestClassify_JSONScalar(t *testing.T) {
	if Classify([]byte(`"hello"`)) != LineTypeRaw {
		t.Error("expected LineTypeRaw for JSON scalar")
	}
}
