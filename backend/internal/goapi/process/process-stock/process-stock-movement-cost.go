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

func ProcessProductMovement(shopId string, fromDate string, endDate string, movementOnly bool, itemCodeList []string, warehouseList []models.WarehouseListItemStruct) models.ResultModel {

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

	itemCodeDataList := []models.ProcessStockMovementStruct{}
	{
		// ดึงรายการสินค้าทั้งหมด ที่มีการเคลื่อนไหว
		var query string
		var args []any

		if len(itemCodeList) > 0 {
			// Build parameterized query for item code list
			placeholders := make([]string, len(itemCodeList))
			for i, itemCode := range itemCodeList {
				placeholders[i] = fmt.Sprintf("$%d", i+3)
				args = append(args, itemCode)
			}
			query = "SELECT DISTINCT itemcode FROM processstockcost WHERE itemcode IN (" + strings.Join(placeholders, ",") + ") AND docdatetime::date >= $1::date AND docdatetime::date <= $2::date ORDER BY itemcode"
			// Prepend fromDate and endDate to args
			args = append([]any{fromDate, endDate}, args...)
		} else {
			query = "SELECT DISTINCT itemcode FROM processstockcost WHERE docdatetime::date >= $1::date AND docdatetime::date <= $2::date ORDER BY itemcode"
			args = []any{fromDate, endDate}
		}

		logger.Info("query: %s", query)

		rows, err := mypg.QuerySelectAll(db, query, args...)
		if err != nil {
			logger.Info("Failed to query PostgreSQL: %v", err)
			return models.ResultModel{
				Success:     false,
				TotalPage:   0,
				TotalRecord: 0,
				Guid:        "",
			}
		}
		for _, row := range rows {
			itemCodeDataList = append(itemCodeDataList, models.ProcessStockMovementStruct{
				ItemCode: mypg.GetStringValue(row, "itemcode"),
				Name:     "",
				UnitCode: "",
				UnitName: "",
				Details:  []models.ProcessStockMovementDetailStruct{},
			})
		}
	}
	{
		// ดึงรายละเอียดสินค้าทั้งหมดในครั้งเดียว (ปรับปรุงแล้ว)
		if len(itemCodeDataList) > 0 {
			// สร้าง list ของ itemcodes
			var itemCodes []string
			itemCodeMap := make(map[string]int) // map itemcode -> index ใน itemCodeDataList

			for i, item := range itemCodeDataList {
				itemCodes = append(itemCodes, item.ItemCode)
				itemCodeMap[item.ItemCode] = i
			}

			// สร้าง placeholders สำหรับ IN clause
			placeholders := make([]string, len(itemCodes))
			args := make([]interface{}, len(itemCodes))

			for i, itemCode := range itemCodes {
				placeholders[i] = fmt.Sprintf("$%d", i+1)
				args[i] = itemCode
			}

			queryAllProducts := fmt.Sprintf(`SELECT 
				p1.itemcode,
				CASE 
					WHEN COUNT(DISTINCT p1.barcode) > 0 THEN 
						STRING_AGG(DISTINCT p1.name0, ' / ' ORDER BY p1.name0) || ' [' || STRING_AGG(DISTINCT p1.barcode, ', ' ORDER BY p1.barcode) || ']'
					ELSE 
						STRING_AGG(DISTINCT p1.name0, ' / ' ORDER BY p1.name0)
				END AS name,
				MAX(CASE WHEN p1.barcoderefunitstand = 1 AND p1.barcoderefunitdivide = 1 THEN p1.unitcode END) AS unitcode,
				MAX(CASE WHEN p1.barcoderefunitstand = 1 AND p1.barcoderefunitdivide = 1 THEN p1.unitname END) AS unitname
			FROM productbarcode p1			
			WHERE p1.itemcode IN (%s)
			GROUP BY p1.itemcode`, strings.Join(placeholders, ","))

			logger.Info("Querying product details for %d items...", len(itemCodes))
			rows, err := mypg.QuerySelectAll(db, queryAllProducts, args...)
			if err != nil {
				logger.Error("querying product details: %v", err)
			} else {
				// แจกจ่ายข้อมูลไปยัง itemCodeDataList
				for _, row := range rows {
					itemCode := row["itemcode"].(string)
					if productIndex, exists := itemCodeMap[itemCode]; exists {
						itemCodeDataList[productIndex].Name = mypg.GetStringValue(row, "name")
						itemCodeDataList[productIndex].UnitCode = mypg.GetStringValue(row, "unitcode")
						itemCodeDataList[productIndex].UnitName = mypg.GetStringValue(row, "unitname")
					}
				}
			}
		}
	}
	{
		// ดึงยอดคงเหลือจาก processstockcost ทั้งหมดในครั้งเดียว (ปรับปรุงแล้ว)
		if len(itemCodeDataList) > 0 {
			// สร้าง list ของ itemcodes
			var itemCodes []string
			itemCodeMap := make(map[string]int) // map itemcode -> index ใน itemCodeDataList

			for i, item := range itemCodeDataList {
				itemCodes = append(itemCodes, item.ItemCode)
				itemCodeMap[item.ItemCode] = i
			}

			// สร้าง placeholders สำหรับ IN clause
			placeholders := make([]string, len(itemCodes))
			args := make([]interface{}, 0, len(itemCodes)+2)
			args = append(args, fromDate, endDate)

			for i, itemCode := range itemCodes {
				placeholders[i] = fmt.Sprintf("$%d", i+3)
				args = append(args, itemCode)
			}

			// Query เดียวสำหรับทุก itemcode
			queryAllStockMovement := fmt.Sprintf(`SELECT 
				docdatetime,
				barcode,
				itemcode,
				docno,
				transflag,
				unitcode,
				whcode,
				locationcode,
				totalqty,
				unitstand,
				unitdivide,
				price,
				averagecost,
				balanceqty, 
				calcamount,
				balanceamount,
				unitcost,
				docref
			FROM processstockcost 
			WHERE 
				docdatetime::date >= $1::date 
				AND docdatetime::date <= $2::date 
				AND itemcode IN (%s)
			ORDER BY itemcode, docdatetime, linenumber, totalqty`,
				strings.Join(placeholders, ","))

			logger.Info("Querying stock movement for %d items from %s to %s...", len(itemCodes), fromDate, endDate)
			rows, err := mypg.QuerySelectAll(db, queryAllStockMovement, args...)
			if err != nil {
				logger.Error("querying stock movement: %v", err)
			} else {
				logger.Info("Found %d stock movement records", len(rows))
				// แจกจ่ายข้อมูลไปยัง itemCodeDataList
				for _, row := range rows {
					itemCode := mypg.GetStringValue(row, "itemcode")
					if productIndex, exists := itemCodeMap[itemCode]; exists {
						unitCodeFull := mypg.GetStringValue(row, "unitcode")
						unitStand := mypg.GetFloat64Value(row, "unitstand")
						unitDivide := mypg.GetFloat64Value(row, "unitdivide")
						if unitStand != 1 || unitDivide != 1 {
							unitCodeFull = unitCodeFull + " (" + myglobal.FormatNumber(unitStand, 0) + ":" + myglobal.FormatNumber(unitDivide, 0) + ")"
						}
						docRefFull := mypg.GetStringValue(row, "docref")
						if docRefFull != "" {
							docRefFull = " "
						}
						docRefFull += "[" + mypg.GetStringValue(row, "barcode") + "]"
						itemCodeDataList[productIndex].Details = append(itemCodeDataList[productIndex].Details, models.ProcessStockMovementDetailStruct{
							DocDateTime:   mypg.GetTimeValue(row, "docdatetime"),
							DocNo:         mypg.GetStringValue(row, "docno"),
							ItemCode:      mypg.GetStringValue(row, "itemcode"),
							TransFlag:     int(mypg.GetInt64Value(row, "transflag")),
							UnitCode:      unitCodeFull,
							WhCode:        mypg.GetStringValue(row, "whcode"),
							LocationCode:  mypg.GetStringValue(row, "locationcode"),
							TotalQty:      mypg.GetFloat64Value(row, "totalqty"),
							UnitStand:     unitStand,
							UnitDivide:    unitDivide,
							Price:         mypg.GetFloat64Value(row, "price"),
							AverageCost:   mypg.GetFloat64Value(row, "averagecost"),
							BalanceQty:    mypg.GetFloat64Value(row, "balanceqty"),
							CalcAmount:    mypg.GetFloat64Value(row, "calcamount"),
							BalanceAmount: mypg.GetFloat64Value(row, "balanceamount"),
							UnitCost:      mypg.GetFloat64Value(row, "unitcost"),
							DocRef:        docRefFull,
						})
					}
				}
			}
		}
	}

	// Helper to guard JSON encoding against NaN/Inf values
	sanitizeFloat := func(val float64) float64 {
		if math.IsNaN(val) || math.IsInf(val, 0) {
			return 0
		}
		return val
	}

	sanitizeDetail := func(detail models.ProcessStockMovementDetailStruct) models.ProcessStockMovementDetailStruct {
		detail.TotalQty = sanitizeFloat(detail.TotalQty)
		detail.Price = sanitizeFloat(detail.Price)
		detail.UnitStand = sanitizeFloat(detail.UnitStand)
		detail.UnitDivide = sanitizeFloat(detail.UnitDivide)
		detail.AverageCost = sanitizeFloat(detail.AverageCost)
		detail.CalcAmount = sanitizeFloat(detail.CalcAmount)
		detail.BalanceAmount = sanitizeFloat(detail.BalanceAmount)
		detail.BalanceQty = sanitizeFloat(detail.BalanceQty)
		detail.UnitCost = sanitizeFloat(detail.UnitCost)
		return detail
	}

	sanitizeRecord := func(record models.ProcessStockMovementStruct) models.ProcessStockMovementStruct {
		for i := range record.Details {
			record.Details[i] = sanitizeDetail(record.Details[i])
		}
		return record
	}

	// write to PostgreSQL result table
	guid := myglobal.GenUUID()
	lineNumber := 0
	totalLine := len(itemCodeDataList)

	// Use bulk insert for PostgreSQL
	columns := []string{"guid", "docdatetime", "linenumber", "datajson"}
	var records [][]any

	for _, b := range itemCodeDataList {
		lineNumber++
		cleanRecord := sanitizeRecord(b)
		jsonData, errMarshal := json.Marshal(cleanRecord)
		if errMarshal != nil {
			logger.Error("marshal movement record item=%s line=%d: %v", b.ItemCode, lineNumber, errMarshal)
			continue
		}

		record := []any{
			guid,
			time.Now(),
			lineNumber,
			string(jsonData), // แปลง []byte เป็น string สำหรับ JSONB
		}
		records = append(records, record)

		// Insert in batches of 10000 or when reaching the end
		if len(records) == 10000 || lineNumber == totalLine {
			logger.Info("Inserting batch of %d records to result table", len(records))
			err := mypg.BulkInsertWithCopy(ctx, db, "result", columns, records)
			if err != nil {
				logger.Error("inserting batch to result table: %v", err)
			}
			records = [][]any{}
		}
	}
	logger.Info("Total lines processed: %d", totalLine)

	result := models.ResultModel{
		Success:     true,
		TotalPage:   int(math.Ceil(float64(totalLine) / 10000.0)),
		TotalRecord: totalLine,
		Guid:        guid,
	}
	return result
}
