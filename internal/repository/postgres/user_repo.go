package postgres

import (
	"context"
	"time"

	"github.com/dermasense/backend/internal/domain/entity"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserRepo implements repository.UserRepository using PostgreSQL.
type UserRepo struct {
	pool *pgxpool.Pool
}

// NewUserRepo creates a new UserRepo.
func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

// Create inserts a new user into the database.
func (r *UserRepo) Create(ctx context.Context, user *entity.User) error {
	query := `
		INSERT INTO users (id, email, name, picture, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (email) DO UPDATE SET
			name = EXCLUDED.name,
			picture = EXCLUDED.picture
		RETURNING id, created_at`

	return r.pool.QueryRow(ctx, query,
		user.ID, user.Email, user.Name, user.Picture, time.Now(),
	).Scan(&user.ID, &user.CreatedAt)
}

// GetByID retrieves a user by their ID.
func (r *UserRepo) GetByID(ctx context.Context, id string) (*entity.User, error) {
	user := &entity.User{}
	query := `SELECT id, email, name, picture, created_at FROM users WHERE id = $1`

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.Name, &user.Picture, &user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetByEmail retrieves a user by their email.
func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	user := &entity.User{}
	query := `SELECT id, email, name, picture, created_at FROM users WHERE email = $1`

	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.Name, &user.Picture, &user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}
