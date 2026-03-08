package mydb

import (
	"database/sql"
	"fmt"
	"sync"

	"smlcloudplatform/internal/goapi/logger"
)

// ManagerPool จัดการ DatabaseManager หลายๆ ตัว (หนึ่งตัวต่อ shop)
type ManagerPool struct {
	managers map[string]*DatabaseManager
	mu       sync.RWMutex
	config   DatabaseManagerConfig
}

// DatabaseManagerConfig - การตั้งค่าสำหรับ DatabaseManager
type DatabaseManagerConfig struct {
	PostgreSQLHost     string
	PostgreSQLPort     string
	PostgreSQLUser     string
	PostgreSQLPassword string
	PostgreSQLSSLMode  string

	ClickHouseHost     string
	ClickHousePort     string
	ClickHouseUser     string
	ClickHousePassword string
}

var (
	globalManagerPool *ManagerPool
	globalPoolOnce    sync.Once
)

// InitManagerPool - เริ่มต้น global manager pool
func InitManagerPool(config DatabaseManagerConfig) {
	globalPoolOnce.Do(func() {
		globalManagerPool = &ManagerPool{
			managers: make(map[string]*DatabaseManager),
			config:   config,
		}
		logger.Success("✅ Global Manager Pool เริ่มต้นเรียบร้อย")
	})
}

// GetGlobalManagerPool - ดึง global manager pool (สร้างอัตโนมัติหากยังไม่มี)
func GetGlobalManagerPool() *ManagerPool {
	if globalManagerPool == nil {
		// Auto-initialize with default config from environment
		defaultConfig := DatabaseManagerConfig{
			PostgreSQLHost:     getEnv("POSTGRES_HOST", "localhost"),
			PostgreSQLPort:     getEnv("POSTGRES_PORT", "5432"),
			PostgreSQLUser:     getEnv("POSTGRES_USER", "postgres"),
			PostgreSQLPassword: getEnv("POSTGRES_PASSWORD", ""),
			PostgreSQLSSLMode:  getEnv("POSTGRES_SSL_MODE", "disable"),

			ClickHouseHost:     getEnv("CLICKHOUSE_HOST", "localhost"),
			ClickHousePort:     getEnv("CLICKHOUSE_PORT", "9000"),
			ClickHouseUser:     getEnv("CLICKHOUSE_USER", "default"),
			ClickHousePassword: getEnv("CLICKHOUSE_PASSWORD", ""),
		}
		InitManagerPool(defaultConfig)
	}
	return globalManagerPool
}

// GetManager - ดึง DatabaseManager สำหรับ shop (สร้างใหม่หากยังไม่มี)
func (p *ManagerPool) GetManager(shopId string) (*DatabaseManager, error) {
	// ลอง read lock ก่อน
	p.mu.RLock()
	if manager, exists := p.managers[shopId]; exists {
		p.mu.RUnlock()

		// ตรวจสอบสุขภาพ
		if manager.IsHealthy() {
			return manager, nil
		}

		// ถ้า unhealthy ให้ลบออกและสร้างใหม่
		logger.Warn("DatabaseManager สำหรับ shop=%s ไม่ healthy, จะสร้างใหม่", shopId)
		p.mu.Lock()
		delete(p.managers, shopId)
		p.mu.Unlock()
	} else {
		p.mu.RUnlock()
	}

	// สร้าง manager ใหม่ (ใช้ write lock)
	p.mu.Lock()
	defer p.mu.Unlock()

	// Double-check หลัง lock
	if manager, exists := p.managers[shopId]; exists {
		return manager, nil
	}

	// สร้าง DatabaseConfig สำหรับ shop นี้
	dbConfig := DatabaseConfig{
		PostgreSQLHost:     p.config.PostgreSQLHost,
		PostgreSQLPort:     p.config.PostgreSQLPort,
		PostgreSQLUser:     p.config.PostgreSQLUser,
		PostgreSQLPassword: p.config.PostgreSQLPassword,
		PostgreSQLDatabase: shopId,
		PostgreSQLSSLMode:  p.config.PostgreSQLSSLMode,

		ClickHouseHost:     p.config.ClickHouseHost,
		ClickHousePort:     p.config.ClickHousePort,
		ClickHouseUser:     p.config.ClickHouseUser,
		ClickHousePassword: p.config.ClickHousePassword,
		ClickHouseDatabase: shopId,
	}

	logger.Info("สร้าง DatabaseManager ใหม่สำหรับ shop=%s", shopId)
	manager, err := NewDatabaseManager(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create DatabaseManager for shop=%s: %w", shopId, err)
	}

	p.managers[shopId] = manager
	logger.Success("✅ สร้าง DatabaseManager สำหรับ shop=%s เรียบร้อย (รวม %d shops)", shopId, len(p.managers))

	return manager, nil
}

// GetConnection - ดึง PostgreSQL connection สำหรับ shop
func (p *ManagerPool) GetConnection(shopId string) (*sql.DB, error) {
	manager, err := p.GetManager(shopId)
	if err != nil {
		return nil, err
	}
	return manager.GetPostgreSQLConnection()
}

// CloseManager - ปิด manager สำหรับ shop เฉพาะ
func (p *ManagerPool) CloseManager(shopId string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if manager, exists := p.managers[shopId]; exists {
		if err := manager.Close(); err != nil {
			return fmt.Errorf("failed to close manager for shop=%s: %w", shopId, err)
		}
		delete(p.managers, shopId)
		logger.Info("ปิด DatabaseManager สำหรับ shop=%s", shopId)
	}

	return nil
}

// CloseAll - ปิด managers ทั้งหมด
func (p *ManagerPool) CloseAll() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	var errors []error
	for shopId, manager := range p.managers {
		if err := manager.Close(); err != nil {
			errors = append(errors, fmt.Errorf("shop=%s: %w", shopId, err))
		} else {
			logger.Info("ปิด DatabaseManager สำหรับ shop=%s", shopId)
		}
	}

	p.managers = make(map[string]*DatabaseManager)

	if len(errors) > 0 {
		return fmt.Errorf("errors closing managers: %v", errors)
	}

	logger.Success("ปิด DatabaseManagers ทั้งหมด (%d shops)", len(p.managers))
	return nil
}

// GetStats - ดึงสถิติของ pool
func (p *ManagerPool) GetStats() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	stats := map[string]interface{}{
		"total_managers": len(p.managers),
		"shops":          make([]string, 0, len(p.managers)),
	}

	for shopId := range p.managers {
		stats["shops"] = append(stats["shops"].([]string), shopId)
	}

	return stats
}

// IsHealthy - ตรวจสอบว่า manager healthy หรือไม่
func (dm *DatabaseManager) IsHealthy() bool {
	pingResults := dm.PingAll()

	// ต้อง PostgreSQL healthy อย่างน้อย
	if err := pingResults["PostgreSQL"]; err != nil {
		return false
	}

	return true
}

// Global helper functions สำหรับ backward compatibility
// ฟังก์ชันเหล่านี้จะถูกเรียกจาก mypg.PgSqlFastConnectV2()

// GetGlobalConnection - ดึง PostgreSQL connection สำหรับ shop
// ใช้ global manager pool
func GetGlobalConnectionFromPool(shopId string) (*sql.DB, error) {
	pool := GetGlobalManagerPool()
	return pool.GetConnection(shopId)
}

// GetGlobalManagerFromPool - ดึง DatabaseManager สำหรับ shop
// ใช้ global manager pool
func GetGlobalManagerFromPool(shopId string) (*DatabaseManager, error) {
	pool := GetGlobalManagerPool()
	return pool.GetManager(shopId)
}

// CloseAllManagers - ปิด managers ทั้งหมดใน global pool
func CloseAllManagers() error {
	if globalManagerPool != nil {
		return globalManagerPool.CloseAll()
	}
	return nil
}
