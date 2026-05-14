package head

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
	if !strings.Contains(out.String(), "10") || strings.Contains(out.String(), "11") {
		t.Error("output mismatch")
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

func TestRunBytes(t *testing.T) {
	in := strings.NewReader("abcdefghij")
	var out bytes.Buffer
	lines, err := runBytes(in, &out, 5)
	if err != nil {
		t.Fatal(err)
	}
	if lines != nil {
		t.Error("runBytes should return nil lines")
	}
	if out.String() != "abcde" {
		t.Errorf("expected 'abcde', got %q", out.String())
	}
}

func TestRunBytesExact(t *testing.T) {
	in := strings.NewReader("abc")
	var out bytes.Buffer
	_, err := runBytes(in, &out, 5)
	if err != nil {
		t.Fatal(err)
	}
	if out.String() != "abc" {
		t.Errorf("expected 'abc', got %q", out.String())
	}
}

func TestRunBytesEmpty(t *testing.T) {
	in := strings.NewReader("")
	var out bytes.Buffer
	_, err := runBytes(in, &out, 5)
	if err != nil {
		t.Fatal(err)
	}
	if out.String() != "" {
		t.Errorf("expected empty, got %q", out.String())
	}
}

func TestRunNegative(t *testing.T) {
	in := strings.NewReader("1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n")
	var out bytes.Buffer
	lines, err := runNegative(in, &out, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 8 {
		t.Errorf("expected 8 lines (10-2), got %d", len(lines))
	}
}

func TestRunNegativeSkipAll(t *testing.T) {
	in := strings.NewReader("a\nb\n")
	var out bytes.Buffer
	lines, err := runNegative(in, &out, 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 0 {
		t.Errorf("expected 0 lines, got %d", len(lines))
	}
}

func TestRunZeroLines(t *testing.T) {
	in := strings.NewReader("1\n2\n3\n")
	var out bytes.Buffer
	lines, _ := Run(in, &out, 0)
	if len(lines) != 0 {
		t.Errorf("expected 0 lines, got %d", len(lines))
	}
}

func TestRunEmpty(t *testing.T) {
	in := strings.NewReader("")
	var out bytes.Buffer
	lines, err := Run(in, &out, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 0 {
		t.Errorf("expected 0 lines, got %d", len(lines))
	}
}
