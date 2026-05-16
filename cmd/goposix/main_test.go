package main

import (
	"os"
	"strings"
	"testing"

	"github.com/ramayac/goposix"
)

// captureRun runs goposix.Run with argv and captures stdout.
func captureRun(argv []string) (int, string) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	code := goposix.Run(argv)
	w.Close()
	os.Stdout = oldStdout
	var buf []byte
	for {
		b := make([]byte, 1024)
		n, _ := r.Read(b)
		if n == 0 {
			break
		}
		buf = append(buf, b[:n]...)
	}
	return code, string(buf)
}

func TestRun_Help(t *testing.T) {
	code, out := captureRun([]string{"goposix", "--help"})
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if !strings.Contains(out, "goposix") {
		t.Errorf("expected 'goposix' in help, got: %s", out)
	}
}

func TestRun_HelpShortFlag(t *testing.T) {
	code, out := captureRun([]string{"goposix", "-h"})
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if !strings.Contains(out, "goposix") {
		t.Errorf("expected help, got: %s", out)
	}
}

func TestRun_Version(t *testing.T) {
	code, out := captureRun([]string{"goposix", "--version"})
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if !strings.Contains(out, "version") {
		t.Errorf("expected version string, got: %s", out)
	}
}

func TestRun_ListCommands(t *testing.T) {
	code, out := captureRun([]string{"goposix", "--list-commands"})
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if !strings.Contains(out, "ls") && !strings.Contains(out, "echo") {
		t.Errorf("expected 'ls' or 'echo' in command list, got: %s", out)
	}
}

func TestRun_NoArgs(t *testing.T) {
	code, out := captureRun([]string{"goposix"})
	if code != 0 {
		t.Errorf("expected exit 0 for no args, got %d", code)
	}
	if !strings.Contains(out, "goposix") {
		t.Errorf("expected help, got: %s", out)
	}
}

func TestRun_Subcommand(t *testing.T) {
	code, _ := captureRun([]string{"goposix", "echo", "hello"})
	if code != 0 {
		t.Errorf("expected exit 0 for subcommand, got %d", code)
	}
}

func TestRun_SymlinkMode(t *testing.T) {
	// Symlink mode: binary name IS the command (not a well-known name)
	code, _ := captureRun([]string{"/some/path/echo", "hello"})
	if code != 0 {
		t.Errorf("expected exit 0 for symlink mode, got %d", code)
	}
}

func TestRun_UnknownCommand(t *testing.T) {
	code, _ := captureRun([]string{"goposix", "nonexistent_command_zzz"})
	if code != 127 {
		t.Errorf("expected exit 127 for unknown command, got %d", code)
	}
}

func TestRun_UnknownSymlink(t *testing.T) {
	code, _ := captureRun([]string{"/bin/nonexistent_cmd_zzz"})
	if code != 127 {
		t.Errorf("expected exit 127 for unknown symlink, got %d", code)
	}
}

func TestRun_UnknownSymlinkNoArgs(t *testing.T) {
	code, _ := captureRun([]string{"/bin/nonexistent_cmd_zzz"})
	if code != 127 {
		t.Errorf("expected exit 127, got %d", code)
	}
}

func TestRun_CommandWithArgs(t *testing.T) {
	code, _ := captureRun([]string{"goposix", "true"})
	if code != 0 {
		t.Errorf("expected exit 0 for goposix true, got %d", code)
	}
}

func TestRun_CommandReturnsError(t *testing.T) {
	code, _ := captureRun([]string{"goposix", "rm", "/nonexistent/file/zzz"})
	// rm should return non-zero for missing file
	if code == 0 {
		t.Error("expected non-zero exit for missing file")
	}
}
