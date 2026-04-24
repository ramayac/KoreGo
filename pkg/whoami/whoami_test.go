package whoami

import "testing"

func TestRunReturnsUser(t *testing.T) {
	result, err := Run()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.User == "" {
		t.Error("expected non-empty username")
	}
	if result.UID < 0 {
		t.Errorf("expected UID >= 0, got %d", result.UID)
	}
}
