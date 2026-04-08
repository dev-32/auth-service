package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/dev-32/auth-service/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type UserRepo struct {
	db *sqlx.DB
}

func NewUserRepo(db *sqlx.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Migrate() error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		name           TEXT NOT NULL,
		email          TEXT NOT NULL UNIQUE,
		password       TEXT NOT NULL DEFAULT '',
		role           TEXT NOT NULL DEFAULT 'user',
		provider       TEXT NOT NULL DEFAULT 'local',
		provider_id    TEXT NOT NULL DEFAULT '',
		email_verified BOOLEAN NOT NULL DEFAULT FALSE,
		created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
	CREATE INDEX IF NOT EXISTS idx_users_provider ON users(provider, provider_id);
	`
	_, err := r.db.Exec(query)
	return err
}

func (r *UserRepo) Create(ctx context.Context, user *domain.User) error {
	query := `
	INSERT INTO users (id, name, email, password, role, provider, provider_id, email_verified)
	VALUES (:id, :name, :email, :password, :role, :provider, :provider_id, :email_verified)`
	_, err := r.db.NamedExecContext(ctx, query, user)
	return err
}

func (r *UserRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var user domain.User
	err := r.db.GetContext(ctx, &user, `SELECT * FROM users WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	return &user, err
}

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	err := r.db.GetContext(ctx, &user, `SELECT * FROM users WHERE email = $1`, email)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	return &user, err
}

func (r *UserRepo) FindByProviderID(ctx context.Context, provider domain.Provider, providerID string) (*domain.User, error) {
	var user domain.User
	err := r.db.GetContext(ctx, &user,
		`SELECT * FROM users WHERE provider = $1 AND provider_id = $2`,
		provider, providerID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	return &user, err
}

func (r *UserRepo) Update(ctx context.Context, user *domain.User) error {
	user.UpdatedAt = time.Now()
	query := `
	UPDATE users SET
		name = :name,
		email = :email,
		password = :password,
		role = :role,
		email_verified = :email_verified,
		updated_at = :updated_at
	WHERE id = :id`
	_, err := r.db.NamedExecContext(ctx, query, user)
	return err
}

func (r *UserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, id)
	return err

}

func (r *UserRepo) List(ctx context.Context, limit, offset int) ([]domain.User, error) {
	var users []domain.User
	err := r.db.SelectContext(ctx, &users,
		`SELECT * FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	return users, err
}
