package shows

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shenith404/seat-booking/internal/common"
	"github.com/shenith404/seat-booking/internal/hold"
	"github.com/shenith404/seat-booking/internal/movies"
)

// Service defines the interface for show business logic
type Service interface {
	Create(ctx context.Context, req *CreateShowRequest) (*ShowResponse, *common.AppError)
	GetByID(ctx context.Context, id string) (*ShowResponse, *common.AppError)
	GetAll(ctx context.Context, page, perPage int) ([]ShowResponse, *common.Meta, *common.AppError)
	GetByDate(ctx context.Context, date string, page, perPage int) ([]ShowResponse, *common.Meta, *common.AppError)
	Delete(ctx context.Context, id string) *common.AppError
	GetShowSeats(ctx context.Context, showID, sessionID string) (*ShowSeatsResponse, *common.AppError)
}

// ServiceImpl implements the Service interface
type ServiceImpl struct {
	repo        Repository
	movieRepo   movies.Repository
	holdService hold.Service
}

// NewService creates a new show service
func NewService(repo Repository, movieRepo movies.Repository, holdService hold.Service) *ServiceImpl {
	return &ServiceImpl{
		repo:        repo,
		movieRepo:   movieRepo,
		holdService: holdService,
	}
}

// Create creates a new show
func (s *ServiceImpl) Create(ctx context.Context, req *CreateShowRequest) (*ShowResponse, *common.AppError) {
	movieID, err := uuid.Parse(req.MovieID)
	if err != nil {
		return nil, common.NewBadRequestError("Invalid movie ID")
	}

	hallID, err := uuid.Parse(req.HallID)
	if err != nil {
		return nil, common.NewBadRequestError("Invalid hall ID")
	}

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		return nil, common.NewBadRequestError("Invalid start time format. Use ISO 8601 (RFC3339)")
	}

	// Get movie to calculate end time
	movie, err := s.movieRepo.GetByID(ctx, movieID)
	if err != nil {
		return nil, common.NewInternalError("Failed to get movie").WithCause(err)
	}
	if movie == nil {
		return nil, common.NewNotFoundError("Movie not found")
	}

	endTime := startTime.Add(time.Duration(movie.DurationMinutes) * time.Minute)

	show := &Show{
		ID:        uuid.New(),
		MovieID:   movieID,
		HallID:    hallID,
		StartTime: startTime,
		EndTime:   endTime,
	}

	if err := s.repo.Create(ctx, show); err != nil {
		return nil, common.NewInternalError("Failed to create show").WithCause(err)
	}

	return show.ToResponse(), nil
}

// GetByID retrieves a show by ID with details
func (s *ServiceImpl) GetByID(ctx context.Context, id string) (*ShowResponse, *common.AppError) {
	showID, err := uuid.Parse(id)
	if err != nil {
		return nil, common.NewBadRequestError("Invalid show ID")
	}

	show, movie, hall, err := s.repo.GetShowWithDetails(ctx, showID)
	if err != nil {
		return nil, common.NewInternalError("Failed to get show").WithCause(err)
	}

	if show == nil {
		return nil, common.NewNotFoundError("Show not found")
	}

	response := show.ToResponse()
	response.Movie = movie
	response.Hall = hall

	return response, nil
}

// GetAll retrieves all shows with pagination
func (s *ServiceImpl) GetAll(ctx context.Context, page, perPage int) ([]ShowResponse, *common.Meta, *common.AppError) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	offset := (page - 1) * perPage

	shows, total, err := s.repo.GetAll(ctx, perPage, offset)
	if err != nil {
		return nil, nil, common.NewInternalError("Failed to get shows").WithCause(err)
	}

	responses := make([]ShowResponse, len(shows))
	for i, sh := range shows {
		responses[i] = *sh.ToResponse()
	}

	meta := common.NewPaginationMeta(page, perPage, total)
	return responses, meta, nil
}

// GetByDate retrieves shows for a specific date
func (s *ServiceImpl) GetByDate(ctx context.Context, dateStr string, page, perPage int) ([]ShowResponse, *common.Meta, *common.AppError) {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, nil, common.NewBadRequestError("Invalid date format. Use YYYY-MM-DD")
	}

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	offset := (page - 1) * perPage

	shows, total, err := s.repo.GetByDate(ctx, date, perPage, offset)
	if err != nil {
		return nil, nil, common.NewInternalError("Failed to get shows").WithCause(err)
	}

	responses := make([]ShowResponse, len(shows))
	for i, sh := range shows {
		responses[i] = *sh.ToResponse()
	}

	meta := common.NewPaginationMeta(page, perPage, total)
	return responses, meta, nil
}

// Delete deletes a show
func (s *ServiceImpl) Delete(ctx context.Context, id string) *common.AppError {
	showID, err := uuid.Parse(id)
	if err != nil {
		return common.NewBadRequestError("Invalid show ID")
	}

	show, err := s.repo.GetByID(ctx, showID)
	if err != nil {
		return common.NewInternalError("Failed to get show").WithCause(err)
	}

	if show == nil {
		return common.NewNotFoundError("Show not found")
	}

	if err := s.repo.Delete(ctx, showID); err != nil {
		return common.NewInternalError("Failed to delete show").WithCause(err)
	}

	return nil
}

// GetShowSeats retrieves seat availability for a show
func (s *ServiceImpl) GetShowSeats(ctx context.Context, showID, sessionID string) (*ShowSeatsResponse, *common.AppError) {
	id, err := uuid.Parse(showID)
	if err != nil {
		return nil, common.NewBadRequestError("Invalid show ID")
	}

	// Get seats from database (with booked status)
	seats, err := s.repo.GetShowSeats(ctx, id)
	if err != nil {
		return nil, common.NewInternalError("Failed to get seats").WithCause(err)
	}

	// Get held seats from Redis
	heldSeats, appErr := s.holdService.GetShowSeatsStatus(ctx, showID, sessionID)
	if appErr != nil {
		return nil, appErr
	}

	// Build held seats map
	heldMap := make(map[string]hold.SeatStatusResponse)
	for _, hs := range heldSeats {
		heldMap[hs.SeatID] = hs
	}

	// Merge status (held takes precedence over available, booked takes precedence over held)
	for i := range seats {
		if seats[i].Status == "available" {
			if held, ok := heldMap[seats[i].ID]; ok {
				seats[i].Status = held.Status
			}
		}
	}

	return &ShowSeatsResponse{
		ShowID: showID,
		Seats:  seats,
	}, nil
}
