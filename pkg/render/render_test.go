package render

import (
	"bytes"
	"strings"
	"testing"

	"github.com/tquery/tquery/pkg/parser"
)

func TestAllocateColumnWidths(t *testing.T) {
	naturalWidths := []int{10, 300, 15}
	headers := []string{"ID", "Labels", "Status"}

	// Terminal width 80 -> available content budget = 80 - 10 = 70
	allocated := allocateColumnWidths(naturalWidths, headers, 80)
	if len(allocated) != 3 {
		t.Fatalf("Expected 3 allocated column widths, got %d", len(allocated))
	}

	// ID and Status should retain their small natural sizes, Labels should absorb the reduction
	if allocated[0] > 10 {
		t.Errorf("ID width should be <= 10, got %d", allocated[0])
	}
	if allocated[1] >= 300 {
		t.Errorf("Labels width should be truncated from 300, got %d", allocated[1])
	}
	if allocated[2] > 15 {
		t.Errorf("Status width should be <= 15, got %d", allocated[2])
	}
}

func TestTruncateString(t *testing.T) {
	longStr := "io.modelcontextprotocol.server.name=github-mcp-server"
	truncated := truncateString(longStr, 20)
	if len([]rune(truncated)) > 20 {
		t.Errorf("Truncated string visual width exceeded 20 runes: %q", truncated)
	}
	if !strings.HasSuffix(truncated, "…") {
		t.Errorf("Expected ellipsis suffix, got %q", truncated)
	}
}

func TestTruncateCellWithMatch(t *testing.T) {
	val := "prefix_data_something_long_very_deep_running_container_state"
	patterns := []string{"running"}

	truncated := truncateCellWithMatch(val, 25, patterns, true)
	if !strings.Contains(truncated, "running") {
		t.Errorf("Truncation should keep matched keyword visible, got: %q", truncated)
	}
}

func TestRenderMarkdown(t *testing.T) {
	ds := &parser.DataStructure{
		Headers: []string{"id", "owned_by"},
		Rows: [][]string{
			{"MiniMax-M2.7", "gonka"},
			{"Kimi-K2.6", "gonka"},
		},
	}

	var buf bytes.Buffer
	opts := RenderOptions{Format: FormatMarkdown, ShowHeader: true}
	if err := Render(&buf, ds, opts); err != nil {
		t.Fatalf("Render error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "| id | owned_by |") {
		t.Errorf("Expected markdown header in output, got:\n%s", out)
	}
	if !strings.Contains(out, "| MiniMax-M2.7 | gonka |") {
		t.Errorf("Expected row in markdown output, got:\n%s", out)
	}
}

func TestRenderCSV(t *testing.T) {
	ds := &parser.DataStructure{
		Headers: []string{"id", "owned_by"},
		Rows: [][]string{
			{"MiniMax-M2.7", "gonka"},
		},
	}

	var buf bytes.Buffer
	opts := RenderOptions{Format: FormatCSV, ShowHeader: true}
	if err := Render(&buf, ds, opts); err != nil {
		t.Fatalf("Render error: %v", err)
	}

	out := buf.String()
	expected := "id,owned_by\nMiniMax-M2.7,gonka\n"
	if out != expected {
		t.Errorf("Expected %q, got %q", expected, out)
	}
}

func TestBuildTreePreservesIntegerFormat(t *testing.T) {
	data := map[string]any{
		"created": float64(735790403),
		"id":      "01-ai/yi-large",
		"ratio":   float64(1.5),
	}

	out := BuildTree(data, false, nil, false)

	if strings.Contains(out, "e+08") || strings.Contains(out, "e+09") {
		t.Errorf("tree output contains scientific notation:\n%s", out)
	}
	for _, want := range []string{"created: 735790403", "ratio: 1.5"} {
		if !strings.Contains(out, want) {
			t.Errorf("tree output missing %q:\n%s", want, out)
		}
	}
}
