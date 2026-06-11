package store

import "github.com/smallestfisher/relaydeck/backend/internal/domain"

type UserReader interface {
	Users() []domain.User
	UserByEmail(email string) (domain.User, bool)
	UserByID(id string) (domain.User, bool)
}

type GatewayConfigReader interface {
	APIKeys() []domain.APIKey
	Models() []domain.Model
	Sites() []domain.UpstreamSite
}

type AdminStore struct {
	users     UserReader
	gateway   GatewayConfigReader
	upstreams UpstreamAccountStore
}

func NewAdminStore(users UserReader, gateway GatewayConfigReader, upstreams UpstreamAccountStore) *AdminStore {
	return &AdminStore{users: users, gateway: gateway, upstreams: upstreams}
}

func (s *AdminStore) APIKeys() []domain.APIKey {
	if s.gateway == nil {
		return nil
	}
	return s.gateway.APIKeys()
}

func (s *AdminStore) Models() []domain.Model {
	if s.gateway == nil {
		return nil
	}
	return s.gateway.Models()
}

func (s *AdminStore) Sites() []domain.UpstreamSite {
	if s.gateway == nil {
		return nil
	}
	return s.gateway.Sites()
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

func (s *AdminStore) Upstreams() UpstreamAccountStore {
	return s.upstreams
}
