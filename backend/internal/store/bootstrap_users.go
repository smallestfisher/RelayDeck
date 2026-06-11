package store

import (
	"time"

	"github.com/smallestfisher/relaydeck/backend/internal/auth"
	"github.com/smallestfisher/relaydeck/backend/internal/domain"
)

const (
	DefaultOwnerEmail    = "owner@example.com"
	DefaultOwnerPassword = "change-me"
)

type BootstrapUserStore struct {
	users []domain.User
}

func NewBootstrapUserStore(ownerEmail string, ownerPassword string, now time.Time) *BootstrapUserStore {
	if ownerEmail == "" {
		ownerEmail = DefaultOwnerEmail
	}
	if ownerPassword == "" {
		ownerPassword = DefaultOwnerPassword
	}
	hash, err := auth.HashPassword(ownerPassword)
	if err != nil {
		panic(err)
	}
	return &BootstrapUserStore{users: []domain.User{
		{
			ID:           "user_admin",
			Email:        ownerEmail,
			PasswordHash: hash,
			Role:         domain.UserRoleOwner,
			Status:       domain.UserStatusActive,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	}}
}

func (s *BootstrapUserStore) Users() []domain.User {
	return append([]domain.User(nil), s.users...)
}

func (s *BootstrapUserStore) UserByEmail(email string) (domain.User, bool) {
	for _, user := range s.users {
		if user.Email == email {
			return user, true
		}
	}
	return domain.User{}, false
}

func (s *BootstrapUserStore) UserByID(id string) (domain.User, bool) {
	for _, user := range s.users {
		if user.ID == id {
			return user, true
		}
	}
	return domain.User{}, false
}
