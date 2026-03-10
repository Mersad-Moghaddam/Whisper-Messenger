package domain

import "time"

type Connection struct {
	ID        string
	UserID    string
	CreatedAt time.Time
}
