package inmemory

import (
	"context"
	"testing"
)

func TestRateLimiter(t *testing.T) {
	rl := NewRateLimiter(2)
	ctx := context.Background()
	if ok, _ := rl.Allow(ctx, "u1"); !ok {
		t.Fatalf("expected first allow")
	}
	if ok, _ := rl.Allow(ctx, "u1"); !ok {
		t.Fatalf("expected second allow")
	}
	if ok, _ := rl.Allow(ctx, "u1"); ok {
		t.Fatalf("expected third deny")
	}
}
