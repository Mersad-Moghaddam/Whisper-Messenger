package ports

import (
	"context"

	"whisper/libs/domain/valueobject"
)

type GatewayUseCase interface {
	ValidateToken(context.Context, string) (valueobject.UserID, error)
	HandleEnvelope(context.Context, valueobject.UserID, Envelope) error
}

type Envelope struct {
	Type    string         `json:"type"`
	Payload map[string]any `json:"payload"`
	Nonce   string         `json:"nonce"`
}

type RateLimiter interface {
	Allow(context.Context, string) (bool, error)
}

type WebSocketHub interface {
	JoinUser(userID valueobject.UserID, connID string)
	LeaveUser(userID valueobject.UserID, connID string)
	BroadcastToUser(userID valueobject.UserID, data []byte) error
	BroadcastToConversation(conversationID valueobject.ConversationID, data []byte) error
}
