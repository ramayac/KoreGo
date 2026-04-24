package cut

import (
	"strings"
	"testing"
)

func TestCutFields(t *testing.T) {
	in := "a:b:c\n1:2:3\n"
	lines, _ := Run(strings.NewReader(in), "1,3", ":", "", "")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines")
	}
	if lines[0].Fields[0] != "a:c" {
		t.Errorf("got %v", lines[0].Fields[0])
	}
}

func TestCutChars(t *testing.T) {
	in := "abcdef\n"
	lines, _ := Run(strings.NewReader(in), "", "", "1-3,5", "")
	if lines[0].Fields[0] != "abce" {
		t.Errorf("got %v", lines[0].Fields[0])
	}
}

func TestCutBytes(t *testing.T) {
	in := "hello\n"
	lines, _ := Run(strings.NewReader(in), "", "", "", "2")
	if lines[0].Fields[0] != "e" {
		t.Errorf("got %v", lines[0].Fields[0])
	}
}
