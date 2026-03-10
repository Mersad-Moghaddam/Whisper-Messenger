package http

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

type InMemoryMetrics struct {
	mu       sync.Mutex
	requests map[string]int
	rejected int
	errors   int
}

func NewInMemoryMetrics() *InMemoryMetrics {
	return &InMemoryMetrics{requests: map[string]int{}}
}

func (m *InMemoryMetrics) IncRequests(path string) { m.mu.Lock(); m.requests[path]++; m.mu.Unlock() }
func (m *InMemoryMetrics) IncRejected()            { m.mu.Lock(); m.rejected++; m.mu.Unlock() }
func (m *InMemoryMetrics) IncErrors()              { m.mu.Lock(); m.errors++; m.mu.Unlock() }

func (m *InMemoryMetrics) Render() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	var b strings.Builder
	b.WriteString("# TYPE whisper_auth_requests_total counter\n")
	keys := make([]string, 0, len(m.requests))
	for k := range m.requests {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		b.WriteString(fmt.Sprintf("whisper_auth_requests_total{path=\"%s\"} %d\n", k, m.requests[k]))
	}
	b.WriteString(fmt.Sprintf("whisper_auth_rejected_total %d\n", m.rejected))
	b.WriteString(fmt.Sprintf("whisper_auth_errors_total %d\n", m.errors))
	return b.String()
}
