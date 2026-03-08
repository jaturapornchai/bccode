package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	mypg "smlcloudplatform/internal/goapi/mypg"

	"github.com/labstack/echo/v4"
)

// PurchaseHistoryRequest - request สำหรับดึงประวัติการสั่งซื้อ
type PurchaseHistoryRequest struct {
	ShopID   string   `json:"shop_id"`
	Barcodes []string `json:"barcodes"`
	Months   int      `json:"months"` // จำนวนเดือนย้อนหลัง (default 3)
}

// PurchaseHistoryItem - รายการประวัติการสั่งซื้อ
type PurchaseHistoryItem struct {
	Barcode      string  `json:"barcode"`
	ItemCode     string  `json:"itemcode"`
	DocNo        string  `json:"docno"`
	DocDate      string  `json:"docdate"`
	CustCode     string  `json:"custcode"`     // รหัสผู้ขาย/เจ้าหนี้
	TotalQty     float64 `json:"totalqty"`
	UnitCode     string  `json:"unitcode"`
	Price        float64 `json:"price"`
	SumAmount    float64 `json:"sumamount"`
	TransFlag    int     `json:"transflag"`
}

// PurchaseHistoryHandler - ดึงประวัติการสั่งซื้อสินค้า
// ใช้สำหรับ AI วิเคราะห์เปรียบเทียบราคา
func PurchaseHistoryHandler(c echo.Context) error {
	var req PurchaseHistoryRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Invalid request payload",
			"code":    "INVALID_PAYLOAD",
		})
	}

	// Validate required fields
	if req.ShopID == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Missing required parameter: shop_id",
			"code":    "MISSING_SHOP_ID",
		})
	}

	if len(req.Barcodes) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Missing required parameter: barcodes",
			"code":    "MISSING_BARCODES",
		})
	}

	// Default months
	if req.Months <= 0 {
		req.Months = 3
	}
	if req.Months > 12 {
		req.Months = 12 // Maximum 12 months
	}

	// Limit barcodes to prevent abuse
	if len(req.Barcodes) > 100 {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Too many barcodes (max 100)",
			"code":    "TOO_MANY_BARCODES",
		})
	}

	// Connect to database (shop_id is used as database name)
	db, err := mypg.PgSqlFastConnect(req.ShopID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   "Database connection failed",
			"code":    "DB_CONNECTION_ERROR",
		})
	}

	// Build query with parameterized placeholders for barcodes
	// Query doc (เอกสาร) join docdetail (รายละเอียดสินค้า)
	// transflag: 12=ซื้อ, 16=ซื้อคืน, 20=สั่งซื้อ
	placeholders := make([]string, len(req.Barcodes))
	args := make([]interface{}, len(req.Barcodes)+2)

	for i, barcode := range req.Barcodes {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = barcode
	}

	// Date range: months ago from now
	dateFrom := time.Now().AddDate(0, -req.Months, 0)
	args[len(req.Barcodes)] = dateFrom.Format("2006-01-02")
	args[len(req.Barcodes)+1] = 12 // transflag for purchase

	query := fmt.Sprintf(`
		SELECT
			dd.barcode,
			dd.itemcode,
			d.docno,
			d.docdatetime::date as docdate,
			d.custcode,
			dd.totalqty,
			dd.unitcode,
			dd.price,
			dd.sumamount,
			d.transflag
		FROM docdetail dd
		INNER JOIN doc d ON dd.docno = d.docno AND dd.transflag = d.transflag
		WHERE dd.barcode IN (%s)
		AND d.docdatetime >= $%d
		AND d.transflag = $%d
		AND d.iscancel = false
		AND (d.isdelete = false OR d.isdelete IS NULL)
		ORDER BY dd.barcode, d.docdatetime DESC
		LIMIT 500
	`, strings.Join(placeholders, ","), len(req.Barcodes)+1, len(req.Barcodes)+2)

	// Execute query with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Query failed: %v", err),
			"code":    "QUERY_ERROR",
		})
	}
	defer rows.Close()

	// Collect results
	results := []PurchaseHistoryItem{}
	for rows.Next() {
		var item PurchaseHistoryItem
		var docDate time.Time

		err := rows.Scan(
			&item.Barcode,
			&item.ItemCode,
			&item.DocNo,
			&docDate,
			&item.CustCode,
			&item.TotalQty,
			&item.UnitCode,
			&item.Price,
			&item.SumAmount,
			&item.TransFlag,
		)
		if err != nil {
			continue // Skip row on error
		}

		item.DocDate = docDate.Format("2006-01-02")
		results = append(results, item)
	}

	// Group by barcode for easier analysis
	groupedResults := make(map[string][]PurchaseHistoryItem)
	for _, item := range results {
		groupedResults[item.Barcode] = append(groupedResults[item.Barcode], item)
	}

	// Calculate statistics per barcode
	stats := make(map[string]map[string]interface{})
	for barcode, items := range groupedResults {
		if len(items) == 0 {
			continue
		}

		var totalQty, totalAmount float64
		var minPrice, maxPrice, sumPrice float64
		minPrice = items[0].Price
		maxPrice = items[0].Price

		for i, item := range items {
			totalQty += item.TotalQty
			totalAmount += item.SumAmount
			sumPrice += item.Price
			if item.Price < minPrice {
				minPrice = item.Price
			}
			if item.Price > maxPrice {
				maxPrice = item.Price
			}
			// Keep only last 5 purchases per barcode
			if i >= 5 {
				groupedResults[barcode] = items[:5]
				break
			}
		}

		avgPrice := sumPrice / float64(len(items))

		stats[barcode] = map[string]interface{}{
			"count":        len(items),
			"total_qty":    totalQty,
			"total_amount": totalAmount,
			"avg_price":    roundTo2Decimals(avgPrice),
			"min_price":    minPrice,
			"max_price":    maxPrice,
			"last_price":   items[0].Price,
			"last_date":    items[0].DocDate,
			"last_vendor":  items[0].CustCode, // ใช้ custcode แทน (ไม่มี custname ใน schema)
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    results,
		"grouped": groupedResults,
		"stats":   stats,
		"total":   len(results),
		"months":  req.Months,
	})
}
