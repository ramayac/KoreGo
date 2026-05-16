package find

import (
	"bytes"
	"strings"
	"testing"
)

func TestFindRun(t *testing.T) {
	var out bytes.Buffer
	rc := run([]string{"."}, &out)
	if rc != 0 {
		t.Errorf("expected 0, got %d", rc)
	}
	if !strings.Contains(out.String(), ".") {
		t.Error("expected output with .")
	}
}

func TestFindJSON(t *testing.T) {
	var out bytes.Buffer
	rc := run([]string{"-j", "."}, &out)
	if rc != 0 {
		t.Errorf("expected 0, got %d", rc)
	}
	if !strings.Contains(out.String(), "command") {
		t.Errorf("expected JSON, got %s", out.String())
	}
}

func TestBuildExecArgs(t *testing.T) {
	files := []FileInfo{{Path: "a.txt"}, {Path: "b.txt"}}

	// {} replacement
	args := buildExecArgs([]string{"echo", "{}"}, files)
	if len(args) != 3 {
		t.Errorf("expected 3 args, got %d: %v", len(args), args)
	}
	if args[1] != "a.txt" || args[2] != "b.txt" {
		t.Errorf("expected a.txt b.txt, got %v", args[1:])
	}

	// No {} placeholder
	args2 := buildExecArgs([]string{"ls", "-la"}, files)
	if len(args2) != 2 || args2[0] != "ls" || args2[1] != "-la" {
		t.Errorf("expected passthrough, got %v", args2)
	}

	// Empty
	args3 := buildExecArgs([]string{}, files)
	if len(args3) != 0 {
		t.Errorf("expected 0 args, got %d", len(args3))
	}
}

func TestBuildExecArgsMultiplePlaceholders(t *testing.T) {
	files := []FileInfo{{Path: "x"}, {Path: "y"}}
	args := buildExecArgs([]string{"cmd", "{}", "--", "{}"}, files)
	if len(args) != 6 {
		t.Errorf("expected 6 args, got %d: %v", len(args), args)
	}
}

// BusyBox hardening: find -xdev should be accepted as a valid flag.
func TestFindXdevFlag(t *testing.T) {
	var out bytes.Buffer
	code := run([]string{".", "-xdev"}, &out)
	if code != 0 {
		t.Fatalf("find -xdev exited with %d, want 0", code)
	}
}
