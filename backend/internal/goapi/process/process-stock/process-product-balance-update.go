package processstock

import (
	"database/sql"
	"fmt"
	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/models"
	"smlcloudplatform/internal/goapi/myglobal"
	"smlcloudplatform/internal/goapi/mypg"
	"strings"
	"time"
)

// productQtyInfo — ข้อมูลยอดของสินค้า 1 รายการ
type productQtyInfo struct {
	ItemCode       string
	BalanceQty     float64
	PendingRecvQty float64
	PendingSendQty float64
}

// ==================== Full Rebuild (ทำทั้งหมด) ====================

// ProcessProductBalanceUpdate — คำนวณยอดคงเหลือ + ค้างรับ + ค้างส่ง ทั้งหมด แล้ว UPDATE ลง product
func ProcessProductBalanceUpdate(holdingCode, businessCode string) error {
	holdingCode = strings.TrimSpace(holdingCode)
	businessCode = strings.ToUpper(strings.TrimSpace(businessCode))
	if holdingCode == "" || businessCode == "" {
		return fmt.Errorf("holdingcode and businesscode are required")
	}
	startTime := time.Now()
	logger.Info("ProcessProductBalanceUpdate: start (holdingCode=%s)", holdingCode)

	db, err := mypg.PgSqlFastConnect(holdingCode)
	if err != nil {
		return fmt.Errorf("connect PG: %w", err)
	}

	// 1. Query ยอดคงเหลือ
	balanceMap, err := queryBalanceAll(db, businessCode)
	if err != nil {
		return fmt.Errorf("query balance: %w", err)
	}

	// 2. Query ค้างรับ (PO - received)
	pendingRecvMap, err := queryPendingRecvAll(db, businessCode)
	if err != nil {
		logger.Error("ProcessProductBalanceUpdate: pending recv: %v", err)
		pendingRecvMap = map[string]float64{}
	}

	// 3. Query ค้างส่ง (SO - delivered)
	pendingSendMap, err := queryPendingSendAll(db, businessCode)
	if err != nil {
		logger.Error("ProcessProductBalanceUpdate: pending send: %v", err)
		pendingSendMap = map[string]float64{}
	}

	// 4. รวม itemcodes ทั้งหมด
	allCodes := map[string]bool{}
	for code := range balanceMap {
		allCodes[code] = true
	}
	for code := range pendingRecvMap {
		allCodes[code] = true
	}
	for code := range pendingSendMap {
		allCodes[code] = true
	}

	var items []productQtyInfo
	var itemCodes []string
	for code := range allCodes {
		items = append(items, productQtyInfo{
			ItemCode:       code,
			BalanceQty:     balanceMap[code],
			PendingRecvQty: pendingRecvMap[code],
			PendingSendQty: pendingSendMap[code],
		})
		itemCodes = append(itemCodes, code)
	}

	logger.Info("ProcessProductBalanceUpdate: %d items (balance=%d, pendingRecv=%d, pendingSend=%d)",
		len(items), len(balanceMap), len(pendingRecvMap), len(pendingSendMap))

	// 5. Build packing + unitname
	packingCache := BuildAutoPackingCacheCompany(db, holdingCode, businessCode, itemCodes)
	unitNameMap := buildUnitNameMapCompany(db, holdingCode, businessCode, itemCodes)

	// 6. Batch UPDATE
	updated, err := batchUpdateProductCompany(db, holdingCode, businessCode, items, packingCache, unitNameMap)
	if err != nil {
		return fmt.Errorf("batch update: %w", err)
	}

	// 7. Reset ยอดเป็น 0 สำหรับ product ที่ไม่มีข้อมูล
	zeroed, err := zeroMissingProductCompany(db, holdingCode, businessCode, itemCodes)
	if err != nil {
		logger.Error("ProcessProductBalanceUpdate: zero missing: %v", err)
	}

	elapsed := time.Since(startTime)
	logger.Success("ProcessProductBalanceUpdate: done in %v (updated=%d, zeroed=%d)", elapsed, updated, zeroed)
	return nil
}

// ==================== Partial Update (ทำตามรหัส) ====================

// ProcessProductBalanceUpdateByItems — คำนวณ + UPDATE เฉพาะ itemcodes ที่ระบุ (เรียกจาก Kafka consumer)
func ProcessProductBalanceUpdateByItems(db *sql.DB, holdingCode, businessCode string, itemCodes []string) error {
	holdingCode = strings.TrimSpace(holdingCode)
	businessCode = strings.ToUpper(strings.TrimSpace(businessCode))
	if holdingCode == "" || businessCode == "" {
		return fmt.Errorf("holdingcode and businesscode are required")
	}
	if len(itemCodes) == 0 {
		return nil
	}

	// 1. Query ยอดคงเหลือ
	balanceMap, err := queryBalanceByItems(db, businessCode, itemCodes)
	if err != nil {
		return fmt.Errorf("query balance by items: %w", err)
	}

	// 2. Query ค้างรับ
	pendingRecvMap, err := queryPendingRecvByItems(db, businessCode, itemCodes)
	if err != nil {
		logger.Error("ProcessProductBalanceUpdateByItems: pending recv: %v", err)
		pendingRecvMap = map[string]float64{}
	}

	// 3. Query ค้างส่ง
	pendingSendMap, err := queryPendingSendByItems(db, businessCode, itemCodes)
	if err != nil {
		logger.Error("ProcessProductBalanceUpdateByItems: pending send: %v", err)
		pendingSendMap = map[string]float64{}
	}

	// 4. Build items list
	var items []productQtyInfo
	for _, code := range itemCodes {
		items = append(items, productQtyInfo{
			ItemCode:       code,
			BalanceQty:     balanceMap[code],
			PendingRecvQty: pendingRecvMap[code],
			PendingSendQty: pendingSendMap[code],
		})
	}

	// 5. Build packing + unitname
	packingCache := BuildAutoPackingCacheCompany(db, holdingCode, businessCode, itemCodes)
	unitNameMap := buildUnitNameMapCompany(db, holdingCode, businessCode, itemCodes)

	// 6. Batch UPDATE
	updated, err := batchUpdateProductCompany(db, holdingCode, businessCode, items, packingCache, unitNameMap)
	if err != nil {
		return fmt.Errorf("batch update by items: %w", err)
	}

	logger.Info("ProcessProductBalanceUpdateByItems: updated %d products (%d itemcodes)", updated, len(itemCodes))
	return nil
}

// ProcessProductBalanceUpdateByItemsAsync — เรียกแบบ goroutine (ไม่ block Kafka consumer)
// OnBalanceUpdated — optional hook invoked after product balances are updated.
// Set by the handlers package to invalidate the product search cache without
// an import cycle (process-stock must not import handlers).
var OnBalanceUpdated func()

func ProcessProductBalanceUpdateByItemsAsync(db *sql.DB, holdingCode, businessCode string, itemCodes []string) {
	go func() {
		if err := ProcessProductBalanceUpdateByItems(db, holdingCode, businessCode, itemCodes); err != nil {
			logger.Error("ProcessProductBalanceUpdateByItemsAsync: %v", err)
			return
		}
		if OnBalanceUpdated != nil {
			OnBalanceUpdated()
		}
	}()
}

// ==================== Query: ยอดคงเหลือ ====================

func queryBalanceAll(db *sql.DB, businessCode string) (map[string]float64, error) {
	transFlagList := myglobal.GetTransFlagsForQuery()
	query := fmt.Sprintf(`
		SELECT itemcode, SUM((totalqty * calcflag) * unitstand / NULLIF(unitdivide, 0)) as balance
		FROM docdetail WHERE businesscode = $1 AND transflag IN (%s) GROUP BY itemcode
	`, transFlagList)
	return queryItemQtyMapWithArgs(db, query, []interface{}{businessCode})
}

func queryBalanceByItems(db *sql.DB, businessCode string, itemCodes []string) (map[string]float64, error) {
	transFlagList := myglobal.GetTransFlagsForQuery()
	ph, itemArgs := buildPlaceholdersOffset(itemCodes, 1)
	query := fmt.Sprintf(`
		SELECT itemcode, SUM((totalqty * calcflag) * unitstand / NULLIF(unitdivide, 0)) as balance
		FROM docdetail WHERE businesscode = $1 AND transflag IN (%s) AND itemcode IN (%s) GROUP BY itemcode
	`, transFlagList, ph)
	args := append([]interface{}{businessCode}, itemArgs...)
	return queryItemQtyMapWithArgs(db, query, args)
}

// ==================== Query: ค้างรับ (PO - received) ====================
// ค้างรับ = จำนวนสั่งซื้อใน PO (transflag=6, ยังไม่ปิด) - จำนวนรับแล้ว (transflag=12,310 ที่อ้าง PO)

func queryPendingRecvAll(db *sql.DB, businessCode string) (map[string]float64, error) {
	query := `
		WITH openpo AS (
			SELECT dd.itemcode, dd.docno,
				SUM(dd.totalqty * COALESCE(dd.unitstand,1) / NULLIF(COALESCE(dd.unitdivide,1), 0)) as orderedqty
			FROM docdetail dd
			JOIN doc d ON d.businesscode = dd.businesscode AND d.docno = dd.docno AND d.transflag = dd.transflag
			WHERE dd.businesscode = $1 AND dd.transflag = 6 AND d.isclosed = false
			GROUP BY dd.itemcode, dd.docno
		),
		received AS (
			SELECT dr.docnoref as docno, dd.itemcode,
				SUM(dd.totalqty * COALESCE(dd.unitstand,1) / NULLIF(COALESCE(dd.unitdivide,1), 0)) as receivedqty
			FROM docdetail dd
			JOIN docref dr ON dr.businesscode = dd.businesscode AND dd.docno = dr.docno
			WHERE dd.businesscode = $1 AND dd.transflag IN (12, 310)
			GROUP BY dr.docnoref, dd.itemcode
		)
		SELECT po.itemcode, SUM(po.orderedqty - COALESCE(r.receivedqty, 0)) as pending
		FROM openpo po
		LEFT JOIN received r ON po.docno = r.docno AND po.itemcode = r.itemcode
		GROUP BY po.itemcode
		HAVING SUM(po.orderedqty - COALESCE(r.receivedqty, 0)) > 0
	`
	return queryItemQtyMapWithArgs(db, query, []interface{}{businessCode})
}

func queryPendingRecvByItems(db *sql.DB, businessCode string, itemCodes []string) (map[string]float64, error) {
	ph, args := buildPlaceholdersOffset(itemCodes, 1)
	ph2, args2 := buildPlaceholdersOffset(itemCodes, len(itemCodes)+1)
	query := fmt.Sprintf(`
		WITH openpo AS (
			SELECT dd.itemcode, dd.docno,
				SUM(dd.totalqty * COALESCE(dd.unitstand,1) / NULLIF(COALESCE(dd.unitdivide,1), 0)) as orderedqty
			FROM docdetail dd
			JOIN doc d ON d.businesscode = dd.businesscode AND d.docno = dd.docno AND d.transflag = dd.transflag
			WHERE dd.businesscode = $1 AND dd.transflag = 6 AND d.isclosed = false AND dd.itemcode IN (%s)
			GROUP BY dd.itemcode, dd.docno
		),
		received AS (
			SELECT dr.docnoref as docno, dd.itemcode,
				SUM(dd.totalqty * COALESCE(dd.unitstand,1) / NULLIF(COALESCE(dd.unitdivide,1), 0)) as receivedqty
			FROM docdetail dd
			JOIN docref dr ON dr.businesscode = dd.businesscode AND dd.docno = dr.docno
			WHERE dd.businesscode = $1 AND dd.transflag IN (12, 310) AND dd.itemcode IN (%s)
			GROUP BY dr.docnoref, dd.itemcode
		)
		SELECT po.itemcode, SUM(po.orderedqty - COALESCE(r.receivedqty, 0)) as pending
		FROM openpo po
		LEFT JOIN received r ON po.docno = r.docno AND po.itemcode = r.itemcode
		GROUP BY po.itemcode
		HAVING SUM(po.orderedqty - COALESCE(r.receivedqty, 0)) > 0
	`, ph, ph2)
	// ส่ง args 2 ชุด (สำหรับ openpo + received) — placeholder groups are now distinct
	doubleArgs := append([]interface{}{businessCode}, args...)
	doubleArgs = append(doubleArgs, args2...)
	return queryItemQtyMapWithArgs(db, query, doubleArgs)
}

// ==================== Query: ค้างส่ง (SO - delivered) ====================
// ค้างส่ง = จำนวนสั่งขายใน SO (transflag=36, ยังไม่ปิด) - จำนวนส่งแล้ว (transflag=44 ที่อ้าง SO)

func queryPendingSendAll(db *sql.DB, businessCode string) (map[string]float64, error) {
	query := `
		WITH openso AS (
			SELECT dd.itemcode, dd.docno,
				SUM(dd.totalqty * COALESCE(dd.unitstand,1) / NULLIF(COALESCE(dd.unitdivide,1), 0)) as orderedqty
			FROM docdetail dd
			JOIN doc d ON d.businesscode = dd.businesscode AND d.docno = dd.docno AND d.transflag = dd.transflag
			WHERE dd.businesscode = $1 AND dd.transflag = 36 AND d.isclosed = false
			GROUP BY dd.itemcode, dd.docno
		),
		delivered AS (
			SELECT dr.docnoref as docno, dd.itemcode,
				SUM(dd.totalqty * COALESCE(dd.unitstand,1) / NULLIF(COALESCE(dd.unitdivide,1), 0)) as deliveredqty
			FROM docdetail dd
			JOIN docref dr ON dr.businesscode = dd.businesscode AND dd.docno = dr.docno
			WHERE dd.businesscode = $1 AND dd.transflag = 44
			GROUP BY dr.docnoref, dd.itemcode
		)
		SELECT so.itemcode, SUM(so.orderedqty - COALESCE(d.deliveredqty, 0)) as pending
		FROM openso so
		LEFT JOIN delivered d ON so.docno = d.docno AND so.itemcode = d.itemcode
		GROUP BY so.itemcode
		HAVING SUM(so.orderedqty - COALESCE(d.deliveredqty, 0)) > 0
	`
	return queryItemQtyMapWithArgs(db, query, []interface{}{businessCode})
}

func queryPendingSendByItems(db *sql.DB, businessCode string, itemCodes []string) (map[string]float64, error) {
	ph, args := buildPlaceholdersOffset(itemCodes, 1)
	ph2, args2 := buildPlaceholdersOffset(itemCodes, len(itemCodes)+1)
	query := fmt.Sprintf(`
		WITH openso AS (
			SELECT dd.itemcode, dd.docno,
				SUM(dd.totalqty * COALESCE(dd.unitstand,1) / NULLIF(COALESCE(dd.unitdivide,1), 0)) as orderedqty
			FROM docdetail dd
			JOIN doc d ON d.businesscode = dd.businesscode AND d.docno = dd.docno AND d.transflag = dd.transflag
			WHERE dd.businesscode = $1 AND dd.transflag = 36 AND d.isclosed = false AND dd.itemcode IN (%s)
			GROUP BY dd.itemcode, dd.docno
		),
		delivered AS (
			SELECT dr.docnoref as docno, dd.itemcode,
				SUM(dd.totalqty * COALESCE(dd.unitstand,1) / NULLIF(COALESCE(dd.unitdivide,1), 0)) as deliveredqty
			FROM docdetail dd
			JOIN docref dr ON dr.businesscode = dd.businesscode AND dd.docno = dr.docno
			WHERE dd.businesscode = $1 AND dd.transflag = 44 AND dd.itemcode IN (%s)
			GROUP BY dr.docnoref, dd.itemcode
		)
		SELECT so.itemcode, SUM(so.orderedqty - COALESCE(d.deliveredqty, 0)) as pending
		FROM openso so
		LEFT JOIN delivered d ON so.docno = d.docno AND so.itemcode = d.itemcode
		GROUP BY so.itemcode
		HAVING SUM(so.orderedqty - COALESCE(d.deliveredqty, 0)) > 0
	`, ph, ph2)
	doubleArgs := append([]interface{}{businessCode}, args...)
	doubleArgs = append(doubleArgs, args2...)
	return queryItemQtyMapWithArgs(db, query, doubleArgs)
}

// ==================== Batch UPDATE ====================

// batchUpdateProduct — UPDATE product ทั้ง 6 fields (balance + pending recv + pending send)
func batchUpdateProduct(db *sql.DB, items []productQtyInfo, packingCache map[string][]models.ProductBarcodePackingStruct, unitNameMap map[string]string) (int, error) {
	return batchUpdateProductScoped(db, "", "", items, packingCache, unitNameMap)
}

func batchUpdateProductCompany(db *sql.DB, holdingCode, businessCode string, items []productQtyInfo, packingCache map[string][]models.ProductBarcodePackingStruct, unitNameMap map[string]string) (int, error) {
	return batchUpdateProductScoped(db, holdingCode, businessCode, items, packingCache, unitNameMap)
}

func batchUpdateProductScoped(db *sql.DB, holdingCode, businessCode string, items []productQtyInfo, packingCache map[string][]models.ProductBarcodePackingStruct, unitNameMap map[string]string) (int, error) {
	if len(items) == 0 {
		return 0, nil
	}

	updated := 0
	batchSize := 100

	for i := 0; i < len(items); i += batchSize {
		end := i + batchSize
		if end > len(items) {
			end = len(items)
		}
		batch := items[i:end]

		balQtyClauses := make([]string, 0, len(batch))
		balWordClauses := make([]string, 0, len(batch))
		recvQtyClauses := make([]string, 0, len(batch))
		recvWordClauses := make([]string, 0, len(batch))
		sendQtyClauses := make([]string, 0, len(batch))
		sendWordClauses := make([]string, 0, len(batch))
		inCodes := make([]string, 0, len(batch))
		args := make([]interface{}, 0, len(batch)*7)
		argIdx := 1

		for _, item := range batch {
			balWord := formatBalanceWord(item.BalanceQty, item.ItemCode, packingCache, unitNameMap)
			recvWord := formatBalanceWord(item.PendingRecvQty, item.ItemCode, packingCache, unitNameMap)
			sendWord := formatBalanceWord(item.PendingSendQty, item.ItemCode, packingCache, unitNameMap)

			balQtyClauses = append(balQtyClauses, fmt.Sprintf("WHEN itemcode = $%d THEN $%d::numeric", argIdx, argIdx+1))
			balWordClauses = append(balWordClauses, fmt.Sprintf("WHEN itemcode = $%d THEN $%d", argIdx, argIdx+2))
			recvQtyClauses = append(recvQtyClauses, fmt.Sprintf("WHEN itemcode = $%d THEN $%d::numeric", argIdx, argIdx+3))
			recvWordClauses = append(recvWordClauses, fmt.Sprintf("WHEN itemcode = $%d THEN $%d", argIdx, argIdx+4))
			sendQtyClauses = append(sendQtyClauses, fmt.Sprintf("WHEN itemcode = $%d THEN $%d::numeric", argIdx, argIdx+5))
			sendWordClauses = append(sendWordClauses, fmt.Sprintf("WHEN itemcode = $%d THEN $%d", argIdx, argIdx+6))
			inCodes = append(inCodes, fmt.Sprintf("$%d", argIdx))
			args = append(args, item.ItemCode, item.BalanceQty, balWord, item.PendingRecvQty, recvWord, item.PendingSendQty, sendWord)
			argIdx += 7
		}

		companyWhere := ""
		if businessCode != "" {
			companyWhere = fmt.Sprintf("holding_code = $%d AND businesscode = $%d AND ", argIdx, argIdx+1)
			args = append(args, holdingCode, businessCode)
		}
		query := fmt.Sprintf(`
			UPDATE product SET
				balanceqty = CASE %s END,
				balanceqtyword = CASE %s END,
				pendingrecvqty = CASE %s END,
				pendingrecvqtyword = CASE %s END,
				pendingsendqty = CASE %s END,
				pendingsendqtyword = CASE %s END
			WHERE %sitemcode IN (%s)
		`,
			strings.Join(balQtyClauses, " "),
			strings.Join(balWordClauses, " "),
			strings.Join(recvQtyClauses, " "),
			strings.Join(recvWordClauses, " "),
			strings.Join(sendQtyClauses, " "),
			strings.Join(sendWordClauses, " "),
			companyWhere,
			strings.Join(inCodes, ","),
		)

		result, err := db.Exec(query, args...)
		if err != nil {
			logger.Error("batchUpdateProduct: batch %d: %v", i/batchSize, err)
			continue
		}
		affected, _ := result.RowsAffected()
		updated += int(affected)
	}

	return updated, nil
}

// zeroMissingProduct — reset ยอดทั้ง 6 fields เป็น 0
func zeroMissingProduct(db *sql.DB, activeItemCodes []string) (int, error) {
	zeroFields := "balanceqty = 0, balanceqtyword = '', pendingrecvqty = 0, pendingrecvqtyword = '', pendingsendqty = 0, pendingsendqtyword = ''"
	hasData := "(balanceqty != 0 OR balanceqtyword != '' OR pendingrecvqty != 0 OR pendingrecvqtyword != '' OR pendingsendqty != 0 OR pendingsendqtyword != '')"

	if len(activeItemCodes) == 0 {
		result, err := db.Exec(fmt.Sprintf("UPDATE product SET %s WHERE %s", zeroFields, hasData))
		if err != nil {
			return 0, err
		}
		affected, _ := result.RowsAffected()
		return int(affected), nil
	}

	ph, args := buildPlaceholders(activeItemCodes)
	query := fmt.Sprintf("UPDATE product SET %s WHERE %s AND itemcode NOT IN (%s)", zeroFields, hasData, ph)
	result, err := db.Exec(query, args...)
	if err != nil {
		return 0, err
	}
	affected, _ := result.RowsAffected()
	return int(affected), nil
}

func zeroMissingProductCompany(db *sql.DB, holdingCode, businessCode string, activeItemCodes []string) (int, error) {
	zeroFields := "balanceqty = 0, balanceqtyword = '', pendingrecvqty = 0, pendingrecvqtyword = '', pendingsendqty = 0, pendingsendqtyword = ''"
	hasData := "(balanceqty != 0 OR balanceqtyword != '' OR pendingrecvqty != 0 OR pendingrecvqtyword != '' OR pendingsendqty != 0 OR pendingsendqtyword != '')"
	args := []interface{}{holdingCode, businessCode}
	if len(activeItemCodes) == 0 {
		result, err := db.Exec(fmt.Sprintf("UPDATE product SET %s WHERE holding_code = $1 AND businesscode = $2 AND %s", zeroFields, hasData), args...)
		if err != nil {
			return 0, err
		}
		affected, _ := result.RowsAffected()
		return int(affected), nil
	}

	ph, itemArgs := buildPlaceholdersOffset(activeItemCodes, 2)
	args = append(args, itemArgs...)
	query := fmt.Sprintf("UPDATE product SET %s WHERE holding_code = $1 AND businesscode = $2 AND %s AND itemcode NOT IN (%s)", zeroFields, hasData, ph)
	result, err := db.Exec(query, args...)
	if err != nil {
		return 0, err
	}
	affected, _ := result.RowsAffected()
	return int(affected), nil
}

// ==================== Helpers ====================

func buildPlaceholders(items []string) (string, []interface{}) {
	return buildPlaceholdersOffset(items, 0)
}

// buildPlaceholdersOffset — placeholder group with distinct indices ($offset+1..$offset+N).
// Two-CTE queries need two DISTINCT groups: pq treats each $N as one parameter,
// so reusing $1..$N in both CTEs while passing 2N args fails with
// "got 2 parameters but the statement requires 1".
func buildPlaceholdersOffset(items []string, offset int) (string, []interface{}) {
	placeholders := make([]string, len(items))
	args := make([]interface{}, len(items))
	for i, item := range items {
		placeholders[i] = fmt.Sprintf("$%d", offset+i+1)
		args[i] = item
	}
	return strings.Join(placeholders, ","), args
}

func queryItemQtyMap(db *sql.DB, query string) (map[string]float64, error) {
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := map[string]float64{}
	for rows.Next() {
		var code string
		var qty float64
		if err := rows.Scan(&code, &qty); err != nil {
			continue
		}
		result[code] = qty
	}
	return result, rows.Err()
}

func queryItemQtyMapWithArgs(db *sql.DB, query string, args []interface{}) (map[string]float64, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := map[string]float64{}
	for rows.Next() {
		var code string
		var qty float64
		if err := rows.Scan(&code, &qty); err != nil {
			continue
		}
		result[code] = qty
	}
	return result, rows.Err()
}

func buildUnitNameMap(db *sql.DB, itemCodes []string) map[string]string {
	return buildUnitNameMapScoped(db, "", "", itemCodes)
}

func buildUnitNameMapCompany(db *sql.DB, holdingCode, businessCode string, itemCodes []string) map[string]string {
	return buildUnitNameMapScoped(db, holdingCode, businessCode, itemCodes)
}

func buildUnitNameMapScoped(db *sql.DB, holdingCode, businessCode string, itemCodes []string) map[string]string {
	result := map[string]string{}
	if len(itemCodes) == 0 {
		return result
	}
	batchSize := 500
	for i := 0; i < len(itemCodes); i += batchSize {
		end := i + batchSize
		if end > len(itemCodes) {
			end = len(itemCodes)
		}
		offset := 0
		args := []interface{}{}
		companyWhere := ""
		if businessCode != "" {
			offset = 2
			args = append(args, holdingCode, businessCode)
			companyWhere = "holding_code = $1 AND businesscode = $2 AND "
		}
		ph, itemArgs := buildPlaceholdersOffset(itemCodes[i:end], offset)
		args = append(args, itemArgs...)
		query := fmt.Sprintf("SELECT itemcode, COALESCE(unitname,'') FROM product WHERE %sitemcode IN (%s)", companyWhere, ph)
		rows, err := db.Query(query, args...)
		if err != nil {
			logger.Error("buildUnitNameMap: %v", err)
			continue
		}
		for rows.Next() {
			var code, name string
			if err := rows.Scan(&code, &name); err == nil {
				result[code] = name
			}
		}
		rows.Close()
	}
	return result
}

func formatBalanceWord(qty float64, itemCode string, packingCache map[string][]models.ProductBarcodePackingStruct, unitNameMap map[string]string) string {
	if qty == 0 {
		return ""
	}
	packUnits := packingCache[itemCode]
	if len(packUnits) > 0 {
		return myglobal.CalcStockQtyWord(qty, packUnits)
	}
	unitName := unitNameMap[itemCode]
	return fmt.Sprintf("%s %s", myglobal.CalcStockQtyWordCut(qty), unitName)
}
