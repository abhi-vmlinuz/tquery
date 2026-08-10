package filter

import (
	"encoding/json"
	"testing"
)

func TestPruneJSON(t *testing.T) {
	dockerJSON := []byte(`{
		"Id": "123456",
		"State": {
			"Running": true,
			"Status": "running"
		},
		"HostConfig": {
			"PortBindings": {
				"1521/tcp": [{"HostPort": "1521"}]
			}
		},
		"NetworkSettings": {
			"Ports": {
				"1521/tcp": [{"HostPort": "1521"}]
			}
		}
	}`)

	var data any
	if err := json.Unmarshal(dockerJSON, &data); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	pruned, matched := PruneJSON(data, "port", true, false)
	if !matched {
		t.Fatalf("Expected match for 'port'")
	}

	m, ok := pruned.(map[string]any)
	if !ok {
		t.Fatalf("Expected map result, got %T", pruned)
	}

	// HostConfig and NetworkSettings should remain because they contain PortBindings and Ports
	if _, exists := m["HostConfig"]; !exists {
		t.Errorf("Expected HostConfig in pruned tree")
	}
	if _, exists := m["NetworkSettings"]; !exists {
		t.Errorf("Expected NetworkSettings in pruned tree")
	}

	// State and Id should be pruned away!
	if _, exists := m["Id"]; exists {
		t.Errorf("Id should have been pruned away")
	}
	if _, exists := m["State"]; exists {
		t.Errorf("State should have been pruned away")
	}
}
