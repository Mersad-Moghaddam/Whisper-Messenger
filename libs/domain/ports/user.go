package ports

import (
	"context"
	"time"

	"whisper/libs/domain/entity"
	"whisper/libs/domain/valueobject"
)

type UpdateProfileCommand struct {
	DisplayName string
	AvatarURL   string
}

type UserUseCase interface {
	GetMe(context.Context, valueobject.UserID) (entity.User, error)
	GetByID(context.Context, valueobject.UserID) (entity.User, error)
	UpdateMe(context.Context, valueobject.UserID, UpdateProfileCommand) (entity.User, error)
	SetPresence(context.Context, valueobject.UserID, PresenceState) error
}

type UserRepository interface {
	GetByID(context.Context, valueobject.UserID) (entity.User, error)
	GetByEmail(context.Context, string) (entity.User, error)
	GetByUsername(context.Context, string) (entity.User, error)
	Create(context.Context, entity.User, string) (entity.User, error)
	Update(context.Context, entity.User) (entity.User, error)
}

type PresenceState string

const (
	PresenceOnline  PresenceState = "online"
	PresenceIdle    PresenceState = "idle"
	PresenceOffline PresenceState = "offline"
)

type PresenceRepository interface {
	SetState(context.Context, valueobject.UserID, PresenceState, time.Duration) error
	GetState(context.Context, valueobject.UserID) (PresenceState, error)
}
