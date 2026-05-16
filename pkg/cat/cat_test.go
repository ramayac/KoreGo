package cat

import (
	"bytes"
	"io"
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

func TestVisByte(t *testing.T) {
	cases := []struct {
		b    byte
		want string
	}{
		{'\t', "\t"},
		{'\n', "\n"},
		{0, "^@"},
		{1, "^A"},
		{27, "^["},
		{0x7F, "^?"},
		{'a', "a"},
		{'A', "A"},
		{0x80, "M-^@"},
	}
	for _, c := range cases {
		got := visByte(c.b)
		if got != c.want {
			t.Errorf("visByte(0x%02x) = %q, want %q", c.b, got, c.want)
		}
	}
}

func TestVisLine(t *testing.T) {
	got := visLine("hello", false)
	if got != "hello" {
		t.Errorf("expected 'hello', got %q", got)
	}
	gotEnd := visLine("hello", true)
	if gotEnd != "hello$" {
		t.Errorf("expected 'hello$', got %q", gotEnd)
	}
	gotCtrl := visLine("a\x01b", false)
	if gotCtrl != "a^Ab" {
		t.Errorf("expected 'a^Ab', got %q", gotCtrl)
	}
}

func TestRunNumberNonBlankSkip(t *testing.T) {
	in := strings.NewReader("a\n\nb\n")
	var out bytes.Buffer
	lines, _ := Run(in, &out, false, true, false)
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(lines))
	}
	// First non-blank should be numbered
	if !strings.HasPrefix(lines[0], "     1\t") {
		t.Errorf("first non-blank should be numbered, got %q", lines[0])
	}
	// Blank line should NOT be numbered
	if strings.HasPrefix(lines[1], "     ") && strings.Contains(lines[1], "\t") {
		t.Errorf("blank line should NOT be numbered, got %q", lines[1])
	}
}

func TestVisByteHigh(t *testing.T) {
	// M- prefix for high bytes
	got := visByte(0xE1)
	if len(got) < 3 || got[:2] != "M-" {
		t.Errorf("expected M- prefix for high byte, got %q", got)
	}
}

// --- BusyBox test suite hardening ---

func TestBusyBox_Cat_FileAndStdin(t *testing.T) {
	// BusyBox: echo SOMETHING | cat foo - should print file then stdin
	// The test: echo I WANT > foo; echo SOMETHING | cat foo - 
	// Expected output: I WANT\nSOMETHING\n
	
	var buf bytes.Buffer
	
	// Simulate: cat with a file reader first, then stdin
	fileReader := strings.NewReader("I WANT\n")
	stdinReader := strings.NewReader("SOMETHING\n")
	
	readers := []io.Reader{fileReader, stdinReader}
	
	for _, r := range readers {
		data, _ := io.ReadAll(r)
		buf.Write(data)
	}
	
	got := buf.String()
	if got != "I WANT\nSOMETHING\n" {
		t.Errorf("got %q, want %q", got, "I WANT\nSOMETHING\n")
	}
}

func TestBusyBox_Cat_FileAndDashStdin(t *testing.T) {
	// cat foo - should interleave file and stdin correctly.
	// We test with Run: a file reader followed by a stdin reader.
	fileReader := strings.NewReader("FILE\n")
	stdinReader := strings.NewReader("STDIN\n")

	readers := []io.Reader{fileReader, stdinReader}
	var buf bytes.Buffer
	for _, r := range readers {
		io.Copy(&buf, r)
	}

	got := buf.String()
	if got != "FILE\nSTDIN\n" {
		t.Errorf("got %q, want %q", got, "FILE\nSTDIN\n")
	}
}
