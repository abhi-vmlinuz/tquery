package engine

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestEvaluateSimple(t *testing.T) {
	raw := []byte(`{
		"data": [
			{"id": "model-a", "created": 100},
			{"id": "model-b", "created": 200}
		]
	}`)

	var input any
	if err := json.Unmarshal(raw, &input); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	result, err := Evaluate(".data[].id", input)
	if err != nil {
		t.Fatalf("Evaluate error: %v", err)
	}

	expected := []any{"model-a", "model-b"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Got %v, expected %v", result, expected)
	}
}

func TestEvaluateFilter(t *testing.T) {
	raw := []byte(`{"models": [{"name": "gpt", "score": 90}, {"name": "claude", "score": 95}]}`)

	resultBytes, _, err := EvaluateRaw(`.models[] | select(.score > 90) | .name`, raw)
	if err != nil {
		t.Fatalf("EvaluateRaw error: %v", err)
	}

	if string(resultBytes) != `"claude"` {
		t.Errorf("Got %s, expected \"claude\"", string(resultBytes))
	}
}
