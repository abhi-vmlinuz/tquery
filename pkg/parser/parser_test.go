package parser

import (
	"reflect"
	"testing"
)

func TestParseArrayOfObjects(t *testing.T) {
	jsonInput := []byte(`[
		{"id": "model-1", "owned_by": "gonka", "created": 1677610602},
		{"id": "model-2", "owned_by": "gonka", "created": 1677610603}
	]`)

	ds, err := Parse(jsonInput, true)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if ds.Type != KindTable {
		t.Errorf("Expected KindTable, got %s", ds.Type)
	}

	expectedHeaders := []string{"created", "id", "owned_by"}
	if !reflect.DeepEqual(ds.Headers, expectedHeaders) {
		t.Errorf("Headers mismatch. Got %v, expected %v", ds.Headers, expectedHeaders)
	}

	if len(ds.Rows) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(ds.Rows))
	}
}

func TestParseAutoUnwrapRoot(t *testing.T) {
	jsonInput := []byte(`{
		"object": "list",
		"data": [
			{"id": "MiniMaxAI/MiniMax-M2.7", "object": "model", "created": 1677610602, "owned_by": "gonka"},
			{"id": "moonshotai/Kimi-K2.6", "object": "model", "created": 1677610602, "owned_by": "gonka"}
		]
	}`)

	ds, err := Parse(jsonInput, true)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if ds.Type != KindTable {
		t.Errorf("Expected auto-unwrapped KindTable, got %s", ds.Type)
	}

	if len(ds.Rows) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(ds.Rows))
	}

	// Verify first row contains the model id
	foundID := false
	for _, cell := range ds.Rows[0] {
		if cell == "MiniMaxAI/MiniMax-M2.7" {
			foundID = true
			break
		}
	}
	if !foundID {
		t.Errorf("Expected row to contain MiniMaxAI model ID")
	}
}

func TestParseKeyValueObject(t *testing.T) {
	jsonInput := []byte(`{"name": "tquery", "version": "1.0.0", "active": true}`)

	ds, err := Parse(jsonInput, false)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if ds.Type != KindKeyValue {
		t.Errorf("Expected KindKeyValue, got %s", ds.Type)
	}

	if len(ds.Rows) != 3 {
		t.Errorf("Expected 3 rows, got %d", len(ds.Rows))
	}
}
