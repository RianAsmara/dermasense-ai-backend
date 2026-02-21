package repository

import (
	"context"

	"github.com/dermasense/backend/internal/domain/entity"
)

// UserRepository defines the interface for user data access.
type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, id string) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
}
