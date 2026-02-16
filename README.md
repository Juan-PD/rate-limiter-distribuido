[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/Juan-PD/rate-limiter-distribuido)
# Rate Limiter Distribuido

Servidor HTTP en Go con rate limiting configurable. Soporta despliegues de un solo nodo (token bucket en memoria) y distribuidos (ventana fija con Redis). Identifica clientes por IP y aplica límites por ventana de tiempo. [1](#0-0) 

## Arquitectura

- **Entrada**: `cmd/server/main.go` inicia el servidor, carga configuración y registra handlers `/ping` y `/hello`. [2](#0-1) 
- **Middleware**: `internal/http/middleware.go` aplica rate limiting a cada request. [3](#0-2) 
- **Limiters**:
  - Local: `internal/limiter/token_bucket_local.go` (token bucket en memoria). [4](#0-3) 
  - Distribuido: `internal/limiter/fixed_window_redis.go` (ventana fija con Redis). [5](#0-4) 
- **Configuración**: `internal/config/config.go` lee variables de entorno. [6](#0-5) 

## Configuración (variables de entorno)

| Variable | Defecto | Descripción |
|----------|---------|-------------|
| `PORT` | `8080` | Puerto del servidor HTTP |
| `RATE_LIMIT_REQUESTS` | `10` | Máximo de peticiones por ventana |
| `RATE_LIMIT_WINDOW_SECONDS` | `1` | Duración de la ventana (segundos) |
| `REDIS_ADDR` | `localhost:6379` | Dirección de Redis (solo modo distribuido) |
| `REDIS_PASSWORD` | `""` | Contraseña de Redis |
| `REDIS_DB` | `0` | Base de datos Redis |

## Pasos para usarlo

### 1. Clonar y dependencias
```bash
git clone <repo-url>
cd rate-limiter-distribuido
go mod tidy
```

### 2. Ejecutar en modo local (sin Redis)
```bash
go run cmd/server/main.go
```
Por defecto usa `localBuckets` (token bucket en memoria). [7](#0-6) 

### 3. Ejecutar en modo distribuido (con Redis)
a) Asegúrate de que Redis esté accesible en `REDIS_ADDR`. [8](#0-7) 
b) Modifica `internal/http/middleware.go` para usar `RedisFixedWindowLimiter`:
```go
// En NewRateLimitMiddleware, reemplaza:
rl := limiter.NewTokenBucketLocal(...)
// Por:
rl, err := limiter.NewRedisFixedWindowLimiter(cfg)
if err != nil { log.Fatal(err) }
```
c) Ejecuta:
```bash
go run cmd/server/main.go
```

### 4. Probar endpoints
```bash
curl http://localhost:8080/ping   # Health check
curl http://localhost:8080/hello  # Endpoint de ejemplo
```
Si superas el límite, recibirás `429 Too Many Requests`. [9](#0-8) 

## Notas
- Para cambiar de algoritmo, edita `NewRateLimitMiddleware` en `internal/http/middleware.go`. [7](#0-6) 
- En producción con Redis, considera usar scripts Lua para atomicidad (ver comentario en `fixed_window_redis.go`). [10](#0-9) 

Wiki pages you might want to explore:
- [Overview (Juan-PD/rate-limiter-distribuido)](/wiki/Juan-PD/rate-limiter-distribuido#1)
- [Rate Limiting (Juan-PD/rate-limiter-distribuido)](/wiki/Juan-PD/rate-limiter-distribuido#4)
- [Distributed Redis Fixed Window (Juan-PD/rate-limiter-distribuido)](/wiki/Juan-PD/rate-limiter-distribuido#4.3)

### Citations

**File:** cmd/server/main.go (L13-36)
```go
func main() {
	cfg := config.Load()
	logger := utils.NewLogger()
	logger.Infof("starting server on :%s", cfg.Port)

	mux := http.NewServeMux()
	mux.HandleFunc("/ping", httpHandlers.PingHandler)
	mux.HandleFunc("/hello", httpHandlers.HelloHandler)

	// Wrap mux with our middleware
	handler := httpHandlers.NewRateLimitMiddleware(cfg, logger)(mux)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: handler,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}

	// ensure context cancel on graceful shutdown in future iterations
	_ = context.Background()
}
```

**File:** internal/http/middleware.go (L13-36)
```go
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
```

**File:** internal/limiter/token_bucket_local.go (L18-25)
```go
// NewTokenBucketLocal creates a simple token bucket for each key (map-based)
func NewTokenBucketLocal(capacity int, refillPerSec float64) *localBuckets {
	return &localBuckets{
		buckets:    make(map[string]*tokenBucket),
		capacity:   capacity,
		refillRate: refillPerSec,
	}
}
```

**File:** internal/limiter/fixed_window_redis.go (L13-14)
```go
// RedisFixedWindowLimiter — implementación simple de ventana fija usando INCR + EXPIRE
// Nota: Es simple y fácil de entender; para producción considera comandos LUA para atomizar por precisión.
```

**File:** internal/limiter/fixed_window_redis.go (L23-43)
```go
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
```

**File:** internal/config/config.go (L17-31)
```go
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
```
