package common

import (
	"fmt"
	"net/http"
)

// ErrorCode represents application-specific error codes
type ErrorCode string

const (
	// General errors
	ErrCodeInternal     ErrorCode = "INTERNAL_ERROR"
	ErrCodeBadRequest   ErrorCode = "BAD_REQUEST"
	ErrCodeNotFound     ErrorCode = "NOT_FOUND"
	ErrCodeUnauthorized ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden    ErrorCode = "FORBIDDEN"
	ErrCodeConflict     ErrorCode = "CONFLICT"
	ErrCodeValidation   ErrorCode = "VALIDATION_ERROR"

	// Hold-specific errors
	ErrCodeHoldSessionExpired  ErrorCode = "HOLD_SESSION_EXPIRED"
	ErrCodeHoldSessionMaxTime  ErrorCode = "HOLD_SESSION_MAX_TIME_EXCEEDED"
	ErrCodeHoldToggleLimitHit  ErrorCode = "HOLD_TOGGLE_LIMIT_EXCEEDED"
	ErrCodeHoldSeatUnavailable ErrorCode = "HOLD_SEAT_UNAVAILABLE"
	ErrCodeHoldSeatNotHeld     ErrorCode = "HOLD_SEAT_NOT_HELD"

	// Booking-specific errors
	ErrCodeBookingSeatsNotHeld ErrorCode = "BOOKING_SEATS_NOT_HELD"
	ErrCodeBookingSeatsTaken   ErrorCode = "BOOKING_SEATS_ALREADY_BOOKED"
	ErrCodeBookingFailed       ErrorCode = "BOOKING_FAILED"

	// Rate limiting
	ErrCodeRateLimitExceeded ErrorCode = "RATE_LIMIT_EXCEEDED"
)

// AppError is the application's custom error type
type AppError struct {
	Code       ErrorCode `json:"code"`
	Message    string    `json:"message"`
	Details    any       `json:"details,omitempty"`
	HTTPStatus int       `json:"-"`
	Err        error     `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (cause: %v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the wrapped error
func (e *AppError) Unwrap() error {
	return e.Err
}

// WithDetails adds details to the error
func (e *AppError) WithDetails(details any) *AppError {
	e.Details = details
	return e
}

// WithCause wraps an underlying error
func (e *AppError) WithCause(err error) *AppError {
	e.Err = err
	return e
}

// NewAppError creates a new AppError
func NewAppError(code ErrorCode, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

// Pre-defined error constructors
func NewInternalError(message string) *AppError {
	return NewAppError(ErrCodeInternal, message, http.StatusInternalServerError)
}

func NewBadRequestError(message string) *AppError {
	return NewAppError(ErrCodeBadRequest, message, http.StatusBadRequest)
}

func NewNotFoundError(message string) *AppError {
	return NewAppError(ErrCodeNotFound, message, http.StatusNotFound)
}

func NewValidationError(message string, details any) *AppError {
	return NewAppError(ErrCodeValidation, message, http.StatusBadRequest).WithDetails(details)
}

func NewConflictError(message string) *AppError {
	return NewAppError(ErrCodeConflict, message, http.StatusConflict)
}

func NewUnauthorizedError(message string) *AppError {
	return NewAppError(ErrCodeUnauthorized, message, http.StatusUnauthorized)
}

func NewForbiddenError(message string) *AppError {
	return NewAppError(ErrCodeForbidden, message, http.StatusForbidden)
}

// Hold-specific errors
func NewHoldSessionExpiredError() *AppError {
	return NewAppError(ErrCodeHoldSessionExpired, "Hold session has expired", http.StatusGone)
}

func NewHoldSessionMaxTimeError() *AppError {
	return NewAppError(ErrCodeHoldSessionMaxTime, "Session has exceeded maximum time limit of 10 minutes", http.StatusGone)
}

func NewHoldToggleLimitError() *AppError {
	return NewAppError(ErrCodeHoldToggleLimitHit, "Maximum toggle limit of 15 actions reached", http.StatusTooManyRequests)
}

func NewHoldSeatUnavailableError(seatID string) *AppError {
	return NewAppError(ErrCodeHoldSeatUnavailable, "Seat is already held by another user", http.StatusConflict).
		WithDetails(map[string]string{"seat_id": seatID})
}

func NewHoldSeatNotHeldError(seatID string) *AppError {
	return NewAppError(ErrCodeHoldSeatNotHeld, "Seat is not held by this session", http.StatusBadRequest).
		WithDetails(map[string]string{"seat_id": seatID})
}

// Booking-specific errors
func NewBookingSeatsNotHeldError() *AppError {
	return NewAppError(ErrCodeBookingSeatsNotHeld, "No seats are held for this session", http.StatusBadRequest)
}

func NewBookingSeatsTakenError(seatIDs []string) *AppError {
	return NewAppError(ErrCodeBookingSeatsTaken, "Some seats have already been booked", http.StatusConflict).
		WithDetails(map[string]any{"seat_ids": seatIDs})
}

func NewBookingFailedError(reason string) *AppError {
	return NewAppError(ErrCodeBookingFailed, reason, http.StatusInternalServerError)
}

// Rate limiting
func NewRateLimitError() *AppError {
	return NewAppError(ErrCodeRateLimitExceeded, "Rate limit exceeded. Please try again later", http.StatusTooManyRequests)
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// AsAppError attempts to convert an error to AppError
func AsAppError(err error) (*AppError, bool) {
	appErr, ok := err.(*AppError)
	return appErr, ok
}
