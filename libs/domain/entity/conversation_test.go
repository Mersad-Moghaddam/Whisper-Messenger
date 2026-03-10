package entity

import (
	"testing"

	"whisper/libs/domain/valueobject"
)

func TestConversationValidate(t *testing.T) {
	valid := Conversation{ID: valueobject.NewConversationID(), Type: ConversationTypeDM}
	if err := valid.Validate(); err != nil {
		t.Fatalf("expected valid conversation: %v", err)
	}

	invalid := Conversation{ID: valueobject.NewConversationID(), Type: "weird"}
	if err := invalid.Validate(); err == nil {
		t.Fatalf("expected invalid conversation error")
	}
}
