package daemon

import (
	"testing"
	"time"
)

func TestRateLimiter(t *testing.T) {
	rl := NewRateLimiter(10.0, 5) // 10 req/s, burst of 5

	// Burst should be allowed
	for i := 0; i < 5; i++ {
		if !rl.Allow() {
			t.Errorf("expected request %d to be allowed", i)
		}
	}

	// 6th request should be denied
	if rl.Allow() {
		t.Errorf("expected 6th request to be denied")
	}

	// Wait 100ms (1 token)
	time.Sleep(105 * time.Millisecond)

	// Should be allowed now
	if !rl.Allow() {
		t.Errorf("expected request to be allowed after sleeping")
	}

	// But next one denied
	if rl.Allow() {
		t.Errorf("expected request to be denied after using regenerated token")
	}
}
