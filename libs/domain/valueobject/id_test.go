package valueobject

import "testing"

func TestNewAndParseUserID(t *testing.T) {
	id := NewUserID()
	if len(id) != 36 {
		t.Fatalf("unexpected id length: %d", len(id))
	}
	parsed, err := ParseUserID(string(id))
	if err != nil {
		t.Fatalf("expected valid parse, got err: %v", err)
	}
	if parsed != id {
		t.Fatalf("parsed id mismatch")
	}
}

func TestParseConversationID_InvalidPrefix(t *testing.T) {
	if _, err := ParseConversationID("usr_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"); err == nil {
		t.Fatalf("expected invalid prefix error")
	}
}
