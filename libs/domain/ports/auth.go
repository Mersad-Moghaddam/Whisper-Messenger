package ports

import (
	"context"
	"time"

	"whisper/libs/domain/entity"
	"whisper/libs/domain/valueobject"
)

type RegisterUserCommand struct {
	Username string
	Email    string
	Password string
}

type LoginCommand struct {
	Email    string
	Password string
}

type AuthTokens struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

// Inbound Auth use-cases.
type AuthUseCase interface {
	Register(context.Context, RegisterUserCommand) (entity.User, AuthTokens, error)
	Login(context.Context, LoginCommand) (entity.User, AuthTokens, error)
	Refresh(context.Context, string) (AuthTokens, error)
	Logout(context.Context, valueobject.SessionID) error
}

// Outbound dependencies for auth application service.
type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, hash string) error
}

type TokenIssuer interface {
	IssueAccessToken(subject valueobject.UserID) (string, error)
	IssueRefreshToken(subject valueobject.UserID, sessionID valueobject.SessionID) (string, error)
	ParseRefreshToken(token string) (valueobject.UserID, valueobject.SessionID, error)
}

type AuthSessionRepository interface {
	Save(context.Context, valueobject.SessionID, valueobject.UserID, string, time.Time) error
	Delete(context.Context, valueobject.SessionID) error
	Exists(context.Context, valueobject.SessionID, valueobject.UserID) (bool, error)
}
