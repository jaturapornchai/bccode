package cheque

import (
	"errors"
	"net/http"

	"github.com/lib/pq"
)

// Error คือความผิดที่ผู้ใช้แก้ได้: Key อยู่ใน languages.tsv (ตอบเป็น errorcode และแปลตามภาษาของคำขอ),
// Field ชี้ช่องที่ต้องแก้, Args แทนค่า {0}, {1}, … ในข้อความ
type Error struct {
	Key    string
	Field  string
	Status int
	Args   []string
}

func (e *Error) Error() string { return e.Key }

func newError(status int, key, field string, args ...string) *Error {
	return &Error{Key: key, Field: field, Status: status, Args: args}
}

func badRequest(key, field string, args ...string) *Error {
	return newError(http.StatusBadRequest, key, field, args...)
}

func notFound(key, field string, args ...string) *Error {
	return newError(http.StatusNotFound, key, field, args...)
}

// versionConflict: แถวถูกคนอื่นแก้ไปแล้วหลังจากที่ผู้ใช้เปิดอ่าน (optimistic lock)
func versionConflict() *Error {
	return newError(http.StatusConflict, "cheque_err_version_conflict", "version")
}

// AsError คืน *Error ถ้า err เป็นความผิดของผู้ใช้
func AsError(err error) (*Error, bool) {
	var e *Error
	if errors.As(err, &e) {
		return e, true
	}
	return nil, false
}

// isUniqueViolation: PostgreSQL 23505 (ชนคีย์หลัก/unique index)
func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}
