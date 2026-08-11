package parser

import (
	"reflect"
	"strings"
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

func TestParseJSONL(t *testing.T) {
	jsonlInput := []byte("{\"ID\":\"abc\",\"Image\":\"nginx\",\"Status\":\"running\"}\n{\"ID\":\"def\",\"Image\":\"redis\",\"Status\":\"running\"}\n")

	ds, err := Parse(jsonlInput, true)
	if err != nil {
		t.Fatalf("Parse error on JSONL: %v", err)
	}

	if ds.Type != KindTable {
		t.Errorf("Expected KindTable for NDJSON stream of maps, got %s", ds.Type)
	}

	if len(ds.Rows) != 2 {
		t.Fatalf("Expected 2 rows from NDJSON stream, got %d", len(ds.Rows))
	}

	expectedHeaders := []string{"ID", "Image", "Status"}
	if !reflect.DeepEqual(ds.Headers, expectedHeaders) {
		t.Errorf("Headers mismatch. Got %v, expected %v", ds.Headers, expectedHeaders)
	}
}

func TestParseConcatenatedJSON(t *testing.T) {
	concatInput := []byte(`{"a": 1} {"a": 2} {"a": 3}`)

	ds, err := Parse(concatInput, true)
	if err != nil {
		t.Fatalf("Parse error on concatenated JSON: %v", err)
	}

	if ds.Type != KindTable {
		t.Errorf("Expected KindTable, got %s", ds.Type)
	}

	if len(ds.Rows) != 3 {
		t.Errorf("Expected 3 rows from concatenated JSON, got %d", len(ds.Rows))
	}
}

func TestParseInvalidJSONStream(t *testing.T) {
	invalidInput := []byte("{\"foo\":\"bar\"}\nNOT_VALID_JSON\n{\"foo\":\"baz\"}")

	_, err := Parse(invalidInput, true)
	if err == nil {
		t.Fatalf("Expected error for invalid JSON stream, got nil")
	}

	if !strings.Contains(err.Error(), "invalid json stream") {
		t.Errorf("Expected error to mention 'invalid json stream', got: %v", err)
	}
}
