// Package apperr provides centralized application error types for the BC Ai Account backend.
// It replaces ad-hoc string errors with structured, machine-readable error codes
// that map to HTTP statuses and support bilingual (TH/EN) messages.
package apperr

import (
	"errors"
	"fmt"
	"net/http"
)

// AppError is the centralized application error type.
// All service-layer errors should use this type so the HTTP middleware
// can produce consistent JSON error responses.
type AppError struct {
	// Code is a machine-readable error identifier (e.g. "NOT_FOUND", "DUPLICATE_CODE").
	Code string `json:"errorcode"`
	// Message is the human-readable English message.
	Message string `json:"message"`
	// ThaiMsg is the optional human-readable Thai message.
	ThaiMsg string `json:"message_th,omitempty"`
	// HTTPStatus is the HTTP status code this error maps to.
	HTTPStatus int `json:"-"`
	// Field optionally indicates which input field caused the error.
	Field string `json:"field,omitempty"`
	// Wrapped is the original underlying error for logging (not serialized to client).
	Wrapped error `json:"-"`
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Wrapped != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Wrapped)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap allows errors.Is / errors.As to inspect the wrapped error.
func (e *AppError) Unwrap() error {
	return e.Wrapped
}

// WithField returns a shallow copy with the Field set.
func (e *AppError) WithField(field string) *AppError {
	cp := *e
	cp.Field = field
	return &cp
}

// WithMessage returns a shallow copy with a custom message override.
func (e *AppError) WithMessage(msg string) *AppError {
	cp := *e
	cp.Message = msg
	return &cp
}

// WithThaiMessage returns a shallow copy with a custom Thai message.
func (e *AppError) WithThaiMessage(msg string) *AppError {
	cp := *e
	cp.ThaiMsg = msg
	return &cp
}

// WithWrap returns a shallow copy wrapping the given underlying error.
func (e *AppError) WithWrap(err error) *AppError {
	cp := *e
	cp.Wrapped = err
	return &cp
}

// Withf returns a shallow copy with a formatted message.
func (e *AppError) Withf(format string, args ...interface{}) *AppError {
	cp := *e
	cp.Message = fmt.Sprintf(format, args...)
	return &cp
}

// New creates a new AppError with the given code, HTTP status, and messages.
func New(code string, httpStatus int, message string, thaiMsg string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		ThaiMsg:    thaiMsg,
		HTTPStatus: httpStatus,
	}
}

// Newf creates a new AppError with a formatted message.
func Newf(code string, httpStatus int, thaiMsg string, format string, args ...interface{}) *AppError {
	return &AppError{
		Code:       code,
		Message:    fmt.Sprintf(format, args...),
		ThaiMsg:    thaiMsg,
		HTTPStatus: httpStatus,
	}
}

// FromError attempts to extract an *AppError from an error chain.
// Returns nil if the error is not an AppError.
func FromError(err error) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return nil
}

// Is checks whether the target error matches this AppError by code.
func (e *AppError) Is(target error) bool {
	var t *AppError
	if errors.As(target, &t) {
		return e.Code == t.Code
	}
	return false
}

// Response is the JSON structure returned to clients on error.
type Response struct {
	Success    bool   `json:"success"`
	ErrorCode  string `json:"errorcode"`
	Message    string `json:"message"`
	MessageTH  string `json:"message_th,omitempty"`
	Field      string `json:"field,omitempty"`
	StatusCode int    `json:"statuscode"`
}

// ToResponse converts an AppError to a client-facing JSON response struct.
func (e *AppError) ToResponse() Response {
	return Response{
		Success:    false,
		ErrorCode:  e.Code,
		Message:    e.Message,
		MessageTH:  e.ThaiMsg,
		Field:      e.Field,
		StatusCode: e.HTTPStatus,
	}
}

// StatusCode returns the HTTP status, defaulting to 500 if unset.
func (e *AppError) StatusCode() int {
	if e.HTTPStatus == 0 {
		return http.StatusInternalServerError
	}
	return e.HTTPStatus
}
