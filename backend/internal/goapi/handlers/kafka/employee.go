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

// OnConsumeMessageEmployeeCreateOrUpdate - handles employee create/update messages
func OnConsumeMessageEmployeeCreateOrUpdate(msg string) error {
	return SafeConsumerWrapper("EMPLOYEE", func(msg string) error {
		return ProcessEmployeeMasterData(msg)
	})(msg)
}

// OnConsumeMessageEmployeeDelete - handles employee delete messages
func OnConsumeMessageEmployeeDelete(msg string) error {
	// รับ Message จาก Kafka ที่เป็นการลบข้อมูล Employee
	// msg จะเป็น JSON string ที่มีข้อมูล เช่น {"holdingcode": "shop123", "code": "E0001"}

	logger.Info("OnConsumeMessageEmployeeDelete: %s", msg)

	employeeData := TransEmployeeDecode(msg)

	if employeeData.HoldingCode == "" || employeeData.Code == "" {
		logger.Error("ข้อมูล employee ไม่ถูกต้อง - ไม่มี HoldingCode หรือ Code")
		return fmt.Errorf("invalid employee data")
	}

	// ลบข้อมูลใน PostgreSQL
	db, err := mypg.PgSqlFastConnect(employeeData.HoldingCode)
	if err != nil {
		logger.Error("ไม่สามารถเชื่อมต่อ PostgreSQL: %v", err)
		return err
	}

	query := "DELETE FROM erp_user WHERE code = $1"
	_, err = db.ExecContext(context.Background(), query, employeeData.Code)
	if err != nil {
		logger.Error("ไม่สามารถลบ employee: %v", err)
		return err
	}

	logger.Info("ลบ employee สำเร็จ: HoldingCode=%s, Code=%s", employeeData.HoldingCode, employeeData.Code)
	return nil
}

// ProcessEmployeeMasterData - processes employee master data from Kafka
func ProcessEmployeeMasterData(msg string) error {
	// ประมวลผลข้อมูล Master Data: Employee (พนักงาน)
	logger.Info("--- ProcessEmployeeMasterData START ---")

	defer func() {
		if r := recover(); r != nil {
			logger.Info("PANIC recovered in ProcessEmployeeMasterData: %v", r)
			logger.Info("Stack trace: %s", debug.Stack())
			logger.Info("Message length: %d characters", len(msg))
		}
	}()

	// Decode employee from JSON message
	logger.Debug("Step 1: กำลัง Decode JSON message... ")
	employeeData := TransEmployeeDecode(msg)
	logger.Info("Step 1: Decode employee สำเร็จ - HoldingCode=%s, Code=%s, Name=%s",
		employeeData.HoldingCode, employeeData.Code, employeeData.Name)

	if employeeData.HoldingCode == "" || employeeData.Code == "" {
		logger.Error("ข้อมูล employee ไม่ถูกต้อง - HoldingCode='%s', Code='%s'", employeeData.HoldingCode, employeeData.Code)
		return fmt.Errorf("invalid employee data - missing HoldingCode or Code")
	}

	// Connect to database
	logger.Debug("Step 2: กำลังเชื่อมต่อ PostgreSQL...")
	db, err := mypg.PgSqlFastConnect(employeeData.HoldingCode)
	if err != nil {
		logger.Error("ไม่สามารถเชื่อมต่อ database: %v", err)
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	logger.Debug("Step 2 เสร็จสิ้น: เชื่อมต่อ PostgreSQL สำเร็จ")

	ctx := context.Background()

	// Insert or Update employee to PostgreSQL
	logger.Debug("Step 3: กำลัง Insert/Update employee ไปยัง PostgreSQL...")
	err = insertOrUpdateEmployee(ctx, db, employeeData)
	if err != nil {
		logger.Error("ไม่สามารถ Insert/Update employee: %v", err)
		return err
	}
	logger.Debug("Step 3 เสร็จสิ้น: Insert/Update employee สำเร็จ")

	logger.Info("--- ProcessEmployeeMasterData COMPLETED SUCCESSFULLY: %s ---", employeeData.Code)
	return nil
}

// insertOrUpdateEmployee - inserts or updates employee in PostgreSQL
func insertOrUpdateEmployee(ctx context.Context, db *sql.DB, employee models.ProcessMongoEmployeeModel) error {
	// Set default values
	name := employee.Name
	if name == "" {
		name = employee.Code
	}
	password := employee.Password
	if password == "" {
		password = "123456" // default password
		logger.Warn("Employee %s has empty password, using default", employee.Code)
	}

	query := `
        INSERT INTO erp_user (code, name, password, position, department, approval_role, max_approval_amount)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        ON CONFLICT (code) DO UPDATE SET
            name = EXCLUDED.name,
            password = EXCLUDED.password,
            position = EXCLUDED.position,
            department = EXCLUDED.department,
            approval_role = EXCLUDED.approval_role,
            max_approval_amount = EXCLUDED.max_approval_amount
    `

	_, err := db.ExecContext(ctx, query,
		employee.Code,
		name,
		password,
		employee.Position,
		employee.Department,
		employee.ApprovalRole,
		employee.MaxApprovalAmount,
	)
	if err != nil {
		logger.Error("ไม่สามารถ insert employee: %v", err)
		return err
	}

	logger.Info("Insert/Update employee สำเร็จ: Code=%s, Name=%s, Position=%s, ApprovalRole=%d",
		employee.Code, name, employee.Position, employee.ApprovalRole)
	return nil
}

// TransEmployeeDecode - decodes employee JSON data
func TransEmployeeDecode(jsonData string) models.ProcessMongoEmployeeModel {
	jsonDecode := map[string]any{}
	err := json.Unmarshal([]byte(jsonData), &jsonDecode)
	if err != nil {
		logger.Error("unmarshal transEmployee: %v", err)
		return models.ProcessMongoEmployeeModel{}
	}

	fixedData := myglobal.FixMongoData(jsonDecode)

	var employeeData models.ProcessMongoEmployeeModel
	decoderConfig := mapstructure.DecoderConfig{
		Squash:           true,
		WeaklyTypedInput: true,
		Result:           &employeeData,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeHookFunc(time.RFC3339),
		),
	}
	decoder, err := mapstructure.NewDecoder(&decoderConfig)
	if err != nil {
		logger.Error("creating decoder: %v", err)
		return models.ProcessMongoEmployeeModel{}
	}

	err = decoder.Decode(fixedData)
	if err != nil {
		logger.Error("decoding employee: %v", err)
		return models.ProcessMongoEmployeeModel{}
	}

	return employeeData
}
