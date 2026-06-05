package handlers

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	mypg "smlcloudplatform/internal/goapi/mypg"

	"github.com/labstack/echo/v4"
)

// isValidDatabaseName checks that a database name contains only safe characters.
var validDBNamePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

func isValidDatabaseName(name string) bool {
	return validDBNamePattern.MatchString(name)
}

// Dangerous SQL patterns - expanded list for better security
var dangerousPatterns = []string{
	"drop table", "drop database", "drop schema", "drop index",
	"truncate table", "truncate ",
	"alter table", "alter database", "alter schema",
	"create table", "create database", "create schema",
	"grant ", "revoke ",
	"exec ", "execute ",
	"xp_", "sp_",
	"--", "/*", "*/",
	"char(", "nchar(",
	"varchar(", "nvarchar(",
	"cast(", "convert(",
	"waitfor ", "delay ",
	"shutdown", "kill ",
	"load_file", "into outfile", "into dumpfile",
	"information_schema", "pg_catalog", "pg_shadow",
	"union select", "union all select",
}

// Regex patterns for SQL injection detection
var sqlInjectionPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i);\s*drop\s+`),
	regexp.MustCompile(`(?i);\s*delete\s+`),
	regexp.MustCompile(`(?i);\s*update\s+`),
	regexp.MustCompile(`(?i);\s*insert\s+`),
	regexp.MustCompile(`(?i)'\s*or\s+'?\d*'?\s*=\s*'?\d*`),
	regexp.MustCompile(`(?i)'\s*or\s+.*--`),
	regexp.MustCompile(`(?i)1\s*=\s*1`),
	regexp.MustCompile(`(?i)'\s*;\s*--`),
}

// Allowed tables for SELECT queries (whitelist approach)
var allowedSelectTables = map[string]bool{
	"doc": true, "docdetail": true, "docref": true, "docpayment": true,
	"product": true, "productbarcode": true, "productunit": true,
	"customer": true, "supplier": true, "branch": true,
	"warehouse": true, "location": true, "category": true,
	"stockcard": true, "stockbalance": true,
	"employee": true, "salechannel": true,
}

// validateSelectQuery checks if a query is a valid SELECT query
func validateSelectQuery(query string) (bool, string) {
	queryLower := strings.ToLower(strings.TrimSpace(query))

	// Must start with SELECT
	if !strings.HasPrefix(queryLower, "select ") {
		return false, "Only SELECT queries are allowed"
	}

	// Check for dangerous patterns
	for _, pattern := range dangerousPatterns {
		if strings.Contains(queryLower, pattern) {
			return false, "Query contains forbidden pattern"
		}
	}

	// Check regex patterns for SQL injection
	for _, re := range sqlInjectionPatterns {
		if re.MatchString(query) {
			return false, "Potential SQL injection detected"
		}
	}

	// Check for multiple statements (semicolon injection)
	if strings.Count(query, ";") > 1 {
		return false, "Multiple statements not allowed"
	}

	return true, ""
}

// validateExecQuery checks if a query is valid for execution
func validateExecQuery(query string) (bool, string) {
	queryLower := strings.ToLower(strings.TrimSpace(query))

	// Check for dangerous patterns
	for _, pattern := range dangerousPatterns {
		if strings.Contains(queryLower, pattern) {
			return false, "Query contains forbidden pattern"
		}
	}

	// Check regex patterns for SQL injection
	for _, re := range sqlInjectionPatterns {
		if re.MatchString(query) {
			return false, "Potential SQL injection detected"
		}
	}

	// Only allow INSERT, UPDATE, DELETE
	allowedPrefixes := []string{"insert ", "update ", "delete "}
	isAllowed := false
	for _, prefix := range allowedPrefixes {
		if strings.HasPrefix(queryLower, prefix) {
			isAllowed = true
			break
		}
	}

	if !isAllowed {
		return false, "Only INSERT, UPDATE, DELETE queries are allowed"
	}

	// Check for multiple statements
	if strings.Count(query, ";") > 1 {
		return false, "Multiple statements not allowed"
	}

	return true, ""
}

// PostgreSQL Select Handler - supports multiple SELECT queries
func PgSelectHandler(c echo.Context) error {
	var payload struct {
		Database string   `json:"database"`
		Queries  []string `json:"queries"`
	}

	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid JSON payload",
			"code":  "INVALID_JSON",
		})
	}

	// Validate database parameter
	if payload.Database == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Missing required parameter: database",
			"code":  "MISSING_DATABASE",
		})
	}

	// Security: Database name validation (length + pattern)
	if len(payload.Database) < 2 || len(payload.Database) > 100 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid database name length",
			"code":  "INVALID_DATABASE",
		})
	}
	if !isValidDatabaseName(payload.Database) {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid database name: only alphanumeric, underscore, and hyphen allowed",
			"code":  "INVALID_DATABASE",
		})
	}

	var queries []string

	// Handle queries array
	if len(payload.Queries) > 0 {
		queries = payload.Queries
	} else {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Missing required parameter: queries",
			"code":  "MISSING_QUERIES",
		})
	}

	// Validate queries
	if len(queries) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "No queries provided",
			"code":  "EMPTY_QUERIES",
		})
	}

	// Limit number of queries to prevent abuse
	if len(queries) > 20 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Too many queries (max 20)",
			"code":  "TOO_MANY_QUERIES",
		})
	}

	// Connect to database
	db, err := mypg.PgSqlFastConnect(payload.Database)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Database connection failed",
			"code":  "DB_CONNECTION_ERROR",
		})
	}

	// Execute queries and collect results
	results := make([]map[string]any, len(queries))

	for i, query := range queries {
		// Enhanced query validation
		queryTrimmed := strings.TrimSpace(query)
		if queryTrimmed == "" {
			results[i] = map[string]any{
				"error":    "Empty query",
				"query_id": i,
				"rows":     []map[string]any{},
			}
			continue
		}

		// Validate query for SQL injection and dangerous patterns
		if valid, errMsg := validateSelectQuery(queryTrimmed); !valid {
			results[i] = map[string]any{
				"error":    errMsg,
				"query_id": i,
				"rows":     []map[string]any{},
			}
			continue
		}

		// Execute query with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		rows, err := db.QueryContext(ctx, query)
		if err != nil {
			results[i] = map[string]any{
				"error":    fmt.Sprintf("Query failed: %v", err),
				"query_id": i,
				"rows":     []map[string]any{},
			}
			continue
		}
		defer rows.Close()

		// Get column names
		columns, err := rows.Columns()
		if err != nil {
			results[i] = map[string]any{
				"error":    fmt.Sprintf("Column retrieval failed: %v", err),
				"query_id": i,
				"rows":     []map[string]any{},
			}
			continue
		}

		// Prepare scan destinations
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for j := range columns {
			valuePtrs[j] = &values[j]
		}

		// Collect rows
		queryRows := []map[string]any{}
		rowCount := 0
		maxRows := 5000 // Limit rows to prevent memory exhaustion

		for rows.Next() && rowCount < maxRows {
			if err := rows.Scan(valuePtrs...); err != nil {
				continue
			}

			row := make(map[string]any)
			for j, col := range columns {
				val := values[j]
				if b, ok := val.([]byte); ok {
					row[col] = string(b)
				} else {
					row[col] = val
				}
			}
			queryRows = append(queryRows, row)
			rowCount++
		}

		results[i] = map[string]any{
			"query_id": i,
			"rows":     queryRows,
			"count":    len(queryRows),
		}

		if rowCount >= maxRows {
			resultMap := results[i]
			resultMap["warning"] = "Result truncated to 5000 rows"
			results[i] = resultMap
		}
	}

	// Return results
	response := map[string]any{
		"status":   "success",
		"database": payload.Database,
		"results":  results,
		"total":    len(queries),
	}

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	return c.JSON(http.StatusOK, response)
}

// PostgreSQL Execute Handler - supports multiple execution queries
func PgExecHandler(c echo.Context) error {
	// รับ JSON payload จาก request body
	var payload struct {
		Database string   `json:"database"`
		Queries  []string `json:"queries"`
	}

	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid JSON payload",
			"code":  "INVALID_JSON",
		})
	}

	// Validate database parameter
	if payload.Database == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Missing required parameter: database",
			"code":  "MISSING_DATABASE",
		})
	}

	// Security: Database name validation (length + pattern)
	if len(payload.Database) < 2 || len(payload.Database) > 100 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid database name length",
			"code":  "INVALID_DATABASE",
		})
	}
	if !isValidDatabaseName(payload.Database) {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid database name: only alphanumeric, underscore, and hyphen allowed",
			"code":  "INVALID_DATABASE",
		})
	}

	var queries []string

	// Handle queries array
	if len(payload.Queries) > 0 {
		queries = payload.Queries
	} else {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Missing required parameter: queries",
			"code":  "MISSING_QUERIES",
		})
	}

	// Validate queries
	if len(queries) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "No queries provided",
			"code":  "EMPTY_QUERIES",
		})
	}

	// Limit number of queries to prevent abuse
	if len(queries) > 50 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Too many queries (max 50)",
			"code":  "TOO_MANY_QUERIES",
		})
	}

	// Connect to database
	db, err := mypg.PgSqlFastConnect(payload.Database)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Database connection failed",
			"code":  "DB_CONNECTION_ERROR",
		})
	}

	// Execute queries and collect results
	results := make([]map[string]any, len(queries))

	for i, query := range queries {
		// Enhanced query validation
		queryTrimmed := strings.TrimSpace(query)
		if queryTrimmed == "" {
			results[i] = map[string]any{
				"error":    "Empty query",
				"query_id": i,
				"affected": 0,
				"success":  false,
			}
			continue
		}

		// Validate query for SQL injection and dangerous patterns
		if valid, errMsg := validateExecQuery(queryTrimmed); !valid {
			results[i] = map[string]any{
				"error":    errMsg,
				"query_id": i,
				"affected": 0,
				"success":  false,
			}
			continue
		}

		// Execute query with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		// Use transaction for safety
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			results[i] = map[string]any{
				"error":    fmt.Sprintf("Transaction failed: %v", err),
				"query_id": i,
				"affected": 0,
				"success":  false,
			}
			continue
		}

		result, err := tx.ExecContext(ctx, query)
		if err != nil {
			tx.Rollback()
			results[i] = map[string]any{
				"error":    fmt.Sprintf("Query failed: %v", err),
				"query_id": i,
				"affected": 0,
				"success":  false,
			}
			continue
		}

		// Commit transaction
		if err = tx.Commit(); err != nil {
			results[i] = map[string]any{
				"error":    fmt.Sprintf("Commit failed: %v", err),
				"query_id": i,
				"affected": 0,
				"success":  false,
			}
			continue
		}

		// Get affected rows
		affected, _ := result.RowsAffected()

		// Get last insert ID (if applicable)
		lastInsertId, err := result.LastInsertId()
		if err != nil {
			// PostgreSQL doesn't support LastInsertId, so ignore the error
			lastInsertId = 0
		}

		results[i] = map[string]any{
			"query_id":       i,
			"affected":       affected,
			"last_insert_id": lastInsertId,
			"success":        true,
		}
	}

	// Return results
	response := map[string]any{
		"status":   "success",
		"database": payload.Database,
		"results":  results,
		"total":    len(queries),
	}

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	return c.JSON(http.StatusOK, response)
}
