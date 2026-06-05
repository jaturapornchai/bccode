package handlers

import (
	"fmt"
	"net/http"
	"strings"

	mypg "smlcloudplatform/internal/goapi/mypg"

	"github.com/labstack/echo/v4"
)

// BalanceDetail - รายละเอียดยอดคงเหลือแยกตามคลัง/ที่เก็บ
type BalanceDetail struct {
	WHCode       string  `json:"whcode"`
	LocationCode string  `json:"locationcode"`
	Balance      float64 `json:"balance"`
}

// ProductBalance - ยอดคงเหลือของสินค้าพร้อมรายละเอียด
type ProductBalance struct {
	ItemCode     string          `json:"itemcode"`
	TotalBalance float64         `json:"totalbalance"`
	Details      []BalanceDetail `json:"details"`
}

// GetProductBalancesHandler - ดึงยอดคงเหลือของสินค้าหลายรายการพร้อม aggregate ใน Go
func GetProductBalancesHandler(c echo.Context) error {
	var payload struct {
		Database  string   `json:"database"`
		ItemCodes []string `json:"itemcodes"`
	}

	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"code":    400,
			"message": "Invalid JSON payload",
			"error":   err.Error(),
		})
	}

	// Validate inputs
	if payload.Database == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"code":    400,
			"message": "Missing required parameter: database",
		})
	}

	if len(payload.ItemCodes) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"code":    400,
			"message": "Missing required parameter: itemcodes",
		})
	}

	// Limit number of itemcodes to prevent abuse
	if len(payload.ItemCodes) > 500 {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"code":    400,
			"message": "Too many itemcodes (max 500)",
		})
	}

	// Connect to PostgreSQL
	db, err := mypg.PgSqlFastConnect(payload.Database)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"code":    500,
			"message": "Database connection failed",
			"error":   err.Error(),
		})
	}

	// Build IN clause for itemcodes
	fmt.Printf("[DEBUG] Received %d itemcodes from request\n", len(payload.ItemCodes))
	if len(payload.ItemCodes) > 0 {
		previewCount := 5
		if len(payload.ItemCodes) < 5 {
			previewCount = len(payload.ItemCodes)
		}
		fmt.Printf("[DEBUG] First %d itemcodes: %v\n", previewCount, payload.ItemCodes[:previewCount])
	}

	itemCodesQuoted := make([]string, len(payload.ItemCodes))
	for i, code := range payload.ItemCodes {
		// Escape single quotes
		code = strings.ReplaceAll(code, "'", "''")
		itemCodesQuoted[i] = fmt.Sprintf("'%s'", code)
	}
	itemCodesCondition := strings.Join(itemCodesQuoted, ",")

	fmt.Printf("[DEBUG] itemCodesCondition length: %d\n", len(itemCodesCondition))
	if len(itemCodesCondition) > 200 {
		fmt.Printf("[DEBUG] itemCodesCondition (first 200 chars): %s...\n", itemCodesCondition[:200])
	} else if len(itemCodesCondition) > 0 {
		fmt.Printf("[DEBUG] itemCodesCondition: %s\n", itemCodesCondition)
	}

	// Build query
	query := fmt.Sprintf(`
		SELECT
			COALESCE(itemcode, '') as itemcode,
			COALESCE(whcode, '') as whcode,
			COALESCE(locationcode, '') as locationcode,
			SUM((totalqty * calcflag) * unitstand / unitdivide) as balance
		FROM docdetail
		WHERE iscalcstock = 1
			AND itemcode IN (%s)
		GROUP BY itemcode, whcode, locationcode
		HAVING SUM((totalqty * calcflag) * unitstand / unitdivide) != 0
		ORDER BY itemcode, whcode, locationcode
	`, itemCodesCondition)

	// Execute query
	fmt.Printf("[DEBUG] Executing query on database '%s'\n", payload.Database)
	fmt.Printf("[DEBUG] Query length: %d characters\n", len(query))
	if len(query) > 500 {
		fmt.Printf("[DEBUG] Query (first 500 chars): %s...\n", query[:500])
	} else {
		fmt.Printf("[DEBUG] Full Query: %s\n", query)
	}

	rows, err := db.Query(query)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"code":    500,
			"message": "Query execution failed",
			"error":   err.Error(),
			"query":   query,
		})
	}
	defer rows.Close()

	// Aggregate data in memory
	balanceMap := make(map[string]*ProductBalance)
	rowCount := 0

	for rows.Next() {
		var itemCode, whCode, locationCode string
		var balance float64

		if err := rows.Scan(&itemCode, &whCode, &locationCode, &balance); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"status":  "error",
				"code":    500,
				"message": "Failed to scan row",
				"error":   err.Error(),
			})
		}

		rowCount++
		fmt.Printf("[DEBUG] Row %d: itemcode=%s, whcode=%s, locationcode=%s, balance=%.2f\n",
			rowCount, itemCode, whCode, locationCode, balance)

		// Initialize ProductBalance if not exists
		if _, exists := balanceMap[itemCode]; !exists {
			balanceMap[itemCode] = &ProductBalance{
				ItemCode:     itemCode,
				TotalBalance: 0.0,
				Details:      []BalanceDetail{},
			}
		}

		// Add detail
		balanceMap[itemCode].Details = append(balanceMap[itemCode].Details, BalanceDetail{
			WHCode:       whCode,
			LocationCode: locationCode,
			Balance:      balance,
		})

		// Add to total
		balanceMap[itemCode].TotalBalance += balance
	}

	fmt.Printf("[DEBUG] Total rows returned: %d, Unique items: %d\n", rowCount, len(balanceMap))

	// Check for errors during iteration
	if err := rows.Err(); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"code":    500,
			"message": "Error iterating rows",
			"error":   err.Error(),
		})
	}

	// Return aggregated data
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": "success",
		"code":   200,
		"count":  len(balanceMap),
		"data":   balanceMap,
	})
}
