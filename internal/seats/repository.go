package seats

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines the interface for seat/hall data access
type Repository interface {
	// Hall operations
	CreateHall(ctx context.Context, hall *Hall) error
	CreateHallWithSeats(ctx context.Context, hall *Hall, seats []Seat) error
	GetHallByID(ctx context.Context, id uuid.UUID) (*Hall, error)
	GetAllHalls(ctx context.Context) ([]Hall, error)
	DeleteHall(ctx context.Context, id uuid.UUID) error

	// Seat operations
	UpdateHallWithSeats(ctx context.Context, hallID uuid.UUID, hallName string, seats []Seat) error
	CreateSeats(ctx context.Context, seats []Seat) error
	GetSeatsByHallID(ctx context.Context, hallID uuid.UUID) ([]Seat, error)
	GetSeatByID(ctx context.Context, id uuid.UUID) (*Seat, error)
}

// PostgresRepository implements Repository using PostgreSQL
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

// CreateHall creates a new hall
func (r *PostgresRepository) CreateHall(ctx context.Context, hall *Hall) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO halls (id, name, total_seats, created_at, updated_at)
		 VALUES ($1, $2, $3, NOW(), NOW())`,
		pgtype.UUID{Bytes: hall.ID, Valid: true},
		hall.Name,
		hall.TotalSeats,
	)
	return err
}

// CreateHallWithSeats creates a new hall with seats in a single transaction
func (r *PostgresRepository) CreateHallWithSeats(ctx context.Context, hall *Hall, seats []Seat) error {
	// Start transaction
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Create hall
	_, err = tx.Exec(ctx,
		`INSERT INTO halls (id, name, total_seats, created_at, updated_at)
		 VALUES ($1, $2, $3, NOW(), NOW())`,
		pgtype.UUID{Bytes: hall.ID, Valid: true},
		hall.Name,
		hall.TotalSeats,
	)
	if err != nil {
		return fmt.Errorf("failed to create hall: %w", err)
	}

	// Create seats using batch
	if len(seats) > 0 {
		batch := &pgx.Batch{}
		for _, seat := range seats {
			batch.Queue(
				`INSERT INTO seats (id, hall_id, row_name, seat_number, created_at, updated_at)
				 VALUES ($1, $2, $3, $4, NOW(), NOW())`,
				pgtype.UUID{Bytes: seat.ID, Valid: true},
				pgtype.UUID{Bytes: seat.HallID, Valid: true},
				seat.RowName,
				seat.SeatNumber,
			)
		}

		br := tx.SendBatch(ctx, batch)
		for range seats {
			if _, err := br.Exec(); err != nil {
				br.Close()
				return fmt.Errorf("failed to insert seat: %w", err)
			}
		}
		br.Close()
	}

	// Commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetHallByID retrieves a hall by ID
func (r *PostgresRepository) GetHallByID(ctx context.Context, id uuid.UUID) (*Hall, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, name, total_seats, created_at, updated_at
		 FROM halls WHERE id = $1`,
		pgtype.UUID{Bytes: id, Valid: true},
	)

	var hall Hall
	var pgID pgtype.UUID
	var pgCreatedAt, pgUpdatedAt pgtype.Timestamptz

	err := row.Scan(&pgID, &hall.Name, &hall.TotalSeats, &pgCreatedAt, &pgUpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get hall: %w", err)
	}

	hall.ID = pgID.Bytes
	hall.CreatedAt = pgCreatedAt.Time
	hall.UpdatedAt = pgUpdatedAt.Time

	return &hall, nil
}

// GetAllHalls retrieves all halls
func (r *PostgresRepository) GetAllHalls(ctx context.Context) ([]Hall, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, total_seats, created_at, updated_at
		 FROM halls ORDER BY name`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get halls: %w", err)
	}
	defer rows.Close()

	var halls []Hall
	for rows.Next() {
		var hall Hall
		var pgID pgtype.UUID
		var pgCreatedAt, pgUpdatedAt pgtype.Timestamptz

		if err := rows.Scan(&pgID, &hall.Name, &hall.TotalSeats, &pgCreatedAt, &pgUpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan hall: %w", err)
		}

		hall.ID = pgID.Bytes
		hall.CreatedAt = pgCreatedAt.Time
		hall.UpdatedAt = pgUpdatedAt.Time
		halls = append(halls, hall)
	}

	return halls, rows.Err()
}

// DeleteHall deletes a hall
func (r *PostgresRepository) DeleteHall(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM halls WHERE id = $1`,
		pgtype.UUID{Bytes: id, Valid: true},
	)
	return err
}

// UpdateSeats updates seats for a hall
func (r *PostgresRepository) UpdateHallWithSeats(ctx context.Context, hallID uuid.UUID, hallName string, seats []Seat) error {
	// Start a transaction
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	//update hall name and total seats
	_, err = tx.Exec(ctx,
		`UPDATE halls SET name = $1, total_seats = $2, updated_at = NOW() WHERE id = $3`,
		hallName,
		len(seats),
		pgtype.UUID{Bytes: hallID, Valid: true},
	)

	// Delete existing seats for the hall
	_, err = tx.Exec(ctx,
		`DELETE FROM seats WHERE hall_id = $1`,
		pgtype.UUID{Bytes: hallID, Valid: true},
	)
	if err != nil {
		return fmt.Errorf("failed to delete existing seats: %w", err)
	}

	// Insert new seats using batch
	if len(seats) > 0 {
		batch := &pgx.Batch{}
		for _, seat := range seats {
			batch.Queue(
				`INSERT INTO seats (id, hall_id, row_name, seat_number, created_at, updated_at)
				 VALUES ($1, $2, $3, $4, NOW(), NOW())`,
				pgtype.UUID{Bytes: seat.ID, Valid: true},
				pgtype.UUID{Bytes: seat.HallID, Valid: true},
				seat.RowName,
				seat.SeatNumber,
			)
		}

		br := tx.SendBatch(ctx, batch)
		for range seats {
			if _, err := br.Exec(); err != nil {
				br.Close()
				return fmt.Errorf("failed to insert seat: %w", err)
			}
		}
		br.Close()
	}

	// Commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// CreateSeats creates multiple seats
func (r *PostgresRepository) CreateSeats(ctx context.Context, seats []Seat) error {
	batch := &pgx.Batch{}
	for _, seat := range seats {
		batch.Queue(
			`INSERT INTO seats (id, hall_id, row_name, seat_number, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, NOW(), NOW())`,
			pgtype.UUID{Bytes: seat.ID, Valid: true},
			pgtype.UUID{Bytes: seat.HallID, Valid: true},
			seat.RowName,
			seat.SeatNumber,
		)
	}

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	for range seats {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("failed to create seat: %w", err)
		}
	}

	return nil
}

// GetSeatsByHallID retrieves all seats for a hall
func (r *PostgresRepository) GetSeatsByHallID(ctx context.Context, hallID uuid.UUID) ([]Seat, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, hall_id, row_name, seat_number, created_at, updated_at
		 FROM seats WHERE hall_id = $1 ORDER BY row_name, seat_number`,
		pgtype.UUID{Bytes: hallID, Valid: true},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get seats: %w", err)
	}
	defer rows.Close()

	var seats []Seat
	for rows.Next() {
		var seat Seat
		var pgID, pgHallID pgtype.UUID
		var pgCreatedAt, pgUpdatedAt pgtype.Timestamptz

		if err := rows.Scan(&pgID, &pgHallID, &seat.RowName, &seat.SeatNumber, &pgCreatedAt, &pgUpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan seat: %w", err)
		}

		seat.ID = pgID.Bytes
		seat.HallID = pgHallID.Bytes
		seat.CreatedAt = pgCreatedAt.Time
		seat.UpdatedAt = pgUpdatedAt.Time
		seats = append(seats, seat)
	}

	return seats, rows.Err()
}

// GetSeatByID retrieves a seat by ID
func (r *PostgresRepository) GetSeatByID(ctx context.Context, id uuid.UUID) (*Seat, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, hall_id, row_name, seat_number, created_at, updated_at
		 FROM seats WHERE id = $1`,
		pgtype.UUID{Bytes: id, Valid: true},
	)

	var seat Seat
	var pgID, pgHallID pgtype.UUID
	var pgCreatedAt, pgUpdatedAt pgtype.Timestamptz

	err := row.Scan(&pgID, &pgHallID, &seat.RowName, &seat.SeatNumber, &pgCreatedAt, &pgUpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get seat: %w", err)
	}

	seat.ID = pgID.Bytes
	seat.HallID = pgHallID.Bytes
	seat.CreatedAt = pgCreatedAt.Time
	seat.UpdatedAt = pgUpdatedAt.Time

	return &seat, nil
}
