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
	RateLimitPerMin   int
	PostgresAddr      string
	RedisAddr         string
	DependencyRetries int
}

func Load() (Config, error) {
	cfg := Config{
		ServiceName:       env("SERVICE_NAME", "gateway"),
		Env:               env("ENV", "local"),
		HTTPAddr:          env("HTTP_ADDR", ":8080"),
		JWTSecret:         env("JWT_SECRET", "dev-secret"),
		RateLimitPerMin:   envInt("RATE_LIMIT_PER_MIN", 240),
		PostgresAddr:      env("POSTGRES_ADDR", "localhost:5432"),
		RedisAddr:         env("REDIS_ADDR", "localhost:6379"),
		DependencyRetries: envInt("DEPENDENCY_RETRIES", 5),
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
	parsed, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return parsed
}
