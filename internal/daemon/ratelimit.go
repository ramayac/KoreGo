package daemon

import (
	"sync"
	"time"
)

// RateLimiter implements a simple token bucket.
type RateLimiter struct {
	mu         sync.Mutex
	rate       float64
	burst      int
	tokens     float64
	lastUpdate time.Time
}

// NewRateLimiter creates a new rate limiter with the given rate (tokens/sec) and burst limit.
func NewRateLimiter(rate float64, burst int) *RateLimiter {
	return &RateLimiter{
		rate:       rate,
		burst:      burst,
		tokens:     float64(burst),
		lastUpdate: time.Now(),
	}
}

// Allow returns true if a request is allowed by the rate limiter, false otherwise.
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastUpdate).Seconds()

	rl.tokens += elapsed * rl.rate
	if rl.tokens > float64(rl.burst) {
		rl.tokens = float64(rl.burst)
	}
	rl.lastUpdate = now

	if rl.tokens >= 1.0 {
		rl.tokens -= 1.0
		return true
	}

	return false
}
