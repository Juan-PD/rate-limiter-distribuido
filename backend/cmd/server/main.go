package main

import (
	"context"
	"log"
	"net/http"

	"rate-limiter-distribuido/internal/config"
	httpHandlers "rate-limiter-distribuido/internal/http"
	"rate-limiter-distribuido/pkg/utils"
)

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
