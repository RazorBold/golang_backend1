package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/RazorBold/golang_backend1/internal/cache"
	"github.com/RazorBold/golang_backend1/internal/config"
	"github.com/RazorBold/golang_backend1/internal/model"
	"github.com/RazorBold/golang_backend1/internal/pkg/password"
	"github.com/RazorBold/golang_backend1/internal/pkg/token"
	"github.com/RazorBold/golang_backend1/internal/repository"
)

var (
	ErrEmailTaken   = errors.New("email already registered")
	ErrInvalidCreds = errors.New("invalid email or password")
	ErrInvalidToken = errors.New("invalid or expired token")
)

type AuthService struct {
	userRepo  repository.UserRepository
	tokenRepo repository.RefreshTokenRepository
	redis     cache.Cache
	cfg       config.JWTConfig
}

func NewAuthService(
	userRepo repository.UserRepository,
	tokenRepo repository.RefreshTokenRepository,
	redis cache.Cache,
	cfg config.JWTConfig,
) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		redis:     redis,
		cfg:       cfg,
	}
}

func (s *AuthService) Register(ctx context.Context, req model.RegisterRequest) (*model.UserResponse, error) {
	existing, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrEmailTaken
	}

	hashed, err := password.Hash(req.Password)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashed,
		Role:     "user",
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return &model.UserResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Role:  user.Role,
	}, nil
}

func (s *AuthService) Login(ctx context.Context, req model.LoginRequest) (*model.AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil || !password.Compare(user.Password, req.Password) {
		return nil, ErrInvalidCreds
	}
	return s.issueTokenPair(ctx, user.ID, user.Role)
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*model.AuthResponse, error) {
	rt, err := s.tokenRepo.FindByToken(ctx, refreshToken)
	if err != nil {
		return nil, err
	}
	if rt == nil || time.Now().After(rt.ExpiresAt) {
		return nil, ErrInvalidToken
	}

	user, err := s.userRepo.FindByID(ctx, rt.UserID)
	if err != nil || user == nil {
		return nil, ErrInvalidToken
	}

	// rotate: hapus token lama sebelum issue baru
	if err := s.tokenRepo.DeleteByToken(ctx, refreshToken); err != nil {
		return nil, err
	}
	return s.issueTokenPair(ctx, user.ID, user.Role)
}

func (s *AuthService) Logout(ctx context.Context, refreshToken, jti string, expiresAt time.Time) error {
	if err := s.tokenRepo.DeleteByToken(ctx, refreshToken); err != nil {
		return err
	}
	// blacklist JTI di Redis sampai access token expired
	if ttl := time.Until(expiresAt); ttl > 0 {
		return s.redis.Set(ctx, fmt.Sprintf("blacklist:%s", jti), "1", ttl)
	}
	return nil
}

func (s *AuthService) issueTokenPair(ctx context.Context, userID, role string) (*model.AuthResponse, error) {
	accessToken, _, err := token.Generate(userID, role, s.cfg.Secret, s.cfg.AccessTokenTTL)
	if err != nil {
		return nil, err
	}

	refreshToken, err := generateSecureToken()
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(time.Duration(s.cfg.RefreshTokenTTL) * 24 * time.Hour)
	if err := s.tokenRepo.Create(ctx, userID, refreshToken, expiresAt); err != nil {
		return nil, err
	}

	return &model.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    s.cfg.AccessTokenTTL * 60,
	}, nil
}

func generateSecureToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
