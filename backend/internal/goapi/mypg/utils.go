package mypg

import (
	"context"
	"database/sql"
	"fmt"
	"smlcloudplatform/internal/goapi/logger"
	"strconv"
	"strings"
	"time"

	serviceConfig "smlcloudplatform/internal/goapi/config"

	"github.com/lib/pq"
)

// Connect สร้างการเชื่อมต่อกับ PostgreSQL พร้อม connection pooling ที่เหมาะสม
// Connect - เชื่อมต่อ PostgreSQL โดยใช้ unified database manager
// ⚠️ MIGRATED: ใช้ mydb.GlobalManagerPool แทน sql.Open()
// คืนค่า *sql.DB ให้ caller รับผิดชอบปิดด้วย db.Close() (deprecated - connection managed by pool)
func Connect(databaseName string) (*sql.DB, error) {
	// ใช้ unified manager pool แทน (migration complete)
	logger.Info("Connect: Using unified database manager for db=%s", databaseName)
	return PgSqlFastConnectV2(databaseName)
}

// ConnectLegacy - เก็บไว้สำหรับ reference (deprecated)
// ⚠️ DEPRECATED: ใช้ Connect() หรือ PgSqlFastConnect() แทน
func ConnectLegacy(databaseName string) (*sql.DB, error) {
	svcConfig := serviceConfig.NewServiceConfig()

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s "+
			"TimeZone=UTC "+ // Timezone Iron Rule: DB stores UTC+0
			"connect_timeout=10 "+
			"statement_timeout=300000 "+
			"idle_in_transaction_session_timeout=60000",
		svcConfig.PostgresHost(),
		svcConfig.PostgresPort(),
		svcConfig.PostgresUser(),
		svcConfig.PostgresPassword(),
		databaseName,
		svcConfig.PostgresSSLMode(),
	)
	logger.Info("กำลังเชื่อมต่อไปยัง Postgres host=%s db=%s (พร้อม pooling)", svcConfig.PostgresHost(), databaseName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// ⚡ OPTIMIZED: ใช้ connection pooling ที่ standardized (ลดลงครึ่งหนึ่งเพื่อไม่กระทบระบบอื่น)
	// - MaxOpenConns: 25 (ลดลงครึ่งหนึ่ง)
	// - MaxIdleConns: 7 (ลดลงครึ่งหนึ่ง)
	// - ConnMaxLifetime: 10min (refresh connections เป็นระยะ)
	// - ConnMaxIdleTime: 5min (ปล่อย idle connections ช้าลง)
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(7)
	db.SetConnMaxLifetime(10 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	// ทดสอบการเชื่อมต่อ (ping)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Success("เชื่อมต่อ PostgreSQL สำเร็จ (pool: max_open=50, max_idle=15, lifetime=10m)")
	return db, nil
}

// Close ปิดการเชื่อมต่อฐานข้อมูล (caller สามารถเรียก db.Close() เองได้)
func Close(db *sql.DB) error {
	if db != nil {
		return db.Close()
	}
	return nil
}

// QuerySelectAll ดึงข้อมูลทั้งหมดจาก query และแปลงเป็น map (Optimized version)
func QuerySelectAll(db *sql.DB, query string, args ...any) ([]map[string]any, error) {
	// เพิ่ม timeout เป็น 10 นาที สำหรับ query ที่ซับซ้อน (เช่น รายงาน)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, fmt.Errorf("failed to get column types: %w", err)
	}

	// Pre-allocate with estimated capacity to reduce re-allocations
	dataRows := make([]map[string]any, 0, 100)

	// Pre-calculate column count
	colCount := len(columns)

	// Reuse slices for better performance
	values := make([]any, colCount)
	valuePtrs := make([]any, colCount)
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	// Pre-build type mapping for faster type checking
	typeMap := make([]string, colCount)
	for i := range columnTypes {
		typeMap[i] = columnTypes[i].DatabaseTypeName()
	}

	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Pre-allocate map with exact size
		rowData := make(map[string]any, colCount)

		for i := 0; i < colCount; i++ {
			val := values[i]

			if val == nil {
				rowData[columns[i]] = nil
				continue
			}

			// Use pre-built type map for faster comparison
			switch typeMap[i] {
			case "INT4", "INT8", "BIGINT", "INTEGER", "SMALLINT", "INT2":
				switch v := val.(type) {
				case int64:
					rowData[columns[i]] = v
				case int32:
					rowData[columns[i]] = int64(v)
				case int16:
					rowData[columns[i]] = int64(v)
				case int:
					rowData[columns[i]] = int64(v)
				default:
					rowData[columns[i]] = val
				}

			case "FLOAT4", "FLOAT8", "NUMERIC", "DECIMAL", "REAL", "DOUBLE PRECISION":
				switch v := val.(type) {
				case float64:
					rowData[columns[i]] = v
				case float32:
					rowData[columns[i]] = float64(v)
				case string:
					if floatVal, parseErr := strconv.ParseFloat(v, 64); parseErr == nil {
						rowData[columns[i]] = floatVal
					} else {
						rowData[columns[i]] = v
					}
				case []byte:
					if floatVal, parseErr := strconv.ParseFloat(string(v), 64); parseErr == nil {
						rowData[columns[i]] = floatVal
					} else {
						rowData[columns[i]] = string(v)
					}
				default:
					rowData[columns[i]] = val
				}

			case "BOOL", "BOOLEAN":
				if v, ok := val.(bool); ok {
					rowData[columns[i]] = v
				} else {
					rowData[columns[i]] = val
				}

			case "TIMESTAMP", "TIMESTAMPTZ", "DATE", "TIME", "TIMETZ":
				if v, ok := val.(time.Time); ok {
					rowData[columns[i]] = v
				} else {
					rowData[columns[i]] = val
				}

			case "TEXT", "VARCHAR", "CHAR", "BPCHAR":
				switch v := val.(type) {
				case []byte:
					rowData[columns[i]] = string(v)
				case string:
					rowData[columns[i]] = v
				default:
					rowData[columns[i]] = val
				}

			default:
				// Handle other types
				if v, ok := val.([]byte); ok {
					rowData[columns[i]] = string(v)
				} else {
					rowData[columns[i]] = val
				}
			}
		}
		dataRows = append(dataRows, rowData)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return dataRows, nil
}

// QuerySelectOne ดึงข้อมูลเพียงแถวเดียว
func QuerySelectOne(db *sql.DB, query string, args ...any) (map[string]any, error) {
	results, err := QuerySelectAll(db, query, args...)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, sql.ErrNoRows
	}

	return results[0], nil
}

// ExecuteCommand รันคำสั่ง SQL (INSERT, UPDATE, DELETE)
func ExecuteCommand(db *sql.DB, query string, args ...any) error {
	result, err := db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to execute command: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Warn("ไม่สามารถดึงจำนวน rows ที่ถูกกระทำ: %v", err)
	} else {
		logger.Info("คำสั่งดำเนินการสำเร็จ, กระทบ %d แถว", rowsAffected)
	}

	return nil
}

// Insert ฟังก์ชันสำหรับ insert ข้อมูลแบบง่าย
func Insert(db *sql.DB, tableName string, data map[string]any) error {
	if len(data) == 0 {
		return fmt.Errorf("no data to insert")
	}

	columns := make([]string, 0, len(data))
	placeholders := make([]string, 0, len(data))
	values := make([]any, 0, len(data))

	i := 1
	for column, value := range data {
		columns = append(columns, column)
		placeholders = append(placeholders, fmt.Sprintf("$%d", i))
		values = append(values, value)
		i++
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		tableName,
		fmt.Sprintf("%s", columns),
		fmt.Sprintf("%s", placeholders))

	return ExecuteCommand(db, query, values...)
}

// Update ฟังก์ชันสำหรับ update ข้อมูลแบบง่าย
func Update(db *sql.DB, tableName string, data map[string]any, where string, whereArgs ...any) error {
	if len(data) == 0 {
		return fmt.Errorf("no data to update")
	}

	setParts := make([]string, 0, len(data))
	values := make([]any, 0, len(data))

	i := 1
	for column, value := range data {
		setParts = append(setParts, fmt.Sprintf("%s = $%d", column, i))
		values = append(values, value)
		i++
	}

	// เพิ่ม where args
	values = append(values, whereArgs...)

	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s",
		tableName,
		fmt.Sprintf("%s", setParts),
		where)

	return ExecuteCommand(db, query, values...)
}

// Delete ฟังก์ชันสำหรับ delete ข้อมูลแบบง่าย
func Delete(db *sql.DB, tableName string, where string, whereArgs ...any) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE %s", tableName, where)
	return ExecuteCommand(db, query, whereArgs...)
}

// BeginTransaction เริ่ม transaction
func BeginTransaction(db *sql.DB) (*sql.Tx, error) {
	return db.Begin()
}

// IsTableExists ตรวจสอบว่าตารางมีอยู่หรือไม่
func IsTableExists(db *sql.DB, tableName string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_schema = 'public'
			AND table_name = $1
		);
	`

	var exists bool
	err := db.QueryRow(query, tableName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check table existence: %w", err)
	}

	return exists, nil
}

// GetTableColumns ดึงรายชื่อคอลัมน์ของตาราง
func GetTableColumns(db *sql.DB, tableName string) ([]string, error) {
	query := `
		SELECT column_name
		FROM information_schema.columns
		WHERE table_schema = 'public'
		AND table_name = $1
		ORDER BY ordinal_position;
	`

	rows, err := db.Query(query, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get table columns: %w", err)
	}
	defer rows.Close()

	var columns []string
	for rows.Next() {
		var column string
		if err := rows.Scan(&column); err != nil {
			return nil, fmt.Errorf("failed to scan column name: %w", err)
		}
		columns = append(columns, column)
	}

	return columns, nil
}

// Ping ทดสอบการเชื่อมต่อ
func Ping(db *sql.DB) error {
	return db.Ping()
}

// Stats ดูข้อมูลสถิติการเชื่อมต่อ
func Stats(db *sql.DB) sql.DBStats {
	return db.Stats()
}

// BulkInsertConfig กำหนดค่าสำหรับ bulk insert
type BulkInsertConfig struct {
	BatchSize  int           // ขนาด batch (default: 10000)
	MaxRetries int           // จำนวน retry สูงสุด (default: 3)
	RetryDelay time.Duration // เวลารอระหว่าง retry (default: 100ms)
}

// DefaultBulkInsertConfig ค่า default สำหรับ bulk insert
var DefaultBulkInsertConfig = BulkInsertConfig{
	BatchSize:  10000,
	MaxRetries: 3,
	RetryDelay: 100 * time.Millisecond,
}

// BulkInsertWithCopy ใช้ COPY FROM เพื่อ insert ข้อมูลจำนวนมากอย่างมีประสิทธิภาพ
// tableName: ชื่อตาราง
// columns: รายชื่อคอลัมน์
// data: ข้อมูลที่จะ insert โดยแต่ละ row เป็น []any
func BulkInsertWithCopy(ctx context.Context, conn interface{}, tableName string, columns []string, data [][]any) error {
	return BulkInsertWithCopyConfig(ctx, conn, tableName, columns, data, DefaultBulkInsertConfig)
}

// BulkInsertWithCopyConfig ใช้ COPY FROM พร้อม config ที่กำหนดเอง
// รองรับ retry mechanism สำหรับ deadlock และ connection errors
func BulkInsertWithCopyConfig(ctx context.Context, conn interface{}, tableName string, columns []string, data [][]any, config BulkInsertConfig) error {
	if len(data) == 0 {
		return nil
	}

	// ใช้ค่า default ถ้าไม่ได้กำหนด
	if config.BatchSize <= 0 {
		config.BatchSize = DefaultBulkInsertConfig.BatchSize
	}
	if config.MaxRetries <= 0 {
		config.MaxRetries = DefaultBulkInsertConfig.MaxRetries
	}
	if config.RetryDelay <= 0 {
		config.RetryDelay = DefaultBulkInsertConfig.RetryDelay
	}

	// แบ่งข้อมูลเป็น batches ถ้าเกิน BatchSize
	totalRows := len(data)
	if totalRows > config.BatchSize {
		logger.Info("Splitting bulk insert into batches: %d rows -> %d batches of %d",
			totalRows, (totalRows+config.BatchSize-1)/config.BatchSize, config.BatchSize)

		for start := 0; start < totalRows; start += config.BatchSize {
			end := start + config.BatchSize
			if end > totalRows {
				end = totalRows
			}
			batch := data[start:end]

			if err := bulkInsertWithRetry(ctx, conn, tableName, columns, batch, config); err != nil {
				return fmt.Errorf("batch %d-%d failed: %w", start, end, err)
			}

			logger.Debug("Inserted batch %d-%d of %d", start, end, totalRows)
		}
		return nil
	}

	return bulkInsertWithRetry(ctx, conn, tableName, columns, data, config)
}

// bulkInsertWithRetry ทำ bulk insert พร้อม retry mechanism
func bulkInsertWithRetry(ctx context.Context, conn interface{}, tableName string, columns []string, data [][]any, config BulkInsertConfig) error {
	var lastErr error

	for attempt := 0; attempt < config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			delay := config.RetryDelay * time.Duration(1<<uint(attempt-1))
			logger.Warn("Retry attempt %d/%d for table %s after %v (error: %v)",
				attempt+1, config.MaxRetries, tableName, delay, lastErr)

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}

		err := bulkInsertSingleAttempt(ctx, conn, tableName, columns, data)
		if err == nil {
			return nil
		}

		lastErr = err

		// ตรวจสอบว่าเป็น error ที่ควร retry หรือไม่
		if !isRetryableError(err) {
			return err
		}
	}

	return fmt.Errorf("bulk insert failed after %d retries: %w", config.MaxRetries, lastErr)
}

// isRetryableError ตรวจสอบว่า error ควร retry หรือไม่
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	retryablePatterns := []string{
		"deadlock",
		"lock",
		"connection",
		"timeout",
		"too many clients",
		"server closed",
		"bad connection",
		"connection refused",
		"connection reset",
	}

	for _, pattern := range retryablePatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}

	return false
}

// bulkInsertSingleAttempt ทำ bulk insert ครั้งเดียว
func bulkInsertSingleAttempt(ctx context.Context, conn interface{}, tableName string, columns []string, data [][]any) error {
	switch c := conn.(type) {
	case *sql.Tx:
		return bulkInsertWithCopyUsingTx(ctx, c, tableName, columns, data)
	case *sql.DB:
		txn, err := c.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("error beginning transaction: %w", err)
		}
		if err := bulkInsertWithCopyUsingTx(ctx, txn, tableName, columns, data); err != nil {
			txn.Rollback()
			return err
		}
		return txn.Commit()
	default:
		return fmt.Errorf("unsupported connection type %T", conn)
	}
}

func bulkInsertWithCopyUsingTx(ctx context.Context, tx *sql.Tx, tableName string, columns []string, data [][]any) error {
	stmt, err := tx.PrepareContext(ctx, pq.CopyIn(tableName, columns...))
	if err != nil {
		return fmt.Errorf("error preparing COPY statement: %w", err)
	}
	defer stmt.Close()

	for i, row := range data {
		_, err = stmt.ExecContext(ctx, row...)
		if err != nil {
			logger.Error("แถว %d ล้มเหลว: %v", i, err)
			logger.Error("Row data: %+v", row)
			return fmt.Errorf("error executing COPY at row %d: %w", i, err)
		}
	}

	_, err = stmt.ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("error finalizing COPY: %w", err)
	}

	return nil
}

// Helper function to replace query parameters for debugging
func ReplaceQueryParams(query string, args ...any) string {
	result := query
	for i, arg := range args {
		placeholder := fmt.Sprintf("$%d", i+1)
		var value string
		switch v := arg.(type) {
		case string:
			value = fmt.Sprintf("'%s'", v)
		case nil:
			value = "NULL"
		default:
			value = fmt.Sprintf("%v", v)
		}
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}

func DatabaseIsReady(holdingCode string) bool {
	db, err := PgSqlFastConnect(holdingCode)
	if err != nil {
		return false
	}

	ctx := context.Background()
	err = db.PingContext(ctx)
	return err == nil
}

func GetStringValue(row map[string]any, key string) string {
	if val, ok := row[key].(string); ok {
		return val
	}
	return ""
}

func GetFloat64Value(row map[string]any, key string) float64 {
	if val, ok := row[key].(float64); ok {
		return val
	}
	if val, ok := row[key].(int64); ok {
		return float64(val)
	}
	return 0.0
}

func GetIntValue(row map[string]any, key string) int {
	if val, ok := row[key].(int64); ok {
		return int(val)
	}
	if val, ok := row[key].(int); ok {
		return val
	}
	return 0
}

func GetBoolValue(row map[string]any, key string) bool {
	if val, ok := row[key].(bool); ok {
		return val
	}
	return false
}

func GetTimeValue(row map[string]any, key string) time.Time {
	if val, ok := row[key].(time.Time); ok {
		return val
	}
	return time.Time{}
}

func GetInt64Value(row map[string]any, key string) int64 {
	if val, ok := row[key].(int64); ok {
		return val
	}
	return 0
}

// BulkUpsertWithCopy ใช้ COPY INTO temp table แล้ว UPSERT ไปยัง target table
// เพื่อป้องกัน data loss จาก DELETE+INSERT pattern
func BulkUpsertWithCopy(ctx context.Context, db *sql.DB, tableName string, columns []string, data [][]any, conflictColumns []string) error {
	if len(data) == 0 {
		return nil
	}

	txn, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
	}
	defer txn.Rollback()

	// สร้าง temp table ชื่อพิเศษ
	tempTableName := fmt.Sprintf("temp_%s_%d", tableName, time.Now().UnixNano())

	// สร้าง temp table ที่มี schema เหมือนกับ target table
	createTempTable := fmt.Sprintf(`
		CREATE TEMP TABLE %s (LIKE %s INCLUDING DEFAULTS)
		ON COMMIT DROP
	`, tempTableName, tableName)

	_, err = txn.ExecContext(ctx, createTempTable)
	if err != nil {
		return fmt.Errorf("error creating temp table: %w", err)
	}

	// COPY ข้อมูลเข้า temp table
	stmt, err := txn.PrepareContext(ctx, pq.CopyIn(tempTableName, columns...))
	if err != nil {
		return fmt.Errorf("error preparing COPY statement: %w", err)
	}
	defer stmt.Close()

	for i, row := range data {
		_, err = stmt.ExecContext(ctx, row...)
		if err != nil {
			logger.Error("แถว %d ล้มเหลว: %v", i, err)
			logger.Error("Row data: %+v", row)
			return fmt.Errorf("error executing COPY at row %d: %w", i, err)
		}
	}

	// Finalize COPY
	_, err = stmt.ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("error finalizing COPY: %w", err)
	}

	// สร้าง UPSERT query (INSERT ... ON CONFLICT ... DO UPDATE)
	columnsList := strings.Join(columns, ", ")

	// สร้าง excluded values สำหรับ UPDATE SET
	updateParts := make([]string, 0, len(columns))
	for _, col := range columns {
		// ไม่ update conflict columns
		isConflict := false
		for _, confCol := range conflictColumns {
			if col == confCol {
				isConflict = true
				break
			}
		}
		if !isConflict {
			updateParts = append(updateParts, fmt.Sprintf("%s = EXCLUDED.%s", col, col))
		}
	}

	conflictList := strings.Join(conflictColumns, ", ")
	updateSet := strings.Join(updateParts, ", ")

	upsertQuery := fmt.Sprintf(`
		INSERT INTO %s (%s)
		SELECT %s FROM %s
		ON CONFLICT (%s) DO UPDATE SET %s
	`, tableName, columnsList, columnsList, tempTableName, conflictList, updateSet)

	_, err = txn.ExecContext(ctx, upsertQuery)
	if err != nil {
		return fmt.Errorf("error executing UPSERT: %w", err)
	}

	// Commit transaction
	return txn.Commit()
}
