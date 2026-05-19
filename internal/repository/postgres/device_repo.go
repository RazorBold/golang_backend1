package postgres

import (
	"context"
	"errors"

	"github.com/RazorBold/golang_backend1/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DeviceRepo struct {
	db *pgxpool.Pool
}

func NewDeviceRepo(db *pgxpool.Pool) *DeviceRepo {
	return &DeviceRepo{db: db}
}

func (r *DeviceRepo) Create(ctx context.Context, device *model.Device) error {
	q := `INSERT INTO devices (application_id, name, type, metadata)
	      VALUES ($1, $2, $3, $4)
	      RETURNING id, status, created_at, updated_at`
	var metadataRaw []byte
	if device.Metadata != nil {
		metadataRaw = device.Metadata
	}
	return r.db.QueryRow(ctx, q, device.ApplicationID, device.Name, device.Type, metadataRaw).
		Scan(&device.ID, &device.Status, &device.CreatedAt, &device.UpdatedAt)
}

func (r *DeviceRepo) FindByID(ctx context.Context, id string) (*model.Device, error) {
	q := `SELECT id, application_id, name, type, status, last_seen_at, metadata, created_at, updated_at
	      FROM devices WHERE id = $1`
	return r.scanOne(r.db.QueryRow(ctx, q, id))
}

func (r *DeviceRepo) FindByAppID(ctx context.Context, appID string) ([]model.Device, error) {
	q := `SELECT id, application_id, name, type, status, last_seen_at, metadata, created_at, updated_at
	      FROM devices WHERE application_id = $1
	      ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, q, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []model.Device
	for rows.Next() {
		d, err := r.scanRow(rows)
		if err != nil {
			return nil, err
		}
		devices = append(devices, *d)
	}
	return devices, rows.Err()
}

func (r *DeviceRepo) FindByIDAndAppID(ctx context.Context, id, appID string) (*model.Device, error) {
	q := `SELECT id, application_id, name, type, status, last_seen_at, metadata, created_at, updated_at
	      FROM devices WHERE id = $1 AND application_id = $2`
	return r.scanOne(r.db.QueryRow(ctx, q, id, appID))
}

func (r *DeviceRepo) Update(ctx context.Context, device *model.Device) error {
	q := `UPDATE devices SET name=$1, type=$2, metadata=$3, updated_at=NOW()
	      WHERE id=$4 AND application_id=$5
	      RETURNING updated_at`
	var metadataRaw []byte
	if device.Metadata != nil {
		metadataRaw = device.Metadata
	}
	err := r.db.QueryRow(ctx, q, device.Name, device.Type, metadataRaw, device.ID, device.ApplicationID).
		Scan(&device.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return pgx.ErrNoRows
	}
	return err
}

func (r *DeviceRepo) Delete(ctx context.Context, id, appID string) error {
	q := `DELETE FROM devices WHERE id=$1 AND application_id=$2`
	tag, err := r.db.Exec(ctx, q, id, appID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *DeviceRepo) UpdateLastSeen(ctx context.Context, id string) error {
	q := `UPDATE devices SET status='active', last_seen_at=NOW(), updated_at=NOW() WHERE id=$1`
	_, err := r.db.Exec(ctx, q, id)
	return err
}

func (r *DeviceRepo) scanOne(row pgx.Row) (*model.Device, error) {
	d := &model.Device{}
	var metadataRaw []byte
	err := row.Scan(&d.ID, &d.ApplicationID, &d.Name, &d.Type, &d.Status,
		&d.LastSeenAt, &metadataRaw, &d.CreatedAt, &d.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	d.Metadata = metadataRaw
	return d, nil
}

func (r *DeviceRepo) scanRow(rows pgx.Rows) (*model.Device, error) {
	d := &model.Device{}
	var metadataRaw []byte
	err := rows.Scan(&d.ID, &d.ApplicationID, &d.Name, &d.Type, &d.Status,
		&d.LastSeenAt, &metadataRaw, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, err
	}
	d.Metadata = metadataRaw
	return d, nil
}
