package hold

import (
	"time"

	"github.com/google/uuid"
)

// Session represents a user's hold session stored in Redis
type Session struct {
	SessionID   string    `json:"session_id"`
	ShowID      string    `json:"show_id"`
	HeldSeats   []string  `json:"held_seats"`
	ToggleCount int       `json:"toggle_count"`
	CreatedAt   time.Time `json:"created_at"`
	LastActive  time.Time `json:"last_active"`
}

// SeatHold represents a single seat hold
type SeatHold struct {
	SeatID    string `json:"seat_id"`
	SessionID string `json:"session_id"`
	ShowID    string `json:"show_id"`
}

// Constants for hold management
const (
	MaxToggleCount = 15               // Maximum seat toggles per session
	IdleTTL        = 2 * time.Minute  // Sliding window TTL
	MaxSessionTime = 10 * time.Minute // Absolute session maximum
)

// IsSessionExpired checks if session has exceeded max time
func (s *Session) IsSessionExpired() bool {
	return time.Since(s.CreatedAt) > MaxSessionTime
}

// CanToggle checks if session can still toggle seats
func (s *Session) CanToggle() bool {
	return s.ToggleCount < MaxToggleCount
}

// HoldSeatRequest represents a request to hold/release a seat
type HoldSeatRequest struct {
	SessionID string `json:"session_id"`
	ShowID    string `json:"show_id"`
	SeatID    string `json:"seat_id"`
}

// Validate validates the request
func (r *HoldSeatRequest) Validate() error {
	if _, err := uuid.Parse(r.SessionID); err != nil {
		return err
	}
	if _, err := uuid.Parse(r.ShowID); err != nil {
		return err
	}
	if _, err := uuid.Parse(r.SeatID); err != nil {
		return err
	}
	return nil
}

// ExtendSessionRequest represents a request to extend session
type ExtendSessionRequest struct {
	SessionID string `json:"session_id"`
	ShowID    string `json:"show_id"`
}

// HoldStatusResponse represents the current hold status
type HoldStatusResponse struct {
	SessionID     string   `json:"session_id"`
	ShowID        string   `json:"show_id"`
	HeldSeats     []string `json:"held_seats"`
	ToggleCount   int      `json:"toggle_count"`
	RemainingTime int      `json:"remaining_time_seconds"`
	MaxTime       int      `json:"max_time_seconds"`
	CanExtend     bool     `json:"can_extend"`
}

// SeatStatusResponse represents seat availability status
type SeatStatusResponse struct {
	SeatID    string `json:"seat_id"`
	Status    string `json:"status"`               // "available", "held", "booked"
	SessionID string `json:"session_id,omitempty"` // Only if held by current session
}
