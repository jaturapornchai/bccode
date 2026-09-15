package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"smlcloudplatform/internal/goapi/mypg"

	"github.com/labstack/echo/v4"
)

const (
	taxRegisterDefaultLimit = 200
	taxRegisterMaxLimit     = 500
)

// ---------------------------------------------------------------------------
// POST /api/report/tax/vat-register — รายงานภาษีซื้อ/ขายรายเอกสาร
// ---------------------------------------------------------------------------

// TaxVatRegisterRequest - request สำหรับรายงานภาษีซื้อ/ขาย (VAT Register)
type TaxVatRegisterRequest struct {
	HoldingCode  string `json:"holdingcode"`
	BusinessCode string `json:"businesscode"`
	Year         int    `json:"year"`
	Month        int    `json:"month"`
	Type         string `json:"type"` // "sale" หรือ "purchase"
	Limit        int    `json:"limit,omitempty"`
	Offset       int    `json:"offset,omitempty"`
}

// TaxVatRegisterRow - แถวรายงานภาษีซื้อ/ขายต่อเอกสาร
type TaxVatRegisterRow struct {
	DocDate          string  `json:"docdate"`
	TaxInvoiceNo     string  `json:"taxinvoiceno"`
	CounterpartyName string  `json:"counterpartyname"`
	TaxID            string  `json:"taxid"`
	BranchNo         string  `json:"branchno"`
	AmountBeforeVat  float64 `json:"amountbeforevat"`
	VatAmount        float64 `json:"vatamount"`
	TotalAmount      float64 `json:"totalamount"`
}

// TaxVatRegisterHandler - POST /api/report/tax/vat-register
// รายงานภาษีซื้อ/ขายรายเอกสาร อ่านจาก saleinvoicetransaction/purchasetransaction จริงใน PostgreSQL
// (ไม่ mock ข้อมูล — ฟิลด์ใดไม่มีจริงในตารางจะคืนค่าว่าง/0 ผ่าน COALESCE)
func TaxVatRegisterHandler(c echo.Context) error {
	var req TaxVatRegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"code":    "INVALID_PAYLOAD",
			"message": "Invalid request payload",
		})
	}

	holdingCode, _, scopeErr := authenticatedCompanyContext(c, req.HoldingCode, req.BusinessCode)
	if scopeErr != nil {
		return c.JSON(scopeErr.Status, map[string]any{
			"success": false,
			"code":    scopeErr.Code,
			"message": scopeErr.Message,
		})
	}

	if req.Type != "sale" && req.Type != "purchase" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"code":    "INVALID_TYPE",
			"message": "type must be 'sale' or 'purchase'",
		})
	}

	if !isValidReportPeriod(req.Year, req.Month) {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"code":    "INVALID_PERIOD",
			"message": "year and month are required (month 1-12)",
		})
	}

	limit, offset := normalizeVatRegisterPaging(req.Limit, req.Offset)

	db, err := mypg.PgSqlFastConnect(holdingCode)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"code":    "DB_CONNECTION_ERROR",
			"message": "Database connection failed",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query, args := buildVatRegisterQuery(req.Type, req.Year, req.Month, limit, offset)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"code":    "QUERY_ERROR",
			"message": "Query execution failed",
		})
	}
	defer rows.Close()

	data := make([]TaxVatRegisterRow, 0)
	for rows.Next() {
		var row TaxVatRegisterRow
		if err := rows.Scan(
			&row.DocDate,
			&row.TaxInvoiceNo,
			&row.CounterpartyName,
			&row.TaxID,
			&row.BranchNo,
			&row.AmountBeforeVat,
			&row.VatAmount,
			&row.TotalAmount,
		); err != nil {
			continue
		}
		data = append(data, row)
	}

	return c.JSON(http.StatusOK, map[string]any{
		"status": "success",
		"data":   data,
		"count":  len(data),
		"limit":  limit,
		"offset": offset,
	})
}

// buildVatRegisterQuery - สร้าง query รายเอกสารสำหรับ VAT register (parameterized ทั้งหมด ป้องกัน SQL injection;
// ชื่อ table มาจาก whitelist คงที่ในโค้ดเท่านั้น ไม่ประกอบจาก input ของผู้ใช้)
//
// docType "sale"     -> public.saleinvoicetransaction join public.debtor
// docType "purchase" -> public.purchasetransaction   join public.creditor
//
// หมายเหตุ: คอลัมน์รหัสคู่ค้าบน saleinvoicetransaction ชื่อจริงในฐานข้อมูลคือ "creditorcode" แม้จะเก็บรหัส "ลูกค้า"
// ก็ตาม (Go struct field ชื่อ DebtorCode แต่ gorm/json tag คือ creditorcode) — ดู
// backend/internal/transaction/models/transaction_saleinvoice_postgres.go:14
func buildVatRegisterQuery(docType string, year, month, limit, offset int) (string, []any) {
	table := "public.saleinvoicetransaction"
	masterTable := "public.debtor"
	if docType == "purchase" {
		table = "public.purchasetransaction"
		masterTable = "public.creditor"
	}

	query := fmt.Sprintf(`
SELECT
  TO_CHAR(t.docdate + INTERVAL '7 hour', 'YYYY-MM-DD') AS docdate,
  COALESCE(t.taxdocno, '') AS taxinvoiceno,
  COALESCE(m.name0, '') AS counterpartyname,
  COALESCE(m.taxid, '') AS taxid,
  COALESCE(m.branchnumber, '') AS branchno,
  COALESCE(t.totalbeforevat, 0) AS amountbeforevat,
  COALESCE(t.totalvatvalue, 0) AS vatamount,
  COALESCE(t.totalaftervat, 0) AS totalamount
FROM %s t
LEFT JOIN %s m ON m.code = t.creditorcode
WHERE t.iscancel = false
  AND t.totalvatvalue <> 0
  AND EXTRACT(YEAR FROM t.docdate + INTERVAL '7 hour') = $1
  AND EXTRACT(MONTH FROM t.docdate + INTERVAL '7 hour') = $2
ORDER BY t.docdate ASC, t.docno ASC
LIMIT $3 OFFSET $4`, table, masterTable)

	return query, []any{year, month, limit, offset}
}

// ---------------------------------------------------------------------------
// POST /api/report/tax/pp30-summary — สรุปยอด ภ.พ.30 รายเดือน
// ---------------------------------------------------------------------------

// PP30SummaryRequest - request สำหรับสรุปยอดภาษีมูลค่าเพิ่มประจำเดือน (ภ.พ.30)
type PP30SummaryRequest struct {
	HoldingCode  string `json:"holdingcode"`
	BusinessCode string `json:"businesscode"`
	Year         int    `json:"year"`
	Month        int    `json:"month"`
}

// PP30SummaryData - สรุปยอดภาษีซื้อ/ขายประจำเดือน (ภ.พ.30)
type PP30SummaryData struct {
	Year            int     `json:"year"`
	Month           int     `json:"month"`
	SalesTaxable    float64 `json:"salestaxable"`
	SalesZeroRated  float64 `json:"saleszerorated"`
	SalesExempt     float64 `json:"salesexempt"`
	OutputVat       float64 `json:"outputvat"`
	PurchaseTaxable float64 `json:"purchasetaxable"`
	InputVat        float64 `json:"inputvat"`
	NetVat          float64 `json:"netvat"`
	Payable         float64 `json:"payable"`
	Creditable      float64 `json:"creditable"`
}

// PP30SummaryHandler - POST /api/report/tax/pp30-summary
// outputvat/inputvat เป็นผลรวมของยอด VAT ที่บันทึกจริงรายเอกสาร (SUM(totalvatvalue)) เสมอ
// ห้ามคำนวณจากฐาน x 7% ตามกฎบัญชี (ดู prompt ผู้สั่งงาน)
func PP30SummaryHandler(c echo.Context) error {
	var req PP30SummaryRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"code":    "INVALID_PAYLOAD",
			"message": "Invalid request payload",
		})
	}

	holdingCode, _, scopeErr := authenticatedCompanyContext(c, req.HoldingCode, req.BusinessCode)
	if scopeErr != nil {
		return c.JSON(scopeErr.Status, map[string]any{
			"success": false,
			"code":    scopeErr.Code,
			"message": scopeErr.Message,
		})
	}

	if !isValidReportPeriod(req.Year, req.Month) {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"code":    "INVALID_PERIOD",
			"message": "year and month are required (month 1-12)",
		})
	}

	db, err := mypg.PgSqlFastConnect(holdingCode)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"code":    "DB_CONNECTION_ERROR",
			"message": "Database connection failed",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	salesQuery, salesArgs := buildPP30SalesQuery(req.Year, req.Month)

	var salesTaxable, salesZeroRated, salesExempt, outputVat float64
	if err := db.QueryRowContext(ctx, salesQuery, salesArgs...).Scan(&salesTaxable, &salesZeroRated, &salesExempt, &outputVat); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"code":    "QUERY_ERROR",
			"message": "Sales aggregation query failed",
		})
	}

	purchaseQuery, purchaseArgs := buildPP30PurchaseQuery(req.Year, req.Month)

	var purchaseTaxable, inputVat float64
	if err := db.QueryRowContext(ctx, purchaseQuery, purchaseArgs...).Scan(&purchaseTaxable, &inputVat); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"code":    "QUERY_ERROR",
			"message": "Purchase aggregation query failed",
		})
	}

	netVat, payable, creditable := computeVatSettlement(outputVat, inputVat)

	return c.JSON(http.StatusOK, map[string]any{
		"status": "success",
		"data": PP30SummaryData{
			Year:            req.Year,
			Month:           req.Month,
			SalesTaxable:    salesTaxable,
			SalesZeroRated:  salesZeroRated,
			SalesExempt:     salesExempt,
			OutputVat:       outputVat,
			PurchaseTaxable: purchaseTaxable,
			InputVat:        inputVat,
			NetVat:          netVat,
			Payable:         payable,
			Creditable:      creditable,
		},
	})
}

// buildPP30SalesQuery - รวมยอดขายตามงวดจาก saleinvoicetransaction
//
// salesexempt = SUM(totalexceptvat) — ฟิลด์ "มูลค่ายกเว้นภาษี" ที่มีอยู่จริงในเอกสารแต่ละใบ
// outputvat   = SUM(totalvatvalue)  — ยอด VAT ที่บันทึกจริงรายเอกสาร (ตรงกับ vat-register เป๊ะ)
//
// salestaxable/saleszerorated: backend ไม่มี const/comment ที่ยืนยันความหมายของ vattype (0/1/...) ที่ชัดเจน
// (grep ทั่ว repo ไม่พบ enum ใน Go แม้แต่ docs/kms/decisions/2026-09-03-product-two-user-groups-plan-proposed.md:207
// เองก็บันทึกไว้ว่ายังไม่ยืนยัน) จึงไม่ใช้ vattype แบ่งกลุ่ม แต่ใช้ vatrate+totalvatvalue ซึ่งมีความหมายชัดจากชื่อ
// คอลัมน์จริง: เอกสารที่ totalbeforevat>0 แต่ vatrate=0 และ totalvatvalue=0 ถือเป็น "0%" (zero-rated), ที่เหลือถือเป็น
// "ปกติ" (taxable). ถ้าความเข้าใจนี้ผิดต้องแก้ตาม business rule ที่ยืนยันแล้วเท่านั้น
func buildPP30SalesQuery(year, month int) (string, []any) {
	query := `
SELECT
  COALESCE(SUM(CASE WHEN NOT (s.vatrate = 0 AND s.totalvatvalue = 0) THEN s.totalbeforevat ELSE 0 END), 0) AS salestaxable,
  COALESCE(SUM(CASE WHEN s.vatrate = 0 AND s.totalvatvalue = 0 AND s.totalbeforevat <> 0 THEN s.totalbeforevat ELSE 0 END), 0) AS saleszerorated,
  COALESCE(SUM(s.totalexceptvat), 0) AS salesexempt,
  COALESCE(SUM(s.totalvatvalue), 0) AS outputvat
FROM public.saleinvoicetransaction s
WHERE s.iscancel = false
  AND EXTRACT(YEAR FROM s.docdate + INTERVAL '7 hour') = $1
  AND EXTRACT(MONTH FROM s.docdate + INTERVAL '7 hour') = $2`
	return query, []any{year, month}
}

// buildPP30PurchaseQuery - รวมยอดซื้อตามงวดจาก purchasetransaction
// purchasetaxable = SUM(totalbeforevat) รวมทุกอัตรา (contract ไม่ได้ขอแยก zero-rated/exempt ฝั่งซื้อ)
// inputvat        = SUM(totalvatvalue) ยอด VAT ที่บันทึกจริงรายเอกสาร
func buildPP30PurchaseQuery(year, month int) (string, []any) {
	query := `
SELECT
  COALESCE(SUM(p.totalbeforevat), 0) AS purchasetaxable,
  COALESCE(SUM(p.totalvatvalue), 0) AS inputvat
FROM public.purchasetransaction p
WHERE p.iscancel = false
  AND EXTRACT(YEAR FROM p.docdate + INTERVAL '7 hour') = $1
  AND EXTRACT(MONTH FROM p.docdate + INTERVAL '7 hour') = $2`
	return query, []any{year, month}
}

// computeVatSettlement - netvat = output - input; บวก = ต้องชำระ (payable), ลบ = ขอคืน/ยกไป (creditable)
func computeVatSettlement(outputVat, inputVat float64) (netVat, payable, creditable float64) {
	netVat = outputVat - inputVat
	if netVat > 0 {
		return netVat, netVat, 0
	}
	if netVat < 0 {
		return netVat, 0, -netVat
	}
	return 0, 0, 0
}

// ---------------------------------------------------------------------------
// shared helpers
// ---------------------------------------------------------------------------

// isValidReportPeriod - ตรวจปี/เดือนของงวดภาษี (ปีอยู่ในช่วงสมเหตุสมผล, เดือน 1-12)
func isValidReportPeriod(year, month int) bool {
	return year >= 2000 && year <= 2100 && month >= 1 && month <= 12
}

// normalizeVatRegisterPaging - จำกัด limit ไม่เกิน taxRegisterMaxLimit ตามกฎ security ของโปรเจ็กต์
func normalizeVatRegisterPaging(limit, offset int) (int, int) {
	if limit <= 0 || limit > taxRegisterMaxLimit {
		limit = taxRegisterDefaultLimit
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}
