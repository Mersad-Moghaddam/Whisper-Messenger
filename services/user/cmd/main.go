package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"whisper/libs/domain/entity"
	"whisper/libs/domain/valueobject"
	"whisper/libs/shared/logger"
	userhttp "whisper/services/user/internal/adapters/http"
	"whisper/services/user/internal/adapters/infrastructure"
	"whisper/services/user/internal/adapters/presence"
	"whisper/services/user/internal/adapters/repository"
	"whisper/services/user/internal/application"
	"whisper/services/user/internal/config"
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
	seed := []entity.User{{ID: valueobject.NewUserID(), Username: "demo", Email: "demo@whisper.local", DisplayName: "Demo User", CreatedAt: time.Now().UTC(), LastSeen: time.Now().UTC()}}
	users := repository.NewUserRepository(seed)
	presenceRepo := presence.NewPresenceRepository()
	svc := application.NewService(users, presenceRepo, cfg.PresenceTTL)
	metrics := userhttp.NewInMemoryMetrics()
	probe := infrastructure.NewMultiProbe(cfg.PostgresAddr, cfg.RedisAddr, cfg.DependencyRetries)
	server := userhttp.NewServer(cfg.HTTPAddr, svc, presenceRepo, metrics, probe)
	errCh := make(chan error, 1)
	go func() {
		log.Info("user service starting", map[string]any{"addr": cfg.HTTPAddr})
		errCh <- server.Start()
	}()
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-sigCh:
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Error("server error", map[string]any{"error": err.Error()})
			os.Exit(1)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Error("shutdown error", map[string]any{"error": err.Error()})
		os.Exit(1)
	}
	log.Info("user service stopped", nil)
}
