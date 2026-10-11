package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"smlcloudplatform/internal/doctype"
	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/mypg"

	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
)

// SalesReportRequest - request สำหรับรายงานขาย
type SalesReportRequest struct {
	HoldingCode   string   `json:"holdingcode"`
	FromDate      string   `json:"fromdate"`
	ToDate        string   `json:"todate"`
	BranchCodes   []string `json:"branchcodes,omitempty"`
	ProductCodes  []string `json:"productcodes,omitempty"`
	SortAscending bool     `json:"sortascending"`
	ReportType    string   `json:"reporttype"` // "header" or "detail"
	Limit         int      `json:"limit,omitempty"`
	Offset        int      `json:"offset,omitempty"`
}

// SalesReportByDocumentHandler - รายงานขายแยกตามเอกสาร (ปลอดภัยจาก SQL injection)
func SalesReportByDocumentHandler(c echo.Context) error {
	var req SalesReportRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload",
			"code":  "INVALID_PAYLOAD",
		})
	}

	// Validate required fields
	if req.HoldingCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "holdingcode is required",
			"code":  "MISSING_HOLDING_CODE",
		})
	}

	if req.FromDate == "" || req.ToDate == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "from_date and to_date are required",
			"code":  "MISSING_DATE_RANGE",
		})
	}

	// Parse dates
	fromDate, err := time.Parse("2006-01-02", req.FromDate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid from_date format. Use YYYY-MM-DD",
			"code":  "INVALID_FROM_DATE",
		})
	}

	toDate, err := time.Parse("2006-01-02", req.ToDate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid to_date format. Use YYYY-MM-DD",
			"code":  "INVALID_TO_DATE",
		})
	}

	// Add 1 day to toDate to include the entire last day
	toDate = toDate.AddDate(0, 0, 1)

	// Set default limits
	if req.Limit <= 0 || req.Limit > 10000 {
		req.Limit = 1000
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	businessCode, scopeErr := reportCompanyScope(c, req.HoldingCode)
	if scopeErr != nil {
		return taxScopeFail(c, scopeErr)
	}

	// Connect to database
	db, err := mypg.PgSqlFastConnect(req.HoldingCode)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Database connection failed",
			"code":  "DB_CONNECTION_ERROR",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	var query string
	var args []any

	// Build appropriate query based on report type
	if req.ReportType == "detail" {
		query, args = buildSalesReportDetailQuery(fromDate, toDate, businessCode, req.BranchCodes, req.ProductCodes, req.SortAscending, req.Limit, req.Offset)
	} else {
		query, args = buildSalesReportHeaderQuery(fromDate, toDate, businessCode, req.BranchCodes, req.ProductCodes, req.SortAscending, req.Limit, req.Offset)
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		logger.Error("SalesReport: query failed for %s/%s: %v", req.HoldingCode, businessCode, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Query execution failed",
			"code":  "QUERY_ERROR",
		})
	}
	defer rows.Close()

	results, err := scanSalesReportRows(rows)
	if err != nil {
		logger.Error("SalesReport: read rows failed for %s/%s: %v", req.HoldingCode, businessCode, err)
		return taxReportFail(c, http.StatusInternalServerError, "QUERY_ERROR", "load_failed")
	}

	return c.JSON(http.StatusOK, map[string]any{
		"status":      "success",
		"data":        results,
		"count":       len(results),
		"limit":       req.Limit,
		"offset":      req.Offset,
		"report_type": req.ReportType,
	})
}

// scanSalesReportRows - อ่านทุกแถวของ query รายงานขายเป็น map ตามชื่อคอลัมน์ แล้วแปลงคอลัมน์เงินด้วย formatSalesReportRow
// แถวที่อ่าน/แปลงไม่ได้ = รายงานเงินไม่ครบ → คืนข้อผิดพลาด ห้ามทิ้งแถวเงียบ ๆ
func scanSalesReportRows(rows *sql.Rows) ([]map[string]any, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	results := make([]map[string]any, 0)
	for rows.Next() {
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}
		row := make(map[string]any, len(columns))
		for i, col := range columns {
			if b, ok := values[i].([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = values[i]
			}
		}
		if err := formatSalesReportRow(row); err != nil {
			return nil, err
		}
		results = append(results, row)
	}
	return results, rows.Err()
}

// buildSalesReportHeaderQuery - สร้าง query สำหรับรายงานขายแบบสรุป (ป้องกัน SQL injection)
func buildSalesReportHeaderQuery(fromDate, toDate time.Time, businessCode string, branchCodes, productCodes []string, sortAscending bool, limit, offset int) (string, []any) {
	args := make([]any, 0)
	argIndex := 1

	// จำนวนและมูลค่าในสมุดสต็อกเป็นบวกเสมอ ทิศทางอยู่ที่คอลัมน์ direction
	// รายงานขายจึงไม่ต้องคูณ -1 เหมือนตารางเดิมที่เก็บยอดขายเป็นค่าลบ
	//
	// ลบเอกสาร = soft delete (doc.isdelete) แล้ว stock engine ค่อยถอนบรรทัดออกจาก
	// stock_ledger ทีหลัง ระหว่างรอ worker ยอดจึงยังค้างใน ledger — กรองด้วย doc.isdelete
	// อีกชั้นเพื่อไม่ให้รายงานนับใบที่ลบไปแล้ว (ใช้กับทั้งสาม query ในไฟล์นี้)
	//
	// ชื่อลูกค้า = doc.custname ที่บันทึกไว้บนเอกสารของบริษัทเดียวกัน (businesscode + transflag + docno)
	// ไม่ join ตาราง debtor เดิม: ฐานปัจจุบันไม่มีตารางนั้น (query ล้มทั้งรายงาน) และ join ด้วยรหัสลูกค้าอย่างเดียวไม่มีบริษัท
	// ทำให้รหัสซ้ำข้ามบริษัทแตกแถวเอกสารและเห็นชื่อลูกค้าของอีกบริษัท
	// ไม่ส่งราคาเฉลี่ย/ต้นทุนเฉลี่ยต่อใบ (AVG ได้ทศนิยมยาวไม่ใช่ค่าที่บันทึก และไม่มีจอใดใช้)
	query := `
SELECT
  CAST(MAX(p.docdatetime) + INTERVAL '7 hour' AS DATE) as docdate,
  TO_CHAR(MAX(p.docdatetime) + INTERVAL '7 hour', 'HH24:MI:SS') as doctime,
  MAX(p.docdatetime) + INTERVAL '7 hour' as docdatetime,
  p.docno,
  COALESCE(doc.custcode, '') as debtorcode,
  COALESCE(doc.custname, '') as debtorname,
  SUM(p.docvalue) as totalamount,
  SUM(p.amount) as calcamount,
  (SUM(p.docvalue) - SUM(p.amount)) as grossprofit
FROM public.stock_ledger p
LEFT JOIN public.doc doc ON doc.businesscode = p.businesscode AND doc.transflag = p.transflag AND doc.docno = p.docno
WHERE p.transflag = ` + strconv.Itoa(doctype.SaleInvoice) + `
  AND p.businesscode = $3
  AND p.docdatetime >= $1 AND p.docdatetime < $2
  AND NOT EXISTS (
    SELECT 1 FROM public.doc dd
    WHERE dd.businesscode = p.businesscode AND dd.docno = p.docno
      AND dd.transflag = p.transflag AND dd.isdelete
  )`

	args = append(args, fromDate, toDate, businessCode)
	argIndex = 4

	// Add branch filter (parameterized)
	if len(branchCodes) > 0 {
		query += fmt.Sprintf(" AND p.whcode = ANY($%d)", argIndex)
		args = append(args, pq.Array(branchCodes))
		argIndex++
	}

	// Add product filter via subquery (parameterized)
	if len(productCodes) > 0 {
		query += fmt.Sprintf(`
  AND p.docno IN (
    SELECT DISTINCT sub.docno
    FROM public.stock_ledger sub
    WHERE sub.transflag = %d
      AND sub.businesscode = $3
      AND sub.docdatetime >= $1 AND sub.docdatetime < $2
      AND sub.itemcode = ANY($%d)`, doctype.SaleInvoice, argIndex)
		args = append(args, pq.Array(productCodes))
		argIndex++

		// Add branch filter to subquery if present
		if len(branchCodes) > 0 {
			query += fmt.Sprintf(" AND sub.whcode = ANY($%d)", argIndex-2) // reuse branch param
		}
		query += ")"
	}

	query += "\nGROUP BY p.docno, doc.custcode, doc.custname"

	if sortAscending {
		query += "\nORDER BY MAX(p.docdatetime) ASC, p.docno"
	} else {
		query += "\nORDER BY MAX(p.docdatetime) DESC, p.docno"
	}

	query += fmt.Sprintf("\nLIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	return query, args
}

// buildSalesReportDetailQuery - สร้าง query สำหรับรายงานขายแบบรายละเอียด (ป้องกัน SQL injection)
func buildSalesReportDetailQuery(fromDate, toDate time.Time, businessCode string, branchCodes, productCodes []string, sortAscending bool, limit, offset int) (string, []any) {
	args := make([]any, 0)
	argIndex := 1

	// ชื่อสินค้า = คำอธิบายบรรทัดเอกสารของบริษัทเดียวกัน (ฐานปัจจุบันไม่มีตาราง productbarcode เดิม), หน่วย = รหัสหน่วยที่บันทึกในสมุดสต็อก
	query := `
SELECT
  CAST(p.docdatetime + INTERVAL '7 hour' AS DATE) as docdate,
  TO_CHAR(p.docdatetime + INTERVAL '7 hour', 'HH24:MI:SS') as doctime,
  p.docdatetime + INTERVAL '7 hour' as docdatetime,
  p.docno,
  p.linenumber,
  p.itemcode,
  COALESCE(item_lookup.itemname, '') as itemname,
  p.barcode,
  p.unitcode as unitname,
  p.docqty as totalqty,
  p.docvalue as totalamount,
  p.amount as calcamount,
  p.price,
  p.unitcost as averagecost,
  (p.docvalue - p.amount) as grossprofit
FROM public.stock_ledger p
LEFT JOIN LATERAL (
  SELECT dl.description AS itemname
  FROM public.docdetail dl
  WHERE dl.businesscode = p.businesscode AND dl.transflag = p.transflag
    AND dl.docno = p.docno AND dl.itemcode = p.itemcode
    AND COALESCE(dl.description, '') <> ''
  ORDER BY dl.linenumber = p.linenumber DESC, dl.linenumber
  LIMIT 1
) item_lookup ON TRUE
WHERE p.transflag = ` + strconv.Itoa(doctype.SaleInvoice) + `
  AND p.businesscode = $3
  AND p.docdatetime >= $1 AND p.docdatetime < $2
  AND NOT EXISTS (
    SELECT 1 FROM public.doc dd
    WHERE dd.businesscode = p.businesscode AND dd.docno = p.docno
      AND dd.transflag = p.transflag AND dd.isdelete
  )`

	args = append(args, fromDate, toDate, businessCode)
	argIndex = 4

	// Add branch filter (parameterized)
	if len(branchCodes) > 0 {
		query += fmt.Sprintf(" AND p.whcode = ANY($%d)", argIndex)
		args = append(args, pq.Array(branchCodes))
		argIndex++
	}

	// Add product filter via subquery (parameterized)
	if len(productCodes) > 0 {
		query += fmt.Sprintf(`
  AND p.docno IN (
    SELECT DISTINCT sub.docno
    FROM public.stock_ledger sub
    WHERE sub.transflag = %d
      AND sub.businesscode = $3
      AND sub.docdatetime >= $1 AND sub.docdatetime < $2
      AND sub.itemcode = ANY($%d)`, doctype.SaleInvoice, argIndex)
		args = append(args, pq.Array(productCodes))
		argIndex++

		// Add branch filter to subquery if present
		if len(branchCodes) > 0 {
			query += fmt.Sprintf(" AND sub.whcode = ANY($%d)", argIndex-2)
		}
		query += ")"
	}

	if sortAscending {
		query += "\nORDER BY p.docdatetime ASC, p.docno, p.linenumber"
	} else {
		query += "\nORDER BY p.docdatetime DESC, p.docno, p.linenumber"
	}

	query += fmt.Sprintf("\nLIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	return query, args
}

// SalesReportSummaryHandler - สรุปยอดขายรวม
func SalesReportSummaryHandler(c echo.Context) error {
	var req SalesReportRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload",
			"code":  "INVALID_PAYLOAD",
		})
	}

	if req.HoldingCode == "" || req.FromDate == "" || req.ToDate == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "holdingcode, from_date and to_date are required",
			"code":  "MISSING_REQUIRED_FIELDS",
		})
	}

	fromDate, err := time.Parse("2006-01-02", req.FromDate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid from_date format",
			"code":  "INVALID_FROM_DATE",
		})
	}

	toDate, err := time.Parse("2006-01-02", req.ToDate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid to_date format",
			"code":  "INVALID_TO_DATE",
		})
	}
	toDate = toDate.AddDate(0, 0, 1)

	businessCode, scopeErr := reportCompanyScope(c, req.HoldingCode)
	if scopeErr != nil {
		return taxScopeFail(c, scopeErr)
	}

	db, err := mypg.PgSqlFastConnect(req.HoldingCode)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Database connection failed",
			"code":  "DB_CONNECTION_ERROR",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query, args := buildSalesReportSummaryQuery(fromDate, toDate, businessCode, req.BranchCodes, req.ProductCodes)

	var totalDocuments int64
	var totalAmount, totalCost, totalProfit decimal.Decimal

	err = db.QueryRowContext(ctx, query, args...).Scan(
		&totalDocuments, &totalAmount, &totalCost, &totalProfit,
	)
	if err != nil {
		logger.Error("SalesReport: query failed for %s/%s: %v", req.HoldingCode, businessCode, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Query execution failed",
			"code":  "QUERY_ERROR",
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"status": "success",
		"data":   salesSummaryData(totalDocuments, totalAmount, totalCost, totalProfit),
	})
}

// salesSummaryData - ยอดสรุปเป็นข้อความทศนิยม (ไม่ส่งเงินเป็น JSON number) ไม่ปัดทิ้งความละเอียดที่บันทึกไว้
func salesSummaryData(documents int64, amount, cost, profit decimal.Decimal) map[string]any {
	return map[string]any{
		"total_documents": documents,
		"totalamount":     reportMoneyText(amount),
		"totalcost":       reportMoneyText(cost),
		"total_profit":    reportMoneyText(profit),
		"profit_margin":   moneyText(salesMarginPercent(amount, profit)),
	}
}

// buildSalesReportSummaryQuery - สร้าง query สำหรับสรุปยอดขาย
func buildSalesReportSummaryQuery(fromDate, toDate time.Time, businessCode string, branchCodes, productCodes []string) (string, []any) {
	args := make([]any, 0)
	argIndex := 1

	query := `
SELECT
  COUNT(DISTINCT p.docno) as total_documents,
  COALESCE(SUM(p.docvalue), 0) as totalamount,
  COALESCE(SUM(p.amount), 0) as totalcost,
  COALESCE(SUM(p.docvalue) - SUM(p.amount), 0) as total_profit
FROM public.stock_ledger p
WHERE p.transflag = ` + strconv.Itoa(doctype.SaleInvoice) + `
  AND p.businesscode = $3
  AND p.docdatetime >= $1 AND p.docdatetime < $2
  AND NOT EXISTS (
    SELECT 1 FROM public.doc dd
    WHERE dd.businesscode = p.businesscode AND dd.docno = p.docno
      AND dd.transflag = p.transflag AND dd.isdelete
  )`

	args = append(args, fromDate, toDate, businessCode)
	argIndex = 4

	if len(branchCodes) > 0 {
		query += fmt.Sprintf(" AND p.whcode = ANY($%d)", argIndex)
		args = append(args, pq.Array(branchCodes))
		argIndex++
	}

	if len(productCodes) > 0 {
		query += fmt.Sprintf(" AND p.itemcode = ANY($%d)", argIndex)
		args = append(args, pq.Array(productCodes))
	}

	return query, args
}

// salesMoneyColumns - คอลัมน์เงินของรายงานขาย (stock_ledger เก็บ numeric(18,8)) ส่งออกด้วย reportMoneyText:
// อย่างน้อย 2 ตำแหน่ง ไม่ปัดทิ้งความละเอียดที่บันทึกไว้ (ต้นทุน = จำนวน × ต้นทุนต่อหน่วย มีได้ถึง 8 ตำแหน่ง)
// เพื่อให้ ยอดขาย − ต้นทุน = กำไรขั้นต้น ทุกแถว และยอดรวมท้ายรายงานกระทบกันได้
// price/averagecost มีเฉพาะแบบรายละเอียด (ค่าที่บันทึกต่อบรรทัด); แบบสรุปไม่ส่ง AVG ที่ทศนิยมยาวเกินที่บันทึก
// totalqty ของรายงานแบบรายละเอียด = docqty (จำนวน ไม่ใช่เงิน) จึงคงข้อความ numeric จากฐาน
var salesMoneyColumns = []string{"totalamount", "price", "averagecost", "calcamount", "grossprofit"}

// formatSalesReportRow - แปลงคอลัมน์เงินเป็น decimal แล้วส่งเป็นข้อความ "1234.50" / "0.125" (Rule 20: ไม่ใช้ float/JSON number กับเงิน)
// และคำนวณ marginpercent ที่ backend จากยอดเต็มความละเอียด (frontend แสดงผลอย่างเดียว)
func formatSalesReportRow(row map[string]any) error {
	amounts := make(map[string]decimal.Decimal, len(salesMoneyColumns))
	for _, key := range salesMoneyColumns {
		raw, ok := row[key]
		if !ok || raw == nil {
			continue
		}
		text, ok := raw.(string)
		if !ok {
			return fmt.Errorf("%s: unexpected type %T", key, raw)
		}
		value, err := decimal.NewFromString(text)
		if err != nil {
			return fmt.Errorf("%s: %w", key, err)
		}
		amounts[key] = value
		row[key] = reportMoneyText(value)
	}
	total, hasTotal := amounts["totalamount"]
	profit, hasProfit := amounts["grossprofit"]
	if hasTotal && hasProfit {
		row["marginpercent"] = moneyText(salesMarginPercent(total, profit))
	}
	return nil
}

// salesMarginPercent - % กำไรขั้นต้น = กำไร × 100 ÷ ยอดขาย ปัด 2 ตำแหน่งแบบครึ่งปัดออกจากศูนย์ (เท่ากับ moneyText)
// ยอดขายเป็นศูนย์ → 0 (ไม่มีฐานให้คิดอัตรา)
func salesMarginPercent(amount, profit decimal.Decimal) decimal.Decimal {
	if amount.IsZero() {
		return decimal.Zero
	}
	return profit.Mul(decimal.NewFromInt(100)).DivRound(amount, 2)
}
