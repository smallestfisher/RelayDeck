package auth

import (
	"os"
	"testing"
	"time"

	"github.com/smallestfisher/relaydeck/backend/internal/config"
)

func TestRedisSessionLifecycle(t *testing.T) {
	config.LoadDotEnv()
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		t.Skip("REDIS_URL is not set")
	}
	now := fixedSessionNow()
	store, err := NewRedisSessionStore(redisURL, func() time.Time { return now })
	if err != nil {
		t.Fatalf("new redis session store: %v", err)
	}
	defer store.Close()
	session := Session{
		Token:      "redis-session-test-token",
		UserID:     "user_redis",
		Email:      "redis-owner@example.com",
		Role:       "owner",
		IssuedAt:   now,
		ExpiresAt:  now.Add(30 * time.Minute),
		LastSeenAt: now,
	}
	store.Delete(session.Token)

	store.Create(session)
	got, ok := store.Get(session.Token)
	if !ok {
		t.Fatal("expected session to be stored in Redis")
	}
	if got.UserID != session.UserID || got.Email != session.Email || got.Role != session.Role {
		t.Fatalf("unexpected session: %+v", got)
	}

	store.Delete(session.Token)
	if _, ok := store.Get(session.Token); ok {
		t.Fatal("expected session to be deleted from Redis")
	}
}
