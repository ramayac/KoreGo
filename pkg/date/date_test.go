package date

import (
	"bytes"
	"strings"
	"testing"
)

func TestDateRun(t *testing.T) {
	var out bytes.Buffer
	rc := run([]string{"-u"}, &out)
	if rc != 0 {
		t.Errorf("expected 0, got %d", rc)
	}
	if out.String() == "" {
		t.Error("expected output")
	}
}

func TestDateJSON(t *testing.T) {
	var out bytes.Buffer
	rc := run([]string{"-j"}, &out)
	if rc != 0 {
		t.Errorf("expected 0, got %d", rc)
	}
	if !strings.Contains(out.String(), "jsonrpc") && !strings.Contains(out.String(), "command") {
		t.Errorf("expected JSON, got %s", out.String())
	}
}
