package sleep

import (
	"bytes"
	"testing"
)

func TestSleepMissingArgs(t *testing.T) {
	var out bytes.Buffer
	rc := run([]string{}, &out)
	if rc != 1 {
		t.Errorf("expected 1, got %d", rc)
	}
}

func TestSleepGoDuration(t *testing.T) {
	var out bytes.Buffer
	rc := run([]string{"1ms"}, &out)
	if rc != 0 {
		t.Errorf("expected 0, got %d", rc)
	}
}

func TestSleepFloat(t *testing.T) {
	var out bytes.Buffer
	rc := run([]string{"0.001"}, &out)
	if rc != 0 {
		t.Errorf("expected 0, got %d", rc)
	}
}
