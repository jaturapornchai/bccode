package myglobal

import (
	"database/sql"
	"fmt"
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

// TransFlags ที่มีการคำนวณต้นทุน (ตาม calccost.md)
// เพิ่มจำนวนคงเหลือ: 54(ยกมา), 12=ซื้อสินค้า, 310=รับสินค้า(พาเชียล), 48=รับคืน, 60=รับเข้า(สำเร็จรูป), 58=รับคืนจากการเบิก, 66=ปรับสต็อก(เพิ่ม)
// ลดจำนวนคงเหลือ: 44=ขายสินค้า, 16=ส่งคืนสินค้า, 56=เบิกออก, 68=ปรับสต็อก(ลด), 72=โอนย้าย
// หมายเหตุ: 866(ปรับปรุงต้นทุนเพิ่ม), 868(ปรับปรุงต้นทุนลด) ไม่คำนวณต้นทุน

var TransFlagsToProcess = []int{54, 12, 310, 48, 60, 58, 66, 44, 16, 56, 68, 72}

// Global Database Manager (สำหรับ unified database management)
var (
	globalDBManager      *sql.DB
	globalDBManagerMu    sync.RWMutex
	globalDBProviderFunc func() (*sql.DB, error)
)

// GetTransFlagsForQuery สร้าง string สำหรับใช้ใน SQL IN clause จาก TransFlagsToProcess
// ตัวอย่าง: "12,16,44,48,54,56,58,60,66,68,72,310,866,868"
func GetTransFlagsForQuery() string {
	transFlagStrings := make([]string, len(TransFlagsToProcess))
	for i, flag := range TransFlagsToProcess {
		transFlagStrings[i] = fmt.Sprintf("%d", flag)
	}
	return strings.Join(transFlagStrings, ",")
}

// IsCalcStockTransFlag ตรวจสอบว่า TransFlag นี้มีการคำนวณสต็อกหรือไม่
// คืนค่า 1 ถ้ามีการคำนวณสต็อก, 0 ถ้าไม่มี
func IsCalcStockTransFlag(transFlag int) int8 {
	switch transFlag {
	case 12, 16, 44, 48, 54, 56, 58, 60, 66, 68, 72, 310, 866, 868:
		// ซื้อ, ส่งคืน, ขาย, รับสินค้า, ยกมา, เบิก, รับคืน, รับเข้า, ปรับสต็อก, โอนย้าย, รับพาเชียล, ผลิต
		return 1
	default:
		// เอกสารอื่นๆ เช่น ใบสั่งซื้อ (10), ใบสั่งขาย (14) ไม่มีการคำนวณสต็อก
		return 0
	}
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
