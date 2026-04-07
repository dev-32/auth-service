package main

import (
	"log/slog"

	"github.com/dev-32/auth-service/internal/config"
	"github.com/dev-32/auth-service/internal/logger"
)

func main() {
	logger.New()

	cfg := config.Load()

	slog.Info("auth service starting", "port", cfg.AppPort, "env", cfg.AppEnv)

}
