package telemetry

import "net/http"

// HTTPMiddleware is a no-op wrapper that preserves an HTTP middleware contract.
func HTTPMiddleware(_ string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler { return next }
}
