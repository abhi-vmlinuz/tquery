package cmd

import (
	"os"
	"testing"
)

func TestNormalizeFormat(t *testing.T) {
	cases := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{"auto", "auto", false},
		{"", "auto", false},
		{"table", "table", false},
		{"tree", "tree", false},
		{"json", "json", false},
		{"raw", "json", false},
		{"markdown", "markdown", false},
		// Documented alias for --markdown must not fall back to table
		{"md", "markdown", false},
		{"csv", "csv", false},
		{"tsv", "tsv", false},
		// Case and surrounding whitespace are tolerated
		{"MD", "markdown", false},
		{" Tree ", "tree", false},
		// Unknown formats must fail loudly
		{"xml", "", true},
		{"yml", "", true},
	}

	for _, c := range cases {
		got, err := normalizeFormat(c.in)
		if (err != nil) != c.wantErr {
			t.Errorf("normalizeFormat(%q) error = %v, wantErr %v", c.in, err, c.wantErr)
			continue
		}
		if err == nil && got != c.want {
			t.Errorf("normalizeFormat(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestParseFlags(t *testing.T) {
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	os.Args = []string{"tq", "-c", "id,status", "-s", "-created", "-cb", "-5"}
	cfg := parseFlags()

	if len(cfg.Columns) != 2 || cfg.Columns[0] != "id" || cfg.Columns[1] != "status" {
		t.Errorf("Columns parse failed, got %v", cfg.Columns)
	}
	if cfg.SortBy != "created" || !cfg.SortDesc {
		t.Errorf("Sort parse failed, got sortBy=%q, sortDesc=%v", cfg.SortBy, cfg.SortDesc)
	}
	if !cfg.Clipboard {
		t.Errorf("Clipboard flag not set")
	}
	if cfg.Limit != 5 {
		t.Errorf("Limit parse failed, got %d", cfg.Limit)
	}

	// Test bare -s
	os.Args = []string{"tq", "-s"}
	cfg = parseFlags()
	if !cfg.SortEnabled || cfg.SortBy != "" {
		t.Errorf("Bare -s failed: SortEnabled=%v, SortBy=%q", cfg.SortEnabled, cfg.SortBy)
	}

	// Test bare --desc
	os.Args = []string{"tq", "--desc"}
	cfg = parseFlags()
	if !cfg.SortEnabled || !cfg.SortDesc {
		t.Errorf("Bare --desc failed: SortEnabled=%v, SortDesc=%v", cfg.SortEnabled, cfg.SortDesc)
	}

	// Test -c 2 --desc
	os.Args = []string{"tq", "-c", "2", "--desc"}
	cfg = parseFlags()
	if len(cfg.Columns) != 1 || cfg.Columns[0] != "2" || !cfg.SortEnabled || !cfg.SortDesc {
		t.Errorf("-c 2 --desc failed: %v", cfg)
	}

	// Test findPrimaryColumn
	headers := []string{"created", "id", "object", "owned_by"}
	if col := findPrimaryColumn(headers); col != "id" {
		t.Errorf("Expected primary column 'id', got %q", col)
	}
}

