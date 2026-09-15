package handlers

import (
	"net/http"

	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/mypg"
	"smlcloudplatform/internal/goapi/process/stockengine"

	"github.com/labstack/echo/v4"
)

// StockQueueStatusRequest ใช้ระบุขอบเขตที่ต้องการดูสถานะคิว
type StockQueueStatusRequest struct {
	HoldingCode  string `json:"holdingcode" query:"holdingcode"`
	BusinessCode string `json:"businesscode" query:"businesscode"`
}

// StockQueueStatusHandler คืนจำนวนงานคำนวณต้นทุนที่ค้างอยู่จริง
//
// จอเครื่องมือใช้ข้อมูลนี้แสดงความคืบหน้าตามจริง แทนแถบความคืบหน้าที่เดินเองโดยไม่ดูงานจริง
func StockQueueStatusHandler(c echo.Context) error {
	var req StockQueueStatusRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"code":    "invalid_request",
		})
	}

	holdingCode, businessCode, scopeErr := authenticatedCompanyContext(c, req.HoldingCode, req.BusinessCode)
	if scopeErr != nil {
		return c.JSON(scopeErr.Status, map[string]any{
			"success": false,
			"code":    scopeErr.Code,
			"message": scopeErr.Message,
		})
	}

	db, err := mypg.PgSqlFastConnect(holdingCode)
	if err != nil {
		logger.Error("StockQueueStatusHandler: connect %s: %v", holdingCode, err)
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"code":    "connection_error",
		})
	}

	// นับเฉพาะงานของบริษัทที่ผู้ใช้กำลังทำงานอยู่ ฐานข้อมูลหนึ่งฐานเก็บงานของหลายบริษัทรวมกัน
	status, err := stockengine.LoadQueueStatus(c.Request().Context(), db, businessCode)
	if err != nil {
		logger.Error("StockQueueStatusHandler: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"code":    "load_failed",
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"success":      true,
		"pending":      status.Pending,
		"processing":   status.Processing,
		"failing":      status.Failing,
		"oldestwait":   status.OldestWait,
		"workeractive": stockengine.Enabled(),
	})
}
