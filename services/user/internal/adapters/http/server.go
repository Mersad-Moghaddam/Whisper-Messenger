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
	"whisper/services/user/internal/ports"
)

type Server struct {
	httpServer *http.Server
	users      domainports.UserUseCase
	presence   domainports.PresenceRepository
	metrics    ports.Metrics
	probe      ports.DependencyProbe
}

func NewServer(addr string, users domainports.UserUseCase, presence domainports.PresenceRepository, metrics ports.Metrics, probe ports.DependencyProbe) *Server {
	s := &Server{users: users, presence: presence, metrics: metrics, probe: probe}
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.health)
	mux.HandleFunc("/ready", s.ready)
	mux.HandleFunc("/metrics", s.promMetrics)
	mux.HandleFunc("/v1/users/me", s.me)
	mux.HandleFunc("/v1/users/", s.byID)
	mux.HandleFunc("/v1/users/me/presence", s.setPresence)
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	s.httpServer = &http.Server{Addr: addr, Handler: s.withMetrics(mux), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 90 * time.Second}
	return s
}
func (s *Server) Start() error                       { return s.httpServer.ListenAndServe() }
func (s *Server) Shutdown(ctx context.Context) error { return s.httpServer.Shutdown(ctx) }
func (s *Server) withMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { s.metrics.IncRequests(r.URL.Path); next.ServeHTTP(w, r) })
}

func (s *Server) me(w http.ResponseWriter, r *http.Request) {
	uid, ok := userIDFromHeader(w, r)
	if !ok {
		s.metrics.IncRejected()
		return
	}
	switch r.Method {
	case http.MethodGet:
		u, err := s.users.GetMe(r.Context(), uid)
		if err != nil {
			s.metrics.IncRejected()
			writeErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"user": u})
	case http.MethodPut:
		var req struct {
			DisplayName string `json:"displayName"`
			AvatarURL   string `json:"avatarUrl"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
			s.metrics.IncErrors()
			writeErr(w, apperrors.New(apperrors.KindValidation, "payload_invalid", "invalid payload", err))
			return
		}
		u, err := s.users.UpdateMe(r.Context(), uid, domainports.UpdateProfileCommand{DisplayName: req.DisplayName, AvatarURL: req.AvatarURL})
		if err != nil {
			s.metrics.IncRejected()
			writeErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"user": u})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) byID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	raw := strings.TrimPrefix(r.URL.Path, "/v1/users/")
	uid, err := valueobject.ParseUserID(raw)
	if err != nil {
		s.metrics.IncRejected()
		writeErr(w, apperrors.New(apperrors.KindValidation, "user_id_invalid", "invalid user id", err))
		return
	}
	u, err := s.users.GetByID(r.Context(), uid)
	if err != nil {
		s.metrics.IncRejected()
		writeErr(w, err)
		return
	}
	st, _ := s.presence.GetState(r.Context(), uid)
	writeJSON(w, http.StatusOK, map[string]any{"user": u, "presence": st})
}

func (s *Server) setPresence(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	uid, ok := userIDFromHeader(w, r)
	if !ok {
		s.metrics.IncRejected()
		return
	}
	var req struct {
		State domainports.PresenceState `json:"state"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		s.metrics.IncErrors()
		writeErr(w, apperrors.New(apperrors.KindValidation, "payload_invalid", "invalid payload", err))
		return
	}
	if err := s.users.SetPresence(r.Context(), uid, req.State); err != nil {
		s.metrics.IncRejected()
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
func (s *Server) ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 4*time.Second)
	defer cancel()
	if err := s.probe.Check(ctx); err != nil {
		writeErr(w, apperrors.New(apperrors.KindUnavailable, "deps_unavailable", "dependencies unavailable", err))
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}
func (s *Server) promMetrics(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	_, _ = w.Write([]byte(s.metrics.Render()))
}

func userIDFromHeader(w http.ResponseWriter, r *http.Request) (valueobject.UserID, bool) {
	raw := strings.TrimSpace(r.Header.Get("X-User-ID"))
	if raw == "" {
		writeErr(w, apperrors.New(apperrors.KindUnauthorized, "auth_required", "X-User-ID required", nil))
		return "", false
	}
	uid, err := valueobject.ParseUserID(raw)
	if err != nil {
		writeErr(w, apperrors.New(apperrors.KindUnauthorized, "auth_invalid", "invalid user id", err))
		return "", false
	}
	return uid, true
}
func writeErr(w http.ResponseWriter, err error) {
	writeJSON(w, apperrors.HTTPStatus(err), map[string]string{"error": err.Error()})
}
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
