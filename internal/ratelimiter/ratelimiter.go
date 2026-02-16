package ratelimiter

import (
	"sync"
	"time"
)

// RateLimiter implements a fixed-window rate limiter per key (e.g., IP address)
type RateLimiter struct {
	limit   int
	window  time.Duration
	clients map[string]*clientRecord
	mu      sync.Mutex
}

type clientRecord struct {
	count       int
	windowStart time.Time
}

// NewRateLimiter creates a new RateLimiter with the specified limit and window duration
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		limit:   limit,
		window:  window,
		clients: make(map[string]*clientRecord),
	}
}

// Allow checks if a request from the given key is allowed.
// Returns (allowed bool, remaining int)
func (rl *RateLimiter) Allow(key string) (bool, int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	record, exists := rl.clients[key]
	if !exists {
		// First request from this key
		rl.clients[key] = &clientRecord{
			count:       1,
			windowStart: now,
		}
		return true, rl.limit - 1
	}

	// Check if window has expired
	if now.Sub(record.windowStart) >= rl.window {
		// Reset window
		record.count = 1
		record.windowStart = now
		return true, rl.limit - 1
	}

	// Within window - check limit
	if record.count >= rl.limit {
		return false, 0
	}

	record.count++
	return true, rl.limit - record.count
}
