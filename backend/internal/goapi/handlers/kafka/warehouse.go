package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"smlcloudplatform/internal/goapi/logger"

	"smlcloudplatform/internal/goapi/models"
	"smlcloudplatform/internal/goapi/mypg"
	build "smlcloudplatform/internal/goapi/process/build"
)

// OnConsumeMessageWarehouseCreateOrUpdate - handles warehouse create/update messages
func OnConsumeMessageWarehouseCreateOrUpdate(msg string) error {
	logger.Info("OnConsumeMessageWarehouseCreateOrUpdate: %s", msg)

	// Add safety check for message processing
	defer func() {
		if r := recover(); r != nil {
			logger.Info("Panic recovered in OnConsumeMessageWarehouseCreateOrUpdate: %v", r)
		}
	}()

	WarehouseBuild(msg)
	return nil
}

// OnConsumeMessageWarehouseDelete - handles warehouse delete messages
func OnConsumeMessageWarehouseDelete(msg string) error {
	logger.Info("OnConsumeMessageWarehouseDelete: %s", msg)

	// Add safety check for message processing
	defer func() {
		if r := recover(); r != nil {
			logger.Info("Panic recovered in OnConsumeMessageWarehouseDelete: %v", r)
		}
	}()

	warehouseData := WarehouseDecode(msg)

	if warehouseData.HoldingCode == "" || warehouseData.Code == "" {
		logger.Error("Invalid warehouse data - missing HoldingCode or Code")
		return fmt.Errorf("invalid warehouse data")
	}

	build.DatabaseChecker(warehouseData.HoldingCode, false) // เช็ค database

	err := DeleteWarehouseFromPostgreSQL(warehouseData)
	if err != nil {
		logger.Error("deleting warehouse: %v", err)
		return err
	}

	return nil
}

// WarehouseBuild - builds warehouse data from Kafka message
func WarehouseBuild(msg string) {
	warehouseData := WarehouseDecode(msg)
	if warehouseData.HoldingCode == "" {
		logger.Warn("Warehouse data missing HoldingCode")
		return
	}

	build.DatabaseChecker(warehouseData.HoldingCode, false)
	WarehouseInsertOrUpdateToPostgreSQL(warehouseData)
}

// WarehouseDecode - decodes warehouse message from JSON
func WarehouseDecode(msg string) models.MongoWarehouseModel {
	var warehouse models.MongoWarehouseModel
	err := json.Unmarshal([]byte(msg), &warehouse)
	if err != nil {
		logger.Error("decoding warehouse JSON: %v", err)
		return models.MongoWarehouseModel{}
	}
	return warehouse
}

// WarehouseInsertOrUpdateToPostgreSQL - inserts or updates warehouse data in PostgreSQL
func WarehouseInsertOrUpdateToPostgreSQL(warehouseData models.MongoWarehouseModel) {
	logger.Info("Processing warehouse: HoldingCode=%s, Code=%s", warehouseData.HoldingCode, warehouseData.Code)

	db, err := mypg.PgSqlFastConnect(warehouseData.HoldingCode)
	if err != nil {
		logger.Info("Failed to connect to Postgres: %v", err)
		return
	}

	ctx := context.Background()

	// Get warehouse name (first Thai name)
	warehouseName := ""
	for _, name := range warehouseData.Names {
		if name.Code == "th" {
			warehouseName = name.Name
			break
		}
	}
	if warehouseName == "" && len(warehouseData.Names) > 0 {
		warehouseName = warehouseData.Names[0].Name
	}

	// Delete existing warehouse and shelves
	_, err = db.ExecContext(ctx, "DELETE FROM ic_shelf WHERE whcode = $1", warehouseData.Code)
	if err != nil {
		logger.Error("deleting existing shelves for warehouse %s: %v", warehouseData.Code, err)
	}

	_, err = db.ExecContext(ctx, "DELETE FROM ic_warehouse WHERE code = $1", warehouseData.Code)
	if err != nil {
		logger.Error("deleting existing warehouse %s: %v", warehouseData.Code, err)
	}

	// Insert warehouse
	_, err = db.ExecContext(ctx,
		"INSERT INTO ic_warehouse (code, name_1, status, guid_code) VALUES ($1, $2, $3, $4)",
		warehouseData.Code, warehouseName, 0, "")
	if err != nil {
		logger.Error("inserting warehouse %s: %v", warehouseData.Code, err)
		return
	}

	logger.Info("Inserted/Updated warehouse: %s", warehouseData.Code)

	// Insert shelves if any
	if warehouseData.Location != nil {
		for _, location := range warehouseData.Location {
			locationName := ""
			for _, name := range location.Names {
				if name.Code == "th" {
					locationName = name.Name
					break
				}
			}
			if locationName == "" && len(location.Names) > 0 {
				locationName = location.Names[0].Name
			}

			_, err = db.ExecContext(ctx,
				"INSERT INTO ic_shelf (code, name_1, whcode, status, guid_code) VALUES ($1, $2, $3, $4, $5)",
				location.Code, locationName, warehouseData.Code, 0, "")
			if err != nil {
				logger.Error("inserting shelf %s: %v", location.Code, err)
			} else {
				logger.Info("Inserted shelf: %s for warehouse: %s", location.Code, warehouseData.Code)
			}
		}
	}
}

// DeleteWarehouseFromPostgreSQL - deletes warehouse data from PostgreSQL
func DeleteWarehouseFromPostgreSQL(warehouseData models.MongoWarehouseModel) error {
	logger.Info("Deleting warehouse: HoldingCode=%s, Code=%s", warehouseData.HoldingCode, warehouseData.Code)

	db, err := mypg.PgSqlFastConnect(warehouseData.HoldingCode)
	if err != nil {
		return fmt.Errorf("failed to connect to Postgres: %v", err)
	}

	ctx := context.Background()

	// Delete shelves first (due to foreign key constraint)
	_, err = db.ExecContext(ctx, "DELETE FROM ic_shelf WHERE whcode = $1", warehouseData.Code)
	if err != nil {
		logger.Error("deleting shelves for warehouse %s: %v", warehouseData.Code, err)
	}
	// Delete warehouse
	_, err = db.ExecContext(ctx, "DELETE FROM ic_warehouse WHERE code = $1", warehouseData.Code)
	if err != nil {
		return fmt.Errorf("error deleting warehouse %s: %v", warehouseData.Code, err)
	}

	logger.Info("Deleted warehouse: %s", warehouseData.Code)
	return nil
}
