package tools

import (
	"context"
	"fmt"
	"strings"
	"time"

	"smlcloudplatform/internal/goapi/logger"
	myClickHouse "smlcloudplatform/internal/goapi/myclickhouse"
	"smlcloudplatform/internal/goapi/mypg"
)

// ==================== Dev Database Tools (ไม่จำกัด readonly) ====================
// เครื่องมือสำหรับ dev — รัน SQL ได้ทุกประเภท (SELECT, DELETE, INSERT, UPDATE, ALTER, TRUNCATE, DROP)
// ⚠️ ใช้ด้วยความระมัดระวัง — ไม่มี safety net

// ==================== PostgreSQL Dev Command ====================

type PgCommandResponse struct {
	Query        string                   `json:"query"`
	CommandType  string                   `json:"commandtype"` // SELECT, DELETE, INSERT, etc.
	Rows         []map[string]interface{} `json:"rows,omitempty"`
	RowCount     int                      `json:"rowcount"`
	RowsAffected int64                    `json:"rowsaffected"`
	Truncated    bool                     `json:"truncated"`
	ExecutionMs  int64                    `json:"executionms"`
	GeneratedAt  time.Time                `json:"generatedat"`
}

// ExecutePgCommand รัน SQL query/command บน PostgreSQL (dev — ไม่จำกัด readonly)
func ExecutePgCommand(ctx context.Context, holdingCode, query string, limit int) (*PgCommandResponse, error) {
	if holdingCode == "" {
		return nil, fmt.Errorf("holdingcode is required")
	}
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}

	if limit <= 0 {
		limit = 100
	}
	if limit > 10000 {
		limit = 10000
	}

	normalizedQuery := strings.ToUpper(strings.TrimSpace(query))
	commandType := detectCommandType(normalizedQuery)

	logger.Info("[Dev PG Command] holdingcode=%s, type=%s, query=%s", holdingCode, commandType, query)

	db, err := mypg.PgSqlFastConnect(holdingCode)
	if err != nil {
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	startTime := time.Now()

	// SELECT → return rows
	if commandType == "SELECT" {
		// เพิ่ม LIMIT ถ้าไม่มี
		if !strings.Contains(normalizedQuery, "LIMIT") {
			query = fmt.Sprintf("%s LIMIT %d", strings.TrimSuffix(strings.TrimSpace(query), ";"), limit)
		}

		rows, err := db.QueryContext(ctx, query)
		if err != nil {
			return nil, fmt.Errorf("query failed: %w", err)
		}
		defer rows.Close()

		columns, err := rows.Columns()
		if err != nil {
			return nil, fmt.Errorf("failed to get columns: %w", err)
		}

		var resultRows []map[string]interface{}
		rowCount := 0
		truncated := false

		for rows.Next() {
			if rowCount >= limit {
				truncated = true
				break
			}

			values := make([]interface{}, len(columns))
			valuePtrs := make([]interface{}, len(columns))
			for i := range values {
				valuePtrs[i] = &values[i]
			}

			if err := rows.Scan(valuePtrs...); err != nil {
				logger.Error("[Dev PG Command] Scan error: %v", err)
				continue
			}

			rowMap := make(map[string]interface{})
			for i, col := range columns {
				if b, ok := values[i].([]byte); ok {
					rowMap[col] = string(b)
				} else {
					rowMap[col] = values[i]
				}
			}
			resultRows = append(resultRows, rowMap)
			rowCount++
		}

		if resultRows == nil {
			resultRows = []map[string]interface{}{}
		}

		return &PgCommandResponse{
			Query:       query,
			CommandType: commandType,
			Rows:        resultRows,
			RowCount:    len(resultRows),
			Truncated:   truncated,
			ExecutionMs: time.Since(startTime).Milliseconds(),
			GeneratedAt: time.Now(),
		}, nil
	}

	// Non-SELECT (DELETE, INSERT, UPDATE, TRUNCATE, ALTER, DROP, etc.)
	result, err := db.ExecContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("command failed: %w", err)
	}

	var rowsAffected int64
	if ra, err := result.RowsAffected(); err == nil {
		rowsAffected = ra
	}

	return &PgCommandResponse{
		Query:        query,
		CommandType:  commandType,
		RowsAffected: rowsAffected,
		ExecutionMs:  time.Since(startTime).Milliseconds(),
		GeneratedAt:  time.Now(),
	}, nil
}

// ==================== ClickHouse Dev Command ====================

type ChCommandResponse struct {
	Database    string                   `json:"database"`
	Query       string                   `json:"query"`
	CommandType string                   `json:"commandtype"`
	Rows        []map[string]interface{} `json:"rows,omitempty"`
	RowCount    int                      `json:"rowcount"`
	Truncated   bool                     `json:"truncated"`
	ExecutionMs int64                    `json:"executionms"`
	GeneratedAt time.Time                `json:"generatedat"`
}

// ExecuteChCommand รัน SQL query/command บน ClickHouse (dev — ไม่จำกัด readonly)
func ExecuteChCommand(ctx context.Context, database, query string, limit int) (*ChCommandResponse, error) {
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}

	if limit <= 0 {
		limit = 100
	}
	if limit > 10000 {
		limit = 10000
	}

	if database == "" {
		database = myClickHouse.GetDatabaseName()
	}

	normalizedQuery := strings.ToUpper(strings.TrimSpace(query))
	commandType := detectCommandType(normalizedQuery)

	logger.Info("[Dev CH Command] db=%s, type=%s, query=%s", database, commandType, query)

	conn, err := myClickHouse.ClickHouseFastConnect()
	if err != nil {
		return nil, fmt.Errorf("ClickHouse connection failed: %w", err)
	}

	startTime := time.Now()

	// SELECT/SHOW → return rows
	if commandType == "SELECT" || commandType == "SHOW" {
		// เพิ่ม LIMIT ถ้าไม่มี (เฉพาะ SELECT)
		if commandType == "SELECT" && !strings.Contains(normalizedQuery, "LIMIT") {
			query = fmt.Sprintf("%s LIMIT %d", strings.TrimSuffix(strings.TrimSpace(query), ";"), limit)
		}

		rows, err := myClickHouse.QuerySelectAll(conn, query)
		if err != nil {
			return nil, fmt.Errorf("query failed: %w", err)
		}

		if rows == nil {
			rows = []map[string]interface{}{}
		}

		return &ChCommandResponse{
			Database:    database,
			Query:       query,
			CommandType: commandType,
			Rows:        rows,
			RowCount:    len(rows),
			Truncated:   len(rows) >= limit,
			ExecutionMs: time.Since(startTime).Milliseconds(),
			GeneratedAt: time.Now(),
		}, nil
	}

	// Non-SELECT (ALTER TABLE DELETE, INSERT, DROP, TRUNCATE, etc.)
	err = myClickHouse.ExecuteCommand(ctx, conn, query)
	if err != nil {
		return nil, fmt.Errorf("command failed: %w", err)
	}

	return &ChCommandResponse{
		Database:    database,
		Query:       query,
		CommandType: commandType,
		ExecutionMs: time.Since(startTime).Milliseconds(),
		GeneratedAt: time.Now(),
	}, nil
}

// ==================== Helper ====================

func detectCommandType(normalizedQuery string) string {
	prefixes := []string{"SELECT", "INSERT", "UPDATE", "DELETE", "ALTER", "DROP", "TRUNCATE", "CREATE", "SHOW"}
	for _, p := range prefixes {
		if strings.HasPrefix(normalizedQuery, p) {
			return p
		}
	}
	return "OTHER"
}
