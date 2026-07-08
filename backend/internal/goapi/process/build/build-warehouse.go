package build

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"smlcloudplatform/internal/goapi/config"
	"smlcloudplatform/internal/goapi/myglobal"
	"smlcloudplatform/internal/goapi/mypg"

	"github.com/mitchellh/mapstructure"
)

// Helper function to get Thai name from names array
func getThaiName(names []models.LanguageModel) string {
	for _, name := range names {
		if name.Code == "th" {
			return name.Name
		}
	}
	// If no Thai name found, return first name or empty string
	if len(names) > 0 {
		return names[0].Name
	}
	return ""
}

func processInsertWarehouseList(db *sql.DB, data *[]models.ProcessMongoWarehouseModel) error {
	if len(*data) == 0 {
		return nil
	}

	// เตรียม columns สำหรับ ic_warehouse table
	warehouseColumns := []string{
		"code", "name_1", "status", "guid_code",
	}

	// เตรียม columns สำหรับ ic_shelf table
	shelfColumns := []string{
		"code", "name_1", "whcode", "status", "guid_code",
	}

	// แปลงข้อมูลเป็น format สำหรับ COPY FROM - warehouse
	warehouseRows := make([][]any, 0, len(*data))
	shelfRows := make([][]any, 0)

	for _, warehouse := range *data {
		// Get warehouse name from names array (first thai name)
		warehouseName := getThaiName(warehouse.Names)

		locationCount := 0
		if warehouse.Location != nil {
			locationCount = len(warehouse.Location)
		}

		logger.Info("Processing warehouse: code=%s, name1=%s, location count=%d", warehouse.Code, warehouseName, locationCount)

		// เพิ่มข้อมูล warehouse
		warehouseRows = append(warehouseRows, []any{
			warehouse.Code,
			warehouseName,
			0,  // status - 0 = ใช้งาน
			"", // guid_code - ว่างไว้ก่อน
		})

		// เพิ่มข้อมูล locations (shelves) - เฉพาะเมื่อ location ไม่เป็น nil
		if warehouse.Location != nil {
			for _, location := range warehouse.Location {
				// Get location name from names array (first thai name)
				locationName := getThaiName(location.Names)

				logger.Info("Processing shelf: warehouse_code=%s, shelf_code=%s, shelf_name=%s", warehouse.Code, location.Code, locationName)
				shelfRows = append(shelfRows, []any{
					location.Code,
					locationName,
					warehouse.Code, // whcode - รหัสคลังสินค้า
					0,              // status - 0 = ใช้งาน
					"",             // guid_code - ว่างไว้ก่อน
				})
			}
		} else {
			logger.Info("Warehouse %s has no locations (null)", warehouse.Code)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Insert warehouses
	if len(warehouseRows) > 0 {
		err := mypg.BulkInsertWithCopy(ctx, db, "ic_warehouse", warehouseColumns, warehouseRows)
		if err != nil {
			return fmt.Errorf("failed to insert warehouses: %w", err)
		}
		logger.Info("Inserted %d warehouses", len(warehouseRows))
	}

	// Insert shelves
	if len(shelfRows) > 0 {
		err := mypg.BulkInsertWithCopy(ctx, db, "ic_shelf", shelfColumns, shelfRows)
		if err != nil {
			return fmt.Errorf("failed to insert shelves: %w", err)
		}
		logger.Info("Inserted %d shelves", len(shelfRows))
	} else {
		logger.Info("No shelf data to insert")
	}

	return nil
}

func processWarehouseDecode(jsonData string) models.ProcessMongoWarehouseModel {
	jsonDecode := map[string]any{}
	err := json.Unmarshal([]byte(jsonData), &jsonDecode)
	if err != nil {
		logger.Error("Error unmarshal: %v", err)
		return models.ProcessMongoWarehouseModel{}
	}

	fixedData := myglobal.FixMongoData(jsonDecode)

	var warehouse models.ProcessMongoWarehouseModel
	decoderConfig := mapstructure.DecoderConfig{
		DecodeHook: myglobal.CustomDecodeHookFunc,
		Result:     &warehouse,
		TagName:    "json",
	}
	decoder, err := mapstructure.NewDecoder(&decoderConfig)
	if err != nil {
		logger.Fatal("%v", err)
	}

	err = decoder.Decode(fixedData)
	if err != nil {
		logger.Error("Error decoding map to struct: %v", err)
	}
	return warehouse
}

func ProcessWarehouseRebuildAll(holdingCode string) {
	mongoClient := myglobal.SafeMongoConnectFast() // Use optimized connection
	if mongoClient == nil {
		logger.Error("MongoConnect failed")
		return
	}

	timeStart := time.Now()
	logger.Info("Starting WarehouseRebuildAll for shop %s", holdingCode)

	db, err := mypg.PgSqlFastConnect(holdingCode)
	if err != nil {
		logger.Info("Failed to connect to Postgres: %v", err)
		return
	}

	// delete existing rows for warehouse and shelf tables
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Clear shelves first (due to foreign key constraint)
	_, err = db.ExecContext(ctx, "DELETE FROM ic_shelf")
	if err != nil {
		logger.Error("deleting existing ic_shelf rows: %v", err)
		return
	}
	logger.Info("Cleared ic_shelf rows for holdingCode %s", holdingCode)

	// Clear warehouses
	_, err = db.ExecContext(ctx, "DELETE FROM ic_warehouse")
	if err != nil {
		logger.Error("deleting existing ic_warehouse rows: %v", err)
		return
	}
	logger.Info("Cleared ic_warehouse rows for holdingCode %s", holdingCode)

	// read from MongoDB
	svcConfig := config.NewServiceConfig()
	MongodbDatabaseName := svcConfig.MongodbDatabaseName()
	collection := mongoClient.Database(MongodbDatabaseName).Collection("warehouse")
	cur, err := collection.Find(context.Background(), bson.M{"holdingcode": holdingCode, "deletedat": bson.M{"$exists": false}})
	logger.Info("Finding documents in MongoDB collection warehouse for shop %s", holdingCode)
	if err != nil {
		logger.Error("finding documents: %v", err)
		return
	}
	defer cur.Close(context.Background())

	var warehouses []models.ProcessMongoWarehouseModel

	for cur.Next(context.Background()) {
		warehouses = append(warehouses, processWarehouseDecode(cur.Current.String()))
	}
	if err := cur.Err(); err != nil {
		logger.Info("Cursor error: %v", err)
		return
	}
	logger.Info("Fetched %d warehouses from MongoDB for shop %s", len(warehouses), holdingCode)

	// insert into Postgres
	err = processInsertWarehouseList(db, &warehouses)
	if err != nil {
		logger.Error("inserting warehouses: %v", err)
		return
	}
	logger.Info("Completed WarehouseRebuildAll for shop %s in %v", holdingCode, time.Since(timeStart))
}
