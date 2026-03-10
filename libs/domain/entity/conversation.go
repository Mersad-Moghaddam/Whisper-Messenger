package entity

import (
	"fmt"
	"time"

	"whisper/libs/domain/valueobject"
)

type ConversationType string

const (
	ConversationTypeDM    ConversationType = "dm"
	ConversationTypeGroup ConversationType = "group"
)

type Conversation struct {
	ID        valueobject.ConversationID
	Type      ConversationType
	Title     string
	CreatedAt time.Time
}

func (c Conversation) Validate() error {
	switch c.Type {
	case ConversationTypeDM, ConversationTypeGroup:
		return nil
	default:
		return fmt.Errorf("invalid conversation type: %s", c.Type)
	}
}
