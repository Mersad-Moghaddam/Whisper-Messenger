package infrastructure

import (
	"context"
	"fmt"
	"net"
	"time"
)

type MultiProbe struct {
	addrs   map[string]string
	retries int
}

func NewMultiProbe(pg, redis string, retries int) *MultiProbe {
	if retries <= 0 {
		retries = 3
	}
	return &MultiProbe{addrs: map[string]string{"postgres": pg, "redis": redis}, retries: retries}
}

func (p *MultiProbe) Check(ctx context.Context) error {
	for name, addr := range p.addrs {
		if err := p.checkOne(ctx, name, addr); err != nil {
			return err
		}
	}
	return nil
}

func (p *MultiProbe) checkOne(ctx context.Context, name, addr string) error {
	var lastErr error
	for attempt := 1; attempt <= p.retries; attempt++ {
		conn, err := (&net.Dialer{Timeout: 2 * time.Second}).DialContext(ctx, "tcp", addr)
		if err == nil {
			_ = conn.Close()
			return nil
		}
		lastErr = err
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(attempt) * 300 * time.Millisecond):
		}
	}
	return fmt.Errorf("%s unavailable: %w", name, lastErr)
}
