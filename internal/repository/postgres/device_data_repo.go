package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/RazorBold/golang_backend1/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DeviceDataRepo struct {
	db *pgxpool.Pool
}

func NewDeviceDataRepo(db *pgxpool.Pool) *DeviceDataRepo {
	return &DeviceDataRepo{db: db}
}

func (r *DeviceDataRepo) Create(ctx context.Context, deviceID string, payload []byte, receivedAt time.Time) error {
	q := `INSERT INTO device_data (device_id, payload, received_at) VALUES ($1, $2, $3)`
	_, err := r.db.Exec(ctx, q, deviceID, payload, receivedAt)
	return err
}

func (r *DeviceDataRepo) FindByDeviceID(ctx context.Context, deviceID string, from, to *time.Time, limit int) ([]model.TelemetryRecord, error) {
	q := `SELECT id, device_id, payload, received_at
	      FROM device_data
	      WHERE device_id = $1
	        AND ($2::timestamptz IS NULL OR received_at >= $2)
	        AND ($3::timestamptz IS NULL OR received_at <= $3)
	      ORDER BY received_at DESC
	      LIMIT $4`
	rows, err := r.db.Query(ctx, q, deviceID, from, to, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []model.TelemetryRecord
	for rows.Next() {
		rec := model.TelemetryRecord{}
		var payloadRaw []byte
		if err := rows.Scan(&rec.ID, &rec.DeviceID, &payloadRaw, &rec.ReceivedAt); err != nil {
			return nil, err
		}
		rec.Payload = payloadRaw
		records = append(records, rec)
	}
	return records, rows.Err()
}

func (r *DeviceDataRepo) FindLatest(ctx context.Context, deviceID string) (*model.TelemetryRecord, error) {
	q := `SELECT id, device_id, payload, received_at
	      FROM device_data WHERE device_id = $1
	      ORDER BY received_at DESC LIMIT 1`
	rec := &model.TelemetryRecord{}
	var payloadRaw []byte
	err := r.db.QueryRow(ctx, q, deviceID).
		Scan(&rec.ID, &rec.DeviceID, &payloadRaw, &rec.ReceivedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	rec.Payload = payloadRaw
	return rec, nil
}
