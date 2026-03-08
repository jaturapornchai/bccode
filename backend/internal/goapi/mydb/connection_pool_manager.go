package mydb

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sync"
	"time"

	"smlcloudplatform/internal/goapi/logger"

	_ "github.com/lib/pq"
)

// ConnectionPoolManager จัดการ connection pools แบบขั้นสูง
type ConnectionPoolManager struct {
	pools map[string]*sql.DB
	mutex sync.RWMutex
}

// NewConnectionPoolManager สร้าง connection pool manager ใหม่
func NewConnectionPoolManager() *ConnectionPoolManager {
	return &ConnectionPoolManager{
		pools: make(map[string]*sql.DB),
	}
}

// PoolConfig การตั้งค่า pool
type PoolConfig struct {
	MaxOpenConns     int           // จำนวน connection เปิดสูงสุด
	MaxIdleConns     int           // จำนวน idle connection สูงสุด
	ConnMaxLifetime  time.Duration // อายุของ connection
	ConnMaxIdleTime  time.Duration // เวลา idle สูงสุด
	HealthCheck      bool          // เปิด health check
}

// DefaultHighTrafficConfig การตั้งค่าสำหรับ high traffic (ลดลงครึ่งหนึ่งเพื่อไม่กระทบระบบอื่น)
var DefaultHighTrafficConfig = PoolConfig{
	MaxOpenConns:    75,     // ลดลงครึ่งหนึ่ง
	MaxIdleConns:    25,     // ลดลงครึ่งหนึ่ง
	ConnMaxLifetime: 30 * time.Minute, // ลดจาก 1 ชั่วโมงเป็น 30 นาที
	ConnMaxIdleTime: 10 * time.Minute, // ลดจาก 15 เป็น 10 นาที
	HealthCheck:     true,
}

// GetConnection ดึง connection จาก pool
func (pm *ConnectionPoolManager) GetConnection(databaseName string) (*sql.DB, error) {
	pm.mutex.RLock()
	if db, exists := pm.pools[databaseName]; exists {
		pm.mutex.RUnlock()
		return db, pm.healthCheck(db)
	}
	pm.mutex.RUnlock()

	// สร้าง connection ใหม่
	return pm.createPool(databaseName)
}

// createPool สร้าง pool ใหม่
func (pm *ConnectionPoolManager) createPool(databaseName string) (*sql.DB, error) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	// Double check หลังจาก lock
	if db, exists := pm.pools[databaseName]; exists {
		return db, nil
	}

	// สร้าง connection string
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s "+
			"connect_timeout=5 "+
			"statement_timeout=300000 "+
			"idle_in_transaction_session_timeout=60000",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		databaseName,
		os.Getenv("POSTGRES_SSL_MODE"),
	)

	logger.Info("Creating new connection pool for database: %s", databaseName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// ตั้งค่า pool ตาม high traffic config
	config := DefaultHighTrafficConfig
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)
	db.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	// ทดสอบการเชื่อมต่อ
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	pm.pools[databaseName] = db

	logger.Success("Created connection pool for %s (max_open=%d, max_idle=%d, lifetime=%v)", 
		databaseName, config.MaxOpenConns, config.MaxIdleConns, config.ConnMaxLifetime)

	return db, nil
}

// healthCheck ตรวจสอบสุขภาพของ connection
func (pm *ConnectionPoolManager) healthCheck(db *sql.DB) error {
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("database connection unhealthy: %w", err)
	}

	return nil
}

// ClosePool ปิด pool เฉพาะ
func (pm *ConnectionPoolManager) ClosePool(databaseName string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if db, exists := pm.pools[databaseName]; exists {
		if err := db.Close(); err != nil {
			return fmt.Errorf("failed to close pool for %s: %w", databaseName, err)
		}
		delete(pm.pools, databaseName)
		logger.Info("Closed connection pool for database: %s", databaseName)
	}

	return nil
}

// CloseAll ปิด pools ทั้งหมด
func (pm *ConnectionPoolManager) CloseAll() error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	var errors []error
	for databaseName, db := range pm.pools {
		if err := db.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close pool for %s: %w", databaseName, err))
		} else {
			logger.Info("Closed connection pool for database: %s", databaseName)
		}
	}

	pm.pools = make(map[string]*sql.DB)

	if len(errors) > 0 {
		return fmt.Errorf("errors closing pools: %v", errors)
	}

	logger.Success("Closed all connection pools")
	return nil
}

// GetStats ดึงสถิติของ pools
func (pm *ConnectionPoolManager) GetStats() map[string]interface{} {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	stats := make(map[string]interface{})
	stats["total_pools"] = len(pm.pools)
	stats["pools"] = make(map[string]interface{})

	for databaseName, db := range pm.pools {
		dbStats := db.Stats()
		stats["pools"].(map[string]interface{})[databaseName] = map[string]interface{}{
			"max_open_conns":      dbStats.MaxOpenConnections,
			"open_conns":          dbStats.OpenConnections,
			"idle_conns":          0, // sql.DBStats ไม่มี field นี้
			"in_use_conns":        dbStats.InUse,
			"wait_count":          dbStats.WaitCount,
			"wait_duration":       dbStats.WaitDuration,
			"max_idle_time":       0, // sql.DBStats ไม่มี field นี้
			"max_lifetime":        0, // sql.DBStats ไม่มี field นี้
		}
	}

	return stats
}

// CleanupIdleConnections ทำความสะอาด idle connections
func (pm *ConnectionPoolManager) CleanupIdleConnections() {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	logger.Info("Starting idle connection cleanup for %d pools", len(pm.pools))

	for databaseName, db := range pm.pools {
		// Force close idle connections
		// Note: sql.DB doesn't have a direct method to close idle connections
		// but setting MaxIdleConns to 0 will prevent new idle connections
		oldMaxIdle := 30 // ค่าเริ่มต้น
		
		db.SetMaxIdleConns(0)

		// Restore original setting
		db.SetMaxIdleConns(oldMaxIdle)
		
		logger.Debug("Cleaned idle connections for database: %s", databaseName)
	}

	logger.Success("Completed idle connection cleanup")
}

// PeriodicMaintenance บำรุงรักษาเป็นระยะ
func (pm *ConnectionPoolManager) PeriodicMaintenance(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			logger.Info("Running periodic connection pool maintenance...")
			
			// Cleanup idle connections
			pm.CleanupIdleConnections()
			
			// Log statistics
			stats := pm.GetStats()
			logger.Info("Connection pool statistics: %+v", stats)
		}
	}
}

// DatabasePoolOptimizer optimizer สำหรับ connection pools
type DatabasePoolOptimizer struct {
	pm           *ConnectionPoolManager
	config       PoolConfig
	optimization bool
}

// NewDatabasePoolOptimizer สร้าง optimizer ใหม่
func NewDatabasePoolOptimizer() *DatabasePoolOptimizer {
	return &DatabasePoolOptimizer{
		pm:           NewConnectionPoolManager(),
		config:       DefaultHighTrafficConfig,
		optimization: true,
	}
}

// OptimizePool ปรับปรุงประสิทธิภาพ pool
func (opo *DatabasePoolOptimizer) OptimizePool(databaseName string) error {
	if !opo.optimization {
		return nil
	}

	db, err := opo.pm.GetConnection(databaseName)
	if err != nil {
		return fmt.Errorf("failed to get connection: %w", err)
	}

	stats := db.Stats()
	
	// ปรับปรุง connection limits ตามการใช้งาน
	if stats.OpenConnections < stats.MaxOpenConnections/2 {
		// ลดจำนวน max connections หากไม่ได้ใช้เต็ม
		newMaxOpen := stats.MaxOpenConnections / 2
		if newMaxOpen > 10 { // ขั้นต่ำ 10 connections
			db.SetMaxOpenConns(newMaxOpen)
			logger.Info("Reduced MaxOpenConns for %s from %d to %d",
				databaseName, stats.MaxOpenConnections, newMaxOpen)
		}
	}

	// ปรับปรุง idle time
	if stats.OpenConnections > 20 {
		// ลด idle time หากมี connections เยอะ
		newIdleTime := 5 * time.Minute
		db.SetConnMaxIdleTime(newIdleTime)
		logger.Info("Reduced ConnMaxIdleTime for %s to %v", databaseName, newIdleTime)
	}

	return nil
}

// GetOptimizerStats ดึงสถิติของ optimizer
func (opo *DatabasePoolOptimizer) GetOptimizerStats() map[string]interface{} {
	return map[string]interface{}{
		"optimization_enabled": opo.optimization,
		"pool_stats":          opo.pm.GetStats(),
		"config": map[string]interface{}{
			"max_open_conns":     opo.config.MaxOpenConns,
			"max_idle_conns":     opo.config.MaxIdleConns,
			"conn_max_lifetime":  opo.config.ConnMaxLifetime,
			"conn_max_idle_time": opo.config.ConnMaxIdleTime,
			"health_check":       opo.config.HealthCheck,
		},
	}
}