package common

import (
	"encoding/json"
	"log"
	"net/http"
)

// Response is the standard API response wrapper
type Response struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   *Error `json:"error,omitempty"`
	Meta    *Meta  `json:"meta,omitempty"`
}

// Error represents the error structure in API responses
type Error struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Details any       `json:"details,omitempty"`
}

// Meta holds pagination and other metadata
type Meta struct {
	Page       int   `json:"page,omitempty"`
	PerPage    int   `json:"per_page,omitempty"`
	Total      int64 `json:"total,omitempty"`
	TotalPages int   `json:"total_pages,omitempty"`
}

// JSON writes a JSON response
func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

// Success sends a successful response
func Success(w http.ResponseWriter, status int, data any) {
	resp := Response{
		Success: true,
		Data:    data,
	}
	JSON(w, status, resp)
}

// SuccessWithMeta sends a successful response with metadata
func SuccessWithMeta(w http.ResponseWriter, status int, data any, meta *Meta) {
	resp := Response{
		Success: true,
		Data:    data,
		Meta:    meta,
	}
	JSON(w, status, resp)
}

// OK sends a 200 OK response
func OK(w http.ResponseWriter, data any) {
	Success(w, http.StatusOK, data)
}

// Created sends a 201 Created response
func Created(w http.ResponseWriter, data any) {
	Success(w, http.StatusCreated, data)
}

// NoContent sends a 204 No Content response
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// Err sends an error response from an AppError
func Err(w http.ResponseWriter, err *AppError) {
	resp := Response{
		Success: false,
		Error: &Error{
			Code:    err.Code,
			Message: err.Message,
			Details: err.Details,
		},
	}
	JSON(w, err.HTTPStatus, resp)
}

// HandleError handles any error and sends appropriate response
func HandleError(w http.ResponseWriter, err error) {
	if appErr, ok := AsAppError(err); ok {
		Err(w, appErr)
		return
	}
	log.Printf("Unexpected error: %v", err)
	Err(w, NewInternalError("An unexpected error occurred"))
}

// BadRequest sends a 400 Bad Request response
func BadRequest(w http.ResponseWriter, message string) {
	Err(w, NewBadRequestError(message))
}

// NotFound sends a 404 Not Found response
func NotFound(w http.ResponseWriter, message string) {
	Err(w, NewNotFoundError(message))
}

// InternalError sends a 500 Internal Server Error response
func InternalError(w http.ResponseWriter, message string) {
	Err(w, NewInternalError(message))
}

// ValidationErrorResponse sends a validation error response
func ValidationErrorResponse(w http.ResponseWriter, message string, details any) {
	Err(w, NewValidationError(message, details))
}

// NewPaginationMeta creates Meta for paginated responses
func NewPaginationMeta(page, perPage int, total int64) *Meta {
	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}
	return &Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	}
}
