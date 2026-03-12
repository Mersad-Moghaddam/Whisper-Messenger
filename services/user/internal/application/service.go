package application

import (
	"context"
	"strings"
	"time"

	"whisper/libs/domain/entity"
	domainports "whisper/libs/domain/ports"
	"whisper/libs/domain/valueobject"
	apperrors "whisper/libs/shared/errors"
	"whisper/services/user/internal/adapters/repository"
)

type Service struct {
	users    domainports.UserRepository
	presence domainports.PresenceRepository
	ttl      time.Duration
}

func NewService(users domainports.UserRepository, presence domainports.PresenceRepository, ttl time.Duration) *Service {
	return &Service{users: users, presence: presence, ttl: ttl}
}

func (s *Service) GetMe(ctx context.Context, uid valueobject.UserID) (entity.User, error) {
	user, err := s.users.GetByID(ctx, uid)
	if err != nil {
		return entity.User{}, apperrors.New(apperrors.KindNotFound, "user_not_found", "user not found", err)
	}
	return user, nil
}

func (s *Service) GetByID(ctx context.Context, uid valueobject.UserID) (entity.User, error) {
	return s.GetMe(ctx, uid)
}

func (s *Service) UpdateMe(ctx context.Context, uid valueobject.UserID, cmd domainports.UpdateProfileCommand) (entity.User, error) {
	user, err := s.users.GetByID(ctx, uid)
	if err != nil {
		return entity.User{}, apperrors.New(apperrors.KindNotFound, "user_not_found", "user not found", err)
	}
	if dn := strings.TrimSpace(cmd.DisplayName); dn != "" {
		user.DisplayName = dn
	}
	if avatar := strings.TrimSpace(cmd.AvatarURL); avatar != "" {
		user.AvatarURL = avatar
	}
	user.LastSeen = time.Now().UTC()
	updated, err := s.users.Update(ctx, user)
	if err != nil {
		if err == repository.ErrNotFound {
			return entity.User{}, apperrors.New(apperrors.KindNotFound, "user_not_found", "user not found", err)
		}
		return entity.User{}, apperrors.New(apperrors.KindInternal, "user_update_failed", "could not update user", err)
	}
	return updated, nil
}

func (s *Service) SetPresence(ctx context.Context, uid valueobject.UserID, state domainports.PresenceState) error {
	switch state {
	case domainports.PresenceOnline, domainports.PresenceIdle, domainports.PresenceOffline:
	default:
		return apperrors.New(apperrors.KindValidation, "presence_invalid", "invalid presence state", nil)
	}
	if err := s.presence.SetState(ctx, uid, state, s.ttl); err != nil {
		return apperrors.New(apperrors.KindUnavailable, "presence_store_failed", "presence unavailable", err)
	}
	return nil
}
