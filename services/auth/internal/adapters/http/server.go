package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/pprof"
	"time"

	domainports "whisper/libs/domain/ports"
	"whisper/libs/domain/valueobject"
	apperrors "whisper/libs/shared/errors"
	"whisper/services/auth/internal/ports"
)

type Server struct {
	httpServer *http.Server
	auth       domainports.AuthUseCase
	metrics    ports.Metrics
	probe      ports.DependencyProbe
}

func NewServer(addr string, auth domainports.AuthUseCase, metrics ports.Metrics, probe ports.DependencyProbe) *Server {
	s := &Server{auth: auth, metrics: metrics, probe: probe}
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.health)
	mux.HandleFunc("/ready", s.ready)
	mux.HandleFunc("/metrics", s.promMetrics)
	mux.HandleFunc("/v1/auth/register", s.register)
	mux.HandleFunc("/v1/auth/login", s.login)
	mux.HandleFunc("/v1/auth/refresh", s.refresh)
	mux.HandleFunc("/v1/auth/logout", s.logout)
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)

	s.httpServer = &http.Server{
		Addr:              addr,
		Handler:           s.withMetrics(mux),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       90 * time.Second,
	}
	return s
}

func (s *Server) Start() error                       { return s.httpServer.ListenAndServe() }
func (s *Server) Shutdown(ctx context.Context) error { return s.httpServer.Shutdown(ctx) }

func (s *Server) withMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.metrics.IncRequests(r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		s.metrics.IncErrors()
		writeErr(w, apperrors.New(apperrors.KindValidation, "payload_invalid", "invalid payload", err))
		return
	}
	user, tokens, err := s.auth.Register(r.Context(), domainports.RegisterUserCommand{Username: req.Username, Email: req.Email, Password: req.Password})
	if err != nil {
		s.metrics.IncRejected()
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"user": user, "tokens": tokens})
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		s.metrics.IncErrors()
		writeErr(w, apperrors.New(apperrors.KindValidation, "payload_invalid", "invalid payload", err))
		return
	}
	user, tokens, err := s.auth.Login(r.Context(), domainports.LoginCommand{Email: req.Email, Password: req.Password})
	if err != nil {
		s.metrics.IncRejected()
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": user, "tokens": tokens})
}

func (s *Server) refresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		RefreshToken string `json:"refreshToken"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		s.metrics.IncErrors()
		writeErr(w, apperrors.New(apperrors.KindValidation, "payload_invalid", "invalid payload", err))
		return
	}
	tokens, err := s.auth.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		s.metrics.IncRejected()
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"tokens": tokens})
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		RefreshToken string `json:"refreshToken"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		s.metrics.IncErrors()
		writeErr(w, apperrors.New(apperrors.KindValidation, "payload_invalid", "invalid payload", err))
		return
	}
	parts := split3(req.RefreshToken)
	if len(parts) != 3 {
		s.metrics.IncRejected()
		writeErr(w, apperrors.New(apperrors.KindValidation, "refresh_required", "refreshToken required", nil))
		return
	}
	payload, err := decodePayload(parts[1])
	if err != nil {
		s.metrics.IncRejected()
		writeErr(w, apperrors.New(apperrors.KindValidation, "refresh_invalid", "invalid refresh token", err))
		return
	}
	sid, err := valueobject.ParseSessionID(payload.SID)
	if err != nil {
		s.metrics.IncRejected()
		writeErr(w, apperrors.New(apperrors.KindValidation, "session_invalid", "invalid session", err))
		return
	}
	if err := s.auth.Logout(r.Context(), sid); err != nil {
		s.metrics.IncErrors()
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "logged_out"})
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 4*time.Second)
	defer cancel()
	if err := s.probe.Check(ctx); err != nil {
		writeErr(w, apperrors.New(apperrors.KindUnavailable, "postgres_unavailable", "postgres unavailable", err))
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func (s *Server) promMetrics(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	_, _ = w.Write([]byte(s.metrics.Render()))
}

func writeErr(w http.ResponseWriter, err error) {
	writeJSON(w, apperrors.HTTPStatus(err), map[string]string{"error": err.Error()})
}
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
