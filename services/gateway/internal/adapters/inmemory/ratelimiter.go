package inmemory

import (
	"context"
	"sync"
	"time"
)

type bucket struct {
	count      int
	windowFrom time.Time
}

type RateLimiter struct {
	limit int
	mu    sync.Mutex
	data  map[string]bucket
}

func NewRateLimiter(limitPerMinute int) *RateLimiter {
	if limitPerMinute <= 0 {
		limitPerMinute = 60
	}
	return &RateLimiter{limit: limitPerMinute, data: map[string]bucket{}}
}

func (r *RateLimiter) Allow(_ context.Context, key string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now().UTC()
	b := r.data[key]
	if b.windowFrom.IsZero() || now.Sub(b.windowFrom) >= time.Minute {
		b.windowFrom = now
		b.count = 0
	}
	if b.count >= r.limit {
		r.data[key] = b
		return false, nil
	}
	b.count++
	r.data[key] = b
	return true, nil
}
