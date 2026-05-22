package processstock

import (
	"context"
	"database/sql"
	"fmt"
	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/models"
	"smlcloudplatform/internal/goapi/myclickhouse"
	"smlcloudplatform/internal/goapi/myglobal"
	"smlcloudplatform/internal/goapi/mypg"
	"smlcloudplatform/internal/goapi/mypostgres"
	"smlcloudplatform/internal/goapi/process"
	"math"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type StockSaver interface {
	SaveCost(details []models.ProcessStockCostDetailStruct) error
	SaveLot(lots []models.ProcessStockLotStruct) error
	SaveClickHouse(details []models.ProcessStockCostDetailStruct) error
}

// ProductCalcCost - Calculate stock cost for a single item (legacy interface)
func ProductCalcCost(db *sql.DB, shopId string, itemCodeForProcess string, pointQty int, pointAmount int, pointCost int, deleteFrist bool) {
	ProductCalcCostWithOptions(db, shopId, itemCodeForProcess, pointQty, pointAmount, pointCost, deleteFrist, false, false, nil)
}

// ProductCalcCostIncremental - Calculate stock cost with incremental mode
func ProductCalcCostIncremental(db *sql.DB, shopId string, itemCodeForProcess string, pointQty int, pointAmount int, pointCost int, incremental, minimalLog bool) {
	ProductCalcCostWithOptions(db, shopId, itemCodeForProcess, pointQty, pointAmount, pointCost, !minimalLog, incremental, minimalLog, nil)
}

// ProductCalcCostWithSaver - Calculate stock cost with custom saver (legacy interface)
func ProductCalcCostWithSaver(db *sql.DB, shopId string, itemCodeForProcess string, pointQty int, pointAmount int, pointCost int, deleteFrist bool, saver StockSaver) {
	ProductCalcCostWithOptions(db, shopId, itemCodeForProcess, pointQty, pointAmount, pointCost, deleteFrist, false, false, saver)
}

// ProductCalcCostWithOptions - Full-featured stock cost calculation
// Parameters:
//   - db: database connection
//   - shopId: shop identifier
//   - itemCodeForProcess: item code to calculate
//   - pointQty, pointAmount, pointCost: decimal precision
//   - deleteFirst: if true, delete existing data before insert (legacy mode)
//   - incremental: if true, check checksum and skip unchanged items
//   - minimalLog: if true, use UPSERT to reduce WAL logging
//   - saver: optional custom saver interface
func ProductCalcCostWithOptions(db *sql.DB, shopId string, itemCodeForProcess string, pointQty int, pointAmount int, pointCost int, deleteFirst bool, incremental bool, minimalLog bool, saver StockSaver) {
	logger.Info("Starting CalcCost for item code: %s (incremental=%v, minimalLog=%v)", itemCodeForProcess, incremental, minimalLog)

	// ทำการบันทึกข้อมูลเฉพาะสำหรับ item code นี้
	ctx := context.Background()

	// 🔒 Distributed Lock เพื่อป้องกัน race condition (ใช้ PostgreSQL)
	lockKey := fmt.Sprintf("stock:calc:%s:%s", shopId, itemCodeForProcess)
	lock := mypostgres.NewDistributedLock(db, lockKey, 5*time.Minute) // Lock timeout 5 นาที

	// พยายาม acquire lock (retry 10 ครั้ง, รอครั้งละ 500ms)
	err := lock.AcquireWithRetry(ctx, 10, 500*time.Millisecond)
	if err != nil {
		logger.Error("Failed to acquire lock for %s/%s: %v", shopId, itemCodeForProcess, err)
		return
	}
	defer func() {
		if err := lock.Release(ctx); err != nil {
			logger.Error("Failed to release lock for %s/%s: %v", shopId, itemCodeForProcess, err)
		}
	}()

	logger.Info("✅ Lock acquired for %s/%s", shopId, itemCodeForProcess)

	// Incremental mode: Check if item has changed
	var currentChecksum string
	if incremental {
		changed, checksum, err := CheckItemChanged(ctx, db, shopId, itemCodeForProcess)
		if err != nil {
			logger.Warn("Failed to check item change for %s: %v, will recalculate", itemCodeForProcess, err)
		} else if !changed {
			logger.Info("✅ Item %s unchanged (checksum=%s), skipping calculation", itemCodeForProcess, checksum)
			return
		}
		currentChecksum = checksum
		logger.Info("Item %s has changed (checksum=%s), will recalculate", itemCodeForProcess, checksum)
	}

	// ลบข้อมูลเก่าก่อน เพื่อป้องกันการซ้ำซ้อน (skip if minimalLog mode - will use UPSERT)
	if deleteFirst && !minimalLog {
		// วิธีเดิม: ลบเฉพาะ item code ที่กำลังประมวลผลโดยใช้หลาย query
		deleteQueries := []string{
			"DELETE FROM processstockcost WHERE itemcode = $1",
			"DELETE FROM processstocklot WHERE itemcode = $1",
		}
		for _, deleteQuery := range deleteQueries {
			_, err := db.ExecContext(ctx, deleteQuery, itemCodeForProcess)
			if err != nil {
				logger.Error("delete PostgreSQL: %v", err)
				return
			}
		}
	}

	// ดึงข้อมูลรายวันจาก docdetail และคำนวณค่าสต็อกใหม่แบบ streaming
	query := "SELECT docdatetime, docno, linenumber, calcseq, calcflag, transflag, itemcode, barcode, whcode, locationcode, unitcode, totalqty as qty, unitstand, unitdivide, price, priceexcludevat, docref, sumamount FROM docdetail WHERE itemcode = $1 and transflag IN (%s) ORDER BY itemcode, docdatetime, linenumber, calcflag DESC, docno"

	// ตัวแปรสำหรับการคำนวณต้นทุน
	// lotNumber := 0
	averageCost := 0.0
	balanceQty := 0.0
	balanceAmount := 0.0
	lots := []models.ProcessStockLotStruct{}
	detailList := []models.ProcessStockCostDetailStruct{}
	allDetails := []models.ProcessStockCostDetailStruct{} // เก็บทั้งหมดสำหรับส่งไป ClickHouse
	insertCounter := 0
	totalRecordsProcessed := 0

	// ฟังก์ชันช่วยในการเพิ่ม lot
	/*addLotFunc := func(detail processModel.ProcessStockCostDetail) {
		// เพิ่ม lot ใหม่
		if calcType == 1 {
			amount := myGlobal.RoundFloat64(detail.PriceExcludeVat*detail.Qty, pointAmount)
			qty := myGlobal.RoundFloat64(detail.Qty*(detail.UnitStand/detail.UnitDivide), pointQty)
			price := myGlobal.RoundFloat64((detail.PriceExcludeVat/(detail.UnitStand/detail.UnitDivide))*(detail.UnitStand/detail.UnitDivide), pointCost)
			lotNumber++
			lot := processModel.ProcessStockLot{
				DocDateTime:   detail.DocDateTime,
				LotNumber:     fmt.Sprintf("%06d", lotNumber),
				DocNo:         detail.DocNo,
				TransFlag:     detail.TransFlag,
				BarcodeMain:   barCodeProcess,
				Barcode:       detail.Barcode,
				UnitCode:      detail.UnitCode,
				WhCode:        detail.WhCode,
				LocationCode:  detail.LocationCode,
				Qty:           qty,
				Price:         price,
				UnitStand:     detail.UnitStand,
				UnitDivide:    detail.UnitDivide,
				Cost:          price,
				BalanceAmount: amount,
				BalanceQty:    qty,
				GuidRef:       detail.Guid,
			}
			lots = append(lots, lot)
		}
	}

	// ฟังก์ชันช่วยในการประมวลผล lot
	processLotFunc := func(detail processModel.ProcessStockCostDetail) {
		// ตัดยอดคงเหลือจาก lot ที่มี qty มากก่อน
		if calcType == 1 {
			// FIFO ตัดจาก lot ที่เข้าก่อน
			qty := myGlobal.RoundFloat64(detail.Qty*(detail.UnitStand/detail.UnitDivide), pointQty) * -1
			for qty > 0 {
				hasChange := false
				for lotIndex := 0; lotIndex < len(lots); lotIndex++ {
					if lots[lotIndex].Barcode == detail.Barcode && lots[lotIndex].BalanceQty > 0 {
						hasChange = true
						if lots[lotIndex].BalanceQty >= qty {
							lots[lotIndex].BalanceQty -= qty
							lots[lotIndex].BalanceAmount = myGlobal.RoundFloat64(lots[lotIndex].Price*lots[lotIndex].BalanceQty, pointAmount)
							qty = 0
						} else {
							qty -= lots[lotIndex].BalanceQty
							lots[lotIndex].BalanceAmount = 0
							lots[lotIndex].BalanceQty = 0
						}
					}
				}
				if !hasChange {
					break
				}
			}
		}
	} */

	// ฟังก์ชันสำหรับบันทึกข้อมูลเป็นชุด
	insertBatch := func() error {
		// ประมวลเสร็จแล้ว ให้บันทึกข้อมูลเป็นชุด
		if len(detailList) == 0 {
			return nil
		}

		// ถ้ามี saver ให้ใช้ saver แทนการ insert เอง
		if saver != nil {
			if err := saver.SaveCost(detailList); err != nil {
				logger.Error("saver.SaveCost failed: %v", err)
				return err
			}
			return nil
		}

		// แบ่งชุดข้อมูลเป็นกลุ่มๆ ละ 1000 รายการ (ลดลงสำหรับ PostgreSQL)
		smallBatchSize := 1000
		totalItems := len(detailList)

		for start := 0; start < totalItems; start += smallBatchSize {
			end := start + smallBatchSize
			if end > totalItems {
				end = totalItems
			}

			batchItems := detailList[start:end]

			// ใช้ bulk insert สำหรับ PostgreSQL
			columns := []string{
				"docdatetime", "docno", "line_number", "transflag", "itemcode", "barcode",
				"unitcode", "whcode", "locationcode", "totalqty", "price", "unitstand",
				"unitdivide", "averagecost", "calcamount", "balanceamount", "balance_qty",
				"guid", "unitcost", "docref",
			}

			var records [][]any
			for _, detail := range batchItems {
				record := []any{
					detail.DocDateTime,
					detail.DocNo,
					detail.LineNumber,
					detail.TransFlag,
					detail.ItemCode,
					detail.Barcode,
					detail.UnitCode,
					detail.WhCode,
					detail.LocationCode,
					detail.Qty,
					detail.PriceExcludeVat,
					detail.UnitStand,
					detail.UnitDivide,
					detail.AverageCost,
					detail.CalcAmount,
					detail.BalanceAmount,
					detail.BalanceQty,
					detail.Guid,
					detail.UnitCost,
					detail.DocRef,
				}
				records = append(records, record)
			}

			if len(records) > 0 {
				// เพิ่มกลไก retry และจัดการ memory error
				maxRetries := 3
				retryDelay := time.Second * 5
				var err error

				for retry := 0; retry < maxRetries; retry++ {
					err = mypg.BulkInsertWithCopy(ctx, db, "processstockcost", columns, records)
					if err == nil {
						logger.Info("Inserted batch %d-%d successfully", start+insertCounter, end+insertCounter)
						break
					}

					// ตรวจสอบว่าเป็น memory error หรือไม่
					if strings.Contains(err.Error(), "memory limit exceeded") {
						if retry < maxRetries-1 {
							logger.Info("Memory limit error on attempt %d, waiting %v before retry", retry+1, retryDelay)
							time.Sleep(retryDelay)
							// เพิ่มเวลารอในแต่ละครั้ง
							retryDelay = retryDelay * 2
						} else {
							logger.Info("Failed to insert after %d retries: %v", maxRetries, err)
							return err
						}
					} else {
						// ถ้าเป็น error อื่นที่ไม่ใช่เรื่อง memory ให้แจ้ง error เลย
						logger.Error("insert batch %d-%d to PostgreSQL: %v", start+insertCounter, end+insertCounter, err)
						return err
					}
				}
			} else {
				logger.Info("No values to insert for batch %d-%d", start+insertCounter, end+insertCounter)
			}
		}

		insertCounter += len(detailList)
		return nil
	}

	transFlagStrings := make([]string, len(myglobal.TransFlagsToProcess))
	for i, flag := range myglobal.TransFlagsToProcess {
		transFlagStrings[i] = strconv.Itoa(flag)
	}
	transFlagForQuery := strings.Join(transFlagStrings, ",")
	logger.Info("TransFlagsToProcess: %s", transFlagForQuery)

	logger.Info("Executing query for itemCode: %s", itemCodeForProcess)
	dataRows, err := mypg.QuerySelectAll(db, fmt.Sprintf(query, transFlagForQuery), itemCodeForProcess)
	if err != nil {
		logger.Error("select PostgreSQL: %v", err)
		return
	}

	detailSorted := []models.ProcessStockCostDetailStruct{}
	for index := 0; index < len(dataRows); index++ {
		row := dataRows[index]
		detailSorted = append(detailSorted, models.ProcessStockCostDetailStruct{
			ShopID:          shopId,
			DocDateTime:     mypg.GetTimeValue(row, "docdatetime"),
			DocNo:           mypg.GetStringValue(row, "docno"),
			LineNumber:      index + 1,
			TransFlag:       int(row["transflag"].(int64)), // PostgreSQL type conversion
			ItemCode:        mypg.GetStringValue(row, "itemcode"),
			Barcode:         mypg.GetStringValue(row, "barcode"),
			BarcodeMain:     mypg.GetStringValue(row, "barcode"), // ใช้ barcode เป็น barcodemain
			UnitCode:        mypg.GetStringValue(row, "unitcode"),
			WhCode:          mypg.GetStringValue(row, "whcode"),
			LocationCode:    mypg.GetStringValue(row, "locationcode"),
			Qty:             mypg.GetFloat64Value(row, "qty"),
			Price:           mypg.GetFloat64Value(row, "price"),
			PriceExcludeVat: mypg.GetFloat64Value(row, "priceexcludevat"),
			UnitStand:       mypg.GetFloat64Value(row, "unitstand"),
			UnitDivide:      mypg.GetFloat64Value(row, "unitdivide"),
			SumAmount:       mypg.GetFloat64Value(row, "sum_amount"),
			AverageCost:     0.0,
			UnitCost:        0.0,
			CalcAmount:      0.0,
			BalanceAmount:   0.0,
			BalanceQty:      0.0,
			Guid:            myglobal.GenUUID(),
			DocRef:          mypg.GetStringValue(row, "docref"),
		})
	}

	// รอ ** เรียงลำดับข้อมูลตามยอดคงเหลือ กรณีติดลบ หรือต้องการเรียงใหม่ หรือมีการกระจายจำนวน กรณีสินค้าหมด ต้องเพิ่มบรรทัด

	// เริ่มคำนวณต้นทุน
	for index := 0; index < len(detailSorted); index++ {
		detail := detailSorted[index]

		// 1. Standardize Qty
		factor := 1.0
		if detail.UnitDivide != 0 {
			factor = detail.UnitStand / detail.UnitDivide
		}
		qty := myglobal.RoundFloat64(detail.Qty*factor, pointQty)

		// 2. Determine CalcAmount and UnitCost
		var calcAmount float64
		var unitCost float64

		// Check TransFlag for specific logic
		switch detail.TransFlag {
		case 44, 56, 72, 20, 21, 62: // Out (Sell, Issue, Transfer Out)
			// Ensure Qty is negative for Out
			if qty > 0 {
				qty = qty * -1
			}
			// Use Current Average Cost
			calcAmount = myglobal.RoundFloat64(qty*averageCost, pointAmount)

		case 12, 310, 60, 61, 54, 66, 48: // In (Buy, Receive, Opening, Adjust Up, Return Out?)
			// Ensure Qty is positive for In
			if qty < 0 {
				qty = qty * -1
			}
			// Use SumAmount from doc (Cost)
			calcAmount = detail.SumAmount
			// Fallback if SumAmount is 0 but Price exists
			if calcAmount == 0 && detail.PriceExcludeVat != 0 {
				calcAmount = myglobal.RoundFloat64(qty*detail.PriceExcludeVat, pointAmount)
			}

		case 16: // Return In (from Customer)
			if qty < 0 {
				qty = qty * -1
			}
			// Try to use original cost logic (simplified here to use SumAmount or AverageCost)
			// If SumAmount is available and non-zero, use it.
			if detail.SumAmount != 0 {
				calcAmount = detail.SumAmount
			} else {
				// Use current Average Cost as fallback
				calcAmount = myglobal.RoundFloat64(qty*averageCost, pointAmount)
			}

		case 866, 868: // Cost Adjustment
			qty = 0
			calcAmount = detail.SumAmount

		default:
			// Default behavior
			calcAmount = detail.SumAmount
		}

		// 3. Update Balance
		balanceQty += qty

		// Handle near-zero balance (Stock Empty)
		if math.Abs(balanceQty) < 0.000001 {
			balanceQty = 0
			// Force BalanceAmount to 0 by adjusting calcAmount
			// This ensures we don't have leftover value when stock is 0
			calcAmount = myglobal.RoundFloat64(balanceAmount*-1, pointAmount)
		}

		balanceAmount += calcAmount

		// Prevent -0.00
		if math.Abs(balanceAmount) < 0.0001 {
			balanceAmount = 0
		}

		// 4. Calculate UnitCost for this transaction (for display/recording)
		if qty != 0 {
			unitCost = myglobal.RoundFloat64(calcAmount/qty, pointCost)
		}

		// 5. Calculate NEW Average Cost (Weighted Average)
		if balanceQty > 0 {
			averageCost = myglobal.RoundFloat64(balanceAmount/balanceQty, pointCost)
		}
		// If balanceQty <= 0, keep previous averageCost

		// 6. Prepare Record
		updatedDetail := detail
		updatedDetail.Qty = qty
		updatedDetail.UnitCost = unitCost
		updatedDetail.AverageCost = averageCost
		updatedDetail.CalcAmount = calcAmount
		updatedDetail.BalanceAmount = balanceAmount
		updatedDetail.BalanceQty = balanceQty

		// Append
		detailList = append(detailList, updatedDetail)
		allDetails = append(allDetails, updatedDetail)
		totalRecordsProcessed++

		// Flush to Postgres
		if len(detailList) >= 25000 {
			if err := insertBatch(); err != nil {
				return
			}
			detailList = []models.ProcessStockCostDetailStruct{}
		}

		// Memory Optimization
		if totalRecordsProcessed%50000 == 0 {
			runtime.GC()                       // บังคับ garbage collection
			time.Sleep(100 * time.Millisecond) // พักเล็กน้อย
		}
	}

	// สำหรับข้อมูลที่เหลือ
	if len(detailList) > 0 {
		if err := insertBatch(); err != nil {
			return
		}
	}

	logger.Info("CalcCost completed for itemCode: %s, processed %d records", itemCodeForProcess, totalRecordsProcessed)

	// บันทึกข้อมูล lot (ถ้ามีและต้องการบันทึก)
	if len(lots) > 0 {
		logger.Info("Saving lot data for itemCode: %s", itemCodeForProcess)

		if saver != nil {
			if err := saver.SaveLot(lots); err != nil {
				logger.Error("saver.SaveLot failed: %v", err)
			}
		} else {
			// แบ่งชุดข้อมูล lot
			smallBatchSize := 1000
			totalLots := len(lots)

			for start := 0; start < totalLots; start += smallBatchSize {
				end := start + smallBatchSize
				if end > totalLots {
					end = totalLots
				}

				batchLots := lots[start:end]

				columns := []string{
					"docdatetime", "lot_number", "docno", "transflag", "itemcode",
					"unitcode", "whcode", "locationcode", "qty", "price", "unitstand",
					"unitdivide", "cost", "balanceamount", "balance_qty", "guid_ref",
				}

				var records [][]any
				for _, lot := range batchLots {
					record := []any{
						lot.DocDateTime,
						lot.LotNumber,
						lot.DocNo,
						lot.TransFlag,
						lot.ItemCode,
						lot.UnitCode,
						lot.WhCode,
						lot.LocationCode,
						lot.Qty,
						lot.Price,
						lot.UnitStand,
						lot.UnitDivide,
						lot.Cost,
						lot.BalanceAmount,
						lot.BalanceQty,
						lot.GuidRef,
					}
					records = append(records, record)
				}

				if len(records) > 0 {
					err := mypg.BulkInsertWithCopy(ctx, db, "processstocklot", columns, records)
					if err != nil {
						logger.Error("insert lot batch %d-%d to PostgreSQL: %v", start, end, err)
					} else {
						logger.Info("Inserted lot batch %d-%d successfully", start, end)
					}
				}
			}
		}

		// ล้างข้อมูล lots เพื่อคืนหน่วยความจำ
		lots = nil
	}

	// save to clickhouse - ใช้ข้อมูลที่คำนวณไว้แล้วใน allDetails
	logger.Info("Saving %d records to ClickHouse for itemCode: %s", len(allDetails), itemCodeForProcess)

	if len(allDetails) > 0 {
		if saver != nil {
			if err := saver.SaveClickHouse(allDetails); err != nil {
				logger.Error("saver.SaveClickHouse failed: %v", err)
			}
		} else {
			// Async ClickHouse Processing
			// Capture data to avoid race condition with allDetails = nil
			detailsToSave := allDetails
			go func(details []models.ProcessStockCostDetailStruct) {
				// เชื่อมต่อ ClickHouse (ใช้ connection pool ไม่ต้อง close)
				clickHouseDB, err := myclickhouse.ClickHouseFastConnect()
				if err != nil {
					logger.Error("Failed to connect to ClickHouse: %v", err)
					return
				}
				result := process.ReplaceProcessStockCostPartition(context.Background(), clickHouseDB, shopId, itemCodeForProcess, details)
				if result.Success {
					logger.Info("ClickHouse partition replaced successfully: %s", result.Message)
				} else {
					logger.Error("Failed to replace ClickHouse partition: %s", result.Error)
				}
			}(detailsToSave)
		}
	} else {
		logger.Info("No data to save to ClickHouse for itemCode: %s", itemCodeForProcess)
	}

	// Update checksum after successful calculation (incremental mode)
	if incremental && currentChecksum != "" {
		if err := UpdateItemChecksum(ctx, db, shopId, itemCodeForProcess, currentChecksum); err != nil {
			logger.Warn("Failed to update checksum for %s: %v", itemCodeForProcess, err)
		} else {
			logger.Debug("Updated checksum for %s: %s", itemCodeForProcess, currentChecksum)
		}
	}

	// สุดท้าย ล้างข้อมูล detailList เพื่อคืนหน่วยความจำ
	detailList = nil
	allDetails = nil
}
