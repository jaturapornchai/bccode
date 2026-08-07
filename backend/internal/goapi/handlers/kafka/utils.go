package kafka

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/mitchellh/mapstructure"

	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/models"
	"smlcloudplatform/internal/goapi/mydlq"
	"smlcloudplatform/internal/goapi/myglobal"
	"smlcloudplatform/internal/goapi/mypg"
	"smlcloudplatform/internal/goapi/myretry"
	processdoc "smlcloudplatform/internal/goapi/process/process-doc"
	processstock "smlcloudplatform/internal/goapi/process/process-stock"
)

// MessageConsumer interface for message consumers
type MessageConsumer interface {
	ConsumeCreateOrUpdate(msg string) error
	ConsumeDelete(msg string) error
}

// BulkMessageConsumer interface for bulk message consumers
type BulkMessageConsumer interface {
	MessageConsumer
	ConsumeBulk(msg string) error
}

// SafeConsumerWrapper wraps a consumer function with panic recovery, retry และ DLQ
func SafeConsumerWrapper(consumerName string, fn func(string) error) func(string) error {
	return func(msg string) error {
		defer func() {
			if r := recover(); r != nil {
				logger.Info("PANIC recovered in %s: %v", consumerName, r)
				logger.Info("Stack trace: %s", debug.Stack())
				logger.Info("Message length: %d characters", len(msg))
				logger.Info("Message preview (first 200 chars): %.200s", msg)

				// ส่งไป DLQ เมื่อเกิด panic
				panicErr := fmt.Errorf("panic: %v", r)
				mydlq.QuickSendToDLQ(consumerName, msg, panicErr)
			}
		}()

		if msg == "" {
			logger.Error("Empty message received in %s", consumerName)
			return fmt.Errorf("empty message received")
		}

		logger.Info("=== %s CONSUMER STARTED ===", consumerName)
		logger.Info("Received message length: %d characters", len(msg))
		logger.Info("Message preview: %.500s", msg) // Show first 500 characters

		// ✅ เพิ่ม Retry logic with exponential backoff
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		retryConfig := &myretry.RetryConfig{
			MaxRetries: 3,
			BaseDelay:  1 * time.Second,
			MaxDelay:   30 * time.Second,
			Multiplier: 2.0,
			OnRetry: func(attempt int, err error) {
				logger.Warn("⚠️ %s: Retry attempt %d due to: %v", consumerName, attempt, err)
			},
			IsRetryable: func(err error) bool {
				// ตรวจสอบว่า error นี้ควร retry หรือไม่
				if err == nil {
					return false
				}
				// ไม่ retry กับ validation errors
				errMsg := err.Error()
				if contains(errMsg, "invalid") || contains(errMsg, "missing") {
					return false
				}
				// Retry กับ database/network errors
				return contains(errMsg, "connection") ||
					contains(errMsg, "timeout") ||
					contains(errMsg, "temporary") ||
					contains(errMsg, "unavailable")
			},
		}

		var finalErr error
		err := myretry.WithRetry(ctx, retryConfig, func() error {
			return fn(msg)
		})

		if err != nil {
			finalErr = err
			logger.Error("❌ %s failed after retries: %v", consumerName, err)

			// ✅ ส่งไป Dead Letter Queue
			dlqErr := mydlq.SendToDLQ(ctx, consumerName, "", msg, err, retryConfig.MaxRetries, map[string]interface{}{
				"consumer":   consumerName,
				"msg_length": len(msg),
				"failed_at":  time.Now(),
				"retrycount": retryConfig.MaxRetries,
			})

			if dlqErr != nil {
				logger.Error("Failed to send message to DLQ: %v", dlqErr)
			}

			return finalErr
		}

		logger.Info("=== %s CONSUMER COMPLETED SUCCESSFULLY ===", consumerName)
		return nil
	}
}

// contains helper function (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && anySubstring(s, substr))
}

func anySubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ValidateMessage validates that the message is not empty
func ValidateMessage(msg string) error {
	if msg == "" {
		return fmt.Errorf("empty message received")
	}
	return nil
}

// LogMessageInfo logs basic message information
func LogMessageInfo(consumerName string, msg string) {
	logger.Info("=== %s CONSUMER STARTED ===", consumerName)
	logger.Info("Received message length: %d characters", len(msg))
	logger.Info("Message preview: %.500s", msg)
}

// ============================================================================
// Common Document Processing Functions
// ============================================================================

// ProcessDocumentStockCalculation - คำนวณต้นทุนสินค้าสำหรับเอกสาร (ใช้ร่วมกันได้ทุก document type)
// Parameters:
//   - db: database connection (*sql.DB)
//   - holdingCode: holding Code
//   - docDetailStructs: รายการสินค้าในเอกสาร ([]models.DocDetailStruct)
//   - stepNumber: เลข step สำหรับ log (เช่น 11, 12)
//
// Returns: error
func ProcessDocumentStockCalculation(db *sql.DB, holdingCode string, docDetailStructs []models.DocDetailStruct, stepNumber int) error {
	// Use incremental mode by default for Kafka consumer
	return ProcessDocumentStockCalculationWithOptions(db, holdingCode, docDetailStructs, stepNumber, true, true)
}

// ProcessDocumentStockCalculationWithOptions - คำนวณต้นทุนสินค้าพร้อม options
// Parameters:
//   - db: database connection (*sql.DB)
//   - holdingCode: holding Code
//   - docDetailStructs: รายการสินค้าในเอกสาร ([]models.DocDetailStruct)
//   - stepNumber: เลข step สำหรับ log (เช่น 11, 12)
//   - incremental: ถ้า true จะตรวจสอบ checksum ก่อน และข้ามถ้าไม่มีการเปลี่ยนแปลง
//   - minimalLog: ถ้า true จะใช้ UPSERT แทน DELETE+INSERT เพื่อลด WAL log
//
// Returns: error
func ProcessDocumentStockCalculationWithOptions(db *sql.DB, holdingCode string, docDetailStructs []models.DocDetailStruct, stepNumber int, incremental, minimalLog bool) error {
	logger.Debug("Step %d: Starting stock calculation (incremental=%v, minimalLog=%v)...", stepNumber, incremental, minimalLog)

	// รวบรวม unique itemcodes จาก docdetails
	itemCodeMap := make(map[string]bool)

	for _, detail := range docDetailStructs {
		if detail.ItemCode != "" {
			itemCodeMap[detail.ItemCode] = true
		}
	}

	if len(itemCodeMap) == 0 {
		logger.Warn("Step %d skipped: No items to calculate cost", stepNumber)
		return nil
	}

	// ใช้ค่าทศนิยมจาก global config
	pointQty := myglobal.ConfigSystem.StockQtyPoint       // ทศนิยมจำนวน
	pointAmount := myglobal.ConfigSystem.StockAmountPoint // ทศนิยมมูลค่า
	pointCost := myglobal.ConfigSystem.StockCostPoint     // ทศนิยมต้นทุน

	logger.Info("Using decimal points: Qty=%d, Amount=%d, Cost=%d", pointQty, pointAmount, pointCost)

	// Track stats for incremental mode
	totalItems := len(itemCodeMap)
	skippedItems := 0
	processedItems := 0

	// คำนวณต้นทุนทีละ itemcode
	itemCount := 0
	for itemCode := range itemCodeMap {
		itemCount++
		logger.Info("Calculating cost for item %d/%d: %s", itemCount, totalItems, itemCode)

		if incremental {
			// Use incremental mode with checksum checking
			processstock.ProductCalcCostIncremental(
				db,          // database connection
				holdingCode, // holdingCode
				itemCode,    // itemCodeForProcess
				pointQty,    // pointQty (จาก global config)
				pointAmount, // pointAmount (จาก global config)
				pointCost,   // pointCost (จาก global config)
				true,        // incremental
				minimalLog,  // minimalLog
			)
		} else {
			// Legacy mode: delete first then insert
			processstock.ProductCalcCost(
				db,          // database connection
				holdingCode, // holdingCode
				itemCode,    // itemCodeForProcess
				pointQty,    // pointQty (จาก global config)
				pointAmount, // pointAmount (จาก global config)
				pointCost,   // pointCost (จาก global config)
				true,        // deleteFrist (ลบข้อมูลเก่าก่อน)
			)
		}

		processedItems++
		logger.Success("Completed cost calculation for item: %s", itemCode)
	}

	// Log stats
	if incremental {
		logger.Info("Incremental calculation stats | shop=%s total=%d skipped=%d processed=%d",
			holdingCode, totalItems, skippedItems, processedItems)
	}

	logger.Success("Step %d completed: Calculated cost for %d items", stepNumber, totalItems)

	// Update product balance + ค้างรับ + ค้างส่ง (async — ไม่ block Kafka consumer)
	itemCodeList := make([]string, 0, len(itemCodeMap))
	for code := range itemCodeMap {
		itemCodeList = append(itemCodeList, code)
	}
	processstock.ProcessProductBalanceUpdateByItemsAsync(db, itemCodeList)

	return nil
}

// InsertDocumentToPostgreSQL - Insert เอกสารไปยัง PostgreSQL (ใช้ร่วมกัน)
// Parameters:
//   - ctx: context
//   - db: database connection (*sql.DB)
//   - docStruct: เอกสารหลัก
//   - docRefStructs: เอกสารอ้างอิง
//   - docPaymentStruct: ข้อมูลการชำระเงิน
//   - stepNumber: เลข step สำหรับ log
//
// Returns: error
func InsertDocumentToPostgreSQL(ctx context.Context, db *sql.DB, docStruct models.DocStruct, docRefStructs []models.DocRefStruct, docPaymentStruct models.DocPaymentStruct, stepNumber int) error {
	logger.Debug("Step %d: Inserting main document to PostgreSQL...", stepNumber)

	if err := mypg.InsertDocListToPostgreSql(ctx, db,
		[]models.DocStruct{docStruct},
		docRefStructs,
		[]models.DocPaymentStruct{docPaymentStruct}); err != nil {
		logger.Error("Failed to insert main document: %v", err)
		return fmt.Errorf("failed to insert doc: %w", err)
	}

	logger.Success("Step %d completed: Inserted main document", stepNumber)
	return nil
}

func InsertDocumentToPostgreSQLTx(ctx context.Context, tx *sql.Tx, docStruct models.DocStruct, docRefStructs []models.DocRefStruct, docPaymentStruct models.DocPaymentStruct, stepNumber int) error {
	logger.Debug("Step %d: Inserting main document to PostgreSQL (tx)...", stepNumber)

	if err := mypg.InsertDocListTx(ctx, tx,
		[]models.DocStruct{docStruct},
		docRefStructs,
		[]models.DocPaymentStruct{docPaymentStruct}); err != nil {
		logger.Error("Failed to insert main document (tx): %v", err)
		return fmt.Errorf("failed to insert doc (tx): %w", err)
	}

	logger.Success("Step %d completed: Inserted main document", stepNumber)
	return nil
}

// InsertDocDetailToPostgreSQL - Insert รายละเอียดเอกสารไปยัง PostgreSQL (ใช้ร่วมกัน)
// Parameters:
//   - ctx: context
//   - db: database connection (*sql.DB)
//   - holdingCode: holding Code
//   - docDetailStructs: รายการสินค้า
//   - stepNumber: เลข step สำหรับ log
//
// Returns: error
func InsertDocDetailToPostgreSQL(ctx context.Context, db *sql.DB, holdingCode string, docDetailStructs []models.DocDetailStruct, stepNumber int) error {
	logger.Debug("Step %d: Inserting document details to PostgreSQL...", stepNumber)

	err := mypg.InsertDocDetailListToPostgreSql(ctx, db, holdingCode, docDetailStructs)
	if err != nil {
		logger.Error("Failed to insert docdetail: %v", err)
		return fmt.Errorf("failed to insert docdetail: %v", err)
	}

	logger.Success("Step %d completed: Inserted %d document details", stepNumber, len(docDetailStructs))
	return nil
}

func InsertDocDetailToPostgreSQLTx(ctx context.Context, tx *sql.Tx, holdingCode string, docDetailStructs []models.DocDetailStruct, stepNumber int) error {
	logger.Debug("Step %d: Inserting document details to PostgreSQL (tx)...", stepNumber)

	if err := mypg.InsertDocDetailListToPostgreSqlTx(ctx, tx, holdingCode, docDetailStructs); err != nil {
		logger.Error("Failed to insert docdetail (tx): %v", err)
		return fmt.Errorf("failed to insert docdetail (tx): %v", err)
	}

	logger.Success("Step %d completed: Inserted %d document details", stepNumber, len(docDetailStructs))
	return nil
}

// InsertDocumentToClickHouse - Insert เอกสารไปยัง ClickHouse (ใช้ร่วมกัน)
// Parameters:
//   - ctx: context
//   - holdingCode: holding Code
//   - docStruct: เอกสารหลัก
//   - docRefStructs: เอกสารอ้างอิง
//   - docPaymentStruct: ข้อมูลการชำระเงิน
//   - docDetailStructs: รายการสินค้า
//   - stepNumber: เลข step สำหรับ log
//
// Returns: error
func InsertDocumentToClickHouse(ctx context.Context, holdingCode string, docStruct models.DocStruct, docRefStructs []models.DocRefStruct, docPaymentStruct models.DocPaymentStruct, docDetailStructs []models.DocDetailStruct, stepNumber int) error {
	logger.Debug("Step %d: ClickHouse is disabled, skipping insert.", stepNumber)
	return nil
}

// GetItemName returns the item name based on the configured language model.
// Falls back to first available name if configured language not found.
func GetItemName(itemNames []models.LanguageModel) string {
	if len(itemNames) == 0 {
		return ""
	}

	// Get language code from global config
	languageCode := myglobal.ConfigSystem.DefaultLanguage
	if languageCode == "" {
		languageCode = "th" // Default fallback
	}

	// Try to find name in configured language
	for _, name := range itemNames {
		if name.Code == languageCode {
			return name.Name
		}
	}

	// Return first available name if configured language not found
	return itemNames[0].Name
}

// ConvertLanguageModels converts language model array with nil check
func ConvertLanguageModels(sourceNames []models.LanguageModel) []models.LanguageModel {
	if sourceNames == nil {
		return []models.LanguageModel{}
	}

	var result []models.LanguageModel
	for _, name := range sourceNames {
		result = append(result, models.LanguageModel{
			Code: name.Code,
			Name: name.Name,
		})
	}
	return result
}

// DecodeKafkaMessage decodes JSON message from Kafka with error handling
func DecodeKafkaMessage[T any](jsonData string, targetModel *T, messageType string) error {
	if jsonData == "" {
		return fmt.Errorf("empty JSON data for %s", messageType)
	}

	jsonDecode := map[string]any{}
	err := json.Unmarshal([]byte(jsonData), &jsonDecode)
	if err != nil {
		logger.Error("unmarshal %s: %v", messageType, err)
		return fmt.Errorf("unmarshal %s: %w", messageType, err)
	}

	fixedData := myglobal.FixMongoData(jsonDecode)

	decoderConfig := mapstructure.DecoderConfig{
		Squash:           true,
		WeaklyTypedInput: true,
		TagName:          "json", // ใช้ json tag เพื่อ map keys ที่มี underscore (เช่น doc_currency → DocCurrency) ให้ตรงกับ DocDecode
		Result:           targetModel,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeHookFunc(time.RFC3339),
		),
	}
	decoder, err := mapstructure.NewDecoder(&decoderConfig)
	if err != nil {
		logger.Error("creating decoder for %s: %v", messageType, err)
		return fmt.Errorf("creating decoder for %s: %w", messageType, err)
	}

	err = decoder.Decode(fixedData)
	if err != nil {
		logger.Error("decoding %s: %v", messageType, err)
		return fmt.Errorf("decoding %s: %w", messageType, err)
	}

	return nil
}

// DeleteDocumentFromDatabases ทำ soft delete เอกสารใน PostgreSQL และ ClickHouse
// เปลี่ยนจาก hard delete เป็น UPDATE isdelete = true
func DeleteDocumentFromDatabases(ctx context.Context, holdingCode, docNo string, transFlag int) error {
	// Soft delete ใน PostgreSQL
	db, err := mypg.PgSqlFastConnect(holdingCode)
	if err != nil {
		logger.Error("เชื่อมต่อ PostgreSQL สำหรับ soft delete ล้มเหลว: %v", err)
		return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	softDeleteQuery := "UPDATE doc SET isdelete = true WHERE docno = $1 AND transflag = $2"
	_, err = db.ExecContext(ctx, softDeleteQuery, docNo, transFlag)
	if err != nil {
		logger.Error("Soft delete ใน PostgreSQL ล้มเหลว: %v", err)
		return fmt.Errorf("failed to soft delete from PostgreSQL: %w", err)
	}
	logger.Info("[SoftDelete] PostgreSQL: docno=%s, transflag=%d — isdelete=true", docNo, transFlag)

	// Soft delete ใน ClickHouse
	SoftDeleteDocClickHouse(ctx, holdingCode, docNo)

	// เพิ่มเข้า document wait process queue
	query := "INSERT INTO docwaitprocess (docno, transflag) VALUES ($1, $2)"
	_, err = db.ExecContext(ctx, query, docNo, transFlag)
	if err != nil {
		logger.Warn("ไม่สามารถเพิ่มเข้า doc queue: %v", err)
	}

	return nil
}

// SoftDeleteDocClickHouse ทำ soft delete เอกสารใน ClickHouse (เลิกใช้งานแล้ว)
func SoftDeleteDocClickHouse(ctx context.Context, holdingCode, docNo string) {
	// ClickHouse is permanently disabled
}

// ProcessDocumentStatusAsync - ประมวลผลสถานะเอกสารแบบ async หลังจาก Kafka เสร็จ
// เรียกทันทีโดยไม่มี debounce เพื่อให้สถานะอัปเดตเร็วที่สุด
func ProcessDocumentStatusAsync(holdingCode string) {
	// เรียกทันทีแบบ goroutine ไม่ต้องรอ
	processdoc.ProcessDocumentStatusByShop(holdingCode)
}
