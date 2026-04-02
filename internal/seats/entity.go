package seats

import (
	"time"

	"github.com/google/uuid"
)

// Seat represents a seat entity
type Seat struct {
	ID         uuid.UUID `json:"id"`
	HallID     uuid.UUID `json:"hall_id"`
	RowName    string    `json:"row_name"`
	SeatNumber int       `json:"seat_number"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Hall represents a hall/cinema room
type Hall struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	TotalSeats int       `json:"total_seats"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// CreateHallRequest represents the request to create a hall
type CreateHallRequest struct {
	Name       string          `json:"name"`
	SeatLayout []SeatLayoutRow `json:"seat_layout"`
}

type UpdateHallRequest struct {
	Name       string          `json:"name"`
	SeatLayout []SeatLayoutRow `json:"seat_layout"`
}

// SeatLayoutRow represents a row of seats
type SeatLayoutRow struct {
	RowName   string `json:"row_name"`
	SeatCount int    `json:"seat_count"`
}

// CreateSeatsRequest represents the request to create seats
type CreateSeatsRequest struct {
	HallID     string          `json:"hall_id"`
	SeatLayout []SeatLayoutRow `json:"seat_layout"`
}

// HallResponse represents the hall response
type HallResponse struct {
	ID         string         `json:"id"`
	Name       string         `json:"name"`
	TotalSeats int            `json:"total_seats"`
	Seats      []SeatResponse `json:"seats,omitempty"`
}

// SeatResponse represents the seat response
type SeatResponse struct {
	ID         string `json:"id"`
	RowName    string `json:"row_name"`
	SeatNumber int    `json:"seat_number"`
}

// ToResponse converts Hall to HallResponse
func (h *Hall) ToResponse() *HallResponse {
	return &HallResponse{
		ID:         h.ID.String(),
		Name:       h.Name,
		TotalSeats: h.TotalSeats,
	}
}

// ToResponse converts Seat to SeatResponse
func (s *Seat) ToResponse() *SeatResponse {
	return &SeatResponse{
		ID:         s.ID.String(),
		RowName:    s.RowName,
		SeatNumber: s.SeatNumber,
	}
}
