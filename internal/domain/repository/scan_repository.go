package repository

import (
	"context"

	"github.com/dermasense/backend/internal/domain/entity"
)

// ScanRepository defines the interface for scan data access.
type ScanRepository interface {
	Create(ctx context.Context, scan *entity.Scan) error
	GetByID(ctx context.Context, id string) (*entity.Scan, error)
	GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*entity.Scan, error)
	CountByUserID(ctx context.Context, userID string) (int, error)
}
