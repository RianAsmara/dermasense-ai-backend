package database

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

// RunMigrations executes all SQL migration statements.
func RunMigrations(pool *pgxpool.Pool) error {
	ctx := context.Background()

	migrations := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id VARCHAR(36) PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			name VARCHAR(255) NOT NULL,
			picture TEXT DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,

		`CREATE TABLE IF NOT EXISTS scans (
			id VARCHAR(36) PRIMARY KEY,
			user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			image_url TEXT NOT NULL,
			diagnosis VARCHAR(255) NOT NULL DEFAULT '',
			confidence DOUBLE PRECISION NOT NULL DEFAULT 0,
			severity VARCHAR(50) NOT NULL DEFAULT 'unknown',
			explanation TEXT NOT NULL DEFAULT '',
			recommendations JSONB NOT NULL DEFAULT '[]',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,

		`CREATE INDEX IF NOT EXISTS idx_scans_user_id ON scans(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_scans_created_at ON scans(created_at DESC)`,

		`CREATE TABLE IF NOT EXISTS quotas (
			user_id VARCHAR(36) PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
			remaining_scans INTEGER NOT NULL DEFAULT 3,
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
	}

	for i, migration := range migrations {
		if _, err := pool.Exec(ctx, migration); err != nil {
			return fmt.Errorf("migration %d failed: %w", i+1, err)
		}
	}

	log.Println("✓ Database migrations completed successfully")
	return nil
}
