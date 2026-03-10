package inmemory

import (
	"fmt"
	"sync"

	"whisper/libs/domain/valueobject"
)

type Hub struct {
	mu                    sync.Mutex
	userEvents            map[valueobject.UserID][][]byte
	conversationEventFeed map[valueobject.ConversationID][][]byte
}

func NewHub() *Hub {
	return &Hub{
		userEvents:            map[valueobject.UserID][][]byte{},
		conversationEventFeed: map[valueobject.ConversationID][][]byte{},
	}
}

func (h *Hub) JoinUser(_ valueobject.UserID, _ string)  {}
func (h *Hub) LeaveUser(_ valueobject.UserID, _ string) {}

func (h *Hub) BroadcastToUser(userID valueobject.UserID, data []byte) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.userEvents[userID] = append(h.userEvents[userID], data)
	return nil
}

func (h *Hub) BroadcastToConversation(conversationID valueobject.ConversationID, data []byte) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.conversationEventFeed[conversationID] = append(h.conversationEventFeed[conversationID], data)
	return nil
}

func (h *Hub) DebugStats() string {
	h.mu.Lock()
	defer h.mu.Unlock()
	return fmt.Sprintf("users=%d conversations=%d", len(h.userEvents), len(h.conversationEventFeed))
}
