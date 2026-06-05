package inventory

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	m "smlcloudplatform/internal/goapi/inventory/models"
	"smlcloudplatform/internal/goapi/logger"
	myPg "smlcloudplatform/internal/goapi/mypg"
)

// RegisterRoutes — ลงทะเบียน routes กับ Echo group
// ใช้ per-request DB connection ตาม holdingCode (pattern เดียวกับ handlers อื่นๆ)
func RegisterRoutes(g *echo.Group) {
	// Costing Config
	g.GET("/products/:itemcode/costing-config", GetCostingConfig)
	g.PUT("/products/:itemcode/costing-config", UpdateCostingConfig)

	// Cost Layers & Lots
	g.GET("/products/:itemcode/cost-layers", GetCostLayers)

	// Transactions
	g.POST("/inventory/receipt", ProcessReceipt)
	g.POST("/inventory/issue", ProcessIssue)
	g.POST("/inventory/transfer", ProcessTransfer)
	g.POST("/inventory/adjustment", ProcessAdjustment)
	g.POST("/inventory/sales-return", ProcessSalesReturn)
	g.POST("/inventory/purchase-return", ProcessPurchaseReturn)

	// Reports
	g.GET("/reports/inventory-valuation", GetInventoryValuation)
	g.GET("/reports/stock-card/:itemcode", GetStockCard)

	// Database
	g.POST("/inventory/create-tables", CreateTablesHandler)
}

// connectAndService — สร้าง DB connection + service จาก holdingCode
func connectAndService(holdingCode string) (*InventoryCostingService, error) {
	db, err := myPg.PgSqlFastConnect(holdingCode)
	if err != nil {
		return nil, err
	}
	return NewInventoryCostingService(db), nil
}

// === Config Handlers ===

func GetCostingConfig(c echo.Context) error {
	holdingCode := c.QueryParam("holdingcode")
	itemCode := c.Param("itemcode")
	if holdingCode == "" || itemCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "ต้องระบุ holdingcode และ itemcode"})
	}

	svc, err := connectAndService(holdingCode)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	config, err := svc.GetCostingConfig(c.Request().Context(), holdingCode, itemCode)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, config)
}

func UpdateCostingConfig(c echo.Context) error {
	holdingCode := c.QueryParam("holdingcode")
	itemCode := c.Param("itemcode")

	var config m.ProductCostingConfig
	if err := c.Bind(&config); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "ข้อมูลไม่ถูกต้อง"})
	}
	config.ItemCode = itemCode

	svc, err := connectAndService(holdingCode)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	if err := svc.UpdateCostingConfig(c.Request().Context(), holdingCode, &config); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	logger.Info("อัพเดท costing config: " + itemCode + " → " + config.CostingMethod)
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

// === Transaction Handlers ===

func ProcessReceipt(c echo.Context) error {
	var params m.ReceiptParams
	if err := c.Bind(&params); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "ข้อมูลไม่ถูกต้อง"})
	}
	if params.HoldingCode == "" || params.ItemCode == "" || params.Qty <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "ต้องระบุ holdingcode, itemcode, qty > 0"})
	}

	svc, err := connectAndService(params.HoldingCode)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	result, err := svc.ProcessReceipt(c.Request().Context(), params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, result)
}

func ProcessIssue(c echo.Context) error {
	var params m.IssueParams
	if err := c.Bind(&params); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "ข้อมูลไม่ถูกต้อง"})
	}

	svc, err := connectAndService(params.HoldingCode)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	result, err := svc.ProcessIssue(c.Request().Context(), params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, result)
}

func ProcessTransfer(c echo.Context) error {
	var params m.TransferParams
	if err := c.Bind(&params); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "ข้อมูลไม่ถูกต้อง"})
	}

	svc, err := connectAndService(params.HoldingCode)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	result, err := svc.ProcessTransfer(c.Request().Context(), params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, result)
}

func ProcessAdjustment(c echo.Context) error {
	var params m.AdjustmentParams
	if err := c.Bind(&params); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "ข้อมูลไม่ถูกต้อง"})
	}

	svc, err := connectAndService(params.HoldingCode)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	result, err := svc.ProcessAdjustment(c.Request().Context(), params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, result)
}

func ProcessSalesReturn(c echo.Context) error {
	var params m.SalesReturnParams
	if err := c.Bind(&params); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "ข้อมูลไม่ถูกต้อง"})
	}

	svc, err := connectAndService(params.HoldingCode)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	result, err := svc.ProcessSalesReturn(c.Request().Context(), params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, result)
}

func ProcessPurchaseReturn(c echo.Context) error {
	var params m.PurchaseReturnParams
	if err := c.Bind(&params); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "ข้อมูลไม่ถูกต้อง"})
	}

	svc, err := connectAndService(params.HoldingCode)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	result, err := svc.ProcessPurchaseReturn(c.Request().Context(), params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, result)
}

// === Report Handlers ===

func GetInventoryValuation(c echo.Context) error {
	holdingCode := c.QueryParam("holdingcode")
	if holdingCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "ต้องระบุ holdingcode"})
	}

	svc, err := connectAndService(holdingCode)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	report, err := svc.GetInventoryValuation(c.Request().Context(), holdingCode)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, report)
}

func GetStockCard(c echo.Context) error {
	holdingCode := c.QueryParam("holdingcode")
	itemCode := c.Param("itemcode")
	fromDate := c.QueryParam("from")
	toDate := c.QueryParam("to")

	if holdingCode == "" || itemCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "ต้องระบุ holdingcode และ itemcode"})
	}

	from, err := time.Parse("2006-01-02", fromDate)
	if err != nil {
		from = time.Now().AddDate(0, -1, 0)
	}
	to, err := time.Parse("2006-01-02", toDate)
	if err != nil {
		to = time.Now()
	}

	svc, svcErr := connectAndService(holdingCode)
	if svcErr != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": svcErr.Error()})
	}

	report, err := svc.GetStockCard(c.Request().Context(), holdingCode, itemCode, from, to)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, report)
}

func GetCostLayers(c echo.Context) error {
	holdingCode := c.QueryParam("holdingcode")
	itemCode := c.Param("itemcode")
	whCode := c.QueryParam("whcode")

	if holdingCode == "" || itemCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "ต้องระบุ holdingcode และ itemcode"})
	}

	svc, err := connectAndService(holdingCode)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	layers, err := svc.GetCostLayers(c.Request().Context(), holdingCode, itemCode, whCode)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, layers)
}

// === Database Handler ===

func CreateTablesHandler(c echo.Context) error {
	holdingCode := c.QueryParam("holdingcode")
	if holdingCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "ต้องระบุ holdingcode"})
	}

	db, err := myPg.PgSqlFastConnect(holdingCode)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	if err := CreateInventoryCostingTables(db); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "ok", "message": "สร้างตารางสำเร็จ"})
}
