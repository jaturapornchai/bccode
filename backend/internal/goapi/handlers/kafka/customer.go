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
	"runtime/debug"
	"time"

	"github.com/mitchellh/mapstructure"
)

// OnConsumeMessageCustomerCreateOrUpdate - handles customer create/update messages
func OnConsumeMessageCustomerCreateOrUpdate(msg string) error {
	return SafeConsumerWrapper("CUSTOMER", func(msg string) error {
		return ProcessCustomerMasterData(msg)
	})(msg)
}

// OnConsumeMessageCustomerDelete - handles customer delete messages
func OnConsumeMessageCustomerDelete(msg string) error {
	// รับ Message จาก Kafka ที่เป็นการลบข้อมูล Customer
	// msg จะเป็น JSON string ที่มีข้อมูล เช่น {"shopid": "shop123", "code": "C0001"}

	logger.Info("OnConsumeMessageCustomerDelete: %s", msg)

	customerData := TransCustomerDecode(msg)

	if customerData.ShopId == "" || customerData.Code == "" {
		logger.Error("ข้อมูล customer ไม่ถูกต้อง - ไม่มี ShopId หรือ Code")
		return fmt.Errorf("invalid customer data")
	}

	// ลบข้อมูลใน PostgreSQL
	db, err := mypg.PgSqlFastConnect(customerData.ShopId)
	if err != nil {
		logger.Error("ไม่สามารถเชื่อมต่อ PostgreSQL: %v", err)
		return err
	}

	query := "DELETE FROM customer WHERE code = $1"
	_, err = db.ExecContext(context.Background(), query, customerData.Code)
	if err != nil {
		logger.Error("ไม่สามารถลบ customer: %v", err)
		return err
	}

	logger.Info("ลบ customer สำเร็จ: ShopID=%s, Code=%s", customerData.ShopId, customerData.Code)
	return nil
}

// ProcessCustomerMasterData - processes customer master data from Kafka
func ProcessCustomerMasterData(msg string) error {
	// ประมวลผลข้อมูล Master Data: Customer (ลูกค้า)
	logger.Info("--- ProcessCustomerMasterData START ---")

	defer func() {
		if r := recover(); r != nil {
			logger.Info("PANIC recovered in ProcessCustomerMasterData: %v", r)
			logger.Info("Stack trace: %s", debug.Stack())
			logger.Info("Message length: %d characters", len(msg))
		}
	}()

	// Decode customer from JSON message
	logger.Debug("Step 1: กำลัง Decode JSON message... ")
	customerData := TransCustomerDecode(msg)
	logger.Info("Step 1: Decode customer สำเร็จ - ShopID=%s, Code=%s, TaxID=%s, Names=%d",
		customerData.ShopId, customerData.Code, customerData.TaxID, len(customerData.Names))

	if customerData.ShopId == "" || customerData.Code == "" {
		logger.Error("ข้อมูล customer ไม่ถูกต้อง - ShopId='%s', Code='%s'", customerData.ShopId, customerData.Code)
		return fmt.Errorf("invalid customer data - missing ShopId or Code")
	}

	// Connect to database
	logger.Debug("Step 2: กำลังเชื่อมต่อ PostgreSQL...")
	db, err := mypg.PgSqlFastConnect(customerData.ShopId)
	if err != nil {
		logger.Error("ไม่สามารถเชื่อมต่อ database: %v", err)
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	logger.Debug("Step 2 เสร็จสิ้น: เชื่อมต่อ PostgreSQL สำเร็จ")

	ctx := context.Background()

	// Insert or Update customer to PostgreSQL
	logger.Debug("Step 3: กำลัง Insert/Update customer ไปยัง PostgreSQL...")
	err = insertOrUpdateCustomer(ctx, db, customerData)
	if err != nil {
		logger.Error("ไม่สามารถ Insert/Update customer: %v", err)
		return err
	}
	logger.Debug("Step 3 เสร็จสิ้น: Insert/Update customer สำเร็จ")

	logger.Info("--- ProcessCustomerMasterData COMPLETED SUCCESSFULLY: %s ---", customerData.Code)
	return nil
}

// insertOrUpdateCustomer - inserts or updates customer in PostgreSQL
func insertOrUpdateCustomer(ctx context.Context, db *sql.DB, customer models.ProcessMongoCustomerModel) error {
	// Extract first name from names array
	name0 := getCustomerName(customer.Names)

	query := `
        INSERT INTO customer (code, personaltype, name0, taxid, customertype)
        VALUES ($1, $2, $3, $4, $5)
        ON CONFLICT (code) DO UPDATE SET
            personaltype = EXCLUDED.personaltype,
            name0 = EXCLUDED.name0,
            taxid = EXCLUDED.taxid,
            customertype = EXCLUDED.customertype
    `

	_, err := db.ExecContext(ctx, query, customer.Code, customer.PersonalType, name0, customer.TaxID, customer.CustomerType)
	if err != nil {
		logger.Error("ไม่สามารถ insert customer: %v", err)
		return err
	}

	logger.Info("Insert/Update customer สำเร็จ: Code=%s, Name=%s", customer.Code, name0)
	return nil
}

// TransCustomerDecode - decodes customer JSON data
func TransCustomerDecode(jsonData string) models.ProcessMongoCustomerModel {
	jsonDecode := map[string]any{}
	err := json.Unmarshal([]byte(jsonData), &jsonDecode)
	if err != nil {
		logger.Error("unmarshal transCustomer: %v", err)
		return models.ProcessMongoCustomerModel{}
	}

	fixedData := myglobal.FixMongoData(jsonDecode)

	var customerData models.ProcessMongoCustomerModel
	decoderConfig := mapstructure.DecoderConfig{
		Squash:           true,
		WeaklyTypedInput: true,
		Result:           &customerData,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeHookFunc(time.RFC3339),
		),
	}
	decoder, err := mapstructure.NewDecoder(&decoderConfig)
	if err != nil {
		logger.Error("creating decoder: %v", err)
		return models.ProcessMongoCustomerModel{}
	}

	err = decoder.Decode(fixedData)
	if err != nil {
		logger.Error("decoding customer: %v", err)
		return models.ProcessMongoCustomerModel{}
	}

	return customerData
}

// getCustomerName - helper function to get customer name from language models
func getCustomerName(names []models.ProcessMongoCustomerNameModel) string {
	if len(names) == 0 {
		return ""
	}

	// Return first available name
	return names[0].Name
}
