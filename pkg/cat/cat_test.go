package cat

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunBasic(t *testing.T) {
	in := strings.NewReader("hello\nworld\n")
	var out bytes.Buffer
	lines, err := Run(in, &out, false, false, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d", len(lines))
	}
	if !strings.Contains(out.String(), "hello") {
		t.Error("output missing 'hello'")
	}
}

func TestRunNumberAll(t *testing.T) {
	in := strings.NewReader("a\nb\n")
	var out bytes.Buffer
	lines, _ := Run(in, &out, true, false, false)
	if !strings.HasPrefix(lines[0], "     1\t") {
		t.Errorf("expected line numbers, got %q", lines[0])
	}
}

func TestRunNumberNonBlank(t *testing.T) {
	in := strings.NewReader("a\n\nb\n")
	var out bytes.Buffer
	lines, _ := Run(in, &out, false, true, false)
	// blank line should not have a number
	if strings.HasPrefix(lines[1], " ") && strings.Contains(lines[1], "\t") {
		t.Error("blank line should not be numbered with -b")
	}
}

func TestRunSqueezeBlank(t *testing.T) {
	in := strings.NewReader("a\n\n\nb\n")
	var out bytes.Buffer
	lines, _ := Run(in, &out, false, false, true)
	// Two consecutive blanks should collapse to one.
	blanks := 0
	for _, l := range lines {
		if strings.TrimSpace(l) == "" {
			blanks++
		}
	}
	if blanks > 1 {
		t.Errorf("squeeze-blank: expected ≤1 blank line, got %d", blanks)
	}
}

func TestRunEmpty(t *testing.T) {
	in := strings.NewReader("")
	var out bytes.Buffer
	lines, err := Run(in, &out, false, false, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 0 {
		t.Errorf("expected 0 lines for empty input, got %d", len(lines))
	}
}
