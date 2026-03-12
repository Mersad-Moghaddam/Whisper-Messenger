package presence

import (
	"context"
	"sync"
	"time"

	domainports "whisper/libs/domain/ports"
	"whisper/libs/domain/valueobject"
)

type PresenceRepository struct {
	mu    sync.Mutex
	state map[valueobject.UserID]record
}

type record struct {
	state   domainports.PresenceState
	untilAt time.Time
}

func NewPresenceRepository() *PresenceRepository {
	return &PresenceRepository{state: map[valueobject.UserID]record{}}
}

func (r *PresenceRepository) SetState(_ context.Context, uid valueobject.UserID, st domainports.PresenceState, ttl time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.state[uid] = record{state: st, untilAt: time.Now().UTC().Add(ttl)}
	return nil
}

func (r *PresenceRepository) GetState(_ context.Context, uid valueobject.UserID) (domainports.PresenceState, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	rec, ok := r.state[uid]
	if !ok || time.Now().UTC().After(rec.untilAt) {
		return domainports.PresenceOffline, nil
	}
	return rec.state, nil
}
