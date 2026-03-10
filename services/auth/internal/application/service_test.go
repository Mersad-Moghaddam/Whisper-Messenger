package application

import (
	"context"
	"testing"
	"time"

	domainports "whisper/libs/domain/ports"
	"whisper/services/auth/internal/adapters/repository"
	"whisper/services/auth/internal/adapters/security"
)

func newSvc() *Service {
	return NewService(
		repository.NewUserRepository(),
		repository.NewSessionRepository(),
		security.NewPasswordHasher(),
		security.NewTokenIssuer("secret", time.Hour, 24*time.Hour),
	)
}

func TestRegisterAndLogin(t *testing.T) {
	svc := newSvc()
	ctx := context.Background()
	_, tokens, err := svc.Register(ctx, domainports.RegisterUserCommand{Username: "alice", Email: "alice@example.com", Password: "password123"})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Fatalf("expected tokens")
	}
	_, _, err = svc.Login(ctx, domainports.LoginCommand{Email: "alice@example.com", Password: "password123"})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
}

func TestRefresh(t *testing.T) {
	svc := newSvc()
	ctx := context.Background()
	_, tokens, err := svc.Register(ctx, domainports.RegisterUserCommand{Username: "bob", Email: "bob@example.com", Password: "password123"})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	rotated, err := svc.Refresh(ctx, tokens.RefreshToken)
	if err != nil {
		t.Fatalf("refresh failed: %v", err)
	}
	if rotated.AccessToken == "" {
		t.Fatalf("expected new access token")
	}
}
