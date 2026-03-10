package infrastructure

import (
	"context"
	"fmt"
	"net"
	"time"
)

type TCPProbe struct {
	name    string
	addr    string
	retries int
}

func NewTCPProbe(name, addr string, retries int) *TCPProbe {
	if retries <= 0 {
		retries = 3
	}
	return &TCPProbe{name: name, addr: addr, retries: retries}
}

func (p *TCPProbe) Check(ctx context.Context) error {
	var lastErr error
	for attempt := 1; attempt <= p.retries; attempt++ {
		dialer := net.Dialer{Timeout: 2 * time.Second}
		conn, err := dialer.DialContext(ctx, "tcp", p.addr)
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
	return fmt.Errorf("%s unavailable after retries: %w", p.name, lastErr)
}
