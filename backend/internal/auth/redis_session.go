package auth

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const redisSessionKeyPrefix = "relaydeck:session:"

type RedisSessionStore struct {
	client *redis.Client
	now    func() time.Time
}

func NewRedisSessionStore(redisURL string, now func() time.Time) (*RedisSessionStore, error) {
	if now == nil {
		now = time.Now
	}
	options, err := redis.ParseURL(redisURL)
	if err != nil {
		options = &redis.Options{Addr: redisURL}
	}
	return &RedisSessionStore{client: redis.NewClient(options), now: now}, nil
}

func (s *RedisSessionStore) Ping(ctx context.Context) error {
	return s.client.Ping(ctx).Err()
}

func (s *RedisSessionStore) Close() error {
	return s.client.Close()
}

func (s *RedisSessionStore) Create(session Session) error {
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}
	ttl := time.Duration(0)
	if !session.ExpiresAt.IsZero() {
		ttl = time.Until(session.ExpiresAt)
		if s.now != nil {
			ttl = session.ExpiresAt.Sub(s.now())
		}
		if ttl <= 0 {
			return s.client.Del(context.Background(), s.key(session.Token)).Err()
		}
	}
	return s.client.Set(context.Background(), s.key(session.Token), data, ttl).Err()
}

func (s *RedisSessionStore) Get(token string) (Session, bool) {
	data, err := s.client.Get(context.Background(), s.key(token)).Bytes()
	if errors.Is(err, redis.Nil) || err != nil {
		return Session{}, false
	}
	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		_ = s.client.Del(context.Background(), s.key(token)).Err()
		return Session{}, false
	}
	if !session.ExpiresAt.IsZero() && !s.now().Before(session.ExpiresAt) {
		_ = s.client.Del(context.Background(), s.key(token)).Err()
		return Session{}, false
	}
	return session, true
}

func (s *RedisSessionStore) Delete(token string) error {
	return s.client.Del(context.Background(), s.key(token)).Err()
}

func (s *RedisSessionStore) key(token string) string {
	return redisSessionKeyPrefix + strings.TrimSpace(token)
}
