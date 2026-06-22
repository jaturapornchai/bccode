package build

import (
	"context"
	"database/sql"
	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"smlcloudplatform/internal/goapi/config"
	"smlcloudplatform/internal/goapi/myclickhouse"
	"smlcloudplatform/internal/goapi/myglobal"
	"smlcloudplatform/internal/goapi/mypg"

	"github.com/ClickHouse/clickhouse-go/v2"
)

func ProcessInsertBarCodeListForPostgres(db *sql.DB, barcodes *[]models.BarcodeModel) error {
	// เตรียม columns สำหรับ COPY FROM
	columns := []string{
		"barcode", "barcoderef", "itemcode", "name0", "unitcode", "unitname",
		"groupcode", "groupnames", "price1", "price_retail", "barcoderefunitstand", "barcoderefunitdivide",
		"isstock", "itemtype", "checksum", "imageuri",
	}

	// แปลงข้อมูลเป็น format สำหรับ COPY FROM
	rows := make([][]any, len(*barcodes))

	for i, barcode := range *barcodes {
		rows[i] = []any{
			barcode.Barcode,
			barcode.BarcodeRef,
			barcode.ItemCode,
			barcode.Name0,
			barcode.UnitCode,
			barcode.UnitName,
			barcode.GroupCode,
			barcode.GroupNames,
			barcode.Price1,
			barcode.PriceRetail, // ราคาขายปลีก (keynumber=1 จาก MongoDB)
			barcode.BarcodeRefUnitStand,
			barcode.BarcodeRefUnitDivide,
			barcode.IsStock,
			barcode.ItemType,
			barcode.Checksum,
			barcode.ImageUri, // imageuri จาก MongoDB
		}
	}

	// ใช้ BulkInsertWithCopy
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := mypg.BulkInsertWithCopy(ctx, db, "productbarcode", columns, rows)
	if err != nil {
		return err
	}
	return nil
}

func ProcessInsertBarCodeListForClickHouse(chClient clickhouse.Conn, barcodes *[]models.BarcodeModel) {
	// เตรียมข้อมูลสำหรับ batch insert
	batch, err := chClient.PrepareBatch(context.Background(), `
		INSERT INTO productbarcode (
			holding_code, itemcode, barcode, barcoderef, name0, name1, name2, name3, name4, name5,
			checksum, groupcode, groupnames, unitcode, unitname, price1, price_retail, unitstand, unitdivide, imageuri
		)
	`)
	if err != nil {
		logger.Error("preparing ClickHouse batch: %v", err)
		return
	}

	// วนลูปข้อมูลและเพิ่มลงใน batch
	for _, barcode := range *barcodes {
		// เพิ่มข้อมูลลงใน batch
		err = batch.Append(
			barcode.HoldingCode,          // holding_code
			barcode.ItemCode,             // itemcode
			barcode.Barcode,              // barcode
			barcode.BarcodeRef,           // barcoderef
			barcode.Name0,                // name0
			barcode.Name1,                // name1
			barcode.Name2,                // name2
			barcode.Name3,                // name3
			barcode.Name4,                // name4
			barcode.Name5,                // name5
			barcode.Checksum,             // checksum
			barcode.GroupCode,            // groupcode
			barcode.GroupNames,           // groupnames
			barcode.UnitCode,             // unitcode
			barcode.UnitName,             // unitname
			barcode.Price1,               // price1
			barcode.PriceRetail,          // price_retail (keynumber=1 จาก MongoDB)
			barcode.BarcodeRefUnitStand,  // unitstand (from RefBarCodes)
			barcode.BarcodeRefUnitDivide, // unitdivide (from RefBarCodes)
			barcode.ImageUri,             // imageuri
		)
		if err != nil {
			logger.Error("appending to ClickHouse batch: %v", err)
			return
		}
	}

	// Execute batch
	err = batch.Send()
	if err != nil {
		logger.Error("sending ClickHouse batch: %v", err)
		return
	}
}

func ProcessInsertBarCodeList(dbPg *sql.DB, chClient clickhouse.Conn, barcodes *[]models.BarcodeModel) error {
	if len(*barcodes) == 0 {
		return nil
	}

	// Insert ลง PostgreSQL
	if err := ProcessInsertBarCodeListForPostgres(dbPg, barcodes); err != nil {
		return err
	}

	// Insert ลง ClickHouse
	ProcessInsertBarCodeListForClickHouse(chClient, barcodes)
	return nil
}

func ProcessBarcodeRebuildAll(holdingCode string) {
	mongoClient := myglobal.SafeMongoConnectFast() // Use optimized connection
	if mongoClient == nil {
		logger.Error("MongoConnect failed")
		return
	}

	logger.Info("Starting BarcodeRebuildAll for shop %s", holdingCode)

	pgDb, err := mypg.PgSqlFastConnect(holdingCode)
	if err != nil {
		logger.Info("Failed to connect to Postgres: %v", err)
		return
	}

	// postgresql delete existing rows for this shop (แยก database แล้ว ไม่ต้อง where holding_code)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err = pgDb.ExecContext(ctx, "TRUNCATE TABLE productbarcode")
	if err != nil {
		logger.Error("truncating productbarcode: %v", err)
		return
	}
	logger.Info("Cleared productbarcode rows for holdingCode %s", holdingCode)

	// clickhouse delete existing rows for this shop
	chClient, err := myclickhouse.ClickHouseFastConnect()
	if err != nil || chClient == nil {
		logger.Info("Failed to connect to ClickHouse: %v", err)
		return
	}
	// ใช้ connection pool ไม่ต้อง close

	err = chClient.Exec(context.Background(), "ALTER TABLE productbarcode DELETE WHERE holding_code = ?", holdingCode)
	if err != nil {
		logger.Error("deleting existing ClickHouse productbarcode rows: %v", err)
		return
	}

	// read from MongoDB
	svcConfig := config.NewServiceConfig()
	MongodbDatabaseName := svcConfig.MongodbDatabaseName()
	collection := mongoClient.Database(MongodbDatabaseName).Collection("productbarcodes")
	cur, err := collection.Find(context.Background(), bson.M{"holdingcode": holdingCode, "deletedby": bson.M{"$exists": false}})
	logger.Info("Finding documents in MongoDB collection %s", MongodbDatabaseName)
	if err != nil {
		logger.Error("finding documents: %v", err)
		return
	}
	defer cur.Close(context.Background())

	var productBarcodes []models.BarcodeModel

	for cur.Next(context.Background()) {
		barcodeModel, _ := myglobal.MapBarcodeFromMongoToStruct(myglobal.ProcessProductBarcodeDecode(cur.Current.String()))
		productBarcodes = append(productBarcodes, barcodeModel)
		// ไม่ต้องเก็บ barcodeRefModel อีกแล้ว
	}
	if err := cur.Err(); err != nil {
		logger.Info("Cursor error: %v", err)
		return
	}
	logger.Info("Fetched %d product barcodes from MongoDB for shop %s", len(productBarcodes), holdingCode)

	// insert into Postgres and ClickHouse
	if err := ProcessInsertBarCodeList(pgDb, chClient, &productBarcodes); err != nil {
		logger.Error("BulkInsertWithCopy productbarcode error: %v", err)
		return
	}
	logger.Info("Inserted %d product barcodes for shop %s", len(productBarcodes), holdingCode)
}
