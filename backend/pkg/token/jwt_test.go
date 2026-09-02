package token_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/invest/backend/pkg/token"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTMaker_GenerateAndVerifyToken(t *testing.T) {
	secretKey := "supersecretjwtkeywith32characterslong"
	maker := token.NewJWTMaker(secretKey)

	userID := uuid.New()
	email := "investor@example.com"
	duration := 15 * time.Minute

	// Table-driven tests for JWT lifecycle
	t.Run("generate and verify valid access token", func(t *testing.T) {
		tokenString, payload, err := maker.CreateToken(userID, email, token.TokenTypeAccess, duration)
		require.NoError(t, err)
		require.NotEmpty(t, tokenString)
		require.NotNil(t, payload)

		verifiedPayload, err := maker.VerifyToken(tokenString)
		require.NoError(t, err)
		assert.Equal(t, userID, verifiedPayload.UserID)
		assert.Equal(t, email, verifiedPayload.Email)
		assert.Equal(t, token.TokenTypeAccess, verifiedPayload.TokenType)
		assert.WithinDuration(t, time.Now().Add(duration), verifiedPayload.ExpiresAt.Time, 5*time.Second)
	})

	t.Run("generate and verify refresh token", func(t *testing.T) {
		tokenString, _, err := maker.CreateToken(userID, email, token.TokenTypeRefresh, 7*24*time.Hour)
		require.NoError(t, err)
		require.NotEmpty(t, tokenString)

		verifiedPayload, err := maker.VerifyToken(tokenString)
		require.NoError(t, err)
		assert.Equal(t, userID, verifiedPayload.UserID)
		assert.Equal(t, token.TokenTypeRefresh, verifiedPayload.TokenType)
	})

	t.Run("expired token returns error", func(t *testing.T) {
		tokenString, _, err := maker.CreateToken(userID, email, token.TokenTypeAccess, -time.Minute)
		require.NoError(t, err)

		verifiedPayload, err := maker.VerifyToken(tokenString)
		assert.Error(t, err)
		assert.Nil(t, verifiedPayload)
		assert.Equal(t, token.ErrExpiredToken, err)
	})

	t.Run("invalid token signature returns error", func(t *testing.T) {
		tokenString, _, err := maker.CreateToken(userID, email, token.TokenTypeAccess, duration)
		require.NoError(t, err)

		differentMaker := token.NewJWTMaker("anothersupersecretkeywithdifferentcharacters32")
		verifiedPayload, err := differentMaker.VerifyToken(tokenString)
		assert.Error(t, err)
		assert.Nil(t, verifiedPayload)
		assert.Equal(t, token.ErrInvalidToken, err)
	})
}
