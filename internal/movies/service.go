package movies

import (
	"context"

	"github.com/google/uuid"
	"github.com/shenith404/seat-booking/internal/common"
)

// Service defines the interface for movie business logic
type Service interface {
	Create(ctx context.Context, req *CreateMovieRequest) (*MovieResponse, *common.AppError)
	GetByID(ctx context.Context, id string) (*MovieResponse, *common.AppError)
	GetAll(ctx context.Context, page, perPage int) ([]MovieResponse, *common.Meta, *common.AppError)
	Update(ctx context.Context, id string, req *UpdateMovieRequest) (*MovieResponse, *common.AppError)
	Delete(ctx context.Context, id string) *common.AppError
}

// ServiceImpl implements the Service interface
type ServiceImpl struct {
	repo Repository
}

// NewService creates a new movie service
func NewService(repo Repository) *ServiceImpl {
	return &ServiceImpl{repo: repo}
}

// Create creates a new movie
func (s *ServiceImpl) Create(ctx context.Context, req *CreateMovieRequest) (*MovieResponse, *common.AppError) {
	movie := &Movie{
		ID:              uuid.New(),
		Title:           req.Title,
		DurationMinutes: req.DurationMinutes,
		Description:     req.Description,
	}

	if err := s.repo.Create(ctx, movie); err != nil {
		return nil, common.NewInternalError("Failed to create movie").WithCause(err)
	}

	return movie.ToResponse(), nil
}

// GetByID retrieves a movie by ID
func (s *ServiceImpl) GetByID(ctx context.Context, id string) (*MovieResponse, *common.AppError) {
	movieID, err := uuid.Parse(id)
	if err != nil {
		return nil, common.NewBadRequestError("Invalid movie ID")
	}

	movie, err := s.repo.GetByID(ctx, movieID)
	if err != nil {
		return nil, common.NewInternalError("Failed to get movie").WithCause(err)
	}

	if movie == nil {
		return nil, common.NewNotFoundError("Movie not found")
	}

	return movie.ToResponse(), nil
}

// GetAll retrieves all movies with pagination
func (s *ServiceImpl) GetAll(ctx context.Context, page, perPage int) ([]MovieResponse, *common.Meta, *common.AppError) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	offset := (page - 1) * perPage

	movies, total, err := s.repo.GetAll(ctx, perPage, offset)
	if err != nil {
		return nil, nil, common.NewInternalError("Failed to get movies").WithCause(err)
	}

	responses := make([]MovieResponse, len(movies))
	for i, m := range movies {
		responses[i] = *m.ToResponse()
	}

	meta := common.NewPaginationMeta(page, perPage, total)
	return responses, meta, nil
}

// Update updates a movie
func (s *ServiceImpl) Update(ctx context.Context, id string, req *UpdateMovieRequest) (*MovieResponse, *common.AppError) {
	movieID, err := uuid.Parse(id)
	if err != nil {
		return nil, common.NewBadRequestError("Invalid movie ID")
	}

	movie, err := s.repo.GetByID(ctx, movieID)
	if err != nil {
		return nil, common.NewInternalError("Failed to get movie").WithCause(err)
	}

	if movie == nil {
		return nil, common.NewNotFoundError("Movie not found")
	}

	// Apply updates
	if req.Title != nil {
		movie.Title = *req.Title
	}
	if req.DurationMinutes != nil {
		movie.DurationMinutes = *req.DurationMinutes
	}
	if req.Description != nil {
		movie.Description = *req.Description
	}

	if err := s.repo.Update(ctx, movie); err != nil {
		return nil, common.NewInternalError("Failed to update movie").WithCause(err)
	}

	return movie.ToResponse(), nil
}

// Delete deletes a movie
func (s *ServiceImpl) Delete(ctx context.Context, id string) *common.AppError {
	movieID, err := uuid.Parse(id)
	if err != nil {
		return common.NewBadRequestError("Invalid movie ID")
	}

	movie, err := s.repo.GetByID(ctx, movieID)
	if err != nil {
		return common.NewInternalError("Failed to get movie").WithCause(err)
	}

	if movie == nil {
		return common.NewNotFoundError("Movie not found")
	}

	if err := s.repo.Delete(ctx, movieID); err != nil {
		return common.NewInternalError("Failed to delete movie").WithCause(err)
	}

	return nil
}
