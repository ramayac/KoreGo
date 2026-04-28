package printf

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestFormatBasicString(t *testing.T) {
	got := Format("hello %s", []string{"world"})
	if got != "hello world" {
		t.Errorf("got %q, want %q", got, "hello world")
	}
}

func TestFormatMultipleSpecifiers(t *testing.T) {
	got := Format("%s is %d", []string{"age", "30"})
	if got != "age is 30" {
		t.Errorf("got %q, want %q", got, "age is 30")
	}
}

func TestFormatEscapeNewline(t *testing.T) {
	got := Format("line1\\nline2", nil)
	if got != "line1\nline2" {
		t.Errorf("got %q, want %q", got, "line1\nline2")
	}
}

func TestFormatEscapeTab(t *testing.T) {
	got := Format("col1\\tcol2", nil)
	if got != "col1\tcol2" {
		t.Errorf("got %q, want %q", got, "col1\tcol2")
	}
}

func TestFormatEscapeBackslash(t *testing.T) {
	got := Format("path\\\\file", nil)
	if got != "path\\file" {
		t.Errorf("got %q, want %q", got, "path\\file")
	}
}

func TestFormatOctalEscape(t *testing.T) {
	got := Format("A\\0101B", nil) // \0101 = 'A' in octal (65)
	if got != "AAB" {
		t.Errorf("got %q, want %q", got, "AAB")
	}
}

func TestFormatOctalNUL(t *testing.T) {
	got := Format("a\\0b", nil)
	if got != "a\x00b" {
		t.Errorf("got %q, want %q", got, "a\\x00b")
	}
}

func TestFormatHex(t *testing.T) {
	got := Format("%x", []string{"255"})
	if got != "ff" {
		t.Errorf("got %q, want %q", got, "ff")
	}
}

func TestFormatHexUpper(t *testing.T) {
	got := Format("%X", []string{"255"})
	if got != "FF" {
		t.Errorf("got %q, want %q", got, "FF")
	}
}

func TestFormatOctal(t *testing.T) {
	got := Format("%o", []string{"8"})
	if got != "10" {
		t.Errorf("got %q, want %q", got, "10")
	}
}

func TestFormatFloat(t *testing.T) {
	got := Format("%.2f", []string{"3.14159"})
	if got != "3.14" {
		t.Errorf("got %q, want %q", got, "3.14")
	}
}

func TestFormatWidth(t *testing.T) {
	got := Format("%10s", []string{"hi"})
	if got != "        hi" {
		t.Errorf("got %q, want %q", got, "        hi")
	}
}

func TestFormatLeftAlign(t *testing.T) {
	got := Format("%-10s|", []string{"hi"})
	if got != "hi        |" {
		t.Errorf("got %q, want %q", got, "hi        |")
	}
}

func TestFormatPercentLiteral(t *testing.T) {
	got := Format("100%%", nil)
	if got != "100%" {
		t.Errorf("got %q, want %q", got, "100%")
	}
}

func TestFormatArgCycling(t *testing.T) {
	// When more specifiers than args, cycle
	got := Format("%s %s %s", []string{"a", "b"})
	if got != "a b a" {
		t.Errorf("got %q, want %q", got, "a b a")
	}
}

func TestFormatNoArgs(t *testing.T) {
	got := Format("plain text", nil)
	if got != "plain text" {
		t.Errorf("got %q, want %q", got, "plain text")
	}
}

func TestFormatCharSpec(t *testing.T) {
	got := Format("%c", []string{"ABC"})
	if got != "A" {
		t.Errorf("got %q, want %q", got, "A")
	}
}

func TestRunCLI(t *testing.T) {
	var buf bytes.Buffer
	code := run([]string{"%s=%d\\n", "x", "42"}, &buf)
	if code != 0 {
		t.Fatalf("exit code %d, want 0", code)
	}
	if buf.String() != "x=42\n" {
		t.Errorf("got %q, want %q", buf.String(), "x=42\n")
	}
}

func TestRunCLIJSON(t *testing.T) {
	var buf bytes.Buffer
	code := run([]string{"--json", "hello %s", "world"}, &buf)
	if code != 0 {
		t.Fatalf("exit code %d, want 0", code)
	}
	// Should be valid JSON envelope
	var env map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	data := env["data"].(map[string]interface{})
	if data["output"] != "hello world" {
		t.Errorf("got %q, want %q", data["output"], "hello world")
	}
}

func TestRunCLIMissingOperand(t *testing.T) {
	var buf bytes.Buffer
	code := run([]string{}, &buf)
	if code != 1 {
		t.Fatalf("exit code %d, want 1", code)
	}
}

func TestFormatIntegerSpecifiers(t *testing.T) {
	tests := []struct {
		format string
		args   []string
		want   string
	}{
		{"%d", []string{"42"}, "42"},
		{"%05d", []string{"42"}, "00042"},
		{"%+d", []string{"42"}, "+42"},
		{"%i", []string{"10"}, "10"},   // %i treated as %d
		{"%d", []string{"0xff"}, "255"}, // hex input
		{"%d", []string{"077"}, "63"},   // octal input
	}
	for _, tc := range tests {
		got := Format(tc.format, tc.args)
		if got != tc.want {
			t.Errorf("Format(%q, %v) = %q, want %q", tc.format, tc.args, got, tc.want)
		}
	}
}

func TestProcessEscapes(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`\a`, "\a"},
		{`\b`, "\b"},
		{`\f`, "\f"},
		{`\v`, "\v"},
		{`\r`, "\r"},
		{`no escapes`, "no escapes"},
	}
	for _, tc := range tests {
		got := processEscapes(tc.input)
		if got != tc.want {
			t.Errorf("processEscapes(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestFormatEmptyString(t *testing.T) {
	got := Format("%s", []string{""})
	if got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

func TestFormatScientific(t *testing.T) {
	got := Format("%e", []string{"1000"})
	if !strings.Contains(got, "e+") {
		t.Errorf("got %q, want scientific notation", got)
	}
}
