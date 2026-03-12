package application

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"crypto/hmac"
	"crypto/sha256"

	domainports "whisper/libs/domain/ports"
	"whisper/libs/domain/valueobject"
	apperrors "whisper/libs/shared/errors"
)

type RateLimiter interface {
	Allow(context.Context, string) (bool, error)
}

type Hub interface {
	BroadcastToUser(userID valueobject.UserID, data []byte) error
	BroadcastToConversation(conversationID valueobject.ConversationID, data []byte) error
}

type Service struct {
	jwtSecret []byte
	limiter   RateLimiter
	hub       Hub
}

func NewService(jwtSecret string, limiter RateLimiter, hub Hub) *Service {
	return &Service{jwtSecret: []byte(jwtSecret), limiter: limiter, hub: hub}
}

func (s *Service) ValidateToken(_ context.Context, token string) (valueobject.UserID, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", apperrors.New(apperrors.KindUnauthorized, "token_malformed", "malformed token", nil)
	}

	mac := hmac.New(sha256.New, s.jwtSecret)
	payload := parts[0] + "." + parts[1]
	_, _ = mac.Write([]byte(payload))
	expected := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(parts[2])) {
		return "", apperrors.New(apperrors.KindUnauthorized, "token_invalid", "invalid token signature", nil)
	}

	rawClaims, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", apperrors.New(apperrors.KindUnauthorized, "token_payload_invalid", "invalid token payload", err)
	}

	var claims struct {
		Sub string `json:"sub"`
		Exp int64  `json:"exp"`
	}
	if err := json.Unmarshal(rawClaims, &claims); err != nil {
		return "", apperrors.New(apperrors.KindUnauthorized, "token_claims_invalid", "invalid token claims", err)
	}
	if claims.Exp <= time.Now().Unix() {
		return "", apperrors.New(apperrors.KindUnauthorized, "token_expired", "token expired", nil)
	}
	userID, err := valueobject.ParseUserID(claims.Sub)
	if err != nil {
		return "", apperrors.New(apperrors.KindUnauthorized, "token_subject_invalid", "invalid subject", err)
	}
	return userID, nil
}

func (s *Service) HandleEnvelope(ctx context.Context, userID valueobject.UserID, envelope domainports.Envelope) error {
	allowed, err := s.limiter.Allow(ctx, string(userID))
	if err != nil {
		return apperrors.New(apperrors.KindUnavailable, "rate_limiter_unavailable", "rate limiter unavailable", err)
	}
	if !allowed {
		return apperrors.New(apperrors.KindForbidden, "rate_limited", "rate limit exceeded", nil)
	}

	data, err := json.Marshal(map[string]any{
		"type":    envelope.Type,
		"payload": envelope.Payload,
		"nonce":   envelope.Nonce,
		"userId":  userID,
	})
	if err != nil {
		return apperrors.New(apperrors.KindInternal, "envelope_encode_failed", "failed to encode envelope", err)
	}

	switch {
	case strings.HasPrefix(envelope.Type, "message."):
		conversationRaw, ok := envelope.Payload["conversationId"].(string)
		if !ok || conversationRaw == "" {
			return apperrors.New(apperrors.KindValidation, "conversation_id_required", "conversationId is required", nil)
		}
		conversationID, err := valueobject.ParseConversationID(conversationRaw)
		if err != nil {
			return apperrors.New(apperrors.KindValidation, "conversation_id_invalid", "invalid conversationId", err)
		}
		if err := s.hub.BroadcastToConversation(conversationID, data); err != nil {
			return apperrors.New(apperrors.KindUnavailable, "broadcast_failed", "conversation broadcast failed", err)
		}
		return nil
	case strings.HasPrefix(envelope.Type, "presence."):
		if err := s.hub.BroadcastToUser(userID, data); err != nil {
			return apperrors.New(apperrors.KindUnavailable, "presence_broadcast_failed", "presence broadcast failed", err)
		}
		return nil
	default:
		return apperrors.New(apperrors.KindValidation, "envelope_type_unsupported", fmt.Sprintf("unsupported envelope type: %s", envelope.Type), errors.New("unsupported type"))
	}
}
