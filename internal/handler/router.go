package handler

import (
	"net/http"

	"github.com/dev-32/auth-service/internal/domain"
	"github.com/dev-32/auth-service/internal/middleware"
	"github.com/dev-32/auth-service/internal/service"
	jwtpkg "github.com/dev-32/auth-service/pkg/jwt"
	"github.com/gin-gonic/gin"
)

type Router struct {
	authHandler *AuthHandler
	userHandler *UserHandler

	adminHandler   *AdminHandler
	authMiddleware *middleware.AuthMiddleware
}

func NewRouter(
	authService domain.AuthService,
	// oauthService domain.OAuthService,
	adminService *service.AdminService,
	jwt *jwtpkg.Manager,
	tokenCache domain.TokenCache,
	frontendURL string,
) *gin.Engine {

	// Set Gin to release mode in production
	gin.SetMode(gin.DebugMode)

	r := gin.New()

	// Global middleware
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	// Init handlers
	authHandler := NewAuthHandler(authService)
	userHandler := NewUserHandler(authService)
	adminHandler := NewAdminHandler(adminService)

	// Init middleware
	authMW := middleware.NewAuthMiddleware(jwt, tokenCache)

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API v1
	v1 := r.Group("/api/v1")
	{
		// Public auth routes
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.Refresh)

			// // OAuth routes
			// auth.GET("/oauth/google", oauthHandler.GoogleLogin)
			// auth.GET("/oauth/google/callback", oauthHandler.GoogleCallback)
			// auth.GET("/oauth/github", oauthHandler.GitHubLogin)
			// auth.GET("/oauth/github/callback", oauthHandler.GitHubCallback)
		}

		// Protected routes — requires valid JWT
		protected := v1.Group("/")
		protected.Use(authMW.Authenticate())
		{
			protected.POST("/auth/logout", authHandler.Logout)
			protected.GET("/me", userHandler.GetMe)
			protected.PUT("/me", userHandler.UpdateMe)
		}

		// Admin routes — requires admin role
		admin := v1.Group("/admin")
		admin.Use(authMW.Authenticate())
		admin.Use(middleware.RequireRole("admin"))
		{
			admin.GET("/users", adminHandler.ListUsers)
			admin.DELETE("/users/:id", adminHandler.DeleteUser)
		}
	}

	return r
}
