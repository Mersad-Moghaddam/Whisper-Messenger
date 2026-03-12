package application

import (
	"context"
	"testing"
	"time"

	"whisper/libs/domain/entity"
	domainports "whisper/libs/domain/ports"
	"whisper/libs/domain/valueobject"
	"whisper/services/user/internal/adapters/presence"
	"whisper/services/user/internal/adapters/repository"
)

func TestUpdateAndPresence(t *testing.T) {
	uid := valueobject.NewUserID()
	repo := repository.NewUserRepository([]entity.User{{ID: uid, Username: "u", Email: "u@x", DisplayName: "U"}})
	pres := presence.NewPresenceRepository()
	svc := NewService(repo, pres, 2*time.Minute)
	ctx := context.Background()

	u, err := svc.UpdateMe(ctx, uid, domainports.UpdateProfileCommand{DisplayName: "Updated"})
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if u.DisplayName != "Updated" {
		t.Fatalf("display name not updated")
	}
	if err := svc.SetPresence(ctx, uid, domainports.PresenceOnline); err != nil {
		t.Fatalf("presence failed: %v", err)
	}
	st, _ := pres.GetState(ctx, uid)
	if st != domainports.PresenceOnline {
		t.Fatalf("presence mismatch")
	}
}
