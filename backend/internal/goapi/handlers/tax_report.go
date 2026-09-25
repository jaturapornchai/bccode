package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"smlcloudplatform/internal/generalledger"
	"smlcloudplatform/internal/goapi/language"
	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/mypg"
	branchmodels "smlcloudplatform/internal/organization/branch/models"
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
	// View - "" = ทะเบียนตามงวดภาษี, "reversed_later" = ใบที่ยื่นในงวดก่อนแล้วกลับรายการในเดือนนี้ (Champ 5522/5523 ยกเลิกข้ามงวด)
	View string `json:"view,omitempty"`
	// BranchCode - สถานประกอบการ (เลขสาขา 5 หลักตาม ภ.พ.20 = สาขาของใบสำคัญ); ว่าง = ทุกสถานประกอบการรวมกัน
	// ม.87 วรรคสาม + ประกาศฯ VAT ฉบับที่ 89 ข้อ 5: รายงานภาษีซื้อ/ขายต้องจัดทำเป็นรายสถานประกอบการ
	BranchCode string `json:"branchcode,omitempty"`
	Limit      int    `json:"limit,omitempty"`
	Offset     int    `json:"offset,omitempty"`
}

// TaxVatRegisterRow - แถวรายงานภาษีซื้อ/ขายต่อเอกสาร
type TaxVatRegisterRow struct {
	DocDate          string `json:"docdate"`
	TaxInvoiceNo     string `json:"taxinvoiceno"`
	CounterpartyName string `json:"counterpartyname"`
	TaxID            string `json:"taxid"`
	BranchNo         string `json:"branchno"`
	AmountBeforeVat  string `json:"amountbeforevat"` // มูลค่าสินค้าหรือบริการทั้งใบ ทศนิยม 2 ตำแหน่งแบบ string (ห้ามส่งเงินเป็น JSON number)
	// ZeroAmount/ExemptAmount - ส่วนของมูลค่าที่เป็นอัตรา 0% / ได้รับยกเว้น (รวมอยู่ใน AmountBeforeVat แล้ว) แยกแสดงแบบ Champ GLRepInputTaxView
	ZeroAmount   string `json:"zeroamount"`
	ExemptAmount string `json:"exemptamount"`
	VatAmount    string `json:"vatamount"`
	TotalAmount  string `json:"totalamount"`
	// DuplicateDocNos - เลขที่ใบสำคัญที่บันทึกใบกำกับภาษีฉบับเดียวกัน (ผู้ออก + เลขที่ + วันที่) — มีเลขที่ใบสำคัญของแถวนี้เอง = ซ้ำในใบสำคัญเดียวกัน; [] = ไม่ซ้ำ; เตือนให้ตรวจ ไม่บล็อก
	DuplicateDocNos []string `json:"duplicatedocnos"`
	// TaxMonth - งวดภาษีที่ยื่นไว้ (YYYY-MM) — มีเฉพาะมุมมองยกเลิกข้ามงวด
	TaxMonth string `json:"taxmonth,omitempty"`
	// ReversalDocNo/ReversalDate - ใบกลับรายการของใบสำคัญนี้ (กลับในเดือนหลังงวดภาษี = แถวยังอยู่ในงวดที่ยื่นแล้ว แบบ Champ CancelOutPeriod)
	ReversalDocNo string `json:"reversaldocno,omitempty"`
	ReversalDate  string `json:"reversaldate,omitempty"`
	ReversedMonth string `json:"reversedmonth,omitempty"`
	// Note - หมายเหตุของแถวในภาษาผู้ใช้ (กลับรายการภายหลัง / เหตุผลที่กลับรายการ)
	Note string `json:"note,omitempty"`
}

// TaxVatRegisterSummary - ยอดรวมทั้งงวด (ไม่ใช่เฉพาะหน้าที่แสดง) คำนวณใน PostgreSQL
type TaxVatRegisterSummary struct {
	AmountBeforeVat string `json:"amountbeforevat"`
	ZeroAmount      string `json:"zeroamount"`
	ExemptAmount    string `json:"exemptamount"`
	VatAmount       string `json:"vatamount"`
	TotalAmount     string `json:"totalamount"`
	DuplicateCount  int    `json:"duplicatecount"` // จำนวนแถวของงวดที่มีใบกำกับฉบับเดียวกันในใบสำคัญอื่น
}

// TaxVatRegisterHandler - POST /api/report/tax/vat-register
// รายงานภาษีซื้อ/ขายรายใบกำกับ อ่านจากรายละเอียดภาษีมูลค่าเพิ่มของใบสำคัญ GL ที่ผ่านรายการแล้ว (details.vats)
// ของบริษัทที่ผู้ใช้เลือก ตามงวดภาษีที่บันทึก (ไม่ใช่วันที่ใบสำคัญ) — ภาษีซื้อนับเฉพาะรายการที่ใช้สิทธิในงวดนี้
func TaxVatRegisterHandler(c echo.Context) error {
	var req TaxVatRegisterRequest
	if err := c.Bind(&req); err != nil {
		return taxReportFail(c, http.StatusBadRequest, "INVALID_PAYLOAD", "tax_form_payload_invalid")
	}

	holdingCode, businessCode, scopeErr := authenticatedCompanyContext(c, req.HoldingCode, req.BusinessCode)
	if scopeErr != nil {
		return taxScopeFail(c, scopeErr)
	}

	if req.Type != "sale" && req.Type != "purchase" {
		return taxReportFail(c, http.StatusBadRequest, "INVALID_TYPE", "tax_report_type_invalid")
	}

	if req.View != "" && req.View != vatViewReversedLater {
		return taxReportFail(c, http.StatusBadRequest, "INVALID_VIEW", "tax_report_view_invalid")
	}

	if !isValidReportPeriod(req.Year, req.Month) {
		return taxReportFail(c, http.StatusBadRequest, "INVALID_PERIOD", "tax_form_period_invalid")
	}

	branch := ""
	if strings.TrimSpace(req.BranchCode) != "" {
		code, err := branchmodels.NormalizeThaiTaxBranchCode(req.BranchCode)
		if err != nil {
			return taxReportFail(c, http.StatusBadRequest, "INVALID_BRANCH", "tax_report_branch_invalid")
		}
		branch = code
	}

	limit, offset := normalizeVatRegisterPaging(req.Limit, req.Offset)

	db, err := mypg.PgSqlFastConnect(holdingCode)
	if err != nil {
		logger.Error("TaxVatRegister: db: %v", err)
		return taxReportFail(c, http.StatusInternalServerError, "DB_CONNECTION_ERROR", "tax_report_failed")
	}
	// ห้าม db.Close(): PgSqlFastConnect คืน pool กลางของ holding ที่ทุก request ใช้ร่วมกัน

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// กลุ่มกิจการใหม่ที่ยังไม่เคยเปิด GL ต้องเห็นรายงานว่าง ไม่ใช่ error เพราะยังไม่มีตาราง gl_*
	if err := generalledger.EnsureSchema(ctx, db); err != nil {
		logger.Error("TaxVatRegister: ensure GL schema: %v", err)
		return taxReportFail(c, http.StatusInternalServerError, "QUERY_ERROR", "tax_report_failed")
	}

	taxType := 2
	if req.Type == "purchase" {
		taxType = 1
	}
	var records []generalledger.VatRecord
	if req.View == vatViewReversedLater {
		records, err = generalledger.VatCrossPeriodCancellations(ctx, db, businessCode, req.Year, req.Month, taxType)
	} else {
		records, err = generalledger.VatRecordsForPeriod(ctx, db, businessCode, req.Year, req.Month, taxType)
	}
	if err != nil {
		logger.Error("TaxVatRegister: %v", err)
		return taxReportFail(c, http.StatusInternalServerError, "QUERY_ERROR", "tax_report_failed")
	}

	company, err := loadCompanyHeader(ctx, holdingCode, businessCode)
	if err != nil {
		logger.Error("TaxVatRegister: company header: %v", err)
		return taxReportFail(c, http.StatusInternalServerError, "QUERY_ERROR", "tax_report_failed")
	}
	establishments, err := loadTaxEstablishments(ctx, holdingCode, businessCode)
	if err != nil {
		logger.Error("TaxVatRegister: establishments: %v", err)
		return taxReportFail(c, http.StatusInternalServerError, "QUERY_ERROR", "tax_report_failed")
	}

	// ยอดรวมท้ายรายงานเป็นของทั้งงวด (ของสถานประกอบการที่เลือก) ไม่ใช่ผลรวมเฉพาะหน้าที่ browser ได้รับ
	rows, summary := buildVatRegister(vatRecordsOfBranch(records, branch))
	lang := taxRequestLanguage(c)
	note := vatRegisterNote(req.View, rows, lang)
	if branch == "" && len(establishments) > 1 {
		note = strings.TrimSpace(note + " " + language.Text("tax_vat_note_all_establishments", lang))
	}
	data := pageVatRegisterRows(rows, limit, offset)
	for i := range data {
		data[i].Note = vatRegisterRowNote(req.View, data[i], lang)
	}
	return c.JSON(http.StatusOK, map[string]any{
		"status":  "success",
		"data":    data,
		"count":   len(data),
		"total":   len(rows),
		"summary": summary,
		"limit":   limit,
		"offset":  offset,
		"note":    note,
		// หัวรายงานตามแบบท้ายประกาศฯ VAT ฉบับที่ 202: ชื่อผู้ประกอบการ เลขผู้เสียภาษี ชื่อสถานประกอบการ สำนักงานใหญ่/สาขา
		"company":        company,
		"establishments": establishments,
		"branchcode":     branch,
	})
}

// vatRecordsOfBranch - แถวของสถานประกอบการเดียว (สาขาของใบสำคัญ); branch ว่าง = ทุกสถานประกอบการรวมกัน
func vatRecordsOfBranch(records []generalledger.VatRecord, branch string) []generalledger.VatRecord {
	if branch == "" {
		return records
	}
	out := make([]generalledger.VatRecord, 0, len(records))
	for _, r := range records {
		if code, err := branchmodels.NormalizeThaiTaxBranchCode(r.BranchCode); err == nil && code == branch {
			out = append(out, r)
		}
	}
	return out
}

// vatViewReversedLater - มุมมองใบที่ยื่นในงวดก่อนแล้วกลับรายการในเดือนที่เลือก
const vatViewReversedLater = "reversed_later"

// vatRegisterNote - หมายเหตุของรายงานในภาษาผู้ใช้: ทะเบียนปกติบอกจำนวนแถวที่กลับรายการภายหลัง; มุมมองยกเลิกข้ามงวดอธิบายว่ายอดงวดเดิมไม่เปลี่ยน
func vatRegisterNote(view string, rows []TaxVatRegisterRow, lang string) string {
	if view == vatViewReversedLater {
		if len(rows) == 0 {
			return ""
		}
		return language.Text("tax_vat_note_cross_period_view", lang)
	}
	reversed := 0
	for _, r := range rows {
		if r.ReversedMonth != "" {
			reversed++
		}
	}
	if reversed == 0 {
		return ""
	}
	return strings.ReplaceAll(language.Text("tax_vat_note_reversed_later", lang), "{count}", strconv.Itoa(reversed))
}

// vatRegisterRowNote - ทะเบียนปกติ: "กลับรายการภายหลังในเดือน ..." (key เดียวกับภาษีหัก ณ ที่จ่าย);
// มุมมองยกเลิกข้ามงวด: งวดภาษีที่ยื่นไว้ + เลขที่ใบกลับรายการ + เหตุผลที่ผู้ใช้บันทึกตอนกลับรายการ (ถ้ามี)
func vatRegisterRowNote(view string, row TaxVatRegisterRow, lang string) string {
	if view == vatViewReversedLater {
		note := strings.NewReplacer("{month}", taxMonthLabel(row.TaxMonth, lang), "{docno}", row.ReversalDocNo).
			Replace(language.Text("tax_vat_row_cancelled", lang))
		if reason := strings.TrimSpace(row.Note); reason != "" {
			note += " — " + reason
		}
		return note
	}
	if row.ReversedMonth == "" {
		return ""
	}
	return strings.ReplaceAll(language.Text("tax_wht_row_reversed_later", lang), "{month}", taxMonthLabel(row.ReversedMonth, lang))
}

// vatSign - ใบลดหนี้ (document_type 3) หักออกจากยอด; ใบกำกับภาษีและใบเพิ่มหนี้บวกเพิ่ม (vat.sql เก็บยอดบวกเสมอ)
func vatSign(documentType int) decimal.Decimal {
	if documentType == 3 {
		return decimal.NewFromInt(-1)
	}
	return decimal.NewFromInt(1)
}

// vatMoney - ยอดที่บันทึก ปัด 2 ตำแหน่งต่อรายการก่อนรวม (ยอดท้ายรายงานจึงเท่ากับผลบวกของแถวเสมอ) แล้วใส่เครื่องหมาย
func vatMoney(amount generalledger.Amount, sign decimal.Decimal) decimal.Decimal {
	return amount.Decimal().Round(2).Mul(sign)
}

// vatAmountOf - ยอดภาษีที่บันทึก (backend คำนวณให้ตอนบันทึกเสมอ; ค่าว่างจากข้อมูลเสียถือเป็น 0 ไม่เดา)
func vatAmountOf(record generalledger.VatRecord, sign decimal.Decimal) decimal.Decimal {
	if record.VatAmount == nil {
		return decimal.Zero
	}
	return vatMoney(*record.VatAmount, sign)
}

// buildVatRegister - แถวรายงานภาษีซื้อ/ขาย + ยอดรวมทั้งงวด (decimal) จากรายการภาษีที่บันทึกในใบสำคัญ
// มูลค่าก่อนภาษี = ฐานภาษี + ยอดอัตรา 0% + ยอดยกเว้น ของใบกำกับ (มูลค่าสินค้า/บริการ ไม่รวม VAT); 0% และยกเว้นแยกแสดงอีกช่อง
func buildVatRegister(records []generalledger.VatRecord) ([]TaxVatRegisterRow, TaxVatRegisterSummary) {
	rows := make([]TaxVatRegisterRow, 0, len(records))
	var sumBefore, sumZero, sumExempt, sumVat decimal.Decimal
	for _, r := range records {
		sign := vatSign(r.DocumentType)
		zero, exempt := vatMoney(r.ZeroRateAmount, sign), vatMoney(r.ExemptAmount, sign)
		before := vatMoney(r.BaseAmount, sign).Add(zero).Add(exempt)
		vat := vatAmountOf(r, sign)
		rows = append(rows, TaxVatRegisterRow{
			DocDate:          r.TaxInvoiceDate,
			TaxInvoiceNo:     r.TaxInvoiceNo,
			CounterpartyName: r.PartnerName,
			TaxID:            r.PartnerTaxID,
			BranchNo:         r.PartnerBranchNo,
			AmountBeforeVat:  moneyText(before),
			ZeroAmount:       moneyText(zero),
			ExemptAmount:     moneyText(exempt),
			VatAmount:        moneyText(vat),
			TotalAmount:      moneyText(before.Add(vat)),
			DuplicateDocNos:  append([]string{}, r.DuplicateDocNos...),
			TaxMonth:         vatTaxMonth(r),
			ReversalDocNo:    r.ReversalDocNo,
			ReversalDate:     r.ReversalDate,
			ReversedMonth:    firstN(r.ReversalDate, 7),
			Note:             r.ReversalReason,
		})
		sumBefore, sumVat = sumBefore.Add(before), sumVat.Add(vat)
		sumZero, sumExempt = sumZero.Add(zero), sumExempt.Add(exempt)
	}
	return rows, TaxVatRegisterSummary{
		AmountBeforeVat: moneyText(sumBefore),
		ZeroAmount:      moneyText(sumZero),
		ExemptAmount:    moneyText(sumExempt),
		VatAmount:       moneyText(sumVat),
		TotalAmount:     moneyText(sumBefore.Add(sumVat)),
		DuplicateCount:  countDuplicateInvoices(records),
	}
}

// vatTaxMonth - งวดภาษีที่บันทึก (YYYY-MM) เฉพาะใบที่มีใบกลับรายการ — ทะเบียนปกติไม่ต้องแสดงซ้ำกับงวดที่เลือก
func vatTaxMonth(r generalledger.VatRecord) string {
	if r.ReversalDocNo == "" || r.TaxPeriodYear == 0 || r.TaxPeriodMonth == 0 {
		return ""
	}
	return fmt.Sprintf("%04d-%02d", r.TaxPeriodYear, r.TaxPeriodMonth)
}

// firstN - n ตัวอักษรแรก (ASCII วันที่ ISO) หรือทั้งสตริงถ้าสั้นกว่า
func firstN(s string, n int) string {
	if len(s) < n {
		return s
	}
	return s[:n]
}

// pageVatRegisterRows - ตัดหน้าหลังคำนวณครบทั้งงวดแล้ว
func pageVatRegisterRows(rows []TaxVatRegisterRow, limit, offset int) []TaxVatRegisterRow {
	if offset >= len(rows) {
		return []TaxVatRegisterRow{}
	}
	end := offset + limit
	if end > len(rows) {
		end = len(rows)
	}
	return rows[offset:end]
}

// pp30Totals - ยอด ภ.พ.30 ข้อ 2–7 รวมจากรายการภาษีที่บันทึก (decimal) — ใช้เติมแบบ ภ.พ.30 ใน tax_form_vat.go
type pp30Totals struct {
	salesTaxable, salesZeroRated, salesExempt, outputVat, purchaseTaxable, inputVat decimal.Decimal
}

// sumPP30 - ขาย: ข้อ 4 = ฐานภาษี, ข้อ 2 = ยอดอัตรา 0%, ข้อ 3 = ยอดยกเว้น, ข้อ 5 = VAT ที่บันทึก
// ซื้อ (เฉพาะรายการที่ใช้สิทธิงวดนี้): ข้อ 6 = ฐานภาษีที่มีสิทธินำภาษีซื้อมาหัก, ข้อ 7 = VAT ที่บันทึก
// ใช้การปัดต่อรายการเดียวกับรายงานภาษีซื้อ/ขาย ยอด ภ.พ.30 จึงเท่ากับผลรวมของรายงานรายใบเป๊ะ
func sumPP30(sales, purchases []generalledger.VatRecord) pp30Totals {
	var t pp30Totals
	for _, r := range sales {
		sign := vatSign(r.DocumentType)
		t.salesTaxable = t.salesTaxable.Add(vatMoney(r.BaseAmount, sign))
		t.salesZeroRated = t.salesZeroRated.Add(vatMoney(r.ZeroRateAmount, sign))
		t.salesExempt = t.salesExempt.Add(vatMoney(r.ExemptAmount, sign))
		t.outputVat = t.outputVat.Add(vatAmountOf(r, sign))
	}
	for _, r := range purchases {
		sign := vatSign(r.DocumentType)
		t.purchaseTaxable = t.purchaseTaxable.Add(vatMoney(r.BaseAmount, sign))
		t.inputVat = t.inputVat.Add(vatAmountOf(r, sign))
	}
	return t
}

// ---------------------------------------------------------------------------
// shared helpers
// ---------------------------------------------------------------------------

// taxRequestLanguage - ภาษาของข้อความถึงผู้ใช้: ?lang ก่อน แล้ว Accept-Language (แบบเดียวกับ shop/shopuser_http.go)
func taxRequestLanguage(c echo.Context) string {
	if lang := strings.TrimSpace(c.QueryParam("lang")); lang != "" {
		return language.Normalize(lang)
	}
	return language.Normalize(c.Request().Header.Get("Accept-Language"))
}

// taxReportFail - ตอบข้อผิดพลาดด้วยรหัสเดิม (สัญญากับ frontend/ระบบภายนอก) + ข้อความจาก languages.tsv ตามภาษาผู้ใช้
func taxReportFail(c echo.Context, status int, code, key string) error {
	return c.JSON(status, map[string]any{"success": false, "code": code, "message": language.Text(key, taxRequestLanguage(c))})
}

// taxScopeKeys - ข้อความของข้อผิดพลาดขอบเขตบริษัท (authenticatedCompanyContext คืนข้อความภาษาเดียว)
var taxScopeKeys = map[string]string{"UNAUTHORIZED": "unauthorized", "FORBIDDEN": "gl_err_company_forbidden", "COMPANY_REQUIRED": "company_required"}

func taxScopeFail(c echo.Context, e *companyContextError) error {
	key, ok := taxScopeKeys[e.Code]
	if !ok {
		return e.respond(c)
	}
	return taxReportFail(c, e.Status, e.Code, key)
}

// isValidReportPeriod - ตรวจปี/เดือนของงวดภาษี (ปีอยู่ในช่วงสมเหตุสมผล, เดือน 1-12)
func isValidReportPeriod(year, month int) bool {
	return year >= 2000 && year <= 2100 && month >= 1 && month <= 12
}

// normalizeVatRegisterPaging - จำกัด limit ไม่เกิน taxRegisterMaxLimit ตามกฎ security ของโปรเจ็กต์
// ขอเกินเพดาน = ได้เพดาน (ไม่ใช่ถอยกลับค่าเริ่มต้น 200 ซึ่งทำให้แถวหายเงียบ ๆ — UAT S22 2026-09-24)
func normalizeVatRegisterPaging(limit, offset int) (int, int) {
	switch {
	case limit <= 0:
		limit = taxRegisterDefaultLimit
	case limit > taxRegisterMaxLimit:
		limit = taxRegisterMaxLimit
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
	Forms        []string `json:"forms,omitempty"` // กรองตามแบบยื่น เช่น ["53"], ["3"] — ว่าง = ทุกแบบ (ภ.ง.ด.2/3/53)
	Limit        int      `json:"limit,omitempty"`
	Offset       int      `json:"offset,omitempty"`
}

// TaxWithholdingRow - แถวภาษีหัก ณ ที่จ่ายหนึ่งแถวต่อรายการที่บันทึกในใบสำคัญ (details.withholdings)
type TaxWithholdingRow struct {
	JournalID string `json:"journalid"`
	// id ของรายการใน details.withholdings — จอ 50 ทวิ ใช้จับคู่รายการแบบตรงตัว ไม่ต้องเดาจากยอด
	WithholdingID string `json:"withholdingid,omitempty"`
	DocNo         string `json:"docno"`
	DocDate       string `json:"docdate"`
	PartnerCode   string `json:"partnercode"`
	PartnerName   string `json:"partnername"`
	TaxID         string `json:"taxid"`
	Address       string `json:"address"`
	// คำนำหน้าชื่อ + อำเภอ/จังหวัด/รหัสไปรษณีย์แยกช่องจากทะเบียนคู่ค้า (payload title_name/addr_district/addr_province/addr_postcode)
	// ใช้กับใบแนบ ภ.ง.ด. และไฟล์ยื่นด้วยสื่อ (Format กลาง บังคับคำนำหน้าทุกแบบ และที่อยู่ 3 ช่องของ ภ.ง.ด.3)
	Title string `json:"title,omitempty"`
	// PartnerFullName - ชื่อสำหรับแสดงบนจอ/CSV/พิมพ์ทะเบียน = คำนำหน้า + ชื่อ (generalledger.PartnerFullName ไม่เติมซ้ำ);
	// partnername/title ยังแยกช่องเหมือนเดิมสำหรับใบแนบ ภ.ง.ด. และไฟล์ยื่นด้วยสื่อ
	PartnerFullName string `json:"partnerfullname,omitempty"`
	// เลขสาขาภาษีของคู่ค้า (5 หลัก, 00000 = สำนักงานใหญ่): snapshot ในรายการก่อน แล้วทะเบียนคู่ค้า — เดิมไม่มีช่องนี้ จอรายงานแสดงสาขาว่าง
	BranchNo      string `json:"branchno,omitempty"`
	District      string `json:"district,omitempty"`
	Province      string `json:"province,omitempty"`
	Postcode      string `json:"postcode,omitempty"`
	Description   string `json:"description"`
	BaseAmount    string `json:"baseamount"`           // ทศนิยม 2 ตำแหน่งแบบ string
	WhtAmount     string `json:"whtamount"`            // ยอดหักที่บันทึกในรายการ
	WhtText       string `json:"whtamounttext"`        // ยอดหักเป็นตัวอักษรไทย สำหรับหนังสือรับรอง 50 ทวิ
	NetAmount     string `json:"netamount"`            // ยอดจ่ายสุทธิ = ฐาน - ยอดหัก ("" เมื่อไม่รู้ฐาน — withholdingNetKnown)
	RatePercent   string `json:"ratepercent"`          // อัตราที่บันทึก (ปัด 2 ตำแหน่ง)
	FormType      string `json:"formtype,omitempty"`   // PND2/PND3/PND53 ตามที่บันทึก
	IncomeType    string `json:"incometype,omitempty"` // รหัสประเภทเงินได้ของแบบ 50 ทวิ
	Condition     int    `json:"condition,omitempty"`  // 1=หัก ณ ที่จ่าย 2=ออกให้ตลอดไป 3=ออกให้ครั้งเดียว
	PaidDate      string `json:"paiddate,omitempty"`
	CertificateNo string `json:"certificateno,omitempty"`
	// ReversedMonth - ใบนี้ถูกกลับรายการในเดือนหลังเดือนที่ยื่น (YYYY-MM): แถวยังอยู่ในรายงาน/แบบของเดือนที่จ่ายเงิน
	// เพราะแบบของเดือนนั้นยื่นแล้ว และเดือนที่กลับรายการไม่มีแถวติดลบ (whtReversalFilter)
	ReversedMonth string `json:"reversedmonth,omitempty"`
	// Note - หมายเหตุของแถวตามภาษาผู้ใช้ (handler เติม เช่น กลับรายการภายหลัง)
	Note string `json:"note,omitempty"`

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
	NoteKey string // key ใน languages.tsv — handler แปลตามภาษาผู้ใช้
}

// whtReportMaxRows - เพดานรายการต่อเดือน (ภาษีหักของ SME ต่อเดือนหลักสิบถึงหลักร้อยใบ) กันงานหนักผิดปกติ
const whtReportMaxRows = 5000

// TaxWithholdingHandler - POST /api/report/tax/wht
// อ่านเฉพาะรายการภาษีหัก ณ ที่จ่ายที่ผู้ใช้บันทึกในใบสำคัญ (details.withholdings) — ไม่ประมาณจากชื่อ/รหัสบัญชี
// (ทะเบียนเดียวแบบ Champ BCAPWTaxList) พร้อม snapshot/ทะเบียนคู่ค้า (gl_subledger_partners)
func TaxWithholdingHandler(c echo.Context) error {
	var req TaxWithholdingRequest
	if err := c.Bind(&req); err != nil {
		return taxReportFail(c, http.StatusBadRequest, "INVALID_PAYLOAD", "tax_form_payload_invalid")
	}

	holdingCode, businessCode, scopeErr := authenticatedCompanyContext(c, req.HoldingCode, req.BusinessCode)
	if scopeErr != nil {
		return taxScopeFail(c, scopeErr)
	}

	if req.Direction != "paid" && req.Direction != "received" {
		return taxReportFail(c, http.StatusBadRequest, "INVALID_TYPE", "tax_report_direction_invalid")
	}
	for _, f := range req.Forms {
		if !validWhtForms[f] {
			return taxReportFail(c, http.StatusBadRequest, "INVALID_FORM", "tax_report_form_invalid")
		}
	}

	if !isValidReportPeriod(req.Year, req.Month) {
		return taxReportFail(c, http.StatusBadRequest, "INVALID_PERIOD", "tax_form_period_invalid")
	}

	limit, offset := normalizeVatRegisterPaging(req.Limit, req.Offset)

	db, err := mypg.PgSqlFastConnect(holdingCode)
	if err != nil {
		logger.Error("TaxWithholding: db: %v", err)
		return taxReportFail(c, http.StatusInternalServerError, "DB_CONNECTION_ERROR", "tax_report_failed")
	}
	// ห้าม db.Close(): PgSqlFastConnect คืน pool กลางของ holding ที่ทุก request ใช้ร่วมกัน

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// กลุ่มกิจการใหม่ที่ยังไม่เคยเปิด GL ต้องเห็นรายงานว่าง ไม่ใช่ error เพราะยังไม่มีตาราง gl_*
	if err := generalledger.EnsureSchema(ctx, db); err != nil {
		logger.Error("TaxWithholding: ensure GL schema: %v", err)
		return taxReportFail(c, http.StatusInternalServerError, "QUERY_ERROR", "tax_report_failed")
	}

	report, err := buildWithholdingReport(ctx, db, businessCode, req.Year, req.Month, req.Direction, req.Forms)
	if err != nil {
		logger.Error("TaxWithholding: %v", err)
		return taxReportFail(c, http.StatusInternalServerError, "QUERY_ERROR", "tax_report_failed")
	}

	company, err := loadCompanyHeader(ctx, holdingCode, businessCode)
	if err != nil {
		logger.Error("TaxWithholding: company header: %v", err)
		return taxReportFail(c, http.StatusInternalServerError, "QUERY_ERROR", "tax_report_failed")
	}

	lang := taxRequestLanguage(c)
	note := strings.Join(withholdingReportNotes(report, lang), " ")
	data := pageWithholdingRows(report.Rows, limit, offset)
	// คำอธิบายว่างของรายการที่บันทึก → ชื่อประเภทเงินได้ตามรหัสที่บันทึก ในภาษาของผู้ใช้ (ไม่ใช่คำบรรยายบรรทัดบัญชี)
	for i := range data {
		data[i].Description = incomeTypeText(data[i], lang)
		if data[i].ReversedMonth != "" {
			data[i].Note = strings.ReplaceAll(language.Text("tax_wht_row_reversed_later", lang), "{month}", taxMonthLabel(data[i].ReversedMonth, lang))
		}
	}
	return c.JSON(http.StatusOK, map[string]any{
		"status":  "success",
		"data":    data,
		"total":   len(report.Rows),
		"summary": report.Summary,
		"company": company,
		"limit":   limit,
		"offset":  offset,
		"note":    note,
	})
}

// withholdingReportNotes - หมายเหตุของรายงานตามภาษาผู้ใช้: ยังไม่มีรายการที่บันทึก, รายการที่กลับรายการในเดือนหลัง
func withholdingReportNotes(report taxWithholdingReport, lang string) []string {
	notes := []string{}
	if report.NoteKey != "" {
		notes = append(notes, language.Text(report.NoteKey, lang))
	}
	count := func(key string, n int) string {
		return strings.ReplaceAll(language.Text(key, lang), "{count}", strconv.Itoa(n))
	}
	if reversed := reversedLaterCount(report.Rows); reversed > 0 {
		notes = append(notes, count("tax_wht_note_reversed_later", reversed))
	}
	return notes
}

// reversedLaterCount - แถวที่ใบถูกกลับรายการในเดือนหลัง (ยังอยู่ในแบบของเดือนที่จ่ายเงิน)
func reversedLaterCount(rows []TaxWithholdingRow) int {
	n := 0
	for _, r := range rows {
		if r.ReversedMonth != "" {
			n++
		}
	}
	return n
}

// taxMonthKeys - key ชื่อเดือนใน languages.tsv (มกราคม…ธันวาคม)
var taxMonthKeys = [12]string{"month_january", "month_february", "month_march", "month_april", "month_may", "month_june",
	"month_july", "month_august", "month_september", "month_october", "month_november", "month_december"}

// taxMonthLabel - "2026-10" → "ตุลาคม 2569" (ภาษาไทยใช้ปี พ.ศ. แบบเอกสารภาษีไทย) / "October 2026"; รูปแบบผิดคืนค่าเดิม
func taxMonthLabel(yearMonth, lang string) string {
	t, err := time.Parse("2006-01", yearMonth)
	if err != nil {
		return yearMonth
	}
	year := t.Year()
	if language.Normalize(lang) == "th" {
		year += 543
	}
	return language.Text(taxMonthKeys[t.Month()-1], lang) + " " + strconv.Itoa(year)
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

// validWhtForms - แบบยื่นที่บันทึกในรายการภาษีหักได้ (generalledger withholdingForms: PND2/PND3/PND53)
// ไม่มี ภ.ง.ด.1 — ระบบเงินเดือนอยู่นอกขอบเขตผลิตภัณฑ์ (AGENTS.md "ไม่ทำระบบเงินเดือน")
var validWhtForms = map[string]bool{"2": true, "3": true, "53": true}

// defaultWhtForms - แบบยื่นทั้งหมดของฝั่ง "เราหักผู้อื่น" เมื่อไม่ระบุแบบ
var defaultWhtForms = []string{"2", "3", "53"}

// whtReversalMonthSQL - เดือน (YYYY-MM) ของใบกลับรายการของใบต้นฉบับ r ที่สถานะ reversed ( = ไม่ใช่ใบที่ถูกกลับ/หาใบกลับไม่พบ)
// งวดภาษีหัก ณ ที่จ่าย = เดือนที่จ่ายเงินได้ (ประกาศกระทรวงการคลังเรื่องขยายกำหนดเวลานำส่ง — docs/kms/21 RD-MOF-WHT-EXT;
// ยื่นทางอินเทอร์เน็ต: ประกาศอธิบดีฯ เกี่ยวกับภาษีเงินได้ ฉบับที่ 111 https://www.rd.go.th/9041.html
// "ภายในเจ็ดวันนับแต่วันสิ้นเดือนของเดือนที่จ่ายเงินได้พึงประเมิน") จึงตัดสินรายเดือนที่ยื่น:
//   - กลับรายการในเดือนที่ยื่นเดียวกันหรือก่อนหน้า → ภาษีของใบนั้นไม่เคยต้องยื่น ไม่แสดงทั้งต้นฉบับและใบกลับ
//   - กลับรายการในเดือนหลัง → แบบของเดือนเดิมยื่นไปแล้ว ต้นฉบับคงอยู่ในเดือนของตัวเอง เดือนที่กลับไม่มีแถวติดลบ;
//     การแก้ภาษีที่ยื่นแล้วทำนอกระบบ: นำส่งเกิน/ผิด/ซ้ำ ขอคืนด้วยคำร้อง ค.10 ภายใน 3 ปี (ประมวลรัษฎากร มาตรา 27 ตรี
//     https://www.rd.go.th/5943.html, แบบ ค.10 https://www.rd.go.th/fileadmin/tax_pdf/others/K10_061260.pdf) — หมายเหตุ tax_wht_note_reversed_later
//
// ใบกลับรายการไม่มี details.withholdings จึงไม่เกิดแถวเองอยู่แล้ว
const whtReversalMonthSQL = generalledger.ReversalMonthSQL

// whtNonTaxJournalKinds - ใบที่เกิดจากการหักภาษีจริงเท่านั้นที่เข้ารายงาน: ใบกลับรายการหักล้างต้นฉบับ,
// ใบยอดยกมา/ปิดบัญชีแค่ยกยอดคงเหลือของงวดก่อน (หนี้ภาษีหักค้างจ่ายของ ธ.ค. ปีก่อนไม่ใช่การหักภาษีใหม่ของ ม.ค. — UAT S24 2026-09-24)
var whtNonTaxJournalKinds = []string{"reversal", "opening", "closing"}

// buildWithholdingReport - รายการภาษีหัก ณ ที่จ่ายทั้งงวด (แยกออกจาก handler เพื่อทดสอบ integration ตรงกับ *sql.DB)
//
// อ่านเฉพาะรายการที่ผู้ใช้บันทึก (details.withholdings) = หลักฐานรายฉบับ ใช้ฐาน/ยอดหัก/อัตราที่บันทึกตรง ๆ หนึ่งแถวต่อรายการ
// เข้างวดตาม "เดือนที่จ่ายเงินได้" (payment_date — docs/kms/21-thai-tax-form-references.md §9) และกติกากลับรายการ whtReversalMonthSQL
// ไม่ประมาณจากบรรทัดบัญชีภาษีหักอีกต่อไป (2026-09-25): ชื่อ/รหัสบัญชีต่างกันทุกผัง และฐาน/อัตรา/คู่ค้าที่ประมาณได้ไม่ใช่หลักฐาน
// — Champ ก็อ่านจากทะเบียนภาษีหักที่บันทึก (BCAPWTaxList) อย่างเดียว; ใบที่ลงบัญชีภาษีหักแต่ยังไม่บันทึกรายการ ต้องบันทึกในใบสำคัญ
func buildWithholdingReport(ctx context.Context, db *sql.DB, company string, year, month int, direction string, forms []string) (taxWithholdingReport, error) {
	report := taxWithholdingReport{Rows: []TaxWithholdingRow{}, Summary: TaxWithholdingSummary{ByRate: []TaxWithholdingRateGroup{}}}
	want, formTypes := 1, []string{}
	if direction == "received" {
		want = 2 // ฝั่งถูกหักไม่มีแบบยื่นของเรา — ทุกแบบ
	} else {
		if len(forms) == 0 {
			forms = defaultWhtForms
		}
		for _, f := range forms {
			if !validWhtForms[f] {
				return report, fmt.Errorf("invalid form %q", f)
			}
			formTypes = append(formTypes, "PND"+f)
		}
	}
	rows, err := recordedWithholdingRows(ctx, db, company, year, month, want, formTypes)
	if err != nil {
		return report, err
	}
	if len(rows) > whtReportMaxRows {
		return report, fmt.Errorf("wht rows exceed %d for %04d-%02d", whtReportMaxRows, year, month)
	}
	report.Rows = rows
	if len(rows) == 0 {
		report.NoteKey = "tax_wht_note_no_records"
	}
	report.Summary = summarizeWithholding(report.Rows)
	return report, nil
}

// recordedWithholdingRows - รายการภาษีหักที่บันทึกในรายละเอียดใบสำคัญ (details.withholdings) ของใบที่ยังมีผล ตามทิศทาง/แบบที่ขอ
// งวด = เดือนของวันที่จ่ายในรายการ (payment_date): ประกาศกระทรวงการคลังเรื่องขยายกำหนดเวลาการนำส่งภาษีเงินได้หัก ณ ที่จ่าย
// ให้ยื่น "ภายในเจ็ดวัน นับแต่วันสิ้นเดือนของเดือนที่จ่ายเงินได้พึงประเมิน" (docs/kms/21-thai-tax-form-references.md §9)
// — วันที่จ่ายว่าง (ข้อมูลเสีย) ใช้วันที่ใบแทน และเรียงตามวันที่จ่ายเดียวกันนี้; อ่านครั้งเดียวทั้งงวดพร้อมทะเบียนคู่ค้า ไม่โหลดทีละใบ
// ใบยอดยกมา/ปิดบัญชี/กลับรายการไม่นับแม้มีรายการที่บันทึก (whtNonTaxJournalKinds); ใบที่ถูกกลับรายการตาม whtReversalMonthSQL
// เทียบกับเดือนที่จ่ายของรายการ (ไม่ใช่วันที่ใบ)
// ทิศทาง/แบบ/เดือนที่จ่ายกรองใน SQL ตั้งแต่ระดับรายการ (CTE items) — เฉพาะรายการที่วันที่จ่ายว่าง/ผิดรูปแบบเท่านั้นที่ต้องหาวันที่ใบจาก
// gl_lines ต่อ (เดิมทุกรายการของทุกงวดต้องหาวันที่ใบก่อนแล้วค่อยกรองเดือน) และ Go แปลง JSON เฉพาะรายการของงวดที่ขอ
func recordedWithholdingRows(ctx context.Context, db *sql.DB, company string, year, month, direction int, formTypes []string) ([]TaxWithholdingRow, error) {
	rows, err := db.QueryContext(ctx, `
WITH items AS (
  SELECT r.id, r.code, r.payload->>'docno' AS docno, LEFT(r.payload->>'date', 10) AS journal_date, w.item, w.n,
    CASE WHEN COALESCE(w.item->>'payment_date', '') ~ '^[0-9]{4}-[0-9]{2}-[0-9]{2}' THEN LEFT(w.item->>'payment_date', 10) END AS payment_date,
    r.payload->>'status' AS status, `+whtReversalMonthSQL+` AS reversal_month
  FROM gl_records r
  CROSS JOIN LATERAL jsonb_array_elements(CASE WHEN jsonb_typeof(r.payload->'details'->'withholdings') = 'array'
    THEN r.payload->'details'->'withholdings' ELSE '[]'::jsonb END) WITH ORDINALITY AS w(item, n)
  WHERE r.company = $1 AND r.kind = 'journals'
    AND r.payload->>'status' IN ('posted', 'reversed')
    AND NOT COALESCE((r.payload->>'isdeleted')::boolean, false)
    AND NOT (COALESCE(r.payload->>'kind', '') = ANY($6::text[]))
    AND r.payload->'details'->'withholdings' @> jsonb_build_array(jsonb_build_object('wht_direction', $2::int))
    AND w.item @> jsonb_build_object('wht_direction', $2::int)
    AND (COALESCE(cardinality($3::text[]), 0) = 0 OR w.item->>'form_type' = ANY($3::text[]))
    AND (LEFT(COALESCE(w.item->>'payment_date', ''), 7) = $4
      OR COALESCE(w.item->>'payment_date', '') !~ '^[0-9]{4}-[0-9]{2}-[0-9]{2}')
)
SELECT i.id, x.doc_no, x.doc_date, i.item, CASE WHEN i.status = 'reversed' THEN i.reversal_month ELSE '' END, COALESCE(p.payload->>'name_th', ''), COALESCE(p.payload->>'tax_id', ''), COALESCE(p.payload->>'address', ''),
  COALESCE(p.payload->>'title_name', ''), COALESCE(p.payload->>'addr_district', ''), COALESCE(p.payload->>'addr_province', ''), COALESCE(p.payload->>'addr_postcode', ''),
  COALESCE(p.payload->>'tax_branch_no', '')
FROM items i
LEFT JOIN LATERAL (
  SELECT l.doc_no, TO_CHAR(l.entry_date, 'YYYY-MM-DD') AS doc_date FROM gl_lines l
  WHERE l.company = $1 AND l.journal_id = i.id ORDER BY l.line_no LIMIT 1) first_line ON true
CROSS JOIN LATERAL (SELECT COALESCE(first_line.doc_no, i.docno, i.code) AS doc_no,
  COALESCE(first_line.doc_date, i.journal_date, '') AS doc_date) x
CROSS JOIN LATERAL (SELECT COALESCE(i.payment_date, x.doc_date) AS paid_date) pd
LEFT JOIN gl_subledger_partners p ON p.company = $1 AND p.code = i.item->>'partner_code'
WHERE LEFT(pd.paid_date, 7) = $4
  AND (i.status = 'posted' OR i.reversal_month > $4)
ORDER BY pd.paid_date, x.doc_date, i.id, i.n
LIMIT $5`, company, direction, pqArray(formTypes), fmt.Sprintf("%04d-%02d", year, month), whtReportMaxRows+1, pqArray(whtNonTaxJournalKinds))
	if err != nil {
		return nil, fmt.Errorf("read recorded withholdings: %w", err)
	}
	defer rows.Close()
	out := []TaxWithholdingRow{}
	for rows.Next() {
		var row TaxWithholdingRow
		var raw []byte
		if err := rows.Scan(&row.JournalID, &row.DocNo, &row.DocDate, &raw, &row.ReversedMonth, &row.PartnerName, &row.TaxID, &row.Address,
			&row.Title, &row.District, &row.Province, &row.Postcode, &row.BranchNo); err != nil {
			return nil, fmt.Errorf("scan recorded withholding: %w", err)
		}
		var item generalledger.SubledgerWithholding
		if err := json.Unmarshal(raw, &item); err != nil {
			return nil, fmt.Errorf("parse recorded withholding of %s: %w", row.JournalID, err)
		}
		// คำอธิบายว่างคงว่าง — handler/ใบแนบเติมชื่อประเภทเงินได้ตามภาษา (incomeTypeText) ไม่ฝังข้อความไทยที่นี่
		row.WithholdingID = item.ID
		applyWithholdingPartySnapshot(&row, item)
		row.PartnerCode, row.Description = item.PartnerCode, item.Description
		row.FormType, row.IncomeType, row.Condition = item.FormType, item.IncomeType, item.Condition
		row.PaidDate, row.CertificateNo, row.base = item.PaymentDate, item.CertificateNo, item.BaseAmount.Decimal()
		if item.TaxAmount != nil {
			row.wht = item.TaxAmount.Decimal()
		}
		out = append(out, finishWithholdingRow(row, moneyText(item.Rate.Decimal())))
	}
	return out, rows.Err()
}

// applyWithholdingPartySnapshot - ชื่อ/เลขภาษี/ที่อยู่ของคู่ค้าตามที่บันทึกในรายการ (snapshot ณ วันบันทึก ตาม wht.sql)
// ชนะทะเบียนคู่ค้าปัจจุบันทีละช่อง: แก้ทะเบียนภายหลังต้องไม่เปลี่ยนเลขภาษีในรายงาน/แบบยื่นของใบเก่า
// ทิศทาง 1 คู่ค้าเป็นผู้รับเงิน (payee), ทิศทาง 2 คู่ค้าเป็นผู้จ่ายเงิน (payer); ช่องที่ว่างใช้ค่าจากทะเบียนเหมือนเดิม
// คำนำหน้า/อำเภอ/จังหวัด/รหัสไปรษณีย์ไม่มีใน snapshot: ใช้ของทะเบียนเฉพาะเมื่อชื่อ/ที่อยู่ใน snapshot ยังตรงกับทะเบียน
// (ทะเบียนถูกแก้ภายหลัง = เป็นของชื่อ/ที่อยู่ใหม่ ผสมกับที่อยู่เดิมจะได้ที่อยู่ที่ไม่มีจริงในใบแนบ/ไฟล์ยื่น) — เว้นว่างให้ผู้ใช้กรอกในใบแนบ
func applyWithholdingPartySnapshot(row *TaxWithholdingRow, item generalledger.SubledgerWithholding) {
	party := item.Payee()
	if item.PartnerIsPayer() {
		party = item.Payer()
	}
	// snapshot ใหม่เก็บชื่อ/ที่อยู่เต็ม (คำนำหน้า + ชื่อ, ที่อยู่ + อำเภอ/จังหวัด/รหัสไปรษณีย์ — fillPartnerSnapshot); ตรงกับทะเบียนแบบเต็ม
	// = ข้อมูลเดียวกัน → แยกกลับเป็นชื่อ/ที่อยู่ตามทะเบียน + ช่องคำนำหน้า/อำเภอ/จังหวัดของใบแนบและไฟล์ยื่น (ไม่ให้คำนำหน้าซ้อนในชื่อ)
	if party.Name != "" {
		switch {
		case sameSpacedText(party.Name, row.PartnerName):
		case row.Title != "" && sameSpacedText(party.Name, generalledger.PartnerFullName(row.Title, row.PartnerName)):
			party.Name = row.PartnerName
		default:
			row.Title = ""
		}
		row.PartnerName = party.Name
	}
	if party.TaxID != "" {
		row.TaxID = party.TaxID
	}
	if party.BranchNo != "" {
		row.BranchNo = party.BranchNo
	}
	if party.Address != "" {
		switch {
		case sameSpacedText(party.Address, row.Address):
		case sameSpacedText(party.Address, generalledger.PartnerFullAddress(row.Address, row.District, row.Province, row.Postcode)):
			party.Address = row.Address
		default:
			row.District, row.Province, row.Postcode = "", "", ""
		}
		row.Address = party.Address
	}
}

// sameSpacedText - ข้อความเดียวกันเมื่อไม่นับช่องว่างซ้ำ/หัวท้าย
func sameSpacedText(a, b string) bool {
	return strings.Join(strings.Fields(a), " ") == strings.Join(strings.Fields(b), " ")
}

// summarizeWithholding - ยอดรวมทั้งงวด + สรุปตามอัตรา (สำหรับหน้าสรุปแบบยื่น) จากยอด decimal ของทุกแถว
func summarizeWithholding(rows []TaxWithholdingRow) TaxWithholdingSummary {
	baseTotal, whtTotal, netTotal := decimal.Zero, decimal.Zero, decimal.Zero
	payees := map[string]bool{}
	type group struct {
		count     int
		base, wht decimal.Decimal
	}
	groups := map[string]*group{}
	var order []string
	for _, r := range rows {
		baseTotal, whtTotal = baseTotal.Add(r.base), whtTotal.Add(r.wht)
		// สุทธิรวมเฉพาะแถวที่รู้ฐาน (ตรงกับยอดสุทธิที่แสดงรายแถว) — แถวฐาน 0 ไม่ดึงยอดรวมให้ติดลบ
		if withholdingNetKnown(r) {
			netTotal = netTotal.Add(r.base.Sub(r.wht))
		}
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
		NetTotal:     moneyText(netTotal),
		PayeeCount:   len(payees),
		ByRate:       byRate,
	}
}

// finishWithholdingRow - ชื่อแสดงผล (คำนำหน้า + ชื่อ) และยอดข้อความ/สุทธิ/อัตราที่บันทึก
func finishWithholdingRow(row TaxWithholdingRow, recordedRate string) TaxWithholdingRow {
	row.PartnerFullName = generalledger.PartnerFullName(row.Title, row.PartnerName)
	row.WhtAmount, row.BaseAmount = moneyText(row.wht), moneyText(row.base)
	row.WhtText = whtcert.BahtText(row.wht)
	row.NetAmount = ""
	if withholdingNetKnown(row) {
		row.NetAmount = moneyText(row.base.Sub(row.wht))
	}
	row.RatePercent = recordedRate
	return row
}

// withholdingNetKnown - ยอดจ่ายสุทธิรู้ได้เมื่อรู้ฐานเท่านั้น: ฐาน 0 (แบ่งฐานไม่ได้/ยังไม่ได้บันทึก) หรือฐานน้อยกว่ายอดหัก
// (ข้อมูลไม่สอดคล้อง) → สุทธิว่าง ไม่แสดงค่าติดลบที่ไม่มีความหมาย (UAT S14/S25 2026-09-24)
func withholdingNetKnown(row TaxWithholdingRow) bool {
	return row.base.Sign() > 0 && row.base.GreaterThanOrEqual(row.wht)
}

// pqArray - แปลงรายการรหัสบัญชีเป็น text[] สำหรับ ANY($n) (ใช้ lib/pq เวอร์ชันเดียวกับทั้งโปรเจกต์)
func pqArray(items []string) interface{} {
	return pq.StringArray(items)
}
