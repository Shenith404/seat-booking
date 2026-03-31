package shows

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines the interface for show data access
type Repository interface {
	Create(ctx context.Context, show *Show) error
	GetByID(ctx context.Context, id uuid.UUID) (*Show, error)
	GetAll(ctx context.Context, limit, offset int) ([]Show, int64, error)
	GetByDate(ctx context.Context, date time.Time, limit, offset int) ([]Show, int64, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetShowWithDetails(ctx context.Context, id uuid.UUID) (*Show, *MovieInfo, *HallInfo, error)
	GetShowSeats(ctx context.Context, showID uuid.UUID) ([]ShowSeatStatus, error)
	GetBookedSeatIDs(ctx context.Context, showID uuid.UUID) ([]uuid.UUID, error)
}

// PostgresRepository implements Repository using PostgreSQL
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

// Create creates a new show
func (r *PostgresRepository) Create(ctx context.Context, show *Show) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO shows (id, movie_id, hall_id, start_time, end_time, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, NOW(), NOW())`,
		pgtype.UUID{Bytes: show.ID, Valid: true},
		pgtype.UUID{Bytes: show.MovieID, Valid: true},
		pgtype.UUID{Bytes: show.HallID, Valid: true},
		show.StartTime,
		show.EndTime,
	)
	return err
}

// GetByID retrieves a show by ID
func (r *PostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*Show, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, movie_id, hall_id, start_time, end_time, created_at, updated_at
		 FROM shows WHERE id = $1`,
		pgtype.UUID{Bytes: id, Valid: true},
	)

	var show Show
	var pgID, pgMovieID, pgHallID pgtype.UUID
	var pgCreatedAt, pgUpdatedAt pgtype.Timestamptz

	err := row.Scan(&pgID, &pgMovieID, &pgHallID, &show.StartTime, &show.EndTime, &pgCreatedAt, &pgUpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get show: %w", err)
	}

	show.ID = pgID.Bytes
	show.MovieID = pgMovieID.Bytes
	show.HallID = pgHallID.Bytes
	show.CreatedAt = pgCreatedAt.Time
	show.UpdatedAt = pgUpdatedAt.Time

	return &show, nil
}

// GetAll retrieves all shows with pagination
func (r *PostgresRepository) GetAll(ctx context.Context, limit, offset int) ([]Show, int64, error) {
	var total int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM shows`).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count shows: %w", err)
	}

	rows, err := r.pool.Query(ctx,
		`SELECT id, movie_id, hall_id, start_time, end_time, created_at, updated_at
		 FROM shows ORDER BY start_time DESC LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get shows: %w", err)
	}
	defer rows.Close()

	return r.scanShows(rows, total)
}

// GetByDate retrieves shows for a specific date
func (r *PostgresRepository) GetByDate(ctx context.Context, date time.Time, limit, offset int) ([]Show, int64, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	var total int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM shows WHERE start_time >= $1 AND start_time < $2`,
		startOfDay, endOfDay,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count shows: %w", err)
	}

	rows, err := r.pool.Query(ctx,
		`SELECT id, movie_id, hall_id, start_time, end_time, created_at, updated_at
		 FROM shows WHERE start_time >= $1 AND start_time < $2
		 ORDER BY start_time LIMIT $3 OFFSET $4`,
		startOfDay, endOfDay, limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get shows: %w", err)
	}
	defer rows.Close()

	return r.scanShows(rows, total)
}

// Delete deletes a show
func (r *PostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM shows WHERE id = $1`,
		pgtype.UUID{Bytes: id, Valid: true},
	)
	return err
}

// GetShowWithDetails retrieves a show with movie and hall details
func (r *PostgresRepository) GetShowWithDetails(ctx context.Context, id uuid.UUID) (*Show, *MovieInfo, *HallInfo, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT s.id, s.movie_id, s.hall_id, s.start_time, s.end_time, s.created_at, s.updated_at,
		        m.id, m.title, m.duration_minutes,
		        h.id, h.name, h.total_seats
		 FROM shows s
		 JOIN movies m ON s.movie_id = m.id
		 JOIN halls h ON s.hall_id = h.id
		 WHERE s.id = $1`,
		pgtype.UUID{Bytes: id, Valid: true},
	)

	var show Show
	var movie MovieInfo
	var hall HallInfo
	var pgShowID, pgMovieID, pgHallID pgtype.UUID
	var pgMID, pgHID pgtype.UUID
	var pgCreatedAt, pgUpdatedAt pgtype.Timestamptz

	err := row.Scan(
		&pgShowID, &pgMovieID, &pgHallID, &show.StartTime, &show.EndTime, &pgCreatedAt, &pgUpdatedAt,
		&pgMID, &movie.Title, &movie.DurationMinutes,
		&pgHID, &hall.Name, &hall.TotalSeats,
	)
	if err == pgx.ErrNoRows {
		return nil, nil, nil, nil
	}
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get show details: %w", err)
	}

	show.ID = pgShowID.Bytes
	show.MovieID = pgMovieID.Bytes
	show.HallID = pgHallID.Bytes
	show.CreatedAt = pgCreatedAt.Time
	show.UpdatedAt = pgUpdatedAt.Time
	movie.ID = uuid.UUID(pgMID.Bytes).String()
	hall.ID = uuid.UUID(pgHID.Bytes).String()

	return &show, &movie, &hall, nil
}

// GetShowSeats retrieves all seats for a show with their status
func (r *PostgresRepository) GetShowSeats(ctx context.Context, showID uuid.UUID) ([]ShowSeatStatus, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT s.id, s.row_name, s.seat_number,
		        CASE WHEN t.id IS NOT NULL THEN 'booked' ELSE 'available' END as status
		 FROM seats s
		 JOIN shows sh ON s.hall_id = sh.hall_id
		 LEFT JOIN tickets t ON s.id = t.seat_id AND t.show_id = sh.id
		 WHERE sh.id = $1
		 ORDER BY s.row_name, s.seat_number`,
		pgtype.UUID{Bytes: showID, Valid: true},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get show seats: %w", err)
	}
	defer rows.Close()

	var seats []ShowSeatStatus
	for rows.Next() {
		var seat ShowSeatStatus
		var pgID pgtype.UUID

		if err := rows.Scan(&pgID, &seat.RowName, &seat.SeatNumber, &seat.Status); err != nil {
			return nil, fmt.Errorf("failed to scan seat: %w", err)
		}

		seat.ID = uuid.UUID(pgID.Bytes).String()
		seats = append(seats, seat)
	}

	return seats, rows.Err()
}

// GetBookedSeatIDs retrieves IDs of booked seats for a show
func (r *PostgresRepository) GetBookedSeatIDs(ctx context.Context, showID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT seat_id FROM tickets WHERE show_id = $1`,
		pgtype.UUID{Bytes: showID, Valid: true},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get booked seats: %w", err)
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var pgID pgtype.UUID
		if err := rows.Scan(&pgID); err != nil {
			return nil, fmt.Errorf("failed to scan seat id: %w", err)
		}
		ids = append(ids, pgID.Bytes)
	}

	return ids, rows.Err()
}

func (r *PostgresRepository) scanShows(rows pgx.Rows, total int64) ([]Show, int64, error) {
	var shows []Show
	for rows.Next() {
		var show Show
		var pgID, pgMovieID, pgHallID pgtype.UUID
		var pgCreatedAt, pgUpdatedAt pgtype.Timestamptz

		if err := rows.Scan(&pgID, &pgMovieID, &pgHallID, &show.StartTime, &show.EndTime, &pgCreatedAt, &pgUpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan show: %w", err)
		}

		show.ID = pgID.Bytes
		show.MovieID = pgMovieID.Bytes
		show.HallID = pgHallID.Bytes
		show.CreatedAt = pgCreatedAt.Time
		show.UpdatedAt = pgUpdatedAt.Time
		shows = append(shows, show)
	}

	return shows, total, rows.Err()
}
