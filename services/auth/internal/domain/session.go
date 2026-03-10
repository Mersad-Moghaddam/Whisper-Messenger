package domain

import (
	"time"

	"whisper/libs/domain/valueobject"
)

type Session struct {
	ID           valueobject.SessionID
	UserID       valueobject.UserID
	RefreshToken string
	ExpiresAt    time.Time
}
