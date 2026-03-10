package valueobject

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

type UserID string
type ConversationID string
type MessageID string
type AttachmentID string
type SessionID string

func NewUserID() UserID                 { return UserID(newID("usr")) }
func NewConversationID() ConversationID { return ConversationID(newID("cnv")) }
func NewMessageID() MessageID           { return MessageID(newID("msg")) }
func NewAttachmentID() AttachmentID     { return AttachmentID(newID("att")) }
func NewSessionID() SessionID           { return SessionID(newID("ses")) }

func ParseUserID(v string) (UserID, error) {
	if err := validateID(v, "usr"); err != nil {
		return "", err
	}
	return UserID(v), nil
}

func ParseConversationID(v string) (ConversationID, error) {
	if err := validateID(v, "cnv"); err != nil {
		return "", err
	}
	return ConversationID(v), nil
}

func ParseMessageID(v string) (MessageID, error) {
	if err := validateID(v, "msg"); err != nil {
		return "", err
	}
	return MessageID(v), nil
}

func ParseAttachmentID(v string) (AttachmentID, error) {
	if err := validateID(v, "att"); err != nil {
		return "", err
	}
	return AttachmentID(v), nil
}

func ParseSessionID(v string) (SessionID, error) {
	if err := validateID(v, "ses"); err != nil {
		return "", err
	}
	return SessionID(v), nil
}

func newID(prefix string) string {
	buf := make([]byte, 16)
	_, _ = rand.Read(buf)
	return fmt.Sprintf("%s_%s", prefix, hex.EncodeToString(buf))
}

func validateID(v, prefix string) error {
	if len(v) != 36 {
		return fmt.Errorf("invalid %s id length", prefix)
	}
	wantPrefix := prefix + "_"
	if v[:4] != wantPrefix {
		return fmt.Errorf("invalid %s id prefix", prefix)
	}
	return nil
}
