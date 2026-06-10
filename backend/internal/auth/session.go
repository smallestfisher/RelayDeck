package auth

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

type Session struct {
	Token      string
	UserID     string
	Email      string
	Role       string
	IssuedAt   time.Time
	ExpiresAt  time.Time
	LastSeenAt time.Time
}

type SessionStore interface {
	Create(session Session)
	Get(token string) (Session, bool)
	Delete(token string)
}

type MemorySessionStore struct {
	mu   sync.Mutex
	now  func() time.Time
	data map[string]Session
}

func NewMemorySessionStore(now func() time.Time) *MemorySessionStore {
	if now == nil {
		now = time.Now
	}
	return &MemorySessionStore{now: now, data: map[string]Session{}}
}

func NewSessionToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func (s *MemorySessionStore) Create(session Session) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[session.Token] = session
}

func (s *MemorySessionStore) Get(token string) (Session, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.data[token]
	if !ok {
		return Session{}, false
	}
	if !session.ExpiresAt.IsZero() && !s.now().Before(session.ExpiresAt) {
		delete(s.data, token)
		return Session{}, false
	}
	return session, true
}

func (s *MemorySessionStore) Delete(token string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, token)
}
