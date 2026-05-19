package token_test

import (
	"testing"
	"time"

	"github.com/RazorBold/golang_backend1/internal/pkg/token"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSecret = "test-secret-32-bytes-minimum-len"

func TestGenerate_ValidToken(t *testing.T) {
	signed, jti, err := token.Generate("user-123", "user", testSecret, 15)

	require.NoError(t, err)
	assert.NotEmpty(t, signed)
	assert.NotEmpty(t, jti)
}

func TestValidate_Success(t *testing.T) {
	signed, jti, err := token.Generate("user-123", "user", testSecret, 15)
	require.NoError(t, err)

	claims, err := token.Validate(signed, testSecret)

	require.NoError(t, err)
	assert.Equal(t, "user-123", claims.UserID)
	assert.Equal(t, "user", claims.Role)
	assert.Equal(t, jti, claims.JTI)
	assert.True(t, claims.ExpiresAt.After(time.Now()))
}

func TestValidate_WrongSecret(t *testing.T) {
	signed, _, err := token.Generate("user-123", "user", testSecret, 15)
	require.NoError(t, err)

	_, err = token.Validate(signed, "wrong-secret-32-bytes-minimum-len")

	assert.Error(t, err)
}

func TestValidate_MalformedToken(t *testing.T) {
	_, err := token.Validate("not.a.valid.token", testSecret)
	assert.Error(t, err)
}

func TestValidate_ExpiredToken(t *testing.T) {
	// TTL -1 = sudah expired sejak 1 menit lalu
	signed, _, err := token.Generate("user-123", "user", testSecret, -1)
	require.NoError(t, err)

	_, err = token.Validate(signed, testSecret)
	assert.Error(t, err)
}
