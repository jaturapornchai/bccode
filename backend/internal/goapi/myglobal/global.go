package myglobal

import (
	"database/sql"
	"fmt"
	"smlcloudplatform/internal/doctype"
	"smlcloudplatform/internal/goapi/logger"
	"strings"
	"sync"
)

type ConfigSystemModel struct {
	WareHouseCostType int    // 1=รวมคลัง, 2=แยกคลัง
	StockQtyPoint     int    // ทศนิยมจำนวน
	StockAmountPoint  int    // ทศนิยมมูลค่า
	StockCostPoint    int    // ทศนิยมต้นทุน
	DefaultLanguage   string // ภาษาที่ใช้แสดงชื่อ เช่น "th", "en"
}

var ConfigSystem ConfigSystemModel = ConfigSystemModel{
	WareHouseCostType: 1,
	StockQtyPoint:     2,
	StockAmountPoint:  2,
	StockCostPoint:    2,
	DefaultLanguage:   "th",
}

// TransFlagsToProcess คือประเภทเอกสารที่เครื่องคิดต้นทุนนำมาคิด มาจากทะเบียน doctype ที่เดียว
var TransFlagsToProcess = doctype.StockFlags()

// Global Database Manager (สำหรับ unified database management)
var (
	globalDBManager      *sql.DB
	globalDBManagerMu    sync.RWMutex
	globalDBProviderFunc func() (*sql.DB, error)
)

// GetTransFlagsForQuery สร้าง string สำหรับใช้ใน SQL IN clause จาก TransFlagsToProcess
func GetTransFlagsForQuery() string {
	transFlagStrings := make([]string, len(TransFlagsToProcess))
	for i, flag := range TransFlagsToProcess {
		transFlagStrings[i] = fmt.Sprintf("%d", flag)
	}
	return strings.Join(transFlagStrings, ",")
}

// SetGlobalDatabaseConnection - ตั้งค่า global database connection
func SetGlobalDatabaseConnection(db *sql.DB) {
	globalDBManagerMu.Lock()
	defer globalDBManagerMu.Unlock()
	globalDBManager = db
}

// SetGlobalDatabaseProvider - กำหนด provider สำหรับสร้าง connection ใหม่เมื่อจำเป็น
func SetGlobalDatabaseProvider(provider func() (*sql.DB, error)) {
	globalDBManagerMu.Lock()
	defer globalDBManagerMu.Unlock()
	globalDBProviderFunc = provider
}

// RefreshGlobalDatabaseConnection - ดึง connection ใหม่ผ่าน provider และอัปเดต global connection
func RefreshGlobalDatabaseConnection() (*sql.DB, error) {
	globalDBManagerMu.Lock()
	defer globalDBManagerMu.Unlock()

	if globalDBProviderFunc == nil {
		return nil, fmt.Errorf("global database provider not configured")
	}

	logger.Warn("รีเฟรชการเชื่อมต่อ Global Database queue...")
	newDB, err := globalDBProviderFunc()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh global database connection: %w", err)
	}

	globalDBManager = newDB
	logger.Success("✅ รีเฟรชการเชื่อมต่อ Global Database สำเร็จ")
	return globalDBManager, nil
}

// GetGlobalDatabaseConnection - ดึง global database connection
func GetGlobalDatabaseConnection() (*sql.DB, error) {
	globalDBManagerMu.RLock()
	db := globalDBManager
	providerConfigured := globalDBProviderFunc != nil
	globalDBManagerMu.RUnlock()

	if db != nil {
		return db, nil
	}

	if providerConfigured {
		return RefreshGlobalDatabaseConnection()
	}

	return nil, fmt.Errorf("global database manager not initialized")
}
