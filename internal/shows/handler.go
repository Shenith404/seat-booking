package shows

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/shenith404/seat-booking/internal/common"
)

// Handler handles HTTP requests for show operations
type Handler struct {
	service Service
}

// NewHandler creates a new show handler
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers show routes
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/shows", h.Create)
	r.Get("/shows", h.GetAll)
	r.Get("/shows/{id}", h.GetByID)
	r.Delete("/shows/{id}", h.Delete)
	r.Get("/shows/{id}/seats", h.GetShowSeats)
}

// Create godoc
// @Summary Create a show
// @Description Create a new movie show/screening
// @Tags shows
// @Accept json
// @Produce json
// @Param request body CreateShowRequest true "Create show request"
// @Success 201 {object} common.Response{data=ShowResponse}
// @Failure 400 {object} common.Response{error=common.Error}
// @Failure 404 {object} common.Response{error=common.Error}
// @Router /api/v1/shows [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	req, appErr := common.DecodeAndValidate[CreateShowRequest](r, func(req *CreateShowRequest, v *common.Validator) {
		v.Required(req.MovieID, "movie_id")
		v.UUID(req.MovieID, "movie_id")
		v.Required(req.HallID, "hall_id")
		v.UUID(req.HallID, "hall_id")
		v.Required(req.StartTime, "start_time")
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
// @Summary Get all shows
// @Description Get all shows with pagination, optionally filtered by date
// @Tags shows
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Param date query string false "Filter by date (YYYY-MM-DD)"
// @Success 200 {object} common.Response{data=[]ShowResponse}
// @Router /api/v1/shows [get]
func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	date := r.URL.Query().Get("date")

	var results []ShowResponse
	var meta *common.Meta
	var appErr *common.AppError

	if date != "" {
		results, meta, appErr = h.service.GetByDate(r.Context(), date, page, perPage)
	} else {
		results, meta, appErr = h.service.GetAll(r.Context(), page, perPage)
	}

	if appErr != nil {
		common.Err(w, appErr)
		return
	}

	common.SuccessWithMeta(w, http.StatusOK, results, meta)
}

// GetByID godoc
// @Summary Get a show
// @Description Get a show by ID with movie and hall details
// @Tags shows
// @Produce json
// @Param id path string true "Show ID"
// @Success 200 {object} common.Response{data=ShowResponse}
// @Failure 404 {object} common.Response{error=common.Error}
// @Router /api/v1/shows/{id} [get]
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	result, appErr := h.service.GetByID(r.Context(), id)
	if appErr != nil {
		common.Err(w, appErr)
		return
	}

	common.OK(w, result)
}

// Delete godoc
// @Summary Delete a show
// @Description Delete a show by ID
// @Tags shows
// @Param id path string true "Show ID"
// @Success 204
// @Failure 404 {object} common.Response{error=common.Error}
// @Router /api/v1/shows/{id} [delete]
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if appErr := h.service.Delete(r.Context(), id); appErr != nil {
		common.Err(w, appErr)
		return
	}

	common.NoContent(w)
}

// GetShowSeats godoc
// @Summary Get show seats
// @Description Get seat availability for a show (includes held and booked status)
// @Tags shows
// @Produce json
// @Param id path string true "Show ID"
// @Param session_id query string false "Session ID to identify own holds"
// @Success 200 {object} common.Response{data=ShowSeatsResponse}
// @Failure 404 {object} common.Response{error=common.Error}
// @Router /api/v1/shows/{id}/seats [get]
func (h *Handler) GetShowSeats(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	sessionID := r.URL.Query().Get("session_id")

	result, appErr := h.service.GetShowSeats(r.Context(), id, sessionID)
	if appErr != nil {
		common.Err(w, appErr)
		return
	}

	common.OK(w, result)
}
