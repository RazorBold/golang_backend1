package postgres

import (
	"context"
	"errors"

	"github.com/RazorBold/golang_backend1/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, user *model.User) error {
	q := `INSERT INTO users (name, email, password, role)
	      VALUES ($1, $2, $3, $4)
	      RETURNING id, created_at, updated_at`
	return r.db.QueryRow(ctx, q, user.Name, user.Email, user.Password, user.Role).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	q := `SELECT id, name, email, password, role, created_at, updated_at
	      FROM users WHERE email = $1`
	u := &model.User{}
	err := r.db.QueryRow(ctx, q, email).
		Scan(&u.ID, &u.Name, &u.Email, &u.Password, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return u, err
}

func (r *UserRepo) FindByID(ctx context.Context, id string) (*model.User, error) {
	q := `SELECT id, name, email, password, role, created_at, updated_at
	      FROM users WHERE id = $1`
	u := &model.User{}
	err := r.db.QueryRow(ctx, q, id).
		Scan(&u.ID, &u.Name, &u.Email, &u.Password, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return u, err
}
