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

func ProcessProductBalanceByItemAndWareHouseAndLocation(shopId string, condition int, finalDate string, balanceOnly bool, itemCodeList []string, warehouseList []models.WarehouseListItemStruct) (result models.ResultModel) {
	// Default to Thailand timezone for backward compatibility
	return ProcessProductBalanceByItemAndWareHouseAndLocationWithTimezone(shopId, condition, finalDate, balanceOnly, itemCodeList, warehouseList, "TH")
}

// Deprecated: ใช้ ProcessProductBalanceByBarcodeWhCodeLocationCodeWithTimezone แทน
func ProcessProductBalanceByItemAndWareHouseAndLocationWithCountry(shopId string, condition int, finalDate string, balanceOnly bool, itemCodeList []string, warehouseList []models.WarehouseListItemStruct, countryCode string) (result models.ResultModel) {
	// Convert country code to timezone code for backward compatibility
	timezoneCode := countryCode
	switch countryCode {
	case "TH":
		timezoneCode = "TH"
	case "US":
		timezoneCode = "US_EST" // Default to Eastern Time
	}
	return ProcessProductBalanceByItemAndWareHouseAndLocationWithTimezone(shopId, condition, finalDate, balanceOnly, itemCodeList, warehouseList, timezoneCode)
}

func ProcessProductBalanceByItemAndWareHouseAndLocationWithTimezone(shopId string, condition int, finalDate string, balanceOnly bool, itemCodeList []string, warehouseList []models.WarehouseListItemStruct, timezoneCode string) (result models.ResultModel) {
	overallStart := time.Now()

	ctx := context.Background()

	db, err := mypg.PgSqlFastConnect(shopId)
	if err != nil {
		logger.Info("Failed to connect to PostgreSQL: %v", err)
		return
	}

	itemCodeMainList := []models.ProductBalanceByCodeStruct{}
	{
		// ดึงรายละเอียดสินค้า
		logger.Info("ดึงรายละเอียดสินค้า")
		detailStart := time.Now()

		var query string
		var args []any

		query = `
		WITH barcode_list AS (
			SELECT
				p.itemcode,
				STRING_AGG(DISTINCT p.barcode, ', ') AS barcodelist
			FROM productbarcode p
			WHERE p.itemcode IS NOT NULL 
			AND p.itemcode <> ''
			{OTHER_CONDITION}
			GROUP BY p.itemcode
		),
		item_names AS (
			SELECT
				p.itemcode,
				STRING_AGG(DISTINCT p.name0, ', ') AS itemname
			FROM productbarcode p
			WHERE p.itemcode IS NOT NULL 
			AND p.itemcode <> ''
			{OTHER_CONDITION}
			GROUP BY p.itemcode
		),
		auto_packing AS (
			SELECT
				p.itemcode,
				COUNT(DISTINCT CASE 
					WHEN p.barcoderefunitstand > 0 AND p.barcoderefunitdivide > 0 
					AND (p.barcoderefunitstand != p.barcoderefunitdivide)
					THEN p.barcoderefunitstand || '-' || p.barcoderefunitdivide 
				END) AS countpacking
			FROM productbarcode p
			WHERE p.itemcode IS NOT NULL 
			AND p.itemcode <> ''
			{OTHER_CONDITION}
			GROUP BY p.itemcode
		)
		SELECT
			b.itemcode,
			b.barcodelist,
			n.itemname,
			MAX(p.unitcode) AS unitcode,
			MAX(p.unitname) AS unitname,
			COALESCE(MAX(ap.countpacking), 0) AS countpacking
		FROM barcode_list b
		JOIN item_names n ON b.itemcode = n.itemcode
		JOIN productbarcode p ON b.itemcode = p.itemcode
		LEFT JOIN auto_packing ap ON b.itemcode = ap.itemcode
		WHERE p.barcoderefunitstand = 1
		AND p.barcoderefunitdivide = 1
		GROUP BY b.itemcode, b.barcodelist, n.itemname
		ORDER BY b.itemcode;
		`

		otherCondition := ""
		if len(itemCodeList) > 0 {
			// มีหลาย itemcode
			// Create placeholders for IN clause
			placeholders := make([]string, len(itemCodeList))
			for i := range itemCodeList {
				placeholders[i] = fmt.Sprintf("$%d", i+1)
			}

			otherCondition = ` AND itemcode IN (` + strings.Join(placeholders, ",") + `)`

			// Convert itemCodeList to any slice
			args = make([]any, len(itemCodeList))
			for i, v := range itemCodeList {
				args[i] = v
			}
		}
		query = strings.ReplaceAll(query, "{OTHER_CONDITION}", otherCondition)

		logger.Info("Query: %s", query)
		logger.Debug("Product detail query preview: %s", mypg.ReplaceQueryParams(query, args...))

		rows, err := mypg.QuerySelectAll(db, query, args...)
		if err != nil {
			logger.Error("querying PostgreSQL: %v", err)
			return
		}
		logger.Info("Product detail query returned %d rows (elapsed=%s)", len(rows), time.Since(detailStart))

		for _, row := range rows {
			var countpacking int64
			if row["countpacking"] != nil {
				// Auto packing
				countpacking = row["countpacking"].(int64)
			}

			itemCodeMainList = append(itemCodeMainList, models.ProductBalanceByCodeStruct{
				ItemCode:      mypg.GetStringValue(row, "itemcode"),
				ItemName:      mypg.GetStringValue(row, "itemname"),
				BarcodeList:   mypg.GetStringValue(row, "barcodelist"),
				UnitCode:      mypg.GetStringValue(row, "unitcode"),
				UnitName:      mypg.GetStringValue(row, "unitname"),
				BalanceQty:    0,
				AverageCost:   0,
				BalanceAmount: 0,
				BalanceWord:   "",
				IsAutoPacking: uint64(countpacking),
				WareHouses:    []models.ProductBalanceByCodeWareHouseStruct{},
			})
		}
	}
	{
		var wareHouseDataList []models.ProductBalanceByCodeWareHouseGetStruct
		var locationDataList []models.ProductBalanceByCodeLocationGetStruct
		if condition == 2 || condition == 3 {
			// 2=ดึงรายการคลังสินค้า
			// ใช้ระบบ timezone ที่รองรับหลายประเทศ
			dateCondition, conditionArgs, err := mypg.BuildDateConditionSQLWithTimeZone("docdatetime", finalDate, timezoneCode)
			if err != nil {
				logger.Error("building date condition for timezone %s: %v", timezoneCode, err)
				return
			}

			transFlagList := myglobal.GetTransFlagsForQuery()

			query := `SELECT 
						lc.itemcode,
						lc.whcode,
						lc.averagecost,
						sb.total_balance AS balanceqty, 
						sb.total_balance * lc.averagecost as balanceamount
					FROM 
						(
							SELECT 
								itemcode,
								whcode,
								averagecost,
								balanceamount,
								ROW_NUMBER() OVER (PARTITION BY itemcode, whcode ORDER BY docdatetime DESC) AS rn
							FROM processstockcost
							WHERE 
								` + dateCondition + `
						) AS lc
					INNER JOIN 
						(
							SELECT 
								itemcode,
								whcode,
								SUM(totalqty * (unitstand / NULLIF(unitdivide, 0))) AS total_balance
							FROM docdetail 
							WHERE 
								` + dateCondition + `
								AND transflag IN (` + transFlagList + `)
							GROUP BY itemcode, whcode
						) AS sb 
					ON lc.itemcode = sb.itemcode AND lc.whcode = sb.whcode
					WHERE lc.rn = 1`

			logger.Info("WareHouse Query: %s", query)
			logger.Debug("WareHouse Query preview: %s", mypg.ReplaceQueryParams(query, conditionArgs...))
			wareHouseQueryStart := time.Now()
			rows, err := mypg.QuerySelectAll(db, query, conditionArgs...)
			if err != nil {
				logger.Error("querying PostgreSQL: %v", err)
				return
			}
			logger.Info("Warehouse cost rows=%d (elapsed=%s)", len(rows), time.Since(wareHouseQueryStart))

			for _, row := range rows {
				itemCode := mypg.GetStringValue(row, "itemcode")
				whcode := mypg.GetStringValue(row, "whcode")
				averagecost := mypg.GetFloat64Value(row, "averagecost")
				balanceqty := mypg.GetFloat64Value(row, "balanceqty")
				balanceamount := mypg.GetFloat64Value(row, "balanceamount")

				found := false
				if len(warehouseList) > 0 {
					for _, wh := range warehouseList {
						if wh.Code == whcode && wh.IsSelected {
							found = true
							break
						}
					}
					if !found {
						continue
					}
				} else {
					// ถ้าไม่มีคลังสินค้าให้ดึงทั้งหมด
					found = true
				}
				if found {
					wareHouseDataList = append(wareHouseDataList, models.ProductBalanceByCodeWareHouseGetStruct{
						ItemCode:      itemCode,
						WareHouseCode: whcode,
						BalanceQty:    balanceqty,
						AverageCost:   averagecost,
						BalanceAmount: balanceamount,
					})
				}
			}
		}
		if condition == 3 {
			// 3=ดึงรายการคลังสินค้าและ Location
			dateCondition2, conditionArgs2, err := mypg.BuildDateConditionSQLWithTimeZone("docdatetime", finalDate, timezoneCode)
			if err != nil {
				logger.Error("building date condition for timezone %s: %v", timezoneCode, err)
				return
			}

			transFlagList := myglobal.GetTransFlagsForQuery()

			query := `SELECT 
						lc.itemcode,
						lc.whcode,
						lc.locationcode,
						sb.total_balance AS balanceqty
					FROM 
						(
							SELECT 
								itemcode,
								whcode,
								locationcode,
								ROW_NUMBER() OVER (PARTITION BY itemcode, whcode, locationcode ORDER BY docdatetime DESC) AS rn
							FROM processstockcost
							WHERE 
								` + dateCondition2 + `
						) AS lc
					INNER JOIN 
						(
							SELECT 
								itemcode,
								whcode,
								locationcode,
								SUM(totalqty * (unitstand / NULLIF(unitdivide, 0))) AS total_balance
							FROM docdetail 
							WHERE 
								` + dateCondition2 + `
								AND transflag IN (` + transFlagList + `)
							GROUP BY itemcode, whcode, locationcode
						) AS sb 
					ON lc.itemcode = sb.itemcode AND lc.whcode = sb.whcode AND lc.locationcode = sb.locationcode
					WHERE lc.rn = 1`

			logger.Info("Query: %s", query)
			logger.Debug("Warehouse+location query preview: %s", mypg.ReplaceQueryParams(query, conditionArgs2...))
			locationQueryStart := time.Now()
			rows, err := mypg.QuerySelectAll(db, query, conditionArgs2...)
			if err != nil {
				logger.Error("querying PostgreSQL: %v", err)
				return
			}
			logger.Info("Warehouse+location rows=%d (elapsed=%s)", len(rows), time.Since(locationQueryStart))

			for _, row := range rows {
				itemcode := mypg.GetStringValue(row, "itemcode")
				whcode := mypg.GetStringValue(row, "whcode")
				locationcode := mypg.GetStringValue(row, "locationcode")
				balanceqty := mypg.GetFloat64Value(row, "balanceqty")

				// ค้นหาคลังสินค้า + location
				found := false
				if len(warehouseList) > 0 {
					for _, wh := range warehouseList {
						if wh.Code == whcode {
							// ค้นหา location
							if len(wh.Locations) > 0 {
								for _, loc := range wh.Locations {
									if loc.Code == locationcode && loc.IsSelected {
										found = true
										break
									}
								}
							} else {
								found = true
							}
							break
						}
					}
					if !found {
						continue
					}
				} else {
					// ถ้าไม่มีคลังสินค้าให้ดึงทั้งหมด
					found = true
				}
				if found {
					locationDataList = append(locationDataList, models.ProductBalanceByCodeLocationGetStruct{
						ItemCode:      itemcode,
						WareHouseCode: whcode,
						LocationCode:  locationcode,
						BalanceQty:    balanceqty,
					})
				}
			}
		}
		// ดึงยอดคงเหลือจาก processstockcost
		logger.Info("ดึงยอดคงเหลือจาก processstockcost")
		dateCondition3, conditionArgs3, err := mypg.BuildDateConditionSQLWithTimeZone("docdatetime", finalDate, timezoneCode)
		if err != nil {
			logger.Error("building date condition for timezone %s: %v", timezoneCode, err)
			return
		}

		queryProduct := `SELECT 
						itemcode,
						averagecost,
						balanceqty,
						balanceamount
					FROM 
						(
							SELECT 
								itemcode,
								averagecost,
								balanceqty,
								balanceamount,
								ROW_NUMBER() OVER (PARTITION BY itemcode ORDER BY docdatetime DESC) AS rn
							FROM processstockcost
							WHERE 
								` + dateCondition3 + `
						) AS lc
					WHERE lc.rn = 1`

		logger.Debug("Final balance query preview: %s", mypg.ReplaceQueryParams(queryProduct, conditionArgs3...))
		finalBalanceStart := time.Now()
		rows, err := mypg.QuerySelectAll(db, queryProduct, conditionArgs3...)
		if err != nil {
			logger.Error("querying PostgreSQL: %v", err)
			return
		}
		logger.Info("Final balance rows=%d (elapsed=%s)", len(rows), time.Since(finalBalanceStart))

		for _, row := range rows {
			itemCode := mypg.GetStringValue(row, "itemcode")
			averagecost := mypg.GetFloat64Value(row, "averagecost")
			balanceqty := mypg.GetFloat64Value(row, "balanceqty")
			balanceamount := mypg.GetFloat64Value(row, "balanceamount")

			index := -1
			for i, datax := range itemCodeMainList {
				if datax.ItemCode == itemCode {
					index = i
					break
				}
			}
			if index != -1 {
				itemCodeMainList[index].AverageCost = averagecost
				itemCodeMainList[index].BalanceQty = balanceqty
				itemCodeMainList[index].BalanceAmount = balanceamount
			}
		}
		if condition == 2 || condition == 3 {
			// 2=ดึงรายการคลังสินค้า
			for productIndex, product := range itemCodeMainList {
				for _, wh := range wareHouseDataList {
					if product.ItemCode == wh.ItemCode {
						// เพิ่มข้อมูลที่เก็บสินค้า
						locationList := []models.ProductBalanceByCodeLocationStruct{}
						if condition == 3 {
							// 3=ดึงรายการคลังสินค้าและ Location
							for _, location := range locationDataList {
								if wh.WareHouseCode == location.WareHouseCode && product.ItemCode == location.ItemCode {
									locationList = append(locationList, models.ProductBalanceByCodeLocationStruct{
										LocationCode: location.LocationCode,
										BalanceQty:   location.BalanceQty,
										BalanceWord:  "",
									})
								}
							}
						}
						// เพิ่มข้อมูลคลังสินค้า
						itemCodeMainList[productIndex].WareHouses = append(itemCodeMainList[productIndex].WareHouses, models.ProductBalanceByCodeWareHouseStruct{
							WareHouseCode: wh.WareHouseCode,
							BalanceQty:    wh.BalanceQty,
							AverageCost:   wh.AverageCost,
							BalanceAmount: wh.BalanceAmount,
							BalanceWord:   "",
							Locations:     locationList,
						})
					}
				}
			}
		}
	}
	if balanceOnly {
		// ลบรายการที่ไม่มียอดคงเหลือ
		filteredList := []models.ProductBalanceByCodeStruct{}
		for _, p := range itemCodeMainList {
			if condition == 1 {
				// 1=ตามสินค้า
				if p.BalanceQty != 0 {
					filteredList = append(filteredList, p)
				}
			} else {
				if p.BalanceQty != 0 && (condition == 2 || condition == 3) {
					// 2=ตามคลัง
					// 3=ตามคลังและตำแหน่ง
					if len(p.WareHouses) > 0 {
						filteredList = append(filteredList, p)
					}
				}
			}
		}
		itemCodeMainList = filteredList
	}

	// เตรียมข้อมูล auto packing ล่วงหน้าเพื่อลดจำนวน query
	uniqueAutoPackingItems := make([]string, 0, len(itemCodeMainList))
	seenAutoPacking := make(map[string]struct{})
	for _, p := range itemCodeMainList {
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
	formatQtyWithUnit := func(qty float64, unitName string) string {
		return fmt.Sprintf("%s %s", myglobal.CalcStockQtyWordCut(qty), unitName)
	}
	logger.Info("Formatting qty words for %d products", len(itemCodeMainList))
	// qty word
	for indexProduct, p := range itemCodeMainList {
		if p.IsAutoPacking > 0 {
			barcodePacking := fetchPackingForItem(p.ItemCode)
			if len(barcodePacking) == 0 {
				itemCodeMainList[indexProduct].BalanceWord = formatQtyWithUnit(p.BalanceQty, p.UnitName)
				for whIndex, wh := range p.WareHouses {
					itemCodeMainList[indexProduct].WareHouses[whIndex].BalanceWord = formatQtyWithUnit(wh.BalanceQty, p.UnitName)
					for locationIndex, location := range wh.Locations {
						itemCodeMainList[indexProduct].WareHouses[whIndex].Locations[locationIndex].BalanceWord = formatQtyWithUnit(location.BalanceQty, p.UnitName)
					}
				}
				continue
			}

			// คำนวณค่า BalanceWord
			itemCodeMainList[indexProduct].BalanceWord = myglobal.CalcStockQtyWord(p.BalanceQty, barcodePacking)
			// qty word ตาม Warehouse
			for whIndex, wh := range p.WareHouses {
				itemCodeMainList[indexProduct].WareHouses[whIndex].BalanceWord = myglobal.CalcStockQtyWord(wh.BalanceQty, barcodePacking)
				// qty word ตาม Location
				for locationIndex, location := range wh.Locations {
					itemCodeMainList[indexProduct].WareHouses[whIndex].Locations[locationIndex].BalanceWord = myglobal.CalcStockQtyWord(location.BalanceQty, barcodePacking)
				}
			}
		} else {
			// ไม่ auto packing - มีแค่หน่วยเดียว
			itemCodeMainList[indexProduct].BalanceWord = formatQtyWithUnit(p.BalanceQty, p.UnitName)
			// qty word ตาม Warehouse
			for whIndex, wh := range p.WareHouses {
				itemCodeMainList[indexProduct].WareHouses[whIndex].BalanceWord = formatQtyWithUnit(wh.BalanceQty, p.UnitName)
				// qty word ตาม Location
				for locationIndex, location := range wh.Locations {
					itemCodeMainList[indexProduct].WareHouses[whIndex].Locations[locationIndex].BalanceWord = formatQtyWithUnit(location.BalanceQty, p.UnitName)
				}
			}
		}
	}

	// Helper to guard against NaN/Inf values that json.Marshal cannot encode
	sanitizeFloat := func(val float64) float64 {
		if math.IsNaN(val) || math.IsInf(val, 0) {
			return 0
		}
		return val
	}

	sanitizeRecord := func(record models.ProductBalanceByCodeStruct) models.ProductBalanceByCodeStruct {
		record.BalanceQty = sanitizeFloat(record.BalanceQty)
		record.AverageCost = sanitizeFloat(record.AverageCost)
		record.BalanceAmount = sanitizeFloat(record.BalanceAmount)
		for whIdx := range record.WareHouses {
			record.WareHouses[whIdx].BalanceQty = sanitizeFloat(record.WareHouses[whIdx].BalanceQty)
			record.WareHouses[whIdx].AverageCost = sanitizeFloat(record.WareHouses[whIdx].AverageCost)
			record.WareHouses[whIdx].BalanceAmount = sanitizeFloat(record.WareHouses[whIdx].BalanceAmount)
			for locIdx := range record.WareHouses[whIdx].Locations {
				record.WareHouses[whIdx].Locations[locIdx].BalanceQty = sanitizeFloat(record.WareHouses[whIdx].Locations[locIdx].BalanceQty)
			}
		}
		return record
	}

	// write to PostgreSQL result table
	guid := myglobal.GenUUID()
	lineNumber := 0
	totalLine := len(itemCodeMainList)
	logger.Info("Prepared %d rows for bulk insert (elapsed=%s)", totalLine, time.Since(overallStart))

	// Use bulk insert for PostgreSQL
	columns := []string{"guid", "docdatetime", "linenumber", "datajson"}
	var records [][]any

	bulkStart := time.Now()
	logger.Info("Starting bulk insert guid=%s", guid)
	for _, b := range itemCodeMainList {
		lineNumber++
		cleanRecord := sanitizeRecord(b)
		jsonData, errMarshal := json.Marshal(cleanRecord)
		if errMarshal != nil {
			logger.Error("marshal balance record item=%s line=%d: %v", b.ItemCode, lineNumber, errMarshal)
			continue
		}
		if !json.Valid(jsonData) {
			logger.Error("invalid JSON payload item=%s line=%d", b.ItemCode, lineNumber)
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
			logger.Info("Inserted batch of %d rows (progress=%d/%d) elapsed=%s", 10000, lineNumber, totalLine, time.Since(bulkStart))
			records = [][]any{}
		}
	}

	// Insert remaining records
	if len(records) > 0 {
		err := mypg.BulkInsertWithCopy(ctx, db, "result", columns, records)
		if err != nil {
			logger.Error("inserting final batch to result table: %v", err)
		}
		logger.Info("Inserted final batch of %d rows (total=%d) bulkElapsed=%s", len(records), totalLine, time.Since(bulkStart))
	}

	logger.Info("Item-WH-Location report stored guid=%s totalLine=%d (total elapsed=%s)", guid, totalLine, time.Since(overallStart))

	return models.ResultModel{
		Success:     true,
		TotalPage:   int(math.Ceil(float64(totalLine) / 10000.0)),
		TotalRecord: totalLine,
		Guid:        guid,
	}
}
