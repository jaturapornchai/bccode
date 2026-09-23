package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"smlcloudplatform/internal/generalledger"
	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/mypg"
	"smlcloudplatform/internal/whtcert"

	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
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
	DocDate          string `json:"docdate"`
	TaxInvoiceNo     string `json:"taxinvoiceno"`
	CounterpartyName string `json:"counterpartyname"`
	TaxID            string `json:"taxid"`
	BranchNo         string `json:"branchno"`
	AmountBeforeVat  string `json:"amountbeforevat"` // ทศนิยม 2 ตำแหน่งแบบ string (ห้ามส่งเงินเป็น JSON number)
	VatAmount        string `json:"vatamount"`
	TotalAmount      string `json:"totalamount"`
}

// TaxVatRegisterSummary - ยอดรวมทั้งงวด (ไม่ใช่เฉพาะหน้าที่แสดง) คำนวณใน PostgreSQL
type TaxVatRegisterSummary struct {
	AmountBeforeVat string `json:"amountbeforevat"`
	VatAmount       string `json:"vatamount"`
	TotalAmount     string `json:"totalamount"`
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
		var beforeVat, vat, total decimal.Decimal
		if err := rows.Scan(
			&row.DocDate,
			&row.TaxInvoiceNo,
			&row.CounterpartyName,
			&row.TaxID,
			&row.BranchNo,
			&beforeVat,
			&vat,
			&total,
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
		row.AmountBeforeVat, row.VatAmount, row.TotalAmount = moneyText(beforeVat), moneyText(vat), moneyText(total)
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

	// ยอดรวมท้ายรายงานต้องเป็นของทั้งงวด ไม่ใช่ผลรวมเฉพาะหน้าที่ browser ได้รับ
	summaryQuery, summaryArgs := buildVatRegisterSummaryQuery(req.Type, req.Year, req.Month)
	var totalRows int
	var sumBefore, sumVat, sumTotal decimal.Decimal
	if err := db.QueryRowContext(ctx, summaryQuery, summaryArgs...).Scan(&totalRows, &sumBefore, &sumVat, &sumTotal); err != nil {
		logger.Error("TaxVatRegister: summary: %v", err)
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
		"total":  totalRows,
		"summary": TaxVatRegisterSummary{
			AmountBeforeVat: moneyText(sumBefore),
			VatAmount:       moneyText(sumVat),
			TotalAmount:     moneyText(sumTotal),
		},
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
	table, masterTable := vatRegisterTables(docType)

	// คอลัมน์ยอดเงินของตารางเอกสาร ERP ยังเป็น double precision จึงแปลงเป็น numeric แล้วปัด 2 ตำแหน่งใน SQL
	// ต่อใบก่อนส่งออก — ผลรวมทุกจุดใช้ค่าที่ปัดแล้วชุดเดียวกัน ยอดท้ายรายงานจึงเท่ากับผลบวกของแถวเสมอ
	query := fmt.Sprintf(`
SELECT
  TO_CHAR(t.docdate + INTERVAL '7 hour', 'YYYY-MM-DD') AS docdate,
  COALESCE(t.taxdocno, '') AS taxinvoiceno,
  COALESCE(m.name0, '') AS counterpartyname,
  COALESCE(m.taxid, '') AS taxid,
  COALESCE(m.branchnumber, '') AS branchno,
  %s AS amountbeforevat,
  %s AS vatamount,
  %s AS totalamount
FROM %s t
LEFT JOIN %s m ON m.code = t.creditorcode
%s
ORDER BY t.docdate ASC, t.docno ASC
LIMIT $3 OFFSET $4`, moneySQL("t.totalbeforevat"), moneySQL("t.totalvatvalue"), moneySQL("t.totalaftervat"), table, masterTable, vatRegisterWhere)

	return query, []any{year, month, limit, offset}
}

// buildVatRegisterSummaryQuery - จำนวนใบและยอดรวมทั้งงวด (เงื่อนไขเดียวกับรายการ ไม่มี LIMIT)
func buildVatRegisterSummaryQuery(docType string, year, month int) (string, []any) {
	table, _ := vatRegisterTables(docType)
	query := fmt.Sprintf(`
SELECT COUNT(*), COALESCE(SUM(%s), 0), COALESCE(SUM(%s), 0), COALESCE(SUM(%s), 0)
FROM %s t
%s`, moneySQL("t.totalbeforevat"), moneySQL("t.totalvatvalue"), moneySQL("t.totalaftervat"), table, vatRegisterWhere)
	return query, []any{year, month}
}

const vatRegisterWhere = `WHERE t.iscancel = false
  AND t.totalvatvalue <> 0
  AND EXTRACT(YEAR FROM t.docdate + INTERVAL '7 hour') = $1
  AND EXTRACT(MONTH FROM t.docdate + INTERVAL '7 hour') = $2`

// vatRegisterTables - ชื่อตารางมาจาก whitelist คงที่ในโค้ดเท่านั้น
func vatRegisterTables(docType string) (table, masterTable string) {
	if docType == "purchase" {
		return "public.purchasetransaction", "public.creditor"
	}
	return "public.saleinvoicetransaction", "public.debtor"
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
	// ภาษีชำระเกินยกมาจากเดือนก่อน (ภ.พ.30 ข้อ 8) — ผู้ใช้กรอกเป็นทศนิยม string เช่น "1250.50"; ว่าง = 0
	CreditBroughtForward string `json:"creditbroughtforward,omitempty"`
}

// PP30SummaryData - แบบ ภ.พ.30 ข้อ 1–10 ประจำเดือน ยอดเงินทุกช่องเป็นทศนิยม 2 ตำแหน่งแบบ string
type PP30SummaryData struct {
	Year                 int           `json:"year"`
	Month                int           `json:"month"`
	Company              CompanyHeader `json:"company"`
	SalesGross           string        `json:"salesgross"`           // ข้อ 1 ยอดขายเดือนนี้ (= 2 + 3 + 4)
	SalesZeroRated       string        `json:"saleszerorated"`       // ข้อ 2 ยอดขายอัตราร้อยละ 0
	SalesExempt          string        `json:"salesexempt"`          // ข้อ 3 ยอดขายที่ได้รับยกเว้น
	SalesTaxable         string        `json:"salestaxable"`         // ข้อ 4 ยอดขายที่ต้องเสียภาษี
	OutputVat            string        `json:"outputvat"`            // ข้อ 5 ภาษีขาย (ยอดที่บันทึกจริงรายใบ)
	PurchaseTaxable      string        `json:"purchasetaxable"`      // ข้อ 6 ยอดซื้อที่มีสิทธินำภาษีซื้อมาหัก
	InputVat             string        `json:"inputvat"`             // ข้อ 7 ภาษีซื้อ (ยอดที่บันทึกจริงรายใบ)
	CreditBroughtForward string        `json:"creditbroughtforward"` // ข้อ 8 ภาษีชำระเกินยกมา
	NetVat               string        `json:"netvat"`               // 5 - (7 + 8)
	Payable              string        `json:"payable"`              // ข้อ 9 ภาษีที่ต้องชำระ
	Creditable           string        `json:"creditable"`           // ข้อ 10 ภาษีชำระเกิน
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

	holdingCode, businessCode, scopeErr := authenticatedCompanyContext(c, req.HoldingCode, req.BusinessCode)
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

	creditForward, ok := parseMoneyInput(req.CreditBroughtForward)
	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"code":    "INVALID_AMOUNT",
			"message": "creditbroughtforward must be a non-negative amount with at most 2 decimals",
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

	var salesTaxable, salesZeroRated, salesExempt, outputVat decimal.Decimal
	if err := db.QueryRowContext(ctx, salesQuery, salesArgs...).Scan(&salesTaxable, &salesZeroRated, &salesExempt, &outputVat); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"code":    "QUERY_ERROR",
			"message": "Sales aggregation query failed",
		})
	}

	purchaseQuery, purchaseArgs := buildPP30PurchaseQuery(req.Year, req.Month)

	var purchaseTaxable, inputVat decimal.Decimal
	if err := db.QueryRowContext(ctx, purchaseQuery, purchaseArgs...).Scan(&purchaseTaxable, &inputVat); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"code":    "QUERY_ERROR",
			"message": "Purchase aggregation query failed",
		})
	}

	company, err := loadCompanyHeader(ctx, holdingCode, businessCode)
	if err != nil {
		logger.Error("PP30Summary: company header: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"code":    "QUERY_ERROR",
			"message": "Company lookup failed",
		})
	}

	netVat, payable, creditable := computeVatSettlement(outputVat, inputVat, creditForward)

	return c.JSON(http.StatusOK, map[string]any{
		"status": "success",
		"data": PP30SummaryData{
			Year:                 req.Year,
			Month:                req.Month,
			Company:              company,
			SalesGross:           moneyText(salesTaxable.Add(salesZeroRated).Add(salesExempt)),
			SalesZeroRated:       moneyText(salesZeroRated),
			SalesExempt:          moneyText(salesExempt),
			SalesTaxable:         moneyText(salesTaxable),
			OutputVat:            moneyText(outputVat),
			PurchaseTaxable:      moneyText(purchaseTaxable),
			InputVat:             moneyText(inputVat),
			CreditBroughtForward: moneyText(creditForward),
			NetVat:               moneyText(netVat),
			Payable:              moneyText(payable),
			Creditable:           moneyText(creditable),
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
	// ปัดต่อใบก่อนรวม (moneySQL) ให้ยอด ภ.พ.30 เท่ากับผลรวมของรายงานภาษีขายรายใบเป๊ะ
	base, vat, exempt := moneySQL("s.totalbeforevat"), moneySQL("s.totalvatvalue"), moneySQL("s.totalexceptvat")
	query := fmt.Sprintf(`
SELECT
  COALESCE(SUM(CASE WHEN NOT (s.vatrate = 0 AND s.totalvatvalue = 0) THEN %[1]s ELSE 0 END), 0) AS salestaxable,
  COALESCE(SUM(CASE WHEN s.vatrate = 0 AND s.totalvatvalue = 0 AND s.totalbeforevat <> 0 THEN %[1]s ELSE 0 END), 0) AS saleszerorated,
  COALESCE(SUM(%[3]s), 0) AS salesexempt,
  COALESCE(SUM(%[2]s), 0) AS outputvat
FROM public.saleinvoicetransaction s
WHERE s.iscancel = false
  AND EXTRACT(YEAR FROM s.docdate + INTERVAL '7 hour') = $1
  AND EXTRACT(MONTH FROM s.docdate + INTERVAL '7 hour') = $2`, base, vat, exempt)
	return query, []any{year, month}
}

// buildPP30PurchaseQuery - รวมยอดซื้อตามงวดจาก purchasetransaction
// purchasetaxable = SUM(totalbeforevat) รวมทุกอัตรา (contract ไม่ได้ขอแยก zero-rated/exempt ฝั่งซื้อ)
// inputvat        = SUM(totalvatvalue) ยอด VAT ที่บันทึกจริงรายเอกสาร
func buildPP30PurchaseQuery(year, month int) (string, []any) {
	query := fmt.Sprintf(`
SELECT
  COALESCE(SUM(%s), 0) AS purchasetaxable,
  COALESCE(SUM(%s), 0) AS inputvat
FROM public.purchasetransaction p
WHERE p.iscancel = false
  AND EXTRACT(YEAR FROM p.docdate + INTERVAL '7 hour') = $1
  AND EXTRACT(MONTH FROM p.docdate + INTERVAL '7 hour') = $2`, moneySQL("p.totalbeforevat"), moneySQL("p.totalvatvalue"))
	return query, []any{year, month}
}

// computeVatSettlement - ภ.พ.30: netvat = ข้อ5 - (ข้อ7 + ข้อ8); บวก = ข้อ9 ต้องชำระ, ลบ = ข้อ10 ชำระเกิน
func computeVatSettlement(outputVat, inputVat, creditForward decimal.Decimal) (netVat, payable, creditable decimal.Decimal) {
	netVat = outputVat.Sub(inputVat).Sub(creditForward)
	switch netVat.Sign() {
	case 1:
		return netVat, netVat, decimal.Zero
	case -1:
		return netVat, decimal.Zero, netVat.Neg()
	}
	return decimal.Zero, decimal.Zero, decimal.Zero
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
	JournalID   string `json:"journalid"`
	DocNo       string `json:"docno"`
	DocDate     string `json:"docdate"`
	PartnerCode string `json:"partnercode"`
	PartnerName string `json:"partnername"`
	TaxID       string `json:"taxid"`
	Address     string `json:"address"`
	Description string `json:"description"`
	BaseAmount  string `json:"baseamount"`    // ทศนิยม 2 ตำแหน่งแบบ string
	WhtAmount   string `json:"whtamount"`     // ยอดหักที่บันทึกจริงในบรรทัด GL
	WhtText     string `json:"whtamounttext"` // ยอดหักเป็นตัวอักษรไทย สำหรับหนังสือรับรอง 50 ทวิ
	NetAmount   string `json:"netamount"`     // ยอดจ่ายสุทธิ = ฐาน - ยอดหัก
	RatePercent string `json:"ratepercent"`   // อัตรา = ยอดหัก / ฐาน × 100 ปัด 2 ตำแหน่ง ("" เมื่อไม่มีฐาน)
	// ฐานภาษีมาจากไหน: recorded = ผู้ใช้บันทึกในรายละเอียดใบสำคัญ (แก้ได้เสมอ), inferred = ระบบประมาณจากบรรทัดบัญชี
	TaxBaseSource string `json:"taxbasesource"`
	FormType      string `json:"formtype,omitempty"`   // PND2/PND3/PND53 ตามที่บันทึก
	IncomeType    string `json:"incometype,omitempty"` // รหัสประเภทเงินได้ของแบบ 50 ทวิ
	Condition     int    `json:"condition,omitempty"`  // 1=หัก ณ ที่จ่าย 2=ออกให้ตลอดไป 3=ออกให้ครั้งเดียว
	PaidDate      string `json:"paiddate,omitempty"`
	CertificateNo string `json:"certificateno,omitempty"`

	base, wht decimal.Decimal
}

// TaxWithholdingRateGroup - สรุปตามอัตราภาษี สำหรับหน้าสรุปแบบยื่น ภ.ง.ด.
type TaxWithholdingRateGroup struct {
	RatePercent string `json:"ratepercent"`
	Count       int    `json:"count"`
	BaseAmount  string `json:"baseamount"`
	WhtAmount   string `json:"whtamount"`
}

// TaxWithholdingSummary - ยอดรวมทั้งงวด (ทุกแถว ไม่ใช่เฉพาะหน้าที่แสดง)
type TaxWithholdingSummary struct {
	BaseTotal    string                    `json:"basetotal"`
	WhtTotal     string                    `json:"whttotal"`
	WhtTotalText string                    `json:"whttotaltext"`
	NetTotal     string                    `json:"nettotal"`
	PayeeCount   int                       `json:"payeecount"`
	ByRate       []TaxWithholdingRateGroup `json:"byrate"`
}

// taxWithholdingReport - ผลคำนวณรายงานทั้งงวดก่อนตัดหน้า
type taxWithholdingReport struct {
	Rows    []TaxWithholdingRow
	Summary TaxWithholdingSummary
	Note    string
}

// whtReportMaxRows - เพดานรายการต่อเดือน (ภาษีหักของ SME ต่อเดือนหลักสิบถึงหลักร้อยใบ) กันงานหนักผิดปกติ
const whtReportMaxRows = 5000

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

	holdingCode, businessCode, scopeErr := authenticatedCompanyContext(c, req.HoldingCode, req.BusinessCode)
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
	for _, f := range req.Forms {
		if !validWhtForms[f] {
			return c.JSON(http.StatusBadRequest, map[string]any{
				"success": false,
				"code":    "INVALID_FORM",
				"message": "forms must be one of 1, 2, 3, 53, 54",
			})
		}
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
	// ห้าม db.Close(): PgSqlFastConnect คืน pool กลางของ holding ที่ทุก request ใช้ร่วมกัน

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// กลุ่มกิจการใหม่ที่ยังไม่เคยเปิด GL ต้องเห็นรายงานว่าง ไม่ใช่ error เพราะยังไม่มีตาราง gl_*
	if err := generalledger.EnsureSchema(ctx, db); err != nil {
		logger.Error("TaxWithholding: ensure GL schema: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]any{"success": false, "code": "QUERY_ERROR", "message": "Query execution failed"})
	}

	report, err := buildWithholdingReport(ctx, db, businessCode, req.Year, req.Month, req.Direction, req.Forms)
	if err != nil {
		logger.Error("TaxWithholding: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"code":    "QUERY_ERROR",
			"message": "Query execution failed",
		})
	}

	company, err := loadCompanyHeader(ctx, holdingCode, businessCode)
	if err != nil {
		logger.Error("TaxWithholding: company header: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"code":    "QUERY_ERROR",
			"message": "Company lookup failed",
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"status":  "success",
		"data":    pageWithholdingRows(report.Rows, limit, offset),
		"total":   len(report.Rows),
		"summary": report.Summary,
		"company": company,
		"limit":   limit,
		"offset":  offset,
		"note":    report.Note,
	})
}

// pageWithholdingRows - ตัดหน้าหลังคำนวณครบทั้งงวดแล้ว (ยอดรวมจึงเป็นของทั้งงวดเสมอ)
func pageWithholdingRows(rows []TaxWithholdingRow, limit, offset int) []TaxWithholdingRow {
	if offset >= len(rows) {
		return []TaxWithholdingRow{}
	}
	end := offset + limit
	if end > len(rows) {
		end = len(rows)
	}
	return rows[offset:end]
}

// validWhtForms - แบบยื่นที่ระบบรองรับการจับคู่บัญชี (จับจากชื่อบัญชีที่มีรูปแบบ "ภ.ง.ด.<แบบ>")
var validWhtForms = map[string]bool{"1": true, "2": true, "3": true, "53": true, "54": true}

// buildWithholdingReport - คำนวณรายการหัก ณ ที่จ่ายทั้งงวดจาก GL จริง (แยกออกจาก handler เพื่อทดสอบ integration ตรงกับ *sql.DB)
//
// หลักการอ่านข้อมูล (ตรวจกับโครง runtime จริง backend/internal/generalledger/schema.sql):
//   - บัญชีภาษีหัก = บัญชี liability ที่ชื่อภาษาไทยอ้างแบบยื่น เช่น "ภ.ง.ด.53" (ผังมาตรฐานแยกบัญชีตามแบบยื่น:
//     ภ.ง.ด.1 เงินเดือน / ภ.ง.ด.3 บุคคลธรรมดา / ภ.ง.ด.53 นิติบุคคล / ภ.ง.ด.2 ดอกเบี้ยปันผล / ภ.ง.ด.54 ต่างประเทศ)
//     ฝั่ง received คือบัญชี asset ที่ชื่อมีคำว่า "ภาษีถูกหัก" — ชั่วคราวจนกว่าจะมีตาราง wht_records ตาม
//     mydocs/datamodels/gl/wht.sql (หนังสือรับรองแยกรายฉบับ) ซึ่งจะเลิกการหาจากชื่อบัญชี
//   - ยอดหักมาจาก gl_lines เฉพาะใบที่ผ่านรายการแล้ว (projection เก็บเฉพาะ posted)
//   - คู่ค้า/เลขผู้เสียภาษีจากหลักฐานประกอบของใบเดียวกัน (documents/settlements → gl_subledger_partners)
//   - ฐานภาษี = ยอดตัดยอดตาม settlements ของใบ; ถ้าไม่มีการตัดยอดใช้ผลรวมเดบิตของใบ (ไม่รวมบรรทัดภาษีหักเอง)
//     ยอดหักเป็นยอดที่บันทึกจริงเสมอ ไม่คำนวณย้อนจากอัตรา; ทุกยอดเป็น decimal ไม่มี float
func buildWithholdingReport(ctx context.Context, db *sql.DB, company string, year, month int, direction string, forms []string) (taxWithholdingReport, error) {
	report := taxWithholdingReport{Rows: []TaxWithholdingRow{}, Summary: TaxWithholdingSummary{ByRate: []TaxWithholdingRateGroup{}}}
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
				return report, fmt.Errorf("invalid form %q", f)
			}
			formPatterns = append(formPatterns, "ภ.ง.ด."+f)
		}
	}

	accountType := "liability"
	if direction == "received" {
		accountType = ""
	}
	whtAccounts, err := findWithholdingAccounts(ctx, db, company, accountType, formPatterns)
	if err != nil {
		return report, err
	}
	if len(whtAccounts) == 0 {
		report.Note = "ไม่พบบัญชีภาษีหัก ณ ที่จ่ายในผังบัญชี (ค้นตามแบบยื่น ภ.ง.ด.) รายงานจึงว่าง — เพิ่มบัญชีภาษีหักในผังบัญชีแล้วรายงานจะแสดงทันที"
		report.Summary = summarizeWithholding(report.Rows)
		return report, nil
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
LIMIT $5`
	lineRows, err := db.QueryContext(ctx, lineQuery, company, pqArray(whtAccounts), year, month, whtReportMaxRows+1)
	if err != nil {
		return report, fmt.Errorf("read wht lines: %w", err)
	}
	type lineRow struct {
		journalID, docNo, docDate, description string
		credit, debit                          decimal.Decimal
	}
	var lines []lineRow
	for lineRows.Next() {
		var r lineRow
		if err := lineRows.Scan(&r.journalID, &r.docNo, &r.docDate, &r.credit, &r.debit, &r.description); err != nil {
			lineRows.Close()
			return report, fmt.Errorf("scan wht line: %w", err)
		}
		lines = append(lines, r)
	}
	lineRows.Close()
	if err := lineRows.Err(); err != nil {
		return report, fmt.Errorf("read wht lines: %w", err)
	}
	if len(lines) > whtReportMaxRows {
		return report, fmt.Errorf("wht lines exceed %d for %04d-%02d", whtReportMaxRows, year, month)
	}

	recorded := map[string]bool{}
	for _, line := range lines {
		if recorded[line.journalID] {
			continue
		}
		// ใบที่บันทึกรายการภาษีหักไว้ ใช้ฐาน/ยอดหัก/อัตราที่ผู้ใช้บันทึกตรง ๆ (หนึ่งแถวต่อรายการเงินได้)
		items, err := recordedWithholdings(ctx, db, company, line.journalID, direction, forms)
		if err != nil {
			return report, err
		}
		if len(items) > 0 {
			recorded[line.journalID] = true
			for _, item := range items {
				row := TaxWithholdingRow{JournalID: line.journalID, DocNo: line.docNo, DocDate: line.docDate, Description: item.Description,
					TaxBaseSource: "recorded", FormType: item.FormType, IncomeType: item.IncomeType, Condition: item.Condition,
					PaidDate: item.PaymentDate, CertificateNo: item.CertificateNo, base: item.BaseAmount.Decimal()}
				if item.TaxAmount != nil {
					row.wht = item.TaxAmount.Decimal()
				}
				if row.Description == "" {
					row.Description = line.description
				}
				if err := fillPartnerMaster(ctx, db, company, item.PartnerCode, &row); err != nil {
					return report, err
				}
				report.Rows = append(report.Rows, finishWithholdingRow(row, moneyText(item.Rate.Decimal())))
			}
			continue
		}
		row := TaxWithholdingRow{
			JournalID:     line.journalID,
			DocNo:         line.docNo,
			DocDate:       line.docDate,
			Description:   line.description,
			TaxBaseSource: "inferred",
			wht:           line.debit,
		}
		if direction == "paid" {
			row.wht = line.credit
		}
		if err := fillWithholdingEvidence(ctx, db, company, line.journalID, direction, &row); err != nil {
			return report, err
		}
		report.Rows = append(report.Rows, finishWithholdingRow(row, ""))
	}
	report.Summary = summarizeWithholding(report.Rows)
	return report, nil
}

// findWithholdingAccounts - รหัสบัญชีภาษีหักจากผังบัญชีจริง (ค้นจากชื่อบัญชีตามแบบยื่น)
func findWithholdingAccounts(ctx context.Context, db *sql.DB, company, accountType string, patterns []string) ([]string, error) {
	rows, err := db.QueryContext(ctx, `
SELECT payload->>'accountcode' AS code
FROM gl_records
WHERE company = $1 AND kind = 'accounts'
  AND (payload->>'accounttype' = $2 OR $2 = '')
  AND EXISTS (
    SELECT 1 FROM jsonb_array_elements(payload->'names') n, unnest($3::text[]) f
    WHERE n->>'name' LIKE '%' || f || '%'
  )`, company, accountType, pqArray(patterns))
	if err != nil {
		return nil, fmt.Errorf("find wht accounts: %w", err)
	}
	defer rows.Close()
	codes := []string{}
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, fmt.Errorf("scan wht account: %w", err)
		}
		codes = append(codes, code)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read wht accounts: %w", err)
	}
	return codes, nil
}

// summarizeWithholding - ยอดรวมทั้งงวด + สรุปตามอัตรา (สำหรับหน้าสรุปแบบยื่น) จากยอด decimal ของทุกแถว
func summarizeWithholding(rows []TaxWithholdingRow) TaxWithholdingSummary {
	baseTotal, whtTotal := decimal.Zero, decimal.Zero
	payees := map[string]bool{}
	type group struct {
		count     int
		base, wht decimal.Decimal
	}
	groups := map[string]*group{}
	var order []string
	for _, r := range rows {
		baseTotal, whtTotal = baseTotal.Add(r.base), whtTotal.Add(r.wht)
		payee := r.TaxID
		if payee == "" {
			payee = r.PartnerCode
		}
		if payee == "" {
			payee = "journal:" + r.JournalID
		}
		payees[payee] = true
		g, ok := groups[r.RatePercent]
		if !ok {
			g = &group{}
			groups[r.RatePercent] = g
			order = append(order, r.RatePercent)
		}
		g.count++
		g.base, g.wht = g.base.Add(r.base), g.wht.Add(r.wht)
	}
	byRate := make([]TaxWithholdingRateGroup, 0, len(order))
	for _, rate := range order {
		g := groups[rate]
		byRate = append(byRate, TaxWithholdingRateGroup{RatePercent: rate, Count: g.count, BaseAmount: moneyText(g.base), WhtAmount: moneyText(g.wht)})
	}
	return TaxWithholdingSummary{
		BaseTotal:    moneyText(baseTotal),
		WhtTotal:     moneyText(whtTotal),
		WhtTotalText: whtcert.BahtText(whtTotal),
		NetTotal:     moneyText(baseTotal.Sub(whtTotal)),
		PayeeCount:   len(payees),
		ByRate:       byRate,
	}
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
	settleBase := decimal.Zero
	partnerCode := ""
	var evidenceDocID string
	if journal.Details != nil {
		for _, s := range journal.Details.Settlements {
			if v, err := decimal.NewFromString(strings.Trim(string(s.Amount), `"`)); err == nil {
				settleBase = settleBase.Add(v)
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
	if settleBase.Sign() > 0 {
		row.base = settleBase
	} else {
		// ไม่มีการตัดยอด: ใช้ผลรวมฝั่งตรงข้ามของบรรทัดภาษีหักเป็นฐาน (เช่น เดบิตเจ้าหนี้ในใบจ่าย)
		var base decimal.NullDecimal
		opposite := "debit"
		if direction == "received" {
			opposite = "credit"
		}
		if err := db.QueryRowContext(ctx, `SELECT COALESCE(SUM(`+opposite+`),0) FROM gl_lines WHERE company=$1 AND journal_id=$2 AND account_code NOT IN (
			SELECT payload->>'accountcode' FROM gl_records WHERE company=$1 AND kind='accounts' AND EXISTS (
				SELECT 1 FROM jsonb_array_elements(payload->'names') n WHERE n->>'name' LIKE '%ภาษีหัก ณ ที่จ่าย%' OR n->>'name' LIKE '%ภาษีถูกหัก%'))`,
			company, journalID).Scan(&base); err == nil && base.Valid {
			row.base = base.Decimal
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
	return fillPartnerMaster(ctx, db, company, partnerCode, row)
}

// fillPartnerMaster - ชื่อ/เลขผู้เสียภาษี/ที่อยู่ของคู่ค้าจากทะเบียนคู่ค้าของห้องบัญชี
func fillPartnerMaster(ctx context.Context, db *sql.DB, company, partnerCode string, row *TaxWithholdingRow) error {
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

// recordedWithholdings - รายการภาษีหักที่ผู้ใช้บันทึกในรายละเอียดใบสำคัญ ตามทิศทางและแบบยื่นที่ขอ
func recordedWithholdings(ctx context.Context, db *sql.DB, company, journalID, direction string, forms []string) ([]generalledger.SubledgerWithholding, error) {
	var raw []byte
	if err := db.QueryRowContext(ctx, `SELECT COALESCE(payload->'details'->'withholdings','[]'::jsonb) FROM gl_records WHERE company=$1 AND kind='journals' AND id=$2`, company, journalID).Scan(&raw); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("load recorded withholdings: %w", err)
	}
	var all []generalledger.SubledgerWithholding
	if err := json.Unmarshal(raw, &all); err != nil {
		return nil, fmt.Errorf("parse recorded withholdings: %w", err)
	}
	want, allowed := 1, map[string]bool{}
	if direction == "received" {
		want = 2
	}
	for _, f := range forms {
		allowed["PND"+f] = true
	}
	items := all[:0]
	for _, item := range all {
		if item.Direction == want && (want == 2 || allowed[item.FormType]) {
			items = append(items, item)
		}
	}
	return items, nil
}

// finishWithholdingRow - ยอดข้อความ/สุทธิ/อัตรา; อัตราที่บันทึกไว้ชนะอัตราที่คำนวณย้อนจากยอด
func finishWithholdingRow(row TaxWithholdingRow, recordedRate string) TaxWithholdingRow {
	row.WhtAmount, row.BaseAmount = moneyText(row.wht), moneyText(row.base)
	row.WhtText = whtcert.BahtText(row.wht)
	row.NetAmount = moneyText(row.base.Sub(row.wht))
	switch {
	case recordedRate != "":
		row.RatePercent = recordedRate
	case row.base.Sign() > 0:
		row.RatePercent = moneyText(row.wht.Mul(decimal.NewFromInt(100)).Div(row.base))
	}
	return row
}

// pqArray - แปลงรายการรหัสบัญชีเป็น text[] สำหรับ ANY($n) (ใช้ lib/pq เวอร์ชันเดียวกับทั้งโปรเจกต์)
func pqArray(items []string) interface{} {
	return pq.StringArray(items)
}
