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
	PresenceTTL       time.Duration
	DependencyRetries int
	PostgresAddr      string
	RedisAddr         string
}

func Load() (Config, error) {
	cfg := Config{
		ServiceName:       env("SERVICE_NAME", "user"),
		Env:               env("ENV", "local"),
		HTTPAddr:          env("HTTP_ADDR", ":8082"),
		PresenceTTL:       time.Duration(envInt("PRESENCE_TTL_SEC", 90)) * time.Second,
		DependencyRetries: envInt("DEPENDENCY_RETRIES", 5),
		PostgresAddr:      env("POSTGRES_ADDR", "localhost:5432"),
		RedisAddr:         env("REDIS_ADDR", "localhost:6379"),
		ShutdownTimeout:   time.Duration(envInt("SHUTDOWN_TIMEOUT_SEC", 10)) * time.Second,
	}
	if cfg.PresenceTTL < 10*time.Second {
		return Config{}, fmt.Errorf("PRESENCE_TTL_SEC must be >= 10")
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
