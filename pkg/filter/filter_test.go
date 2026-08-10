package filter

import (
	"testing"

	"github.com/tquery/tquery/pkg/parser"
)

func TestFilterLiteral(t *testing.T) {
	ds := &parser.DataStructure{
		Headers: []string{"id", "status"},
		Rows: [][]string{
			{"node-1", "Running"},
			{"node-2", "Stopped"},
			{"node-3", "Running"},
		},
		Unwrapped: []any{
			map[string]any{"id": "node-1", "status": "Running"},
			map[string]any{"id": "node-2", "status": "Stopped"},
			map[string]any{"id": "node-3", "status": "Running"},
		},
	}

	filtered, err := Filter(ds, MultiOptions{Patterns: []string{"Running"}})
	if err != nil {
		t.Fatalf("Filter error: %v", err)
	}

	if len(filtered.Rows) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(filtered.Rows))
	}
}

func TestFilterMulti_OR_and_Strict(t *testing.T) {
	ds := &parser.DataStructure{
		Headers: []string{"name", "status"},
		Rows: [][]string{
			{"nginx", "Running"},
			{"nginx", "Stopped"},
			{"redis", "Running"},
		},
	}

	// 1. OR mode
	filteredOR, err := Filter(ds, MultiOptions{
		Patterns: []string{"nginx", "Running"},
		Strict:   false,
	})
	if err != nil {
		t.Fatalf("OR Filter error: %v", err)
	}
	if len(filteredOR.Rows) != 3 {
		t.Errorf("Expected 3 rows in OR mode, got %d", len(filteredOR.Rows))
	}

	// 2. Strict mode (AND)
	filteredStrict, err := Filter(ds, MultiOptions{
		Patterns: []string{"nginx", "Running"},
		Strict:   true,
	})
	if err != nil {
		t.Fatalf("Strict Filter error: %v", err)
	}
	if len(filteredStrict.Rows) != 1 {
		t.Fatalf("Expected 1 row in Strict mode, got %d", len(filteredStrict.Rows))
	}
	if filteredStrict.Rows[0][0] != "nginx" || filteredStrict.Rows[0][1] != "Running" {
		t.Errorf("Unexpected row in strict mode: %v", filteredStrict.Rows[0])
	}
}

func TestFilterInvert(t *testing.T) {
	ds := &parser.DataStructure{
		Headers: []string{"id", "status"},
		Rows: [][]string{
			{"node-1", "Running"},
			{"node-2", "Stopped"},
			{"node-3", "Running"},
		},
	}

	filtered, err := Filter(ds, MultiOptions{Patterns: []string{"Stopped"}, InvertMatch: true})
	if err != nil {
		t.Fatalf("Filter error: %v", err)
	}

	if len(filtered.Rows) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(filtered.Rows))
	}
}

func TestFilterColumnScoped(t *testing.T) {
	ds := &parser.DataStructure{
		Headers: []string{"name", "role"},
		Rows: [][]string{
			{"admin-user", "guest"},
			{"john-doe", "admin"},
		},
	}

	filtered, err := Filter(ds, MultiOptions{Patterns: []string{"role:admin"}})
	if err != nil {
		t.Fatalf("Filter error: %v", err)
	}

	if len(filtered.Rows) != 1 {
		t.Fatalf("Expected 1 row, got %d", len(filtered.Rows))
	}

	if filtered.Rows[0][0] != "john-doe" {
		t.Errorf("Expected john-doe, got %s", filtered.Rows[0][0])
	}
}

func TestFilterRegex(t *testing.T) {
	ds := &parser.DataStructure{
		Headers: []string{"id"},
		Rows: [][]string{
			{"model-v1.0"},
			{"model-v2.5"},
			{"other-app"},
		},
	}

	filtered, err := Filter(ds, MultiOptions{Patterns: []string{`^model-v\d+\.\d+`}})
	if err != nil {
		t.Fatalf("Filter error: %v", err)
	}

	if len(filtered.Rows) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(filtered.Rows))
	}
}
