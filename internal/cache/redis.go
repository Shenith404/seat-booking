package cache

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
	"github.com/shenith404/seat-booking/internal/config"
)

// RedisCache wraps the Redis client
type RedisCache struct {
	Client *redis.Client
}

// NewRedisCache creates a new Redis client connection
func NewRedisCache(ctx context.Context, cfg config.RedisConfig) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Println("Connected to Redis successfully")

	return &RedisCache{Client: client}, nil
}

// Close closes the Redis connection
func (c *RedisCache) Close() error {
	if c.Client != nil {
		log.Println("Redis connection closed")
		return c.Client.Close()
	}
	return nil
}

// Health checks if the Redis connection is healthy
func (c *RedisCache) Health(ctx context.Context) error {
	return c.Client.Ping(ctx).Err()
}
