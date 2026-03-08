package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"smlcloudplatform/internal/goapi/mypg"

	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
)

// StockReportBarcodesRequest - request สำหรับดึง barcodes ตาม item codes
type StockReportBarcodesRequest struct {
	ShopID    string   `json:"shop_id"`
	ItemCodes []string `json:"item_codes"`
}

// StockReportWarehousesRequest - request สำหรับดึง warehouses + locations ตาม barcodes
type StockReportWarehousesRequest struct {
	ShopID   string   `json:"shop_id"`
	Barcodes []string `json:"barcodes"`
}

// StockReportBarcodesHandler - ดึง barcodes จาก productbarcodeprocess ตาม item codes
// POST /api/stock-report/barcodes
func StockReportBarcodesHandler(c echo.Context) error {
	var req StockReportBarcodesRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload",
			"code":  "INVALID_PAYLOAD",
		})
	}

	if req.ShopID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "shop_id is required",
			"code":  "MISSING_SHOP_ID",
		})
	}

	if len(req.ItemCodes) == 0 {
		return c.JSON(http.StatusOK, map[string]any{
			"status":   "success",
			"barcodes": []string{},
		})
	}

	db, err := mypg.PgSqlFastConnect(req.ShopID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Database connection failed",
			"code":  "DB_CONNECTION_ERROR",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := `SELECT barcode FROM productbarcodeprocess WHERE shopid = $1 AND barcoderef = ANY($2) ORDER BY barcode`
	rows, err := db.QueryContext(ctx, query, req.ShopID, pq.Array(req.ItemCodes))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Query execution failed",
			"code":  "QUERY_ERROR",
		})
	}
	defer rows.Close()

	barcodes := make([]string, 0)
	for rows.Next() {
		var barcode string
		if err := rows.Scan(&barcode); err != nil {
			continue
		}
		barcodes = append(barcodes, barcode)
	}

	return c.JSON(http.StatusOK, map[string]any{
		"status":   "success",
		"barcodes": barcodes,
	})
}

// StockReportWarehousesHandler - ดึง warehouses + locations จาก processstockcost ตาม barcodes
// POST /api/stock-report/warehouses
func StockReportWarehousesHandler(c echo.Context) error {
	var req StockReportWarehousesRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload",
			"code":  "INVALID_PAYLOAD",
		})
	}

	if req.ShopID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "shop_id is required",
			"code":  "MISSING_SHOP_ID",
		})
	}

	db, err := mypg.PgSqlFastConnect(req.ShopID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Database connection failed",
			"code":  "DB_CONNECTION_ERROR",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Query warehouses and locations
	var query string
	var args []any

	if len(req.Barcodes) > 0 {
		query = `SELECT DISTINCT whcode, locationcode FROM processstockcost WHERE barcode = ANY($1) ORDER BY whcode, locationcode`
		args = []any{pq.Array(req.Barcodes)}
	} else {
		query = `SELECT DISTINCT whcode, locationcode FROM processstockcost ORDER BY whcode, locationcode`
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Query execution failed",
			"code":  "QUERY_ERROR",
		})
	}
	defer rows.Close()

	// Group locations by warehouse
	type warehouseData struct {
		WHCode    string   `json:"whcode"`
		Locations []string `json:"locations"`
	}

	whMap := make(map[string]*warehouseData)
	whOrder := make([]string, 0)

	for rows.Next() {
		var whcode, locationcode string
		if err := rows.Scan(&whcode, &locationcode); err != nil {
			continue
		}
		if _, exists := whMap[whcode]; !exists {
			whMap[whcode] = &warehouseData{WHCode: whcode, Locations: make([]string, 0)}
			whOrder = append(whOrder, whcode)
		}
		if locationcode != "" {
			whMap[whcode].Locations = append(whMap[whcode].Locations, locationcode)
		}
	}

	warehouses := make([]warehouseData, 0, len(whOrder))
	for _, code := range whOrder {
		warehouses = append(warehouses, *whMap[code])
	}

	return c.JSON(http.StatusOK, map[string]any{
		"status":     "success",
		"warehouses": warehouses,
	})
}

// StockReportItemBarcodesHandler - ดึง item codes + barcodes ทั้งหมด สำหรับ filter dropdown
// POST /api/stock-report/item-barcodes
func StockReportItemBarcodesHandler(c echo.Context) error {
	var req struct {
		ShopID string `json:"shop_id"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload",
			"code":  "INVALID_PAYLOAD",
		})
	}

	if req.ShopID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "shop_id is required",
			"code":  "MISSING_SHOP_ID",
		})
	}

	db, err := mypg.PgSqlFastConnect(req.ShopID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Database connection failed",
			"code":  "DB_CONNECTION_ERROR",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := fmt.Sprintf(`SELECT DISTINCT barcode, barcoderef FROM productbarcodeprocess WHERE shopid = $1 ORDER BY barcode LIMIT 10000`)
	rows, err := db.QueryContext(ctx, query, req.ShopID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Query execution failed",
			"code":  "QUERY_ERROR",
		})
	}
	defer rows.Close()

	type itemBarcode struct {
		Barcode    string `json:"barcode"`
		BarcodeRef string `json:"barcoderef"`
	}
	items := make([]itemBarcode, 0)
	for rows.Next() {
		var ib itemBarcode
		if err := rows.Scan(&ib.Barcode, &ib.BarcodeRef); err != nil {
			continue
		}
		items = append(items, ib)
	}

	return c.JSON(http.StatusOK, map[string]any{
		"status": "success",
		"items":  items,
	})
}
