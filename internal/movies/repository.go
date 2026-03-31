package movies

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines the interface for movie data access
type Repository interface {
	Create(ctx context.Context, movie *Movie) error
	GetByID(ctx context.Context, id uuid.UUID) (*Movie, error)
	GetAll(ctx context.Context, limit, offset int) ([]Movie, int64, error)
	Update(ctx context.Context, movie *Movie) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// PostgresRepository implements Repository using PostgreSQL
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

// Create creates a new movie
func (r *PostgresRepository) Create(ctx context.Context, movie *Movie) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO movies (id, title, duration_minutes, description, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, NOW(), NOW())`,
		pgtype.UUID{Bytes: movie.ID, Valid: true},
		movie.Title,
		movie.DurationMinutes,
		movie.Description,
	)
	return err
}

// GetByID retrieves a movie by ID
func (r *PostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*Movie, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, title, duration_minutes, description, created_at, updated_at
		 FROM movies WHERE id = $1`,
		pgtype.UUID{Bytes: id, Valid: true},
	)

	var movie Movie
	var pgID pgtype.UUID
	var pgDesc pgtype.Text
	var pgCreatedAt, pgUpdatedAt pgtype.Timestamptz

	err := row.Scan(&pgID, &movie.Title, &movie.DurationMinutes, &pgDesc, &pgCreatedAt, &pgUpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get movie: %w", err)
	}

	movie.ID = pgID.Bytes
	movie.Description = pgDesc.String
	movie.CreatedAt = pgCreatedAt.Time
	movie.UpdatedAt = pgUpdatedAt.Time

	return &movie, nil
}

// GetAll retrieves all movies with pagination
func (r *PostgresRepository) GetAll(ctx context.Context, limit, offset int) ([]Movie, int64, error) {
	// Get total count
	var total int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM movies`).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count movies: %w", err)
	}

	// Get movies
	rows, err := r.pool.Query(ctx,
		`SELECT id, title, duration_minutes, description, created_at, updated_at
		 FROM movies ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get movies: %w", err)
	}
	defer rows.Close()

	var movies []Movie
	for rows.Next() {
		var movie Movie
		var pgID pgtype.UUID
		var pgDesc pgtype.Text
		var pgCreatedAt, pgUpdatedAt pgtype.Timestamptz

		if err := rows.Scan(&pgID, &movie.Title, &movie.DurationMinutes, &pgDesc, &pgCreatedAt, &pgUpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan movie: %w", err)
		}

		movie.ID = pgID.Bytes
		movie.Description = pgDesc.String
		movie.CreatedAt = pgCreatedAt.Time
		movie.UpdatedAt = pgUpdatedAt.Time
		movies = append(movies, movie)
	}

	return movies, total, rows.Err()
}

// Update updates a movie
func (r *PostgresRepository) Update(ctx context.Context, movie *Movie) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE movies SET title = $1, duration_minutes = $2, description = $3, updated_at = NOW()
		 WHERE id = $4`,
		movie.Title,
		movie.DurationMinutes,
		movie.Description,
		pgtype.UUID{Bytes: movie.ID, Valid: true},
	)
	return err
}

// Delete deletes a movie
func (r *PostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM movies WHERE id = $1`,
		pgtype.UUID{Bytes: id, Valid: true},
	)
	return err
}
