package password_test

import (
	"testing"

	"github.com/RazorBold/golang_backend1/internal/pkg/password"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHash_ProducesNonEmptyHash(t *testing.T) {
	hash, err := password.Hash("secret123")

	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, "secret123", hash)
}

func TestHash_DifferentHashEachTime(t *testing.T) {
	hash1, _ := password.Hash("secret123")
	hash2, _ := password.Hash("secret123")

	// bcrypt menggunakan salt random → hash selalu berbeda
	assert.NotEqual(t, hash1, hash2)
}

func TestCompare_CorrectPassword(t *testing.T) {
	hash, err := password.Hash("secret123")
	require.NoError(t, err)

	assert.True(t, password.Compare(hash, "secret123"))
}

func TestCompare_WrongPassword(t *testing.T) {
	hash, err := password.Hash("secret123")
	require.NoError(t, err)

	assert.False(t, password.Compare(hash, "wrongpass"))
}

func TestCompare_EmptyPassword(t *testing.T) {
	hash, err := password.Hash("secret123")
	require.NoError(t, err)

	assert.False(t, password.Compare(hash, ""))
}
