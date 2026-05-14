package cut

import (
	"strings"
	"testing"
)

func TestCutFields(t *testing.T) {
	in := "a:b:c\n1:2:3\n"
	lines, _ := Run(strings.NewReader(in), "1,3", ":", "", "", false, false, false)
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines")
	}
	if lines[0].Fields[0] != "a:c" {
		t.Errorf("got %v", lines[0].Fields[0])
	}
}

func TestCutChars(t *testing.T) {
	in := "abcdef\n"
	lines, _ := Run(strings.NewReader(in), "", "", "1-3,5", "", false, false, false)
	if lines[0].Fields[0] != "abce" {
		t.Errorf("got %v", lines[0].Fields[0])
	}
}

func TestCutBytes(t *testing.T) {
	in := "hello\n"
	lines, _ := Run(strings.NewReader(in), "", "", "", "2", false, false, false)
	if lines[0].Fields[0] != "e" {
		t.Errorf("got %v", lines[0].Fields[0])
	}
}

func TestParseList(t *testing.T) {
	cases := []struct {
		in    string
		count int
		valid bool
	}{
		{"1", 1, true},
		{"1,3,5", 3, true},
		{"1-5", 1, true},
		{"1-", 1, true},
		{"-5", 1, true},
		{"1-3,5-7", 2, true},
		{"", 0, false},
		{"0", 0, false},
	}
	for _, c := range cases {
		specs, err := parseList(c.in)
		if c.valid && err != nil {
			t.Errorf("parseList(%q) unexpected error: %v", c.in, err)
		}
		if !c.valid && err == nil {
			t.Errorf("parseList(%q) expected error, got %v", c.in, specs)
		}
		if c.valid && len(specs) != c.count {
			t.Errorf("parseList(%q) = %d specs, want %d", c.in, len(specs), c.count)
		}
	}
}

func TestInRange(t *testing.T) {
	specs := []rangeSpec{{1, 3}, {5, 5}}
	if !inRange(1, specs) {
		t.Error("1 should be in range")
	}
	if !inRange(3, specs) {
		t.Error("3 should be in range")
	}
	if inRange(4, specs) {
		t.Error("4 should NOT be in range")
	}
	if !inRange(5, specs) {
		t.Error("5 should be in range")
	}
}

func TestCutFieldsDelimiter(t *testing.T) {
	in := "a,b,c\nd,e,f\n"
	lines, _ := Run(strings.NewReader(in), "2", ",", "", "", false, false, false)
	if lines[0].Fields[0] != "b" {
		t.Errorf("got %v", lines[0].Fields[0])
	}
}

func TestCutOnlyDelimited(t *testing.T) {
	in := "a:b\nno-delim\nc:d\n"
	lines, _ := Run(strings.NewReader(in), "1", ":", "", "", true, false, false)
	if len(lines) != 2 {
		t.Errorf("expected 2 lines with -s, got %d", len(lines))
	}
}

func TestCutCharsRange(t *testing.T) {
	in := "12345\n"
	lines, _ := Run(strings.NewReader(in), "", "", "2-4", "", false, false, false)
	if lines[0].Fields[0] != "234" {
		t.Errorf("got %v", lines[0].Fields[0])
	}
}

func TestCutEmpty(t *testing.T) {
	lines, err := Run(strings.NewReader(""), "1", "\t", "", "", false, false, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 0 {
		t.Errorf("expected 0 lines, got %d", len(lines))
	}
}

func TestCutWhitespaceFields(t *testing.T) {
	in := "hello   world\nfoo bar\n"
	lines, _ := Run(strings.NewReader(in), "2", "", "", "", false, false, true)
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines")
	}
	if lines[0].Fields[0] != "world" {
		t.Errorf("got %q", lines[0].Fields[0])
	}
	if lines[1].Fields[0] != "bar" {
		t.Errorf("got %q", lines[1].Fields[0])
	}
}
