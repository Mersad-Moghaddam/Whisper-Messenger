package security

import (
	"testing"
	"time"

	"whisper/libs/domain/valueobject"
)

func TestRefreshTokenRoundTrip(t *testing.T) {
	issuer := NewTokenIssuer("secret", time.Minute, time.Hour)
	uid := valueobject.NewUserID()
	sid := valueobject.NewSessionID()
	tok, err := issuer.IssueRefreshToken(uid, sid)
	if err != nil {
		t.Fatalf("issue refresh: %v", err)
	}
	uid2, sid2, err := issuer.ParseRefreshToken(tok)
	if err != nil {
		t.Fatalf("parse refresh: %v", err)
	}
	if uid2 != uid || sid2 != sid {
		t.Fatalf("claims mismatch")
	}
}
