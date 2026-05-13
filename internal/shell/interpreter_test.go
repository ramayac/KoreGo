package shell

import (
	"os"
	"strings"
	"testing"
)

func TestExecBasic(t *testing.T) {
	result := Exec("echo hello", "", nil)
	if result.Stdout != "hello\n" {
		t.Errorf("expected 'hello\\n', got %q", result.Stdout)
	}
	if result.ExitCode != 0 {
		t.Errorf("expected exit 0, got %d", result.ExitCode)
	}
}

func TestTimeout(t *testing.T) {
	os.Setenv("KOREGO_SHELL_TIMEOUT", "500ms")
	defer os.Unsetenv("KOREGO_SHELL_TIMEOUT")

	result := Exec("sleep 10", "", nil)
	if result.ExitCode == 0 {
		t.Error("expected non-zero exit code from timeout, got 0")
	}
	if !strings.Contains(result.Stderr, "deadline") && !strings.Contains(result.Stderr, "killed") && !strings.Contains(result.Stderr, "signal") {
		t.Logf("stderr from timeout: %q", result.Stderr)
	}
}

func TestTimeoutViaEnv(t *testing.T) {
	os.Setenv("KOREGO_SHELL_TIMEOUT", "100ms")
	defer os.Unsetenv("KOREGO_SHELL_TIMEOUT")

	result := Exec("sleep 5", "", nil)
	if result.ExitCode == 0 {
		t.Error("expected non-zero exit from 100ms timeout")
	}
}

func TestOutputWithinLimits(t *testing.T) {
	// Verify that output within the 128MB LimitWriter cap works correctly.
	result := Exec("echo hello && echo world", "", nil)
	if result.ExitCode != 0 {
		t.Errorf("unexpected exit %d: %s", result.ExitCode, result.Stderr)
	}
	if !strings.Contains(result.Stdout, "hello") || !strings.Contains(result.Stdout, "world") {
		t.Errorf("unexpected stdout: %q", result.Stdout)
	}
}

func TestPathEscape(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(tmpDir+"/allowed.txt", []byte("ok"), 0644)

	result := Exec("cat allowed.txt", tmpDir, nil)
	if result.ExitCode != 0 {
		t.Fatalf("allowed read failed: %s", result.Stderr)
	}
	if result.Stdout != "ok" {
		t.Errorf("expected 'ok', got %q", result.Stdout)
	}
}

func TestPathEscapeBlocked(t *testing.T) {
	tmpDir := t.TempDir()

	// openHandler intercepts shell-level file opens (redirections like <, >).
	// Use a shell redirection to test path traversal blocking.
	result := Exec("cat < ../../../etc/passwd", tmpDir, nil)
	if result.ExitCode == 0 {
		t.Error("expected non-zero exit for path traversal attempt")
	}
	errOut := strings.ToLower(result.Stderr)
	if !strings.Contains(errOut, "traversal") && !strings.Contains(errOut, "permission") && !strings.Contains(errOut, "no such file") && !strings.Contains(errOut, "not found") {
		t.Logf("stderr from path escape attempt: %q", result.Stderr)
	}
}

func TestEnvVarInjection(t *testing.T) {
	result := Exec("echo $TEST_VAR", "", map[string]string{"TEST_VAR": "injected"})
	if result.Stdout != "injected\n" {
		t.Errorf("expected 'injected\\n', got %q", result.Stdout)
	}
}

func TestStderrCapture(t *testing.T) {
	result := Exec("echo error >&2", "", nil)
	if result.Stderr != "error\n" {
		t.Errorf("expected 'error\\n' on stderr, got %q", result.Stderr)
	}
	if result.ExitCode != 0 {
		t.Errorf("expected exit 0, got %d", result.ExitCode)
	}
}

func TestNonZeroExit(t *testing.T) {
	result := Exec("exit 42", "", nil)
	if result.ExitCode != 42 {
		t.Errorf("expected exit 42, got %d", result.ExitCode)
	}
}

func TestSyntaxError(t *testing.T) {
	result := Exec("{{{{", "", nil)
	if result.ExitCode != 127 {
		t.Errorf("expected exit 127 for syntax error, got %d", result.ExitCode)
	}
}
