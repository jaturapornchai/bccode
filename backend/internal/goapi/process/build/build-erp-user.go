package build

import (
	"context"
	"database/sql"
	"encoding/json"
	"smlcloudplatform/internal/goapi/config"
	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/models"
	"smlcloudplatform/internal/goapi/myglobal"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"smlcloudplatform/internal/goapi/mypg"

	"github.com/mitchellh/mapstructure"
)

func processInsertErpUserList(db *sql.DB, data *[]models.ProcessMongoEmployeeModel) error {
	if len(*data) == 0 {
		return nil
	}

	// เตรียม columns สำหรับ COPY FROM (รวมฟิลด์ระบบอนุมัติ)
	columns := []string{
		"code", "name", "password",
		"position", "department", "approval_role", "max_approval_amount",
	}

	// แปลงข้อมูลเป็น format สำหรับ COPY FROM
	rows := make([][]any, len(*data))

	for i, employee := range *data {
		employeeCode := employee.Code
		employeeName := employee.Name
		employeePassword := employee.Password

		// ตรวจสอบข้อมูลที่จำเป็น
		if employeeCode == "" {
			logger.Warn("Employee at index %d has empty code, skipping", i)
			continue
		}
		if employeeName == "" {
			employeeName = employeeCode // ใช้ code เป็น name ถ้าไม่มี name
		}
		if employeePassword == "" {
			employeePassword = "123456" // default password (ควรเปลี่ยนในการใช้งานจริง)
			logger.Warn("Employee %s has empty password, using default", employeeCode)
		}

		rows[i] = []any{
			employeeCode,
			employeeName,
			employeePassword,
			employee.Position,
			employee.Department,
			employee.ApprovalRole,
			employee.MaxApprovalAmount,
		}
	}

	// ใช้ BulkInsertWithCopy
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return mypg.BulkInsertWithCopy(ctx, db, "erp_user", columns, rows)
}

func processEmployeeDecode(jsonData string) models.ProcessMongoEmployeeModel {
	jsonDecode := map[string]any{}
	err := json.Unmarshal([]byte(jsonData), &jsonDecode)
	if err != nil {
		logger.Error("Error unmarshal: %v", err)
		return models.ProcessMongoEmployeeModel{}
	}

	fixedData := myglobal.FixMongoData(jsonDecode)

	var employee models.ProcessMongoEmployeeModel
	decoderConfig := mapstructure.DecoderConfig{
		DecodeHook: myglobal.CustomDecodeHookFunc,
		Result:     &employee,
		TagName:    "json",
	}
	decoder, err := mapstructure.NewDecoder(&decoderConfig)
	if err != nil {
		logger.Error("creating decoder: %v", err)
		return models.ProcessMongoEmployeeModel{}
	}

	err = decoder.Decode(fixedData)
	if err != nil {
		logger.Error("Error decoding map to struct: %v", err)
	}
	return employee
}

func ProcessErpUserRebuildAll(holdingCode string) {
	mongoClient := myglobal.SafeMongoConnectFast() // Use optimized connection
	if mongoClient == nil {
		logger.Error("MongoConnect failed")
		return
	}

	timeStart := time.Now()
	logger.Info("Starting ErpUserRebuildAll for shop %s", holdingCode)

	db, err := mypg.PgSqlFastConnect(holdingCode)
	if err != nil {
		logger.Info("Failed to connect to Postgres: %v", err)
		return
	}

	// delete existing rows for this shop
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_, err = db.ExecContext(ctx, "TRUNCATE TABLE erp_user")
	if err != nil {
		logger.Error("deleting existing erp_user rows: %v", err)
		return
	}
	logger.Info("Cleared erp_user rows for holdingCode %s", holdingCode)

	// read from MongoDB
	svcConfig := config.NewServiceConfig()
	MongodbDatabaseName := svcConfig.MongodbDatabaseName()
	collection := mongoClient.Database(MongodbDatabaseName).Collection("employees")
	cur, err := collection.Find(context.Background(), bson.M{"holdingcode": holdingCode, "deletedat": bson.M{"$exists": false}})
	logger.Info("Finding documents in MongoDB collection %s", MongodbDatabaseName)
	if err != nil {
		logger.Error("finding documents: %v", err)
		return
	}
	defer cur.Close(context.Background())

	var employees []models.ProcessMongoEmployeeModel

	for cur.Next(context.Background()) {
		employees = append(employees, processEmployeeDecode(cur.Current.String()))
	}
	if err := cur.Err(); err != nil {
		logger.Info("Cursor error: %v", err)
		return
	}
	logger.Info("Fetched %d employees from MongoDB for shop %s", len(employees), holdingCode)

	// insert into Postgres
	err = processInsertErpUserList(db, &employees)
	if err != nil {
		logger.Error("inserting employees: %v", err)
		return
	}
	logger.Info("Inserted %d employees into Postgres for shop %s", len(employees), holdingCode)
	logger.Info("Completed ErpUserRebuildAll for shop %s in %v", holdingCode, time.Since(timeStart))
}
