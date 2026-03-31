package movies

import (
	"time"

	"github.com/google/uuid"
)

// Movie represents a movie entity
type Movie struct {
	ID              uuid.UUID `json:"id"`
	Title           string    `json:"title"`
	DurationMinutes int       `json:"duration_minutes"`
	Description     string    `json:"description,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// CreateMovieRequest represents the request to create a movie
type CreateMovieRequest struct {
	Title           string `json:"title"`
	DurationMinutes int    `json:"duration_minutes"`
	Description     string `json:"description,omitempty"`
}

// UpdateMovieRequest represents the request to update a movie
type UpdateMovieRequest struct {
	Title           *string `json:"title,omitempty"`
	DurationMinutes *int    `json:"duration_minutes,omitempty"`
	Description     *string `json:"description,omitempty"`
}

// MovieResponse represents the movie response
type MovieResponse struct {
	ID              string `json:"id"`
	Title           string `json:"title"`
	DurationMinutes int    `json:"duration_minutes"`
	Description     string `json:"description,omitempty"`
}

// ToResponse converts Movie to MovieResponse
func (m *Movie) ToResponse() *MovieResponse {
	return &MovieResponse{
		ID:              m.ID.String(),
		Title:           m.Title,
		DurationMinutes: m.DurationMinutes,
		Description:     m.Description,
	}
}
