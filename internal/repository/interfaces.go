package repository

import (
	"context"
	"time"

	"github.com/RazorBold/golang_backend1/internal/model"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByID(ctx context.Context, id string) (*model.User, error)
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, userID, token string, expiresAt time.Time) error
	FindByToken(ctx context.Context, token string) (*model.RefreshToken, error)
	DeleteByToken(ctx context.Context, token string) error
}

type ApplicationRepository interface {
	Create(ctx context.Context, app *model.Application) error
	FindByID(ctx context.Context, id string) (*model.Application, error)
	FindByUserID(ctx context.Context, userID string) ([]model.Application, error)
	FindByAPIKey(ctx context.Context, apiKey string) (*model.Application, error)
	Update(ctx context.Context, app *model.Application) error
	Delete(ctx context.Context, id, userID string) error
	UpdateAPIKey(ctx context.Context, id, apiKey string) error
}

type DeviceRepository interface {
	Create(ctx context.Context, device *model.Device) error
	FindByID(ctx context.Context, id string) (*model.Device, error)
	FindByAppID(ctx context.Context, appID string) ([]model.Device, error)
	FindByIDAndAppID(ctx context.Context, id, appID string) (*model.Device, error)
	Update(ctx context.Context, device *model.Device) error
	Delete(ctx context.Context, id, appID string) error
	UpdateLastSeen(ctx context.Context, id string) error
}

type DeviceDataRepository interface {
	Create(ctx context.Context, deviceID string, payload []byte, receivedAt time.Time) error
	FindByDeviceID(ctx context.Context, deviceID string, from, to *time.Time, limit int) ([]model.TelemetryRecord, error)
	FindLatest(ctx context.Context, deviceID string) (*model.TelemetryRecord, error)
}
