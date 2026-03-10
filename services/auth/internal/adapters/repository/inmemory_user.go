package repository

import (
	"context"
	"sync"

	"whisper/libs/domain/entity"
	"whisper/libs/domain/valueobject"
)

type UserRepository struct {
	mu         sync.Mutex
	usersByID  map[valueobject.UserID]entity.User
	hashByID   map[valueobject.UserID]string
	idByEmail  map[string]valueobject.UserID
	idByHandle map[string]valueobject.UserID
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		usersByID:  map[valueobject.UserID]entity.User{},
		hashByID:   map[valueobject.UserID]string{},
		idByEmail:  map[string]valueobject.UserID{},
		idByHandle: map[string]valueobject.UserID{},
	}
}

func (r *UserRepository) GetByID(_ context.Context, id valueobject.UserID) (entity.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.usersByID[id]
	if !ok {
		return entity.User{}, ErrNotFound
	}
	return u, nil
}

func (r *UserRepository) GetByEmail(_ context.Context, email string) (entity.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	id, ok := r.idByEmail[email]
	if !ok {
		return entity.User{}, ErrNotFound
	}
	return r.usersByID[id], nil
}

func (r *UserRepository) GetByUsername(_ context.Context, username string) (entity.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	id, ok := r.idByHandle[username]
	if !ok {
		return entity.User{}, ErrNotFound
	}
	return r.usersByID[id], nil
}

func (r *UserRepository) Create(_ context.Context, user entity.User, hash string) (entity.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.idByEmail[user.Email]; ok {
		return entity.User{}, ErrConflict
	}
	if _, ok := r.idByHandle[user.Username]; ok {
		return entity.User{}, ErrConflict
	}
	r.usersByID[user.ID] = user
	r.hashByID[user.ID] = hash
	r.idByEmail[user.Email] = user.ID
	r.idByHandle[user.Username] = user.ID
	return user, nil
}

func (r *UserRepository) Update(_ context.Context, user entity.User) (entity.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.usersByID[user.ID]; !ok {
		return entity.User{}, ErrNotFound
	}
	r.usersByID[user.ID] = user
	return user, nil
}

func (r *UserRepository) PasswordHashByUserID(id valueobject.UserID) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	h, ok := r.hashByID[id]
	if !ok {
		return "", ErrNotFound
	}
	return h, nil
}
