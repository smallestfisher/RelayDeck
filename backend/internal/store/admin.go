package store

import "github.com/smallestfisher/relaydeck/backend/internal/domain"

type UserReader interface {
	Users() []domain.User
	UserByEmail(email string) (domain.User, bool)
	UserByID(id string) (domain.User, bool)
}

type AdminStore struct {
	config *MemoryStore
	users  UserReader
}

func NewAdminStore(config *MemoryStore, users UserReader) *AdminStore {
	return &AdminStore{config: config, users: users}
}

func (s *AdminStore) APIKeys() []domain.APIKey {
	return s.config.APIKeys()
}

func (s *AdminStore) Models() []domain.Model {
	return s.config.Models()
}

func (s *AdminStore) Sites() []domain.UpstreamSite {
	return s.config.Sites()
}

func (s *AdminStore) Users() []domain.User {
	return s.users.Users()
}

func (s *AdminStore) UserByEmail(email string) (domain.User, bool) {
	return s.users.UserByEmail(email)
}

func (s *AdminStore) UserByID(id string) (domain.User, bool) {
	return s.users.UserByID(id)
}
