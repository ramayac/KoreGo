package du

import (
	"bytes"
	"strings"
	"testing"
)

func TestDuRun(t *testing.T) {
	var out bytes.Buffer
	rc := run([]string{"."}, &out)
	if rc != 0 {
		t.Errorf("expected 0, got %d", rc)
	}
	if !strings.Contains(out.String(), ".") {
		t.Error("expected output with .")
	}
}

func TestDuJSON(t *testing.T) {
	var out bytes.Buffer
	rc := run([]string{"-j", "."}, &out)
	if rc != 0 {
		t.Errorf("expected 0, got %d", rc)
	}
	if !strings.Contains(out.String(), "command") {
		t.Errorf("expected JSON, got %s", out.String())
	}
}
