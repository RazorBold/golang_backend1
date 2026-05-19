package service

import (
	"context"
	"errors"

	"github.com/RazorBold/golang_backend1/internal/model"
	"github.com/RazorBold/golang_backend1/internal/repository"
	"github.com/jackc/pgx/v5"
)

type DeviceService struct {
	appRepo    repository.ApplicationRepository
	deviceRepo repository.DeviceRepository
}

func NewDeviceService(appRepo repository.ApplicationRepository, deviceRepo repository.DeviceRepository) *DeviceService {
	return &DeviceService{appRepo: appRepo, deviceRepo: deviceRepo}
}

func (s *DeviceService) Create(ctx context.Context, userID, appID string, req model.CreateDeviceRequest) (*model.DeviceResponse, error) {
	if err := s.verifyAppOwnership(ctx, userID, appID); err != nil {
		return nil, err
	}

	device := &model.Device{
		ApplicationID: appID,
		Name:          req.Name,
		Type:          req.Type,
		Metadata:      req.Metadata,
	}
	if err := s.deviceRepo.Create(ctx, device); err != nil {
		return nil, err
	}
	return toDeviceResponse(device), nil
}

func (s *DeviceService) List(ctx context.Context, userID, appID string) ([]model.DeviceResponse, error) {
	if err := s.verifyAppOwnership(ctx, userID, appID); err != nil {
		return nil, err
	}

	devices, err := s.deviceRepo.FindByAppID(ctx, appID)
	if err != nil {
		return nil, err
	}
	result := make([]model.DeviceResponse, len(devices))
	for i, d := range devices {
		result[i] = *toDeviceResponse(&d)
	}
	return result, nil
}

func (s *DeviceService) Get(ctx context.Context, userID, appID, deviceID string) (*model.DeviceResponse, error) {
	if err := s.verifyAppOwnership(ctx, userID, appID); err != nil {
		return nil, err
	}

	device, err := s.deviceRepo.FindByIDAndAppID(ctx, deviceID, appID)
	if err != nil {
		return nil, err
	}
	if device == nil {
		return nil, ErrNotFound
	}
	return toDeviceResponse(device), nil
}

func (s *DeviceService) Update(ctx context.Context, userID, appID, deviceID string, req model.UpdateDeviceRequest) (*model.DeviceResponse, error) {
	if err := s.verifyAppOwnership(ctx, userID, appID); err != nil {
		return nil, err
	}

	device, err := s.deviceRepo.FindByIDAndAppID(ctx, deviceID, appID)
	if err != nil {
		return nil, err
	}
	if device == nil {
		return nil, ErrNotFound
	}

	if req.Name != "" {
		device.Name = req.Name
	}
	if req.Type != "" {
		device.Type = req.Type
	}
	if req.Metadata != nil {
		device.Metadata = req.Metadata
	}

	if err := s.deviceRepo.Update(ctx, device); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toDeviceResponse(device), nil
}

func (s *DeviceService) Delete(ctx context.Context, userID, appID, deviceID string) error {
	if err := s.verifyAppOwnership(ctx, userID, appID); err != nil {
		return err
	}
	err := s.deviceRepo.Delete(ctx, deviceID, appID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func (s *DeviceService) verifyAppOwnership(ctx context.Context, userID, appID string) error {
	app, err := s.appRepo.FindByID(ctx, appID)
	if err != nil {
		return err
	}
	if app == nil || app.UserID != userID {
		return ErrNotFound
	}
	return nil
}

func toDeviceResponse(d *model.Device) *model.DeviceResponse {
	return &model.DeviceResponse{
		ID:            d.ID,
		ApplicationID: d.ApplicationID,
		Name:          d.Name,
		Type:          d.Type,
		Status:        d.Status,
		LastSeenAt:    d.LastSeenAt,
		Metadata:      d.Metadata,
		CreatedAt:     d.CreatedAt,
		UpdatedAt:     d.UpdatedAt,
	}
}
