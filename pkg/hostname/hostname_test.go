package hostname

import "testing"

func TestRunReturnsHostname(t *testing.T) {
	result, err := Run()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Name == "" {
		t.Error("expected non-empty hostname")
	}
}
