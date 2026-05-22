package process

import (
	"context"
	"fmt"
	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/models"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/google/uuid"
)

// ReplacePartitionResult ผลลัพธ์การ replace
type ReplacePartitionResult struct {
	Success bool    `json:"success"`
	ShopID string  `json:"shopid"`
	ItemCode string  `json:"itemcode"`
	Rows int64   `json:"rows"`
	Duration float64 `json:"duration"`
	Message string  `json:"message"`
	Error string  `json:"error,omitempty"`
}

// ReplaceProcessStockCostPartition replace partition ของ processstockcost table โดยส่ง array ข้อมูลเข้ามาโดยตรง
//
// Parameters:
//   - ctx: context สำหรับ timeout/cancellation
//   - conn: clickhouse.Conn ClickHouse connection
//   - shopID: รหัสร้าน
//   - itemCode: รหัสสินค้า
//   - data: array ของข้อมูลที่จะ insert
//
// Returns:
//   - *ReplacePartitionResult: ผลลัพธ์การทำงาน
//
// Example:
//
//	data := []ProcessStockCostDetailStruct{
//	    {ShopID: "SHOP001", ItemCode: "ITEM001", DocNo: "PO001", ...},
//	    {ShopID: "SHOP001", ItemCode: "ITEM001", DocNo: "PO002", ...},
//	}
//	result := ReplaceProcessStockCostPartition(ctx, conn, "SHOP001", "ITEM001", data)
func ReplaceProcessStockCostPartition(
	ctx context.Context,
	conn clickhouse.Conn,
	shopID string,
	itemCode string,
	data []models.ProcessStockCostDetailStruct,
) *ReplacePartitionResult {

	startTime := time.Now()

	result := &ReplacePartitionResult{
		ShopID:   shopID,
		ItemCode: itemCode,
	}

	// ตรวจสอบข้อมูล
	if len(data) == 0 {
		result.Error = "no data provided"
		result.Duration = time.Since(startTime).Seconds()
		return result
	}

	// สร้างชื่อ staging table ที่ unique
	stagingTable := "staging_" + strings.ReplaceAll(uuid.New().String(), "-", "_")

	// Defer cleanup staging table
	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		dropQuery := fmt.Sprintf("DROP TABLE IF EXISTS %s", stagingTable)
		if err := conn.Exec(cleanupCtx, dropQuery); err != nil {
			logger.Warn("Failed to drop staging table %s: %v", stagingTable, err)
		}
	}()

	// 1. สร้าง staging table
	createQuery := fmt.Sprintf("CREATE TABLE %s AS processstockcost", stagingTable)
	if err := conn.Exec(ctx, createQuery); err != nil {
		result.Error = fmt.Sprintf("failed to create staging table: %v", err)
		result.Duration = time.Since(startTime).Seconds()
		return result
	}

	// 2. Prepare batch insert
	insertQuery := fmt.Sprintf(`
        INSERT INTO %s (
            shopid, itemcode, docdatetime, docno, linenumber, transflag,
            barcodemain, barcode, unitcode, whcode, locationcode,
            totalqty, unitstand, unitdivide, price, averagecost,
            calcamount, balanceqty, balanceamount, guid, unitcost, docref
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `, stagingTable)

	batch, err := conn.PrepareBatch(ctx, insertQuery)
	if err != nil {
		result.Error = fmt.Sprintf("failed to prepare batch: %v", err)
		result.Duration = time.Since(startTime).Seconds()
		return result
	}

	// 3. Batch insert ข้อมูล
	for _, row := range data {
		err := batch.Append(
			shopID,              // shopid
			itemCode,            // itemcode
			row.DocDateTime,     // docdatetime
			row.DocNo,           // docno
			row.LineNumber,      // linenumber
			row.TransFlag,       // transflag
			row.BarcodeMain,     // barcodemain (ถ้าไม่มีให้ใช้ Barcode)
			row.Barcode,         // barcode
			row.UnitCode,        // unitcode
			row.WhCode,          // whcode
			row.LocationCode,    // locationcode
			row.Qty,             // totalqty
			row.UnitStand,       // unitstand
			row.UnitDivide,      // unitdivide
			row.PriceExcludeVat, // price
			row.AverageCost,     // averagecost
			row.CalcAmount,      // calcamount
			row.BalanceQty,      // balanceqty
			row.BalanceAmount,   // balanceamount
			row.Guid,            // guid
			row.UnitCost,        // unitcost
			row.DocRef,          // docref
		)
		if err != nil {
			result.Error = fmt.Sprintf("failed to append row: %v", err)
			result.Duration = time.Since(startTime).Seconds()
			return result
		}
	}

	// Send batch
	if err := batch.Send(); err != nil {
		result.Error = fmt.Sprintf("failed to send batch: %v", err)
		result.Duration = time.Since(startTime).Seconds()
		return result
	}

	// 4. REPLACE PARTITION
	replaceQuery := fmt.Sprintf(`
        ALTER TABLE processstockcost
        REPLACE PARTITION ('%s', '%s')
        FROM %s
    `, shopID, itemCode, stagingTable)

	if err := conn.Exec(ctx, replaceQuery); err != nil {
		result.Error = fmt.Sprintf("failed to replace partition: %v", err)
		result.Duration = time.Since(startTime).Seconds()
		return result
	}

	// Success
	result.Success = true
	result.Rows = int64(len(data))
	result.Duration = time.Since(startTime).Seconds()
	result.Message = fmt.Sprintf("Replaced %d rows in %.2fs", result.Rows, result.Duration)

	return result
}
