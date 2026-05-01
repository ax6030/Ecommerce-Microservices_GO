package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	pb_user "github.com/jason/ecommerce/proto/user"
)

type AuthMiddleware struct {
	userClient pb_user.UserServiceClient
}

func NewAuthMiddleware(userClient pb_user.UserServiceClient) *AuthMiddleware {
	return &AuthMiddleware{userClient: userClient}
}

func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
			return
		}

		resp, err := m.userClient.ValidateToken(context.Background(), &pb_user.ValidateTokenRequest{Token: parts[1]})
		if err != nil || !resp.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set("user_id", resp.UserId)
		c.Next()
	}
}
