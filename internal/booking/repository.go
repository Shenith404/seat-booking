package booking

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shenith404/seat-booking/internal/db"
)

// Repository defines the interface for booking data access
type Repository interface {
	// CreateBookingWithTickets creates a booking and tickets in a single transaction
	// Uses SELECT FOR UPDATE to lock seats and prevent double booking
	CreateBookingWithTickets(ctx context.Context, booking CreateBookingParams, seatIDs []uuid.UUID) (*Booking, []Ticket, error)

	// GetBooking retrieves a booking by ID
	GetBooking(ctx context.Context, id uuid.UUID) (*Booking, error)

	// GetBookingTickets retrieves all tickets for a booking
	GetBookingTickets(ctx context.Context, bookingID uuid.UUID) ([]Ticket, error)

	// CheckSeatsAvailable checks if seats are already booked (without locking)
	CheckSeatsAvailable(ctx context.Context, showID uuid.UUID, seatIDs []uuid.UUID) (bool, error)

	// UpdateTicketQRHash updates the QR hash for a ticket
	UpdateTicketQRHash(ctx context.Context, ticketID uuid.UUID, qrHash string) error
}

// CreateBookingParams holds parameters for creating a booking
type CreateBookingParams struct {
	ShowID        uuid.UUID
	CustomerEmail string
	CustomerPhone string
}

// PostgresRepository implements Repository using PostgreSQL
type PostgresRepository struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{
		pool:    pool,
		queries: db.New(pool),
	}
}

// CreateBookingWithTickets creates a booking with tickets using ACID transaction
func (r *PostgresRepository) CreateBookingWithTickets(
	ctx context.Context,
	params CreateBookingParams,
	seatIDs []uuid.UUID,
) (*Booking, []Ticket, error) {
	var booking *Booking
	var tickets []Ticket

	// Start transaction
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.Serializable, // Strictest isolation level
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	// Convert seatIDs to pgtype.UUID slice
	pgSeatIDs := make([]pgtype.UUID, len(seatIDs))
	for i, id := range seatIDs {
		pgSeatIDs[i] = pgtype.UUID{Bytes: id, Valid: true}
	}

	// SELECT FOR UPDATE to lock existing tickets for these seats
	// If any tickets exist, these seats are already booked
	existingTickets, err := qtx.CheckSeatsAvailableForUpdate(ctx, db.CheckSeatsAvailableForUpdateParams{
		ShowID:  pgtype.UUID{Bytes: params.ShowID, Valid: true},
		Column2: pgSeatIDs,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to check seat availability: %w", err)
	}

	if len(existingTickets) > 0 {
		return nil, nil, ErrSeatsAlreadyBooked
	}

	// Create booking
	dbBooking, err := qtx.CreateBooking(ctx, db.CreateBookingParams{
		ShowID:        pgtype.UUID{Bytes: params.ShowID, Valid: true},
		CustomerEmail: params.CustomerEmail,
		CustomerPhone: params.CustomerPhone,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create booking: %w", err)
	}

	booking = &Booking{
		ID:            dbBooking.ID.Bytes,
		ShowID:        dbBooking.ShowID.Bytes,
		CustomerEmail: dbBooking.CustomerEmail,
		CustomerPhone: dbBooking.CustomerPhone,
		Status:        dbBooking.Status,
		CreatedAt:     dbBooking.CreatedAt.Time,
	}

	// Create tickets for each seat
	tickets = make([]Ticket, 0, len(seatIDs))
	for _, seatID := range seatIDs {
		// Create ticket with placeholder QR hash (will be updated by worker)
		dbTicket, err := qtx.CreateTicket(ctx, db.CreateTicketParams{
			BookingID:  dbBooking.ID,
			ShowID:     pgtype.UUID{Bytes: params.ShowID, Valid: true},
			SeatID:     pgtype.UUID{Bytes: seatID, Valid: true},
			QrCodeHash: "pending", // Will be updated by background worker
		})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create ticket: %w", err)
		}

		tickets = append(tickets, Ticket{
			ID:         dbTicket.ID.Bytes,
			BookingID:  dbTicket.BookingID.Bytes,
			ShowID:     dbTicket.ShowID.Bytes,
			SeatID:     dbTicket.SeatID.Bytes,
			QRCodeHash: dbTicket.QrCodeHash,
			CreatedAt:  dbTicket.CreatedAt.Time,
		})
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return nil, nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return booking, tickets, nil
}

// GetBooking retrieves a booking by ID
func (r *PostgresRepository) GetBooking(ctx context.Context, id uuid.UUID) (*Booking, error) {
	// We need to add this query to queries.sql
	row := r.pool.QueryRow(ctx,
		`SELECT id, show_id, customer_email, customer_phone, status, created_at 
		 FROM booking WHERE id = $1`,
		pgtype.UUID{Bytes: id, Valid: true},
	)

	var booking Booking
	var pgID, pgShowID pgtype.UUID
	var pgCreatedAt pgtype.Timestamptz

	err := row.Scan(&pgID, &pgShowID, &booking.CustomerEmail, &booking.CustomerPhone, &booking.Status, &pgCreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}

	booking.ID = pgID.Bytes
	booking.ShowID = pgShowID.Bytes
	booking.CreatedAt = pgCreatedAt.Time

	return &booking, nil
}

// GetBookingTickets retrieves all tickets for a booking
func (r *PostgresRepository) GetBookingTickets(ctx context.Context, bookingID uuid.UUID) ([]Ticket, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, booking_id, show_id, seat_id, qr_code_hash, created_at 
		 FROM tickets WHERE booking_id = $1`,
		pgtype.UUID{Bytes: bookingID, Valid: true},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get tickets: %w", err)
	}
	defer rows.Close()

	var tickets []Ticket
	for rows.Next() {
		var t Ticket
		var pgID, pgBookingID, pgShowID, pgSeatID pgtype.UUID
		var pgCreatedAt pgtype.Timestamptz

		if err := rows.Scan(&pgID, &pgBookingID, &pgShowID, &pgSeatID, &t.QRCodeHash, &pgCreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan ticket: %w", err)
		}

		t.ID = pgID.Bytes
		t.BookingID = pgBookingID.Bytes
		t.ShowID = pgShowID.Bytes
		t.SeatID = pgSeatID.Bytes
		t.CreatedAt = pgCreatedAt.Time
		tickets = append(tickets, t)
	}

	return tickets, rows.Err()
}

// CheckSeatsAvailable checks if seats are already booked
func (r *PostgresRepository) CheckSeatsAvailable(ctx context.Context, showID uuid.UUID, seatIDs []uuid.UUID) (bool, error) {
	pgSeatIDs := make([]pgtype.UUID, len(seatIDs))
	for i, id := range seatIDs {
		pgSeatIDs[i] = pgtype.UUID{Bytes: id, Valid: true}
	}

	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM tickets WHERE show_id = $1 AND seat_id = ANY($2)`,
		pgtype.UUID{Bytes: showID, Valid: true},
		pgSeatIDs,
	).Scan(&count)

	if err != nil {
		return false, fmt.Errorf("failed to check seat availability: %w", err)
	}

	return count == 0, nil
}

// UpdateTicketQRHash updates the QR hash for a ticket
func (r *PostgresRepository) UpdateTicketQRHash(ctx context.Context, ticketID uuid.UUID, qrHash string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE tickets SET qr_code_hash = $1 WHERE id = $2`,
		qrHash,
		pgtype.UUID{Bytes: ticketID, Valid: true},
	)
	return err
}

// ErrSeatsAlreadyBooked is returned when seats are already booked
var ErrSeatsAlreadyBooked = fmt.Errorf("seats already booked")
