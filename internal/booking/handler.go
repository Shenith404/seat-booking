package booking

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/shenith404/seat-booking/internal/common"
)

// Handler handles HTTP requests for booking operations
type Handler struct {
	service Service
}

// NewHandler creates a new booking handler
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers booking routes
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/bookings", h.CreateBooking)
	r.Get("/bookings/{id}", h.GetBooking)
}

// CreateBooking godoc
// @Summary Create a booking
// @Description Create a booking from held seats
// @Tags booking
// @Accept json
// @Produce json
// @Param request body CreateBookingRequest true "Create booking request"
// @Success 201 {object} common.Response{data=BookingResponse}
// @Failure 400 {object} common.Response{error=common.Error}
// @Failure 409 {object} common.Response{error=common.Error}
// @Router /api/v1/bookings [post]
func (h *Handler) CreateBooking(w http.ResponseWriter, r *http.Request) {
	req, appErr := common.DecodeAndValidate[CreateBookingRequest](r, func(req *CreateBookingRequest, v *common.Validator) {
		v.Required(req.SessionID, "session_id")
		v.UUID(req.SessionID, "session_id")
		v.Required(req.ShowID, "show_id")
		v.UUID(req.ShowID, "show_id")
		v.Required(req.CustomerEmail, "customer_email")
		v.Email(req.CustomerEmail, "customer_email")
		v.Required(req.CustomerPhone, "customer_phone")
		v.Phone(req.CustomerPhone, "customer_phone")
	})
	if appErr != nil {
		common.Err(w, appErr)
		return
	}

	result, appErr := h.service.CreateBooking(r.Context(), req)
	if appErr != nil {
		common.Err(w, appErr)
		return
	}

	common.Created(w, result)
}

// GetBooking godoc
// @Summary Get a booking
// @Description Get booking details by ID
// @Tags booking
// @Produce json
// @Param id path string true "Booking ID"
// @Success 200 {object} common.Response{data=BookingResponse}
// @Failure 404 {object} common.Response{error=common.Error}
// @Router /api/v1/bookings/{id} [get]
func (h *Handler) GetBooking(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	v := common.NewValidator()
	v.Required(id, "id")
	v.UUID(id, "id")

	if !v.Valid() {
		common.Err(w, v.ToAppError())
		return
	}

	result, appErr := h.service.GetBooking(r.Context(), id)
	if appErr != nil {
		common.Err(w, appErr)
		return
	}

	common.OK(w, result)
}
