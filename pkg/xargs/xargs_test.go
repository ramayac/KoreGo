package xargs

import (
	"bytes"
	"strings"
	"testing"
)

func TestXargsJSON(t *testing.T) {
	var out bytes.Buffer
	rc := run([]string{"-j", "true"}, &out)
	if rc != 0 {
		t.Errorf("expected 0, got %d", rc)
	}
	if !strings.Contains(out.String(), "command") {
		t.Errorf("expected JSON, got %s", out.String())
	}
}
