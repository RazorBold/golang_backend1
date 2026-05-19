package mocks

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

type MockCache struct {
	mock.Mock
}

func (m *MockCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	return m.Called(ctx, key, value, ttl).Error(0)
}

func (m *MockCache) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockCache) Del(ctx context.Context, keys ...string) error {
	args := make([]any, len(keys)+1)
	args[0] = ctx
	for i, k := range keys {
		args[i+1] = k
	}
	return m.Called(args...).Error(0)
}

func (m *MockCache) Exists(ctx context.Context, key string) (bool, error) {
	args := m.Called(ctx, key)
	return args.Bool(0), args.Error(1)
}

func (m *MockCache) Incr(ctx context.Context, key string) (int64, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return m.Called(ctx, key, ttl).Error(0)
}

func (m *MockCache) Ping(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

func (m *MockCache) Close() error {
	return m.Called().Error(0)
}
