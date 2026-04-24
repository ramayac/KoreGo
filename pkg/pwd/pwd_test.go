package pwd

import (
	"os"
	"testing"
)

func TestRunReturnsCwd(t *testing.T) {
	expected, _ := os.Getwd()
	result, err := Run(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Path != expected {
		t.Errorf("got %q, want %q", result.Path, expected)
	}
}

func TestRunPhysical(t *testing.T) {
	result, err := Run(true)
	if err != nil {
		t.Fatalf("unexpected error: %v (physical mode)", err)
	}
	if result.Path == "" {
		t.Error("expected non-empty path in physical mode")
	}
}
