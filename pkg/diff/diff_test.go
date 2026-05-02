package diff

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDiffEqual(t *testing.T) {
	differ, hunks := GenerateDiff("a\nb\nc", "a\nb\nc", 3, false, false)
	if differ {
		t.Errorf("Expected differ to be false")
	}
	if len(hunks) != 0 {
		t.Errorf("Expected 0 hunks, got %d", len(hunks))
	}
}

func TestDiffSimpleDiff(t *testing.T) {
	differ, hunks := GenerateDiff("a\nb\nc", "a\nx\nc", 3, false, false)
	if !differ {
		t.Errorf("Expected differ to be true")
	}
	if len(hunks) != 1 {
		t.Fatalf("Expected 1 hunk, got %d", len(hunks))
	}
	h := hunks[0]
	if h.OldStart != 1 || h.OldLines != 3 || h.NewStart != 1 || h.NewLines != 3 {
		t.Errorf("Hunk numbers wrong: %+v", h)
	}
	expectedLines := []string{" a", "-b", "+x", " c"}
	for i, l := range expectedLines {
		if h.Lines[i] != l {
			t.Errorf("Hunk line %d: got %q, want %q", i, h.Lines[i], l)
		}
	}
}

func TestDiffContextLines(t *testing.T) {
	a := "1\n2\n3\n4\n5\n6\n7\n8\n9"
	b := "1\n2\n3\n4\nx\n6\n7\n8\n9"
	differ, hunks := GenerateDiff(a, b, 2, false, false)
	if !differ {
		t.Errorf("Expected differ to be true")
	}
	if len(hunks) != 1 {
		t.Fatalf("Expected 1 hunk, got %d", len(hunks))
	}
	h := hunks[0]
	if h.OldStart != 3 || h.OldLines != 5 || h.NewStart != 3 || h.NewLines != 5 {
		t.Errorf("Hunk numbers wrong: %+v", h)
	}
	expectedLines := []string{" 3", " 4", "-5", "+x", " 6", " 7"}
	for i, l := range expectedLines {
		if h.Lines[i] != l {
			t.Errorf("Hunk line %d: got %q, want %q", i, h.Lines[i], l)
		}
	}
}

func TestRunCLI(t *testing.T) {
	tmpDir := t.TempDir()
	f1 := filepath.Join(tmpDir, "a.txt")
	f2 := filepath.Join(tmpDir, "b.txt")
	os.WriteFile(f1, []byte("a\nb\nc\n"), 0644)
	os.WriteFile(f2, []byte("a\nx\nc\n"), 0644)

	var buf bytes.Buffer
	code := run([]string{"-u", f1, f2}, &buf)
	if code != 1 {
		t.Errorf("exit code %d, want 1", code)
	}
	out := buf.String()
	if !bytes.Contains([]byte(out), []byte("--- "+f1)) || !bytes.Contains([]byte(out), []byte("+++ "+f2)) {
		t.Errorf("missing headers in output: %s", out)
	}
	if !bytes.Contains([]byte(out), []byte("@@ -1,3 +1,3 @@")) {
		t.Errorf("missing hunk header in output: %s", out)
	}
}

func TestRunCLIJSON(t *testing.T) {
	tmpDir := t.TempDir()
	f1 := filepath.Join(tmpDir, "a.txt")
	f2 := filepath.Join(tmpDir, "b.txt")
	os.WriteFile(f1, []byte("a\nb\nc\n"), 0644)
	os.WriteFile(f2, []byte("a\nx\nc\n"), 0644)

	var buf bytes.Buffer
	code := run([]string{"-j", f1, f2}, &buf)
	if code != 1 {
		t.Errorf("exit code %d, want 1", code)
	}

	var env map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	data := env["data"].(map[string]interface{})
	if data["differ"] != true {
		t.Errorf("expected differ true")
	}
}
