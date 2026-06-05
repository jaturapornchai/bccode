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

// OnConsumeMessageDebtorCreateOrUpdate - handles debtor create/update messages
func OnConsumeMessageDebtorCreateOrUpdate(msg string) error {
	return SafeConsumerWrapper("DEBTOR", func(msg string) error {
		return ProcessDebtorMasterData(msg)
	})(msg)
}

// OnConsumeMessageDebtorDelete - handles debtor delete messages
func OnConsumeMessageDebtorDelete(msg string) error {
	// รับ Message จาก Kafka ที่เป็นการลบข้อมูล Debtor
	// msg จะเป็น JSON string ที่มีข้อมูล เช่น {"holdingcode": "shop123", "code": "AR0001"}

	logger.Info("OnConsumeMessageDebtorDelete: %s", msg)

	debtorData := TransDebtorDecode(msg)

	if debtorData.HoldingCode == "" || debtorData.Code == "" {
		logger.Error("ข้อมูล debtor ไม่ถูกต้อง - ไม่มี HoldingCode หรือ Code")
		return fmt.Errorf("invalid debtor data")
	}

	// ลบข้อมูลใน PostgreSQL
	db, err := mypg.PgSqlFastConnect(debtorData.HoldingCode)
	if err != nil {
		logger.Error("ไม่สามารถเชื่อมต่อ PostgreSQL: %v", err)
		return err
	}

	query := "DELETE FROM debtor WHERE code = $1"
	_, err = db.ExecContext(context.Background(), query, debtorData.Code)
	if err != nil {
		logger.Error("ไม่สามารถลบ debtor: %v", err)
		return err
	}

	logger.Info("ลบ debtor สำเร็จ: HoldingCode=%s, Code=%s", debtorData.HoldingCode, debtorData.Code)
	return nil
}

// ProcessDebtorMasterData - processes debtor master data from Kafka
func ProcessDebtorMasterData(msg string) error {
	// ประมวลผลข้อมูล Master Data: Debtor (ลูกหนี้)
	logger.Info("--- ProcessDebtorMasterData START ---")

	defer func() {
		if r := recover(); r != nil {
			logger.Info("PANIC recovered in ProcessDebtorMasterData: %v", r)
			logger.Info("Stack trace: %s", debug.Stack())
			logger.Info("Message length: %d characters", len(msg))
		}
	}()

	// Decode debtor from JSON message
	logger.Debug("Step 1: กำลัง Decode JSON message... ")
	debtorData := TransDebtorDecode(msg)
	logger.Info("Step 1: Decode debtor สำเร็จ - HoldingCode=%s, Code=%s, TaxID=%s, Names=%d",
		debtorData.HoldingCode, debtorData.Code, debtorData.TaxID, len(debtorData.Names))

	if debtorData.HoldingCode == "" || debtorData.Code == "" {
		logger.Error("ข้อมูล debtor ไม่ถูกต้อง - HoldingCode='%s', Code='%s'", debtorData.HoldingCode, debtorData.Code)
		return fmt.Errorf("invalid debtor data - missing HoldingCode or Code")
	}

	// Connect to database
	logger.Debug("Step 2: กำลังเชื่อมต่อ PostgreSQL...")
	db, err := mypg.PgSqlFastConnect(debtorData.HoldingCode)
	if err != nil {
		logger.Error("ไม่สามารถเชื่อมต่อ database: %v", err)
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	logger.Debug("Step 2 เสร็จสิ้น: เชื่อมต่อ PostgreSQL สำเร็จ")

	ctx := context.Background()

	// Insert or Update debtor to PostgreSQL
	logger.Debug("Step 3: กำลัง Insert/Update debtor ไปยัง PostgreSQL...")
	err = insertOrUpdateDebtor(ctx, db, debtorData)
	if err != nil {
		logger.Error("ไม่สามารถ Insert/Update debtor: %v", err)
		return err
	}
	logger.Debug("Step 3 เสร็จสิ้น: Insert/Update debtor สำเร็จ")

	logger.Info("--- ProcessDebtorMasterData COMPLETED SUCCESSFULLY: %s ---", debtorData.Code)
	return nil
}

// insertOrUpdateDebtor - inserts or updates debtor in PostgreSQL
func insertOrUpdateDebtor(ctx context.Context, db *sql.DB, debtor models.ProcessMongoDebtorModel) error {
	name0 := getDebtorName(debtor.Names)

	query := `
        INSERT INTO debtor (code, name0, taxid, createdat, updatedat)
        VALUES ($1, $2, $3, NOW(), NOW())
        ON CONFLICT (code) DO UPDATE SET
            name0 = EXCLUDED.name0,
            taxid = EXCLUDED.taxid,
            updatedat = NOW()
    `

	_, err := db.ExecContext(ctx, query, debtor.Code, name0, debtor.TaxID)
	if err != nil {
		logger.Error("ไม่สามารถ insert/update debtor: %v", err)
		return err
	}

	logger.Info("Insert/Update debtor สำเร็จ: Code=%s, Name=%s", debtor.Code, name0)
	return nil
}

// TransDebtorDecode - decodes debtor JSON data
func TransDebtorDecode(jsonData string) models.ProcessMongoDebtorModel {
	jsonDecode := map[string]any{}
	err := json.Unmarshal([]byte(jsonData), &jsonDecode)
	if err != nil {
		logger.Error("unmarshal transDebtor: %v", err)
		return models.ProcessMongoDebtorModel{}
	}

	fixedData := myglobal.FixMongoData(jsonDecode)

	var debtorData models.ProcessMongoDebtorModel
	decoderConfig := mapstructure.DecoderConfig{
		Squash:           true,
		WeaklyTypedInput: true,
		Result:           &debtorData,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeHookFunc(time.RFC3339),
		),
	}
	decoder, err := mapstructure.NewDecoder(&decoderConfig)
	if err != nil {
		logger.Error("creating decoder: %v", err)
		return models.ProcessMongoDebtorModel{}
	}

	err = decoder.Decode(fixedData)
	if err != nil {
		logger.Error("decoding debtor: %v", err)
		return models.ProcessMongoDebtorModel{}
	}

	return debtorData
}

// OnConsumeMessageDebtorBulkCreateOrUpdate - handles bulk debtor create/update messages
// รับ JSON array ของ debtor documents จาก Kafka topic: when-debtor-bulk-created
func OnConsumeMessageDebtorBulkCreateOrUpdate(msg string) error {
	return SafeConsumerWrapper("DEBTOR_BULK", func(msg string) error {
		logger.Info("--- ProcessDebtorBulkMasterData START ---")

		var rawArray []map[string]interface{}
		if err := json.Unmarshal([]byte(msg), &rawArray); err != nil {
			logger.Info("ไม่ใช่ JSON array — fallback เป็น single debtor")
			return ProcessDebtorMasterData(msg)
		}

		logger.Info("Received %d debtors for bulk processing", len(rawArray))

		for i, raw := range rawArray {
			rawJSON, err := json.Marshal(raw)
			if err != nil {
				logger.Error("marshal debtor[%d] ล้มเหลว: %v", i, err)
				continue
			}
			if err := ProcessDebtorMasterData(string(rawJSON)); err != nil {
				logger.Error("process debtor[%d] ล้มเหลว: %v", i, err)
			}
		}

		logger.Info("--- ProcessDebtorBulkMasterData COMPLETED: %d items ---", len(rawArray))
		return nil
	})(msg)
}

// OnConsumeMessageDebtorBulkDelete - handles bulk debtor delete messages
// รับ JSON array ของ debtor documents จาก Kafka topic: when-debtor-bulk-deleted
func OnConsumeMessageDebtorBulkDelete(msg string) error {
	return SafeConsumerWrapper("DEBTOR_BULK_DELETE", func(msg string) error {
		logger.Info("--- ProcessDebtorBulkDelete START ---")

		var rawArray []map[string]interface{}
		if err := json.Unmarshal([]byte(msg), &rawArray); err != nil {
			logger.Info("ไม่ใช่ JSON array — fallback เป็น single debtor delete")
			return OnConsumeMessageDebtorDelete(msg)
		}

		logger.Info("Received %d debtors for bulk deletion", len(rawArray))

		for i, raw := range rawArray {
			rawJSON, err := json.Marshal(raw)
			if err != nil {
				logger.Error("marshal debtor[%d] ล้มเหลว: %v", i, err)
				continue
			}

			debtorData := TransDebtorDecode(string(rawJSON))
			if debtorData.HoldingCode == "" || debtorData.Code == "" {
				logger.Warn("debtor[%d] ไม่มี HoldingCode/Code — ข้าม", i)
				continue
			}

			db, err := mypg.PgSqlFastConnect(debtorData.HoldingCode)
			if err != nil {
				logger.Error("debtor[%d] เชื่อมต่อ PG ล้มเหลว: %v", i, err)
				continue
			}

			query := "DELETE FROM debtor WHERE code = $1"
			_, err = db.ExecContext(context.Background(), query, debtorData.Code)
			if err != nil {
				logger.Error("debtor[%d] ลบล้มเหลว: %v", i, err)
				continue
			}
			logger.Info("ลบ debtor สำเร็จ: Code=%s", debtorData.Code)
		}

		logger.Info("--- ProcessDebtorBulkDelete COMPLETED: %d items ---", len(rawArray))
		return nil
	})(msg)
}

// getDebtorName - helper function to get debtor name from language models
func getDebtorName(names []models.ProcessMongoDebtorNameModel) string {
	if len(names) == 0 {
		return ""
	}

	// Return first available name
	return names[0].Name
}
