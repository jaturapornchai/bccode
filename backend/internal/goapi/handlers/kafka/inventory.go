package kafka

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/models"
	"smlcloudplatform/internal/goapi/myglobal"
	"smlcloudplatform/internal/goapi/mypg"
	build "smlcloudplatform/internal/goapi/process/build"
	processstock "smlcloudplatform/internal/goapi/process/process-stock"
)

// OnConsumeMessageInventoryCreateOrUpdate - handles inventory create/update messages
func OnConsumeMessageInventoryCreateOrUpdate(msg string) error {
	logger.Debug("OnConsumeMessageInventoryCreateOrUpdate: %s", msg)
	ProductBarcodeBuild(msg)
	return nil
}

// OnConsumeMessageInventoryDelete - handles inventory delete messages
func OnConsumeMessageInventoryDelete(msg string) error {
	logger.Info("OnConsumeMessageInventoryDelete received message: %s", msg)

	// Decode the product barcode delete message
	var productData models.MongoProductBarcodeModel
	err := json.Unmarshal([]byte(msg), &productData)
	if err != nil {
		logger.Error("unmarshaling product barcode for deletion: %v", err)
		return err
	}

	if productData.HoldingCode == "" {
		logger.Warn("Product barcode delete data missing HoldingCode")
		return fmt.Errorf("missing HoldingCode for product deletion")
	}

	if productData.Barcode == "" {
		logger.Warn("Product barcode delete data missing Barcode")
		return fmt.Errorf("missing Barcode for product deletion")
	}

	build.DatabaseChecker(productData.HoldingCode, false)

	// Process single barcode deletion
	err = ProductBarcodeDeleteFromPostgreSQL(productData)
	if err != nil {
		logger.Error("deleting product barcode: %v", err)
		return err
	}

	logger.Info("Successfully deleted product barcode: %s", productData.Barcode)

	return nil
}

// OnConsumeMessageInventoryBulkCreateOrUpdate - handles bulk inventory create/update messages
func OnConsumeMessageInventoryBulkCreateOrUpdate(msg string) error {
	logger.Info("OnConsumeMessageInventoryBulkCreateOrUpdate received message")

	var barcodes []models.MongoProductBarcodeModel
	err := json.Unmarshal([]byte(msg), &barcodes)
	if err != nil {
		logger.Error("unmarshaling bulk barcodes: %v", err)
		return err
	}

	logger.Info("Received %d barcodes for bulk processing", len(barcodes))

	for i, barcode := range barcodes {
		name := ""
		if len(barcode.Names) > 0 {
			name = barcode.Names[0].Name
		}

		groupName := ""
		if len(barcode.GroupNames) > 0 {
			groupName = barcode.GroupNames[0].Name
		}

		unitName := ""
		if len(barcode.ItemUnitNames) > 0 {
			unitName = barcode.ItemUnitNames[0].Name
		}

		price := 0.0
		if len(barcode.Prices) > 0 {
			price = barcode.Prices[0].Price
		}

		logger.Info("Barcode %d: HoldingCode: %s, Barcode: %s, ItemCode: %s, Name: %s, GroupCode: %s, GroupName: %s,UnitCode: %s, UnitName: %s, Price: %.2f",
			i+1, barcode.HoldingCode, barcode.Barcode, barcode.ItemCode, name, barcode.GroupCode, groupName, barcode.ItemUnitCode, unitName, price)
	}

	err = ProductBarcodeBulkUpdateWithLogging(barcodes)
	if err != nil {
		logger.Error("in ProductBarcodeBulkUpdate: %v", err)
		return err
	}

	return nil
}

// OnConsumeMessageInventoryBulkDelete - handles bulk inventory delete messages
func OnConsumeMessageInventoryBulkDelete(msg string) error {
	logger.Info("OnConsumeMessageInventoryBulkDelete received message")

	var barcodes []models.MongoProductBarcodeModel
	err := json.Unmarshal([]byte(msg), &barcodes)
	if err != nil {
		logger.Error("unmarshaling bulk barcodes for deletion: %v", err)
		return err
	}

	logger.Info("Received %d barcodes for bulk deletion", len(barcodes))

	if len(barcodes) == 0 {
		logger.Warn("No barcodes provided for bulk deletion")
		return nil
	}

	// ตรวจสอบ HoldingCode ให้ตรงกันทุกตัว
	holdingCode := barcodes[0].HoldingCode
	for i, barcode := range barcodes {
		if barcode.HoldingCode != holdingCode {
			logger.Error("HoldingCode mismatch in bulk delete: barcode %d has HoldingCode %s, expected %s",
				i+1, barcode.HoldingCode, holdingCode)
			return fmt.Errorf("HoldingCode mismatch in bulk delete")
		}
	}

	err = ProductBarcodeBulkDeleteWithLogging(barcodes)
	if err != nil {
		logger.Error("in ProductBarcodeBulkDelete: %v", err)
		return err
	}

	return nil
}

// OnProductBarcodeCreateUpdateMessageConsume - wrapper function for product barcode processing
func OnProductBarcodeCreateUpdateMessageConsume(msg string) error {
	return SafeConsumerWrapper("PRODUCT_BARCODE_WAREHOUSE", func(msg string) error {
		ProductBarcodeBuild(msg)
		return nil
	})(msg)
}

// OnProductBarcodeDeleteMessageConsume - wrapper function for product barcode deletion processing
func OnProductBarcodeDeleteMessageConsume(msg string) error {
	return SafeConsumerWrapper("PRODUCT_BARCODE_WAREHOUSE", func(msg string) error {
		return OnConsumeMessageInventoryDelete(msg)
	})(msg)
}

// OnProductBarcodeBulkDeleteMessageConsume - wrapper function for bulk product barcode deletion processing
func OnProductBarcodeBulkDeleteMessageConsume(msg string) error {
	return SafeConsumerWrapper("PRODUCT_BARCODE_WAREHOUSE", func(msg string) error {
		return OnConsumeMessageInventoryBulkDelete(msg)
	})(msg)
}

// ProductBarcodeBuild - processes single product barcode update
func ProductBarcodeBuild(msg string) {
	// Decode the product barcode message
	var productData models.MongoProductBarcodeModel
	err := json.Unmarshal([]byte(msg), &productData)
	if err != nil {
		logger.Error("unmarshaling product barcode: %v", err)
		return
	}

	if productData.HoldingCode == "" {
		logger.Warn("Product barcode data missing HoldingCode")
		return
	}

	build.DatabaseChecker(productData.HoldingCode, false)

	// Process single barcode update
	err = ProductBarcodeInsertOrUpdateToPostgreSQL(productData)
	if err != nil {
		logger.Error("updating product barcode: %v", err)
	}

	// Update product balance (packing/unitname อาจเปลี่ยน → ต้อง recalc word)
	if productData.ItemCode != "" {
		db, dbErr := mypg.PgSqlFastConnect(productData.HoldingCode)
		if dbErr == nil {
			processstock.ProcessProductBalanceUpdateByItemsAsync(db, []string{productData.ItemCode})
		}
	}
}

// ProductBarcodeInsertOrUpdateToPostgreSQL - inserts or updates product barcode data in PostgreSQL
func ProductBarcodeInsertOrUpdateToPostgreSQL(productData models.MongoProductBarcodeModel) error {
	logger.Info("Processing product barcode: HoldingCode=%s, Barcode=%s, ItemCode=%s",
		productData.HoldingCode, productData.Barcode, productData.ItemCode)

	db, err := mypg.PgSqlFastConnect(productData.HoldingCode)
	if err != nil {
		return fmt.Errorf("failed to connect to Postgres: %v", err)
	}

	ctx := context.Background()

	// Calculate checksum
	checksum := myglobal.CalculateMD5(fmt.Sprintf("%v", productData))

	// Get product name (first available name)
	productName := ""
	if len(productData.Names) > 0 {
		productName = productData.Names[0].Name
	}

	// Get group name (first available name)
	groupName := ""
	if len(productData.GroupNames) > 0 {
		groupName = productData.GroupNames[0].Name
	}

	// Get unit name (first available name)
	unitName := ""
	if len(productData.ItemUnitNames) > 0 {
		unitName = productData.ItemUnitNames[0].Name
	}

	// Get price (first available price)
	price := 0.0
	if len(productData.Prices) > 0 {
		price = productData.Prices[0].Price
	}

	// Delete existing record
	_, err = db.ExecContext(ctx, "DELETE FROM productbarcode WHERE holding_code = $1 AND barcode = $2", productData.HoldingCode, productData.Barcode)
	if err != nil {
		logger.Warn("Could not delete existing barcode %s: %v", productData.Barcode, err)
	}

	// Insert new record (PostgreSQL)
	_, err = db.ExecContext(ctx,
		`INSERT INTO productbarcode (holding_code, barcode, itemcode, name0, checksum, groupcode, groupnames, unitcode, unitname, price1, barcoderefunitstand, barcoderefunitdivide, itemtype)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		productData.HoldingCode,
		productData.Barcode,
		productData.ItemCode,
		productName,
		checksum,
		productData.GroupCode,
		groupName,
		productData.ItemUnitCode,
		unitName,
		price,
		productData.StandValue,
		productData.DivideValue,
		productData.ItemType)

	if err != nil {
		return fmt.Errorf("error inserting product barcode %s: %v", productData.Barcode, err)
	}

	logger.Info("Inserted/Updated product barcode (PG): %s", productData.Barcode)

	// ClickHouse: insert/update ด้วย (best-effort — ไม่ fail ถ้า CH พัง)
	clickHouseInsertOrUpdate(productData.HoldingCode, productData.Barcode, productData.ItemCode, productName,
		productData.ItemUnitCode, unitName, productData.GroupCode, groupName,
		price, productData.StandValue, productData.DivideValue, checksum)

	return nil
}

// ProductBarcodeBulkUpdateWithLogging - performs bulk update of product barcodes with logging
func ProductBarcodeBulkUpdateWithLogging(productDataList []models.MongoProductBarcodeModel) error {
	if len(productDataList) == 0 {
		return nil
	}

	// Get holding Code from first item
	holdingCode := productDataList[0].HoldingCode

	// Connect to PostgreSQL
	db, err := mypg.PgSqlFastConnect(holdingCode)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %v", err)
	}

	ctx := context.Background()

	for i := 0; i < len(productDataList); i += BATCH_SIZE {
		end := i + BATCH_SIZE
		if end > len(productDataList) {
			end = len(productDataList)
		}

		batch := productDataList[i:end]
		logger.Info("Processing batch %d to %d", i+1, end)

		if err := productBarcodeBulkUpdateInternalWithLogging(ctx, db, batch); err != nil {
			return fmt.Errorf("error processing batch %d-%d: %v", i+1, end, err)
		}
	}

	return nil
}

// productBarcodeBulkUpdateInternalWithLogging - internal function for bulk update with logging
func productBarcodeBulkUpdateInternalWithLogging(ctx context.Context, db *sql.DB, productDataList []models.MongoProductBarcodeModel) error {
	logger.Info("Starting bulk update for %d records", len(productDataList))

	// Prepare data for COPY FROM
	var records [][]any

	for _, productData := range productDataList {
		checksum := myglobal.CalculateMD5(fmt.Sprintf("%v", productData))

		productName := ""
		if len(productData.Names) > 0 {
			productName = productData.Names[0].Name
		}

		groupName := ""
		if len(productData.GroupNames) > 0 {
			groupName = productData.GroupNames[0].Name
		}

		unitName := ""
		if len(productData.ItemUnitNames) > 0 {
			unitName = productData.ItemUnitNames[0].Name
		}

		price := 0.0
		if len(productData.Prices) > 0 {
			price = productData.Prices[0].Price
		}

		// For bulk update, use prepared statement to delete existing records
		deleteQuery := "DELETE FROM productbarcode WHERE holding_code = $1 AND barcode = $2"
		_, err := db.ExecContext(ctx, deleteQuery, productData.HoldingCode, productData.Barcode)
		if err != nil {
			logger.Warn("Could not delete existing barcode %s: %v", productData.Barcode, err)
		}

		record := []any{
			productData.HoldingCode,
			productData.Barcode,
			productData.ItemCode,
			productName,
			checksum,
			productData.GroupCode,
			groupName,
			productData.ItemUnitCode,
			unitName,
			price,
			productData.StandValue,
			productData.DivideValue,
			productData.ItemType,
		}
		records = append(records, record)
	}

	// Use COPY FROM for bulk insert (PostgreSQL)
	columns := []string{
		"holding_code", "barcode", "itemcode", "name0", "checksum",
		"groupcode", "groupnames", "unitcode", "unitname", "price1",
		"barcoderefunitstand", "barcoderefunitdivide", "itemtype",
	}

	err := mypg.BulkInsertWithCopy(ctx, db, "productbarcode", columns, records)
	if err != nil {
		return fmt.Errorf("failed to bulk insert product barcodes: %v", err)
	}
	logger.Info("Bulk update (PG) completed. Processed %d records", len(productDataList))

	// ClickHouse: bulk insert ด้วย (best-effort)
	clickHouseBulkInsert(productDataList)

	return nil
}

// ProductBarcodeDeleteFromPostgreSQL - ลบข้อมูลสินค้าใน PostgreSQL + ClickHouse
func ProductBarcodeDeleteFromPostgreSQL(productData models.MongoProductBarcodeModel) error {
	logger.Info("Deleting product barcode: HoldingCode=%s, Barcode=%s",
		productData.HoldingCode, productData.Barcode)

	db, err := mypg.PgSqlFastConnect(productData.HoldingCode)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %v", err)
	}

	ctx := context.Background()

	// ลบข้อมูลสินค้า (PostgreSQL)
	result, err := db.ExecContext(ctx, "DELETE FROM productbarcode WHERE holding_code = $1 AND barcode = $2", productData.HoldingCode, productData.Barcode)
	if err != nil {
		return fmt.Errorf("error deleting product barcode %s: %v", productData.Barcode, err)
	}

	// ตรวจสอบจำนวนแถวที่ถูกลบ
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Warn("Could not get rows affected for barcode %s: %v", productData.Barcode, err)
	} else {
		logger.Info("Deleted %d row(s) for barcode %s (PG)", rowsAffected, productData.Barcode)
	}

	// ClickHouse: ลบด้วย (best-effort)
	clickHouseDelete(productData.HoldingCode, productData.Barcode)

	return nil
}

// ProductBarcodeBulkDeleteWithLogging - ลบสินค้าแบบเป็นชุดพร้อม logging
func ProductBarcodeBulkDeleteWithLogging(productDataList []models.MongoProductBarcodeModel) error {
	if len(productDataList) == 0 {
		return nil
	}

	// Get holding Code from first item
	holdingCode := productDataList[0].HoldingCode

	// Connect to PostgreSQL
	db, err := mypg.PgSqlFastConnect(holdingCode)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %v", err)
	}

	ctx := context.Background()

	for i := 0; i < len(productDataList); i += BATCH_SIZE {
		end := i + BATCH_SIZE
		if end > len(productDataList) {
			end = len(productDataList)
		}

		batch := productDataList[i:end]
		logger.Info("Processing batch %d to %d for deletion", i+1, end)

		if err := productBarcodeBulkDeleteInternalWithLogging(ctx, db, batch); err != nil {
			return fmt.Errorf("error processing batch %d-%d for deletion: %v", i+1, end, err)
		}
	}

	return nil
}

// productBarcodeBulkDeleteInternalWithLogging - ฟังก์ชันภายในสำหรับการลบแบบ bulk
func productBarcodeBulkDeleteInternalWithLogging(ctx context.Context, db *sql.DB, productDataList []models.MongoProductBarcodeModel) error {
	logger.Info("Starting bulk delete for %d records", len(productDataList))

	var deletedCount int
	var failedBarcodes []string

	for _, productData := range productDataList {
		result, err := db.ExecContext(ctx, "DELETE FROM productbarcode WHERE holding_code = $1 AND barcode = $2", productData.HoldingCode, productData.Barcode)
		if err != nil {
			logger.Error("failed to delete barcode %s: %v", productData.Barcode, err)
			failedBarcodes = append(failedBarcodes, productData.Barcode)
			continue
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			logger.Warn("could not get rows affected for barcode %s: %v", productData.Barcode, err)
			failedBarcodes = append(failedBarcodes, productData.Barcode)
			continue
		}

		if rowsAffected > 0 {
			deletedCount++
			logger.Debug("Successfully deleted barcode %s", productData.Barcode)
		} else {
			logger.Warn("No rows deleted for barcode %s (not found)", productData.Barcode)
		}
	}

	logger.Info("Bulk delete (PG) completed. Successfully deleted %d records, %d failed",
		deletedCount, len(failedBarcodes))

	if len(failedBarcodes) > 0 {
		logger.Warn("Failed to delete barcodes: %v", failedBarcodes)
	}

	// ClickHouse: bulk delete ด้วย (best-effort)
	if len(productDataList) > 0 {
		clickHouseBulkDelete(productDataList[0].HoldingCode, productDataList)
	}

	return nil
}

// ==================== ClickHouse Helper Functions (disabled) ====================

func clickHouseInsertOrUpdate(holdingCode, barcode, itemcode, name0, unitcode, unitname, groupcode, groupnames string,
	price, standValue, divideValue float64, checksum string) {
}

func clickHouseBulkInsert(productDataList []models.MongoProductBarcodeModel) {
}

func clickHouseDelete(holdingCode, barcode string) {
}

func clickHouseBulkDelete(holdingCode string, productDataList []models.MongoProductBarcodeModel) {
}
