package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/dev-32/auth-service/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type TokenRepo struct {
	db *sqlx.DB
}

func NewTokenRepo(db *sqlx.DB) *TokenRepo {
	return &TokenRepo{db: db}
}

func (r *TokenRepo) Migrate() error {
	query := `
	CREATE TABLE IF NOT EXISTS refresh_tokens (
		id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		token      TEXT NOT NULL UNIQUE,
		expires_at TIMESTAMPTZ NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token ON refresh_tokens(token);
	CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);
	`
	_, err := r.db.Exec(query)
	return err
}

func (r *TokenRepo) Create(ctx context.Context, token *domain.RefreshToken) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO refresh_tokens (id, user_id, token, expires_at)
		 VALUES ($1, $2, $3, $4)`,
		token.ID, token.UserID, token.Token, token.ExpiresAt,
	)
	return err
}

func (r *TokenRepo) FindByToken(ctx context.Context, token string) (*domain.RefreshToken, error) {
	var t domain.RefreshToken
	err := r.db.QueryRowxContext(ctx,
		`SELECT id, user_id, token, expires_at, created_at
		 FROM refresh_tokens WHERE token = $1`, token).StructScan(&t)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrInvalidToken
	}
	return &t, err
}

func (r *TokenRepo) DeleteByToken(ctx context.Context, token string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM refresh_tokens WHERE token = $1`, token)
	return err
}

func (r *TokenRepo) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM refresh_tokens WHERE user_id = $1`, userID)
	return err
}
