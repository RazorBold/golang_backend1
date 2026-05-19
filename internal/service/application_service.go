package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"

	"github.com/RazorBold/golang_backend1/internal/cache"
	"github.com/RazorBold/golang_backend1/internal/model"
	"github.com/RazorBold/golang_backend1/internal/repository"
	"github.com/jackc/pgx/v5"
)

var ErrNotFound = errors.New("resource not found")

type ApplicationService struct {
	appRepo repository.ApplicationRepository
	redis   cache.Cache
}

func NewApplicationService(appRepo repository.ApplicationRepository, redis cache.Cache) *ApplicationService {
	return &ApplicationService{appRepo: appRepo, redis: redis}
}

func (s *ApplicationService) Create(ctx context.Context, userID string, req model.CreateApplicationRequest) (*model.ApplicationResponse, error) {
	apiKey, err := generateAPIKey()
	if err != nil {
		return nil, err
	}

	app := &model.Application{
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		APIKey:      apiKey,
	}
	if err := s.appRepo.Create(ctx, app); err != nil {
		return nil, err
	}
	return toApplicationResponse(app), nil
}

func (s *ApplicationService) List(ctx context.Context, userID string) ([]model.ApplicationResponse, error) {
	apps, err := s.appRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	result := make([]model.ApplicationResponse, len(apps))
	for i, a := range apps {
		result[i] = *toApplicationResponse(&a)
	}
	return result, nil
}

func (s *ApplicationService) Get(ctx context.Context, userID, id string) (*model.ApplicationResponse, error) {
	app, err := s.appRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if app == nil || app.UserID != userID {
		return nil, ErrNotFound
	}
	return toApplicationResponse(app), nil
}

func (s *ApplicationService) Update(ctx context.Context, userID, id string, req model.UpdateApplicationRequest) (*model.ApplicationResponse, error) {
	app, err := s.appRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if app == nil || app.UserID != userID {
		return nil, ErrNotFound
	}

	if req.Name != "" {
		app.Name = req.Name
	}
	app.Description = req.Description

	if err := s.appRepo.Update(ctx, app); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toApplicationResponse(app), nil
}

func (s *ApplicationService) Delete(ctx context.Context, userID, id string) error {
	err := s.appRepo.Delete(ctx, id, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func (s *ApplicationService) RegenerateAPIKey(ctx context.Context, userID, id string) (*model.ApplicationResponse, error) {
	app, err := s.appRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if app == nil || app.UserID != userID {
		return nil, ErrNotFound
	}

	// invalidate old key dari Redis cache
	_ = s.redis.Del(ctx, "apikey:"+app.APIKey)

	newKey, err := generateAPIKey()
	if err != nil {
		return nil, err
	}
	if err := s.appRepo.UpdateAPIKey(ctx, id, newKey); err != nil {
		return nil, err
	}
	app.APIKey = newKey
	return toApplicationResponse(app), nil
}

func toApplicationResponse(app *model.Application) *model.ApplicationResponse {
	return &model.ApplicationResponse{
		ID:          app.ID,
		Name:        app.Name,
		Description: app.Description,
		APIKey:      app.APIKey,
		CreatedAt:   app.CreatedAt,
		UpdatedAt:   app.UpdatedAt,
	}
}

func generateAPIKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
