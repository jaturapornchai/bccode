package handlers

import (
	"fmt"
	"net/http"

	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/myclickhouse"

	"github.com/labstack/echo/v4"
)

// ClickHouseQueryHandler - Execute custom SELECT query (รองรับภาษาไทย)
func ClickHouseQueryHandler(c echo.Context) error {
	var reqBody struct {
		Query string `json:"query"` // SQL SELECT query
	}

	// รองรับ UTF-8 encoding
	if err := c.Bind(&reqBody); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"code":    400,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	// Validate query
	if reqBody.Query == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"code":    400,
			"message": "Query is required",
		})
	}

	logger.Info("ClickHouse query: %s", reqBody.Query)

	// Get ClickHouse connection
	conn, err := myclickhouse.ClickHouseFastConnect()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"code":    500,
			"message": "Failed to connect to ClickHouse",
			"error":   err.Error(),
		})
	}

	// Execute query with proper UTF-8 handling
	results, err := myclickhouse.QuerySelectAllWithUTF8(conn, reqBody.Query)
	if err != nil {
		logger.Error("ClickHouse query error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"code":    500,
			"message": "Failed to execute query",
			"error":   err.Error(),
			"query":   reqBody.Query,
		})
	}

	// Set UTF-8 content type
	c.Response().Header().Set(echo.HeaderContentType, "application/json; charset=utf-8")

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": "success",
		"code":   200,
		"count":  len(results),
		"data":   results,
	})
}

// ClickHouseSelectHandler - Simplified SELECT with parameters
func ClickHouseSelectHandler(c echo.Context) error {
	var reqBody struct {
		Table   string                 `json:"table"`   // Table name
		Columns []string               `json:"columns"` // Column names (empty = SELECT *)
		Where   map[string]interface{} `json:"where"`   // WHERE conditions
		OrderBy string                 `json:"orderby"`
		Limit   int                    `json:"limit"`
		Offset  int                    `json:"offset"`
	}

	if err := c.Bind(&reqBody); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"code":    400,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	// Validate table name
	if reqBody.Table == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"code":    400,
			"message": "Table name is required",
		})
	}

	// Build SELECT clause
	selectClause := "*"
	if len(reqBody.Columns) > 0 {
		selectClause = ""
		for i, col := range reqBody.Columns {
			if i > 0 {
				selectClause += ", "
			}
			selectClause += col
		}
	}

	// Build WHERE clause
	whereClause := ""
	if len(reqBody.Where) > 0 {
		whereClause = " WHERE "
		i := 0
		for key, value := range reqBody.Where {
			if i > 0 {
				whereClause += " AND "
			}
			switch v := value.(type) {
			case string:
				whereClause += key + " = '" + v + "'"
			case int, int64, float64:
				whereClause += key + " = " + toString(v)
			default:
				whereClause += key + " = '" + toString(v) + "'"
			}
			i++
		}
	}

	// Build ORDER BY clause
	orderByClause := ""
	if reqBody.OrderBy != "" {
		orderByClause = " ORDER BY " + reqBody.OrderBy
	}

	// Build LIMIT clause
	limitClause := ""
	if reqBody.Limit > 0 {
		limitClause = " LIMIT " + toString(reqBody.Limit)
		if reqBody.Offset > 0 {
			limitClause += " OFFSET " + toString(reqBody.Offset)
		}
	}

	// Build final query
	query := "SELECT " + selectClause + " FROM " + reqBody.Table + whereClause + orderByClause + limitClause

	logger.Info("ClickHouse SELECT query: %s", query)

	// Get ClickHouse connection
	conn, err := myclickhouse.ClickHouseFastConnect()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"code":    500,
			"message": "Failed to connect to ClickHouse",
			"error":   err.Error(),
		})
	}

	// Execute query
	results, err := myclickhouse.QuerySelectAll(conn, query)
	if err != nil {
		logger.Error("ClickHouse SELECT error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"code":    500,
			"message": "Failed to execute SELECT",
			"error":   err.Error(),
			"query":   query,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": "success",
		"code":   200,
		"count":  len(results),
		"data":   results,
		"query":  query,
	})
}

// ClickHouseMultiQueryHandler - Execute multiple queries at once
func ClickHouseMultiQueryHandler(c echo.Context) error {
	var reqBody struct {
		Queries []string `json:"queries"` // Array of SQL SELECT queries
	}

	// รองรับ UTF-8 encoding
	if err := c.Bind(&reqBody); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"code":    400,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	// Validate queries
	if len(reqBody.Queries) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"code":    400,
			"message": "Queries array is required and must not be empty",
		})
	}

	logger.Info("ClickHouse multi-query: executing %d queries", len(reqBody.Queries))

	// Get ClickHouse connection
	conn, err := myclickhouse.ClickHouseFastConnect()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"code":    500,
			"message": "Failed to connect to ClickHouse",
			"error":   err.Error(),
		})
	}

	// Execute all queries and collect results
	type QueryResult struct {
		Index  int                      `json:"index"`
		Query  string                   `json:"query"`
		Status string                   `json:"status"`
		Count  int                      `json:"count,omitempty"`
		Data   []map[string]interface{} `json:"data,omitempty"`
		Error  string                   `json:"error,omitempty"`
	}

	results := make([]QueryResult, len(reqBody.Queries))
	successCount := 0
	errorCount := 0

	for i, query := range reqBody.Queries {
		if query == "" {
			results[i] = QueryResult{
				Index:  i,
				Query:  query,
				Status: "error",
				Error:  "Empty query",
			}
			errorCount++
			continue
		}

		logger.Info("Executing query %d: %s", i+1, query)

		// Execute query with proper UTF-8 handling
		data, err := myclickhouse.QuerySelectAllWithUTF8(conn, query)
		if err != nil {
			logger.Error("Query %d error: %v", i+1, err)
			results[i] = QueryResult{
				Index:  i,
				Query:  query,
				Status: "error",
				Error:  err.Error(),
			}
			errorCount++
		} else {
			results[i] = QueryResult{
				Index:  i,
				Query:  query,
				Status: "success",
				Count:  len(data),
				Data:   data,
			}
			successCount++
		}
	}

	// Set UTF-8 content type
	c.Response().Header().Set(echo.HeaderContentType, "application/json; charset=utf-8")

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":        "success",
		"code":          200,
		"total_queries": len(reqBody.Queries),
		"success_count": successCount,
		"error_count":   errorCount,
		"results":       results,
	})
}

// Helper function to convert value to string
func toString(value interface{}) string {
	return fmt.Sprintf("%v", value)
}
