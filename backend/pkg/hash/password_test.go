package hash_test

import (
	"testing"

	"github.com/invest/backend/pkg/hash"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPasswordHasher(t *testing.T) {
	password := "Secret123@Pass"

	t.Run("hash and check valid password", func(t *testing.T) {
		hashed, err := hash.HashPassword(password)
		require.NoError(t, err)
		require.NotEmpty(t, hashed)
		assert.NotEqual(t, password, hashed)

		err = hash.CheckPassword(password, hashed)
		assert.NoError(t, err)
	})

	t.Run("check invalid password returns error", func(t *testing.T) {
		hashed, err := hash.HashPassword(password)
		require.NoError(t, err)

		err = hash.CheckPassword("WrongPassword123", hashed)
		assert.Error(t, err)
	})

	t.Run("empty password returns error", func(t *testing.T) {
		hashed, err := hash.HashPassword("")
		assert.Error(t, err)
		assert.Empty(t, hashed)
	})
}
