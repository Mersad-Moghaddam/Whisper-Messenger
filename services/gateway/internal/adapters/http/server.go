package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/pprof"
	"strings"
	"time"

	domainports "whisper/libs/domain/ports"
	"whisper/libs/domain/valueobject"
	apperrors "whisper/libs/shared/errors"
	"whisper/libs/shared/logger"
	"whisper/services/gateway/internal/ports"
)

type Server struct {
	httpServer    *http.Server
	gateway       ports.GatewayService
	metrics       ports.Metrics
	postgresProbe ports.DependencyProbe
	redisProbe    ports.DependencyProbe
	log           *logger.Logger
}

func NewServer(addr string, gateway ports.GatewayService, metrics ports.Metrics, postgresProbe, redisProbe ports.DependencyProbe, log *logger.Logger) *Server {
	s := &Server{gateway: gateway, metrics: metrics, postgresProbe: postgresProbe, redisProbe: redisProbe, log: log}
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.health)
	mux.HandleFunc("/ready", s.ready)
	mux.HandleFunc("/metrics", s.promMetrics)
	mux.HandleFunc("/v1/gateway/envelope", s.withAuth(s.handleEnvelope))
	mux.HandleFunc("/v1/gateway/ws", s.wsNotImplemented)

	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	s.httpServer = &http.Server{
		Addr:              addr,
		Handler:           s.withRequestMetrics(mux),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       20 * time.Second,
		WriteTimeout:      20 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	return s
}

func (s *Server) Start() error                       { return s.httpServer.ListenAndServe() }
func (s *Server) Shutdown(ctx context.Context) error { return s.httpServer.Shutdown(ctx) }

func (s *Server) withRequestMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.metrics.IncRequests(r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func (s *Server) withAuth(next func(http.ResponseWriter, *http.Request, valueobject.UserID)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auth := strings.TrimSpace(r.Header.Get("Authorization"))
		if !strings.HasPrefix(auth, "Bearer ") {
			s.metrics.IncRejected()
			writeErr(w, apperrors.New(apperrors.KindUnauthorized, "auth_required", "bearer token required", nil))
			return
		}
		token := strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
		userID, err := s.gateway.ValidateToken(r.Context(), token)
		if err != nil {
			s.metrics.IncRejected()
			writeErr(w, err)
			return
		}
		next(w, r, userID)
	}
}

func (s *Server) handleEnvelope(w http.ResponseWriter, r *http.Request, userID valueobject.UserID) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var envelope domainports.Envelope
	if err := json.NewDecoder(r.Body).Decode(&envelope); err != nil {
		s.metrics.IncErrors()
		writeErr(w, apperrors.New(apperrors.KindValidation, "payload_invalid", "invalid JSON payload", err))
		return
	}
	if err := s.gateway.HandleEnvelope(r.Context(), userID, envelope); err != nil {
		s.metrics.IncRejected()
		writeErr(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
}

func (s *Server) wsNotImplemented(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": "native websocket adapter is planned in next phase"})
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 4*time.Second)
	defer cancel()
	if err := s.postgresProbe.Check(ctx); err != nil {
		writeErr(w, apperrors.New(apperrors.KindUnavailable, "postgres_unavailable", "postgres is unavailable", err))
		return
	}
	if err := s.redisProbe.Check(ctx); err != nil {
		writeErr(w, apperrors.New(apperrors.KindUnavailable, "redis_unavailable", "redis is unavailable", err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
}

func (s *Server) promMetrics(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	_, _ = w.Write([]byte(s.metrics.Render()))
}

func writeErr(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apperrors.HTTPStatus(err))
	_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
}
