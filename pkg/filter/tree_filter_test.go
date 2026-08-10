package filter

import (
	"encoding/json"
	"testing"
)

func TestPruneJSONMulti_OR(t *testing.T) {
	jsonData := []byte(`{
		"image": "nginx:alpine",
		"status": "Running",
		"port": 8080
	}`)

	var data any
	if err := json.Unmarshal(jsonData, &data); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	opts := MultiOptions{
		Patterns:   []string{"nginx", "running"},
		IgnoreCase: true,
		Strict:     false,
	}

	pruned, matched := PruneJSON(data, opts)
	if !matched {
		t.Fatalf("Expected match in OR mode")
	}

	m, ok := pruned.(map[string]any)
	if !ok {
		t.Fatalf("Expected map, got %T", pruned)
	}

	if _, exists := m["image"]; !exists {
		t.Errorf("Expected 'image' in pruned result")
	}
	if _, exists := m["status"]; !exists {
		t.Errorf("Expected 'status' in pruned result")
	}
	if _, exists := m["port"]; exists {
		t.Errorf("'port' should have been pruned away")
	}
}

func TestPruneJSONMulti_Strict(t *testing.T) {
	jsonData := []byte(`{
		"name": "nginx",
		"status": "Running",
		"port": 8080
	}`)

	var data any
	if err := json.Unmarshal(jsonData, &data); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	// 1. Both patterns match -> should succeed
	optsSuccess := MultiOptions{
		Patterns:   []string{"nginx", "running"},
		IgnoreCase: true,
		Strict:     true,
	}
	pruned, matched := PruneJSON(data, optsSuccess)
	if !matched {
		t.Fatalf("Expected match in Strict mode when both patterns exist")
	}
	m := pruned.(map[string]any)
	if _, exists := m["name"]; !exists {
		t.Errorf("Expected 'name' in result")
	}
	if _, exists := m["status"]; !exists {
		t.Errorf("Expected 'status' in result")
	}
	if _, exists := m["port"]; exists {
		t.Errorf("'port' should have been pruned")
	}

	// 2. Only one pattern matches -> should fail in strict mode
	optsFail := MultiOptions{
		Patterns:   []string{"nginx", "stopped"},
		IgnoreCase: true,
		Strict:     true,
	}
	_, matchedFail := PruneJSON(data, optsFail)
	if matchedFail {
		t.Fatalf("Strict mode should fail when only one pattern matches")
	}
}

func TestPruneJSONMulti_ArrayStrictRecords(t *testing.T) {
	arrayJSON := []byte(`[
		{"name": "nginx", "status": "Running"},
		{"name": "nginx", "status": "Stopped"},
		{"name": "redis", "status": "Running"}
	]`)

	var data any
	if err := json.Unmarshal(arrayJSON, &data); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	opts := MultiOptions{
		Patterns:   []string{"nginx", "running"},
		IgnoreCase: true,
		Strict:     true,
	}

	pruned, matched := PruneJSON(data, opts)
	if !matched {
		t.Fatalf("Expected match for array with strict mode")
	}

	slice, ok := pruned.([]any)
	if !ok {
		t.Fatalf("Expected slice result, got %T", pruned)
	}

	if len(slice) != 1 {
		t.Fatalf("Expected exactly 1 matching record in strict mode, got %d", len(slice))
	}

	rec := slice[0].(map[string]any)
	if rec["name"] != "nginx" || rec["status"] != "Running" {
		t.Errorf("Unexpected record content: %v", rec)
	}
}
