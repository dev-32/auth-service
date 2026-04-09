package main

import (
	"fmt"
	"log/slog"

	"github.com/dev-32/auth-service/internal/config"
	"github.com/dev-32/auth-service/internal/handler"
	"github.com/dev-32/auth-service/internal/logger"
	"github.com/dev-32/auth-service/internal/repository/postgres"
	redisrepo "github.com/dev-32/auth-service/internal/repository/redis"
	"github.com/dev-32/auth-service/internal/service"
	jwtpkg "github.com/dev-32/auth-service/pkg/jwt"
)

func main() {
	// Logger
	logger.New()

	// Config
	cfg := config.Load()

	// Validate required config
	if cfg.JWTAccessSecret == "" || cfg.JWTRefreshSecret == "" {
		slog.Error("JWT secrets are not set — refusing to start")
		return
	}

	// PostgreSQL
	db := postgres.Connect(cfg)
	defer db.Close()

	// Redis
	redisClient := redisrepo.NewClient(cfg)
	tokenCache := redisrepo.NewTokenCache(redisClient)

	// Repositories
	userRepo := postgres.NewUserRepo(db)
	tokenRepo := postgres.NewTokenRepo(db)

	// Run migrations
	if err := userRepo.Migrate(); err != nil {
		slog.Error("user migration failed", "error", err.Error())
		return
	}
	if err := tokenRepo.Migrate(); err != nil {
		slog.Error("token migration failed", "error", err.Error())
		return
	}
	slog.Info("migrations complete")

	// JWT manager
	jwtManager, err := jwtpkg.NewManager(
		cfg.JWTAccessSecret,
		cfg.JWTRefreshSecret,
		cfg.JWTAccessExpiry,
		cfg.JWTRefreshExpiry,
	)
	if err != nil {
		slog.Error("jwt manager failed", "error", err.Error())
		return
	}

	// // OAuth providers
	// googleProvider := oauth.NewGoogleProvider(
	// 	cfg.GoogleClientID,
	// 	cfg.GoogleClientSecret,
	// 	cfg.GoogleRedirectURL,
	// )
	// githubProvider := oauth.NewGitHubProvider(
	// 	cfg.GitHubClientID,
	// 	cfg.GitHubClientSecret,
	// 	cfg.GitHubRedirectURL,
	// )

	// Services
	authService := service.NewAuthService(userRepo, tokenRepo, tokenCache, jwtManager)
	// oauthService := service.NewOAuthService(userRepo, tokenRepo, tokenCache, jwtManager, googleProvider, githubProvider)
	adminService := service.NewAdminService(userRepo)

	// Router
	router := handler.NewRouter(
		authService,
		adminService,
		jwtManager,
		tokenCache,
		cfg.FrontendURL,
	)

	// Start server
	addr := fmt.Sprintf(":%s", cfg.AppPort)
	slog.Info("auth service starting",
		"port", cfg.AppPort,
		"env", cfg.AppEnv,
	)

	if err := router.Run(addr); err != nil {
		slog.Error("server failed", "error", err.Error())
	}
}
