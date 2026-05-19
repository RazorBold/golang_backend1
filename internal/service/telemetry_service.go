package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/RazorBold/golang_backend1/internal/broker"
	"github.com/RazorBold/golang_backend1/internal/cache"
	"github.com/RazorBold/golang_backend1/internal/metrics"
	"github.com/RazorBold/golang_backend1/internal/model"
	"github.com/RazorBold/golang_backend1/internal/repository"
	"github.com/google/uuid"
)

type TelemetryService struct {
	deviceRepo repository.DeviceRepository
	dataRepo   repository.DeviceDataRepository
	appRepo    repository.ApplicationRepository
	publisher  *broker.Publisher
	redis      cache.Cache
}

func NewTelemetryService(
	deviceRepo repository.DeviceRepository,
	dataRepo repository.DeviceDataRepository,
	appRepo repository.ApplicationRepository,
	publisher *broker.Publisher,
	redis cache.Cache,
) *TelemetryService {
	return &TelemetryService{
		deviceRepo: deviceRepo,
		dataRepo:   dataRepo,
		appRepo:    appRepo,
		publisher:  publisher,
		redis:      redis,
	}
}

// Ingest menerima data dari device, publish ke RabbitMQ, cache latest di Redis.
func (s *TelemetryService) Ingest(ctx context.Context, appID, deviceID string, data json.RawMessage) (*model.IngestResponse, error) {
	// validasi device milik application ini
	device, err := s.deviceRepo.FindByIDAndAppID(ctx, deviceID, appID)
	if err != nil {
		return nil, err
	}
	if device == nil {
		return nil, ErrNotFound
	}

	receivedAt := time.Now().UTC()

	// update last_seen_at dan status='active' (non-blocking, best effort)
	go func() {
		_ = s.deviceRepo.UpdateLastSeen(context.Background(), deviceID)
	}()

	// publish ke RabbitMQ
	msg := broker.TelemetryMessage{
		EventID:    uuid.New().String(),
		DeviceID:   deviceID,
		AppID:      appID,
		ReceivedAt: receivedAt,
		Data:       data,
	}
	if err := s.publisher.PublishTelemetry(ctx, msg); err != nil {
		metrics.TelemetryIngestErrors.Inc()
		return nil, err
	}
	metrics.TelemetryIngestTotal.Inc()

	// cache latest telemetry (best effort, non-blocking)
	go func() {
		_ = s.redis.Set(context.Background(),
			fmt.Sprintf("device:%s:latest", deviceID),
			string(data),
			10*time.Second,
		)
	}()

	return &model.IngestResponse{Status: "ok", ReceivedAt: receivedAt}, nil
}

// GetData query telemetry dari DB dengan filter waktu.
func (s *TelemetryService) GetData(ctx context.Context, userID, deviceID string, from, to *time.Time, limit int) ([]model.TelemetryResponse, error) {
	if err := s.verifyDeviceOwnership(ctx, userID, deviceID); err != nil {
		return nil, err
	}

	records, err := s.dataRepo.FindByDeviceID(ctx, deviceID, from, to, limit)
	if err != nil {
		return nil, err
	}

	result := make([]model.TelemetryResponse, len(records))
	for i, r := range records {
		result[i] = model.TelemetryResponse{
			ID:         r.ID,
			DeviceID:   r.DeviceID,
			Payload:    r.Payload,
			ReceivedAt: r.ReceivedAt,
		}
	}
	return result, nil
}

// GetLatest cek Redis cache dulu, fallback ke DB.
func (s *TelemetryService) GetLatest(ctx context.Context, userID, deviceID string) (*model.TelemetryResponse, error) {
	if err := s.verifyDeviceOwnership(ctx, userID, deviceID); err != nil {
		return nil, err
	}

	// try Redis cache
	cacheKey := fmt.Sprintf("device:%s:latest", deviceID)
	cached, err := s.redis.Get(ctx, cacheKey)
	if err == nil && cached != "" {
		return &model.TelemetryResponse{
			DeviceID:   deviceID,
			Payload:    json.RawMessage(cached),
			ReceivedAt: time.Now(), // approximate, cache tidak simpan timestamp
		}, nil
	}

	// fallback ke DB
	rec, err := s.dataRepo.FindLatest(ctx, deviceID)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, ErrNotFound
	}
	return &model.TelemetryResponse{
		ID:         rec.ID,
		DeviceID:   rec.DeviceID,
		Payload:    rec.Payload,
		ReceivedAt: rec.ReceivedAt,
	}, nil
}

func (s *TelemetryService) verifyDeviceOwnership(ctx context.Context, userID, deviceID string) error {
	device, err := s.deviceRepo.FindByID(ctx, deviceID)
	if err != nil {
		return err
	}
	if device == nil {
		return ErrNotFound
	}
	app, err := s.appRepo.FindByID(ctx, device.ApplicationID)
	if err != nil {
		return err
	}
	if app == nil || app.UserID != userID {
		return ErrNotFound
	}
	return nil
}
