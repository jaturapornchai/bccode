package build

import (
	"context"
	"database/sql"
	"encoding/json"
	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"smlcloudplatform/internal/goapi/config"
	"smlcloudplatform/internal/goapi/myglobal"
	"smlcloudplatform/internal/goapi/mypg"

	"github.com/mitchellh/mapstructure"
)

func processInsertCustomerList(db *sql.DB, data *[]models.ProcessMongoCustomerModel) error {
	if len(*data) == 0 {
		return nil
	}

	// เตรียม columns สำหรับ COPY FROM
	columns := []string{
		"code",
		"personal_type",
		"name0",
		"tax_id",
		"customer_type",
	}

	// แปลงข้อมูลเป็น format สำหรับ COPY FROM
	rows := make([][]any, len(*data))

	for i, customer := range *data {
		customerCode := customer.Code

		// ตรวจสอบข้อมูลที่จำเป็น
		if customerCode == "" {
			logger.Warn("Customer at index %d has empty code, skipping", i)
			continue
		}
		personalType := customer.PersonalType
		name0 := ""
		if len(customer.Names) > 0 {
			name0 = customer.Names[0].Name
		}

		customerType := customer.CustomerType
		taxID := customer.TaxID

		rows[i] = []any{
			customerCode,
			personalType,
			name0,
			taxID,
			customerType,
		}
	}

	// ใช้ BulkInsertWithCopy
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return mypg.BulkInsertWithCopy(ctx, db, "customer", columns, rows)
}

func processCustomerDecode(jsonData string) models.ProcessMongoCustomerModel {
	jsonDecode := map[string]any{}
	err := json.Unmarshal([]byte(jsonData), &jsonDecode)
	if err != nil {
		logger.Error("Error unmarshal: %v", err)
		return models.ProcessMongoCustomerModel{}
	}

	fixedData := myglobal.FixMongoData(jsonDecode)

	var customer models.ProcessMongoCustomerModel
	decoderConfig := mapstructure.DecoderConfig{
		DecodeHook: myglobal.CustomDecodeHookFunc,
		Result:     &customer,
		TagName:    "json",
	}
	decoder, err := mapstructure.NewDecoder(&decoderConfig)
	if err != nil {
		logger.Error("creating decoder: %v", err)
		return models.ProcessMongoCustomerModel{}
	}

	err = decoder.Decode(fixedData)
	if err != nil {
		logger.Error("Error decoding map to struct: %v", err)
	}
	return customer
}

func ProcessCustomerRebuildAll(holdingCode string) {
	mongoClient := myglobal.SafeMongoConnectFast() // Use optimized connection
	if mongoClient == nil {
		logger.Error("MongoConnect failed")
		return
	}

	timeStart := time.Now()
	logger.Info("Starting CustomerRebuildAll for shop %s", holdingCode)

	db, err := mypg.PgSqlFastConnect(holdingCode)
	if err != nil {
		logger.Info("Failed to connect to Postgres: %v", err)
		return
	}

	// delete existing rows for this shop
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_, err = db.ExecContext(ctx, "TRUNCATE TABLE customer")
	if err != nil {
		logger.Error("deleting existing customer rows: %v", err)
		return
	}
	logger.Info("Cleared customer rows for holdingCode %s", holdingCode)

	// read from MongoDB
	svcConfig := config.NewServiceConfig()
	MongodbDatabaseName := svcConfig.MongodbDatabaseName()
	collection := mongoClient.Database(MongodbDatabaseName).Collection("customers")
	cur, err := collection.Find(context.Background(), bson.M{"holdingcode": holdingCode, "deletedat": bson.M{"$exists": false}})
	logger.Info("Finding documents in MongoDB collection %s", MongodbDatabaseName)
	if err != nil {
		logger.Error("finding documents: %v", err)
		return
	}
	defer cur.Close(context.Background())

	var customers []models.ProcessMongoCustomerModel

	for cur.Next(context.Background()) {
		customers = append(customers, processCustomerDecode(cur.Current.String()))
	}
	if err := cur.Err(); err != nil {
		logger.Info("Cursor error: %v", err)
		return
	}
	logger.Info("Fetched %d customers from MongoDB for shop %s", len(customers), holdingCode)

	// insert into Postgres
	err = processInsertCustomerList(db, &customers)
	if err != nil {
		logger.Error("inserting customers: %v", err)
		return
	}
	logger.Info("Inserted %d customers into Postgres for shop %s", len(customers), holdingCode)
	logger.Info("Completed CustomerRebuildAll for shop %s in %v", holdingCode, time.Since(timeStart))
}
