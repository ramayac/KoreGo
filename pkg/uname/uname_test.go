package uname

import "testing"

func TestRunReturnsFields(t *testing.T) {
	result, err := Run()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Sysname == "" {
		t.Error("expected non-empty Sysname")
	}
	if result.Machine == "" {
		t.Error("expected non-empty Machine")
	}
}
