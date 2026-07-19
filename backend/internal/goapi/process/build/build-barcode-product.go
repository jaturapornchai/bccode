package build

import (
	"context"
	"database/sql"
	"fmt"
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

func ProcessInsertBarCodeListForPostgres(conn interface{}, barcodes *[]models.BarcodeModel) error {
	// เตรียม columns สำหรับ COPY FROM
	columns := []string{
		"holding_code", "barcode", "barcoderef", "itemcode", "name0", "unitcode", "unitname",
		"groupcode", "groupnames", "price1", "price_retail", "barcoderefunitstand", "barcoderefunitdivide",
		"isstock", "itemtype", "checksum", "imageuri",
	}

	// แปลงข้อมูลเป็น format สำหรับ COPY FROM
	rows := make([][]any, len(*barcodes))

	for i, barcode := range *barcodes {
		rows[i] = []any{
			barcode.HoldingCode,
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

	err := mypg.BulkInsertWithCopy(ctx, conn, "productbarcode", columns, rows)
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
	logger.Info("Starting BarcodeRebuildAll for shop %s", holdingCode)
	productBarcodes, err := loadProductBarcodesFromMongo(holdingCode)
	if err != nil {
		logger.Error("loading product barcodes from MongoDB: %v", err)
		return
	}
	if err := rebuildProductBarcodePostgres(holdingCode, &productBarcodes); err != nil {
		logger.Error("rebuilding PostgreSQL productbarcode: %v", err)
		return
	}

	chClient, err := myclickhouse.ClickHouseFastConnect()
	if err != nil || chClient == nil {
		logger.Error("PostgreSQL productbarcode rebuilt, but ClickHouse connection failed: %v", err)
		return
	}
	if err := chClient.Exec(context.Background(), "ALTER TABLE productbarcode DELETE WHERE holding_code = ?", holdingCode); err != nil {
		logger.Error("PostgreSQL productbarcode rebuilt, but ClickHouse cleanup failed: %v", err)
		return
	}
	ProcessInsertBarCodeListForClickHouse(chClient, &productBarcodes)
	logger.Info("Rebuilt %d product barcodes for shop %s", len(productBarcodes), holdingCode)
}

func ProcessBarcodePostgresRebuildAll(holdingCode string) (int, error) {
	productBarcodes, err := loadProductBarcodesFromMongo(holdingCode)
	if err != nil {
		return 0, err
	}
	if err := rebuildProductBarcodePostgres(holdingCode, &productBarcodes); err != nil {
		return 0, err
	}
	logger.Info("Rebuilt %d PostgreSQL product barcodes for shop %s", len(productBarcodes), holdingCode)
	return len(productBarcodes), nil
}

func loadProductBarcodesFromMongo(holdingCode string) ([]models.BarcodeModel, error) {
	mongoClient := myglobal.SafeMongoConnectFast()
	if mongoClient == nil {
		return nil, fmt.Errorf("connect MongoDB for product barcode projection")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	databaseName := config.NewServiceConfig().MongodbDatabaseName()
	cur, err := mongoClient.Database(databaseName).Collection("productbarcodes").Find(
		ctx,
		bson.M{"holdingcode": holdingCode, "deletedat": nil},
	)
	if err != nil {
		return nil, fmt.Errorf("find MongoDB product barcodes: %w", err)
	}
	defer cur.Close(ctx)

	productBarcodes := make([]models.BarcodeModel, 0)
	for cur.Next(ctx) {
		barcodeModel, _ := myglobal.MapBarcodeFromMongoToStruct(myglobal.ProcessProductBarcodeDecode(cur.Current.String()))
		productBarcodes = append(productBarcodes, barcodeModel)
	}
	if err := cur.Err(); err != nil {
		return nil, fmt.Errorf("iterate MongoDB product barcodes: %w", err)
	}
	return productBarcodes, nil
}

func rebuildProductBarcodePostgres(holdingCode string, productBarcodes *[]models.BarcodeModel) error {
	pgDB, err := mypg.PgSqlFastConnect(holdingCode)
	if err != nil {
		return fmt.Errorf("connect PostgreSQL product barcode projection: %w", err)
	}
	if err := TableProductBarcodeCreate(pgDB); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	tx, err := pgDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin PostgreSQL product barcode rebuild: %w", err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, "TRUNCATE TABLE productbarcode"); err != nil {
		return fmt.Errorf("clear PostgreSQL product barcode projection: %w", err)
	}
	if err := ProcessInsertBarCodeListForPostgres(tx, productBarcodes); err != nil {
		return fmt.Errorf("insert PostgreSQL product barcode projection: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit PostgreSQL product barcode projection: %w", err)
	}
	return nil
}
