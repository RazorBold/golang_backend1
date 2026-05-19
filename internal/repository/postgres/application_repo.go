package postgres

import (
	"context"
	"errors"

	"github.com/RazorBold/golang_backend1/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ApplicationRepo struct {
	db *pgxpool.Pool
}

func NewApplicationRepo(db *pgxpool.Pool) *ApplicationRepo {
	return &ApplicationRepo{db: db}
}

func (r *ApplicationRepo) Create(ctx context.Context, app *model.Application) error {
	q := `INSERT INTO applications (user_id, name, description, api_key)
	      VALUES ($1, $2, $3, $4)
	      RETURNING id, created_at, updated_at`
	return r.db.QueryRow(ctx, q, app.UserID, app.Name, app.Description, app.APIKey).
		Scan(&app.ID, &app.CreatedAt, &app.UpdatedAt)
}

func (r *ApplicationRepo) FindByID(ctx context.Context, id string) (*model.Application, error) {
	q := `SELECT id, user_id, name, description, api_key, created_at, updated_at
	      FROM applications WHERE id = $1`
	return r.scanOne(r.db.QueryRow(ctx, q, id))
}

func (r *ApplicationRepo) FindByUserID(ctx context.Context, userID string) ([]model.Application, error) {
	q := `SELECT id, user_id, name, description, api_key, created_at, updated_at
	      FROM applications WHERE user_id = $1
	      ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apps []model.Application
	for rows.Next() {
		app := model.Application{}
		if err := rows.Scan(&app.ID, &app.UserID, &app.Name, &app.Description,
			&app.APIKey, &app.CreatedAt, &app.UpdatedAt); err != nil {
			return nil, err
		}
		apps = append(apps, app)
	}
	return apps, rows.Err()
}

func (r *ApplicationRepo) FindByAPIKey(ctx context.Context, apiKey string) (*model.Application, error) {
	q := `SELECT id, user_id, name, description, api_key, created_at, updated_at
	      FROM applications WHERE api_key = $1`
	return r.scanOne(r.db.QueryRow(ctx, q, apiKey))
}

func (r *ApplicationRepo) Update(ctx context.Context, app *model.Application) error {
	q := `UPDATE applications SET name=$1, description=$2, updated_at=NOW()
	      WHERE id=$3 AND user_id=$4
	      RETURNING updated_at`
	err := r.db.QueryRow(ctx, q, app.Name, app.Description, app.ID, app.UserID).
		Scan(&app.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return pgx.ErrNoRows
	}
	return err
}

func (r *ApplicationRepo) Delete(ctx context.Context, id, userID string) error {
	q := `DELETE FROM applications WHERE id=$1 AND user_id=$2`
	tag, err := r.db.Exec(ctx, q, id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *ApplicationRepo) UpdateAPIKey(ctx context.Context, id, apiKey string) error {
	q := `UPDATE applications SET api_key=$1, updated_at=NOW() WHERE id=$2`
	_, err := r.db.Exec(ctx, q, apiKey, id)
	return err
}

func (r *ApplicationRepo) scanOne(row pgx.Row) (*model.Application, error) {
	app := &model.Application{}
	err := row.Scan(&app.ID, &app.UserID, &app.Name, &app.Description,
		&app.APIKey, &app.CreatedAt, &app.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return app, err
}
