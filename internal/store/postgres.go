package store

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shenith404/seat-booking/internal/config"
)

// PostgresStore wraps the pgxpool for database operations
type PostgresStore struct {
	Pool *pgxpool.Pool
}

// NewPostgresStore creates a new PostgreSQL connection pool
func NewPostgresStore(ctx context.Context, cfg config.PostgresConfig) (*PostgresStore, error) {
	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
		cfg.SSLMode,
	)

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse postgres config: %w", err)
	}

	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	log.Println("Connected to PostgreSQL successfully")

	return &PostgresStore{Pool: pool}, nil
}

// Close closes the database connection pool
func (s *PostgresStore) Close() {
	if s.Pool != nil {
		s.Pool.Close()
		log.Println("PostgreSQL connection pool closed")
	}
}

// Health checks if the database connection is healthy
func (s *PostgresStore) Health(ctx context.Context) error {
	return s.Pool.Ping(ctx)
}
