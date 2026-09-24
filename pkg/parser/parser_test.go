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

func TestFormatValue(t *testing.T) {
	// Whole-number float64s (how json.Decoder decodes integers into any)
	// must keep integer representation instead of scientific notation.
	cases := []struct {
		in   any
		want string
	}{
		{float64(735790403), "735790403"},
		{float64(0), "0"},
		{float64(-42), "-42"},
		{float64(1.5), "1.5"},
		{"text", "text"},
		{true, "true"},
		{nil, "null"},
	}

	for _, c := range cases {
		if got := FormatValue(c.in); got != c.want {
			t.Errorf("FormatValue(%v) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestProjectColumns(t *testing.T) {
	ds := &DataStructure{
		Headers: []string{"id", "status", "created", "owned_by"},
		Rows: [][]string{
			{"model-1", "ready", "100", "orgA"},
			{"model-2", "pending", "200", "orgB"},
		},
		Unwrapped: []any{
			map[string]any{"id": "model-1", "status": "ready", "created": 100, "owned_by": "orgA"},
			map[string]any{"id": "model-2", "status": "pending", "created": 200, "owned_by": "orgB"},
		},
	}

	// 1. Project subset of columns (case-insensitive)
	projected := ProjectColumns(ds, []string{"STATUS", "id"})
	if len(projected.Headers) != 2 {
		t.Fatalf("Expected 2 headers, got %v", projected.Headers)
	}
	if projected.Headers[0] != "status" || projected.Headers[1] != "id" {
		t.Errorf("Unexpected headers: %v", projected.Headers)
	}
	if len(projected.Rows) != 2 {
		t.Fatalf("Expected 2 rows, got %d", len(projected.Rows))
	}
	if projected.Rows[0][0] != "ready" || projected.Rows[0][1] != "model-1" {
		t.Errorf("Unexpected row 0: %v", projected.Rows[0])
	}

	// Verify unwrapped slice of maps projected
	slice, ok := projected.Unwrapped.([]any)
	if !ok || len(slice) != 2 {
		t.Fatalf("Expected unwrapped slice of 2 items")
	}
	m := slice[0].(map[string]any)
	if _, hasCreated := m["created"]; hasCreated {
		t.Errorf("'created' should have been projected away")
	}
	if m["status"] != "ready" || m["id"] != "model-1" {
		t.Errorf("Unexpected projected map content: %v", m)
	}

	// 2. Non-existent columns return original
	unchanged := ProjectColumns(ds, []string{"nonexistent"})
	if len(unchanged.Headers) != 4 {
		t.Errorf("Expected unchanged headers for nonexistent column, got %v", unchanged.Headers)
	}
}

func TestSortDataStructure(t *testing.T) {
	ds := &DataStructure{
		Headers: []string{"name", "score"},
		Rows: [][]string{
			{"alice", "95"},
			{"bob", "100"},
			{"charlie", "80"},
		},
		Unwrapped: []any{
			map[string]any{"name": "alice", "score": 95},
			map[string]any{"name": "bob", "score": 100},
			map[string]any{"name": "charlie", "score": 80},
		},
	}

	// 1. Numeric sort ascending
	sortedAsc := SortDataStructure(ds, "score", false)
	if sortedAsc.Rows[0][0] != "charlie" || sortedAsc.Rows[1][0] != "alice" || sortedAsc.Rows[2][0] != "bob" {
		t.Errorf("Ascending sort failed: %v", sortedAsc.Rows)
	}

	// 2. Numeric sort descending
	sortedDesc := SortDataStructure(ds, "score", true)
	if sortedDesc.Rows[0][0] != "bob" || sortedDesc.Rows[1][0] != "alice" || sortedDesc.Rows[2][0] != "charlie" {
		t.Errorf("Descending sort failed: %v", sortedDesc.Rows)
	}

	// 3. String sort ascending
	sortedAlpha := SortDataStructure(ds, "name", false)
	if sortedAlpha.Rows[0][0] != "alice" || sortedAlpha.Rows[1][0] != "bob" || sortedAlpha.Rows[2][0] != "charlie" {
		t.Errorf("String sort failed: %v", sortedAlpha.Rows)
	}

	// 4. Sort by 1-based index (2 = score) descending
	sortedCol2 := SortDataStructure(ds, "2", true)
	if sortedCol2.Rows[0][0] != "bob" || sortedCol2.Rows[2][0] != "charlie" {
		t.Errorf("Index 2 descending sort failed: %v", sortedCol2.Rows)
	}
}

func TestResolveColumnIndex(t *testing.T) {
	headers := []string{"created", "id", "object", "owned_by"}

	// Direct match
	if idx := ResolveColumnIndex(headers, "id"); idx != 1 {
		t.Errorf("Expected 1 for 'id', got %d", idx)
	}

	// 1-based index
	if idx := ResolveColumnIndex(headers, "2"); idx != 1 {
		t.Errorf("Expected 1 for '2', got %d", idx)
	}

	// Non-existent column returns -1 (no silent alias guessing)
	if idx := ResolveColumnIndex(headers, "name"); idx != -1 {
		t.Errorf("Expected -1 for 'name' when not in headers, got %d", idx)
	}

	// Non-existent
	if idx := ResolveColumnIndex(headers, "nonexistent"); idx != -1 {
		t.Errorf("Expected -1 for nonexistent, got %d", idx)
	}
}

