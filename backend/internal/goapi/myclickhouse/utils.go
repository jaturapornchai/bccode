package myclickhouse

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/models"
	"smlcloudplatform/internal/goapi/myglobal"
	"strings"
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/google/uuid"
)

// Global ClickHouse connection pool
var (
	clickhouseConn  clickhouse.Conn
	clickhouseOnce  sync.Once
	clickhouseMutex sync.RWMutex
)

// GetDatabaseName คืนชื่อ database ClickHouse จาก env variable CH_DATABASE_NAME
// ถ้าไม่มีจะ default เป็น "bcbidev"
func GetDatabaseName() string {
	dbName := os.Getenv("CH_DATABASE_NAME")
	if dbName == "" {
		return "bcbidev"
	}
	return dbName
}

// TableName คืนชื่อ table แบบ fully-qualified (database.table) เช่น "bcbi.doc"
func TableName(table string) string {
	return fmt.Sprintf("%s.%s", GetDatabaseName(), table)
}

func QuerySelectAll(conn clickhouse.Conn, query string) ([]map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	rows, err := conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	dataRows := make([]map[string]interface{}, 0)
	columnTypes := rows.ColumnTypes()

	vars := make([]interface{}, len(columnTypes))
	for i := range columnTypes {
		scanType := columnTypes[i].ScanType()
		columnName := columnTypes[i].Name()

		// จัดการ transflag เป็น UInt16 โดยเฉพาะ
		if columnName == "transflag" {
			vars[i] = new(uint16)
			continue
		}

		switch scanType.Kind() {
		case reflect.String:
			vars[i] = new(string)
		case reflect.Int8:
			vars[i] = new(int8)
		case reflect.Int16:
			vars[i] = new(int16)
		case reflect.Int32:
			vars[i] = new(int32)
		case reflect.Int, reflect.Int64:
			vars[i] = new(int64)
		case reflect.Uint8:
			vars[i] = new(uint8)
		case reflect.Uint16:
			vars[i] = new(uint16)
		case reflect.Uint32:
			vars[i] = new(uint32)
		case reflect.Uint, reflect.Uint64:
			vars[i] = new(uint64)
		case reflect.Float32, reflect.Float64:
			vars[i] = new(float64)
		case reflect.Bool:
			vars[i] = new(bool)
		case reflect.TypeOf(time.Time{}).Kind():
			vars[i] = new(time.Time)
		default:
			logger.Error("Unsupported column type %v (scanType=%v) for column %s", scanType.Kind(), scanType, columnName)
			return nil, fmt.Errorf("unsupported type %v for column %s", scanType.Kind(), columnName)
		}
	}

	for rows.Next() {
		if err := rows.Scan(vars...); err != nil {
			logger.Error("Failed to scan row: %v", err)
			return nil, err
		}

		rowData := make(map[string]interface{}, len(columnTypes))
		for i, v := range vars {
			columnName := columnTypes[i].Name()
			switch val := reflect.Indirect(reflect.ValueOf(v)).Interface().(type) {
			case string:
				rowData[columnName] = val
			case int8:
				rowData[columnName] = int(val)
			case int16:
				rowData[columnName] = int(val)
			case int32:
				rowData[columnName] = int(val)
			case int64:
				rowData[columnName] = int(val)
			case int:
				rowData[columnName] = val
			case uint8:
				rowData[columnName] = int(val)
			case uint16:
				rowData[columnName] = int(val)
			case uint32:
				rowData[columnName] = int(val)
			case uint64:
				rowData[columnName] = int(val)
			case float64:
				rowData[columnName] = val
			case bool:
				rowData[columnName] = val
			case time.Time:
				rowData[columnName] = val
			}
		}
		dataRows = append(dataRows, rowData)
	}

	return dataRows, nil
}

// QuerySelectAllWithUTF8 - รองรับ UTF-8 เต็มรูปแบบ
func QuerySelectAllWithUTF8(conn clickhouse.Conn, query string) ([]map[string]interface{}, error) {

	// เพิ่ม timeout และตั้งค่า UTF-8
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	rows, err := conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	dataRows := make([]map[string]interface{}, 0)
	columnTypes := rows.ColumnTypes()

	vars := make([]interface{}, len(columnTypes))
	for i := range columnTypes {
		scanType := columnTypes[i].ScanType()
		columnName := columnTypes[i].Name()

		if columnName == "transflag" {
			vars[i] = new(uint16)
			continue
		}

		switch scanType.Kind() {
		case reflect.String:
			vars[i] = new(string)
		case reflect.Int8:
			vars[i] = new(int8)
		case reflect.Int16:
			vars[i] = new(int16)
		case reflect.Int32:
			vars[i] = new(int32)
		case reflect.Int, reflect.Int64:
			vars[i] = new(int64)
		case reflect.Uint8:
			vars[i] = new(uint8)
		case reflect.Uint16:
			vars[i] = new(uint16)
		case reflect.Uint32:
			vars[i] = new(uint32)
		case reflect.Uint, reflect.Uint64:
			vars[i] = new(uint64)
		case reflect.Float32, reflect.Float64:
			vars[i] = new(float64)
		case reflect.Bool:
			vars[i] = new(bool)
		case reflect.TypeOf(time.Time{}).Kind():
			vars[i] = new(time.Time)
		default:
			return nil, fmt.Errorf("unsupported type %v for column %s", scanType.Kind(), columnName)
		}
	}

	for rows.Next() {
		if err := rows.Scan(vars...); err != nil {
			return nil, err
		}

		rowData := make(map[string]interface{}, len(columnTypes))
		for i, v := range vars {
			columnName := columnTypes[i].Name()
			switch val := reflect.Indirect(reflect.ValueOf(v)).Interface().(type) {
			case string:
				rowData[columnName] = val
			case int8:
				rowData[columnName] = int(val)
			case int16:
				rowData[columnName] = int(val)
			case int32:
				rowData[columnName] = int(val)
			case int64:
				rowData[columnName] = int(val)
			case int:
				rowData[columnName] = val
			case uint8:
				rowData[columnName] = int(val)
			case uint16:
				rowData[columnName] = int(val)
			case uint32:
				rowData[columnName] = int(val)
			case uint64:
				rowData[columnName] = int(val)
			case float64:
				rowData[columnName] = val
			case bool:
				rowData[columnName] = val
			case time.Time:
				rowData[columnName] = val
			}
		}
		dataRows = append(dataRows, rowData)
	}

	return dataRows, nil
}

func ClickHouseFastConnect() (clickhouse.Conn, error) {
	return nil, fmt.Errorf("clickhouse is disabled")
}

// getClickHouseOptions คืน Options สำหรับเชื่อมต่อ ClickHouse
func getClickHouseOptions(database string) *clickhouse.Options {
	return &clickhouse.Options{
		Addr: []string{os.Getenv("CH_SERVER_ADDRESS")},
		Auth: clickhouse.Auth{
			Database: database,
			Username: os.Getenv("CH_USERNAME"),
			Password: os.Getenv("CH_PASSWORD"),
		},
		Settings: clickhouse.Settings{
			"max_execution_time":            600,
			"max_block_size":                100000,
			"max_insert_block_size":         1048576,
			"max_memory_usage":              10000000000,
			"insert_quorum":                 0,
			"insert_quorum_timeout":         600000,
			"select_sequential_consistency": 0,
			"max_threads":                   16,
			"receive_timeout":               600,
			"send_timeout":                  600,
			"enable_http_compression":       1,
			"distributed_product_mode":      "global",
		},
		DialTimeout:          30 * time.Second,
		MaxOpenConns:         25,
		MaxIdleConns:         10,
		ConnMaxLifetime:      2 * time.Hour,
		ConnOpenStrategy:     clickhouse.ConnOpenInOrder,
		BlockBufferSize:      10,
		MaxCompressionBuffer: 10240,
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
	}
}

// ensureDatabase สร้าง database ใน ClickHouse ถ้ายังไม่มี
func ensureDatabase(dbName string) error {
	return fmt.Errorf("clickhouse is disabled")
}

// สร้าง ClickHouse connection ใหม่ — ถ้า database ยังไม่มีจะสร้าง database + tables อัตโนมัติ
func CreateClickHouseConnection() (clickhouse.Conn, error) {
	return nil, fmt.Errorf("clickhouse is disabled")
}

// CloseClickHouseConnection ปิด ClickHouse connection pool
func CloseClickHouseConnection() error {
	return nil
}

func ExecuteCommand(ctx context.Context, conn clickhouse.Conn, query string) error {
	return conn.Exec(ctx, query)
}

// ======================= Async Batch Insert =======================

// AsyncBatchItem ข้อมูลที่รอ insert แบบ async
type AsyncBatchItem struct {
	Table       string
	HoldingCode string
	Columns     []string
	Values      [][]interface{}
	Callback    func(error)
	CreatedAt   time.Time
}

var (
	asyncBatchQueue   = make(chan AsyncBatchItem, 1000) // Queue สำหรับรวบรวม batch
	asyncWorkerOnce   sync.Once
	asyncWorkerCancel context.CancelFunc
	asyncWorkerWg     sync.WaitGroup
)

// initAsyncWorkerPool เริ่มต้น worker pool สำหรับ async insert
func initAsyncWorkerPool() {
	asyncWorkerOnce.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		asyncWorkerCancel = cancel
		for i := 0; i < 2; i++ {
			asyncWorkerWg.Add(1)
			go asyncBatchWorker(ctx, i+1)
		}
	})
}

// closeAsyncWorkerPool ปิด worker pool
func closeAsyncWorkerPool() {
	if asyncWorkerCancel != nil {
		asyncWorkerCancel()
		asyncWorkerWg.Wait()
	}
}

// asyncBatchWorker worker สำหรับประมวลผล batch insert แบบ async
func asyncBatchWorker(ctx context.Context, workerID int) {
	defer asyncWorkerWg.Done()

	// รวบรวม batch จาก queue
	batchBuffer := make(map[string][]AsyncBatchItem) // key = table
	flushTicker := time.NewTicker(500 * time.Millisecond)
	defer flushTicker.Stop()

	const maxBatchSize = 5000 // รวม batch ไม่เกิน 5000 rows

	for {
		select {
		case <-ctx.Done():
			// Flush remaining batches before exit
			for table, items := range batchBuffer {
				if len(items) > 0 {
					flushBatch(items, table, workerID)
				}
			}
			return

		case item := <-asyncBatchQueue:
			// รวม batch ตาม table
			key := item.Table
			batchBuffer[key] = append(batchBuffer[key], item)

			// Flush ถ้า batch ใหญ่พอ
			totalRows := 0
			for _, bi := range batchBuffer[key] {
				totalRows += len(bi.Values)
			}
			if totalRows >= maxBatchSize {
				flushBatch(batchBuffer[key], key, workerID)
				batchBuffer[key] = nil
			}

		case <-flushTicker.C:
			// Flush batch ที่รออยู่ทุก 500ms
			for table, items := range batchBuffer {
				if len(items) > 0 {
					flushBatch(items, table, workerID)
					batchBuffer[table] = nil
				}
			}
		}
	}
}

// flushBatch insert batch ไปยัง ClickHouse
func flushBatch(items []AsyncBatchItem, table string, workerID int) {
	if len(items) == 0 {
		return
	}

	conn, err := ClickHouseFastConnect()
	if err != nil {
		for _, item := range items {
			if item.Callback != nil {
				item.Callback(err)
			}
		}
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	columns := items[0].Columns
	var allValues [][]interface{}
	for _, item := range items {
		allValues = append(allValues, item.Values...)
	}

	insertQuery := fmt.Sprintf("INSERT INTO %s (%s)", table, strings.Join(columns, ", "))
	batch, err := conn.PrepareBatch(ctx, insertQuery)
	if err != nil {
		for _, item := range items {
			if item.Callback != nil {
				item.Callback(err)
			}
		}
		return
	}

	for _, values := range allValues {
		batch.Append(values...)
	}

	if err := batch.Send(); err != nil {
		for _, item := range items {
			if item.Callback != nil {
				item.Callback(err)
			}
		}
		return
	}

	for _, item := range items {
		if item.Callback != nil {
			item.Callback(nil)
		}
	}
}

// AsyncInsert ส่งข้อมูลไปยัง async queue สำหรับ batch insert
// callback จะถูกเรียกเมื่อ insert เสร็จ (หรือ error)
func AsyncInsert(table, holdingCode string, columns []string, values [][]interface{}, callback func(error)) {
	// เริ่ม worker pool ถ้ายังไม่ได้เริ่ม
	initAsyncWorkerPool()

	item := AsyncBatchItem{
		Table:       table,
		HoldingCode: holdingCode,
		Columns:     columns,
		Values:      values,
		Callback:    callback,
		CreatedAt:   time.Now(),
	}

	select {
	case asyncBatchQueue <- item:
	default:
		go func() {
			conn, err := ClickHouseFastConnect()
			if err != nil {
				if callback != nil {
					callback(err)
				}
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
			defer cancel()

			insertQuery := fmt.Sprintf("INSERT INTO %s (%s)", table, strings.Join(columns, ", "))
			batch, err := conn.PrepareBatch(ctx, insertQuery)
			if err != nil {
				if callback != nil {
					callback(err)
				}
				return
			}

			for _, v := range values {
				batch.Append(v...)
			}

			err = batch.Send()
			if callback != nil {
				callback(err)
			}
		}()
	}
}

// AsyncInsertFireAndForget ส่งข้อมูลไปยัง async queue โดยไม่รอผลลัพธ์
func AsyncInsertFireAndForget(table, holdingCode string, columns []string, values [][]interface{}) {
	AsyncInsert(table, holdingCode, columns, values, nil)
}

func DocDeleteClickHouse(ctx context.Context, holdingCode string, docNo string) {
	conn, err := ClickHouseFastConnect()
	if err != nil {
		return
	}

	deleteCommands := []string{
		fmt.Sprintf("ALTER TABLE %s DELETE WHERE holdingcode = '%s' AND docno = '%s'", TableName("doc"), holdingCode, docNo),
		fmt.Sprintf("ALTER TABLE %s DELETE WHERE holdingcode = '%s' AND docno = '%s'", TableName("docdetail"), holdingCode, docNo),
		fmt.Sprintf("ALTER TABLE %s DELETE WHERE holdingcode = '%s' AND docno = '%s'", TableName("docpayment"), holdingCode, docNo),
	}

	for _, cmd := range deleteCommands {
		ExecuteCommand(ctx, conn, cmd)
	}

	time.Sleep(500 * time.Millisecond)
}

// DocUpdate ทำหน้าที่อัพเดทข้อมูลเอกสารในฐานข้อมูล ClickHouse
func DocUpdate(docData models.MongoDocModel) {
	go func(docData models.MongoDocModel) {
		clickHouseConn, err := ClickHouseFastConnect()
		if err != nil {
			return
		}

		ctx := context.Background()
		checkSumMongodb := myglobal.CalculateMD5(fmt.Sprintf("%v", docData))
		if len(checkSumMongodb) != 32 {
			checkSumMongodb = uuid.NewString()
		}

		query := fmt.Sprintf("SELECT checksum FROM %s WHERE holdingcode = '%s' AND docno = '%s'", TableName("doc"), docData.HoldingCode, docData.DocNo)
		dataRows, err := QuerySelectAll(clickHouseConn, query)
		if err != nil {
			return
		}

		if len(dataRows) > 0 {
			existingChecksum := dataRows[0]["checksum"].(string)
			if existingChecksum == checkSumMongodb {
				return
			}
		}

		DocDeleteClickHouse(context.Background(), docData.HoldingCode, docData.DocNo)

		docDateTimeStr := docData.DocDateTime.Format("2006-01-02 15:04:05")
		payCashBalance := docData.PayCashAmount - docData.PayCashChange

		insertCommands := []string{
			fmt.Sprintf(`INSERT INTO %s (
            holdingcode, branchid, docno, docdatetime, perioddatetime,
            totalamount, paycashamount, paycashchange, paycashbalance,
            roundamount, checksum, slipurl, salechannelcode, deliveryamount,
            guidfixed, iscancel, cancelreason, guidpos, guidbranch
        ) VALUES (
            '%s', '%s', '%s', '%s', '%s', %.2f, %.2f, %.2f, %.2f, %.2f,
            '%s', '%s', '%s', %.2f, '%s', %v, '%s', '%s', '%s'
        )`, TableName("doc"),
				docData.HoldingCode, docData.BranchId, docData.DocNo, docDateTimeStr, docDateTimeStr,
				docData.TotalAmount, docData.PayCashAmount, docData.PayCashChange, payCashBalance,
				docData.RoundAmount, checkSumMongodb, docData.SlipUrl, docData.SaleChannelCode,
				docData.DeliveryAmount, docData.GuidFixed, docData.IsCancel,
				strings.Replace(docData.CancelReason, "'", "''", -1),
				docData.GuidPos, docData.Branch.GuidFixed),
		}

		for i, detail := range docData.Details {
			itemName := ""
			if len(detail.ItemNames) > 0 {
				itemName = strings.Replace(detail.ItemNames[0].Name, "'", "''", -1)
			}
			insertCommands = append(insertCommands,
				fmt.Sprintf(`INSERT INTO %s (
                holdingcode, branchid, docno, docdatetime, perioddatetime,
                line_number, barcode, qty, price, sumamount, discountamount,
                itemnames, refguid, sumamountchoice, ischoice, guidfixed, guidpos, guidbranch
            ) VALUES (
                '%s', '%s', '%s', '%s', '%s', %d, '%s', %.2f, %.2f, %.2f, %.2f,
                '%s', '%s', %.2f, %d, '%s', '%s', '%s'
            )`, TableName("docdetail"),
					docData.HoldingCode, docData.BranchId, docData.DocNo, docDateTimeStr, docDateTimeStr,
					i+1, detail.Barcode, detail.Qty, detail.Price, detail.SumAmount, detail.DiscountAmount,
					itemName, detail.RefGuid, detail.SumAmountChoice, detail.IsChoice, docData.GuidFixed,
					docData.GuidPos, docData.Branch.GuidFixed),
			)
		}

		jsonPaymentRaw := strings.Replace(docData.PaymentDetailRaw, "'", "''", -1)
		var payments []map[string]interface{}
		if err := json.Unmarshal([]byte(jsonPaymentRaw), &payments); err == nil {
			for _, payment := range payments {
				amount, _ := payment["amount"].(float64)
				providerName := payment["provider_name"].(string)
				trans_flag := payment["trans_flag"].(float64)
				insertCommands = append(insertCommands,
					fmt.Sprintf(`INSERT INTO %s (
                    holdingcode, branchid, docno, docdatetime, perioddatetime,
                    description, amount, trans_flag, guidfixed, guidbranch
                ) VALUES (
                    '%s', '%s', '%s', '%s', '%s', '%s', %f, %f, '%s', '%s'
                )`, TableName("docpayment"),
						docData.HoldingCode, docData.BranchId, docData.DocNo, docDateTimeStr, docDateTimeStr,
						providerName, amount, trans_flag, docData.GuidFixed, docData.Branch.GuidFixed))
			}
		}

		for _, cmd := range insertCommands {
			ExecuteCommand(ctx, clickHouseConn, cmd)
		}
	}(docData)
}

func ProductBarcodeUpdate(productData models.MongoProductBarcodeModel) error {
	conn, err := ClickHouseFastConnect()
	if err != nil {
		return err
	}
	ctx := context.Background()

	productDataStr := fmt.Sprintf("%v", productData)
	checkSumMongodb := myglobal.CalculateMD5(productDataStr)
	updateClickHouse := false

	query := fmt.Sprintf("SELECT checksum FROM %s WHERE holdingcode = '%s' AND barcode = '%s'", TableName("productbarcode"), productData.HoldingCode, productData.Barcode)
	dataRows, err := QuerySelectAll(conn, query)
	if err != nil {
		return err
	}

	if len(dataRows) > 0 {
		if dataRows[0]["checksum"] != checkSumMongodb {
			updateClickHouse = true
		}
	} else {
		updateClickHouse = true
	}

	if updateClickHouse {

		// สร้างคำสั่ง SQL หลายประเภทที่ต้องการรันพร้อมกัน
		deleteCommands := []string{
			fmt.Sprintf("ALTER TABLE %s DELETE WHERE holdingcode = '%s' AND barcode = '%s'", TableName("productbarcode"), productData.HoldingCode, productData.Barcode),
		}

		productName := ""
		productName1 := ""
		productName2 := ""
		productName3 := ""
		productName4 := ""
		productName5 := ""
		productGroupName := ""
		productUnitName := ""
		price := 0.0

		if len(productData.Names) > 0 {
			productName = productData.Names[0].Name
		}
		if len(productData.Names) > 1 {
			productName1 = productData.Names[1].Name
		}
		if len(productData.Names) > 2 {
			productName2 = productData.Names[2].Name
		}
		if len(productData.Names) > 3 {
			productName3 = productData.Names[3].Name
		}
		if len(productData.Names) > 4 {
			productName4 = productData.Names[4].Name
		}
		if len(productData.Names) > 5 {
			productName5 = productData.Names[5].Name
		}

		if len(productData.GroupNames) > 0 {
			productGroupName = productData.GroupNames[0].Name
		}

		if len(productData.ItemUnitNames) > 0 {
			productUnitName = productData.ItemUnitNames[0].Name
		}

		if len(productData.Prices) > 0 {
			price = productData.Prices[0].Price
		}

		// ดึง barcoderef และ unit conversion จาก RefBarCodes array
		barcodeRef := ""
		unitStand := 1.0
		unitDivide := 1.0
		if len(productData.RefBarCodes) > 0 {
			barcodeRef = productData.RefBarCodes[0].Barcode
			unitStand = productData.RefBarCodes[0].UnitStand
			unitDivide = productData.RefBarCodes[0].UnitDivide
		}

		insertCommands := []string{
			fmt.Sprintf("INSERT INTO %s (holdingcode, itemcode, barcode, barcoderef, name0, name1, name2, name3, name4, name5, checksum, groupcode, groupnames, unitcode, unitname, price1, unitstand, unitdivide) VALUES ('%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', %.2f, %.8f, %.8f)", TableName("productbarcode"),
				productData.HoldingCode,
				productData.ItemCode,
				productData.Barcode,
				barcodeRef,
				productName,
				productName1,
				productName2,
				productName3,
				productName4,
				productName5,
				checkSumMongodb,
				productData.GroupCode,
				productGroupName,
				productData.ItemUnitCode,
				productUnitName,
				price,
				unitStand,
				unitDivide),
		}

		var wg sync.WaitGroup
		var execErr error
		var mu sync.Mutex

		for _, cmd := range deleteCommands {
			wg.Add(1)
			go func(cmd string) {
				defer wg.Done()
				if err := ExecuteCommand(ctx, conn, cmd); err != nil {
					mu.Lock()
					execErr = err
					mu.Unlock()
				}
			}(cmd)
		}
		wg.Wait()
		if execErr != nil {
			return execErr
		}

		for _, cmd := range insertCommands {
			wg.Add(1)
			go func(cmd string) {
				defer wg.Done()
				if err := ExecuteCommand(ctx, conn, cmd); err != nil {
					mu.Lock()
					execErr = err
					mu.Unlock()
				}
			}(cmd)
		}
		wg.Wait()
		if execErr != nil {
			return execErr
		}
	}
	return nil
}
