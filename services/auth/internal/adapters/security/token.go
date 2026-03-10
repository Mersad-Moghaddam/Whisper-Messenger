package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"whisper/libs/domain/valueobject"
)

type TokenIssuer struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewTokenIssuer(secret string, accessTTL, refreshTTL time.Duration) *TokenIssuer {
	return &TokenIssuer{secret: []byte(secret), accessTTL: accessTTL, refreshTTL: refreshTTL}
}

func (i *TokenIssuer) IssueAccessToken(subject valueobject.UserID) (string, error) {
	return i.issue(map[string]any{"sub": subject, "exp": time.Now().Add(i.accessTTL).Unix(), "typ": "access"})
}

func (i *TokenIssuer) IssueRefreshToken(subject valueobject.UserID, sessionID valueobject.SessionID) (string, error) {
	return i.issue(map[string]any{"sub": subject, "sid": sessionID, "exp": time.Now().Add(i.refreshTTL).Unix(), "typ": "refresh"})
}

func (i *TokenIssuer) ParseRefreshToken(token string) (valueobject.UserID, valueobject.SessionID, error) {
	claims, err := i.parse(token)
	if err != nil {
		return "", "", err
	}
	if claims.Typ != "refresh" {
		return "", "", fmt.Errorf("invalid token type")
	}
	uid, err := valueobject.ParseUserID(claims.Sub)
	if err != nil {
		return "", "", err
	}
	sid, err := valueobject.ParseSessionID(claims.SID)
	if err != nil {
		return "", "", err
	}
	return uid, sid, nil
}

type claims struct {
	Sub string `json:"sub"`
	SID string `json:"sid,omitempty"`
	Exp int64  `json:"exp"`
	Typ string `json:"typ"`
}

func (i *TokenIssuer) issue(payload map[string]any) (string, error) {
	header, _ := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	p0 := base64.RawURLEncoding.EncodeToString(header)
	p1 := base64.RawURLEncoding.EncodeToString(body)
	sig := i.sign(p0 + "." + p1)
	return p0 + "." + p1 + "." + sig, nil
}

func (i *TokenIssuer) parse(token string) (claims, error) {
	var c claims
	parts := split3(token)
	if len(parts) != 3 {
		return c, fmt.Errorf("malformed token")
	}
	expected := i.sign(parts[0] + "." + parts[1])
	if !hmac.Equal([]byte(expected), []byte(parts[2])) {
		return c, fmt.Errorf("invalid token signature")
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return c, err
	}
	if err := json.Unmarshal(raw, &c); err != nil {
		return c, err
	}
	if c.Exp <= time.Now().Unix() {
		return c, fmt.Errorf("token expired")
	}
	return c, nil
}

func (i *TokenIssuer) sign(payload string) string {
	mac := hmac.New(sha256.New, i.secret)
	_, _ = mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func split3(v string) []string {
	out := make([]string, 0, 3)
	curr := ""
	for _, ch := range v {
		if ch == '.' {
			out = append(out, curr)
			curr = ""
			continue
		}
		curr += string(ch)
	}
	out = append(out, curr)
	return out
}
