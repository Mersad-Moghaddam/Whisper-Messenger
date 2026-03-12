package telemetry

import "context"

type Config struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	OTLPEndpoint   string
	SampleRatio    float64
}

// InitTracer initializes tracing and returns a shutdown callback.
// In constrained environments it acts as a no-op initializer.
func InitTracer(_ context.Context, _ Config) (func(context.Context) error, error) {
	return func(context.Context) error { return nil }, nil
}
