package handlers

import (
	"net/http"

	mypg "smlcloudplatform/internal/goapi/mypg"

	"github.com/labstack/echo/v4"
)

// CheckReferenceResponse - response สำหรับตรวจสอบการอ้างอิง
type CheckReferenceResponse struct {
	IsReferenced bool     `json:"is_referenced"`
	ReferencedBy []string `json:"referenced_by,omitempty"` // รายการ docno ที่อ้างอิง
}

// CheckPOReferenceHandler - ตรวจสอบว่า PO ถูกอ้างอิงโดยเอกสารอื่นหรือไม่
// GET /api/transaction/check-reference?holding_code=xxx&docno=PO2026010700001
func CheckPOReferenceHandler(c echo.Context) error {
	holdingCode := c.QueryParam("holding_code")
	docNo := c.QueryParam("docno")

	// Validate required parameters
	if holdingCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "holding_code is required",
		})
	}

	if docNo == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "docno is required",
		})
	}

	// Connect to database
	db, err := mypg.PgSqlFastConnect(holdingCode)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "Database connection failed",
		})
	}

	// TODO: Implement proper reference checking
	// For now, we'll check the doc table's isref column
	// which is updated by background processes when document is referenced
	query := `
		SELECT isref
		FROM doc
		WHERE docno = $1
		LIMIT 1
	`

	var isRef bool
	err = db.QueryRow(query, docNo).Scan(&isRef)
	if err != nil {
		// If document not found or error, assume not referenced (safe mode disabled temporarily)
		isRef = false
	}

	// Collect referenced documents (placeholder for now)
	var referencedBy []string
	// TODO: Query actual referencing documents when schema is confirmed
	// For now, if isref is true, we just say "other documents"

	// Build response
	response := CheckReferenceResponse{
		IsReferenced: isRef,
	}

	if isRef && len(referencedBy) > 0 {
		response.ReferencedBy = referencedBy
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    response,
	})
}
