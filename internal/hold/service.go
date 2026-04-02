package hold

import (
	"context"
	"time"

	"github.com/shenith404/seat-booking/internal/common"
	"github.com/shenith404/seat-booking/internal/pubsub"
)

// Service defines the interface for hold business logic
type Service interface {
	HoldSeat(ctx context.Context, req *HoldSeatRequest) (*HoldStatusResponse, *common.AppError)
	ReleaseSeat(ctx context.Context, req *HoldSeatRequest) (*HoldStatusResponse, *common.AppError)
	GetSessionStatus(ctx context.Context, showID, sessionID string) (*HoldStatusResponse, *common.AppError)
	ExtendSession(ctx context.Context, req *ExtendSessionRequest) (*HoldStatusResponse, *common.AppError)
	GetShowSeatsStatus(ctx context.Context, showID, sessionID string) ([]SeatStatusResponse, *common.AppError)
	GetSessionSeats(ctx context.Context, showID, sessionID string) ([]string, *common.AppError)
	ReleaseSession(ctx context.Context, showID, sessionID string) *common.AppError
}

// ServiceImpl implements the Service interface
type ServiceImpl struct {
	repo   Repository
	pubsub *pubsub.PubSub
	config ServiceConfig
}

// ServiceConfig holds service configuration
type ServiceConfig struct {
	IdleTTL        time.Duration
	MaxSessionTime time.Duration
	MaxToggleCount int
}

// DefaultServiceConfig returns default configuration
func DefaultServiceConfig() ServiceConfig {
	return ServiceConfig{
		IdleTTL:        IdleTTL,
		MaxSessionTime: MaxSessionTime,
		MaxToggleCount: MaxToggleCount,
	}
}

// NewService creates a new hold service
func NewService(repo Repository, ps *pubsub.PubSub, cfg ServiceConfig) *ServiceImpl {
	return &ServiceImpl{
		repo:   repo,
		pubsub: ps,
		config: cfg,
	}
}

// HoldSeat holds a seat for a session
func (s *ServiceImpl) HoldSeat(ctx context.Context, req *HoldSeatRequest) (*HoldStatusResponse, *common.AppError) {
	if err := req.Validate(); err != nil {
		return nil, common.NewBadRequestError("Invalid request: " + err.Error())
	}

	// Get or create session
	session, err := s.repo.GetSession(ctx, req.ShowID, req.SessionID)
	if err != nil {
		return nil, common.NewInternalError("Failed to get session").WithCause(err)
	}

	// Create new session if doesn't exist
	if session == nil {
		session = &Session{
			SessionID:   req.SessionID,
			ShowID:      req.ShowID,
			HeldSeats:   []string{},
			ToggleCount: 0,
			CreatedAt:   time.Now(),
			LastActive:  time.Now(),
		}
	}

	// Check if session has exceeded max time
	if session.IsSessionExpired() {
		return nil, common.NewHoldSessionMaxTimeError()
	}

	// Check toggle limit
	if !session.CanToggle() {
		return nil, common.NewHoldToggleLimitError()
	}

	// Check if seat is already held by this session
	for _, seatID := range session.HeldSeats {
		if seatID == req.SeatID {
			return s.buildStatusResponse(session), nil
		}
	}

	// Try to hold the seat
	success, err := s.repo.HoldSeat(ctx, req.ShowID, req.SeatID, req.SessionID, s.config.IdleTTL)
	if err != nil {
		return nil, common.NewInternalError("Failed to hold seat").WithCause(err)
	}

	if !success {
		return nil, common.NewHoldSeatUnavailableError(req.SeatID)
	}

	// Update session
	session.HeldSeats = append(session.HeldSeats, req.SeatID)
	session.ToggleCount++
	session.LastActive = time.Now()

	if err := s.repo.SaveSession(ctx, session, s.config.IdleTTL); err != nil {
		// Rollback seat hold
		s.repo.ReleaseSeat(ctx, req.ShowID, req.SeatID, req.SessionID)
		return nil, common.NewInternalError("Failed to save session").WithCause(err)
	}

	// Extend session and all held seats (including previously held ones)
	if err := s.extendSessionInternal(ctx, session); err != nil {
		return nil, common.NewInternalError("Failed to extend session").WithCause(err)
	}

	// Publish event
	if s.pubsub != nil {
		s.pubsub.PublishSeatHeld(ctx, req.ShowID, req.SeatID)
	}

	return s.buildStatusResponse(session), nil
}

// ReleaseSeat releases a held seat
func (s *ServiceImpl) ReleaseSeat(ctx context.Context, req *HoldSeatRequest) (*HoldStatusResponse, *common.AppError) {
	if err := req.Validate(); err != nil {
		return nil, common.NewBadRequestError("Invalid request: " + err.Error())
	}

	// Get session
	session, err := s.repo.GetSession(ctx, req.ShowID, req.SessionID)
	if err != nil {
		return nil, common.NewInternalError("Failed to get session").WithCause(err)
	}

	if session == nil {
		return nil, common.NewHoldSessionExpiredError()
	}

	// Check if session has exceeded max time
	if session.IsSessionExpired() {
		return nil, common.NewHoldSessionMaxTimeError()
	}

	// Check toggle limit
	if !session.CanToggle() {
		return nil, common.NewHoldToggleLimitError()
	}

	// Check if seat is held by this session
	found := false
	newSeats := make([]string, 0, len(session.HeldSeats))
	for _, seatID := range session.HeldSeats {
		if seatID == req.SeatID {
			found = true
		} else {
			newSeats = append(newSeats, seatID)
		}
	}

	if !found {
		return nil, common.NewHoldSeatNotHeldError(req.SeatID)
	}

	// Release the seat
	if err := s.repo.ReleaseSeat(ctx, req.ShowID, req.SeatID, req.SessionID); err != nil {
		return nil, common.NewInternalError("Failed to release seat").WithCause(err)
	}

	// Update session
	session.HeldSeats = newSeats
	session.ToggleCount++
	session.LastActive = time.Now()

	if err := s.repo.SaveSession(ctx, session, s.config.IdleTTL); err != nil {
		return nil, common.NewInternalError("Failed to save session").WithCause(err)
	}

	// Extend session and all remaining held seats
	if err := s.extendSessionInternal(ctx, session); err != nil {
		return nil, common.NewInternalError("Failed to extend session").WithCause(err)
	}

	// Publish event
	if s.pubsub != nil {
		s.pubsub.PublishSeatReleased(ctx, req.ShowID, req.SeatID)
	}

	return s.buildStatusResponse(session), nil
}

// GetSessionStatus returns current session status
func (s *ServiceImpl) GetSessionStatus(ctx context.Context, showID, sessionID string) (*HoldStatusResponse, *common.AppError) {
	session, err := s.repo.GetSession(ctx, showID, sessionID)
	if err != nil {
		return nil, common.NewInternalError("Failed to get session").WithCause(err)
	}

	if session == nil {
		// Return empty status for new session
		return &HoldStatusResponse{
			SessionID:     sessionID,
			ShowID:        showID,
			HeldSeats:     []string{},
			ToggleCount:   0,
			RemainingTime: int(s.config.IdleTTL.Seconds()),
			MaxTime:       int(s.config.MaxSessionTime.Seconds()),
			CanExtend:     true,
		}, nil
	}

	return s.buildStatusResponse(session), nil
}

// ExtendSession extends the session TTL and all held seats
func (s *ServiceImpl) ExtendSession(ctx context.Context, req *ExtendSessionRequest) (*HoldStatusResponse, *common.AppError) {
	session, err := s.repo.GetSession(ctx, req.ShowID, req.SessionID)
	if err != nil {
		return nil, common.NewInternalError("Failed to get session").WithCause(err)
	}

	if session == nil {
		return nil, common.NewHoldSessionExpiredError()
	}

	// Check if session has exceeded max time
	if session.IsSessionExpired() {
		return nil, common.NewHoldSessionMaxTimeError()
	}

	// Calculate remaining absolute time
	elapsed := time.Since(session.CreatedAt)
	remaining := s.config.MaxSessionTime - elapsed

	// Can only extend if there's time left in absolute window
	if remaining <= 0 {
		return nil, common.NewHoldSessionMaxTimeError()
	}

	// Extend TTL (min of IdleTTL and remaining absolute time)
	extendTTL := s.config.IdleTTL
	if remaining < extendTTL {
		extendTTL = remaining
	}

	// Extend session TTL
	if err := s.repo.ExtendSession(ctx, req.ShowID, req.SessionID, extendTTL); err != nil {
		return nil, common.NewInternalError("Failed to extend session").WithCause(err)
	}

	// Also extend all held seat TTLs
	for _, seatID := range session.HeldSeats {
		s.repo.ExtendSeatHold(ctx, req.ShowID, seatID, req.SessionID, extendTTL)
	}

	session.LastActive = time.Now()
	return s.buildStatusResponse(session), nil
}

// extendSessionInternal is a helper to extend session and seats (called internally)
func (s *ServiceImpl) extendSessionInternal(ctx context.Context, session *Session) error {
	// Check if session has exceeded max time
	if session.IsSessionExpired() {
		return nil // Don't extend if expired
	}

	// Calculate remaining absolute time
	elapsed := time.Since(session.CreatedAt)
	remaining := s.config.MaxSessionTime - elapsed

	// Can only extend if there's time left
	if remaining <= 0 {
		return nil // Don't extend if max time reached
	}

	// Extend TTL (min of IdleTTL and remaining absolute time)
	extendTTL := s.config.IdleTTL
	if remaining < extendTTL {
		extendTTL = remaining
	}

	// Extend session TTL
	if err := s.repo.ExtendSession(ctx, session.ShowID, session.SessionID, extendTTL); err != nil {
		return err
	}

	// Extend all held seat TTLs
	for _, seatID := range session.HeldSeats {
		s.repo.ExtendSeatHold(ctx, session.ShowID, seatID, session.SessionID, extendTTL)
	}

	return nil
}

// GetShowSeatsStatus returns status of all seats for a show
func (s *ServiceImpl) GetShowSeatsStatus(ctx context.Context, showID, sessionID string) ([]SeatStatusResponse, *common.AppError) {
	heldSeats, err := s.repo.GetHeldSeats(ctx, showID)
	if err != nil {
		return nil, common.NewInternalError("Failed to get held seats").WithCause(err)
	}

	var result []SeatStatusResponse
	for seatID, holder := range heldSeats {
		status := SeatStatusResponse{
			SeatID: seatID,
			Status: "held",
		}
		if holder == sessionID {
			status.SessionID = holder
		}
		result = append(result, status)
	}

	return result, nil
}

// GetSessionSeats returns all seats held by a session
func (s *ServiceImpl) GetSessionSeats(ctx context.Context, showID, sessionID string) ([]string, *common.AppError) {
	session, err := s.repo.GetSession(ctx, showID, sessionID)
	if err != nil {
		return nil, common.NewInternalError("Failed to get session").WithCause(err)
	}

	if session == nil {
		return []string{}, nil
	}

	return session.HeldSeats, nil
}

// ReleaseSession releases all holds for a session
func (s *ServiceImpl) ReleaseSession(ctx context.Context, showID, sessionID string) *common.AppError {
	session, err := s.repo.GetSession(ctx, showID, sessionID)
	if err != nil {
		return common.NewInternalError("Failed to get session").WithCause(err)
	}

	if session == nil {
		return nil
	}

	// Release all seats
	if err := s.repo.ReleaseAllSeats(ctx, showID, sessionID, session.HeldSeats); err != nil {
		return common.NewInternalError("Failed to release seats").WithCause(err)
	}

	// Delete session
	if err := s.repo.DeleteSession(ctx, showID, sessionID); err != nil {
		return common.NewInternalError("Failed to delete session").WithCause(err)
	}

	// Publish events
	if s.pubsub != nil {
		for _, seatID := range session.HeldSeats {
			s.pubsub.PublishSeatReleased(ctx, showID, seatID)
		}
	}

	return nil
}

// buildStatusResponse builds the response from session
func (s *ServiceImpl) buildStatusResponse(session *Session) *HoldStatusResponse {
	elapsed := time.Since(session.CreatedAt)
	remaining := s.config.MaxSessionTime - elapsed
	if remaining < 0 {
		remaining = 0
	}

	return &HoldStatusResponse{
		SessionID:     session.SessionID,
		ShowID:        session.ShowID,
		HeldSeats:     session.HeldSeats,
		ToggleCount:   session.ToggleCount,
		RemainingTime: int(remaining.Seconds()),
		MaxTime:       int(s.config.MaxSessionTime.Seconds()),
		CanExtend:     remaining > 0 && session.ToggleCount < s.config.MaxToggleCount,
	}
}
