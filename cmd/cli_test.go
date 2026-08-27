package cmd

import "testing"

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
