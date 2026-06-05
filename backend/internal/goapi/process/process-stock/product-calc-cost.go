package processstock

import (
	"context"
	"smlcloudplatform/internal/goapi/logger"
	"sync"
	"sync/atomic"
	"time"

	"smlcloudplatform/internal/goapi/models"
	"smlcloudplatform/internal/goapi/myclickhouse"
	mypg "smlcloudplatform/internal/goapi/mypg"
	"smlcloudplatform/internal/goapi/process"

	"github.com/lib/pq"
)

// BatchStockSaver implements StockSaver interface for buffered writing
type BatchStockSaver struct {
	CostChan       chan []models.ProcessStockCostDetailStruct
	LotChan        chan []models.ProcessStockLotStruct
	ClickHouseChan chan []models.ProcessStockCostDetailStruct
}

func (b *BatchStockSaver) SaveCost(details []models.ProcessStockCostDetailStruct) error {
	if len(details) > 0 {
		b.CostChan <- details
	}
	return nil
}

func (b *BatchStockSaver) SaveLot(lots []models.ProcessStockLotStruct) error {
	if len(lots) > 0 {
		b.LotChan <- lots
	}
	return nil
}

func (b *BatchStockSaver) SaveClickHouse(details []models.ProcessStockCostDetailStruct) error {
	if len(details) > 0 {
		b.ClickHouseChan <- details
	}
	return nil
}

// StockProgressCallback — callback สำหรับรายงานความคืบหน้าคำนวณต้นทุนสต็อก
type StockProgressCallback func(processed, total int)

func ProcessStockCostAll(holdingCode string) {
	ProcessStockCostAllWithCallback(holdingCode, nil)
}

// ProcessStockCostAllWithCallback — คำนวณต้นทุนสต็อกทั้งหมด พร้อมรายงาน progress ผ่าน callback
func ProcessStockCostAllWithCallback(holdingCode string, callback StockProgressCallback) {
	logger.Info("ProcessStockCostAll: %s\n", holdingCode)
	pointQty := 2
	pointAmount := 2
	pointCost := 2

	// ดึงรายการสินค้าที่ต้องคำนวณสต็อกจาก stockwaitprocess
	db, err := mypg.PgSqlFastConnect(holdingCode)
	if err != nil {
		logger.Error("connect to PostgreSQL: %v", err)
		return
	}

	query := "SELECT DISTINCT itemcode FROM stockwaitprocess"
	dataRows, err := mypg.QuerySelectAll(db, query)
	if err != nil {
		logger.Error("select PostgreSQL: %v", err)
		return
	}

	logger.Info("Found %d item codes to process for holdingCode %s", len(dataRows), holdingCode)

	// ลบทั้งหมดเฉพาะ holdingCode โดยใช้คำสั่งเดียว
	ctx := context.TODO()

	// Extract item codes for bulk delete
	var itemCodeList []string
	for _, row := range dataRows {
		itemCodeList = append(itemCodeList, mypg.GetStringValue(row, "itemcode"))
	}

	if len(itemCodeList) > 0 {
		// Use lib/pq for ANY array parameter
		_, err = db.ExecContext(ctx, `
			DELETE FROM processstockcost
			WHERE itemcode = ANY($1::text[])
		`, pq.Array(itemCodeList))
		if err != nil {
			logger.Error("delete PostgreSQL processstockcost: %v", err)
			return
		}

		_, err = db.ExecContext(ctx, `
			DELETE FROM processstocklot
			WHERE itemcode = ANY($1::text[])
		`, pq.Array(itemCodeList))
		if err != nil {
			logger.Error("delete PostgreSQL processstocklot: %v", err)
			return
		}
		logger.Info("Deleted data for %d items for holdingCode %s", len(itemCodeList), holdingCode)
	}

	// Channels for Shared Writer
	costChan := make(chan []models.ProcessStockCostDetailStruct, 1000)
	lotChan := make(chan []models.ProcessStockLotStruct, 1000)
	clickHouseChan := make(chan []models.ProcessStockCostDetailStruct, 1000)
	deleteWaitChan := make(chan string, 1000)

	// Writer WaitGroup
	var writerWg sync.WaitGroup

	// --- Shared Writer for Cost ---
	writerWg.Add(1)
	go func() {
		defer writerWg.Done()
		var batch []models.ProcessStockCostDetailStruct
		batchSize := 50000

		flush := func() {
			if len(batch) == 0 {
				return
			}

			// Insert batch
			columns := []string{
				"docdatetime", "docno", "linenumber", "transflag", "itemcode", "barcode",
				"unitcode", "whcode", "locationcode", "totalqty", "price", "unitstand",
				"unitdivide", "averagecost", "calcamount", "balanceamount", "balanceqty",
				"guid", "unitcost", "docref",
			}

			var records [][]any
			for _, detail := range batch {
				record := []any{
					detail.DocDateTime, detail.DocNo, detail.LineNumber, detail.TransFlag,
					detail.ItemCode, detail.Barcode, detail.UnitCode, detail.WhCode,
					detail.LocationCode, detail.Qty, detail.PriceExcludeVat, detail.UnitStand,
					detail.UnitDivide, detail.AverageCost, detail.CalcAmount, detail.BalanceAmount,
					detail.BalanceQty, detail.Guid, detail.UnitCost, detail.DocRef,
				}
				records = append(records, record)
			}

			if err := mypg.BulkInsertWithCopy(context.Background(), db, "processstockcost", columns, records); err != nil {
				logger.Error("BulkInsertWithCopy cost failed: %v", err)
			}
			batch = nil // Clear batch
		}

		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case items, ok := <-costChan:
				if !ok {
					flush() // Final flush
					return
				}
				batch = append(batch, items...)
				if len(batch) >= batchSize {
					flush()
				}
			case <-ticker.C:
				flush()
			}
		}
	}()

	// --- Shared Writer for Lot ---
	writerWg.Add(1)
	go func() {
		defer writerWg.Done()
		var batch []models.ProcessStockLotStruct
		batchSize := 50000

		flush := func() {
			if len(batch) == 0 {
				return
			}

			columns := []string{
				"docdatetime", "lotnumber", "docno", "transflag", "itemcode",
				"unitcode", "whcode", "locationcode", "qty", "price", "unitstand",
				"unitdivide", "cost", "balanceamount", "balanceqty", "guidref",
			}

			var records [][]any
			for _, lot := range batch {
				record := []any{
					lot.DocDateTime, lot.LotNumber, lot.DocNo, lot.TransFlag, lot.ItemCode,
					lot.UnitCode, lot.WhCode, lot.LocationCode, lot.Qty, lot.Price,
					lot.UnitStand, lot.UnitDivide, lot.Cost, lot.BalanceAmount,
					lot.BalanceQty, lot.GuidRef,
				}
				records = append(records, record)
			}

			if err := mypg.BulkInsertWithCopy(context.Background(), db, "processstocklot", columns, records); err != nil {
				logger.Error("BulkInsertWithCopy lot failed: %v", err)
			}
			batch = nil
		}

		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case items, ok := <-lotChan:
				if !ok {
					flush()
					return
				}
				batch = append(batch, items...)
				if len(batch) >= batchSize {
					flush()
				}
			case <-ticker.C:
				flush()
			}
		}
	}()

	// --- Shared Writer for ClickHouse ---
	writerWg.Add(1)
	go func() {
		defer writerWg.Done()

		// Use a semaphore to limit concurrent ClickHouse connections
		maxClickHouseWorkers := 20
		sem := make(chan struct{}, maxClickHouseWorkers)
		var chWg sync.WaitGroup

		for details := range clickHouseChan {
			if len(details) == 0 {
				continue
			}

			// Extract itemCode from the first detail
			itemCode := details[0].ItemCode

			chWg.Add(1)
			sem <- struct{}{} // acquire

			go func(d []models.ProcessStockCostDetailStruct, code string) {
				defer chWg.Done()
				defer func() { <-sem }() // release

				defer func() {
					if r := recover(); r != nil {
						logger.Error("Recovered from panic in ClickHouse writer: %v", r)
					}
				}()

				// Connect
				clickHouseDB, err := myclickhouse.ClickHouseFastConnect()
				if err != nil {
					logger.Error("Failed to connect to ClickHouse: %v", err)
					return
				}

				// Use process package to replace partition
				result := process.ReplaceProcessStockCostPartition(context.Background(), clickHouseDB, holdingCode, code, d)
				if result.Success {
					// logger.Info("ClickHouse partition replaced successfully: %s", result.Message)
				} else {
					logger.Error("Failed to replace ClickHouse partition: %s", result.Error)
				}

			}(details, itemCode)
		}
		chWg.Wait()
	}()

	// --- Batch Delete StockWaitProcess ---
	writerWg.Add(1)
	go func() {
		defer writerWg.Done()
		var items []string
		batchSize := 500

		flush := func() {
			if len(items) == 0 {
				return
			}
			_, err := db.ExecContext(context.Background(), `
				DELETE FROM stockwaitprocess
				WHERE itemcode = ANY($1::text[])
			`, pq.Array(items))
			if err != nil {
				logger.Error("Batch delete stockwaitprocess failed: %v", err)
			}
			items = nil
		}

		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case item, ok := <-deleteWaitChan:
				if !ok {
					flush()
					return
				}
				items = append(items, item)
				if len(items) >= batchSize {
					flush()
				}
			case <-ticker.C:
				flush()
			}
		}
	}()

	// ใช้ sync เพื่อป้องกันการค้างเมื่อข้อมูลเยอะ
	maxWorkers := 100 // เพิ่มจาก 50 เป็น 100
	sem := make(chan struct{}, maxWorkers)
	var wg sync.WaitGroup

	logger.Info("Processing %d item codes with max %d concurrent workers for shop %s",
		len(dataRows), maxWorkers, holdingCode)

	totalItems := len(dataRows)
	saver := &BatchStockSaver{
		CostChan:       costChan,
		LotChan:        lotChan,
		ClickHouseChan: clickHouseChan,
	}

	// Atomic counter สำหรับ track จำนวนที่ประมวลผลเสร็จ
	var processedCount atomic.Int64

	for index := range dataRows {
		row := dataRows[index]
		itemCodeProcess := mypg.GetStringValue(row, "itemcode")

		// Progress Tracking
		if index%100 == 0 {
			logger.Info("Progress: %d/%d (%.1f%%) - Processing item %s",
				index, totalItems,
				float64(index)/float64(totalItems)*100,
				itemCodeProcess)
		}

		wg.Add(1)
		sem <- struct{}{} // acquire semaphore

		// คำนวณต้นทุน
		go func(itemCode string) {
			defer wg.Done()
			defer func() { <-sem }() // release semaphore

			// Use ProductCalcCostWithSaver with our shared saver
			ProductCalcCostWithSaver(db, holdingCode, itemCode, pointQty, pointAmount, pointCost, false, saver)

			// Send to delete queue
			deleteWaitChan <- itemCode

			// อัพเดท progress ผ่าน callback (ทุก 50 รายการ หรือรายการสุดท้าย)
			current := int(processedCount.Add(1))
			if callback != nil && (current%50 == 0 || current == totalItems) {
				callback(current, totalItems)
			}
		}(itemCodeProcess)
	}

	// รอให้ทุก goroutines เสร็จสิ้น
	wg.Wait()

	// Close channels to signal writers to finish
	close(costChan)
	close(lotChan)
	close(clickHouseChan)
	close(deleteWaitChan)

	// Wait for writers to finish flushing
	writerWg.Wait()

	logger.Info("ProcessStockCostAll completed for shop %s", holdingCode)
}
