package service_test

import (
	"context"
	"testing"

	"github.com/RazorBold/golang_backend1/internal/mocks"
	"github.com/RazorBold/golang_backend1/internal/model"
	"github.com/RazorBold/golang_backend1/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newAppService(appRepo *mocks.MockApplicationRepository, cache *mocks.MockCache) *service.ApplicationService {
	return service.NewApplicationService(appRepo, cache)
}

func TestCreateApplication_Success(t *testing.T) {
	appRepo := new(mocks.MockApplicationRepository)
	cache := new(mocks.MockCache)

	appRepo.On("Create", mock.Anything, mock.MatchedBy(func(a *model.Application) bool {
		return a.UserID == "user-1" && a.Name == "My App" && len(a.APIKey) == 64
	})).Run(func(args mock.Arguments) {
		a := args.Get(1).(*model.Application)
		a.ID = "app-1"
	}).Return(nil)

	svc := newAppService(appRepo, cache)
	resp, err := svc.Create(context.Background(), "user-1", model.CreateApplicationRequest{
		Name: "My App",
	})

	require.NoError(t, err)
	assert.Equal(t, "app-1", resp.ID)
	assert.Equal(t, "My App", resp.Name)
	assert.Len(t, resp.APIKey, 64) // hex(32 bytes)
}

func TestGetApplication_NotOwner(t *testing.T) {
	appRepo := new(mocks.MockApplicationRepository)
	cache := new(mocks.MockCache)

	app := &model.Application{ID: "app-1", UserID: "other-user"}
	appRepo.On("FindByID", mock.Anything, "app-1").Return(app, nil)

	svc := newAppService(appRepo, cache)
	_, err := svc.Get(context.Background(), "user-1", "app-1")

	assert.ErrorIs(t, err, service.ErrNotFound)
}

func TestGetApplication_Success(t *testing.T) {
	appRepo := new(mocks.MockApplicationRepository)
	cache := new(mocks.MockCache)

	app := &model.Application{ID: "app-1", UserID: "user-1", Name: "Farm App", APIKey: "key123"}
	appRepo.On("FindByID", mock.Anything, "app-1").Return(app, nil)

	svc := newAppService(appRepo, cache)
	resp, err := svc.Get(context.Background(), "user-1", "app-1")

	require.NoError(t, err)
	assert.Equal(t, "Farm App", resp.Name)
}

func TestDeleteApplication_NotFound(t *testing.T) {
	appRepo := new(mocks.MockApplicationRepository)
	cache := new(mocks.MockCache)

	appRepo.On("Delete", mock.Anything, "app-x", "user-1").Return(
		// pgx.ErrNoRows disimulasikan dengan return dari Delete
		service.ErrNotFound,
	)

	svc := newAppService(appRepo, cache)
	err := svc.Delete(context.Background(), "user-1", "app-x")

	assert.ErrorIs(t, err, service.ErrNotFound)
}

func TestRegenerateAPIKey_Success(t *testing.T) {
	appRepo := new(mocks.MockApplicationRepository)
	cache := new(mocks.MockCache)

	app := &model.Application{ID: "app-1", UserID: "user-1", APIKey: "old-key-64chars"}
	appRepo.On("FindByID", mock.Anything, "app-1").Return(app, nil)
	cache.On("Del", mock.Anything, "apikey:old-key-64chars").Return(nil)
	appRepo.On("UpdateAPIKey", mock.Anything, "app-1", mock.AnythingOfType("string")).Return(nil)

	svc := newAppService(appRepo, cache)
	resp, err := svc.RegenerateAPIKey(context.Background(), "user-1", "app-1")

	require.NoError(t, err)
	assert.NotEqual(t, "old-key-64chars", resp.APIKey)
	assert.Len(t, resp.APIKey, 64)
	cache.AssertCalled(t, "Del", mock.Anything, "apikey:old-key-64chars")
}
