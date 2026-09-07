package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/invest/backend/pkg/token"
)

const (
	AuthorizationHeaderKey  = "authorization"
	AuthorizationTypeBearer = "bearer"
	AuthorizationPayloadKey = "authorization_payload"
)

// AuthMiddleware creates a gin middleware for JWT authentication
func AuthMiddleware(tokenMaker *token.JWTMaker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorizationHeader := ctx.GetHeader(AuthorizationHeaderKey)
		if len(authorizationHeader) == 0 {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "cabeçalho de autorização não fornecido"})
			return
		}

		fields := strings.Fields(authorizationHeader)
		if len(fields) < 2 {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "formato de token inválido"})
			return
		}

		authorizationType := strings.ToLower(fields[0])
		if authorizationType != AuthorizationTypeBearer {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "tipo de autorização não suportado"})
			return
		}

		accessToken := fields[1]
		payload, err := tokenMaker.VerifyToken(accessToken)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token expirado ou inválido"})
			return
		}

		ctx.Set(AuthorizationPayloadKey, payload)
		ctx.Next()
	}
}
