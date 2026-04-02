package seats

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/shenith404/seat-booking/internal/common"
)

// Handler handles HTTP requests for seat/hall operations
type Handler struct {
	service Service
}

// NewHandler creates a new seat handler
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers seat/hall routes
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/halls", h.CreateHall)
	r.Put("/halls/{id}", h.UpdateHallWithSeats)
	r.Get("/halls", h.GetAllHalls)
	r.Get("/halls/{id}", h.GetHallByID)
	r.Delete("/halls/{id}", h.DeleteHall)
	r.Get("/halls/{id}/seats", h.GetSeatsByHallID)
}

// CreateHall godoc
// @Summary Create a hall
// @Description Create a new hall with seat layout
// @Tags halls
// @Accept json
// @Produce json
// @Param request body CreateHallRequest true "Create hall request"
// @Success 201 {object} common.Response{data=HallResponse}
// @Failure 400 {object} common.Response{error=common.Error}
// @Router /api/v1/halls [post]
func (h *Handler) CreateHall(w http.ResponseWriter, r *http.Request) {
	req, appErr := common.DecodeAndValidate[CreateHallRequest](r, func(req *CreateHallRequest, v *common.Validator) {
		v.Required(req.Name, "name")
		v.MaxLength(req.Name, 255, "name")
		v.Check(len(req.SeatLayout) > 0, "seat_layout", "At least one row is required")
		for i, row := range req.SeatLayout {
			v.Check(row.RowName != "", "seat_layout", "Row name is required for row "+string(rune('1'+i)))
			v.Check(row.SeatCount > 0, "seat_layout", "Seat count must be positive for row "+row.RowName)
		}
	})
	if appErr != nil {
		common.Err(w, appErr)
		return
	}

	result, appErr := h.service.CreateHall(r.Context(), req)
	if appErr != nil {
		common.Err(w, appErr)
		return
	}

	common.Created(w, result)
}

// UpdateHallWithSeats godoc
// @Summary Update a hall with seats
// @Description Update hall name and seat layout
// @Tags halls
// @Accept json
// @Produce json
// @Param id path string true "Hall ID"
// @Param request body UpdateHallRequest true "Update hall request"
// @Success 204
// @Failure 400 {object} common.Response{error=common.Error}
// @Failure 404 {object} common.Response{error=common.Error}
// @Router /api/v1/halls/{id} [put]
func (h *Handler) UpdateHallWithSeats(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	req, appErr := common.DecodeAndValidate(r, func(req *UpdateHallRequest, v *common.Validator) {
		v.Required(req.Name, "name")
		v.MaxLength(req.Name, 255, "name")
		v.Check(len(req.SeatLayout) > 0, "seat_layout", "At least one row is required")
		for i, row := range req.SeatLayout {
			v.Check(row.RowName != "", "seat_layout", "Row name is required for row "+string(rune('1'+i)))
			v.Check(row.SeatCount > 0, "seat_layout", "Seat count must be positive for row "+row.RowName)
		}
	})
	if appErr != nil {
		common.Err(w, appErr)
		return
	}

	if appErr := h.service.UpdateHallWithSeats(r.Context(), id, req); appErr != nil {
		common.Err(w, appErr)
		return
	}

	common.NoContent(w)
}

// GetAllHalls godoc
// @Summary Get all halls
// @Description Get all halls
// @Tags halls
// @Produce json
// @Success 200 {object} common.Response{data=[]HallResponse}
// @Router /api/v1/halls [get]
func (h *Handler) GetAllHalls(w http.ResponseWriter, r *http.Request) {
	results, appErr := h.service.GetAllHalls(r.Context())
	if appErr != nil {
		common.Err(w, appErr)
		return
	}

	common.OK(w, results)
}

// GetHallByID godoc
// @Summary Get a hall
// @Description Get a hall by ID with its seats
// @Tags halls
// @Produce json
// @Param id path string true "Hall ID"
// @Success 200 {object} common.Response{data=HallResponse}
// @Failure 404 {object} common.Response{error=common.Error}
// @Router /api/v1/halls/{id} [get]
func (h *Handler) GetHallByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	result, appErr := h.service.GetHallByID(r.Context(), id)
	if appErr != nil {
		common.Err(w, appErr)
		return
	}

	common.OK(w, result)
}

// DeleteHall godoc
// @Summary Delete a hall
// @Description Delete a hall by ID
// @Tags halls
// @Param id path string true "Hall ID"
// @Success 204
// @Failure 404 {object} common.Response{error=common.Error}
// @Router /api/v1/halls/{id} [delete]
func (h *Handler) DeleteHall(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if appErr := h.service.DeleteHall(r.Context(), id); appErr != nil {
		common.Err(w, appErr)
		return
	}

	common.NoContent(w)
}

// GetSeatsByHallID godoc
// @Summary Get hall seats
// @Description Get all seats for a hall
// @Tags halls
// @Produce json
// @Param id path string true "Hall ID"
// @Success 200 {object} common.Response{data=[]SeatResponse}
// @Failure 404 {object} common.Response{error=common.Error}
// @Router /api/v1/halls/{id}/seats [get]
func (h *Handler) GetSeatsByHallID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	results, appErr := h.service.GetSeatsByHallID(r.Context(), id)
	if appErr != nil {
		common.Err(w, appErr)
		return
	}

	common.OK(w, results)
}
