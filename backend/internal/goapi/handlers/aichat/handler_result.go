package aichat

// Result fetch endpoint — frontend ดึง full tool result จาก result_store
//
// Endpoint: GET /goapi/api/aichat/result/:id
// Response: { "id": "...", "tool": "...", "row_count": N, "data": <full data>, "stored_at": "..." }
// 404 ถ้าไม่เจอหรือหมด TTL

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// GetAIChatResult handles GET /goapi/api/aichat/result/:id
func GetAIChatResult(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "id is required",
		})
	}

	entry, ok := GetResultMeta(id)
	if !ok {
		return c.JSON(http.StatusNotFound, map[string]any{
			"success": false,
			"error":   "result not found or expired",
			"id":      id,
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"success":   true,
		"id":        entry.id,
		"tool":      entry.tool,
		"row_count": entry.rowCount,
		"stored_at": entry.storedAt,
		"data":      entry.data,
	})
}
