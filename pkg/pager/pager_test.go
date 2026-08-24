package pager

import (
	"bytes"
	"os"
	"testing"
)

func TestShouldPageNoPagerFlag(t *testing.T) {
	if ShouldPage(100, true) {
		t.Errorf("ShouldPage should return false when noPager is true")
	}
}

func TestShouldPageZeroLines(t *testing.T) {
	if ShouldPage(0, false) {
		t.Errorf("ShouldPage should return false for 0 lines")
	}
}

func TestWriteOrPageNonTTY(t *testing.T) {
	var buf bytes.Buffer
	content := "line 1\nline 2\nline 3\n"

	err := WriteOrPage(&buf, content, false)
	if err != nil {
		t.Fatalf("WriteOrPage failed: %v", err)
	}

	if buf.String() != content {
		t.Errorf("Expected %q, got %q", content, buf.String())
	}
}

func TestCustomSystemPager(t *testing.T) {
	os.Setenv("TQ_PAGER", "cat")
	defer os.Unsetenv("TQ_PAGER")

	err := runSystemPager(os.Getenv("TQ_PAGER"), "test content\n")
	if err != nil {
		t.Errorf("runSystemPager failed for cat: %v", err)
	}
}
