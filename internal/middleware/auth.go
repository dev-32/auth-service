package middleware

import (
	"strings"

	"github.com/dev-32/auth-service/internal/domain"
	jwtpkg "github.com/dev-32/auth-service/pkg/jwt"
	"github.com/gin-gonic/gin"
	"net/http"
)

type AuthMiddleware struct {
	jwt        *jwtpkg.Manager
	tokenCache domain.TokenCache
}

func NewAuthMiddleware(jwt *jwtpkg.Manager, tokenCache domain.TokenCache) *AuthMiddleware {
	return &AuthMiddleware{jwt: jwt, tokenCache: tokenCache}
}

// Authenticate validates JWT and sets user info in context
func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearerToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			c.Abort()
			return
		}

		// Check blacklist
		blacklisted, err := m.tokenCache.IsBlacklisted(c.Request.Context(), token)
		if err != nil || blacklisted {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token has been revoked"})
			c.Abort()
			return
		}

		// Parse token
		claims, err := m.jwt.ParseAccessToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		// Set claims in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)

		c.Next()
	}
}

func extractBearerToken(c *gin.Context) string {
	auth := c.GetHeader("Authorization")
	if auth == "" {
		return ""
	}
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}
	return parts[1]
}
