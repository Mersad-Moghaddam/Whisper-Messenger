package application

import (
	"context"
	"strings"
	"time"

	"whisper/libs/domain/entity"
	domainports "whisper/libs/domain/ports"
	"whisper/libs/domain/valueobject"
	apperrors "whisper/libs/shared/errors"
	"whisper/services/auth/internal/adapters/repository"
)

type credentialStore interface {
	PasswordHashByUserID(id valueobject.UserID) (string, error)
}

type Service struct {
	users    domainports.UserRepository
	sessions domainports.AuthSessionRepository
	hasher   domainports.PasswordHasher
	tokens   domainports.TokenIssuer
}

func NewService(users domainports.UserRepository, sessions domainports.AuthSessionRepository, hasher domainports.PasswordHasher, tokens domainports.TokenIssuer) *Service {
	return &Service{users: users, sessions: sessions, hasher: hasher, tokens: tokens}
}

func (s *Service) Register(ctx context.Context, cmd domainports.RegisterUserCommand) (entity.User, domainports.AuthTokens, error) {
	if len(strings.TrimSpace(cmd.Password)) < 8 {
		return entity.User{}, domainports.AuthTokens{}, apperrors.New(apperrors.KindValidation, "password_too_short", "password must be at least 8 characters", nil)
	}
	user := entity.User{
		ID:          valueobject.NewUserID(),
		Username:    cmd.Username,
		Email:       cmd.Email,
		DisplayName: cmd.Username,
		CreatedAt:   time.Now().UTC(),
		LastSeen:    time.Now().UTC(),
	}
	user.Normalize()
	hash, err := s.hasher.Hash(cmd.Password)
	if err != nil {
		return entity.User{}, domainports.AuthTokens{}, apperrors.New(apperrors.KindInternal, "hash_failed", "could not hash password", err)
	}
	created, err := s.users.Create(ctx, user, hash)
	if err != nil {
		if err == repository.ErrConflict {
			return entity.User{}, domainports.AuthTokens{}, apperrors.New(apperrors.KindConflict, "user_exists", "user already exists", err)
		}
		return entity.User{}, domainports.AuthTokens{}, apperrors.New(apperrors.KindInternal, "create_user_failed", "could not create user", err)
	}
	tokens, err := s.issueTokens(ctx, created.ID)
	if err != nil {
		return entity.User{}, domainports.AuthTokens{}, err
	}
	return created, tokens, nil
}

func (s *Service) Login(ctx context.Context, cmd domainports.LoginCommand) (entity.User, domainports.AuthTokens, error) {
	user, err := s.users.GetByEmail(ctx, strings.ToLower(strings.TrimSpace(cmd.Email)))
	if err != nil {
		return entity.User{}, domainports.AuthTokens{}, apperrors.New(apperrors.KindUnauthorized, "invalid_credentials", "invalid email or password", nil)
	}
	cs, ok := s.users.(credentialStore)
	if !ok {
		return entity.User{}, domainports.AuthTokens{}, apperrors.New(apperrors.KindInternal, "credentials_unavailable", "credential store not available", nil)
	}
	hash, err := cs.PasswordHashByUserID(user.ID)
	if err != nil {
		return entity.User{}, domainports.AuthTokens{}, apperrors.New(apperrors.KindUnauthorized, "invalid_credentials", "invalid email or password", nil)
	}
	if err := s.hasher.Verify(cmd.Password, hash); err != nil {
		return entity.User{}, domainports.AuthTokens{}, apperrors.New(apperrors.KindUnauthorized, "invalid_credentials", "invalid email or password", nil)
	}
	tokens, err := s.issueTokens(ctx, user.ID)
	if err != nil {
		return entity.User{}, domainports.AuthTokens{}, err
	}
	return user, tokens, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (domainports.AuthTokens, error) {
	uid, sid, err := s.tokens.ParseRefreshToken(refreshToken)
	if err != nil {
		return domainports.AuthTokens{}, apperrors.New(apperrors.KindUnauthorized, "refresh_invalid", "invalid refresh token", err)
	}
	exists, err := s.sessions.Exists(ctx, sid, uid)
	if err != nil {
		return domainports.AuthTokens{}, apperrors.New(apperrors.KindInternal, "session_check_failed", "session check failed", err)
	}
	if !exists {
		return domainports.AuthTokens{}, apperrors.New(apperrors.KindUnauthorized, "session_not_found", "session not found", nil)
	}
	return s.issueTokens(ctx, uid)
}

func (s *Service) Logout(ctx context.Context, sid valueobject.SessionID) error {
	if err := s.sessions.Delete(ctx, sid); err != nil {
		return apperrors.New(apperrors.KindInternal, "logout_failed", "logout failed", err)
	}
	return nil
}

func (s *Service) issueTokens(ctx context.Context, uid valueobject.UserID) (domainports.AuthTokens, error) {
	sid := valueobject.NewSessionID()
	access, err := s.tokens.IssueAccessToken(uid)
	if err != nil {
		return domainports.AuthTokens{}, apperrors.New(apperrors.KindInternal, "issue_access_failed", "failed to issue access token", err)
	}
	refresh, err := s.tokens.IssueRefreshToken(uid, sid)
	if err != nil {
		return domainports.AuthTokens{}, apperrors.New(apperrors.KindInternal, "issue_refresh_failed", "failed to issue refresh token", err)
	}
	expiresAt := time.Now().UTC().Add(7 * 24 * time.Hour)
	if err := s.sessions.Save(ctx, sid, uid, refresh, expiresAt); err != nil {
		return domainports.AuthTokens{}, apperrors.New(apperrors.KindInternal, "save_session_failed", "failed to persist session", err)
	}
	return domainports.AuthTokens{AccessToken: access, RefreshToken: refresh, ExpiresAt: expiresAt}, nil
}
