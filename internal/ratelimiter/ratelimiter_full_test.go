package ratelimiter

import (
	"testing"
	"time"
)

// Utilidad para crear un rate limiter pequeño y fácil de testear
func newTestLimiter() *RateLimiter {
	return NewRateLimiter(5, time.Second) // 5 req por 1 segundo
}

func TestRateLimiter_AllowsWithinLimit(t *testing.T) {
	rl := newTestLimiter()
	ip := "1.1.1.1"

	for i := 0; i < 5; i++ {
		allowed, remaining := rl.Allow(ip)
		if !allowed {
			t.Fatalf("expected request %d to be allowed", i+1)
		}
		if remaining != 4-i {
			t.Fatalf("expected remaining=%d, got=%d", 4-i, remaining)
		}
	}
}

func TestRateLimiter_BlocksAfterLimit(t *testing.T) {
	rl := newTestLimiter()
	ip := "2.2.2.2"

	for i := 0; i < 5; i++ {
		rl.Allow(ip)
	}

	allowed, remaining := rl.Allow(ip)
	if allowed {
		t.Fatal("expected limit to block request but it allowed")
	}
	if remaining != 0 {
		t.Fatalf("expected remaining=0 after blocking, got %d", remaining)
	}
}

func TestRateLimiter_ResetsAfterWindow(t *testing.T) {
	rl := newTestLimiter()
	ip := "3.3.3.3"

	for i := 0; i < 5; i++ {
		rl.Allow(ip)
	}

	// Debe bloquear aquí
	allowed, _ := rl.Allow(ip)
	if allowed {
		t.Fatal("expected request to be blocked before reset window")
	}

	// Esperamos el reset
	time.Sleep(time.Second + 50*time.Millisecond)

	allowed, remaining := rl.Allow(ip)
	if !allowed {
		t.Fatal("expected request to be allowed after window reset")
	}
	if remaining != 4 {
		t.Fatalf("expected remaining=4 after reset, got=%d", remaining)
	}
}

func TestRateLimiter_SeparatesIPs(t *testing.T) {
	rl := newTestLimiter()

	ip1 := "10.0.0.1"
	ip2 := "10.0.0.2"

	for i := 0; i < 5; i++ {
		a1, _ := rl.Allow(ip1)
		a2, _ := rl.Allow(ip2)

		if !a1 || !a2 {
			t.Fatal("each IP should have independent limits")
		}
	}

	// Ambos deben bloquear ahora
	a1, _ := rl.Allow(ip1)
	a2, _ := rl.Allow(ip2)

	if a1 || a2 {
		t.Fatal("expected both IPs to be blocked independently")
	}
}

func TestRateLimiter_FastRequests(t *testing.T) {
	rl := newTestLimiter()
	ip := "4.4.4.4"

	allowedCount := 0

	for i := 0; i < 20; i++ {
		allowed, _ := rl.Allow(ip)
		if allowed {
			allowedCount++
		}
	}

	if allowedCount != 5 {
		t.Fatalf("expected exactly 5 allowed requests, got %d", allowedCount)
	}
}

func TestRateLimiter_MultipleWindows(t *testing.T) {
	rl := newTestLimiter()
	ip := "5.5.5.5"

	// Primera ventana
	for i := 0; i < 5; i++ {
		rl.Allow(ip)
	}

	// Debe bloquear
	if ok, _ := rl.Allow(ip); ok {
		t.Fatal("should block at end of first window")
	}

	// Esperamos reset
	time.Sleep(time.Second + 20*time.Millisecond)

	// Segunda ventana
	for i := 0; i < 5; i++ {
		if ok, _ := rl.Allow(ip); !ok {
			t.Fatal("should allow requests after new window")
		}
	}
}
