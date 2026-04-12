package aichat

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// ==================== Question History (No-Op) ====================
//
// กฎใหม่: backend ห้ามเขียน DB เลย → ลบ MongoDB-backed question history ออก
// คงไว้แค่ stub เพื่อให้ call sites เดิมไม่พัง และ HTTP endpoint return empty list
//
// ถ้าต้องการ history จริง ให้ frontend เก็บใน local storage เอง

// saveQuestionHistory — no-op (เดิมเขียน MongoDB, ตอนนี้ทิ้ง)
func saveQuestionHistory(_, _, _ string) {
	// intentionally empty: backend is read-only
}

// ListQuestionHistory — POST /api/v1/ai-provider/question-history
// คืน list ว่างเสมอ — frontend ต้องเก็บ history เองฝั่ง client
func ListQuestionHistory(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]any{
		"success":   true,
		"questions": []string{},
		"note":      "backend is read-only — store history client-side",
	})
}
