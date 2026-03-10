package logger

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

type Level string

const (
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
)

type Config struct {
	Service string
	Env     string
	Level   string
	Writer  io.Writer
}

type Logger struct {
	service string
	env     string
	level   Level
	std     *log.Logger
	mu      sync.Mutex
}

func New(cfg Config) (*Logger, error) {
	w := cfg.Writer
	if w == nil {
		w = os.Stdout
	}

	lvl := parseLevel(cfg.Level)
	return &Logger{
		service: cfg.Service,
		env:     cfg.Env,
		level:   lvl,
		std:     log.New(w, "", 0),
	}, nil
}

func parseLevel(v string) Level {
	switch strings.ToLower(v) {
	case "debug":
		return LevelDebug
	case "warn":
		return LevelWarn
	case "error":
		return LevelError
	default:
		return LevelInfo
	}
}

func (l *Logger) log(level Level, msg string, fields map[string]any) {
	if !enabled(l.level, level) {
		return
	}
	entry := map[string]any{
		"ts":      time.Now().UTC().Format(time.RFC3339Nano),
		"service": l.service,
		"env":     l.env,
		"level":   level,
		"msg":     msg,
	}
	for k, v := range fields {
		entry[k] = v
	}
	buf, _ := json.Marshal(entry)
	l.mu.Lock()
	defer l.mu.Unlock()
	l.std.Println(string(buf))
}

func enabled(threshold, current Level) bool {
	order := map[Level]int{LevelDebug: 0, LevelInfo: 1, LevelWarn: 2, LevelError: 3}
	return order[current] >= order[threshold]
}

func (l *Logger) Debug(msg string, fields map[string]any) { l.log(LevelDebug, msg, fields) }
func (l *Logger) Info(msg string, fields map[string]any)  { l.log(LevelInfo, msg, fields) }
func (l *Logger) Warn(msg string, fields map[string]any)  { l.log(LevelWarn, msg, fields) }
func (l *Logger) Error(msg string, fields map[string]any) { l.log(LevelError, msg, fields) }

type ctxKey struct{}

func WithContext(ctx context.Context, l *Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, l)
}

func FromContext(ctx context.Context, fallback *Logger) *Logger {
	if lgr, ok := ctx.Value(ctxKey{}).(*Logger); ok && lgr != nil {
		return lgr
	}
	return fallback
}
