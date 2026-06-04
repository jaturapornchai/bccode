package handlers

import (
	"net/http"

	"smlcloudplatform/internal/goapi/logger"
	processstock "smlcloudplatform/internal/goapi/process/process-stock"

	"github.com/labstack/echo/v4"
)

// ProductBalanceUpdateRequest — request body
type ProductBalanceUpdateRequest struct {
	HoldingCode string `json:"holding_code"`
}

// ProductBalanceUpdateHandler — POST /api/process/product-balance
// คำนวณยอดคงเหลือทั้งหมด แล้ว UPDATE ลง product.balanceqty + product.balanceqtyword
func ProductBalanceUpdateHandler(c echo.Context) error {
	var req ProductBalanceUpdateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Invalid request body",
		})
	}

	holdingCode := req.HoldingCode
	if holdingCode == "" {
		holdingCode = c.QueryParam("holding_code")
	}
	if holdingCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Missing required parameter: holding_code",
		})
	}

	logger.Info("ProductBalanceUpdateHandler: processing holdingCode=%s", holdingCode)

	if err := processstock.ProcessProductBalanceUpdate(holdingCode); err != nil {
		logger.Error("ProductBalanceUpdateHandler: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Product balance updated successfully",
	})
}
