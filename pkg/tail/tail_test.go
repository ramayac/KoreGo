package tail

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunBasic(t *testing.T) {
	in := strings.NewReader("1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n11\n12\n")
	var out bytes.Buffer
	lines, err := Run(in, &out, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 10 {
		t.Errorf("expected 10 lines, got %d", len(lines))
	}
	if !strings.Contains(out.String(), "12") || strings.Contains(out.String(), "\n2\n") || strings.HasPrefix(out.String(), "2\n") {
		t.Errorf("output mismatch: %q", out.String())
	}
}

func TestRunShort(t *testing.T) {
	in := strings.NewReader("1\n2\n")
	var out bytes.Buffer
	lines, _ := Run(in, &out, 10)
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d", len(lines))
	}
}
