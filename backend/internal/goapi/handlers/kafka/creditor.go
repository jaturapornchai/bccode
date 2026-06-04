package kafka

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"runtime/debug"
	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/models"
	"smlcloudplatform/internal/goapi/myglobal"
	"smlcloudplatform/internal/goapi/mypg"
	"time"

	"github.com/mitchellh/mapstructure"
)

// OnConsumeMessageCreditorCreateOrUpdate - handles creditor create/update messages
func OnConsumeMessageCreditorCreateOrUpdate(msg string) error {
	return SafeConsumerWrapper("CREDITOR", func(msg string) error {
		return ProcessCreditorMasterData(msg)
	})(msg)
}

// OnConsumeMessageCreditorDelete - handles creditor delete messages
func OnConsumeMessageCreditorDelete(msg string) error {
	// รับ Message จาก Kafka ที่เป็นการลบข้อมูล Creditor
	// msg จะเป็น JSON string ที่มีข้อมูล เช่น {"holding_code": "shop123", "code": "AP0001"}

	logger.Info("OnConsumeMessageCreditorDelete: %s", msg)

	creditorData := TransCreditorDecode(msg)

	if creditorData.HoldingCode == "" || creditorData.Code == "" {
		logger.Error("ข้อมูล creditor ไม่ถูกต้อง - ไม่มี HoldingCode หรือ Code")
		return fmt.Errorf("invalid creditor data")
	}

	// ลบข้อมูลใน PostgreSQL
	db, err := mypg.PgSqlFastConnect(creditorData.HoldingCode)
	if err != nil {
		logger.Error("ไม่สามารถเชื่อมต่อ PostgreSQL: %v", err)
		return err
	}

	query := "DELETE FROM creditor WHERE code = $1"
	_, err = db.ExecContext(context.Background(), query, creditorData.Code)
	if err != nil {
		logger.Error("ไม่สามารถลบ creditor: %v", err)
		return err
	}

	logger.Info("ลบ creditor สำเร็จ: HoldingCode=%s, Code=%s", creditorData.HoldingCode, creditorData.Code)
	return nil
}

// ProcessCreditorMasterData - processes creditor master data from Kafka
func ProcessCreditorMasterData(msg string) error {
	// ประมวลผลข้อมูล Master Data: Creditor (เจ้าหนี้)
	logger.Info("--- ProcessCreditorMasterData START ---")

	defer func() {
		if r := recover(); r != nil {
			logger.Info("PANIC recovered in ProcessCreditorMasterData: %v", r)
			logger.Info("Stack trace: %s", debug.Stack())
			logger.Info("Message length: %d characters", len(msg))
		}
	}()

	// Decode creditor from JSON message
	logger.Debug("Step 1: กำลัง Decode JSON message... ")
	creditorData := TransCreditorDecode(msg)
	logger.Info("Step 1: Decode creditor สำเร็จ - HoldingCode=%s, Code=%s, TaxID=%s, Names=%d",
		creditorData.HoldingCode, creditorData.Code, creditorData.TaxID, len(creditorData.Names))

	if creditorData.HoldingCode == "" || creditorData.Code == "" {
		logger.Error("ข้อมูล creditor ไม่ถูกต้อง - HoldingCode='%s', Code='%s'", creditorData.HoldingCode, creditorData.Code)
		return fmt.Errorf("invalid creditor data - missing HoldingCode or Code")
	}

	// Connect to database
	logger.Debug("Step 2: กำลังเชื่อมต่อ PostgreSQL...")
	db, err := mypg.PgSqlFastConnect(creditorData.HoldingCode)
	if err != nil {
		logger.Error("ไม่สามารถเชื่อมต่อ database: %v", err)
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	logger.Debug("Step 2 เสร็จสิ้น: เชื่อมต่อ PostgreSQL สำเร็จ")

	ctx := context.Background()

	// Insert or Update creditor to PostgreSQL
	logger.Debug("Step 3: กำลัง Insert/Update creditor ไปยัง PostgreSQL...")
	err = insertOrUpdateCreditor(ctx, db, creditorData)
	if err != nil {
		logger.Error("ไม่สามารถ Insert/Update creditor: %v", err)
		return err
	}
	logger.Debug("Step 3 เสร็จสิ้น: Insert/Update creditor สำเร็จ")

	logger.Info("--- ProcessCreditorMasterData COMPLETED SUCCESSFULLY: %s ---", creditorData.Code)
	return nil
}

// insertOrUpdateCreditor - inserts or updates creditor in PostgreSQL
func insertOrUpdateCreditor(ctx context.Context, db *sql.DB, creditor models.ProcessMongoCreditorModel) error {
	// Extract first name from names array
	name0 := getCreditorName(creditor.Names)

	query := `
        INSERT INTO creditor (code, name0, taxid)
        VALUES ($1, $2, $3)
        ON CONFLICT (code) DO UPDATE SET
            name0 = EXCLUDED.name0,
            taxid = EXCLUDED.taxid
    `

	_, err := db.ExecContext(ctx, query, creditor.Code, name0, creditor.TaxID)
	if err != nil {
		logger.Error("ไม่สามารถ insert creditor: %v", err)
		return err
	}

	logger.Info("Insert/Update creditor สำเร็จ: Code=%s, Name=%s", creditor.Code, name0)
	return nil
}

// TransCreditorDecode - decodes creditor JSON data
func TransCreditorDecode(jsonData string) models.ProcessMongoCreditorModel {
	jsonDecode := map[string]any{}
	err := json.Unmarshal([]byte(jsonData), &jsonDecode)
	if err != nil {
		logger.Error("unmarshal transCreditor: %v", err)
		return models.ProcessMongoCreditorModel{}
	}

	fixedData := myglobal.FixMongoData(jsonDecode)

	var creditorData models.ProcessMongoCreditorModel
	decoderConfig := mapstructure.DecoderConfig{
		Squash:           true,
		WeaklyTypedInput: true,
		Result:           &creditorData,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeHookFunc(time.RFC3339),
		),
	}
	decoder, err := mapstructure.NewDecoder(&decoderConfig)
	if err != nil {
		logger.Error("creating decoder: %v", err)
		return models.ProcessMongoCreditorModel{}
	}

	err = decoder.Decode(fixedData)
	if err != nil {
		logger.Error("decoding creditor: %v", err)
		return models.ProcessMongoCreditorModel{}
	}

	return creditorData
}

// OnConsumeMessageCreditorBulkCreateOrUpdate - handles bulk creditor create/update messages
// รับ JSON array ของ creditor documents จาก Kafka topic: when-creditor-bulk-created
func OnConsumeMessageCreditorBulkCreateOrUpdate(msg string) error {
	return SafeConsumerWrapper("CREDITOR_BULK", func(msg string) error {
		logger.Info("--- ProcessCreditorBulkMasterData START ---")

		// ลอง decode เป็น array ก่อน
		var rawArray []map[string]interface{}
		if err := json.Unmarshal([]byte(msg), &rawArray); err != nil {
			// ถ้าไม่ใช่ array → ลอง single object (fallback)
			logger.Info("ไม่ใช่ JSON array — fallback เป็น single creditor")
			return ProcessCreditorMasterData(msg)
		}

		logger.Info("Received %d creditors for bulk processing", len(rawArray))

		for i, raw := range rawArray {
			rawJSON, err := json.Marshal(raw)
			if err != nil {
				logger.Error("marshal creditor[%d] ล้มเหลว: %v", i, err)
				continue
			}
			if err := ProcessCreditorMasterData(string(rawJSON)); err != nil {
				logger.Error("process creditor[%d] ล้มเหลว: %v", i, err)
				// ไม่ return error — ประมวลผลตัวถัดไปต่อ
			}
		}

		logger.Info("--- ProcessCreditorBulkMasterData COMPLETED: %d items ---", len(rawArray))
		return nil
	})(msg)
}

// OnConsumeMessageCreditorBulkDelete - handles bulk creditor delete messages
// รับ JSON array ของ creditor documents จาก Kafka topic: when-creditor-bulk-deleted
func OnConsumeMessageCreditorBulkDelete(msg string) error {
	return SafeConsumerWrapper("CREDITOR_BULK_DELETE", func(msg string) error {
		logger.Info("--- ProcessCreditorBulkDelete START ---")

		var rawArray []map[string]interface{}
		if err := json.Unmarshal([]byte(msg), &rawArray); err != nil {
			logger.Info("ไม่ใช่ JSON array — fallback เป็น single creditor delete")
			return OnConsumeMessageCreditorDelete(msg)
		}

		logger.Info("Received %d creditors for bulk deletion", len(rawArray))

		for i, raw := range rawArray {
			rawJSON, err := json.Marshal(raw)
			if err != nil {
				logger.Error("marshal creditor[%d] ล้มเหลว: %v", i, err)
				continue
			}

			creditorData := TransCreditorDecode(string(rawJSON))
			if creditorData.HoldingCode == "" || creditorData.Code == "" {
				logger.Warn("creditor[%d] ไม่มี HoldingCode/Code — ข้าม", i)
				continue
			}

			db, err := mypg.PgSqlFastConnect(creditorData.HoldingCode)
			if err != nil {
				logger.Error("creditor[%d] เชื่อมต่อ PG ล้มเหลว: %v", i, err)
				continue
			}

			query := "DELETE FROM creditor WHERE code = $1"
			_, err = db.ExecContext(context.Background(), query, creditorData.Code)
			if err != nil {
				logger.Error("creditor[%d] ลบล้มเหลว: %v", i, err)
				continue
			}
			logger.Info("ลบ creditor สำเร็จ: Code=%s", creditorData.Code)
		}

		logger.Info("--- ProcessCreditorBulkDelete COMPLETED: %d items ---", len(rawArray))
		return nil
	})(msg)
}

// getCreditorName - helper function to get creditor name from language models
func getCreditorName(names []models.ProcessMongoCreditorNameModel) string {
	if len(names) == 0 {
		return ""
	}

	// Return first available name
	return names[0].Name
}
