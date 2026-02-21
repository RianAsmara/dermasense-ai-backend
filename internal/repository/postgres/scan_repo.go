package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/dermasense/backend/internal/domain/entity"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ScanRepo implements repository.ScanRepository using PostgreSQL.
type ScanRepo struct {
	pool *pgxpool.Pool
}

// NewScanRepo creates a new ScanRepo.
func NewScanRepo(pool *pgxpool.Pool) *ScanRepo {
	return &ScanRepo{pool: pool}
}

// Create inserts a new scan result into the database.
func (r *ScanRepo) Create(ctx context.Context, scan *entity.Scan) error {
	recommendations, err := json.Marshal(scan.Recommendations)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO scans (id, user_id, image_url, diagnosis, confidence, severity, explanation, recommendations, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err = r.pool.Exec(ctx, query,
		scan.ID, scan.UserID, scan.ImageURL, scan.Diagnosis,
		scan.Confidence, scan.Severity, scan.Explanation,
		recommendations, time.Now(),
	)
	return err
}

// GetByID retrieves a scan by its ID.
func (r *ScanRepo) GetByID(ctx context.Context, id string) (*entity.Scan, error) {
	scan := &entity.Scan{}
	var recommendations []byte

	query := `
		SELECT id, user_id, image_url, diagnosis, confidence, severity, explanation, recommendations, created_at
		FROM scans WHERE id = $1`

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&scan.ID, &scan.UserID, &scan.ImageURL, &scan.Diagnosis,
		&scan.Confidence, &scan.Severity, &scan.Explanation,
		&recommendations, &scan.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(recommendations, &scan.Recommendations); err != nil {
		return nil, err
	}
	return scan, nil
}

// GetByUserID retrieves scans for a user with pagination.
func (r *ScanRepo) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*entity.Scan, error) {
	query := `
		SELECT id, user_id, image_url, diagnosis, confidence, severity, explanation, recommendations, created_at
		FROM scans WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scans []*entity.Scan
	for rows.Next() {
		scan := &entity.Scan{}
		var recommendations []byte

		err := rows.Scan(
			&scan.ID, &scan.UserID, &scan.ImageURL, &scan.Diagnosis,
			&scan.Confidence, &scan.Severity, &scan.Explanation,
			&recommendations, &scan.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(recommendations, &scan.Recommendations); err != nil {
			return nil, err
		}
		scans = append(scans, scan)
	}

	return scans, rows.Err()
}

// CountByUserID returns the total number of scans for a user.
func (r *ScanRepo) CountByUserID(ctx context.Context, userID string) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM scans WHERE user_id = $1`

	err := r.pool.QueryRow(ctx, query, userID).Scan(&count)
	return count, err
}
