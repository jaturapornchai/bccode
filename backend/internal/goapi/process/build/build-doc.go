package build

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/models"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mitchellh/mapstructure"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"smlcloudplatform/internal/goapi/config"
	"smlcloudplatform/internal/goapi/myclickhouse"
	"smlcloudplatform/internal/goapi/myglobal"
	"smlcloudplatform/internal/goapi/mypg"
)

func DocDecode(jsonData string) models.MongoDocModel {
	// Decode JSON data from MongoDB into MongoDocModel struct

	jsonDecode := map[string]any{}
	err := json.Unmarshal([]byte(jsonData), &jsonDecode)
	if err != nil {
		logger.Error("Failed to unmarshal JSON data: %v", err)
		return models.MongoDocModel{}
	}

	fixedData := myglobal.FixMongoData(jsonDecode)

	var docData models.MongoDocModel
	decoderConfig := mapstructure.DecoderConfig{
		DecodeHook: myglobal.CustomDecodeHookFunc,
		Result:     &docData,
		TagName:    "json",
	}
	decoder, err := mapstructure.NewDecoder(&decoderConfig)
	if err != nil {
		logger.Fatal("Failed to create decoder: %v", err)
	}

	err = decoder.Decode(fixedData)
	if err != nil {
		logger.Error("Failed to decode map to struct: %v", err)
		return models.MongoDocModel{}
	}
	return docData
}

func RebuildDocFromMongoWhereTransFlagsDeleteOldData(postgresConnect *sql.DB, holdingCode string, transaction models.StockTransactionStruct) {
	logger.Debug("Starting deletion of old data for %s (holdingCode: %s)", transaction.Name, holdingCode)

	var flagStrings []string
	for _, flag := range transaction.Flags {
		flagStrings = append(flagStrings, strconv.Itoa(flag))
	}
	flagsInClause := strings.Join(flagStrings, ",")

	// ⚡ OPTIMIZATION: ลบ PostgreSQL และ ClickHouse แบบขนาน
	var wg sync.WaitGroup
	var pgDeleted int64

	// PostgreSQL deletions (sequential within itself due to FK constraints)
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Debug("Deleting old records from PostgreSQL (flags: %s)", flagsInClause)

		// ลบ doc ที่มี transflag ในรายการ
		queryDelete := "DELETE from doc where transflag in (" + flagsInClause + ")"
		result, err := postgresConnect.ExecContext(context.Background(), queryDelete)
		if err != nil {
			logger.Error("Failed to delete old doc records: %v", err)
		} else {
			rowsAffected, _ := result.RowsAffected()
			pgDeleted += rowsAffected
			logger.Debug("Deleted %d old doc records", rowsAffected)
		}

		// ลบ docpayment ที่มี transflag ในรายการ
		queryDeletePayment := "DELETE from docpayment where docno in (SELECT docno from doc where transflag in (" + flagsInClause + "));"
		result, err = postgresConnect.ExecContext(context.Background(), queryDeletePayment)
		if err != nil {
			logger.Error("Failed to delete old docpayment records: %v", err)
		} else {
			rowsAffected, _ := result.RowsAffected()
			pgDeleted += rowsAffected
			logger.Debug("Deleted %d old docpayment records", rowsAffected)
		}

		// ลบ docref ที่มี transflag ในรายการ
		queryDeleteRef := "DELETE from docref where docnotransflag in (" + flagsInClause + ")"
		result, err = postgresConnect.ExecContext(context.Background(), queryDeleteRef)
		if err != nil {
			logger.Error("Failed to delete old docref records: %v", err)
		} else {
			rowsAffected, _ := result.RowsAffected()
			pgDeleted += rowsAffected
			logger.Debug("Deleted %d old docref records", rowsAffected)
		}
	}()

	// ⚡ OPTIMIZATION: ClickHouse deletions - ทำแบบขนานทั้ง 3 ตาราง และไม่รอ sync
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Debug("Deleting old records from ClickHouse (async)")

		clickhouseConn, err := myclickhouse.ClickHouseFastConnect()
		if err != nil {
			logger.Error("Failed to connect to ClickHouse: %v", err)
			return
		}

		// ⚡ ลบทั้ง 3 ตารางพร้อมกัน โดยไม่รอ mutations_sync
		var chWg sync.WaitGroup
		chWg.Add(3)

		go func() {
			defer chWg.Done()
			deleteDocQuery := fmt.Sprintf("ALTER TABLE %s DELETE WHERE holding_code = '%s' AND transflag IN (%s)", myclickhouse.TableName("doc"), holdingCode, flagsInClause)
			if err := myclickhouse.ExecuteCommand(context.Background(), clickhouseConn, deleteDocQuery); err != nil {
				logger.Error("Failed to delete old doc records from ClickHouse: %v", err)
			}
		}()

		go func() {
			defer chWg.Done()
			deletePaymentQuery := fmt.Sprintf("ALTER TABLE %s DELETE WHERE holding_code = '%s' AND trans_flag IN (%s)", myclickhouse.TableName("docpayment"), holdingCode, flagsInClause)
			if err := myclickhouse.ExecuteCommand(context.Background(), clickhouseConn, deletePaymentQuery); err != nil {
				logger.Error("Failed to delete old docpayment records from ClickHouse: %v", err)
			}
		}()

		go func() {
			defer chWg.Done()
			deleteRefQuery := fmt.Sprintf("ALTER TABLE %s DELETE WHERE holding_code = '%s' AND docnotransflag IN (%s)", myclickhouse.TableName("docref"), holdingCode, flagsInClause)
			if err := myclickhouse.ExecuteCommand(context.Background(), clickhouseConn, deleteRefQuery); err != nil {
				logger.Error("Failed to delete old docref records from ClickHouse: %v", err)
			}
		}()

		chWg.Wait()
		logger.Debug("ClickHouse delete mutations submitted (async)")
	}()

	wg.Wait()
	logger.Success("Deleted %d total records from PostgreSQL for %s", pgDeleted, transaction.Name)
}

func RebuildDocFromMongoWhereTransFlags(mongoClient *mongo.Client, postgresDB *sql.DB, holdingCode string, transaction models.StockTransactionStruct) {
	logger.Info("Starting rebuild for %s (Doc + DocDetail) (holdingCode: %s, flags: %v)", transaction.Name, holdingCode, transaction.Flags)

	// รวม Doc และ DocDetail เข้าด้วยกัน เพราะอยู่ใน JSON MongoDB เดียวกัน
	// สร้าง doc, docref, docpayment, docdetail ใน PostgreSQL และ ClickHouse
	// function นี้เป็นการสร้างใหม่ทั้งหมด โดยลบออกแล้วสร้างใหม่

	// ลบข้อมูลเก่าใน postgresql และ clickhouse (ทั้ง doc และ docdetail)
	RebuildDocFromMongoWhereTransFlagsDeleteOldData(postgresDB, holdingCode, transaction)
	RebuildDocDetailFromMongoWhereTransFlagsDeleteOldData(mongoClient, postgresDB, holdingCode, transaction)

	// transaction.Flags เป็น []int
	svcConfig := config.NewServiceConfig()
	MongodbDatabaseName := svcConfig.MongodbDatabaseName()
	collection := mongoClient.Database(MongodbDatabaseName).Collection(transaction.Name)
	// deletedby ถ้ามีค่า ไม่ดึง เพราะถือว่าลบไปแล้ว
	logger.Debug("Querying MongoDB for %s documents", transaction.Name)

	cur, err := collection.Find(context.Background(), bson.M{"holdingcode": holdingCode, "transflag": bson.M{"$in": transaction.Flags}, "deletedat": bson.M{"$exists": false}})
	if err != nil {
		logger.Error("Failed to find documents: %v", err)
		return
	}
	defer cur.Close(context.Background())

	// เอาไว้รับข้อมูลจาก MongoDB เพื่อนำไปประมวบผลต่อไป
	var postgresDocData []models.DocStruct
	var postgresDocRefData []models.DocRefStruct
	var docPaymentDataList []models.DocPaymentStruct
	var postgresDocDetailData []models.DocDetailStruct

	totalDocsProcessed := 0
	totalDetailsProcessed := 0
	batchCount := 0

	// Process documents from MongoDB
	for cur.Next(context.Background()) {
		docData := DocDecode(cur.Current.String())
		if docData.DocNo == "" {
			docData.DocNo = "UNKNOWN"
		}

		// เอกสารที่ถูกลบ (DeletedAt != nil) ให้ข้าม แต่เอกสารที่ยกเลิก (IsCancel) ต้องเก็บไว้
		if docData.DeletedAt == nil {
			transFlag := docData.TransFlag

			// 1. Process Doc, DocRef, DocPayment
			docDataStruct, docPaymentStruct := myglobal.MapDocStructFromMongo(docData, holdingCode)
			postgresDocData = append(postgresDocData, docDataStruct)
			docPaymentDataList = append(docPaymentDataList, docPaymentStruct)

			// ถ้ามีการอ้างอิงเอกสาร ให้เพิ่มใน postgresDocRefData
			for _, refNo := range docData.DocReferences {
				postgresDocRefData = append(postgresDocRefData, myglobal.MapDocRefStruct(docData.DocNo, transFlag, refNo))
			}

			totalDocsProcessed++

			// 2. Process DocDetail from same JSON
			transCalc := myglobal.GetTransactionMultiplier(transFlag)
			calcSeq := 1
			lineNumber := 0

			for _, detail := range docData.Details {
				if transFlag == 72 {
					if detail.CalcFlag == -1 {
						// Transfer transaction - create 2 records
						lineNumber++
						postgresDocDetailData = append(postgresDocDetailData,
							myglobal.MapDocDetailFromMongo(docData, detail, transFlag, -1, 1, "", 1.0, 1.0, detail.WhCode, detail.LocationCode, detail.Qty*-1, lineNumber))
						lineNumber++
						postgresDocDetailData = append(postgresDocDetailData,
							myglobal.MapDocDetailFromMongo(docData, detail, transFlag, 1, 2, "", 1.0, 1.0, detail.ToWhCode, detail.ToLocationCode, detail.Qty, lineNumber))
						totalDetailsProcessed += 2
					}
				} else {
					// Other transactions - create 1 record
					lineNumber++
					detailGenerated := myglobal.MapDocDetailFromMongo(docData, detail, transFlag, transCalc, calcSeq, "", 1.0, 1.0, detail.WhCode, detail.LocationCode, detail.Qty*transCalc, lineNumber)
					postgresDocDetailData = append(postgresDocDetailData, detailGenerated)
					totalDetailsProcessed++
				}
			}

			// Batch insert when reaching 10000 docs
			if len(postgresDocData) >= 10000 {
				batchCount++
				logger.Info("%s Inserting batch #%d (%d docs, %d details) for holdingCode %s", transaction.Name, batchCount, len(postgresDocData), len(postgresDocDetailData), holdingCode)

				// ⚡ OPTIMIZATION: Insert PostgreSQL และ ClickHouse แบบขนาน
				var insertWg sync.WaitGroup

				// Copy data for goroutines
				docDataCopy := make([]models.DocStruct, len(postgresDocData))
				copy(docDataCopy, postgresDocData)
				docRefDataCopy := make([]models.DocRefStruct, len(postgresDocRefData))
				copy(docRefDataCopy, postgresDocRefData)
				docPaymentDataCopy := make([]models.DocPaymentStruct, len(docPaymentDataList))
				copy(docPaymentDataCopy, docPaymentDataList)
				docDetailDataCopy := make([]models.DocDetailStruct, len(postgresDocDetailData))
				copy(docDetailDataCopy, postgresDocDetailData)

				// PostgreSQL insert
				insertWg.Add(1)
				go func() {
					defer insertWg.Done()
					if err := mypg.InsertDocListToPostgreSql(context.Background(), postgresDB, docDataCopy, docRefDataCopy, docPaymentDataCopy); err != nil {
						logger.Error("Failed to insert doc batch to PostgreSQL: %v", err)
					}
					if len(docDetailDataCopy) > 0 {
						if err := mypg.InsertDocDetailListToPostgreSql(context.Background(), postgresDB, holdingCode, docDetailDataCopy); err != nil {
							logger.Error("Failed to insert docdetail batch to PostgreSQL: %v", err)
						}
					}
				}()

				// ClickHouse insert (parallel)
				insertWg.Add(1)
				go func() {
					defer insertWg.Done()
					myclickhouse.InsertDocListToClickHouse(context.Background(), holdingCode, docDataCopy, docRefDataCopy, docPaymentDataCopy)
					if len(docDetailDataCopy) > 0 {
						myclickhouse.InsertDocDetailListToClickHouse(context.Background(), holdingCode, docDetailDataCopy)
					}
				}()

				insertWg.Wait()

				postgresDocData = postgresDocData[:0]
				postgresDocRefData = postgresDocRefData[:0]
				docPaymentDataList = docPaymentDataList[:0]
				postgresDocDetailData = postgresDocDetailData[:0]
			}
		}
	}

	// Insert remaining records
	if len(postgresDocData) > 0 {
		batchCount++
		logger.Info("%s Inserting final batch #%d (%d docs, %d details) for holdingCode %s", transaction.Name, batchCount, len(postgresDocData), len(postgresDocDetailData), holdingCode)

		// ⚡ OPTIMIZATION: Insert PostgreSQL และ ClickHouse แบบขนาน
		var insertWg sync.WaitGroup

		// PostgreSQL insert
		insertWg.Add(1)
		go func() {
			defer insertWg.Done()
			if err := mypg.InsertDocListToPostgreSql(context.Background(), postgresDB, postgresDocData, postgresDocRefData, docPaymentDataList); err != nil {
				logger.Error("Failed to insert final doc batch to PostgreSQL: %v", err)
			}
			if len(postgresDocDetailData) > 0 {
				if err := mypg.InsertDocDetailListToPostgreSql(context.Background(), postgresDB, holdingCode, postgresDocDetailData); err != nil {
					logger.Error("Failed to insert final docdetail batch to PostgreSQL: %v", err)
				}
			}
		}()

		// ClickHouse insert (parallel)
		insertWg.Add(1)
		go func() {
			defer insertWg.Done()
			myclickhouse.InsertDocListToClickHouse(context.Background(), holdingCode, postgresDocData, postgresDocRefData, docPaymentDataList)
			if len(postgresDocDetailData) > 0 {
				myclickhouse.InsertDocDetailListToClickHouse(context.Background(), holdingCode, postgresDocDetailData)
			}
		}()

		insertWg.Wait()
	}

	logger.Success("Completed rebuild for %s: %d docs, %d details processed in %d batches", transaction.Name, totalDocsProcessed, totalDetailsProcessed, batchCount)
}

func RebuildDocDetailFromMongoWhereTransFlagsDeleteOldData(mongoClient *mongo.Client, postgresDB *sql.DB, holdingCode string, transaction models.StockTransactionStruct) {
	logger.Debug("Starting deletion of old docdetail data for %s (holdingCode: %s)", transaction.Name, holdingCode)

	// transaction.Flags เป็น []int
	var flagStrings []string
	for _, flag := range transaction.Flags {
		flagStrings = append(flagStrings, strconv.Itoa(flag))
	}
	flagsInClause := strings.Join(flagStrings, ",")

	// ⚡ OPTIMIZATION: ลบ PostgreSQL และ ClickHouse แบบขนาน
	var wg sync.WaitGroup
	var rowsAffected int64

	// PostgreSQL deletion
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Debug("Deleting old records from PostgreSQL docdetail (flags: %s)", flagsInClause)
		queryDelete := "DELETE FROM docdetail WHERE transflag IN (" + flagsInClause + ");"
		result, err := postgresDB.ExecContext(context.Background(), queryDelete)
		if err != nil {
			logger.Error("Failed to delete from docdetail table: %v", err)
			return
		}
		rowsAffected, _ = result.RowsAffected()
		logger.Debug("Deleted %d records from PostgreSQL docdetail", rowsAffected)
	}()

	// ⚡ ClickHouse deletion - ไม่รอ mutations_sync
	wg.Add(1)
	go func() {
		defer wg.Done()
		clickhouseConn, err := myclickhouse.ClickHouseFastConnect()
		if err != nil {
			logger.Error("Failed to connect to ClickHouse: %v", err)
			return
		}

		// ไม่ใช้ mutations_sync = 1 เพื่อไม่ต้องรอ
		deleteQuery := fmt.Sprintf("ALTER TABLE %s DELETE WHERE holding_code='%s' AND transflag IN (%s)", myclickhouse.TableName("docdetail"), holdingCode, flagsInClause)
		if err := myclickhouse.ExecuteCommand(context.Background(), clickhouseConn, deleteQuery); err != nil {
			logger.Error("Failed to delete from ClickHouse docdetail table: %v", err)
			return
		}
		logger.Debug("ClickHouse docdetail delete mutation submitted (async)")
	}()

	wg.Wait()
	logger.Success("Deleted %d records from PostgreSQL docdetail for %s", rowsAffected, transaction.Name)
}

// DocProgressCallback — callback สำหรับรายงาน progress การนำเข้าเอกสาร
type DocProgressCallback func(current, total int, transactionName string)

func DocRebuildAllFromMongo(holdingCode string, rebuild bool) {
	DocRebuildAllFromMongoWithCallback(holdingCode, rebuild, nil)
}

// DocRebuildAllFromMongoWithCallback — นำเข้าเอกสารจาก MongoDB พร้อม progress callback
func DocRebuildAllFromMongoWithCallback(holdingCode string, rebuild bool, callback DocProgressCallback) {
	logger.Info("Starting complete document + docdetail rebuild for holdingCode: %s (rebuild mode: %v)", holdingCode, rebuild)
	timeStart := time.Now()

	mongoClient := myglobal.SafeMongoConnectFast() // Use optimized connection
	if mongoClient == nil {
		logger.Error("Failed to connect to MongoDB")
		return
	}
	logger.Debug("MongoDB connection established successfully")

	postgresDB, err := mypg.PgSqlFastConnect(holdingCode)
	if err != nil {
		logger.Error("Failed to connect to PostgreSQL for %s: %v", holdingCode, err)
		return
	}
	logger.Debug("PostgreSQL connection established successfully")

	if rebuild {
		// ลบข้อมูลเก่าใน clickhouse และ postgresql ทั้งหมด
		logger.Info("Deleting all old documents and docdetails from PostgreSQL and ClickHouse for holdingCode %s", holdingCode)

		clickhouseConn, err := myclickhouse.ClickHouseFastConnect()
		if err != nil {
			logger.Error("Failed to connect to ClickHouse: %v", err)
			return
		}

		clickHouseQuery := fmt.Sprintf("ALTER TABLE %s DELETE WHERE holding_code = '%s'", myclickhouse.TableName("processstockcost"), holdingCode)
		err = myclickhouse.ExecuteCommand(context.Background(), clickhouseConn, clickHouseQuery)
		if err != nil {
			logger.Error("Failed to delete processstockcost from ClickHouse: %v", err)
		} else {
			logger.Debug("Deleted processstockcost records from ClickHouse (mutation will complete asynchronously)")
		}

		// นับจำนวน transaction types ที่ต้องประมวลผล (ยกเว้น TransFlag 54)
		transactionList := myglobal.StockTransactionList()
		totalTransTypes := 0
		for _, t := range transactionList {
			if t.Flags[0] != 54 {
				totalTransTypes++
			}
		}

		totalTransactions := 0
		for _, transaction := range transactionList {
			// ข้าม TransFlag 54 เพราะมีโครงสร้างต่างจากปกติ
			if transaction.Flags[0] == 54 {
				logger.Info("Skipping TransFlag 54 (handled separately)")
				continue
			}

			totalTransactions++

			// รายงาน progress ผ่าน callback
			if callback != nil {
				callback(totalTransactions, totalTransTypes, transaction.Name)
			}

			logger.Info("Processing transaction %s for shop %s (Doc + DocDetail)", transaction.Name, holdingCode)
			// Function นี้จะสร้างทั้ง Doc และ DocDetail พร้อมกัน
			RebuildDocFromMongoWhereTransFlags(mongoClient, postgresDB, holdingCode, transaction)
			logger.Success("Completed transaction %s for shop %s", transaction.Name, holdingCode)
		}
		logger.Debug("Processed %d transaction types total", totalTransactions)
	}

	// ⚡ OPTIMIZATION: รวม INSERT ทั้ง 2 ตาราง ใน transaction เดียว เพื่อความเร็ว
	logger.Debug("Inserting processed documents and items into queue tables (OPTIMIZED)")

	// เริ่ม transaction
	tx, err := postgresDB.BeginTx(context.Background(), nil)
	if err != nil {
		logger.Error("Failed to begin transaction: %v", err)
	} else {
		// Insert into docwaitprocess (สำหรับ Doc)
		queryDoc := `
			INSERT INTO docwaitprocess (docno,transflag)
			SELECT DISTINCT docno,transflag
			FROM doc
		`
		resultDoc, errDoc := tx.ExecContext(context.Background(), queryDoc)
		if errDoc != nil {
			logger.Error("Failed to insert into docwaitprocess table: %v", errDoc)
			tx.Rollback()
		} else {
			rowsDoc, _ := resultDoc.RowsAffected()

			// Insert into stockwaitprocess (สำหรับ DocDetail)
			queryStock := `
				INSERT INTO stockwaitprocess (itemcode)
				SELECT DISTINCT itemcode
				FROM docdetail
				WHERE itemcode <> '' AND itemcode IS NOT NULL
				GROUP BY itemcode
			`
			resultStock, errStock := tx.ExecContext(context.Background(), queryStock)
			if errStock != nil {
				logger.Error("Failed to insert into stockwaitprocess table: %v", errStock)
				tx.Rollback()
			} else {
				rowsStock, _ := resultStock.RowsAffected()

				// Commit transaction
				if errCommit := tx.Commit(); errCommit != nil {
					logger.Error("Failed to commit transaction: %v", errCommit)
				} else {
					logger.Success("✅ Inserted %d documents + %d items into queue tables (in 1 transaction)", rowsDoc, rowsStock)
				}
			}
		}
	}

	elapsed := time.Since(timeStart)
	logger.Success("Document + DocDetail rebuild completed for holdingCode %s in %v", holdingCode, elapsed)
}

// ==================== TransFlag 54 Functions (Stock Balance - โครงสร้าง JSON ต่างจากปกติ) ====================

func createDocDetailTransFlag54(docDetailData models.ProcessMongoTransDetailTransFlag54Model) models.DocDetailStruct {
	if docDetailData.Qty == 0 {
		logger.Warn("totalQty is zero for docNo %s, lineNumber %v", docDetailData.DocNo, docDetailData.LineNumber)
	}
	var itemCode = strings.TrimSpace(docDetailData.ItemCode)
	if itemCode == "" {
		itemCode = docDetailData.Barcode
		if itemCode == "" {
			itemCode = "UNKNOWN"
		}
	}
	var itemName string = ""
	if len(docDetailData.ItemNames) > 0 {
		itemName = docDetailData.ItemNames[0].Name
	} else {
		itemName = "UNKNOWN"
	}

	var whCode = strings.TrimSpace(docDetailData.WhCode)
	var locationCode = strings.TrimSpace(docDetailData.LocationCode)

	if whCode == "" {
		whCode = "X"
	}
	if locationCode == "" {
		locationCode = "X"
	}

	return models.DocDetailStruct{
		DocDateTime:     docDetailData.DocDateTime,
		DocNo:           docDetailData.DocNo,
		LineNumber:      docDetailData.LineNumber,
		TransFlag:       54,
		CalcFlag:        1,
		CalcSeq:         1,
		ItemCode:        docDetailData.ItemCode,
		Description:     itemName,
		BarcodeMain:     "",
		Barcode:         docDetailData.Barcode,
		UnitCode:        docDetailData.UnitCode,
		WhCode:          whCode,
		LocationCode:    locationCode,
		TotalQty:        docDetailData.Qty,
		Price:           docDetailData.Price,
		PriceExcludeVat: docDetailData.PriceExcludeVat,
		UnitStand:       1.0,
		UnitDivide:      1.0,
		DocRef:          docDetailData.DocRef,
		SumAmount:       docDetailData.SumAmount,
	}
}

func DocDetailTransFlag54Decode(jsonData string) models.ProcessMongoTransDetailTransFlag54Model {
	jsonDecode := map[string]any{}
	err := json.Unmarshal([]byte(jsonData), &jsonDecode)
	if err != nil {
		logger.Error("Failed to unmarshal TransFlag54 JSON: %v", err)
		return models.ProcessMongoTransDetailTransFlag54Model{}
	}

	fixedData := myglobal.FixMongoData(jsonDecode)

	var data models.ProcessMongoTransDetailTransFlag54Model
	decoderConfig := mapstructure.DecoderConfig{
		DecodeHook: myglobal.CustomDecodeHookFunc,
		Result:     &data,
		TagName:    "json",
	}
	decoder, err := mapstructure.NewDecoder(&decoderConfig)
	if err != nil {
		logger.Fatal("Failed to create TransFlag54 decoder: %v", err)
	}

	err = decoder.Decode(fixedData)
	if err != nil {
		logger.Error("Failed to decode TransFlag54 data: %v", err)
	}
	return data
}

func DocDetailTransFlag54RebuildFromMongo(mongoClient *mongo.Client, postgresDB *sql.DB, holdingCode string) {
	logger.Info("** Starting DocDetail rebuild for transflag 54 (%s)", holdingCode)

	// ลบข้อมูลเก่าใน PostgreSQL
	queryDelete := "DELETE FROM docdetail WHERE transflag=54;"
	_, err := postgresDB.ExecContext(context.Background(), queryDelete)
	if err != nil {
		logger.Error("Failed to delete from docdetail table: %v", err)
		return
	}
	logger.Debug("Deleted existing records from PostgreSQL docdetail for transflag 54")

	// ลบข้อมูลเก่าใน ClickHouse docdetail
	clickhouseConn, err := myclickhouse.ClickHouseFastConnect()
	if err != nil {
		logger.Error("Failed to connect to ClickHouse: %v", err)
		return
	}
	// ใช้ connection pool ไม่ต้อง close

	// ใช้ SETTINGS mutations_sync = 1 เพื่อรอให้ลบเสร็จก่อน insert
	deleteQuery := fmt.Sprintf("ALTER TABLE %s DELETE WHERE holding_code='%s' AND transflag=54 SETTINGS mutations_sync = 1", myclickhouse.TableName("docdetail"), holdingCode)
	err = myclickhouse.ExecuteCommand(context.Background(), clickhouseConn, deleteQuery)
	if err != nil {
		logger.Error("Failed to delete from ClickHouse docdetail: %v", err)
		return
	}
	logger.Debug("Deleted existing records from ClickHouse docdetail for transflag 54")

	// ดึงข้อมูลจาก MongoDB
	svcConfig := config.NewServiceConfig()
	MongodbDatabaseName := svcConfig.MongodbDatabaseName()
	collection := mongoClient.Database(MongodbDatabaseName).Collection("transactionstockbalancedetails")

	cur, err := collection.Find(context.Background(), bson.M{"holdingcode": holdingCode, "deletedat": bson.M{"$exists": false}})
	if err != nil {
		logger.Error("Failed to find documents: %v", err)
		return
	}
	defer cur.Close(context.Background())

	var docDetailList []models.DocDetailStruct
	totalProcessed := 0
	batchCount := 0

	timeStart := time.Now()
	for cur.Next(context.Background()) {
		docDetailData := DocDetailTransFlag54Decode(cur.Current.String())
		detailGenerated := createDocDetailTransFlag54(docDetailData)
		docDetailList = append(docDetailList, detailGenerated)
		totalProcessed++

		// Batch insert when reaching 10000 records
		if len(docDetailList) >= 10000 {
			batchCount++
			logger.Info("TransFlag 54 Inserting batch #%d (%d records) for holdingCode %s", batchCount, len(docDetailList), holdingCode)

			// Insert to PostgreSQL
			if err := mypg.InsertDocDetailListToPostgreSql(context.Background(), postgresDB, holdingCode, docDetailList); err != nil {
				logger.Error("Failed to insert batch to PostgreSQL: %v", err)
			} // Insert to ClickHouse
			myclickhouse.InsertDocDetailListToClickHouse(context.Background(), holdingCode, docDetailList)

			docDetailList = docDetailList[:0]
		}
	}

	// Insert remaining records
	if len(docDetailList) > 0 {
		batchCount++
		logger.Info("TransFlag 54 Inserting final batch #%d (%d records) for holdingCode %s", batchCount, len(docDetailList), holdingCode)

		// Insert to PostgreSQL
		if err := mypg.InsertDocDetailListToPostgreSql(context.Background(), postgresDB, holdingCode, docDetailList); err != nil {
			logger.Error("Failed to insert final batch to PostgreSQL: %v", err)
		}

		// Insert to ClickHouse
		myclickhouse.InsertDocDetailListToClickHouse(context.Background(), holdingCode, docDetailList)
	}

	elapsed := time.Since(timeStart)
	logger.Success("Completed TransFlag 54 rebuild: %d records processed in %d batches (%v)", totalProcessed, batchCount, elapsed)
}
