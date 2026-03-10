package inmemory

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

type Metrics struct {
	mu       sync.Mutex
	requests map[string]int
	rejected int
	errors   int
}

func NewMetrics() *Metrics {
	return &Metrics{requests: map[string]int{}}
}

func (m *Metrics) IncRequests(path string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requests[path]++
}
func (m *Metrics) IncRejected() { m.mu.Lock(); m.rejected++; m.mu.Unlock() }
func (m *Metrics) IncErrors()   { m.mu.Lock(); m.errors++; m.mu.Unlock() }

func (m *Metrics) Render() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	var b strings.Builder
	b.WriteString("# TYPE whisper_gateway_requests_total counter\n")
	keys := make([]string, 0, len(m.requests))
	for k := range m.requests {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		b.WriteString(fmt.Sprintf("whisper_gateway_requests_total{path=\"%s\"} %d\n", k, m.requests[k]))
	}
	b.WriteString("# TYPE whisper_gateway_rejected_total counter\n")
	b.WriteString(fmt.Sprintf("whisper_gateway_rejected_total %d\n", m.rejected))
	b.WriteString("# TYPE whisper_gateway_errors_total counter\n")
	b.WriteString(fmt.Sprintf("whisper_gateway_errors_total %d\n", m.errors))
	return b.String()
}
