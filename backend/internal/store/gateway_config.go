package store

import "github.com/smallestfisher/relaydeck/backend/internal/domain"

type StaticGatewayConfigStore struct {
	apiKeys []domain.APIKey
	models  []domain.Model
	sites   []domain.UpstreamSite
}

func NewStaticGatewayConfigStore(apiKeys []domain.APIKey, models []domain.Model, sites []domain.UpstreamSite) *StaticGatewayConfigStore {
	return &StaticGatewayConfigStore{
		apiKeys: append([]domain.APIKey(nil), apiKeys...),
		models:  append([]domain.Model(nil), models...),
		sites:   append([]domain.UpstreamSite(nil), sites...),
	}
}

func (s *StaticGatewayConfigStore) APIKeys() []domain.APIKey {
	return append([]domain.APIKey(nil), s.apiKeys...)
}

func (s *StaticGatewayConfigStore) Models() []domain.Model {
	return append([]domain.Model(nil), s.models...)
}

func (s *StaticGatewayConfigStore) Sites() []domain.UpstreamSite {
	return append([]domain.UpstreamSite(nil), s.sites...)
}
