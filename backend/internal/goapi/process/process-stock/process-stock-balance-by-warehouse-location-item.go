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

func ProcessProductBalanceByLocationAndItem(shopId string, finalDate string, balanceOnly bool, itemCodeList []string, warehouseList []models.WarehouseListItemStruct) (result models.ResultModel) {
	// Default to Thailand timezone for backward compatibility
	return ProcessProductBalanceByLocationAndItemWithTimezone(shopId, finalDate, balanceOnly, itemCodeList, warehouseList, "TH")
}

func ProcessProductBalanceByLocationAndItemWithTimezone(shopId string, finalDate string, balanceOnly bool, itemCodeList []string, warehouseList []models.WarehouseListItemStruct, timezoneCode string) (result models.ResultModel) {
	logger.Info("ProcessProductBalanceByLocationCodeBarcode with Timezone: %s", timezoneCode)
	// whereHouseList
	logger.Info("warehouseList: %+v", warehouseList)

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
			// Start from $2 since $1 is used twice for finalDate in the main query
			placeholders[i] = fmt.Sprintf("$%d", len(itemCodeArgs)+2)
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
				// Start from $2 + len(itemCodeArgs) since $1 is used twice for finalDate
				warehousePlaceholders = append(warehousePlaceholders, fmt.Sprintf("$%d", 2+len(itemCodeArgs)+len(warehouseArgs)))
				warehouseArgs = append(warehouseArgs, wareHouseCode.Code)
			}
		}
		wareHouseCodeWhere = " AND lc.whcode IN (" + strings.Join(warehousePlaceholders, ",") + ")"
	}

	var wareHouseDataList []models.ProductBalanceByWareHouseAndLocationAndBarcodeStruct

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
    lc.whcode AS whcode,
    lc.locationcode AS locationcode,
    lc.itemcode AS itemcode,
    n.itemname AS itemname,
	pb2.unitcode AS unitcode,					
    pb2.unitname AS unitname,
    sb.total_balance AS balanceqty,
    COALESCE(ap.countpacking, 0) AS countpacking,
    b.barcodelist AS barcodelist
FROM
(
    SELECT
        itemcode,
        whcode,
        locationcode,
        ROW_NUMBER() OVER (PARTITION BY itemcode, whcode, locationcode ORDER BY docdatetime DESC) AS rn
    FROM processstockcost
    WHERE ` + dateCondition + `
) AS lc
JOIN (
    SELECT
        itemcode,
        whcode,
        locationcode,
        SUM(totalqty * (unitstand / NULLIF(unitdivide, 0))) AS total_balance
    FROM docdetail
    WHERE ` + dateCondition + `
		AND transflag IN (` + transFlagList + `)
    GROUP BY itemcode, whcode, locationcode
) AS sb ON lc.itemcode = sb.itemcode 
       AND lc.whcode = sb.whcode 
       AND lc.locationcode = sb.locationcode
LEFT JOIN item_names n ON lc.itemcode = n.itemcode
LEFT JOIN barcode_list b ON lc.itemcode = b.itemcode
LEFT JOIN auto_packing ap ON lc.itemcode = ap.itemcode
LEFT JOIN (
	SELECT
		itemcode,
		MAX(unitcode) AS unitcode,
		MAX(unitname) AS unitname
	FROM productbarcode
	WHERE barcoderefunitstand = 1 
	  AND barcoderefunitdivide = 1
	GROUP BY itemcode
) pb2 ON lc.itemcode = pb2.itemcode
WHERE lc.rn = 1` + wareHouseCodeWhere + itemCodeListWhere + `
ORDER BY lc.whcode, lc.locationcode, lc.itemcode
`

	// query for pgadmin
	logger.Info("query: %s", mypg.ReplaceQueryParams(query, allArgs...))

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

	for _, row := range rows {
		whCode := mypg.GetStringValue(row, "whcode")
		locationCode := mypg.GetStringValue(row, "locationcode")
		itemCode := mypg.GetStringValue(row, "itemcode")
		itemName := mypg.GetStringValue(row, "item_name")
		unitCode := mypg.GetStringValue(row, "unitcode")
		unitName := mypg.GetStringValue(row, "unit_name")
		barcodeList := mypg.GetStringValue(row, "barcodelist")
		balanceQty := mypg.GetFloat64Value(row, "balance_qty")

		var countpacking int64
		if row["countpacking"] != nil {
			// Auto packing
			countpacking = mypg.GetInt64Value(row, "countpacking")
		}

		wareHouseDataList = append(wareHouseDataList, models.ProductBalanceByWareHouseAndLocationAndBarcodeStruct{
			WareHouseCode: whCode,
			LocationCode:  locationCode,
			ItemCode:      itemCode,
			ItemName:      itemName,
			UnitCode:      unitCode,
			UnitName:      unitName,
			BarcodeList:   barcodeList,
			BalanceQty:    balanceQty,
			BalanceWord:   "",
			IsAutoPacking: uint64(countpacking),
		})
	}

	if wareHouseSelectedCount > 0 {
		// กรองรายการที่ไม่อยู่ในคลังสินค้าที่เลือก
		// หรือ กรณีที่ไม่มีการเลือกคลังสินค้าเลย ให้ดึงทุกคลังสินค้า
		var wareHouseTempDataList []models.ProductBalanceByWareHouseAndLocationAndBarcodeStruct
		for _, p := range wareHouseDataList {
			found := false
			for _, wareHouse := range warehouseList {
				if p.WareHouseCode == wareHouse.Code && wareHouse.IsSelected {
					for _, location := range wareHouse.Locations {
						if p.LocationCode == location.Code && location.IsSelected {
							found = true
							break
						}
					}
				}
			}
			if found {
				wareHouseTempDataList = append(wareHouseTempDataList, p)
			}
		}
		wareHouseDataList = wareHouseTempDataList
	}

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

	// Helper to ensure JSON encoding never receives NaN/Inf
	sanitizeFloat := func(val float64) float64 {
		if math.IsNaN(val) || math.IsInf(val, 0) {
			return 0
		}
		return val
	}

	sanitizeRecord := func(record models.ProductBalanceByWareHouseAndLocationAndBarcodeStruct) models.ProductBalanceByWareHouseAndLocationAndBarcodeStruct {
		record.BalanceQty = sanitizeFloat(record.BalanceQty)
		return record
	}

	// write to PostgreSQL result table
	guid := myglobal.GenUUID()
	lineNumber := 0
	totalLine := len(wareHouseDataList)
	logger.Info("Total records to insert: %d", totalLine)

	// Use bulk insert for PostgreSQL
	columns := []string{"guid", "docdatetime", "line_number", "datajson"}
	var records [][]any

	for _, b := range wareHouseDataList {
		lineNumber++
		cleanRecord := sanitizeRecord(b)
		jsonData, errMarshal := json.Marshal(cleanRecord)
		if errMarshal != nil {
			logger.Error("marshal balance record wh=%s loc=%s item=%s line=%d: %v", b.WareHouseCode, b.LocationCode, b.ItemCode, lineNumber, errMarshal)
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
			records = [][]any{}
		}
	}

	// Insert remaining records
	if len(records) > 0 {
		err := mypg.BulkInsertWithCopy(ctx, db, "result", columns, records)
		if err != nil {
			logger.Error("inserting final batch to result table: %v", err)
		}
	}

	return models.ResultModel{
		Success:     true,
		TotalPage:   int(math.Ceil(float64(totalLine) / 10000.0)),
		TotalRecord: totalLine,
		Guid:        guid,
	}
}
