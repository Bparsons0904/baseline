package repositories

import (
	"context"
	"server/config"
	. "server/internal/models"
)

type UserRepository interface {
	GetByID(ctx context.Context, id string) (*User, error)
	GetByLogin(ctx context.Context, login string) (*User, error)
	Create(ctx context.Context, user *User, config config.Config) error
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id string) error
}

type AdminRepository interface {
	GetByID(ctx context.Context, id string) (*User, error)
}

type SessionRepository interface {
	Create(ctx context.Context, session *Session, config config.Config) error
	GetByID(ctx context.Context, id string) (*Session, error)
	Delete(ctx context.Context, id string) error
}

