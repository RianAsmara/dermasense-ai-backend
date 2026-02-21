package repository

import (
	"context"

	"github.com/dermasense/backend/internal/domain/entity"
)

// QuotaRepository defines the interface for quota data access.
type QuotaRepository interface {
	Create(ctx context.Context, quota *entity.Quota) error
	GetByUserID(ctx context.Context, userID string) (*entity.Quota, error)
	Update(ctx context.Context, quota *entity.Quota) error
	DecrementScans(ctx context.Context, userID string) error
	IncrementScans(ctx context.Context, userID string, amount int) error
}
