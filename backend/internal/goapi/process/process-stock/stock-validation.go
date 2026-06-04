package processstock

import (
	"context"
	"database/sql"
	"fmt"
	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/mypg"
)

// ValidateStockBeforeSale ตรวจสอบสต็อกคงเหลือก่อนขาย
func ValidateStockBeforeSale(ctx context.Context, db *sql.DB, holdingCode int, itemCode string, qtyToSell float64) error {
	// Query สต็อกปัจจุบันจาก docdetail
	query := `
		SELECT COALESCE(SUM(totalqty * CASE
			WHEN calcflag = 1 THEN 1
			WHEN calcflag = -1 THEN -1
			ELSE 0
		END), 0) as current_stock
		FROM docdetail
		WHERE holding_code = $1
		  AND itemcode = $2
		  AND iscalcstock = 1
	`

	rows, err := mypg.QuerySelectAll(db, query, holdingCode, itemCode)
	if err != nil {
		return fmt.Errorf("ไม่สามารถตรวจสอบสต็อกได้: %w", err)
	}

	if len(rows) == 0 {
		return fmt.Errorf("ไม่พบข้อมูลสินค้า: %s", itemCode)
	}

	currentStock := mypg.GetFloat64Value(rows[0], "current_stock")

	logger.Info("📦 Stock validation for %s: current=%f, requested=%f", itemCode, currentStock, qtyToSell)

	// ตรวจสอบว่าสต็อกพอหรือไม่
	if currentStock < qtyToSell {
		return fmt.Errorf(
			"⚠️ สต็อกไม่เพียงพอสำหรับสินค้า %s: มีอยู่ %.2f แต่ต้องการ %.2f (ขาดอีก %.2f)",
			itemCode, currentStock, qtyToSell, qtyToSell-currentStock,
		)
	}

	// ตรวจสอบว่าจะทำให้สต็อกติดลบหรือไม่
	remainingStock := currentStock - qtyToSell
	if remainingStock < 0 {
		return fmt.Errorf(
			"⚠️ การขายจะทำให้สต็อกติดลบ: สินค้า %s จะเหลือ %.2f",
			itemCode, remainingStock,
		)
	}

	logger.Info("✅ Stock validation passed for %s: remaining will be %.2f", itemCode, remainingStock)
	return nil
}

// ValidateStockForMultipleItems ตรวจสอบสต็อกสำหรับหลายรายการพร้อมกัน
func ValidateStockForMultipleItems(ctx context.Context, db *sql.DB, holdingCode int, items []StockValidationItem) []error {
	var errors []error

	for _, item := range items {
		if item.Qty <= 0 {
			errors = append(errors, fmt.Errorf("จำนวนสินค้า %s ต้องมากกว่า 0 (ได้รับ: %.2f)", item.ItemCode, item.Qty))
			continue
		}

		err := ValidateStockBeforeSale(ctx, db, holdingCode, item.ItemCode, item.Qty)
		if err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

// StockValidationItem โครงสร้างสำหรับตรวจสอบสต็อก
type StockValidationItem struct {
	ItemCode string
	Qty      float64
}

// GetCurrentStock ดึงสต็อกปัจจุบันของสินค้า
func GetCurrentStock(ctx context.Context, db *sql.DB, holdingCode int, itemCode string) (float64, error) {
	query := `
		SELECT COALESCE(SUM(totalqty * CASE
			WHEN calcflag = 1 THEN 1
			WHEN calcflag = -1 THEN -1
			ELSE 0
		END), 0) as current_stock
		FROM docdetail
		WHERE holding_code = $1
		  AND itemcode = $2
		  AND iscalcstock = 1
	`

	rows, err := mypg.QuerySelectAll(db, query, holdingCode, itemCode)
	if err != nil {
		return 0, fmt.Errorf("ไม่สามารถดึงข้อมูลสต็อกได้: %w", err)
	}

	if len(rows) == 0 {
		return 0, nil // ไม่มีสต็อก
	}

	currentStock := mypg.GetFloat64Value(rows[0], "current_stock")
	return currentStock, nil
}

// CheckNegativeStock ตรวจสอบสินค้าที่มีสต็อกติดลบในระบบ
func CheckNegativeStock(ctx context.Context, db *sql.DB, holdingCode int) ([]NegativeStockItem, error) {
	query := `
		SELECT
			itemcode,
			SUM(totalqty * CASE
				WHEN calcflag = 1 THEN 1
				WHEN calcflag = -1 THEN -1
				ELSE 0
			END) as current_stock
		FROM docdetail
		WHERE holding_code = $1
		  AND iscalcstock = 1
		GROUP BY itemcode
		HAVING SUM(totalqty * CASE
			WHEN calcflag = 1 THEN 1
			WHEN calcflag = -1 THEN -1
			ELSE 0
		END) < 0
		ORDER BY current_stock ASC
	`

	rows, err := mypg.QuerySelectAll(db, query, holdingCode)
	if err != nil {
		return nil, fmt.Errorf("ไม่สามารถตรวจสอบสต็อกติดลบได้: %w", err)
	}

	var negativeItems []NegativeStockItem
	for _, row := range rows {
		negativeItems = append(negativeItems, NegativeStockItem{
			ItemCode:     mypg.GetStringValue(row, "itemcode"),
			CurrentStock: mypg.GetFloat64Value(row, "current_stock"),
		})
	}

	if len(negativeItems) > 0 {
		logger.Warn("⚠️ พบสินค้าสต็อกติดลบ %d รายการใน shop %d", len(negativeItems), holdingCode)
		for _, item := range negativeItems {
			logger.Warn("  - %s: %.2f", item.ItemCode, item.CurrentStock)
		}
	}

	return negativeItems, nil
}

// NegativeStockItem โครงสร้างสำหรับสินค้าที่สต็อกติดลบ
type NegativeStockItem struct {
	ItemCode     string
	CurrentStock float64
}
