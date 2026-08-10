package render

import (
	"bytes"
	"strings"
	"testing"

	"github.com/tquery/tquery/pkg/parser"
)

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
