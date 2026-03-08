package handlers

import (
	"net/http"

	"smlcloudplatform/internal/goapi/logger"
	processstock "smlcloudplatform/internal/goapi/process/process-stock"

	"github.com/labstack/echo/v4"
)

// ProductBalanceUpdateRequest — request body
type ProductBalanceUpdateRequest struct {
	ShopID string `json:"shopid"`
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

	shopID := req.ShopID
	if shopID == "" {
		shopID = c.QueryParam("shopid")
	}
	if shopID == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Missing required parameter: shopid",
		})
	}

	logger.Info("ProductBalanceUpdateHandler: processing shopID=%s", shopID)

	if err := processstock.ProcessProductBalanceUpdate(shopID); err != nil {
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
