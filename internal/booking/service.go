package booking

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/shenith404/seat-booking/internal/common"
	"github.com/shenith404/seat-booking/internal/hold"
	"github.com/shenith404/seat-booking/internal/pubsub"
	"github.com/shenith404/seat-booking/internal/worker"
)

// Service defines the interface for booking business logic
type Service interface {
	CreateBooking(ctx context.Context, req *CreateBookingRequest) (*BookingResponse, *common.AppError)
	GetBooking(ctx context.Context, id string) (*BookingResponse, *common.AppError)
}

// ServiceImpl implements the Service interface
type ServiceImpl struct {
	repo        Repository
	holdService hold.Service
	pubsub      *pubsub.PubSub
	worker      *worker.Worker
}

// NewService creates a new booking service
func NewService(repo Repository, holdService hold.Service, ps *pubsub.PubSub, w *worker.Worker) *ServiceImpl {
	return &ServiceImpl{
		repo:        repo,
		holdService: holdService,
		pubsub:      ps,
		worker:      w,
	}
}

// CreateBooking creates a new booking from held seats
func (s *ServiceImpl) CreateBooking(ctx context.Context, req *CreateBookingRequest) (*BookingResponse, *common.AppError) {
	// Validate request
	sessionID, err := uuid.Parse(req.SessionID)
	if err != nil {
		return nil, common.NewBadRequestError("Invalid session ID")
	}

	showID, err := uuid.Parse(req.ShowID)
	if err != nil {
		return nil, common.NewBadRequestError("Invalid show ID")
	}

	// Get held seats from the session
	heldSeats, appErr := s.holdService.GetSessionSeats(ctx, req.ShowID, req.SessionID)
	if appErr != nil {
		return nil, appErr
	}

	if len(heldSeats) == 0 {
		return nil, common.NewBookingSeatsNotHeldError()
	}

	// Parse seat IDs
	seatIDs := make([]uuid.UUID, len(heldSeats))
	for i, seatID := range heldSeats {
		id, err := uuid.Parse(seatID)
		if err != nil {
			return nil, common.NewBadRequestError("Invalid seat ID in session")
		}
		seatIDs[i] = id
	}

	// Create booking with ACID transaction
	booking, tickets, err := s.repo.CreateBookingWithTickets(ctx, CreateBookingParams{
		ShowID:        showID,
		CustomerEmail: req.CustomerEmail,
		CustomerPhone: req.CustomerPhone,
	}, seatIDs)

	if err != nil {
		if errors.Is(err, ErrSeatsAlreadyBooked) {
			return nil, common.NewBookingSeatsTakenError(heldSeats)
		}
		return nil, common.NewBookingFailedError("Failed to create booking").WithCause(err)
	}

	// Release the hold session (seats are now booked)
	s.holdService.ReleaseSession(ctx, req.ShowID, req.SessionID)

	// Publish booking event
	if s.pubsub != nil {
		s.pubsub.PublishSeatsBooked(ctx, req.ShowID, heldSeats)
	}

	// Submit background job for QR codes and email
	ticketIDs := make([]uuid.UUID, len(tickets))
	for i, t := range tickets {
		ticketIDs[i] = t.ID
	}

	if s.worker != nil {
		s.worker.Submit(worker.BookingJob{
			BookingID:     booking.ID,
			CustomerEmail: booking.CustomerEmail,
			CustomerPhone: booking.CustomerPhone,
			ShowID:        showID,
			SeatIDs:       seatIDs,
			TicketIDs:     ticketIDs,
		})
	}

	// Build response
	response := &BookingResponse{
		ID:            booking.ID.String(),
		ShowID:        booking.ShowID.String(),
		CustomerEmail: booking.CustomerEmail,
		CustomerPhone: maskPhone(booking.CustomerPhone),
		Status:        booking.Status,
		CreatedAt:     booking.CreatedAt,
		Tickets:       make([]TicketResponse, len(tickets)),
	}

	for i, t := range tickets {
		response.Tickets[i] = TicketResponse{
			ID:     t.ID.String(),
			SeatID: t.SeatID.String(),
		}
	}

	_ = sessionID // Used for validation
	return response, nil
}

// GetBooking retrieves a booking by ID
func (s *ServiceImpl) GetBooking(ctx context.Context, id string) (*BookingResponse, *common.AppError) {
	bookingID, err := uuid.Parse(id)
	if err != nil {
		return nil, common.NewBadRequestError("Invalid booking ID")
	}

	booking, err := s.repo.GetBooking(ctx, bookingID)
	if err != nil {
		return nil, common.NewInternalError("Failed to get booking").WithCause(err)
	}

	if booking == nil {
		return nil, common.NewNotFoundError("Booking not found")
	}

	tickets, err := s.repo.GetBookingTickets(ctx, bookingID)
	if err != nil {
		return nil, common.NewInternalError("Failed to get tickets").WithCause(err)
	}

	response := &BookingResponse{
		ID:            booking.ID.String(),
		ShowID:        booking.ShowID.String(),
		CustomerEmail: maskEmail(booking.CustomerEmail),
		CustomerPhone: maskPhone(booking.CustomerPhone),
		Status:        booking.Status,
		CreatedAt:     booking.CreatedAt,
		Tickets:       make([]TicketResponse, len(tickets)),
	}

	for i, t := range tickets {
		response.Tickets[i] = TicketResponse{
			ID:     t.ID.String(),
			SeatID: t.SeatID.String(),
		}
	}

	return response, nil
}

// maskPhone masks a phone number for privacy
func maskPhone(phone string) string {
	if len(phone) <= 4 {
		return "****"
	}
	return "******" + phone[len(phone)-4:]
}

// maskEmail masks an email for privacy
func maskEmail(email string) string {
	if len(email) <= 4 {
		return "****"
	}
	atIndex := -1
	for i, c := range email {
		if c == '@' {
			atIndex = i
			break
		}
	}
	if atIndex <= 2 {
		return email
	}
	return email[:2] + "***" + email[atIndex:]
}
