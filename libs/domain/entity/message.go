package entity

import (
	"time"

	"whisper/libs/domain/valueobject"
)

type Message struct {
	ID             valueobject.MessageID
	ConversationID valueobject.ConversationID
	SenderID       valueobject.UserID
	Content        string
	ContentType    valueobject.ContentType
	CreatedAt      time.Time
	EditedAt       *time.Time
}

type ReadReceipt struct {
	MessageID valueobject.MessageID
	UserID    valueobject.UserID
	ReadAt    time.Time
}

type Attachment struct {
	ID          valueobject.AttachmentID
	MessageID   valueobject.MessageID
	StoragePath string
	MIME        string
	Size        int64
	CreatedAt   time.Time
}
