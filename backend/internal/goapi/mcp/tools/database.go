package tools

import (
	"context"
	"fmt"
	"strings"
	"time"

	"smlcloudplatform/internal/goapi/logger"
	mypg "smlcloudplatform/internal/goapi/mypg"
)

// ==================== Database Schema ====================

type DatabaseSchemaRequest struct {
	ShopID string `json:"shop_id"`
	TableName string `json:"table_name"` // Optional - specific table
}

type DatabaseSchemaResponse struct {
	DatabaseName string        `json:"database_name"`
	Tables []TableSchema `json:"tables"`
	TableCount int           `json:"table_count"`
	GeneratedAt time.Time     `json:"generated_at"`
}

type TableSchema struct {
	TableName string         `json:"table_name"`
	TableType string         `json:"table_type"` // BASE TABLE, VIEW
	Columns []ColumnSchema `json:"columns"`
	ColumnCount int            `json:"column_count"`
	RowCount int64          `json:"row_count_estimate"`
	Description string         `json:"description,omitempty"`
}

type ColumnSchema struct {
	ColumnName string `json:"column_name"`
	DataType string `json:"data_type"`
	IsNullable string `json:"is_nullable"`
	ColumnDefault string `json:"column_default,omitempty"`
	MaxLength int    `json:"max_length,omitempty"`
	IsPrimaryKey bool   `json:"is_primary_key"`
	IsForeignKey bool   `json:"is_foreign_key"`
	Description string `json:"description,omitempty"`
}

// GetDatabaseSchema returns the database schema information
func GetDatabaseSchema(ctx context.Context, shopID, tableName string) (*DatabaseSchemaResponse, error) {
	if shopID == "" {
		return nil, fmt.Errorf("shop_id is required")
	}

	logger.Info("[Database Schema] shopid=%s, table=%s", shopID, tableName)

	db, err := mypg.PgSqlFastConnect(shopID)
	if err != nil {
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	response := &DatabaseSchemaResponse{
		DatabaseName: shopID,
		Tables:       []TableSchema{},
		GeneratedAt:  time.Now(),
	}

	// Build query for tables
	tableQuery := `
		SELECT
			table_name,
			table_type
		FROM information_schema.tables
		WHERE table_schema = 'public'
	`
	args := []interface{}{}

	if tableName != "" {
		tableQuery += " AND table_name ILIKE $1"
		args = append(args, "%"+tableName+"%")
	}

	tableQuery += " ORDER BY table_name"

	rows, err := db.Query(tableQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get tables: %w", err)
	}
	defer rows.Close()

	var tableNames []string
	tableTypes := make(map[string]string)

	for rows.Next() {
		var tblName, tblType string
		if err := rows.Scan(&tblName, &tblType); err != nil {
			continue
		}
		tableNames = append(tableNames, tblName)
		tableTypes[tblName] = tblType
	}

	// Get columns for each table
	for _, tbl := range tableNames {
		schema := TableSchema{
			TableName: tbl,
			TableType: tableTypes[tbl],
			Columns:   []ColumnSchema{},
		}

		// Get columns
		colQuery := `
			SELECT
				c.column_name,
				c.data_type,
				c.is_nullable,
				COALESCE(c.column_default, ''),
				COALESCE(c.character_maximum_length, 0),
				CASE WHEN pk.column_name IS NOT NULL THEN true ELSE false END as is_pk
			FROM information_schema.columns c
			LEFT JOIN (
				SELECT ku.column_name
				FROM information_schema.table_constraints tc
				JOIN information_schema.key_column_usage ku ON tc.constraint_name = ku.constraint_name
				WHERE tc.table_name = $1 AND tc.constraint_type = 'PRIMARY KEY'
			) pk ON c.column_name = pk.column_name
			WHERE c.table_schema = 'public' AND c.table_name = $1
			ORDER BY c.ordinal_position
		`

		colRows, err := db.Query(colQuery, tbl)
		if err != nil {
			logger.Error("[Database Schema] Failed to get columns for %s: %v", tbl, err)
			continue
		}

		for colRows.Next() {
			var col ColumnSchema
			var maxLen int
			if err := colRows.Scan(&col.ColumnName, &col.DataType, &col.IsNullable, &col.ColumnDefault, &maxLen, &col.IsPrimaryKey); err != nil {
				continue
			}
			col.MaxLength = maxLen
			schema.Columns = append(schema.Columns, col)
		}
		colRows.Close()

		schema.ColumnCount = len(schema.Columns)

		// Get estimated row count
		var rowCount int64
		db.QueryRow(fmt.Sprintf("SELECT reltuples::bigint FROM pg_class WHERE relname = '%s'", tbl)).Scan(&rowCount)
		schema.RowCount = rowCount

		response.Tables = append(response.Tables, schema)
	}

	response.TableCount = len(response.Tables)

	return response, nil
}

// ==================== Execute Query (Readonly) ====================

type ExecuteQueryRequest struct {
	ShopID string `json:"shop_id"`
	Query string `json:"query"`
	Limit int    `json:"limit"` // Max rows to return
}

type ExecuteQueryResponse struct {
	Query string                   `json:"query"`
	Columns []string                 `json:"columns"`
	Rows []map[string]interface{} `json:"rows"`
	RowCount int                      `json:"row_count"`
	ExecutionMs int64                    `json:"execution_ms"`
	Truncated bool                     `json:"truncated"`
	GeneratedAt time.Time                `json:"generated_at"`
}

// ExecuteReadonlyQuery executes a SELECT query and returns results
func ExecuteReadonlyQuery(ctx context.Context, shopID, query string, limit int) (*ExecuteQueryResponse, error) {
	if shopID == "" {
		return nil, fmt.Errorf("shop_id is required")
	}
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}

	// Security: Only allow SELECT queries
	normalizedQuery := strings.ToUpper(strings.TrimSpace(query))
	if !strings.HasPrefix(normalizedQuery, "SELECT") {
		return nil, fmt.Errorf("only SELECT queries are allowed (readonly)")
	}

	// Check for dangerous patterns
	dangerousPatterns := []string{
		"DROP", "DELETE", "UPDATE", "INSERT", "ALTER", "CREATE", "TRUNCATE",
		"GRANT", "REVOKE", "EXECUTE", "COPY", "PG_", "INTO OUTFILE",
	}
	for _, pattern := range dangerousPatterns {
		if strings.Contains(normalizedQuery, pattern) {
			return nil, fmt.Errorf("query contains forbidden keyword: %s", pattern)
		}
	}

	// Default and max limit
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}

	// Add LIMIT if not present
	if !strings.Contains(normalizedQuery, "LIMIT") {
		query = fmt.Sprintf("%s LIMIT %d", strings.TrimSuffix(query, ";"), limit)
	}

	logger.Info("[Execute Query] shopid=%s, query=%s", shopID, query)

	db, err := mypg.PgSqlFastConnect(shopID)
	if err != nil {
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	startTime := time.Now()

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	response := &ExecuteQueryResponse{
		Query:       query,
		Columns:     columns,
		Rows:        []map[string]interface{}{},
		GeneratedAt: time.Now(),
	}

	// Scan rows
	rowCount := 0
	for rows.Next() {
		if rowCount >= limit {
			response.Truncated = true
			break
		}

		// Create a slice of interface{} to hold the values
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			logger.Error("[Execute Query] Scan error: %v", err)
			continue
		}

		// Convert to map
		rowMap := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			// Convert []byte to string for better JSON serialization
			if b, ok := val.([]byte); ok {
				rowMap[col] = string(b)
			} else {
				rowMap[col] = val
			}
		}

		response.Rows = append(response.Rows, rowMap)
		rowCount++
	}

	response.RowCount = len(response.Rows)
	response.ExecutionMs = time.Since(startTime).Milliseconds()

	return response, nil
}

// ==================== Get Table Sample Data ====================

type TableSampleRequest struct {
	ShopID string `json:"shop_id"`
	TableName string `json:"table_name"`
	Limit int    `json:"limit"`
}

type TableSampleResponse struct {
	TableName string                   `json:"table_name"`
	Columns []string                 `json:"columns"`
	Rows []map[string]interface{} `json:"rows"`
	RowCount int                      `json:"row_count"`
	TotalRows int64                    `json:"total_rows_estimate"`
	GeneratedAt time.Time                `json:"generated_at"`
}

// GetTableSample returns sample data from a table
func GetTableSample(ctx context.Context, shopID, tableName string, limit int) (*TableSampleResponse, error) {
	if shopID == "" {
		return nil, fmt.Errorf("shop_id is required")
	}
	if tableName == "" {
		return nil, fmt.Errorf("table_name is required")
	}

	// Sanitize table name (allow only alphanumeric and underscore)
	for _, c := range tableName {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_') {
			return nil, fmt.Errorf("invalid table name")
		}
	}

	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	logger.Info("[Table Sample] shopid=%s, table=%s, limit=%d", shopID, tableName, limit)

	// Use ExecuteReadonlyQuery for the actual query
	query := fmt.Sprintf("SELECT * FROM %s LIMIT %d", tableName, limit)
	result, err := ExecuteReadonlyQuery(ctx, shopID, query, limit)
	if err != nil {
		return nil, err
	}

	// Get total row count
	db, _ := mypg.PgSqlFastConnect(shopID)
	var totalRows int64
	db.QueryRow(fmt.Sprintf("SELECT reltuples::bigint FROM pg_class WHERE relname = '%s'", tableName)).Scan(&totalRows)

	return &TableSampleResponse{
		TableName:   tableName,
		Columns:     result.Columns,
		Rows:        result.Rows,
		RowCount:    result.RowCount,
		TotalRows:   totalRows,
		GeneratedAt: time.Now(),
	}, nil
}
