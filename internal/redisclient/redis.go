// Package redisclient creates Redis clients for workers and queues.
package redisclient

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"

	"remna-user-panel/internal/config"
)

// Open returns a connected Redis client, or nil when REDIS_URL is empty.
func Open(ctx context.Context, settings config.Settings) (*redis.Client, error) {
	if settings.RedisURL == "" {
		return nil, nil
	}
	options, err := redis.ParseURL(settings.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	client := redis.NewClient(options)
	if err := client.Ping(ctx).Err(); err != nil {
		closeErr := client.Close()
		if closeErr != nil {
			return nil, fmt.Errorf("ping redis: %w; close redis: %w", err, closeErr)
		}
		return nil, fmt.Errorf("ping redis: %w", err)
	}
	return client, nil
}
