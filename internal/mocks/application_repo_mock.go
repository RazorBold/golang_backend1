package mocks

import (
	"context"

	"github.com/RazorBold/golang_backend1/internal/model"
	"github.com/stretchr/testify/mock"
)

type MockApplicationRepository struct {
	mock.Mock
}

func (m *MockApplicationRepository) Create(ctx context.Context, app *model.Application) error {
	return m.Called(ctx, app).Error(0)
}

func (m *MockApplicationRepository) FindByID(ctx context.Context, id string) (*model.Application, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Application), args.Error(1)
}

func (m *MockApplicationRepository) FindByUserID(ctx context.Context, userID string) ([]model.Application, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]model.Application), args.Error(1)
}

func (m *MockApplicationRepository) FindByAPIKey(ctx context.Context, apiKey string) (*model.Application, error) {
	args := m.Called(ctx, apiKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Application), args.Error(1)
}

func (m *MockApplicationRepository) Update(ctx context.Context, app *model.Application) error {
	return m.Called(ctx, app).Error(0)
}

func (m *MockApplicationRepository) Delete(ctx context.Context, id, userID string) error {
	return m.Called(ctx, id, userID).Error(0)
}

func (m *MockApplicationRepository) UpdateAPIKey(ctx context.Context, id, apiKey string) error {
	return m.Called(ctx, id, apiKey).Error(0)
}
