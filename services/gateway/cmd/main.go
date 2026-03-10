package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"whisper/libs/shared/logger"
	gatewayhttp "whisper/services/gateway/internal/adapters/http"
	"whisper/services/gateway/internal/adapters/infrastructure"
	"whisper/services/gateway/internal/adapters/inmemory"
	"whisper/services/gateway/internal/application"
	"whisper/services/gateway/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	log, err := logger.New(logger.Config{Service: cfg.ServiceName, Env: cfg.Env, Level: "info"})
	if err != nil {
		panic(err)
	}

	limiter := inmemory.NewRateLimiter(cfg.RateLimitPerMin)
	hub := inmemory.NewHub()
	metrics := inmemory.NewMetrics()
	svc := application.NewService(cfg.JWTSecret, limiter, hub)

	postgresProbe := infrastructure.NewTCPProbe("postgres", cfg.PostgresAddr, cfg.DependencyRetries)
	redisProbe := infrastructure.NewTCPProbe("redis", cfg.RedisAddr, cfg.DependencyRetries)

	server := gatewayhttp.NewServer(cfg.HTTPAddr, svc, metrics, postgresProbe, redisProbe, log)

	errCh := make(chan error, 1)
	go func() {
		log.Info("gateway starting", map[string]any{"addr": cfg.HTTPAddr})
		errCh <- server.Start()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	select {
	case sig := <-sigCh:
		log.Info("shutdown signal received", map[string]any{"signal": sig.String()})
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Error("server stopped unexpectedly", map[string]any{"error": err.Error()})
			os.Exit(1)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Error("graceful shutdown failed", map[string]any{"error": err.Error()})
		os.Exit(1)
	}
	time.Sleep(100 * time.Millisecond)
	log.Info("gateway stopped", nil)
}
