package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type PostgresClient struct {
	pool   *pgxpool.Pool
	logger zerolog.Logger
}

func NewPostgresClient(url string, logger zerolog.Logger) (*PostgresClient, error) {
	config, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = 5 * time.Minute
	config.MaxConnIdleTime = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info().Msg("Connected to PostgreSQL")

	return &PostgresClient{
		pool:   pool,
		logger: logger,
	}, nil
}

func (p *PostgresClient) Pool() *pgxpool.Pool {
	return p.pool
}

func (p *PostgresClient) Close() {
	p.pool.Close()
}

