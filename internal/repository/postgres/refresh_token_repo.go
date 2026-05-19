package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/RazorBold/golang_backend1/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RefreshTokenRepo struct {
	db *pgxpool.Pool
}

func NewRefreshTokenRepo(db *pgxpool.Pool) *RefreshTokenRepo {
	return &RefreshTokenRepo{db: db}
}

func (r *RefreshTokenRepo) Create(ctx context.Context, userID, token string, expiresAt time.Time) error {
	q := `INSERT INTO refresh_tokens (user_id, token, expires_at) VALUES ($1, $2, $3)`
	_, err := r.db.Exec(ctx, q, userID, token, expiresAt)
	return err
}

func (r *RefreshTokenRepo) FindByToken(ctx context.Context, token string) (*model.RefreshToken, error) {
	q := `SELECT id, user_id, token, expires_at, created_at
	      FROM refresh_tokens WHERE token = $1`
	rt := &model.RefreshToken{}
	err := r.db.QueryRow(ctx, q, token).
		Scan(&rt.ID, &rt.UserID, &rt.Token, &rt.ExpiresAt, &rt.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return rt, err
}

func (r *RefreshTokenRepo) DeleteByToken(ctx context.Context, token string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM refresh_tokens WHERE token = $1`, token)
	return err
}
