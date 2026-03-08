package processdoc

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/models"
	"strings"

	"smlcloudplatform/internal/goapi/mypg"
)

func PurchaseStatusByDocNo(shopID string, docNos []string) []models.PurchaseStatusStruct {
	var statusList []models.PurchaseStatusStruct

	if len(docNos) == 0 {
		logger.Info("No document numbers provided")
		return statusList
	}

	ctx := context.Background()
	db, err := mypg.PgSqlFastConnect(shopID)
	if err != nil {
		logger.Info("Failed to connect to PostgreSQL: %v", err)
		return statusList
	}

	// สร้าง placeholder สำหรับ IN clause ($1, $2, $3, ...)
	placeholders := make([]string, len(docNos))
	args := make([]interface{}, len(docNos))
	for i, docNo := range docNos {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = docNo
	}
	inClause := strings.Join(placeholders, ",")

	// Query หลักสำหรับ status ทั้งหมด
	mainQuery := fmt.Sprintf(`
        SELECT docno, isclosed, iscomparedsuccess
        FROM doc
        WHERE docno IN (%s) AND transflag = 6
    `, inClause)

	logger.Info("Main Query for pgAdmin: SELECT docno, isclosed, iscomparedsuccess FROM doc WHERE docno IN ('%s') AND transflag = 6", strings.Join(docNos, "','"))

	rows, err := db.QueryContext(ctx, mainQuery, args...)
	if err != nil {
		logger.Info("Failed to retrieve purchase status: %v", err)
		return statusList
	}
	defer rows.Close()

	// เก็บ status ของแต่ละ document
	statusMap := make(map[string]models.PurchaseStatusStruct)
	for rows.Next() {
		var status models.PurchaseStatusStruct
		var docNo string
		if err := rows.Scan(&docNo, &status.IsClosed, &status.IsComparedSuccess); err != nil {
			logger.Info("Failed to scan purchase status: %v", err)
			continue
		}
		status.DocNo = docNo
		statusMap[docNo] = status
	}

	// Query 1: รายการสั่งซื้อทั้งหมด
	orderedQuery := fmt.Sprintf(`
		SELECT 
			docno,
			itemcode,
			transflag,
			SUM(COALESCE(totalqty * (COALESCE(unitstand, 1) / NULLIF(COALESCE(unitdivide, 1), 0)), totalqty)) AS qty_ordered
		FROM docdetail
		WHERE docno IN (%s) AND transflag = 6
		GROUP BY docno, itemcode, transflag
		ORDER BY docno, itemcode
	`, inClause)

	logger.Info("Ordered Query: %s", orderedQuery)

	orderedRows, err := db.QueryContext(ctx, orderedQuery, args...)
	if err != nil {
		logger.Info("Failed to retrieve ordered items: %v", err)
		return statusList
	}
	defer orderedRows.Close()

	// เก็บข้อมูลสั่งซื้อ
	type OrderedItem struct {
		DocNo      string
		ItemCode   string
		TransFlag  int
		QtyOrdered float64
	}
	orderedItems := make(map[string]OrderedItem) // key = docno+itemcode

	for orderedRows.Next() {
		var item OrderedItem
		if err := orderedRows.Scan(&item.DocNo, &item.ItemCode, &item.TransFlag, &item.QtyOrdered); err != nil {
			logger.Info("Failed to scan ordered item: %v", err)
			continue
		}
		key := item.DocNo + "|" + item.ItemCode
		orderedItems[key] = item
		logger.Info("ORDERED: %s-%s transflag=%d qty=%.2f", item.DocNo, item.ItemCode, item.TransFlag, item.QtyOrdered)
	}

	// Query 2: รายการรับสินค้าทั้งหมด (summary)
	receivedSummaryQuery := fmt.Sprintf(`
		SELECT 
			dr.docnoref,
			dd.itemcode,
			SUM(COALESCE(dd.totalqty * (COALESCE(dd.unitstand, 1) / NULLIF(COALESCE(dd.unitdivide, 1), 0)), dd.totalqty)) AS qty_received
		FROM docdetail dd
		INNER JOIN (SELECT DISTINCT docno, docnoref FROM docref) dr ON dd.docno = dr.docno
		WHERE dd.transflag IN (310,12) AND dr.docnoref IN (%s)
		GROUP BY dr.docnoref, dd.itemcode
		ORDER BY dr.docnoref, dd.itemcode
	`, inClause)

	logger.Info("Received Summary Query: %s", receivedSummaryQuery)

	receivedSummaryRows, err := db.QueryContext(ctx, receivedSummaryQuery, args...)
	if err != nil {
		logger.Info("Failed to retrieve received summary: %v", err)
		return statusList
	}
	defer receivedSummaryRows.Close()

	// เก็บข้อมูลรับสินค้า (รวม)
	type ReceivedSummary struct {
		DocNoRef    string
		ItemCode    string
		QtyReceived float64
	}
	receivedSummaries := make(map[string]ReceivedSummary) // key = docnoref+itemcode

	for receivedSummaryRows.Next() {
		var summary ReceivedSummary
		if err := receivedSummaryRows.Scan(&summary.DocNoRef, &summary.ItemCode, &summary.QtyReceived); err != nil {
			logger.Info("Failed to scan received summary: %v", err)
			continue
		}
		key := summary.DocNoRef + "|" + summary.ItemCode
		receivedSummaries[key] = summary
		logger.Info("RECEIVED SUMMARY: %s-%s qty=%.2f", summary.DocNoRef, summary.ItemCode, summary.QtyReceived)
	}

	// Query 3: รายการรับสินค้ารายละเอียด (สำหรับ doc_refer)
	receivedDetailQuery := fmt.Sprintf(`
		SELECT 
			dr.docnoref,
			dd.itemcode,
			dd.docno,
			dd.transflag,
			SUM(COALESCE(dd.totalqty * (COALESCE(dd.unitstand, 1) / NULLIF(COALESCE(dd.unitdivide, 1), 0)), dd.totalqty)) AS qty_received_detail
		FROM docdetail dd
		INNER JOIN (SELECT DISTINCT docno, docnoref FROM docref) dr ON dd.docno = dr.docno
		WHERE dd.transflag IN (310,12) AND dr.docnoref IN (%s)
		GROUP BY dr.docnoref, dd.itemcode, dd.docno, dd.transflag
		ORDER BY dr.docnoref, dd.itemcode, dd.docno
	`, inClause)

	logger.Info("Received Detail Query: %s", receivedDetailQuery)

	receivedDetailRows, err := db.QueryContext(ctx, receivedDetailQuery, args...)
	if err != nil {
		logger.Info("Failed to retrieve received details: %v", err)
		return statusList
	}
	defer receivedDetailRows.Close()

	// เก็บข้อมูลรับสินค้ารายละเอียด
	type ReceivedDetail struct {
		DocNoRef          string
		ItemCode          string
		DocNo             string
		TransFlag         int
		QtyReceivedDetail float64
	}
	receivedDetails := make(map[string][]ReceivedDetail) // key = docnoref+itemcode

	for receivedDetailRows.Next() {
		var detail ReceivedDetail
		if err := receivedDetailRows.Scan(&detail.DocNoRef, &detail.ItemCode, &detail.DocNo, &detail.TransFlag, &detail.QtyReceivedDetail); err != nil {
			logger.Info("Failed to scan received detail: %v", err)
			continue
		}
		key := detail.DocNoRef + "|" + detail.ItemCode
		receivedDetails[key] = append(receivedDetails[key], detail)
		logger.Info("RECEIVED DETAIL: %s-%s from %s transflag=%d qty=%.2f", detail.DocNoRef, detail.ItemCode, detail.DocNo, detail.TransFlag, detail.QtyReceivedDetail)
	}

	// ประกอบข้อมูลจาก 3 query
	logger.Info("=== COMBINING DATA ===")
	detailsMap := make(map[string][]models.PurchaseStatusDocDetailStruct)

	// รวบรวม itemcode ทั้งหมดเพื่อ query productbarcode ครั้งเดียว
	itemCodes := make([]string, 0)
	for _, ordered := range orderedItems {
		itemCodes = append(itemCodes, ordered.ItemCode)
	}

	// Query productbarcode ทั้งหมดครั้งเดียว
	productMap := make(map[string]map[string]string) // map[itemcode]map[field]value
	if len(itemCodes) > 0 {
		// สร้าง placeholder สำหรับ IN clause
		productPlaceholders := make([]string, len(itemCodes))
		productArgs := make([]any, len(itemCodes))
		for i, itemCode := range itemCodes {
			productPlaceholders[i] = fmt.Sprintf("$%d", i+1)
			productArgs[i] = itemCode
		}
		productInClause := strings.Join(productPlaceholders, ",")

		productQuery := fmt.Sprintf("SELECT itemcode, name0, unitcode, unitname FROM product WHERE itemcode IN (%s)", productInClause)
		logger.Info("Product Query: %s", productQuery)
		logger.Info("Product Args: %v", productArgs)

		productRows, err := db.QueryContext(ctx, productQuery, productArgs...)
		if err != nil {
			logger.Info("Failed to query product: %v", err)
		} else {
			defer productRows.Close()
			rowCount := 0
			for productRows.Next() {
				rowCount++
				var itemCode string
				var productName, unitCode, unitName sql.NullString
				if err := productRows.Scan(&itemCode, &productName, &unitCode, &unitName); err != nil {
					logger.Info("Failed to scan product: %v", err)
					continue
				}

				productInfo := make(map[string]string)
				if productName.Valid {
					productInfo["name"] = productName.String
				}
				if unitCode.Valid {
					productInfo["unit_code"] = unitCode.String
				}
				if unitName.Valid {
					productInfo["unit_name"] = unitName.String
				}

				productMap[itemCode] = productInfo
				logger.Info("PRODUCT: %s name=%s unitcode=%s unitname=%s",
					itemCode, productInfo["name"], productInfo["unit_code"], productInfo["unit_name"])
			}
			logger.Info("Found %d products in product table", rowCount)

			// ตรวจสอบว่า item codes ที่หาไม่เจอมีอะไรบ้าง
			for _, itemCode := range itemCodes {
				if _, exists := productMap[itemCode]; !exists {
					logger.Info("MISSING PRODUCT: %s not found in product table", itemCode)
				}
			}
		}
	}

	// วนลูปข้อมูลสั่งซื้อ
	for key, ordered := range orderedItems {
		var detail models.PurchaseStatusDocDetailStruct
		detail.DocNo = ordered.DocNo
		detail.ItemCode = ordered.ItemCode
		detail.TransFlag = ordered.TransFlag
		detail.TotalQty = ordered.QtyOrdered

		// หาข้อมูลรับสินค้า (summary)
		if received, exists := receivedSummaries[key]; exists {
			detail.ReceivedQty = received.QtyReceived
		} else {
			detail.ReceivedQty = 0
		}

		// คำนวณ pending
		detail.PendingQty = detail.TotalQty - detail.ReceivedQty

		// หาข้อมูลสินค้าจาก productMap
		if productInfo, exists := productMap[detail.ItemCode]; exists {
			detail.ProductName = productInfo["name"]
			detail.UnitCode = productInfo["unit_code"]
			detail.UnitName = productInfo["unit_name"]
		}

		// สร้าง doc_refer จากข้อมูลรายละเอียด
		detail.DocRefer = []models.PurchaseStatusDocDetailReferStruct{}
		if details, exists := receivedDetails[key]; exists {
			for _, rd := range details {
				refStruct := models.PurchaseStatusDocDetailReferStruct{
					DocNo:       rd.DocNo,
					TransFlag:   rd.TransFlag,
					ReceivedQty: rd.QtyReceivedDetail,
				}
				detail.DocRefer = append(detail.DocRefer, refStruct)
			}
		}

		logger.Info("COMBINED: %s-%s Ordered=%.2f Received=%.2f Pending=%.2f DocRefs=%d",
			detail.DocNo, detail.ItemCode, detail.TotalQty, detail.ReceivedQty, detail.PendingQty, len(detail.DocRefer))

		detailsMap[ordered.DocNo] = append(detailsMap[ordered.DocNo], detail)
	}

	// รวม status กับ details
	for _, docNo := range docNos {
		if status, exists := statusMap[docNo]; exists {
			status.Details = detailsMap[docNo]
			statusList = append(statusList, status)
			logger.Info("Purchase status for %s: IsClosed=%v, IsComparedSuccess=%d, Items=%d",
				docNo, status.IsClosed, status.IsComparedSuccess, len(status.Details))
		} else {
			logger.Info("No purchase document found for %s", docNo)
			// เพิ่ม empty status สำหรับ docNo ที่ไม่พบ
			emptyStatus := models.PurchaseStatusStruct{
				DocNo:             docNo,
				IsClosed:          false,
				IsComparedSuccess: 0,
				Details:           []models.PurchaseStatusDocDetailStruct{},
			}
			statusList = append(statusList, emptyStatus)
		}
	}

	logger.Info("Processed %d purchase documents for shop %s", len(statusList), shopID)
	// log json
	jsonData, _ := json.MarshalIndent(statusList, "", "  ")
	logger.Info("Purchase Status JSON: %s", string(jsonData))
	return statusList
}
