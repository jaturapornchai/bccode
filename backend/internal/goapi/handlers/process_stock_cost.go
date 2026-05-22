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

// ProcessStockCostRequest - request payload for process stock cost queries
type ProcessStockCostRequest struct {
	ShopID string   `json:"shop_id"`
	FromDate string   `json:"from_date"`
	ToDate string   `json:"to_date"`
	BranchCodes []string `json:"branch_codes,omitempty"`
	ProductCodes []string `json:"product_codes,omitempty"`
	Limit int      `json:"limit,omitempty"`
	Offset int      `json:"offset,omitempty"`
}

// ProcessStockCostSummaryRequest - request for summary queries
type ProcessStockCostSummaryRequest struct {
	ShopID string   `json:"shop_id"`
	FromDate string   `json:"from_date"`
	ToDate string   `json:"to_date"`
	BranchCodes []string `json:"branch_codes,omitempty"`
	ProductCodes []string `json:"product_codes,omitempty"`
}

// ProcessStockCostHandler - secure handler for process stock cost queries
// Uses parameterized queries to prevent SQL injection
func ProcessStockCostHandler(c echo.Context) error {
	var req ProcessStockCostRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload",
			"code":  "INVALID_PAYLOAD",
		})
	}

	// Validate required fields
	if req.ShopID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "shop_id is required",
			"code":  "MISSING_SHOP_ID",
		})
	}

	if req.FromDate == "" || req.ToDate == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "from_date and to_date are required",
			"code":  "MISSING_DATE_RANGE",
		})
	}

	// Parse dates
	fromDate, err := time.Parse("2006-01-02", req.FromDate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid from_date format. Use YYYY-MM-DD",
			"code":  "INVALID_FROM_DATE",
		})
	}

	toDate, err := time.Parse("2006-01-02", req.ToDate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid to_date format. Use YYYY-MM-DD",
			"code":  "INVALID_TO_DATE",
		})
	}

	// Add 1 day to toDate to include the entire last day
	toDate = toDate.AddDate(0, 0, 1)

	// Set default limits
	if req.Limit <= 0 || req.Limit > 10000 {
		req.Limit = 1000
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	// Connect to database
	db, err := mypg.PgSqlFastConnect(req.ShopID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Database connection failed",
			"code":  "DB_CONNECTION_ERROR",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build parameterized query
	query, args := buildProcessStockCostQuery(fromDate, toDate, req.BranchCodes, req.ProductCodes, req.Limit, req.Offset)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Query execution failed",
			"code":  "QUERY_ERROR",
		})
	}
	defer rows.Close()

	// Scan results
	results := make([]map[string]any, 0)
	columns, err := rows.Columns()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get columns",
			"code":  "COLUMN_ERROR",
		})
	}

	for rows.Next() {
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}

		row := make(map[string]any)
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		results = append(results, row)
	}

	return c.JSON(http.StatusOK, map[string]any{
		"status": "success",
		"data":   results,
		"count":  len(results),
		"limit":  req.Limit,
		"offset": req.Offset,
	})
}

// ProcessStockCostSummaryHandler - secure handler for summary queries
func ProcessStockCostSummaryHandler(c echo.Context) error {
	var req ProcessStockCostSummaryRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload",
			"code":  "INVALID_PAYLOAD",
		})
	}

	// Validate required fields
	if req.ShopID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "shop_id is required",
			"code":  "MISSING_SHOP_ID",
		})
	}

	if req.FromDate == "" || req.ToDate == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "from_date and to_date are required",
			"code":  "MISSING_DATE_RANGE",
		})
	}

	// Parse dates
	fromDate, err := time.Parse("2006-01-02", req.FromDate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid from_date format. Use YYYY-MM-DD",
			"code":  "INVALID_FROM_DATE",
		})
	}

	toDate, err := time.Parse("2006-01-02", req.ToDate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid to_date format. Use YYYY-MM-DD",
			"code":  "INVALID_TO_DATE",
		})
	}

	toDate = toDate.AddDate(0, 0, 1)

	// Connect to database
	db, err := mypg.PgSqlFastConnect(req.ShopID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Database connection failed",
			"code":  "DB_CONNECTION_ERROR",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build parameterized summary query
	query, args := buildProcessStockCostSummaryQuery(fromDate, toDate, req.BranchCodes, req.ProductCodes)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Query execution failed",
			"code":  "QUERY_ERROR",
		})
	}
	defer rows.Close()

	// Scan results
	results := make([]map[string]any, 0)
	columns, err := rows.Columns()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get columns",
			"code":  "COLUMN_ERROR",
		})
	}

	for rows.Next() {
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}

		row := make(map[string]any)
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		results = append(results, row)
	}

	return c.JSON(http.StatusOK, map[string]any{
		"status": "success",
		"data":   results,
		"count":  len(results),
	})
}

// ProcessStockCostCheckHandler - check data availability
func ProcessStockCostCheckHandler(c echo.Context) error {
	var req ProcessStockCostSummaryRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload",
			"code":  "INVALID_PAYLOAD",
		})
	}

	if req.ShopID == "" || req.FromDate == "" || req.ToDate == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "shop_id, from_date and to_date are required",
			"code":  "MISSING_REQUIRED_FIELDS",
		})
	}

	fromDate, err := time.Parse("2006-01-02", req.FromDate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid from_date format",
			"code":  "INVALID_FROM_DATE",
		})
	}

	toDate, err := time.Parse("2006-01-02", req.ToDate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid to_date format",
			"code":  "INVALID_TO_DATE",
		})
	}

	toDate = toDate.AddDate(0, 0, 1)

	db, err := mypg.PgSqlFastConnect(req.ShopID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Database connection failed",
			"code":  "DB_CONNECTION_ERROR",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	query := `
		SELECT
			COUNT(*) as total_rows,
			COUNT(DISTINCT docno) as total_documents,
			COUNT(DISTINCT itemcode) as total_products,
			MIN(docdatetime) as earliest_date,
			MAX(docdatetime) as latest_date
		FROM public.processstockcost
		WHERE docdatetime >= $1 AND docdatetime < $2
	`

	var totalRows, totalDocs, totalProducts int64
	var earliestDate, latestDate *time.Time

	err = db.QueryRowContext(ctx, query, fromDate, toDate).Scan(
		&totalRows, &totalDocs, &totalProducts, &earliestDate, &latestDate,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Query execution failed",
			"code":  "QUERY_ERROR",
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"status":         "success",
		"total_rows":     totalRows,
		"total_documents": totalDocs,
		"total_products": totalProducts,
		"earliest_date":  earliestDate,
		"latest_date":    latestDate,
	})
}

// buildProcessStockCostQuery - builds parameterized query for process stock cost
func buildProcessStockCostQuery(fromDate, toDate time.Time, branchCodes, productCodes []string, limit, offset int) (string, []any) {
	args := make([]any, 0)
	argIndex := 1

	query := `
		SELECT
			id, docdatetime, docno, docref, linenumber, transflag,
			itemcode, itemname, barcode, unitcode, whcode, locationcode,
			debtorcode, debtorname, totalqty, unitstand, unitdivide,
			price, averagecost, calcamount, balanceqty, balanceamount,
			unitcost, guid
		FROM public.processstockcost
		WHERE docdatetime >= $1 AND docdatetime < $2
	`
	args = append(args, fromDate, toDate)
	argIndex = 3

	// Add branch filter
	if len(branchCodes) > 0 {
		query += fmt.Sprintf(" AND whcode = ANY($%d)", argIndex)
		args = append(args, pq.Array(branchCodes))
		argIndex++
	}

	// Add product filter
	if len(productCodes) > 0 {
		query += fmt.Sprintf(" AND itemcode = ANY($%d)", argIndex)
		args = append(args, pq.Array(productCodes))
		argIndex++
	}

	query += " ORDER BY docdatetime DESC, docno, linenumber"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	return query, args
}

// buildProcessStockCostSummaryQuery - builds parameterized summary query
func buildProcessStockCostSummaryQuery(fromDate, toDate time.Time, branchCodes, productCodes []string) (string, []any) {
	args := make([]any, 0)
	argIndex := 1

	query := `
		SELECT
			itemcode,
			SUM(totalqty) as total_quantity,
			SUM(calcamount) as total_amount,
			AVG(averagecost) as avg_cost,
			COUNT(*) as transaction_count
		FROM public.processstockcost
		WHERE docdatetime >= $1 AND docdatetime < $2
	`
	args = append(args, fromDate, toDate)
	argIndex = 3

	// Add branch filter
	if len(branchCodes) > 0 {
		query += fmt.Sprintf(" AND whcode = ANY($%d)", argIndex)
		args = append(args, pq.Array(branchCodes))
		argIndex++
	}

	// Add product filter
	if len(productCodes) > 0 {
		query += fmt.Sprintf(" AND itemcode = ANY($%d)", argIndex)
		args = append(args, pq.Array(productCodes))
	}

	query += " GROUP BY itemcode ORDER BY total_amount DESC LIMIT 10000"

	return query, args
}
