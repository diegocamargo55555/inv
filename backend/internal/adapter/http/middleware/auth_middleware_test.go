package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/invest/backend/internal/adapter/http/middleware"
	"github.com/invest/backend/pkg/token"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	maker := token.NewJWTMaker("supersecretkeywith32characterslong")

	setupServer := func() *gin.Engine {
		r := gin.New()
		r.GET("/protected", middleware.AuthMiddleware(maker), func(c *gin.Context) {
			payload, exists := c.Get(middleware.AuthorizationPayloadKey)
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "no payload"})
				return
			}
			p := payload.(*token.Payload)
			c.JSON(http.StatusOK, gin.H{"user_id": p.UserID, "email": p.Email})
		})
		return r
	}

	t.Run("valid authorization token succeeds with 200", func(t *testing.T) {
		r := setupServer()
		userID := uuid.New()
		tokenStr, _, _ := maker.CreateToken(userID, "user@test.com", token.TokenTypeAccess, time.Hour)

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+tokenStr)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "user@test.com")
	})

	t.Run("missing authorization header returns 401", func(t *testing.T) {
		r := setupServer()
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("unsupported auth format returns 401", func(t *testing.T) {
		r := setupServer()
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Basic invalidtoken")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("expired token returns 401", func(t *testing.T) {
		r := setupServer()
		userID := uuid.New()
		tokenStr, _, _ := maker.CreateToken(userID, "user@test.com", token.TokenTypeAccess, -time.Minute)

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+tokenStr)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
