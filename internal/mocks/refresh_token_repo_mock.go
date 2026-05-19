package mocks

import (
	"context"
	"time"

	"github.com/RazorBold/golang_backend1/internal/model"
	"github.com/stretchr/testify/mock"
)

type MockRefreshTokenRepository struct {
	mock.Mock
}

func (m *MockRefreshTokenRepository) Create(ctx context.Context, userID, token string, expiresAt time.Time) error {
	return m.Called(ctx, userID, token, expiresAt).Error(0)
}

func (m *MockRefreshTokenRepository) FindByToken(ctx context.Context, token string) (*model.RefreshToken, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.RefreshToken), args.Error(1)
}

func (m *MockRefreshTokenRepository) DeleteByToken(ctx context.Context, token string) error {
	return m.Called(ctx, token).Error(0)
}
