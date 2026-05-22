package processstock

import (
	"context"
	"encoding/json"
	"fmt"
	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/models"
	"math"
	"strings"
	"time"

	"smlcloudplatform/internal/goapi/myglobal"
	"smlcloudplatform/internal/goapi/mypg"
)

func ProcessProductBalanceByWareHouseAndItem(shopId string, finalDate string, balanceOnly bool, itemCodeList []string, warehouseList []models.WarehouseListItemStruct) (result models.ResultModel) {
	// Default to Thailand timezone for backward compatibility
	return ProcessProductBalanceByWareHouseAndItemWithTimezone(shopId, finalDate, balanceOnly, itemCodeList, warehouseList, "TH")
}

// Deprecated: ใช้ ProcessProductBalanceByWhCodeBarcodeWithTimezone แทน
func ProcessProductBalanceByWareHouseAndItemWithCountry(shopId string, finalDate string, balanceOnly bool, itemCodeList []string, warehouseList []models.WarehouseListItemStruct, countryCode string) (result models.ResultModel) {
	// Convert country code to timezone code for backward compatibility
	timezoneCode := countryCode
	switch countryCode {
	case "TH":
		timezoneCode = "TH"
	case "US":
		timezoneCode = "US_EST" // Default to Eastern Time
	}
	return ProcessProductBalanceByWareHouseAndItemWithTimezone(shopId, finalDate, balanceOnly, itemCodeList, warehouseList, timezoneCode)
}

func ProcessProductBalanceByWareHouseAndItemWithTimezone(shopId string, finalDate string, balanceOnly bool, itemCodeList []string, warehouseList []models.WarehouseListItemStruct, timezoneCode string) (result models.ResultModel) {
	logger.Info("ProcessProductBalanceByWhCodeBarcode with Timezone: %s", timezoneCode)
	startTime := time.Now()

	ctx := context.Background()

	db, err := mypg.PgSqlFastConnect(shopId)
	if err != nil {
		logger.Info("Failed to connect to PostgreSQL: %v", err)
		return models.ResultModel{
			Success:     false,
			TotalPage:   0,
			TotalRecord: 0,
			Guid:        "",
		}
	}

	// Build parameterized query for barcode list
	var itemCodeArgs []any
	var itemCodeListWhere string
	if len(itemCodeList) > 0 {
		placeholders := make([]string, len(itemCodeList))
		for i, barcode := range itemCodeList {
			placeholders[i] = fmt.Sprintf("$%d", len(itemCodeArgs)+2) // +2 because finalDate is $1
			itemCodeArgs = append(itemCodeArgs, barcode)
		}
		itemCodeListWhere = " AND lc.itemcode IN (" + strings.Join(placeholders, ",") + ")"
	}

	// Build parameterized query for warehouse codes
	var warehouseArgs []any
	var wareHouseCodeWhere string
	wareHouseSelectedCount := 0
	for _, warehouse := range warehouseList {
		if warehouse.IsSelected {
			wareHouseSelectedCount++
		}
	}
	if wareHouseSelectedCount > 0 {
		var warehousePlaceholders []string
		for _, wareHouseCode := range warehouseList {
			if wareHouseCode.IsSelected {
				warehousePlaceholders = append(warehousePlaceholders, fmt.Sprintf("$%d", len(itemCodeArgs)+len(warehouseArgs)+2))
				warehouseArgs = append(warehouseArgs, wareHouseCode.Code)
			}
		}
		wareHouseCodeWhere = " AND lc.whcode IN (" + strings.Join(warehousePlaceholders, ",") + ")"
	}

	var wareHouseDataList []models.ProductBalanceByWareHouseAndBarcodeStruct

	// สร้าง date condition ที่รองรับ timezone
	dateCondition, conditionArgs, err := mypg.BuildDateConditionSQLWithTimeZone("docdatetime", finalDate, timezoneCode)
	if err != nil {
		logger.Error("building date condition for timezone %s: %v", timezoneCode, err)
		return models.ResultModel{
			Success:     false,
			TotalPage:   0,
			TotalRecord: 0,
			Guid:        "",
		}
	}

	// Add date condition args first, then append other args
	allArgs := conditionArgs
	allArgs = append(allArgs, itemCodeArgs...)
	allArgs = append(allArgs, warehouseArgs...)

	transFlagList := myglobal.GetTransFlagsForQuery()

	query := `
WITH item_names AS (
	SELECT
		itemcode,
		STRING_AGG(DISTINCT name0, ', ') AS itemname
	FROM productbarcode
	WHERE itemcode IS NOT NULL 
	AND itemcode <> ''
	GROUP BY itemcode
),
barcode_list AS (
	SELECT
		itemcode,
		STRING_AGG(DISTINCT barcode, ', ') AS barcodelist
	FROM productbarcode
	WHERE itemcode IS NOT NULL 
	AND itemcode <> ''
	GROUP BY itemcode
),
auto_packing AS (
	SELECT
		itemcode,
		COUNT(DISTINCT CASE 
			WHEN barcoderefunitstand > 0 AND barcoderefunitdivide > 0 
			AND (barcoderefunitstand != barcoderefunitdivide)
			THEN barcoderefunitstand || '-' || barcoderefunitdivide 
		END) AS countpacking
	FROM productbarcode
	WHERE itemcode IS NOT NULL 
	AND itemcode <> ''
	GROUP BY itemcode
)
SELECT 
    main.*,
    n.itemname AS itemname,
    pb2.unitcode AS unitcode,
    pb2.unitname AS unitname,
    b.barcodelist AS barcodelist
FROM (
    SELECT DISTINCT
        lc.whcode AS whcode,
        lc.itemcode AS itemcode,
        sb.total_balance AS balanceqty,
        COALESCE(ap.countpacking, 0) AS countpacking
    FROM (
        -- Latest stock cost records by itemcode and warehouse
        SELECT
            itemcode,
            whcode,
            ROW_NUMBER() OVER (
                PARTITION BY itemcode, whcode 
                ORDER BY docdatetime DESC
            ) AS rn
        FROM processstockcost
        WHERE ` + dateCondition + `
    ) AS lc
    JOIN (
        -- Total balance by itemcode and warehouse
        SELECT
            itemcode,
            whcode,
            SUM(totalqty * (unitstand / NULLIF(unitdivide, 0))) AS total_balance
        FROM docdetail
        WHERE ` + dateCondition + `
			AND transflag IN (` + transFlagList + `)
        GROUP BY itemcode, whcode
    ) AS sb ON lc.itemcode = sb.itemcode
           AND lc.whcode = sb.whcode
	LEFT JOIN auto_packing ap ON lc.itemcode = ap.itemcode
	WHERE lc.rn = 1` + wareHouseCodeWhere + itemCodeListWhere + `
	ORDER BY lc.whcode, lc.itemcode
) AS main
LEFT JOIN item_names n ON main.itemcode = n.itemcode
LEFT JOIN barcode_list b ON main.itemcode = b.itemcode
LEFT JOIN (
	SELECT
		itemcode,
		MAX(unitcode) AS unitcode,
		MAX(unitname) AS unitname
	FROM productbarcode
	WHERE barcoderefunitstand = 1 
	  AND barcoderefunitdivide = 1
	GROUP BY itemcode
) pb2 ON main.itemcode = pb2.itemcode
`

	logger.Info("Executing warehouse balance query (items=%d, warehouses=%d)", len(itemCodeList), wareHouseSelectedCount)
	logger.Debug("Warehouse query preview: %s", mypg.ReplaceQueryParams(query, allArgs...))

	rows, err := mypg.QuerySelectAll(db, query, allArgs...)
	if err != nil {
		logger.Error("querying PostgreSQL: %v", err)
		return models.ResultModel{
			Success:     false,
			TotalPage:   0,
			TotalRecord: 0,
			Guid:        "",
		}
	}

	logger.Info("Warehouse query returned %d rows (elapsed=%s)", len(rows), time.Since(startTime))

	for _, row := range rows {
		whcode := mypg.GetStringValue(row, "whcode")
		itemCode := mypg.GetStringValue(row, "itemcode")
		barcodeList := mypg.GetStringValue(row, "barcodelist")
		balanceqty := mypg.GetFloat64Value(row, "balance_qty")
		itemName := mypg.GetStringValue(row, "item_name")
		unitcode := mypg.GetStringValue(row, "unitcode")
		unitname := mypg.GetStringValue(row, "unit_name")

		var countpacking int64
		if row["countpacking"] != nil {
			countpacking = row["countpacking"].(int64)
		}

		wareHouseDataList = append(wareHouseDataList, models.ProductBalanceByWareHouseAndBarcodeStruct{
			WareHouseCode: whcode,
			ItemCode:      itemCode,
			ItemName:      itemName,
			BarcodeList:   barcodeList,
			UnitCode:      unitcode,
			UnitName:      unitname,
			BalanceQty:    balanceqty,
			BalanceWord:   "",
			IsAutoPacking: uint64(countpacking),
		})
	}

	// เตรียมข้อมูล auto packing ล่วงหน้าเพื่อลดจำนวน query
	uniqueAutoPackingItems := make([]string, 0, len(wareHouseDataList))
	seenAutoPacking := make(map[string]struct{})
	for _, p := range wareHouseDataList {
		if p.IsAutoPacking > 0 && p.ItemCode != "" {
			if _, exists := seenAutoPacking[p.ItemCode]; !exists {
				seenAutoPacking[p.ItemCode] = struct{}{}
				uniqueAutoPackingItems = append(uniqueAutoPackingItems, p.ItemCode)
			}
		}
	}
	packingCache := BuildAutoPackingCache(db, uniqueAutoPackingItems)
	fetchPackingForItem := func(itemCode string) []models.ProductBarcodePackingStruct {
		return FetchPackingForItem(db, packingCache, itemCode)
	}

	// qty word
	logger.Info("Formatting qty words for %d warehouse rows", len(wareHouseDataList))
	for indexProduct, p := range wareHouseDataList {
		if p.IsAutoPacking > 0 {
			barcodePacking := fetchPackingForItem(p.ItemCode)
			if len(barcodePacking) == 0 {
				wareHouseDataList[indexProduct].BalanceWord = fmt.Sprintf("%s %s", myglobal.CalcStockQtyWordCut(p.BalanceQty), p.UnitName)
				continue
			}
			wareHouseDataList[indexProduct].BalanceWord = myglobal.CalcStockQtyWord(p.BalanceQty, barcodePacking)
		} else {
			// ไม่ auto packing - มีแค่หน่วยเดียว
			wareHouseDataList[indexProduct].BalanceWord = fmt.Sprintf("%s %s", myglobal.CalcStockQtyWordCut(p.BalanceQty), p.UnitName)
		}
	}

	// Helper to ensure json.Marshal never sees NaN/Inf values
	sanitizeFloat := func(val float64) float64 {
		if math.IsNaN(val) || math.IsInf(val, 0) {
			return 0
		}
		return val
	}

	sanitizeRecord := func(record models.ProductBalanceByWareHouseAndBarcodeStruct) models.ProductBalanceByWareHouseAndBarcodeStruct {
		record.BalanceQty = sanitizeFloat(record.BalanceQty)
		return record
	}

	// write to PostgreSQL result table
	guid := myglobal.GenUUID()
	lineNumber := 0
	totalLine := len(wareHouseDataList)
	logger.Info("Warehouse report ready: total rows=%d (elapsed=%s)", totalLine, time.Since(startTime))

	// Use bulk insert for PostgreSQL
	columns := []string{"guid", "docdatetime", "line_number", "datajson"}
	var records [][]any

	bulkStart := time.Now()
	logger.Info("Starting bulk insert for guid=%s", guid)
	for _, b := range wareHouseDataList {
		lineNumber++
		cleanRecord := sanitizeRecord(b)
		jsonData, errMarshal := json.Marshal(cleanRecord)
		if errMarshal != nil {
			logger.Error("marshal balance record wh=%s item=%s line=%d: %v", b.WareHouseCode, b.ItemCode, lineNumber, errMarshal)
			continue
		}

		record := []any{
			guid,
			time.Now(),
			lineNumber,
			string(jsonData), // แปลง []byte เป็น string สำหรับ JSONB
		}
		records = append(records, record)

		// Insert in batches of 10000
		if len(records) == 10000 {
			err := mypg.BulkInsertWithCopy(ctx, db, "result", columns, records)
			if err != nil {
				logger.Error("inserting batch to result table: %v", err)
			}
			logger.Info("Inserted batch of %d lines (progress=%d/%d) elapsed=%s", 10000, lineNumber, totalLine, time.Since(bulkStart))
			records = [][]any{}
		}
	}

	// Insert remaining records
	if len(records) > 0 {
		err := mypg.BulkInsertWithCopy(ctx, db, "result", columns, records)
		if err != nil {
			logger.Error("inserting final batch to result table: %v", err)
		}
		logger.Info("Inserted final batch of %d lines (total=%d) bulkElapsed=%s", len(records), totalLine, time.Since(bulkStart))
	}

	logger.Info("Warehouse report stored guid=%s totalLine=%d (total elapsed=%s)", guid, totalLine, time.Since(startTime))

	return models.ResultModel{
		Success:     true,
		TotalPage:   int(math.Ceil(float64(totalLine) / 10000.0)),
		TotalRecord: totalLine,
		Guid:        guid,
	}
}
