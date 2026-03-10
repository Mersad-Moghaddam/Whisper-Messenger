package entity

import (
	"strings"
	"time"

	"whisper/libs/domain/valueobject"
)

type User struct {
	ID          valueobject.UserID
	Username    string
	Email       string
	DisplayName string
	AvatarURL   string
	CreatedAt   time.Time
	LastSeen    time.Time
}

func (u *User) Normalize() {
	u.Username = strings.TrimSpace(strings.ToLower(u.Username))
	u.Email = strings.TrimSpace(strings.ToLower(u.Email))
	u.DisplayName = strings.TrimSpace(u.DisplayName)
}
