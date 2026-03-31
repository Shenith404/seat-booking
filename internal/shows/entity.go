package shows

import (
	"time"

	"github.com/google/uuid"
)

// Show represents a movie show/screening
type Show struct {
	ID        uuid.UUID `json:"id"`
	MovieID   uuid.UUID `json:"movie_id"`
	HallID    uuid.UUID `json:"hall_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateShowRequest represents the request to create a show
type CreateShowRequest struct {
	MovieID   string `json:"movie_id"`
	HallID    string `json:"hall_id"`
	StartTime string `json:"start_time"` // ISO 8601 format
}

// ShowResponse represents the show response
type ShowResponse struct {
	ID        string     `json:"id"`
	MovieID   string     `json:"movie_id"`
	HallID    string     `json:"hall_id"`
	StartTime time.Time  `json:"start_time"`
	EndTime   time.Time  `json:"end_time"`
	Movie     *MovieInfo `json:"movie,omitempty"`
	Hall      *HallInfo  `json:"hall,omitempty"`
}

// MovieInfo represents basic movie info in show response
type MovieInfo struct {
	ID              string `json:"id"`
	Title           string `json:"title"`
	DurationMinutes int    `json:"duration_minutes"`
}

// HallInfo represents basic hall info in show response
type HallInfo struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	TotalSeats int    `json:"total_seats"`
}

// ShowSeatsResponse represents seat availability for a show
type ShowSeatsResponse struct {
	ShowID string           `json:"show_id"`
	Seats  []ShowSeatStatus `json:"seats"`
}

// ShowSeatStatus represents a seat's status for a show
type ShowSeatStatus struct {
	ID         string `json:"id"`
	RowName    string `json:"row_name"`
	SeatNumber int    `json:"seat_number"`
	Status     string `json:"status"` // available, held, booked
}

// ToResponse converts Show to ShowResponse
func (s *Show) ToResponse() *ShowResponse {
	return &ShowResponse{
		ID:        s.ID.String(),
		MovieID:   s.MovieID.String(),
		HallID:    s.HallID.String(),
		StartTime: s.StartTime,
		EndTime:   s.EndTime,
	}
}
