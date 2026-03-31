package common

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

// ValidationError represents a single field validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Validator holds validation errors
type Validator struct {
	Errors []ValidationError `json:"errors"`
}

// NewValidator creates a new Validator
func NewValidator() *Validator {
	return &Validator{
		Errors: make([]ValidationError, 0),
	}
}

// Valid returns true if there are no validation errors
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// AddError adds a validation error
func (v *Validator) AddError(field, message string) {
	v.Errors = append(v.Errors, ValidationError{
		Field:   field,
		Message: message,
	})
}

// Check adds an error if the condition is false
func (v *Validator) Check(ok bool, field, message string) {
	if !ok {
		v.AddError(field, message)
	}
}

// ToAppError converts validation errors to AppError
func (v *Validator) ToAppError() *AppError {
	return NewValidationError("Validation failed", v.Errors)
}

// Required checks if a string is not empty
func (v *Validator) Required(value, field string) {
	v.Check(strings.TrimSpace(value) != "", field, "This field is required")
}

// MinLength checks minimum string length
func (v *Validator) MinLength(value string, min int, field string) {
	v.Check(len(value) >= min, field, fmt.Sprintf("Must be at least %d characters", min))
}

// MaxLength checks maximum string length
func (v *Validator) MaxLength(value string, max int, field string) {
	v.Check(len(value) <= max, field, fmt.Sprintf("Must be at most %d characters", max))
}

// Email validates email format
func (v *Validator) Email(value, field string) {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	v.Check(emailRegex.MatchString(value), field, "Must be a valid email address")
}

// Phone validates phone number (10 digits)
func (v *Validator) Phone(value, field string) {
	phoneRegex := regexp.MustCompile(`^\d{10}$`)
	v.Check(phoneRegex.MatchString(value), field, "Must be a valid 10-digit phone number")
}

// UUID validates UUID format
func (v *Validator) UUID(value, field string) {
	_, err := uuid.Parse(value)
	v.Check(err == nil, field, "Must be a valid UUID")
}

// UUIDSlice validates a slice of UUIDs
func (v *Validator) UUIDSlice(values []string, field string) {
	for i, val := range values {
		_, err := uuid.Parse(val)
		v.Check(err == nil, field, fmt.Sprintf("Item %d must be a valid UUID", i+1))
		if err != nil {
			break
		}
	}
}

// MinSliceLength checks minimum slice length
func (v *Validator) MinSliceLength(values []string, min int, field string) {
	v.Check(len(values) >= min, field, fmt.Sprintf("Must have at least %d items", min))
}

// MaxSliceLength checks maximum slice length
func (v *Validator) MaxSliceLength(values []string, max int, field string) {
	v.Check(len(values) <= max, field, fmt.Sprintf("Must have at most %d items", max))
}

// Positive checks if a number is positive
func (v *Validator) Positive(value int, field string) {
	v.Check(value > 0, field, "Must be a positive number")
}

// InRange checks if a number is within a range
func (v *Validator) InRange(value, min, max int, field string) {
	v.Check(value >= min && value <= max, field, fmt.Sprintf("Must be between %d and %d", min, max))
}

// DecodeAndValidate decodes JSON body and runs validation
func DecodeAndValidate[T any](r *http.Request, validate func(*T, *Validator)) (*T, *AppError) {
	var payload T

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return nil, NewBadRequestError("Invalid JSON payload")
	}

	v := NewValidator()
	validate(&payload, v)

	if !v.Valid() {
		return nil, v.ToAppError()
	}

	return &payload, nil
}

// ParseUUID parses and validates a UUID string
func ParseUUID(value string) (uuid.UUID, *AppError) {
	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, NewBadRequestError("Invalid UUID format")
	}
	return id, nil
}

// ParseUUIDs parses and validates multiple UUID strings
func ParseUUIDs(values []string) ([]uuid.UUID, *AppError) {
	ids := make([]uuid.UUID, len(values))
	for i, val := range values {
		id, err := uuid.Parse(val)
		if err != nil {
			return nil, NewBadRequestError(fmt.Sprintf("Invalid UUID format at position %d", i+1))
		}
		ids[i] = id
	}
	return ids, nil
}
