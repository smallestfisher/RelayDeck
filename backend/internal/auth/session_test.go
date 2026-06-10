package auth

import (
	"testing"
	"time"
)

func TestSessionLifecycle(t *testing.T) {
	now := fixedSessionNow()
	store := NewMemorySessionStore(func() time.Time { return now })
	session := Session{
		Token:      "session-token",
		UserID:     "user_1",
		Email:      "admin@example.com",
		Role:       "owner",
		IssuedAt:   now,
		ExpiresAt:  now.Add(30 * time.Minute),
		LastSeenAt: now,
	}

	store.Create(session)
	got, ok := store.Get("session-token")
	if !ok {
		t.Fatal("expected session to be stored")
	}
	if got.UserID != session.UserID || got.Email != session.Email || got.Role != session.Role {
		t.Fatalf("unexpected session: %+v", got)
	}

	store.Delete("session-token")
	if _, ok := store.Get("session-token"); ok {
		t.Fatal("expected session to be deleted")
	}

	expiredStore := NewMemorySessionStore(func() time.Time { return now.Add(31 * time.Minute) })
	expiredStore.Create(session)
	if _, ok := expiredStore.Get("session-token"); ok {
		t.Fatal("expected expired session to be rejected")
	}
}

func fixedSessionNow() time.Time {
	return time.Date(2026, 6, 11, 12, 0, 0, 0, time.UTC)
}
