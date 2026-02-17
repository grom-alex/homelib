package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/grom-alex/homelib/backend/internal/models"
)

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

func (r *UserRepo) Create(ctx context.Context, user *models.User) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO users (email, username, display_name, password_hash, role)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, created_at, updated_at`,
		user.Email, user.Username, user.DisplayName, user.PasswordHash, user.Role,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var u models.User
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, username, display_name, password_hash, role,
				is_active, last_login_at, created_at, updated_at
		 FROM users WHERE email = $1`, email,
	).Scan(&u.ID, &u.Email, &u.Username, &u.DisplayName, &u.PasswordHash,
		&u.Role, &u.IsActive, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return &u, nil
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (*models.User, error) {
	var u models.User
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, username, display_name, password_hash, role,
				is_active, last_login_at, created_at, updated_at
		 FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Email, &u.Username, &u.DisplayName, &u.PasswordHash,
		&u.Role, &u.IsActive, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return &u, nil
}

func (r *UserRepo) CountUsers(ctx context.Context) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count users: %w", err)
	}
	return count, nil
}

// RegisterUser atomically checks user count and creates a user within a transaction.
// First user gets "admin" role. Returns the created user.
func (r *UserRepo) RegisterUser(ctx context.Context, user *models.User, registrationEnabled bool) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Lock users table to prevent race condition on first registration
	var count int
	err = tx.QueryRow(ctx, "SELECT COUNT(*) FROM users FOR UPDATE").Scan(&count)
	if err != nil {
		return fmt.Errorf("count users: %w", err)
	}

	if !registrationEnabled && count > 0 {
		return fmt.Errorf("registration is disabled")
	}

	if count == 0 {
		user.Role = "admin"
	}

	err = tx.QueryRow(ctx,
		`INSERT INTO users (email, username, display_name, password_hash, role)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, created_at, updated_at`,
		user.Email, user.Username, user.DisplayName, user.PasswordHash, user.Role,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *UserRepo) UpdateLastLogin(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx,
		"UPDATE users SET last_login_at = NOW() WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("update last login: %w", err)
	}
	return nil
}
