package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/RazorBold/golang_backend1/internal/model"
	pgRepo "github.com/RazorBold/golang_backend1/internal/repository/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepo_CreateAndFindByEmail(t *testing.T) {
	ctx := context.Background()
	repo := pgRepo.NewUserRepo(testDB)

	user := &model.User{
		Name:     "Integration User",
		Email:    "integration@test.com",
		Password: "hashed_password",
		Role:     "user",
	}

	err := repo.Create(ctx, user)
	require.NoError(t, err)
	assert.NotEmpty(t, user.ID)
	assert.False(t, user.CreatedAt.IsZero())

	found, err := repo.FindByEmail(ctx, "integration@test.com")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, user.ID, found.ID)
	assert.Equal(t, "Integration User", found.Name)
	assert.Equal(t, "hashed_password", found.Password)
}

func TestUserRepo_FindByEmail_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := pgRepo.NewUserRepo(testDB)

	found, err := repo.FindByEmail(ctx, "nonexistent@test.com")
	require.NoError(t, err)
	assert.Nil(t, found)
}

func TestUserRepo_FindByID(t *testing.T) {
	ctx := context.Background()
	repo := pgRepo.NewUserRepo(testDB)

	user := &model.User{
		Name:     "ID User",
		Email:    "iduser@test.com",
		Password: "hashed",
		Role:     "user",
	}
	require.NoError(t, repo.Create(ctx, user))

	found, err := repo.FindByID(ctx, user.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "ID User", found.Name)
}

func TestUserRepo_FindByID_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := pgRepo.NewUserRepo(testDB)

	found, err := repo.FindByID(ctx, "00000000-0000-0000-0000-000000000000")
	require.NoError(t, err)
	assert.Nil(t, found)
}

func TestRefreshTokenRepo_CreateFindDelete(t *testing.T) {
	ctx := context.Background()
	userRepo := pgRepo.NewUserRepo(testDB)
	tokenRepo := pgRepo.NewRefreshTokenRepo(testDB)

	user := &model.User{
		Name:     "Token User",
		Email:    "tokenuser@test.com",
		Password: "hashed",
		Role:     "user",
	}
	require.NoError(t, userRepo.Create(ctx, user))

	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	err := tokenRepo.Create(ctx, user.ID, "my-unique-token-abc123", expiresAt)
	require.NoError(t, err)

	rt, err := tokenRepo.FindByToken(ctx, "my-unique-token-abc123")
	require.NoError(t, err)
	require.NotNil(t, rt)
	assert.Equal(t, user.ID, rt.UserID)

	err = tokenRepo.DeleteByToken(ctx, "my-unique-token-abc123")
	require.NoError(t, err)

	rt2, err := tokenRepo.FindByToken(ctx, "my-unique-token-abc123")
	require.NoError(t, err)
	assert.Nil(t, rt2)
}
