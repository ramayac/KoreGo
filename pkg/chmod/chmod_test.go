package chmod

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestChmodMissingArgs(t *testing.T) {
	var out bytes.Buffer
	rc := run([]string{}, &out)
	if rc != 1 {
		t.Errorf("expected 1, got %d", rc)
	}
}

func TestChmodJSON(t *testing.T) {
	var out bytes.Buffer
	f, _ := os.CreateTemp("", "chmod")
	defer os.Remove(f.Name())

	rc := run([]string{"-j", "0755", f.Name()}, &out)
	if rc != 0 {
		t.Errorf("expected 0, got %d", rc)
	}
	if !strings.Contains(out.String(), "command") {
		t.Errorf("expected JSON, got %s", out.String())
	}
}
