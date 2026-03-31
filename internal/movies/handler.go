package movies

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/shenith404/seat-booking/internal/common"
)

// Handler handles HTTP requests for movie operations
type Handler struct {
	service Service
}

// NewHandler creates a new movie handler
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers movie routes
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/movies", h.Create)
	r.Get("/movies", h.GetAll)
	r.Get("/movies/{id}", h.GetByID)
	r.Put("/movies/{id}", h.Update)
	r.Delete("/movies/{id}", h.Delete)
}

// Create godoc
// @Summary Create a movie
// @Description Create a new movie
// @Tags movies
// @Accept json
// @Produce json
// @Param request body CreateMovieRequest true "Create movie request"
// @Success 201 {object} common.Response{data=MovieResponse}
// @Failure 400 {object} common.Response{error=common.Error}
// @Router /api/v1/movies [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	req, appErr := common.DecodeAndValidate[CreateMovieRequest](r, func(req *CreateMovieRequest, v *common.Validator) {
		v.Required(req.Title, "title")
		v.MaxLength(req.Title, 255, "title")
		v.Positive(req.DurationMinutes, "duration_minutes")
	})
	if appErr != nil {
		common.Err(w, appErr)
		return
	}

	result, appErr := h.service.Create(r.Context(), req)
	if appErr != nil {
		common.Err(w, appErr)
		return
	}

	common.Created(w, result)
}

// GetAll godoc
// @Summary Get all movies
// @Description Get all movies with pagination
// @Tags movies
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Success 200 {object} common.Response{data=[]MovieResponse}
// @Router /api/v1/movies [get]
func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))

	results, meta, appErr := h.service.GetAll(r.Context(), page, perPage)
	if appErr != nil {
		common.Err(w, appErr)
		return
	}

	common.SuccessWithMeta(w, http.StatusOK, results, meta)
}

// GetByID godoc
// @Summary Get a movie
// @Description Get a movie by ID
// @Tags movies
// @Produce json
// @Param id path string true "Movie ID"
// @Success 200 {object} common.Response{data=MovieResponse}
// @Failure 404 {object} common.Response{error=common.Error}
// @Router /api/v1/movies/{id} [get]
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	result, appErr := h.service.GetByID(r.Context(), id)
	if appErr != nil {
		common.Err(w, appErr)
		return
	}

	common.OK(w, result)
}

// Update godoc
// @Summary Update a movie
// @Description Update a movie by ID
// @Tags movies
// @Accept json
// @Produce json
// @Param id path string true "Movie ID"
// @Param request body UpdateMovieRequest true "Update movie request"
// @Success 200 {object} common.Response{data=MovieResponse}
// @Failure 404 {object} common.Response{error=common.Error}
// @Router /api/v1/movies/{id} [put]
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	req, appErr := common.DecodeAndValidate[UpdateMovieRequest](r, func(req *UpdateMovieRequest, v *common.Validator) {
		if req.Title != nil {
			v.MaxLength(*req.Title, 255, "title")
		}
		if req.DurationMinutes != nil {
			v.Positive(*req.DurationMinutes, "duration_minutes")
		}
	})
	if appErr != nil {
		common.Err(w, appErr)
		return
	}

	result, appErr := h.service.Update(r.Context(), id, req)
	if appErr != nil {
		common.Err(w, appErr)
		return
	}

	common.OK(w, result)
}

// Delete godoc
// @Summary Delete a movie
// @Description Delete a movie by ID
// @Tags movies
// @Param id path string true "Movie ID"
// @Success 204
// @Failure 404 {object} common.Response{error=common.Error}
// @Router /api/v1/movies/{id} [delete]
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if appErr := h.service.Delete(r.Context(), id); appErr != nil {
		common.Err(w, appErr)
		return
	}

	common.NoContent(w)
}
