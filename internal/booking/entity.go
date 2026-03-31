package booking

import (
	"time"

	"github.com/google/uuid"
)

// Booking represents a completed booking
type Booking struct {
	ID            uuid.UUID `json:"id"`
	ShowID        uuid.UUID `json:"show_id"`
	CustomerEmail string    `json:"customer_email"`
	CustomerPhone string    `json:"customer_phone"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}

// Ticket represents a booked ticket
type Ticket struct {
	ID         uuid.UUID `json:"id"`
	BookingID  uuid.UUID `json:"booking_id"`
	ShowID     uuid.UUID `json:"show_id"`
	SeatID     uuid.UUID `json:"seat_id"`
	QRCodeHash string    `json:"qr_code_hash"`
	CreatedAt  time.Time `json:"created_at"`
}

// BookingStatus constants
const (
	StatusCompleted = "completed"
	StatusCancelled = "cancelled"
)

// CreateBookingRequest represents the request to create a booking
type CreateBookingRequest struct {
	SessionID     string `json:"session_id"`
	ShowID        string `json:"show_id"`
	CustomerEmail string `json:"customer_email"`
	CustomerPhone string `json:"customer_phone"`
}

// BookingResponse represents the booking response
type BookingResponse struct {
	ID            string           `json:"id"`
	ShowID        string           `json:"show_id"`
	CustomerEmail string           `json:"customer_email"`
	CustomerPhone string           `json:"customer_phone"`
	Status        string           `json:"status"`
	Tickets       []TicketResponse `json:"tickets"`
	CreatedAt     time.Time        `json:"created_at"`
}

// TicketResponse represents a ticket in the response
type TicketResponse struct {
	ID       string `json:"id"`
	SeatID   string `json:"seat_id"`
	SeatInfo string `json:"seat_info,omitempty"`
}

// BookingSummary represents a summary of a booking
type BookingSummary struct {
	ID            string    `json:"id"`
	ShowID        string    `json:"show_id"`
	CustomerEmail string    `json:"customer_email"`
	SeatCount     int       `json:"seat_count"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}
