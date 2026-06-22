package mydb

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/myclickhouse"

	"github.com/ClickHouse/clickhouse-go/v2"
	_ "github.com/lib/pq"
)

// DatabaseType - ประเภทของ database
type DatabaseType string

const (
	PostgreSQL DatabaseType = "PostgreSQL"
	ClickHouse DatabaseType = "ClickHouse"
)

// DatabaseManager จัดการ database connections แบบ unified พร้อม circuit breaker
type DatabaseManager struct {
	postgreSQLConn  *sql.DB
	clickHouseConn  clickhouse.Conn
	postgresMutex   sync.RWMutex
	clickHouseMutex sync.RWMutex
	perfLogger      *DatabasePerformanceLogger
	statsCollector  *QueryStatsCollector

	// Circuit breakers สำหรับป้องกัน cascade failures
	pgCircuitBreaker *CircuitBreaker
	chCircuitBreaker *CircuitBreaker

	// Retry configuration
	retryConfig RetryConfig
}

// DatabaseConfig configuration สำหรับ databases
type DatabaseConfig struct {
	PostgreSQLHost     string
	PostgreSQLPort     string
	PostgreSQLUser     string
	PostgreSQLPassword string
	PostgreSQLDatabase string
	PostgreSQLSSLMode  string

	ClickHouseHost     string
	ClickHousePort     string
	ClickHouseUser     string
	ClickHousePassword string
	ClickHouseDatabase string
}

// NewDatabaseManager สร้าง database manager ใหม่พร้อม circuit breaker และ retry logic
func NewDatabaseManager(config DatabaseConfig) (*DatabaseManager, error) {
	logger.Info("กำลังสร้าง Database Manager พร้อม Circuit Breaker")

	perfLogger := NewDatabasePerformanceLogger("unified")
	statsCollector := NewQueryStatsCollector()

	dm := &DatabaseManager{
		perfLogger:       perfLogger,
		statsCollector:   statsCollector,
		pgCircuitBreaker: NewCircuitBreaker("PostgreSQL"),
		chCircuitBreaker: NewCircuitBreaker("ClickHouse"),
		retryConfig:      DefaultRetryConfig(),
	}

	// เชื่อมต่อ PostgreSQL พร้อม retry
	err := ExecuteWithRetry(context.Background(), dm.retryConfig, func(ctx context.Context) error {
		return dm.connectPostgreSQL(config)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect PostgreSQL after retries: %w", err)
	}

	// เชื่อมต่อ ClickHouse พร้อม retry
	err = ExecuteWithRetry(context.Background(), dm.retryConfig, func(ctx context.Context) error {
		return dm.connectClickHouse(config)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect ClickHouse after retries: %w", err)
	}

	logger.Success("Database Manager สร้างสำเร็จ (PostgreSQL + ClickHouse + Circuit Breakers)")
	return dm, nil
}

func createPostgreSQLDatabase(config DatabaseConfig, dbName string) error {
	sslMode := config.PostgreSQLSSLMode
	if sslMode == "" {
		sslMode = "disable"
	}

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=postgres sslmode=%s connect_timeout=10 TimeZone=UTC",
		config.PostgreSQLHost,
		config.PostgreSQLPort,
		config.PostgreSQLUser,
		config.PostgreSQLPassword,
		sslMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open admin PostgreSQL connection: %w", err)
	}
	defer db.Close()

	query := fmt.Sprintf(`CREATE DATABASE "%s"`, strings.ReplaceAll(dbName, `"`, `""`))

	_, err = db.Exec(query)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return nil
		}
		return fmt.Errorf("failed to create database %s: %w", dbName, err)
	}

	logger.Success("PostgreSQL database '%s' created successfully", dbName)
	return nil
}

func (dm *DatabaseManager) connectPostgreSQL(config DatabaseConfig) error {
	dm.postgresMutex.Lock()
	defer dm.postgresMutex.Unlock()

	if config.PostgreSQLDatabase == "" {
		return fmt.Errorf("database name is required (holdingcode must not be empty)")
	}

	// Ensure sslmode has a valid value (empty = lib/pq defaults to "require")
	sslMode := config.PostgreSQLSSLMode
	if sslMode == "" {
		sslMode = "disable"
	}

	// สร้าง connection string
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s "+
			"TimeZone=UTC "+ // Timezone Iron Rule: DB stores UTC+0
			"connect_timeout=10 "+
			"statement_timeout=300000 "+
			"idle_in_transaction_session_timeout=60000",
		config.PostgreSQLHost,
		config.PostgreSQLPort,
		config.PostgreSQLUser,
		config.PostgreSQLPassword,
		config.PostgreSQLDatabase,
		sslMode,
	)

	logger.Info("กำลังเชื่อมต่อ PostgreSQL host=%s db=%s (พร้อม connection pooling)",
		config.PostgreSQLHost, config.PostgreSQLDatabase)

	var err error
	dm.postgreSQLConn, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open PostgreSQL connection: %w", err)
	}

	// ตั้งค่า connection pool สำหรับ high traffic (ลดลงครึ่งหนึ่งเพื่อไม่กระทบระบบอื่น)
	dm.postgreSQLConn.SetMaxOpenConns(50)
	dm.postgreSQLConn.SetMaxIdleConns(15)
	dm.postgreSQLConn.SetConnMaxLifetime(1 * time.Hour)
	dm.postgreSQLConn.SetConnMaxIdleTime(15 * time.Minute)

	// ทดสอบการเชื่อมต่อ
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := dm.postgreSQLConn.PingContext(ctx); err != nil {
		dm.postgreSQLConn.Close()

		errStr := err.Error()
		if strings.Contains(errStr, "does not exist") || strings.Contains(errStr, "3D000") {
			logger.Info("Database '%s' does not exist. Attempting to create it...", config.PostgreSQLDatabase)
			if createErr := createPostgreSQLDatabase(config, config.PostgreSQLDatabase); createErr != nil {
				return fmt.Errorf("failed to auto-create database: %w", createErr)
			}

			// Retry connecting
			dm.postgreSQLConn, err = sql.Open("postgres", connStr)
			if err != nil {
				return fmt.Errorf("failed to reopen PostgreSQL connection: %w", err)
			}

			dm.postgreSQLConn.SetMaxOpenConns(50)
			dm.postgreSQLConn.SetMaxIdleConns(15)
			dm.postgreSQLConn.SetConnMaxLifetime(1 * time.Hour)
			dm.postgreSQLConn.SetConnMaxIdleTime(15 * time.Minute)

			ctxRetry, cancelRetry := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancelRetry()
			if pingErr := dm.postgreSQLConn.PingContext(ctxRetry); pingErr != nil {
				dm.postgreSQLConn.Close()
				return fmt.Errorf("failed to ping PostgreSQL after database creation: %w", pingErr)
			}
		} else {
			return fmt.Errorf("failed to ping PostgreSQL: %w", err)
		}
	}

	logger.Success("เชื่อมต่อ PostgreSQL สำเร็จ (pool: max_open=100, max_idle=30, lifetime=1h, idle_time=15m)")
	return nil
}

// connectClickHouse เชื่อมต่อ ClickHouse พร้อม connection pooling (เลิกใช้งานแล้ว)
func (dm *DatabaseManager) connectClickHouse(config DatabaseConfig) error {
	// ปิดและ bypass การเชื่อมต่อ ClickHouse ทั้งหมด
	return nil
}

// GetPostgreSQLConnection ดึง PostgreSQL connection พร้อม performance logging
func (dm *DatabaseManager) GetPostgreSQLConnection() (*sql.DB, error) {
	dm.postgresMutex.RLock()
	defer dm.postgresMutex.RUnlock()

	if dm.postgreSQLConn == nil {
		return nil, fmt.Errorf("PostgreSQL connection is not initialized")
	}

	// ตรวจสอบสุขภาพ
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := dm.postgreSQLConn.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("PostgreSQL connection is unhealthy: %w", err)
	}

	return dm.postgreSQLConn, nil
}

// GetClickHouseConnection ดึง ClickHouse connection พร้อม performance logging
func (dm *DatabaseManager) GetClickHouseConnection() (clickhouse.Conn, error) {
	dm.clickHouseMutex.RLock()
	defer dm.clickHouseMutex.RUnlock()

	if dm.clickHouseConn == nil {
		return nil, fmt.Errorf("ClickHouse connection is not initialized")
	}

	// ตรวจสอบสุขภาพ
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := dm.clickHouseConn.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ClickHouse connection is unhealthy: %w", err)
	}

	return dm.clickHouseConn, nil
}

// QueryPostgreSQL รัน query บน PostgreSQL พร้อม circuit breaker และ performance logging
func (dm *DatabaseManager) QueryPostgreSQL(ctx context.Context, query string, args ...any) ([]map[string]any, error) {
	startTime := time.Now()
	var results []map[string]any

	// ใช้ circuit breaker เพื่อป้องกัน cascade failures
	err := dm.pgCircuitBreaker.Execute(ctx, func(ctx context.Context) error {
		db, err := dm.GetPostgreSQLConnection()
		if err != nil {
			return err
		}

		rows, err := db.QueryContext(ctx, query, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		// แปลงผลลัพธ์เป็น map
		columns, err := rows.Columns()
		if err != nil {
			return err
		}

		results = make([]map[string]any, 0)
		for rows.Next() {
			rowData := make(map[string]any, len(columns))
			values := make([]any, len(columns))
			valuePtrs := make([]any, len(columns))

			for i := range values {
				valuePtrs[i] = &values[i]
			}

			if err := rows.Scan(valuePtrs...); err != nil {
				return err
			}

			for i, col := range columns {
				rowData[col] = values[i]
			}

			results = append(results, rowData)
		}
		return nil
	})

	// บันทึก performance
	status := "success"
	if err != nil {
		status = "error"
	}
	dm.logPerformance(ctx, "PostgreSQL", "Query", query, int64(len(results)), startTime, status, err)

	return results, err
}

// QueryClickHouse รัน query บน ClickHouse พร้อม circuit breaker และ performance logging
func (dm *DatabaseManager) QueryClickHouse(ctx context.Context, query string) ([]map[string]any, error) {
	startTime := time.Now()
	var results []map[string]any

	// ใช้ circuit breaker เพื่อป้องกัน cascade failures
	err := dm.chCircuitBreaker.Execute(ctx, func(ctx context.Context) error {
		conn, err := dm.GetClickHouseConnection()
		if err != nil {
			return err
		}

		// ใช้ existing ClickHouse query function
		results, err = myclickhouse.QuerySelectAll(conn, query)
		return err
	})

	// บันทึก performance
	status := "success"
	if err != nil {
		status = "error"
	}
	dm.logPerformance(ctx, "ClickHouse", "Query", query, int64(len(results)), startTime, status, err)

	return results, err
}

// ExecPostgreSQL รันคำสั่งบน PostgreSQL พร้อม circuit breaker และ performance logging
func (dm *DatabaseManager) ExecPostgreSQL(ctx context.Context, query string, args ...any) (sql.Result, error) {
	startTime := time.Now()
	var result sql.Result
	var rowsAffected int64

	// ใช้ circuit breaker เพื่อป้องกัน cascade failures
	err := dm.pgCircuitBreaker.Execute(ctx, func(ctx context.Context) error {
		db, err := dm.GetPostgreSQLConnection()
		if err != nil {
			return err
		}

		result, err = db.ExecContext(ctx, query, args...)
		if err == nil {
			rowsAffected, _ = result.RowsAffected()
		}
		return err
	})

	// บันทึก performance
	status := "success"
	if err != nil {
		status = "error"
	}
	dm.logPerformance(ctx, "PostgreSQL", "Exec", query, rowsAffected, startTime, status, err)

	return result, err
}

// ExecClickHouse รันคำสั่งบน ClickHouse พร้อม circuit breaker และ performance logging
func (dm *DatabaseManager) ExecClickHouse(ctx context.Context, query string) error {
	startTime := time.Now()

	// ใช้ circuit breaker เพื่อป้องกัน cascade failures
	err := dm.chCircuitBreaker.Execute(ctx, func(ctx context.Context) error {
		conn, err := dm.GetClickHouseConnection()
		if err != nil {
			return err
		}
		return conn.Exec(ctx, query)
	})

	// บันทึก performance
	status := "success"
	if err != nil {
		status = "error"
	}
	dm.logPerformance(ctx, "ClickHouse", "Exec", query, 0, startTime, status, err)

	return err
}

// BatchInsertPostgreSQL แทรกข้อมูลแบบ batch บน PostgreSQL
func (dm *DatabaseManager) BatchInsertPostgreSQL(ctx context.Context, tableName string, columns []string, data [][]any) error {
	startTime := time.Now()

	db, err := dm.GetPostgreSQLConnection()
	if err != nil {
		dm.logPerformance(ctx, "PostgreSQL", "BatchInsert", fmt.Sprintf("INSERT %s", tableName), 0, startTime, "error", err)
		return err
	}

	// ใช้ existing PostgreSQL bulk insert function (ต้อง import จาก mypg package)
	// สำหรับตอนนี้ใช้ simple approach
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		dm.logPerformance(ctx, "PostgreSQL", "BatchInsert", fmt.Sprintf("INSERT %s", tableName), 0, startTime, "error", err)
		return err
	}
	defer tx.Rollback()

	// สร้าง INSERT query
	placeholders := make([]string, len(columns))
	for i := range columns {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", tableName,
		strings.Join(columns, ", "), strings.Join(placeholders, ", "))

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		dm.logPerformance(ctx, "PostgreSQL", "BatchInsert", fmt.Sprintf("INSERT %s", tableName), 0, startTime, "error", err)
		return err
	}
	defer stmt.Close()

	successCount := 0
	for i, row := range data {
		_, err := stmt.ExecContext(ctx, row...)
		if err != nil {
			logger.Error("แถว %d ล้มเหลว: %v", i, err)
			dm.logPerformance(ctx, "PostgreSQL", "BatchInsert", fmt.Sprintf("INSERT %s", tableName), 0, startTime, "error", err)
			return fmt.Errorf("batch insert failed at row %d: %w", i, err)
		}
		successCount++
	}

	if err := tx.Commit(); err != nil {
		dm.logPerformance(ctx, "PostgreSQL", "BatchInsert", fmt.Sprintf("INSERT %s", tableName), 0, startTime, "error", err)
		return err
	}

	dm.logPerformance(ctx, "PostgreSQL", "BatchInsert", fmt.Sprintf("INSERT %s", tableName), int64(successCount), startTime, "success", nil)
	return nil
}

// BatchInsertClickHouse แทรกข้อมูลแบบ batch บน ClickHouse
func (dm *DatabaseManager) BatchInsertClickHouse(ctx context.Context, tableName string, columns []string, data [][]any) error {
	startTime := time.Now()

	conn, err := dm.GetClickHouseConnection()
	if err != nil {
		dm.logPerformance(ctx, "ClickHouse", "BatchInsert", fmt.Sprintf("INSERT %s", tableName), 0, startTime, "error", err)
		return err
	}

	if len(data) == 0 {
		dm.logPerformance(ctx, "ClickHouse", "BatchInsert", fmt.Sprintf("INSERT %s", tableName), 0, startTime, "success", nil)
		return nil
	}

	// เตรียม batch insert statement
	placeholders := make([]string, len(columns))
	for i := range columns {
		placeholders[i] = "?"
	}

	insertQuery := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		tableName, strings.Join(columns, ", "), strings.Join(placeholders, ", "))

	// สร้าง batch
	batch, err := conn.PrepareBatch(ctx, insertQuery)
	if err != nil {
		dm.logPerformance(ctx, "ClickHouse", "BatchInsert", fmt.Sprintf("INSERT %s", tableName), 0, startTime, "error", err)
		return fmt.Errorf("failed to prepare batch: %w", err)
	}
	defer batch.Close()

	successCount := 0
	for i, row := range data {
		err := batch.Append(row...)
		if err != nil {
			logger.Error("ClickHouse แถว %d ล้มเหลว: %v", i, err)
			dm.logPerformance(ctx, "ClickHouse", "BatchInsert", fmt.Sprintf("INSERT %s", tableName), 0, startTime, "error", err)
			return fmt.Errorf("batch insert failed at row %d: %w", i, err)
		}
		successCount++
	}

	// ส่ง batch ทั้งหมด
	err = batch.Send()
	if err != nil {
		dm.logPerformance(ctx, "ClickHouse", "BatchInsert", fmt.Sprintf("INSERT %s", tableName), 0, startTime, "error", err)
		return fmt.Errorf("failed to send batch: %w", err)
	}

	dm.logPerformance(ctx, "ClickHouse", "BatchInsert", fmt.Sprintf("INSERT %s", tableName), int64(successCount), startTime, "success", nil)
	logger.Success("ClickHouse batch insert สำเร็จ: %d/%d rows", successCount, len(data))
	return nil
}

func (dm *DatabaseManager) logPerformance(ctx context.Context, database, operation, query string, rowsAffected int64, startTime time.Time, status string, err error) {
	duration := time.Since(startTime)

	logMessage := fmt.Sprintf("[PERF] %s.%s: %s (duration: %v, rows: %d, status: %s)",
		database, operation, truncateQuery(query), duration.Round(time.Millisecond), rowsAffected, status)

	if err != nil {
		logMessage += fmt.Sprintf(", error: %v", err)
	}

	// แสดงผลตามระดับ performance
	switch {
	case duration < 100*time.Millisecond:
		logger.Debug(logMessage)
	case duration < 1*time.Second:
		logger.Info(logMessage)
	case duration < 5*time.Second:
		logger.Warn(logMessage)
	default:
		logger.Error(logMessage)
		logger.Error("🔍 SLOW QUERY DETECTED - Database: %s, Operation: %s, Duration: %v",
			database, operation, duration.Round(time.Millisecond))
		logger.Error("Query: %s", truncateQuery(query))
	}
}

// truncateQuery ตัด query ที่ยาวเกินไป
func truncateQuery(query string) string {
	if len(query) > 200 {
		return query[:200] + "... [TRUNCATED]"
	}
	return query
}

// GetStats ดึงสถิติ performance
func (dm *DatabaseManager) GetStats(databaseType DatabaseType) interface{} {
	switch databaseType {
	case PostgreSQL:
		dm.postgresMutex.RLock()
		defer dm.postgresMutex.RUnlock()
		if dm.postgreSQLConn != nil {
			return dm.postgreSQLConn.Stats()
		}
	case ClickHouse:
		dm.clickHouseMutex.RLock()
		defer dm.clickHouseMutex.RUnlock()
		// ClickHouse ไม่มี Stats แบบ PostgreSQL
		return map[string]interface{}{
			"status": "available",
			"type":   "ClickHouse",
		}
	}
	return nil
}

// Close ปิดการเชื่อมต่อทั้งหมด
func (dm *DatabaseManager) Close() error {
	dm.postgresMutex.Lock()
	defer dm.postgresMutex.Unlock()

	dm.clickHouseMutex.Lock()
	defer dm.clickHouseMutex.Unlock()

	var errors []error

	if dm.postgreSQLConn != nil {
		if err := dm.postgreSQLConn.Close(); err != nil {
			errors = append(errors, fmt.Errorf("PostgreSQL close error: %w", err))
		}
		dm.postgreSQLConn = nil
	}

	if dm.clickHouseConn != nil {
		if err := dm.clickHouseConn.Close(); err != nil {
			errors = append(errors, fmt.Errorf("ClickHouse close error: %w", err))
		}
		dm.clickHouseConn = nil
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors closing connections: %v", errors)
	}

	logger.Success("ปิด Database Manager สำเร็จ")
	return nil
}

// PingAll ทดสอบการเชื่อมต่อทั้งหมด
func (dm *DatabaseManager) PingAll() map[string]error {
	results := make(map[string]error)

	// ทดสอบ PostgreSQL
	if _, err := dm.GetPostgreSQLConnection(); err != nil {
		results["PostgreSQL"] = err
	} else {
		results["PostgreSQL"] = nil
	}

	// ทดสอบ ClickHouse
	if _, err := dm.GetClickHouseConnection(); err != nil {
		results["ClickHouse"] = err
	} else {
		results["ClickHouse"] = nil
	}

	return results
}

// GetCircuitBreakerStats - ดึงสถิติของ circuit breakers
func (dm *DatabaseManager) GetCircuitBreakerStats() map[string]interface{} {
	return map[string]interface{}{
		"postgresql": dm.pgCircuitBreaker.GetStats(),
		"clickhouse": dm.chCircuitBreaker.GetStats(),
	}
}

func (dm *DatabaseManager) ResetCircuitBreakers() {
	logger.Info("Resetting all circuit breakers...")
	dm.pgCircuitBreaker.Reset()
	dm.chCircuitBreaker.Reset()
	logger.Success("All circuit breakers reset successfully")
}

// GetGlobalConnection - ดึง connection สำหรับ shop (backward compatibility)
func GetGlobalConnection(holdingCode string) (*sql.DB, error) {
	// สร้าง config จาก environment
	config := DatabaseConfig{
		PostgreSQLHost:     getEnv("POSTGRES_HOST", "localhost"),
		PostgreSQLPort:     getEnv("POSTGRES_PORT", "5432"),
		PostgreSQLUser:     getEnv("POSTGRES_USER", "postgres"),
		PostgreSQLPassword: getEnv("POSTGRES_PASSWORD", ""),
		PostgreSQLDatabase: holdingCode,
		PostgreSQLSSLMode:  getEnv("POSTGRES_SSL_MODE", "disable"),

		ClickHouseHost:     getEnv("CLICKHOUSE_HOST", "localhost"),
		ClickHousePort:     getEnv("CLICKHOUSE_PORT", "9000"),
		ClickHouseUser:     getEnv("CLICKHOUSE_USER", "default"),
		ClickHousePassword: getEnv("CLICKHOUSE_PASSWORD", ""),
		ClickHouseDatabase: holdingCode,
	}

	// สร้าง manager และ return PostgreSQL connection
	manager, err := NewDatabaseManager(config)
	if err != nil {
		return nil, err
	}

	return manager.GetPostgreSQLConnection()
}

// GetGlobalManager - ดึง DatabaseManager สำหรับ shop (backward compatibility)
func GetGlobalManager(holdingCode string) (*DatabaseManager, error) {
	// สร้าง config จาก environment
	config := DatabaseConfig{
		PostgreSQLHost:     getEnv("POSTGRES_HOST", "localhost"),
		PostgreSQLPort:     getEnv("POSTGRES_PORT", "5432"),
		PostgreSQLUser:     getEnv("POSTGRES_USER", "postgres"),
		PostgreSQLPassword: getEnv("POSTGRES_PASSWORD", ""),
		PostgreSQLDatabase: holdingCode,
		PostgreSQLSSLMode:  getEnv("POSTGRES_SSL_MODE", "disable"),

		ClickHouseHost:     getEnv("CLICKHOUSE_HOST", "localhost"),
		ClickHousePort:     getEnv("CLICKHOUSE_PORT", "9000"),
		ClickHouseUser:     getEnv("CLICKHOUSE_USER", "default"),
		ClickHousePassword: getEnv("CLICKHOUSE_PASSWORD", ""),
		ClickHouseDatabase: holdingCode,
	}

	return NewDatabaseManager(config)
}

// getEnv helper function สำหรับ environment variables
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
