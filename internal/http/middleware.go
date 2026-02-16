package http

import (
	"context"
	"net/http"

	"rate-limiter-distribuido/internal/config"
	"rate-limiter-distribuido/internal/limiter"
	"rate-limiter-distribuido/pkg/utils"
)

// NewRateLimitMiddleware creates a rate limit middleware using config and logger
func NewRateLimitMiddleware(cfg config.Config, logger *utils.Logger) func(http.Handler) http.Handler {
	// Create a local token bucket limiter using config values
	rl := limiter.NewTokenBucketLocal(cfg.RateLimitRequests, float64(cfg.RateLimitRequests)/float64(cfg.RateLimitWindowSecs))

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			allowed := rl.Allow(ip)

			if !allowed {
				logger.Infof("rate limit exceeded for IP: %s", ip)
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte("429 - Too Many Requests"))
				return
			}

			// Add rate limit info to context
			ctx := context.WithValue(r.Context(), "rate_limited", false)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}
