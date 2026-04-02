package seats

import (
	"context"

	"github.com/google/uuid"
	"github.com/shenith404/seat-booking/internal/common"
)

// Service defines the interface for seat/hall business logic
type Service interface {
	// Hall operations
	CreateHall(ctx context.Context, req *CreateHallRequest) (*HallResponse, *common.AppError)
	GetHallByID(ctx context.Context, id string) (*HallResponse, *common.AppError)
	GetAllHalls(ctx context.Context) ([]HallResponse, *common.AppError)
	DeleteHall(ctx context.Context, id string) *common.AppError
	UpdateHallWithSeats(ctx context.Context, id string, req *UpdateHallRequest) *common.AppError

	// Seat operations
	GetSeatsByHallID(ctx context.Context, hallID string) ([]SeatResponse, *common.AppError)
}

// ServiceImpl implements the Service interface
type ServiceImpl struct {
	repo Repository
}

// NewService creates a new seat service
func NewService(repo Repository) *ServiceImpl {
	return &ServiceImpl{repo: repo}
}

// CreateHall creates a new hall with seats
func (s *ServiceImpl) CreateHall(ctx context.Context, req *CreateHallRequest) (*HallResponse, *common.AppError) {
	hallID := uuid.New()
	totalSeats := 0

	// Calculate total seats
	for _, row := range req.SeatLayout {
		totalSeats += row.SeatCount
	}

	hall := &Hall{
		ID:         hallID,
		Name:       req.Name,
		TotalSeats: totalSeats,
	}

	// Create seats
	var seats []Seat
	for _, row := range req.SeatLayout {
		for i := 1; i <= row.SeatCount; i++ {
			seats = append(seats, Seat{
				ID:         uuid.New(),
				HallID:     hallID,
				RowName:    row.RowName,
				SeatNumber: i,
			})
		}
	}

	// Create hall with seats in a single transaction
	if err := s.repo.CreateHallWithSeats(ctx, hall, seats); err != nil {
		return nil, common.NewInternalError("Failed to create hall with seats").WithCause(err)
	}

	// Build response
	response := hall.ToResponse()
	response.Seats = make([]SeatResponse, len(seats))
	for i, seat := range seats {
		response.Seats[i] = *seat.ToResponse()
	}

	return response, nil
}

func (s *ServiceImpl) UpdateHallWithSeats(ctx context.Context, id string, req *UpdateHallRequest) *common.AppError {
	hallID, err := uuid.Parse(id)
	if err != nil {
		return common.NewBadRequestError("Invalid hall ID")
	}
	// Create seats
	var seats []Seat
	for _, row := range req.SeatLayout {
		for i := 1; i <= row.SeatCount; i++ {
			seats = append(seats, Seat{
				ID:         uuid.New(),
				HallID:     hallID,
				RowName:    row.RowName,
				SeatNumber: i,
			})
		}
	}
	err = s.repo.UpdateHallWithSeats(ctx, hallID, req.Name, seats)
	if err != nil {
		return common.NewInternalError("Failed to update hall with seats").WithCause(err)
	}

	return nil

}

// GetHallByID retrieves a hall by ID with its seats
func (s *ServiceImpl) GetHallByID(ctx context.Context, id string) (*HallResponse, *common.AppError) {
	hallID, err := uuid.Parse(id)
	if err != nil {
		return nil, common.NewBadRequestError("Invalid hall ID")
	}

	hall, err := s.repo.GetHallByID(ctx, hallID)
	if err != nil {
		return nil, common.NewInternalError("Failed to get hall").WithCause(err)
	}

	if hall == nil {
		return nil, common.NewNotFoundError("Hall not found")
	}

	// Get seats
	seats, err := s.repo.GetSeatsByHallID(ctx, hallID)
	if err != nil {
		return nil, common.NewInternalError("Failed to get seats").WithCause(err)
	}

	response := hall.ToResponse()
	response.Seats = make([]SeatResponse, len(seats))
	for i, seat := range seats {
		response.Seats[i] = *seat.ToResponse()
	}

	return response, nil
}

// GetAllHalls retrieves all halls
func (s *ServiceImpl) GetAllHalls(ctx context.Context) ([]HallResponse, *common.AppError) {
	halls, err := s.repo.GetAllHalls(ctx)
	if err != nil {
		return nil, common.NewInternalError("Failed to get halls").WithCause(err)
	}

	responses := make([]HallResponse, len(halls))
	for i, h := range halls {
		responses[i] = *h.ToResponse()
	}

	return responses, nil
}

// DeleteHall deletes a hall
func (s *ServiceImpl) DeleteHall(ctx context.Context, id string) *common.AppError {
	hallID, err := uuid.Parse(id)
	if err != nil {
		return common.NewBadRequestError("Invalid hall ID")
	}

	hall, err := s.repo.GetHallByID(ctx, hallID)
	if err != nil {
		return common.NewInternalError("Failed to get hall").WithCause(err)
	}

	if hall == nil {
		return common.NewNotFoundError("Hall not found")
	}

	if err := s.repo.DeleteHall(ctx, hallID); err != nil {
		return common.NewInternalError("Failed to delete hall").WithCause(err)
	}

	return nil
}

// GetSeatsByHallID retrieves seats for a hall
func (s *ServiceImpl) GetSeatsByHallID(ctx context.Context, hallID string) ([]SeatResponse, *common.AppError) {
	id, err := uuid.Parse(hallID)
	if err != nil {
		return nil, common.NewBadRequestError("Invalid hall ID")
	}

	seats, err := s.repo.GetSeatsByHallID(ctx, id)
	if err != nil {
		return nil, common.NewInternalError("Failed to get seats").WithCause(err)
	}

	responses := make([]SeatResponse, len(seats))
	for i, seat := range seats {
		responses[i] = *seat.ToResponse()
	}

	return responses, nil
}
