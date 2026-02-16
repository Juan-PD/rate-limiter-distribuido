package limiter

import (
	"sync"
	"time"
)

// Simple local token bucket used as fallback or for single-node deployments

type tokenBucket struct {
	capacity   int
	tokens     float64
	ratePerSec float64
	lastRefill time.Time
	mu         sync.Mutex
}

// NewTokenBucketLocal creates a simple token bucket for each key (map-based)
func NewTokenBucketLocal(capacity int, refillPerSec float64) *localBuckets {
	return &localBuckets{
		buckets:    make(map[string]*tokenBucket),
		capacity:   capacity,
		refillRate: refillPerSec,
	}
}

type localBuckets struct {
	buckets    map[string]*tokenBucket
	mu         sync.Mutex
	capacity   int
	refillRate float64
}

func (l *localBuckets) get(key string) *tokenBucket {
	l.mu.Lock()
	defer l.mu.Unlock()
	b, ok := l.buckets[key]
	if !ok {
		b = &tokenBucket{
			capacity:   l.capacity,
			tokens:     float64(l.capacity),
			ratePerSec: l.refillRate,
			lastRefill: time.Now(),
		}
		l.buckets[key] = b
	}
	return b
}

// Allow checks and consumes 1 token if available
func (l *localBuckets) Allow(key string) bool {
	b := l.get(key)
	b.mu.Lock()
	defer b.mu.Unlock()
	// refill
	now := time.Now()
	delta := now.Sub(b.lastRefill).Seconds()
	if delta > 0 {
		b.tokens += delta * b.ratePerSec
		if b.tokens > float64(b.capacity) {
			b.tokens = float64(b.capacity)
		}
		b.lastRefill = now
	}
	if b.tokens >= 1 {
		b.tokens -= 1
		return true
	}
	return false
}
