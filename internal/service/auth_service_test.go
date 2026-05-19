package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/RazorBold/golang_backend1/internal/config"
	"github.com/RazorBold/golang_backend1/internal/mocks"
	"github.com/RazorBold/golang_backend1/internal/model"
	"github.com/RazorBold/golang_backend1/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var testJWTCfg = config.JWTConfig{
	Secret:          "test-secret-32-bytes-minimum-len",
	AccessTokenTTL:  15,
	RefreshTokenTTL: 7,
}

func newAuthService(userRepo *mocks.MockUserRepository, tokenRepo *mocks.MockRefreshTokenRepository, cache *mocks.MockCache) *service.AuthService {
	return service.NewAuthService(userRepo, tokenRepo, cache, testJWTCfg)
}

// ── Register ──────────────────────────────────────────────────

func TestRegister_Success(t *testing.T) {
	userRepo := new(mocks.MockUserRepository)
	tokenRepo := new(mocks.MockRefreshTokenRepository)
	cache := new(mocks.MockCache)

	userRepo.On("FindByEmail", mock.Anything, "new@test.com").Return(nil, nil)
	userRepo.On("Create", mock.Anything, mock.MatchedBy(func(u *model.User) bool {
		return u.Email == "new@test.com" && u.Name == "New User" && u.Role == "user"
	})).Run(func(args mock.Arguments) {
		u := args.Get(1).(*model.User)
		u.ID = "uuid-new"
	}).Return(nil)

	svc := newAuthService(userRepo, tokenRepo, cache)
	resp, err := svc.Register(context.Background(), model.RegisterRequest{
		Name:     "New User",
		Email:    "new@test.com",
		Password: "secret123",
	})

	require.NoError(t, err)
	assert.Equal(t, "uuid-new", resp.ID)
	assert.Equal(t, "new@test.com", resp.Email)
	assert.Equal(t, "user", resp.Role)
	userRepo.AssertExpectations(t)
}

func TestRegister_EmailTaken(t *testing.T) {
	userRepo := new(mocks.MockUserRepository)
	tokenRepo := new(mocks.MockRefreshTokenRepository)
	cache := new(mocks.MockCache)

	existing := &model.User{ID: "existing-id", Email: "taken@test.com"}
	userRepo.On("FindByEmail", mock.Anything, "taken@test.com").Return(existing, nil)

	svc := newAuthService(userRepo, tokenRepo, cache)
	_, err := svc.Register(context.Background(), model.RegisterRequest{
		Name:     "Someone",
		Email:    "taken@test.com",
		Password: "secret123",
	})

	assert.ErrorIs(t, err, service.ErrEmailTaken)
	userRepo.AssertExpectations(t)
}

func TestRegister_RepositoryError(t *testing.T) {
	userRepo := new(mocks.MockUserRepository)
	tokenRepo := new(mocks.MockRefreshTokenRepository)
	cache := new(mocks.MockCache)

	userRepo.On("FindByEmail", mock.Anything, "err@test.com").Return(nil, errors.New("db error"))

	svc := newAuthService(userRepo, tokenRepo, cache)
	_, err := svc.Register(context.Background(), model.RegisterRequest{
		Name:     "User",
		Email:    "err@test.com",
		Password: "secret123",
	})

	assert.Error(t, err)
	assert.NotErrorIs(t, err, service.ErrEmailTaken)
}

// ── Login ─────────────────────────────────────────────────────

func TestLogin_Success(t *testing.T) {
	userRepo := new(mocks.MockUserRepository)
	tokenRepo := new(mocks.MockRefreshTokenRepository)
	cache := new(mocks.MockCache)

	// bcrypt hash untuk "secret123" (cost=12)
	hashedPwd := "$2a$12$7qUWKeC6OZiGmg4K8APY9OJA5HgNQ3VQ0jMUGViC.vOzHNEGJwIr6"
	user := &model.User{ID: "user-1", Email: "user@test.com", Password: hashedPwd, Role: "user"}

	userRepo.On("FindByEmail", mock.Anything, "user@test.com").Return(user, nil)
	tokenRepo.On("Create", mock.Anything, "user-1", mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).Return(nil)

	svc := newAuthService(userRepo, tokenRepo, cache)
	resp, err := svc.Login(context.Background(), model.LoginRequest{
		Email:    "user@test.com",
		Password: "secret123",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
	assert.Equal(t, 900, resp.ExpiresIn) // 15 menit * 60
}

func TestLogin_UserNotFound(t *testing.T) {
	userRepo := new(mocks.MockUserRepository)
	tokenRepo := new(mocks.MockRefreshTokenRepository)
	cache := new(mocks.MockCache)

	userRepo.On("FindByEmail", mock.Anything, "ghost@test.com").Return(nil, nil)

	svc := newAuthService(userRepo, tokenRepo, cache)
	_, err := svc.Login(context.Background(), model.LoginRequest{
		Email:    "ghost@test.com",
		Password: "secret123",
	})

	assert.ErrorIs(t, err, service.ErrInvalidCreds)
}

func TestLogin_WrongPassword(t *testing.T) {
	userRepo := new(mocks.MockUserRepository)
	tokenRepo := new(mocks.MockRefreshTokenRepository)
	cache := new(mocks.MockCache)

	user := &model.User{
		ID:       "user-1",
		Email:    "user@test.com",
		Password: "$2a$12$7qUWKeC6OZiGmg4K8APY9OJA5HgNQ3VQ0jMUGViC.vOzHNEGJwIr6",
	}
	userRepo.On("FindByEmail", mock.Anything, "user@test.com").Return(user, nil)

	svc := newAuthService(userRepo, tokenRepo, cache)
	_, err := svc.Login(context.Background(), model.LoginRequest{
		Email:    "user@test.com",
		Password: "wrongpassword",
	})

	assert.ErrorIs(t, err, service.ErrInvalidCreds)
}

// ── Refresh ───────────────────────────────────────────────────

func TestRefresh_Success(t *testing.T) {
	userRepo := new(mocks.MockUserRepository)
	tokenRepo := new(mocks.MockRefreshTokenRepository)
	cache := new(mocks.MockCache)

	rt := &model.RefreshToken{
		ID:        "rt-1",
		UserID:    "user-1",
		Token:     "valid-refresh-token",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	user := &model.User{ID: "user-1", Email: "user@test.com", Role: "user"}

	tokenRepo.On("FindByToken", mock.Anything, "valid-refresh-token").Return(rt, nil)
	userRepo.On("FindByID", mock.Anything, "user-1").Return(user, nil)
	tokenRepo.On("DeleteByToken", mock.Anything, "valid-refresh-token").Return(nil)
	tokenRepo.On("Create", mock.Anything, "user-1", mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).Return(nil)

	svc := newAuthService(userRepo, tokenRepo, cache)
	resp, err := svc.Refresh(context.Background(), "valid-refresh-token")

	require.NoError(t, err)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
	assert.NotEqual(t, "valid-refresh-token", resp.RefreshToken) // harus token baru
	tokenRepo.AssertExpectations(t)
}

func TestRefresh_TokenNotFound(t *testing.T) {
	userRepo := new(mocks.MockUserRepository)
	tokenRepo := new(mocks.MockRefreshTokenRepository)
	cache := new(mocks.MockCache)

	tokenRepo.On("FindByToken", mock.Anything, "unknown-token").Return(nil, nil)

	svc := newAuthService(userRepo, tokenRepo, cache)
	_, err := svc.Refresh(context.Background(), "unknown-token")

	assert.ErrorIs(t, err, service.ErrInvalidToken)
}

func TestRefresh_ExpiredToken(t *testing.T) {
	userRepo := new(mocks.MockUserRepository)
	tokenRepo := new(mocks.MockRefreshTokenRepository)
	cache := new(mocks.MockCache)

	rt := &model.RefreshToken{
		Token:     "expired-token",
		UserID:    "user-1",
		ExpiresAt: time.Now().Add(-1 * time.Hour), // sudah expired
	}
	tokenRepo.On("FindByToken", mock.Anything, "expired-token").Return(rt, nil)

	svc := newAuthService(userRepo, tokenRepo, cache)
	_, err := svc.Refresh(context.Background(), "expired-token")

	assert.ErrorIs(t, err, service.ErrInvalidToken)
}

// ── Logout ────────────────────────────────────────────────────

func TestLogout_Success(t *testing.T) {
	userRepo := new(mocks.MockUserRepository)
	tokenRepo := new(mocks.MockRefreshTokenRepository)
	cache := new(mocks.MockCache)

	expiresAt := time.Now().Add(10 * time.Minute)
	tokenRepo.On("DeleteByToken", mock.Anything, "my-refresh-token").Return(nil)
	cache.On("Set", mock.Anything, "blacklist:my-jti", "1", mock.AnythingOfType("time.Duration")).Return(nil)

	svc := newAuthService(userRepo, tokenRepo, cache)
	err := svc.Logout(context.Background(), "my-refresh-token", "my-jti", expiresAt)

	require.NoError(t, err)
	tokenRepo.AssertExpectations(t)
	cache.AssertExpectations(t)
}
