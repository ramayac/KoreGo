package tail

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunBasic(t *testing.T) {
	in := strings.NewReader("1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n11\n12\n")
	var out bytes.Buffer
	lines, err := Run(in, &out, 10, 0, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 10 {
		t.Errorf("expected 10 lines, got %d", len(lines))
	}
}

func TestRunShort(t *testing.T) {
	in := strings.NewReader("1\n2\n")
	var out bytes.Buffer
	lines, _ := Run(in, &out, 10, 0, false)
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d", len(lines))
	}
}

func TestRunFromStart(t *testing.T) {
	in := strings.NewReader("1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n11\n12\n")
	var out bytes.Buffer
	lines, err := Run(in, &out, 5, 0, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) < 6 {
		t.Errorf("expected at least 6 lines from +5, got %d", len(lines))
	}
	if lines[0] != "5" {
		t.Errorf("first line should be '5', got %q", lines[0])
	}
}

func TestRunBytesMode(t *testing.T) {
	in := strings.NewReader("abcdefghij")
	var out bytes.Buffer
	lines, err := Run(in, &out, 0, 5, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 1 {
		t.Errorf("expected 1 line from byte mode, got %d", len(lines))
	}
	if out.String() != "fghij" {
		t.Errorf("expected last 5 bytes 'fghij', got %q", out.String())
	}
}

func TestRunBytesFromStart(t *testing.T) {
	in := strings.NewReader("abcdefghij")
	var out bytes.Buffer
	_, err := Run(in, &out, 0, 3, true)
	if err != nil {
		t.Fatal(err)
	}
	if out.String() != "cdefghij" {
		t.Errorf("expected 'cdefghij' (skip first 2 bytes +c3), got %q", out.String())
	}
}

func TestRunBytesExact(t *testing.T) {
	in := strings.NewReader("abc")
	var out bytes.Buffer
	_, err := Run(in, &out, 0, 10, false)
	if err != nil {
		t.Fatal(err)
	}
	if out.String() != "abc" {
		t.Errorf("expected 'abc', got %q", out.String())
	}
}

func TestRunEmpty(t *testing.T) {
	in := strings.NewReader("")
	var out bytes.Buffer
	lines, err := Run(in, &out, 10, 0, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 0 {
		t.Errorf("expected 0 lines, got %d", len(lines))
	}
}

func TestRunFromStartShort(t *testing.T) {
	in := strings.NewReader("1\n2\n")
	var out bytes.Buffer
	lines, err := Run(in, &out, 10, 0, true)
	if err != nil {
		t.Fatal(err)
	}
	// +10 on 2 lines skips all, producing nothing
	if len(lines) != 0 {
		t.Errorf("expected 0 lines (skip 9 of 2), got %d", len(lines))
	}
}
