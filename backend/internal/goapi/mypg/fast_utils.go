package mypg

import (
	"context"
	"database/sql"
	"fmt"
	"smlcloudplatform/internal/goapi/logger"
	"strconv"
	"sync"
	"time"

	serviceConfig "smlcloudplatform/internal/goapi/config"

	_ "github.com/lib/pq"
)

// PreparedStmtMetadata tracks prepared statement usage for LRU eviction
type PreparedStmtMetadata struct {
	stmt      *sql.Stmt
	lastUsed  time.Time
	hitCount  int64
	createdAt time.Time
}

// Connection pools for fast database access
var (
	connectionPools = make(map[string]*sql.DB)
	poolMutex       sync.RWMutex
	preparedStmts   = make(map[string]*PreparedStmtMetadata)
	stmtMutex       sync.RWMutex

	// LRU cache configuration
	maxPreparedStmts = 1000          // Maximum number of prepared statements
	stmtIdleTimeout  = 1 * time.Hour // Idle timeout before cleanup
)

// PgSqlFastConnect ใช้ unified database manager พร้อม circuit breaker และ retry logic
// MIGRATED: ใช้ mydb.GlobalManagerPool แทน connection pool เดิม
func PgSqlFastConnect(databaseName string) (*sql.DB, error) {
	// ใช้ unified manager pool (แบบใหม่)
	return PgSqlFastConnectV2(databaseName)
}

// PgSqlFastConnectLegacy - เก็บไว้สำหรับ backward compatibility (deprecated)
// ⚠️ DEPRECATED: ใช้ PgSqlFastConnect() หรือ PgSqlFastConnectV2() แทน
func PgSqlFastConnectLegacy(databaseName string) (*sql.DB, error) {
	poolMutex.RLock()
	if db, exists := connectionPools[databaseName]; exists {
		// ตรวจสอบว่า connection ยังใช้งานได้
		if isConnectionHealthy(db) {
			poolMutex.RUnlock()
			return db, nil
		}
		// ถ้า connection เสีย ลบออกจาก pool
		logger.Info("การเชื่อมต่อไม่สมบูรณ์สำหรับฐานข้อมูล %s, กำลังลบออกจาก pool", databaseName)
	}
	poolMutex.RUnlock()

	// สร้าง connection ใหม่พร้อม retry logic
	poolMutex.Lock()
	defer poolMutex.Unlock()

	// Double check ใน case ที่มีคนสร้างไปแล้วระหว่างรอ lock
	if db, exists := connectionPools[databaseName]; exists {
		if isConnectionHealthy(db) {
			return db, nil
		}
		// ปิด connection เก่าที่เสีย
		db.Close()
		delete(connectionPools, databaseName)
	}

	// สร้าง connection ใหม่พร้อม retry
	db, err := createNewConnectionWithRetry(databaseName)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection for %s: %w", databaseName, err)
	}

	connectionPools[databaseName] = db
	logger.Info("สร้าง pool การเชื่อมต่อใหม่สำหรับฐานข้อมูล %s", databaseName)
	return db, nil
}

// isConnectionHealthy ตรวจสอบสุขภาพของ connection
func isConnectionHealthy(db *sql.DB) bool {
	if db == nil {
		return false
	}

	// ตรวจสอบ connection stats ก่อน
	stats := db.Stats()

	// ตรวจสอบว่ามี connection หรือไม่
	if stats.OpenConnections == 0 && stats.Idle == 0 {
		return false
	}

	// ทดสอบ ping ด้วย timeout สั้น
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return false
	}

	return true
}

// createNewConnectionWithRetry สร้าง connection ใหม่พร้อม retry logic
func createNewConnectionWithRetry(databaseName string) (*sql.DB, error) {
	maxRetries := 3
	retryDelay := 1 * time.Second

	var db *sql.DB
	var err error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		db, err = ConnectOptimized(databaseName)
		if err == nil {
			return db, nil
		}

		if attempt < maxRetries {
			time.Sleep(retryDelay)
			retryDelay *= 2
		}
	}

	return nil, fmt.Errorf("failed to connect after %d attempts, last error: %w", maxRetries, err)
}

// ConnectOptimized - version ที่ optimize แล้ว
func ConnectOptimized(databaseName string) (*sql.DB, error) {
	if databaseName == "" {
		return nil, fmt.Errorf("database name is required")
	}

	svcConfig := serviceConfig.NewServiceConfig()

	// Ensure sslmode has a valid value (empty = lib/pq defaults to "require")
	sslMode := svcConfig.PostgresSSLMode()
	if sslMode == "" {
		sslMode = "disable"
	}

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s "+
			"TimeZone=UTC "+ // Timezone Iron Rule: DB stores UTC+0
			"connect_timeout=10 "+
			"statement_timeout=300000 "+
			"idle_in_transaction_session_timeout=60000 "+
			"application_name=goapi "+
			"binary_parameters=yes",
		svcConfig.PostgresHost(),
		svcConfig.PostgresPort(),
		svcConfig.PostgresUser(),
		svcConfig.PostgresPassword(),
		databaseName,
		sslMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// ⭐ STANDARDIZED pool settings (ลดลงครึ่งหนึ่งเพื่อไม่กระทบระบบอื่น)
	db.SetMaxOpenConns(25)                  // 25 connections - ลดลงครึ่งหนึ่ง
	db.SetMaxIdleConns(7)                   // 7 idle connections - ลดลงครึ่งหนึ่ง
	db.SetConnMaxLifetime(10 * time.Minute) // 10 นาที - refresh connections เป็นระยะ
	db.SetConnMaxIdleTime(5 * time.Minute)  // 5 นาที - ปล่อย idle connections

	// Quick ping test
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// QuerySelectAllFast - version ที่ใช้ prepared statement with LRU cache
func QuerySelectAllFast(db *sql.DB, query string, args ...any) ([]map[string]any, error) {
	stmtKey := fmt.Sprintf("%p_%s", db, query)

	stmtMutex.RLock()
	metadata, exists := preparedStmts[stmtKey]
	stmtMutex.RUnlock()

	if !exists {
		stmtMutex.Lock()
		// Double check
		if metadata, exists = preparedStmts[stmtKey]; !exists {
			// Check if we need to evict old statements (LRU)
			if len(preparedStmts) >= maxPreparedStmts {
				evictLRUStatement()
			}

			var err error
			stmt, err := db.Prepare(query)
			if err != nil {
				stmtMutex.Unlock()
				// Fallback to normal query
				return QuerySelectAll(db, query, args...)
			}

			now := time.Now()
			metadata = &PreparedStmtMetadata{
				stmt:      stmt,
				lastUsed:  now,
				hitCount:  0,
				createdAt: now,
			}
			preparedStmts[stmtKey] = metadata
		}
		stmtMutex.Unlock()
	}

	// Update usage tracking
	updateStatementUsage(stmtKey)

	rows, err := metadata.stmt.Query(args...)
	if err != nil {
		// Fallback to normal query
		return QuerySelectAll(db, query, args...)
	}
	defer rows.Close()

	return scanRowsToMapFast(rows)
}

// updateStatementUsage updates the last used time and hit count
func updateStatementUsage(stmtKey string) {
	stmtMutex.Lock()
	defer stmtMutex.Unlock()

	if metadata, exists := preparedStmts[stmtKey]; exists {
		metadata.lastUsed = time.Now()
		metadata.hitCount++
	}
}

// evictLRUStatement removes the least recently used statement
// NOTE: Must be called with stmtMutex locked
func evictLRUStatement() {
	if len(preparedStmts) == 0 {
		return
	}

	// Find the statement with oldest lastUsed
	var oldestKey string
	var oldestTime time.Time
	firstIteration := true

	for key, metadata := range preparedStmts {
		if firstIteration || metadata.lastUsed.Before(oldestTime) {
			oldestKey = key
			oldestTime = metadata.lastUsed
			firstIteration = false
		}
	}

	// Remove the oldest statement
	if oldestKey != "" {
		metadata := preparedStmts[oldestKey]
		if metadata.stmt != nil {
			metadata.stmt.Close()
		}
		delete(preparedStmts, oldestKey)
	}
}

// scanRowsToMapFast - optimized version
func scanRowsToMapFast(rows *sql.Rows) ([]map[string]any, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, fmt.Errorf("failed to get column types: %w", err)
	}

	dataRows := make([]map[string]any, 0, 100) // pre-allocate for performance
	values := make([]any, len(columns))
	valuePtrs := make([]any, len(columns))

	for i := range columns {
		valuePtrs[i] = &values[i]
	}

	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		rowData := make(map[string]any, len(columns))
		for i, col := range columns {
			val := values[i]
			if val != nil {
				// Optimized type conversion
				switch columnTypes[i].DatabaseTypeName() {
				case "INT4", "INT8", "BIGINT", "INTEGER":
					if v, ok := val.(int64); ok {
						rowData[col] = v
					} else if v, ok := val.(int32); ok {
						rowData[col] = int64(v)
					}
				case "FLOAT4", "FLOAT8", "NUMERIC", "DECIMAL":
					if v, ok := val.(float64); ok {
						rowData[col] = v
					} else if v, ok := val.(string); ok {
						if floatVal, parseErr := strconv.ParseFloat(v, 64); parseErr == nil {
							rowData[col] = floatVal
						} else {
							rowData[col] = v
						}
					}
				default:
					if v, ok := val.([]byte); ok {
						rowData[col] = string(v)
					} else {
						rowData[col] = val
					}
				}
			} else {
				rowData[col] = nil
			}
		}
		dataRows = append(dataRows, rowData)
	}

	return dataRows, rows.Err()
}

// CloseAllPools ปิด connection pools ทั้งหมด
func CloseAllPools() {
	poolMutex.Lock()
	defer poolMutex.Unlock()

	for _, db := range connectionPools {
		if db != nil {
			db.Close()
		}
	}
	connectionPools = make(map[string]*sql.DB)

	stmtMutex.Lock()
	defer stmtMutex.Unlock()

	for _, metadata := range preparedStmts {
		if metadata.stmt != nil {
			metadata.stmt.Close()
		}
	}
	preparedStmts = make(map[string]*PreparedStmtMetadata)
}

// HealthCheckAndReconnect ตรวจสอบและกู้คืน connection ที่เสีย
func HealthCheckAndReconnect() {
	poolMutex.Lock()
	defer poolMutex.Unlock()

	for dbName, db := range connectionPools {
		if !isConnectionHealthy(db) {
			db.Close()
			delete(connectionPools, dbName)

			newDB, err := createNewConnectionWithRetry(dbName)
			if err == nil {
				connectionPools[dbName] = newDB
			}
		}
	}
}

// GetConnectionStats ดูสถิติ connection pools
func GetConnectionStats() map[string]sql.DBStats {
	poolMutex.RLock()
	defer poolMutex.RUnlock()

	stats := make(map[string]sql.DBStats)
	for name, db := range connectionPools {
		if db != nil {
			stats[name] = db.Stats()
		}
	}
	return stats
}

// CleanupPreparedStatements ทำความสะอาด prepared statements ที่ idle เกินกำหนด
func CleanupPreparedStatements() {
	stmtMutex.Lock()
	defer stmtMutex.Unlock()

	now := time.Now()

	for key, metadata := range preparedStmts {
		if now.Sub(metadata.lastUsed) > stmtIdleTimeout {
			if metadata.stmt != nil {
				metadata.stmt.Close()
			}
			delete(preparedStmts, key)
		}
	}
}

// GetPreparedStatementCount นับจำนวน prepared statements ที่มีอยู่
func GetPreparedStatementCount() int {
	stmtMutex.RLock()
	defer stmtMutex.RUnlock()
	return len(preparedStmts)
}
