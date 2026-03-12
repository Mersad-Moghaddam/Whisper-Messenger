package repository

import (
	"context"
	"sync"
	"time"

	"whisper/libs/domain/entity"
	"whisper/libs/domain/valueobject"
)

type UserRepository struct {
	mu    sync.Mutex
	users map[valueobject.UserID]entity.User
}

func NewUserRepository(seed []entity.User) *UserRepository {
	users := make(map[valueobject.UserID]entity.User, len(seed))
	for _, u := range seed {
		users[u.ID] = u
	}
	return &UserRepository{users: users}
}

func (r *UserRepository) GetByID(_ context.Context, id valueobject.UserID) (entity.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.users[id]
	if !ok {
		return entity.User{}, ErrNotFound
	}
	return u, nil
}

func (r *UserRepository) GetByEmail(_ context.Context, email string) (entity.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, u := range r.users {
		if u.Email == email {
			return u, nil
		}
	}
	return entity.User{}, ErrNotFound
}

func (r *UserRepository) GetByUsername(_ context.Context, username string) (entity.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, u := range r.users {
		if u.Username == username {
			return u, nil
		}
	}
	return entity.User{}, ErrNotFound
}

func (r *UserRepository) Create(_ context.Context, user entity.User, _ string) (entity.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.users[user.ID]; exists {
		return entity.User{}, ErrConflict
	}
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now().UTC()
	}
	r.users[user.ID] = user
	return user, nil
}

func (r *UserRepository) Update(_ context.Context, user entity.User) (entity.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.users[user.ID]; !ok {
		return entity.User{}, ErrNotFound
	}
	r.users[user.ID] = user
	return user, nil
}
