package redis

import (
	"context"
	"log/slog"
	"time"

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

// BlacklistToken stores the token in Redis with an expiry
func (c *TokenCache) BlacklistToken(ctx context.Context, token string, expiry time.Duration) error {
	key := "blacklist:" + token
	return c.client.Set(ctx, key, "1", expiry).Err()
}

// IsBlacklisted checks if a token exists in the blacklist
func (c *TokenCache) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	key := "blacklist:" + token
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil // key doesn't exist — not blacklisted
	}
	if err != nil {
		return false, err // real redis error
	}
	return val == "1", nil
}
