package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port                string
	RedisAddr           string
	RedisPassword       string
	RedisDB             int
	RateLimitRequests   int
	RateLimitWindowSecs int
}

func Load() Config {
	r := getEnv("REDIS_ADDR", "localhost:6379")
	rp := getEnv("REDIS_PASSWORD", "")
	rdb := atoi(getEnv("REDIS_DB", "0"))
	rq := atoi(getEnv("RATE_LIMIT_REQUESTS", "10"))
	rw := atoi(getEnv("RATE_LIMIT_WINDOW_SECONDS", "1"))

	return Config{
		Port:                getEnv("PORT", "8080"),
		RedisAddr:           r,
		RedisPassword:       rp,
		RedisDB:             rdb,
		RateLimitRequests:   rq,
		RateLimitWindowSecs: rw,
	}
}

func getEnv(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func atoi(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return i
}
