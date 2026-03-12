package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"whisper/libs/shared/logger"
	authhttp "whisper/services/auth/internal/adapters/http"
	"whisper/services/auth/internal/adapters/infrastructure"
	"whisper/services/auth/internal/adapters/repository"
	"whisper/services/auth/internal/adapters/security"
	"whisper/services/auth/internal/application"
	"whisper/services/auth/internal/config"
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

	users := repository.NewUserRepository()
	sessions := repository.NewSessionRepository()
	hasher := security.NewPasswordHasher()
	tokens := security.NewTokenIssuer(cfg.JWTSecret, cfg.AccessTTL, cfg.RefreshTTL)
	svc := application.NewService(users, sessions, hasher, tokens)
	metrics := authhttp.NewInMemoryMetrics()
	probe := infrastructure.NewTCPProbe("postgres", cfg.PostgresAddr, cfg.DependencyRetries)
	server := authhttp.NewServer(cfg.HTTPAddr, svc, metrics, probe)

	errCh := make(chan error, 1)
	go func() { log.Info("auth starting", map[string]any{"addr": cfg.HTTPAddr}); errCh <- server.Start() }()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-sigCh:
		log.Info("shutdown signal received", nil)
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
	log.Info("auth stopped", nil)
}
