package limiter

import (
	"context"
	"fmt"
	"time"

	"rate-limiter-distribuido/internal/config"

	"github.com/redis/go-redis/v9"
)

// RedisFixedWindowLimiter — implementación simple de ventana fija usando INCR + EXPIRE
// Nota: Es simple y fácil de entender; para producción considera comandos LUA para atomizar por precisión.

type RedisFixedWindowLimiter struct {
	rdb    *redis.Client
	limit  int
	window time.Duration
	prefix string
}

func NewRedisFixedWindowLimiter(cfg config.Config) (*RedisFixedWindowLimiter, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}

	return &RedisFixedWindowLimiter{
		rdb:    rdb,
		limit:  cfg.RateLimitRequests,
		window: time.Duration(cfg.RateLimitWindowSecs) * time.Second,
		prefix: "rl:",
	}, nil
}

func (r *RedisFixedWindowLimiter) keyFor(k string) string {
	return r.prefix + k
}

func (r *RedisFixedWindowLimiter) Allow(ctx context.Context, key string) (bool, error) {
	k := r.keyFor(key)
	// INCR key
	n := r.rdb.Incr(ctx, k)
	if err := n.Err(); err != nil {
		return false, err
	}
	val := n.Val()
	if val == 1 {
		// set expiry first time
		err := r.rdb.Expire(ctx, k, r.window).Err()
		if err != nil {
			return false, err
		}
	}
	if int(val) > r.limit {
		return false, nil
	}
	return true, nil
}
