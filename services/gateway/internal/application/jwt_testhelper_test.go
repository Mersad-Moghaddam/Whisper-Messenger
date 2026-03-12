package application

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

func sign(secret, payload string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
