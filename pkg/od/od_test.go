package od

import (
	"bytes"
	"strings"
	"testing"
)

// TestOd_Default verifies default 2-byte octal short dump.
func TestOd_Default(t *testing.T) {
	var out bytes.Buffer
	in := strings.NewReader("HELLO")
	code := odRun(nil, in, &out)
	if code != 0 {
		t.Fatalf("exit code %d, want 0", code)
	}
	// Should contain "042510" (octal for HE in little-endian)
	if !strings.Contains(out.String(), "042510") {
		t.Errorf("expected 042510 in output, got: %q", out.String())
	}
	// Should have final offset line
	if !strings.Contains(out.String(), "0000005") {
		t.Errorf("expected final offset 0000005, got: %q", out.String())
	}
}

// TestOd_OctalBytes verifies od -b (1-byte octal).
func TestOd_OctalBytes(t *testing.T) {
	var out bytes.Buffer
	in := strings.NewReader("HELLO")
	code := odRun([]string{"-b"}, in, &out)
	if code != 0 {
		t.Fatalf("exit code %d, want 0", code)
	}
	// H=110, E=105, L=114, L=114, O=117
	expected := "110 105 114 114 117"
	if !strings.Contains(out.String(), expected) {
		t.Errorf("expected %q in output, got: %q", expected, out.String())
	}
}

// TestOd_Char verifies od -c (character dump).
func TestOd_Char(t *testing.T) {
	var out bytes.Buffer
	in := strings.NewReader("HELLO")
	code := odRun([]string{"-c"}, in, &out)
	if code != 0 {
		t.Fatalf("exit code %d, want 0", code)
	}
	// Should contain "  H   E   L   L   O"
	if !strings.Contains(out.String(), "  H") {
		t.Errorf("expected char dump with 'H', got: %q", out.String())
	}
}

// TestOd_CharEscapes verifies od -c escape sequences.
func TestOd_CharEscapes(t *testing.T) {
	var out bytes.Buffer
	in := strings.NewReader("a\nb")
	code := odRun([]string{"-c"}, in, &out)
	if code != 0 {
		t.Fatalf("exit code %d, want 0", code)
	}
	if !strings.Contains(out.String(), "\\n") {
		t.Errorf("expected \\n escape in output, got: %q", out.String())
	}
}

// TestOd_Hex verifies od -x (hex dump).
func TestOd_Hex(t *testing.T) {
	var out bytes.Buffer
	in := strings.NewReader("HELLO")
	code := odRun([]string{"-x"}, in, &out)
	if code != 0 {
		t.Fatalf("exit code %d, want 0", code)
	}
	// H=0x48, E=0x45 → 4548; L=0x4C, L=0x4C → 4c4c; O=0x4F → 004f
	expected := "4548"
	if !strings.Contains(out.String(), expected) {
		t.Errorf("expected %q in output, got: %q", expected, out.String())
	}
}

// TestOd_Count verifies od -N N (byte limit).
func TestOd_Count(t *testing.T) {
	var out bytes.Buffer
	in := strings.NewReader("HELLO")
	code := odRun([]string{"-N", "3"}, in, &out)
	if code != 0 {
		t.Fatalf("exit code %d, want 0", code)
	}
	// First 3 bytes of "HELLO" → just HE + L. Default format.
	if !strings.Contains(out.String(), "0000003") {
		t.Errorf("expected final offset 0000003, got: %q", out.String())
	}
}

// TestOd_FromStdin verifies od reads from stdin when no file given.
func TestOd_FromStdin(t *testing.T) {
	var out bytes.Buffer
	in := strings.NewReader("test")
	code := odRun(nil, in, &out)
	if code != 0 {
		t.Fatalf("exit code %d, want 0", code)
	}
	if out.Len() == 0 {
		t.Error("expected output, got empty")
	}
}

// TestOd_Json verifies --json structured output.
func TestOd_Json(t *testing.T) {
	var out bytes.Buffer
	in := strings.NewReader("AB")
	code := odRun([]string{"--json"}, in, &out)
	if code != 0 {
		t.Fatalf("exit code %d, want 0", code)
	}
	if !strings.Contains(out.String(), "\"records\"") {
		t.Errorf("expected JSON output with 'records', got: %q", out.String())
	}
}
