package redis

import (
	"context"
	"log/slog"

	"github.com/dev-32/auth-service/internal/config"
	"github.com/redis/go-redis/v9"
)

type TokenCache struct {
	client *redis.Client
}

func NewClient(cfg *config.Config) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		slog.Error("failed to connect to redis", "error", err.Error())
		panic(err)
	}

	slog.Info("connected to redis", "addr", cfg.RedisAddr)
	return client
}

func NewTokenCache(client *redis.Client) *TokenCache {
	return &TokenCache{client: client}
}
