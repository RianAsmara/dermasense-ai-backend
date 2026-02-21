package postgres

import (
	"context"
	"time"

	"github.com/dermasense/backend/internal/domain/entity"
	"github.com/jackc/pgx/v5/pgxpool"
)

// QuotaRepo implements repository.QuotaRepository using PostgreSQL.
type QuotaRepo struct {
	pool *pgxpool.Pool
}

// NewQuotaRepo creates a new QuotaRepo.
func NewQuotaRepo(pool *pgxpool.Pool) *QuotaRepo {
	return &QuotaRepo{pool: pool}
}

// Create inserts a new quota record for a user.
func (r *QuotaRepo) Create(ctx context.Context, quota *entity.Quota) error {
	query := `
		INSERT INTO quotas (user_id, remaining_scans, updated_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id) DO NOTHING`

	_, err := r.pool.Exec(ctx, query, quota.UserID, quota.RemainingScans, time.Now())
	return err
}

// GetByUserID retrieves the quota for a user.
func (r *QuotaRepo) GetByUserID(ctx context.Context, userID string) (*entity.Quota, error) {
	quota := &entity.Quota{}
	query := `SELECT user_id, remaining_scans, updated_at FROM quotas WHERE user_id = $1`

	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&quota.UserID, &quota.RemainingScans, &quota.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return quota, nil
}

// Update modifies a user's quota.
func (r *QuotaRepo) Update(ctx context.Context, quota *entity.Quota) error {
	query := `UPDATE quotas SET remaining_scans = $1, updated_at = $2 WHERE user_id = $3`
	_, err := r.pool.Exec(ctx, query, quota.RemainingScans, time.Now(), quota.UserID)
	return err
}

// DecrementScans decreases remaining scans by 1.
func (r *QuotaRepo) DecrementScans(ctx context.Context, userID string) error {
	query := `
		UPDATE quotas
		SET remaining_scans = remaining_scans - 1, updated_at = $1
		WHERE user_id = $2 AND remaining_scans > 0`

	result, err := r.pool.Exec(ctx, query, time.Now(), userID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNoQuotaRemaining
	}
	return nil
}

// IncrementScans increases remaining scans by a given amount.
func (r *QuotaRepo) IncrementScans(ctx context.Context, userID string, amount int) error {
	query := `
		UPDATE quotas
		SET remaining_scans = remaining_scans + $1, updated_at = $2
		WHERE user_id = $3`

	_, err := r.pool.Exec(ctx, query, amount, time.Now(), userID)
	return err
}
