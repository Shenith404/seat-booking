package hold

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/shenith404/seat-booking/internal/common"
)

// Handler handles HTTP requests for hold operations
type Handler struct {
	service Service
}

// NewHandler creates a new hold handler
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers hold routes
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/hold", h.HoldSeat)
	r.Delete("/hold", h.ReleaseSeat)
	r.Get("/hold/status", h.GetStatus)
	r.Post("/hold/extend", h.ExtendSession)
	r.Get("/hold/seats", h.GetShowSeats)
}

// HoldSeat godoc
// @Summary Hold a seat
// @Description Hold a seat for the user's session
// @Tags hold
// @Accept json
// @Produce json
// @Param request body HoldSeatRequest true "Hold seat request"
// @Success 200 {object} common.Response{data=HoldStatusResponse}
// @Failure 400 {object} common.Response{error=common.Error}
// @Failure 409 {object} common.Response{error=common.Error}
// @Failure 429 {object} common.Response{error=common.Error}
// @Router /api/v1/hold [post]
func (h *Handler) HoldSeat(w http.ResponseWriter, r *http.Request) {
	req, appErr := common.DecodeAndValidate(r, func(req *HoldSeatRequest, v *common.Validator) {
		v.Required(req.SessionID, "session_id")
		v.UUID(req.SessionID, "session_id")
		v.Required(req.ShowID, "show_id")
		v.UUID(req.ShowID, "show_id")
		v.Required(req.SeatID, "seat_id")
		v.UUID(req.SeatID, "seat_id")
	})
	if appErr != nil {
		common.Err(w, appErr)
		return
	}

	result, appErr := h.service.HoldSeat(r.Context(), req)
	if appErr != nil {
		common.Err(w, appErr)
		return
	}

	common.OK(w, result)
}

// ReleaseSeat godoc
// @Summary Release a held seat
// @Description Release a seat from the user's session
// @Tags hold
// @Accept json
// @Produce json
// @Param request body HoldSeatRequest true "Release seat request"
// @Success 200 {object} common.Response{data=HoldStatusResponse}
// @Failure 400 {object} common.Response{error=common.Error}
// @Router /api/v1/hold [delete]
func (h *Handler) ReleaseSeat(w http.ResponseWriter, r *http.Request) {
	req, appErr := common.DecodeAndValidate[HoldSeatRequest](r, func(req *HoldSeatRequest, v *common.Validator) {
		v.Required(req.SessionID, "session_id")
		v.UUID(req.SessionID, "session_id")
		v.Required(req.ShowID, "show_id")
		v.UUID(req.ShowID, "show_id")
		v.Required(req.SeatID, "seat_id")
		v.UUID(req.SeatID, "seat_id")
	})
	if appErr != nil {
		common.Err(w, appErr)
		return
	}

	result, appErr := h.service.ReleaseSeat(r.Context(), req)
	if appErr != nil {
		common.Err(w, appErr)
		return
	}

	common.OK(w, result)
}

// GetStatus godoc
// @Summary Get session status
// @Description Get the current hold session status
// @Tags hold
// @Produce json
// @Param session_id query string true "Session ID"
// @Param show_id query string true "Show ID"
// @Success 200 {object} common.Response{data=HoldStatusResponse}
// @Failure 400 {object} common.Response{error=common.Error}
// @Router /api/v1/hold/status [get]
func (h *Handler) GetStatus(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	showID := r.URL.Query().Get("show_id")

	v := common.NewValidator()
	v.Required(sessionID, "session_id")
	v.UUID(sessionID, "session_id")
	v.Required(showID, "show_id")
	v.UUID(showID, "show_id")

	if !v.Valid() {
		common.Err(w, v.ToAppError())
		return
	}

	result, appErr := h.service.GetSessionStatus(r.Context(), showID, sessionID)
	if appErr != nil {
		common.Err(w, appErr)
		return
	}

	common.OK(w, result)
}

// ExtendSession godoc
// @Summary Extend session
// @Description Extend the hold session TTL
// @Tags hold
// @Accept json
// @Produce json
// @Param request body ExtendSessionRequest true "Extend session request"
// @Success 200 {object} common.Response{data=HoldStatusResponse}
// @Failure 400 {object} common.Response{error=common.Error}
// @Failure 410 {object} common.Response{error=common.Error}
// @Router /api/v1/hold/extend [post]
func (h *Handler) ExtendSession(w http.ResponseWriter, r *http.Request) {
	req, appErr := common.DecodeAndValidate[ExtendSessionRequest](r, func(req *ExtendSessionRequest, v *common.Validator) {
		v.Required(req.SessionID, "session_id")
		v.UUID(req.SessionID, "session_id")
		v.Required(req.ShowID, "show_id")
		v.UUID(req.ShowID, "show_id")
	})
	if appErr != nil {
		common.Err(w, appErr)
		return
	}

	result, appErr := h.service.ExtendSession(r.Context(), req)
	if appErr != nil {
		common.Err(w, appErr)
		return
	}

	common.OK(w, result)
}

// GetShowSeats godoc
// @Summary Get show seats status
// @Description Get the status of all seats for a show
// @Tags hold
// @Produce json
// @Param show_id query string true "Show ID"
// @Param session_id query string false "Session ID (optional, to identify own holds)"
// @Success 200 {object} common.Response{data=[]SeatStatusResponse}
// @Failure 400 {object} common.Response{error=common.Error}
// @Router /api/v1/hold/seats [get]
func (h *Handler) GetShowSeats(w http.ResponseWriter, r *http.Request) {
	showID := r.URL.Query().Get("show_id")
	sessionID := r.URL.Query().Get("session_id")

	v := common.NewValidator()
	v.Required(showID, "show_id")
	v.UUID(showID, "show_id")
	if sessionID != "" {
		v.UUID(sessionID, "session_id")
	}

	if !v.Valid() {
		common.Err(w, v.ToAppError())
		return
	}

	result, appErr := h.service.GetShowSeatsStatus(r.Context(), showID, sessionID)
	if appErr != nil {
		common.Err(w, appErr)
		return
	}

	common.OK(w, result)
}
