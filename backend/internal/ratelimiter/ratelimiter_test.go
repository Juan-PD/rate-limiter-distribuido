package ratelimiter

import (
	"testing"
	"time"
)

func TestRateLimiter(t *testing.T) {
	rl := NewRateLimiter(5, time.Minute) // 5 requests por minuto

	ip := "127.0.0.1"

	// Las primeras 5 peticiones deben permitirse
	for i := 0; i < 5; i++ {
		allowed, _ := rl.Allow(ip)
		if !allowed {
			t.Fatalf("Expected request %d to be allowed", i+1)
		}
	}

	// La sexta debe ser bloqueada
	allowed, _ := rl.Allow(ip)
	if allowed {
		t.Fatal("Expected sixth request to be blocked but it was allowed")
	}
}
