package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/mypg"

	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
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
			// รายงานภาษีต้องครบทุกใบ แถวที่อ่านไม่ได้ต้องแจ้งให้รู้ ไม่ใช่ข้ามเงียบ ๆ
			// ทะเบียนภาษีที่ขาดใบกำกับไปเฉย ๆ คือรายงานที่ผิดโดยไม่มีใครเห็น
			logger.Error("TaxVatRegister: scan row: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]any{
				"success": false,
				"code":    "SCAN_ERROR",
				"message": "Query execution failed",
			})
		}
		data = append(data, row)
	}
	if err := rows.Err(); err != nil {
		logger.Error("TaxVatRegister: read rows: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"code":    "QUERY_ERROR",
			"message": "Query execution failed",
		})
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

// ---------------------------------------------------------------------------
// POST /api/report/tax/wht — รายงานภาษีหัก ณ ที่จ่าย (ภ.ง.ด.3 / ภ.ง.ด.53 / ถูกหัก)
// ---------------------------------------------------------------------------

// TaxWithholdingRequest - request รายงานภาษีหัก ณ ที่จ่าย
type TaxWithholdingRequest struct {
	HoldingCode  string   `json:"holdingcode"`
	BusinessCode string   `json:"businesscode"`
	Year         int      `json:"year"`
	Month        int      `json:"month"`
	Direction    string   `json:"direction"`       // "paid" = เราหักคู่ค้า (ภ.ง.ด.3/53), "received" = เราถูกหัก
	Forms        []string `json:"forms,omitempty"` // กรองตามแบบยื่น เช่น ["53"], ["3"] — ว่าง = ทุกแบบที่ระบบค้นพบ
	Limit        int      `json:"limit,omitempty"`
	Offset       int      `json:"offset,omitempty"`
}

// TaxWithholdingRow - แถวภาษีหัก ณ ที่จ่ายต่อใบสำคัญ (ข้อมูลจริงจาก GL และหลักฐานประกอบ)
type TaxWithholdingRow struct {
	JournalID   string  `json:"journalid"`
	DocNo       string  `json:"docno"`
	DocDate     string  `json:"docdate"`
	PartnerCode string  `json:"partnercode"`
	PartnerName string  `json:"partnername"`
	TaxID       string  `json:"taxid"`
	Address     string  `json:"address"`
	Description string  `json:"description"`
	BaseAmount  float64 `json:"baseamount"`
	WhtAmount   float64 `json:"whtamount"`
	RatePercent float64 `json:"ratepercent"`
}

// TaxWithholdingHandler - POST /api/report/tax/wht
// อ่านยอดหัก ณ ที่จ่ายจากบัญชีภาษีหักในบัญชีแยกประเภทที่ผ่านรายการจริง (gl_lines)
// แล้วโยงคู่ค้า/เลขผู้เสียภาษีจากหลักฐานประกอบ (gl_subledger_partners + documents/settlements)
func TaxWithholdingHandler(c echo.Context) error {
	var req TaxWithholdingRequest
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

	if req.Direction != "paid" && req.Direction != "received" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"code":    "INVALID_TYPE",
			"message": "direction must be 'paid' or 'received'",
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
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	rows, total, basetotal, whttotal, note, err := buildWithholdingRows(ctx, db, req.BusinessCode, req.Year, req.Month, req.Direction, req.Forms, limit, offset)
	if err != nil {
		logger.Error("TaxWithholding: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"code":    "QUERY_ERROR",
			"message": "Query execution failed",
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"status": "success",
		"data":   rows,
		"count":  total,
		"summary": map[string]any{
			"basetotal": basetotal,
			"whttotal":  whttotal,
		},
		"limit":  limit,
		"offset": offset,
		"note":   note,
	})
}

// validWhtForms - แบบยื่นที่ระบบรองรับการจับคู่บัญชี (จับจากชื่อบัญชีที่มีรูปแบบ "ภ.ง.ด.<แบบ>")
var validWhtForms = map[string]bool{"1": true, "2": true, "3": true, "53": true, "54": true}

// buildWithholdingRows - คำนวณรายการหัก ณ ที่จ่ายจาก GL จริง (แยกออกจาก handler เพื่อทดสอบ integration ตรงกับ *sql.DB)
//
// หลักการอ่านข้อมูล (ตรวจกับโครง runtime จริง backend/internal/generalledger/schema.sql):
//   - บัญชีภาษีหัก = บัญชี liability ที่ชื่อภาษาไทยอ้างแบบยื่น เช่น "ภ.ง.ด.53" (ผังมาตรฐานแยกบัญชีตามแบบยื่น:
//     ภ.ง.ด.1 เงินเดือน / ภ.ง.ด.3 บุคคลธรรมดา / ภ.ง.ด.53 นิติบุคคล / ภ.ง.ด.2 ดอกเบี้ยปันผล / ภ.ง.ด.54 ต่างประเทศ)
//     ฝั่ง received คือบัญชี asset ที่ชื่อมีคำว่า "ภาษีถูกหัก" — อ่านจากผังบัญชีจริง ไม่ hard-code รหัสบัญชี
//   - ยอดหักมาจาก gl_lines เฉพาะใบที่ผ่านรายการแล้ว (projection เก็บเฉพาะ posted)
//   - คู่ค้า/เลขผู้เสียภาษีจากหลักฐานประกอบของใบเดียวกัน (documents/settlements → gl_subledger_partners)
//   - ฐานภาษี = ยอดตัดยอดตาม settlements ของใบ; ถ้าไม่มีการตัดยอดใช้ผลรวมเดบิตของใบ (ไม่รวมบรรทัดภาษีหักเอง)
//     เอกสารบางใบจ่ายหลายบิลรวมกัน ฐานจึงแสดงรวมต่อใบ — ยอดหักเป็นยอดที่บันทึกจริงเสมอ ไม่คำนวณย้อนจากอัตรา
func buildWithholdingRows(ctx context.Context, db *sql.DB, company string, year, month int, direction string, forms []string, limit, offset int) ([]TaxWithholdingRow, int, float64, float64, string, error) {
	suppress := "credit"
	var formPatterns []string
	if direction == "received" {
		suppress = "debit"
		formPatterns = []string{"ภาษีถูกหัก"}
	} else {
		if len(forms) == 0 {
			forms = []string{"1", "2", "3", "53", "54"}
		}
		for _, f := range forms {
			if !validWhtForms[f] {
				return nil, 0, 0, 0, "", fmt.Errorf("invalid form %q", f)
			}
			formPatterns = append(formPatterns, "ภ.ง.ด."+f)
		}
	}

	accountQuery := `
SELECT payload->>'accountcode' AS code
FROM gl_records
WHERE company = $1 AND kind = 'accounts'
  AND (payload->>'accounttype' = $2 OR $2 = '')
  AND EXISTS (
    SELECT 1 FROM jsonb_array_elements(payload->'names') n, unnest($3::text[]) f
    WHERE n->>'name' LIKE '%' || f || '%'
  )`
	acctRows, err := db.QueryContext(ctx, accountQuery, company, func() string {
		if direction == "received" {
			return ""
		}
		return "liability"
	}(), pqArray(formPatterns))
	if err != nil {
		return nil, 0, 0, 0, "", fmt.Errorf("find wht accounts: %w", err)
	}
	whtAccounts := []string{}
	for acctRows.Next() {
		var code string
		if err := acctRows.Scan(&code); err != nil {
			acctRows.Close()
			return nil, 0, 0, 0, "", fmt.Errorf("scan wht account: %w", err)
		}
		whtAccounts = append(whtAccounts, code)
	}
	acctRows.Close()
	if err := acctRows.Err(); err != nil {
		return nil, 0, 0, 0, "", fmt.Errorf("read wht accounts: %w", err)
	}
	if len(whtAccounts) == 0 {
		note := "ไม่พบบัญชีภาษีหัก ณ ที่จ่ายในผังบัญชี (ค้นตามแบบยื่น ภ.ง.ด.) รายงานจึงว่าง — เพิ่มบัญชีภาษีหักในผังบัญชีแล้วรายงานจะแสดงทันที"
		return []TaxWithholdingRow{}, 0, 0, 0, note, nil
	}

	lineQuery := `
SELECT l.journal_id, l.doc_no, TO_CHAR(l.entry_date, 'YYYY-MM-DD'), l.credit, l.debit, l.description
FROM gl_lines l
WHERE l.company = $1
  AND l.account_code = ANY($2)
  AND l.` + suppress + ` > 0
  AND EXTRACT(YEAR FROM l.entry_date) = $3
  AND EXTRACT(MONTH FROM l.entry_date) = $4
ORDER BY l.entry_date ASC, l.journal_id ASC
LIMIT $5 OFFSET $6`
	lineRows, err := db.QueryContext(ctx, lineQuery, company, pqArray(whtAccounts), year, month, limit, offset)
	if err != nil {
		return nil, 0, 0, 0, "", fmt.Errorf("read wht lines: %w", err)
	}
	type lineRow struct {
		journalID, docNo, docDate, description string
		credit, debit                          float64
	}
	var lines []lineRow
	for lineRows.Next() {
		var r lineRow
		if err := lineRows.Scan(&r.journalID, &r.docNo, &r.docDate, &r.credit, &r.debit, &r.description); err != nil {
			lineRows.Close()
			return nil, 0, 0, 0, "", fmt.Errorf("scan wht line: %w", err)
		}
		lines = append(lines, r)
	}
	lineRows.Close()
	if err := lineRows.Err(); err != nil {
		return nil, 0, 0, 0, "", fmt.Errorf("read wht lines: %w", err)
	}

	results := make([]TaxWithholdingRow, 0, len(lines))
	var basetotal, whttotal float64
	for _, line := range lines {
		row := TaxWithholdingRow{
			JournalID:   line.journalID,
			DocNo:       line.docNo,
			DocDate:     line.docDate,
			Description: line.description,
		}
		if direction == "paid" {
			row.WhtAmount = line.credit
		} else {
			row.WhtAmount = line.debit
		}
		whttotal += row.WhtAmount

		if err := fillWithholdingEvidence(ctx, db, company, line.journalID, direction, &row); err != nil {
			return nil, 0, 0, 0, "", err
		}
		if row.BaseAmount > 0 {
			row.RatePercent = math.Round(row.WhtAmount/row.BaseAmount*1000) / 10
		}
		basetotal += row.BaseAmount
		results = append(results, row)
	}
	return results, len(results), basetotal, whttotal, "", nil
}

// fillWithholdingEvidence - โยงคู่ค้าและฐานภาษีของใบจากหลักฐานประกอบจริง
func fillWithholdingEvidence(ctx context.Context, db *sql.DB, company, journalID, direction string, row *TaxWithholdingRow) error {
	var payload []byte
	if err := db.QueryRowContext(ctx, `SELECT payload FROM gl_records WHERE company=$1 AND kind='journals' AND id=$2`, company, journalID).Scan(&payload); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("load journal payload: %w", err)
	}

	var journal struct {
		Description string `json:"description"`
		Details     *struct {
			Documents []struct {
				ID          string `json:"id"`
				Ledger      string `json:"ledger"`
				PartnerCode string `json:"partner_code"`
			} `json:"documents"`
			Settlements []struct {
				ID                string          `json:"id"`
				PartnerCode       string          `json:"partner_code"`
				DebtDocumentID    string          `json:"debt_document_id"`
				PaymentDocumentID string          `json:"payment_document_id"`
				Amount            json.RawMessage `json:"amount"`
			} `json:"settlements"`
		} `json:"details"`
	}
	if err := json.Unmarshal(payload, &journal); err != nil {
		return fmt.Errorf("parse journal payload: %w", err)
	}
	if row.Description == "" {
		row.Description = journal.Description
	}

	// ฐานภาษี: รวมยอดตัดยอดของใบ (ตัดชำระจริงต่อบิล) — ยอดหักยังเป็นยอดบรรทัดจริงเสมอ
	settleBase := 0.0
	partnerCode := ""
	var evidenceDocID string
	if journal.Details != nil {
		for _, s := range journal.Details.Settlements {
			if v, err := strconv.ParseFloat(strings.Trim(string(s.Amount), `"`), 64); err == nil {
				settleBase += v
			}
			if partnerCode == "" {
				partnerCode = s.PartnerCode
			}
			if direction == "paid" && evidenceDocID == "" && s.DebtDocumentID != "" {
				evidenceDocID = s.DebtDocumentID
			}
			if direction == "received" && evidenceDocID == "" && s.PaymentDocumentID != "" {
				evidenceDocID = s.PaymentDocumentID
			}
		}
		// คู่ค้าจากเอกสารของใบนี้เอง (ใบจ่ายมักผูกเอกสารลดหนี้ฝั่งเจ้าหนี้, ใบถูกหักผูกเอกสารลูกหนี้)
		wantLedger := "ap"
		if direction == "received" {
			wantLedger = "ar"
		}
		for _, d := range journal.Details.Documents {
			if d.Ledger == wantLedger && d.PartnerCode != "" {
				partnerCode = d.PartnerCode
				if evidenceDocID == "" {
					evidenceDocID = d.ID
				}
				break
			}
		}
	}
	if settleBase > 0 {
		row.BaseAmount = settleBase
	} else {
		// ไม่มีการตัดยอด: ใช้ผลรวมฝั่งตรงข้ามของบรรทัดภาษีหักเป็นฐาน (เช่น เดบิตเจ้าหนี้ในใบจ่าย)
		var base sql.NullFloat64
		opposite := "debit"
		if direction == "received" {
			opposite = "credit"
		}
		if err := db.QueryRowContext(ctx, `SELECT COALESCE(SUM(`+opposite+`),0) FROM gl_lines WHERE company=$1 AND journal_id=$2 AND account_code NOT IN (
			SELECT payload->>'accountcode' FROM gl_records WHERE company=$1 AND kind='accounts' AND EXISTS (
				SELECT 1 FROM jsonb_array_elements(payload->'names') n WHERE n->>'name' LIKE '%ภาษีหัก ณ ที่จ่าย%' OR n->>'name' LIKE '%ภาษีถูกหัก%'))`,
			company, journalID).Scan(&base); err == nil && base.Valid {
			row.BaseAmount = base.Float64
		}
	}

	if partnerCode == "" && evidenceDocID != "" {
		if err := db.QueryRowContext(ctx, `SELECT partner_code FROM gl_subledger_documents WHERE company=$1 AND id=$2`, company, evidenceDocID).Scan(&partnerCode); err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("load document partner: %w", err)
		}
	}
	if partnerCode == "" {
		return nil
	}
	row.PartnerCode = partnerCode
	var name, taxID, address sql.NullString
	if err := db.QueryRowContext(ctx, `SELECT COALESCE(payload->>'name_th',''), COALESCE(payload->>'tax_id',''), COALESCE(payload->>'address','') FROM gl_subledger_partners WHERE company=$1 AND code=$2`, company, partnerCode).Scan(&name, &taxID, &address); err != nil {
		if err != sql.ErrNoRows {
			return fmt.Errorf("load partner master: %w", err)
		}
		return nil
	}
	row.PartnerName, row.TaxID, row.Address = name.String, taxID.String, address.String
	return nil
}

// pqArray - แปลงรายการรหัสบัญชีเป็น text[] สำหรับ ANY($n) (ใช้ lib/pq เวอร์ชันเดียวกับทั้งโปรเจกต์)
func pqArray(items []string) interface{} {
	return pq.StringArray(items)
}
