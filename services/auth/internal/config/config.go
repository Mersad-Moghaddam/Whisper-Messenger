package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServiceName       string
	Env               string
	HTTPAddr          string
	ShutdownTimeout   time.Duration
	JWTSecret         string
	AccessTTL         time.Duration
	RefreshTTL        time.Duration
	DependencyRetries int
	PostgresAddr      string
}

func Load() (Config, error) {
	cfg := Config{
		ServiceName:       env("SERVICE_NAME", "auth"),
		Env:               env("ENV", "local"),
		HTTPAddr:          env("HTTP_ADDR", ":8081"),
		JWTSecret:         env("JWT_SECRET", "dev-secret"),
		AccessTTL:         time.Duration(envInt("ACCESS_TOKEN_TTL_MIN", 15)) * time.Minute,
		RefreshTTL:        time.Duration(envInt("REFRESH_TOKEN_TTL_HOUR", 24*7)) * time.Hour,
		DependencyRetries: envInt("DEPENDENCY_RETRIES", 5),
		PostgresAddr:      env("POSTGRES_ADDR", "localhost:5432"),
		ShutdownTimeout:   time.Duration(envInt("SHUTDOWN_TIMEOUT_SEC", 10)) * time.Second,
	}
	if cfg.JWTSecret == "" {
		return Config{}, fmt.Errorf("JWT_SECRET is required")
	}
	return cfg, nil
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
