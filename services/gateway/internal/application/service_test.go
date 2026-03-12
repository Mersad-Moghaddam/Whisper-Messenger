package application

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	domainports "whisper/libs/domain/ports"
	"whisper/libs/domain/valueobject"
)

type limiterStub struct{ allow bool }

func (l limiterStub) Allow(context.Context, string) (bool, error) { return l.allow, nil }

type hubStub struct{ user, conv int }

func (h *hubStub) BroadcastToUser(valueobject.UserID, []byte) error { h.user++; return nil }
func (h *hubStub) BroadcastToConversation(valueobject.ConversationID, []byte) error {
	h.conv++
	return nil
}

func TestValidateToken(t *testing.T) {
	hub := &hubStub{}
	svc := NewService("secret", limiterStub{allow: true}, hub)
	uid := string(valueobject.NewUserID())
	tok := testToken(t, "secret", uid, time.Now().Add(time.Hour).Unix())
	got, err := svc.ValidateToken(context.Background(), tok)
	if err != nil {
		t.Fatalf("expected token valid: %v", err)
	}
	if string(got) != uid {
		t.Fatalf("unexpected uid: got %s want %s", got, uid)
	}
}

func TestHandleEnvelope_Message(t *testing.T) {
	hub := &hubStub{}
	svc := NewService("secret", limiterStub{allow: true}, hub)
	err := svc.HandleEnvelope(context.Background(), valueobject.NewUserID(), domainports.Envelope{
		Type:    "message.send",
		Payload: map[string]any{"conversationId": string(valueobject.NewConversationID())},
		Nonce:   "abc",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hub.conv != 1 {
		t.Fatalf("expected conversation broadcast")
	}
}

type testingTB interface{ Helper() }

func testToken(t testingTB, secret, sub string, exp int64) string {
	if t != nil {
		t.Helper()
	}
	header, _ := json.Marshal(map[string]any{"alg": "HS256", "typ": "JWT"})
	claims, _ := json.Marshal(map[string]any{"sub": sub, "exp": exp})
	p0 := base64.RawURLEncoding.EncodeToString(header)
	p1 := base64.RawURLEncoding.EncodeToString(claims)
	sig := sign(secret, p0+"."+p1)
	return p0 + "." + p1 + "." + sig
}

func BenchmarkValidateToken(b *testing.B) {
	hub := &hubStub{}
	svc := NewService("secret", limiterStub{allow: true}, hub)
	uid := string(valueobject.NewUserID())
	tok := testToken(b, "secret", uid, time.Now().Add(time.Hour).Unix())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = svc.ValidateToken(context.Background(), tok)
	}
}

func BenchmarkHandleEnvelope(b *testing.B) {
	hub := &hubStub{}
	svc := NewService("secret", limiterStub{allow: true}, hub)
	env := domainports.Envelope{Type: "message.send", Payload: map[string]any{"conversationId": string(valueobject.NewConversationID())}, Nonce: "n"}
	uid := valueobject.NewUserID()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = svc.HandleEnvelope(context.Background(), uid, env)
	}
}
