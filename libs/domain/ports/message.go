package ports

import (
	"context"

	"whisper/libs/domain/entity"
	"whisper/libs/domain/valueobject"
)

type CreateConversationCommand struct {
	Type           entity.ConversationType
	Title          string
	ParticipantIDs []valueobject.UserID
}

type SendMessageCommand struct {
	ConversationID valueobject.ConversationID
	SenderID       valueobject.UserID
	Content        string
	ContentType    valueobject.ContentType
	Nonce          string
}

type MessageUseCase interface {
	CreateConversation(context.Context, CreateConversationCommand) (entity.Conversation, error)
	SendMessage(context.Context, SendMessageCommand) (entity.Message, error)
	GetHistory(context.Context, valueobject.ConversationID, valueobject.Cursor) ([]entity.Message, string, error)
	MarkRead(context.Context, valueobject.MessageID, valueobject.UserID) error
}

type ConversationRepository interface {
	Create(context.Context, entity.Conversation, []valueobject.UserID) (entity.Conversation, error)
	IsParticipant(context.Context, valueobject.ConversationID, valueobject.UserID) (bool, error)
}

type MessageRepository interface {
	Create(context.Context, entity.Message) (entity.Message, error)
	ListByConversation(context.Context, valueobject.ConversationID, valueobject.Cursor) ([]entity.Message, string, error)
	CreateReadReceipt(context.Context, entity.ReadReceipt) error
}

type EventPublisher interface {
	PublishMessageCreated(context.Context, entity.Message, string) error
	PublishTyping(context.Context, valueobject.ConversationID, valueobject.UserID) error
}
