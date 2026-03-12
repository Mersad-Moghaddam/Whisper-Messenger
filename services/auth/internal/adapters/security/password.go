package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

type PasswordHasher struct{}

func NewPasswordHasher() *PasswordHasher { return &PasswordHasher{} }

func (h *PasswordHasher) Hash(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	s := hex.EncodeToString(salt)
	digest := sha256.Sum256([]byte(s + ":" + password))
	return fmt.Sprintf("sha256$%s$%s", s, hex.EncodeToString(digest[:])), nil
}

func (h *PasswordHasher) Verify(password, hash string) error {
	parts := strings.Split(hash, "$")
	if len(parts) != 3 || parts[0] != "sha256" {
		return fmt.Errorf("invalid hash format")
	}
	digest := sha256.Sum256([]byte(parts[1] + ":" + password))
	if hex.EncodeToString(digest[:]) != parts[2] {
		return fmt.Errorf("password mismatch")
	}
	return nil
}
