package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type RedisClient struct {
	client *redis.Client
	logger zerolog.Logger
}

func NewRedisClient(url string, logger zerolog.Logger) (*RedisClient, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		// Try parsing as simple host:port
		opts = &redis.Options{
			Addr: url,
		}
	}
	applyDefaultRedisTimeouts(opts)

	client := redis.NewClient(opts)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info().Str("url", url).Msg("Connected to Redis")

	return &RedisClient{
		client: client,
		logger: logger,
	}, nil
}

func (r *RedisClient) Client() *redis.Client {
	return r.client
}

func (r *RedisClient) Close() error {
	return r.client.Close()
}

// applyDefaultRedisTimeouts sets sane network bounds when the URL did not specify them.
// Without Read/Write timeouts a half-open TCP to Redis can stall tunnel handlers indefinitely.
func applyDefaultRedisTimeouts(opts *redis.Options) {
	if opts == nil {
		return
	}
	if opts.DialTimeout <= 0 {
		opts.DialTimeout = 5 * time.Second
	}
	if opts.ReadTimeout <= 0 {
		opts.ReadTimeout = 5 * time.Second
	}
	if opts.WriteTimeout <= 0 {
		opts.WriteTimeout = 5 * time.Second
	}
	if opts.PoolTimeout <= 0 {
		opts.PoolTimeout = 4 * time.Second
	}
}
