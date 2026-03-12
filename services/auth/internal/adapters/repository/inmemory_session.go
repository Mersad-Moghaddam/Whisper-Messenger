package repository

import (
	"context"
	"sync"
	"time"

	"whisper/libs/domain/valueobject"
)

type SessionRepository struct {
	mu       sync.Mutex
	sessions map[valueobject.SessionID]sessionRecord
}

type sessionRecord struct {
	userID valueobject.UserID
	token  string
	exp    time.Time
}

func NewSessionRepository() *SessionRepository {
	return &SessionRepository{sessions: map[valueobject.SessionID]sessionRecord{}}
}

func (r *SessionRepository) Save(_ context.Context, sid valueobject.SessionID, uid valueobject.UserID, token string, exp time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions[sid] = sessionRecord{userID: uid, token: token, exp: exp}
	return nil
}

func (r *SessionRepository) Delete(_ context.Context, sid valueobject.SessionID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.sessions, sid)
	return nil
}

func (r *SessionRepository) Exists(_ context.Context, sid valueobject.SessionID, uid valueobject.UserID) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	rec, ok := r.sessions[sid]
	if !ok {
		return false, nil
	}
	if rec.userID != uid || time.Now().After(rec.exp) {
		return false, nil
	}
	return true, nil
}
