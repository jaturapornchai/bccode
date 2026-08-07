package apperr

import (
	"fmt"
	"net/http"
)

// Predefined sentinel errors for common cases.
// Use these as base errors and customize with .WithField(), .Withf(), .WithWrap() etc.
var (
	// ErrNotFound indicates a requested record/resource does not exist.
	ErrNotFound = New("NOT_FOUND", http.StatusNotFound, "record not found", "ไม่พบข้อมูล")

	// ErrDuplicate indicates a unique constraint violation (code already exists).
	ErrDuplicate = New("DUPLICATE", http.StatusConflict, "code already exists", "รหัสซ้ำ")

	// ErrValidation indicates input validation failure.
	ErrValidation = New("VALIDATION_FAILED", http.StatusBadRequest, "validation failed", "ข้อมูลไม่ถูกต้อง")

	// ErrUnauthorized indicates missing or invalid authentication.
	ErrUnauthorized = New("UNAUTHORIZED", http.StatusUnauthorized, "unauthorized", "ไม่ได้รับอนุญาต")

	// ErrForbidden indicates authenticated but lacking permission.
	ErrForbidden = New("FORBIDDEN", http.StatusForbidden, "forbidden", "ไม่มีสิทธิ์")

	// ErrInternal indicates an unexpected server-side error.
	ErrInternal = New("INTERNAL_ERROR", http.StatusInternalServerError, "internal server error", "เกิดข้อผิดพลาดในระบบ")

	// ErrBadRequest indicates a malformed request payload.
	ErrBadRequest = New("BAD_REQUEST", http.StatusBadRequest, "invalid request payload", "คำขอไม่ถูกต้อง")

	// ErrConflict indicates a state conflict (e.g. record in use, cannot delete).
	ErrConflict = New("CONFLICT", http.StatusConflict, "operation conflicts with current state", "ไม่สามารถดำเนินการได้เนื่องจากสถานะปัจจุบัน")

	// ErrDependencyFailed indicates an external service or database call failed.
	ErrDependencyFailed = New("DEPENDENCY_FAILED", http.StatusBadGateway, "upstream service error", "บริการภายนอกเกิดข้อผิดพลาด")

	// ErrRateLimited indicates too many requests.
	ErrRateLimited = New("RATE_LIMITED", http.StatusTooManyRequests, "rate limit exceeded", "ส่งคำขอถี่เกินไป")

	// ErrDisabled indicates the resource/account is disabled.
	ErrDisabled = New("DISABLED", http.StatusForbidden, "resource is disabled", "ถูกระงับการใช้งาน")

	// ErrExpired indicates the resource/token has expired.
	ErrExpired = New("EXPIRED", http.StatusGone, "resource has expired", "หมดอายุแล้ว")
)

// Domain-specific error constructors for common patterns.

// NotFound creates a NOT_FOUND error for a specific entity type.
func NotFound(entity string) *AppError {
	return &AppError{
		Code:       "NOT_FOUND",
		Message:    entity + " not found",
		ThaiMsg:    "ไม่พบ" + entity,
		HTTPStatus: http.StatusNotFound,
	}
}

// DuplicateCode creates a DUPLICATE error for a specific code field.
func DuplicateCode(field, code string) *AppError {
	return &AppError{
		Code:       "DUPLICATE",
		Message:    fmt.Sprintf("%s '%s' already exists", field, code),
		ThaiMsg:    fmt.Sprintf("%s '%s' ซ้ำ", field, code),
		HTTPStatus: http.StatusConflict,
		Field:      field,
	}
}

// Validation creates a VALIDATION_FAILED error for a specific field.
func Validation(field, reason string) *AppError {
	return &AppError{
		Code:       "VALIDATION_FAILED",
		Message:    reason,
		ThaiMsg:    reason,
		HTTPStatus: http.StatusBadRequest,
		Field:      field,
	}
}

// InternalWrap wraps an unexpected error as INTERNAL_ERROR.
func InternalWrap(err error) *AppError {
	return &AppError{
		Code:       "INTERNAL_ERROR",
		Message:    "internal server error",
		ThaiMsg:    "เกิดข้อผิดพลาดในระบบ",
		HTTPStatus: http.StatusInternalServerError,
		Wrapped:    err,
	}
}
