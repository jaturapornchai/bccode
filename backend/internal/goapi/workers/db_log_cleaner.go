package workers

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/myclickhouse"
	"smlcloudplatform/internal/goapi/mydb"
)

// DBLogCleaner - Worker สำหรับทำความสะอาด database logs
type DBLogCleaner struct {
	interval time.Duration
	stopCh   chan struct{}
	wg       sync.WaitGroup
	running  bool
	mu       sync.Mutex
}

// NewDBLogCleaner - สร้าง DBLogCleaner ใหม่
func NewDBLogCleaner(interval time.Duration) *DBLogCleaner {
	if interval < time.Hour {
		interval = time.Hour // Minimum 1 hour
	}
	return &DBLogCleaner{
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

// Start - เริ่ม worker
func (c *DBLogCleaner) Start() {
	c.mu.Lock()
	if c.running {
		c.mu.Unlock()
		return
	}
	c.running = true
	c.mu.Unlock()

	c.wg.Add(1)
	go c.run()
	logger.Info("DBLogCleaner started (interval: %v)", c.interval)
}

// Stop - หยุด worker
func (c *DBLogCleaner) Stop() {
	c.mu.Lock()
	if !c.running {
		c.mu.Unlock()
		return
	}
	c.running = false
	c.mu.Unlock()

	close(c.stopCh)
	c.wg.Wait()
	logger.Info("DBLogCleaner stopped")
}

func (c *DBLogCleaner) run() {
	defer c.wg.Done()

	// Run immediately on start
	c.cleanup()

	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanup()
		case <-c.stopCh:
			return
		}
	}
}

func (c *DBLogCleaner) cleanup() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Clean ClickHouse logs
	go c.cleanClickHouseLogs(ctx)

	// Note: PostgreSQL WAL cleanup requires superuser and is usually
	// handled by the database server itself via autovacuum and checkpoint
}

// cleanClickHouseLogs - ทำความสะอาด ClickHouse system logs
func (c *DBLogCleaner) cleanClickHouseLogs(ctx context.Context) {
	conn, err := myclickhouse.ClickHouseFastConnect()
	if err != nil {
		logger.Warn("DBLogCleaner: Cannot connect to ClickHouse: %v", err)
		return
	}

	// TTL queries - set retention period for system logs
	ttlQueries := []struct {
		table string
		days  int
	}{
		{"system.query_log", 7},
		{"system.query_thread_log", 3},
		{"system.part_log", 7},
		{"system.trace_log", 1},
		{"system.metric_log", 3},
		{"system.asynchronous_metric_log", 1},
	}

	for _, q := range ttlQueries {
		query := fmt.Sprintf("ALTER TABLE %s MODIFY TTL event_date + INTERVAL %d DAY", q.table, q.days)
		if err := conn.Exec(ctx, query); err != nil {
			// Ignore errors - table might not exist or TTL already set
			continue
		}
	}

	// Optimize tables to reclaim space (run less frequently)
	optimizeQueries := []string{
		"OPTIMIZE TABLE system.query_log FINAL",
		"OPTIMIZE TABLE system.query_thread_log FINAL",
	}

	for _, query := range optimizeQueries {
		if err := conn.Exec(ctx, query); err != nil {
			continue
		}
	}

	logger.Debug("DBLogCleaner: ClickHouse logs TTL configured")
}

// CleanupPostgreSQLForShop - ทำความสะอาด PostgreSQL สำหรับ shop
// เรียกใช้เมื่อต้องการ manual cleanup
func CleanupPostgreSQLForShop(holdingCode string) error {
	db, err := mydb.GetGlobalConnectionFromPool(holdingCode)
	if err != nil {
		return fmt.Errorf("cannot connect to PostgreSQL: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Run VACUUM ANALYZE on main tables
	tables := []string{"doc", "docdetail", "docref", "docpayment"}
	for _, table := range tables {
		query := fmt.Sprintf("VACUUM ANALYZE %s", table)
		if _, err := db.ExecContext(ctx, query); err != nil {
			// Continue on error
			continue
		}
	}

	return nil
}

// GetClickHouseLogStats - ดึงสถิติ system logs ของ ClickHouse
func GetClickHouseLogStats() (map[string]interface{}, error) {
	conn, err := myclickhouse.ClickHouseFastConnect()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := `
		SELECT
			database,
			table,
			sum(bytes_on_disk) as size_bytes,
			sum(rows) as total_rows
		FROM system.parts
		WHERE database = 'system' AND table LIKE '%log%'
		GROUP BY database, table
		ORDER BY sum(bytes_on_disk) DESC
	`

	rows, err := conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make(map[string]interface{})
	var totalSize uint64

	for rows.Next() {
		var database, table string
		var sizeBytes, totalRows uint64
		if err := rows.Scan(&database, &table, &sizeBytes, &totalRows); err != nil {
			continue
		}
		stats[table] = map[string]interface{}{
			"size_bytes": sizeBytes,
			"size_mb":    float64(sizeBytes) / 1024 / 1024,
			"rows":       totalRows,
		}
		totalSize += sizeBytes
	}

	stats["_total_size_mb"] = float64(totalSize) / 1024 / 1024

	return stats, nil
}

// GetPostgreSQLStats - ดึงสถิติ dead tuples ของ PostgreSQL
func GetPostgreSQLStats(db *sql.DB) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := `
		SELECT
			schemaname,
			relname as table_name,
			pg_total_relation_size(schemaname || '.' || relname) as total_size,
			n_dead_tup as dead_tuples,
			n_live_tup as live_tuples
		FROM pg_stat_user_tables
		WHERE n_dead_tup > 100
		ORDER BY n_dead_tup DESC
		LIMIT 10
	`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make(map[string]interface{})
	tables := []map[string]interface{}{}

	for rows.Next() {
		var schema, tableName string
		var totalSize, deadTuples, liveTuples int64
		if err := rows.Scan(&schema, &tableName, &totalSize, &deadTuples, &liveTuples); err != nil {
			continue
		}
		tables = append(tables, map[string]interface{}{
			"table":       tableName,
			"size_mb":     float64(totalSize) / 1024 / 1024,
			"dead_tuples": deadTuples,
			"live_tuples": liveTuples,
		})
	}

	stats["tables"] = tables
	return stats, nil
}
