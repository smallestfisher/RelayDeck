package postgres

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/smallestfisher/relaydeck/backend/internal/auth"
	"github.com/smallestfisher/relaydeck/backend/internal/config"
)

func TestUserStoreBootstrapOwnerPersistsAdmin(t *testing.T) {
	config.LoadDotEnv()
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL is not set")
	}
	if !strings.Contains(databaseURL, "test") {
		t.Fatalf("DATABASE_URL must point to a test database, got %q", databaseURL)
	}
	databaseURL = isolatedDatabaseURL(t, databaseURL)
	ctx := context.Background()
	db, err := Open(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	defer db.Close()
	store := NewUserStore(db)
	if err := store.EnsureSchema(ctx); err != nil {
		t.Fatalf("ensure schema: %v", err)
	}
	if err := store.BootstrapOwner(ctx, "persisted-owner@example.com", "change-me"); err != nil {
		t.Fatalf("bootstrap owner: %v", err)
	}

	user, ok := store.UserByEmail("persisted-owner@example.com")
	if !ok {
		t.Fatal("expected bootstrapped owner to be persisted")
	}
	if user.Role != "owner" || user.Status != "active" {
		t.Fatalf("unexpected user role/status: %+v", user)
	}
	if !auth.VerifyPassword(user.PasswordHash, "change-me") {
		t.Fatal("expected persisted password hash to verify")
	}
}
