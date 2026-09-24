package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"smlcloudplatform/internal/generalledger"
	"smlcloudplatform/internal/goapi/language"
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
	// DuplicateDocNos - เลขที่ใบสำคัญที่บันทึกใบกำกับภาษีฉบับเดียวกัน (ผู้ออก + เลขที่ + วันที่) — มีเลขที่ใบสำคัญของแถวนี้เอง = ซ้ำในใบสำคัญเดียวกัน; [] = ไม่ซ้ำ; เตือนให้ตรวจ ไม่บล็อก
	DuplicateDocNos []string `json:"duplicatedocnos"`
}

// TaxVatRegisterSummary - ยอดรวมทั้งงวด (ไม่ใช่เฉพาะหน้าที่แสดง) คำนวณใน PostgreSQL
type TaxVatRegisterSummary struct {
	AmountBeforeVat string `json:"amountbeforevat"`
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

	if !isValidReportPeriod(req.Year, req.Month) {
		return taxReportFail(c, http.StatusBadRequest, "INVALID_PERIOD", "tax_form_period_invalid")
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
	records, err := generalledger.VatRecordsForPeriod(ctx, db, businessCode, req.Year, req.Month, taxType)
	if err != nil {
		logger.Error("TaxVatRegister: %v", err)
		return taxReportFail(c, http.StatusInternalServerError, "QUERY_ERROR", "tax_report_failed")
	}

	// ยอดรวมท้ายรายงานเป็นของทั้งงวด ไม่ใช่ผลรวมเฉพาะหน้าที่ browser ได้รับ
	rows, summary := buildVatRegister(records)
	data := pageVatRegisterRows(rows, limit, offset)
	return c.JSON(http.StatusOK, map[string]any{
		"status":  "success",
		"data":    data,
		"count":   len(data),
		"total":   len(rows),
		"summary": summary,
		"limit":   limit,
		"offset":  offset,
	})
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
// มูลค่าก่อนภาษี = ฐานภาษี + ยอดอัตรา 0% + ยอดยกเว้น ของใบกำกับ (มูลค่าสินค้า/บริการ ไม่รวม VAT)
func buildVatRegister(records []generalledger.VatRecord) ([]TaxVatRegisterRow, TaxVatRegisterSummary) {
	rows := make([]TaxVatRegisterRow, 0, len(records))
	var sumBefore, sumVat decimal.Decimal
	for _, r := range records {
		sign := vatSign(r.DocumentType)
		before := vatMoney(r.BaseAmount, sign).Add(vatMoney(r.ZeroRateAmount, sign)).Add(vatMoney(r.ExemptAmount, sign))
		vat := vatAmountOf(r, sign)
		rows = append(rows, TaxVatRegisterRow{
			DocDate:          r.TaxInvoiceDate,
			TaxInvoiceNo:     r.TaxInvoiceNo,
			CounterpartyName: r.PartnerName,
			TaxID:            r.PartnerTaxID,
			BranchNo:         r.PartnerBranchNo,
			AmountBeforeVat:  moneyText(before),
			VatAmount:        moneyText(vat),
			TotalAmount:      moneyText(before.Add(vat)),
			DuplicateDocNos:  append([]string{}, r.DuplicateDocNos...),
		})
		sumBefore, sumVat = sumBefore.Add(before), sumVat.Add(vat)
	}
	return rows, TaxVatRegisterSummary{
		AmountBeforeVat: moneyText(sumBefore),
		VatAmount:       moneyText(sumVat),
		TotalAmount:     moneyText(sumBefore.Add(sumVat)),
		DuplicateCount:  countDuplicateInvoices(records),
	}
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
	Forms        []string `json:"forms,omitempty"` // กรองตามแบบยื่น เช่น ["53"], ["3"] — ว่าง = ทุกแบบที่ระบบค้นพบ
	Limit        int      `json:"limit,omitempty"`
	Offset       int      `json:"offset,omitempty"`
}

// TaxWithholdingRow - แถวภาษีหัก ณ ที่จ่ายต่อใบสำคัญ (ข้อมูลจริงจาก GL และหลักฐานประกอบ)
type TaxWithholdingRow struct {
	JournalID string `json:"journalid"`
	// id ของรายการใน details.withholdings (เฉพาะแถวที่บันทึก) — จอ 50 ทวิ ใช้จับคู่รายการแบบตรงตัว ไม่ต้องเดาจากยอด
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
	BranchNo    string `json:"branchno,omitempty"`
	District    string `json:"district,omitempty"`
	Province    string `json:"province,omitempty"`
	Postcode    string `json:"postcode,omitempty"`
	Description string `json:"description"`
	BaseAmount  string `json:"baseamount"`    // ทศนิยม 2 ตำแหน่งแบบ string
	WhtAmount   string `json:"whtamount"`     // ยอดหักที่บันทึกจริงในบรรทัด GL
	WhtText     string `json:"whtamounttext"` // ยอดหักเป็นตัวอักษรไทย สำหรับหนังสือรับรอง 50 ทวิ
	NetAmount   string `json:"netamount"`     // ยอดจ่ายสุทธิ = ฐาน - ยอดหัก ("" เมื่อไม่รู้ฐาน — withholdingNetKnown)
	RatePercent string `json:"ratepercent"`   // อัตรา = ยอดหัก / ฐาน × 100 ปัด 2 ตำแหน่ง ("" เมื่อไม่มีฐาน)
	// ฐานภาษีมาจากไหน: recorded = ผู้ใช้บันทึกในรายละเอียดใบสำคัญ (แก้ได้เสมอ), inferred = ระบบประมาณจากบรรทัดบัญชี
	TaxBaseSource string `json:"taxbasesource"`
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
	// UnknownForm - ยอดภาษีหักที่ระบบไม่รู้แบบยื่น (บัญชีภาษีหักที่ชื่ออ้างหลายแบบ/แบ่งให้แบบใดไม่ได้) ของทิศทางเราหัก:
	// แสดงเฉพาะรายงานทุกแบบ (แบบว่าง) และไม่ถูกเติมลงแบบ ภ.ง.ด. ใด ๆ — ผู้ใช้ต้องบันทึกรายละเอียดพร้อมเลือกแบบเอง
	UnknownForm int
}

// whtReportMaxRows - เพดานรายการต่อเดือน (ภาษีหักของ SME ต่อเดือนหลักสิบถึงหลักร้อยใบ) กันงานหนักผิดปกติ
const whtReportMaxRows = 5000

// TaxWithholdingHandler - POST /api/report/tax/wht
// อ่านยอดหัก ณ ที่จ่ายจากบัญชีภาษีหักในบัญชีแยกประเภทที่ผ่านรายการจริง (gl_lines)
// แล้วโยงคู่ค้า/เลขผู้เสียภาษีจากหลักฐานประกอบ (gl_subledger_partners + documents/settlements)
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

// withholdingReportNotes - หมายเหตุของรายงานตามภาษาผู้ใช้: ไม่พบบัญชี, ยอดที่ไม่รู้แบบยื่น, รายการที่กลับรายการในเดือนหลัง
func withholdingReportNotes(report taxWithholdingReport, lang string) []string {
	notes := []string{}
	if report.NoteKey != "" {
		notes = append(notes, language.Text(report.NoteKey, lang))
	}
	count := func(key string, n int) string {
		return strings.ReplaceAll(language.Text(key, lang), "{count}", strconv.Itoa(n))
	}
	if report.UnknownForm > 0 {
		notes = append(notes, count("tax_wht_note_form_unknown", report.UnknownForm))
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

// validWhtForms - แบบยื่นที่ระบบรองรับการจับคู่บัญชี (จับจากชื่อบัญชีที่มีรูปแบบ "ภ.ง.ด.<แบบ>")
// ไม่มี ภ.ง.ด.1 — ระบบเงินเดือนอยู่นอกขอบเขตผลิตภัณฑ์ (AGENTS.md "ไม่ทำระบบเงินเดือน")
var validWhtForms = map[string]bool{"2": true, "3": true, "53": true, "54": true}

// defaultWhtForms - แบบยื่นทั้งหมดของฝั่ง "เราหักผู้อื่น" เมื่อไม่ระบุแบบ
var defaultWhtForms = []string{"2", "3", "53", "54"}

// whtFormPrefix - คำนำหน้าแบบยื่นในชื่อบัญชีภาษีหักฝั่งเราหัก เช่น "ภาษีเงินได้หัก ณ ที่จ่ายค้างจ่าย - ภ.ง.ด.53"
const whtFormPrefix = "ภ.ง.ด."

// whtReceivedPatterns - ชื่อบัญชีภาษีที่บริษัทถูกหัก (ฝั่ง received) — รับทั้ง "ภาษีถูกหัก" และ "…ถูกหัก ณ ที่จ่าย" ของผังมาตรฐาน
var whtReceivedPatterns = []string{"ภาษีถูกหัก", "ถูกหัก ณ ที่จ่าย"}

// บรรทัดที่ไม่ใช่ฐานภาษีหัก ณ ที่จ่าย ตัดด้วยชื่อ + ประเภทบัญชีในผังบัญชีเท่านั้น (ห้ามใช้รหัสบัญชี — ผังแต่ละบริษัทต่างกัน):
//   - whtAccountWords: บัญชีภาษีหัก/ถูกหัก ณ ที่จ่ายเอง (ทุกประเภทบัญชี)
//   - vatAccountWords: บัญชีภาษีมูลค่าเพิ่ม (ภาษีซื้อ/ภาษีขาย) เพราะฐานภาษีหักไม่รวม VAT — ตัดเฉพาะบัญชีสินทรัพย์/หนี้สิน
//     เพราะบัญชีรายได้/ค่าใช้จ่ายในผังจริงมีชื่อแบบ "รายได้จากการให้บริการ - มีภาษีมูลค่าเพิ่ม 7%" ซึ่งเป็นฐานภาษี ไม่ใช่บรรทัดภาษี
//     (UAT S14 2026-09-24: ตัดบัญชีรายได้ทิ้งทำให้ฐาน 0 และยอดสุทธิติดลบ)
//   - บัญชีภาษีหักที่ค้นพบจากชื่อแบบยื่นของทั้งสองทิศทาง (whtScope.taxAccounts) — ชื่ออย่าง "ภ.ง.ด.53 ค้างจ่าย" ไม่มีคำใน whtAccountWords
var (
	whtAccountWords = []string{"หัก ณ ที่จ่าย", "ถูกหัก"}
	vatAccountWords = []string{"ภาษีซื้อ", "ภาษีขาย", "ภาษีมูลค่าเพิ่ม"}
)

// whtReversalMonthSQL - เดือน (YYYY-MM) ของใบกลับรายการของใบต้นฉบับ r ที่สถานะ reversed ( = ไม่ใช่ใบที่ถูกกลับ/หาใบกลับไม่พบ)
// งวดภาษีหัก ณ ที่จ่าย = เดือนที่จ่ายเงินได้ (ประกาศอธิบดีกรมสรรพากร เกี่ยวกับภาษีเงินได้ ฉบับที่ 111 https://www.rd.go.th/9041.html
// "ภายในเจ็ดวันนับแต่วันสิ้นเดือนของเดือนที่จ่ายเงินได้พึงประเมิน") จึงตัดสินรายเดือนที่ยื่น:
//   - กลับรายการในเดือนที่ยื่นเดียวกันหรือก่อนหน้า → ภาษีของใบนั้นไม่เคยต้องยื่น ไม่แสดงทั้งต้นฉบับและใบกลับ
//   - กลับรายการในเดือนหลัง → แบบของเดือนเดิมยื่นไปแล้ว ต้นฉบับคงอยู่ในเดือนของตัวเอง เดือนที่กลับไม่มีแถวติดลบ;
//     การแก้ภาษีที่ยื่นแล้วทำนอกระบบ: นำส่งเกิน/ผิด/ซ้ำ ขอคืนด้วยคำร้อง ค.10 ภายใน 3 ปี (ประมวลรัษฎากร มาตรา 27 ตรี
//     https://www.rd.go.th/5943.html, แบบ ค.10 https://www.rd.go.th/fileadmin/tax_pdf/others/K10_061260.pdf) — หมายเหตุ tax_wht_note_reversed_later
//
// ใบกลับรายการไม่มี details.withholdings และบรรทัดบัญชีภาษีหักของใบกลับอยู่ฝั่งตรงข้าม จึงไม่เกิดแถวเองอยู่แล้ว
const whtReversalMonthSQL = `COALESCE((SELECT LEFT(MIN(rv.payload->>'date'), 7) FROM gl_records rv
  WHERE r.payload->>'status' = 'reversed' AND rv.company = r.company AND rv.kind = 'journals'
    AND rv.payload->>'reversalof' = r.id AND rv.payload->>'kind' = 'reversal'
    AND NOT COALESCE((rv.payload->>'isdeleted')::boolean, false)), '')`

// whtNonTaxJournalKinds - ใบที่เกิดจากการหักภาษีจริงเท่านั้นที่เข้ารายงาน: ใบกลับรายการหักล้างต้นฉบับ,
// ใบยอดยกมา/ปิดบัญชีแค่ยกยอดคงเหลือของงวดก่อน (หนี้ภาษีหักค้างจ่ายของ ธ.ค. ปีก่อนไม่ใช่การหักภาษีใหม่ของ ม.ค. — UAT S24 2026-09-24)
var whtNonTaxJournalKinds = []string{"reversal", "opening", "closing"}

// whtAccount - บัญชีภาษีหักที่ค้นพบจากชื่อในผังบัญชี (ห้ามใช้รหัสบัญชีเป็นเงื่อนไข — ผังแต่ละบริษัทต่างกัน)
type whtAccount struct {
	code  string
	forms []string // แบบยื่นที่ชื่อบัญชีอ้างถึง เช่น ["53"]; ว่าง = ฝั่งถูกหัก (ไม่มีแบบยื่นของเรา)
}

// formType - แบบยื่นของบัญชีตามชื่อ ("ภ.ง.ด.53" → PND53); ชื่อที่ไม่อ้างแบบหรืออ้างหลายแบบ = "" (ไม่รู้แบบ ไม่เดา)
func (a whtAccount) formType() string {
	if len(a.forms) != 1 {
		return ""
	}
	return "PND" + a.forms[0]
}

// whtScope - ทิศทาง/แบบที่ขอของรายงาน + บัญชีภาษีหักที่ค้นพบจากผังบัญชี
type whtScope struct {
	direction   string
	want        int             // wht_direction ของรายการที่บันทึก: 1 = paid (เราหัก), 2 = received (เราถูกหัก)
	accounts    []whtAccount    // บัญชีภาษีหักของทิศทางนี้ทุกแบบ (ใช้ตัดสินว่าใบมีภาษีหักหลายบัญชีหรือไม่ แม้ขอดูแบบเดียว)
	formTypes   map[string]bool // แบบที่ขอ เช่น PND53 (เฉพาะ paid)
	allForms    bool            // ขอทุกแบบ (หรือฝั่ง received) — แถวที่ไม่รู้แบบแสดงได้เฉพาะกรณีนี้
	taxAccounts []string        // บัญชีภาษีหักของทั้งสองทิศทาง — ไม่ใช่ฐานภาษีของแถว inferred
}

// wantsForm - แถวของแบบนี้อยู่ในรายงานที่ขอหรือไม่ ("" = ไม่รู้แบบ → เฉพาะรายงานทุกแบบ ไม่ยัดเข้าแบบใดแบบหนึ่งเอง)
func (s whtScope) wantsForm(formType string) bool {
	return s.allForms || (formType != "" && s.formTypes[formType])
}

// wantsAccount - บรรทัดของบัญชีนี้อยู่ในรายงานที่ขอหรือไม่ (บัญชีที่ชื่ออ้างแบบที่ขออย่างน้อยหนึ่งแบบ)
func (s whtScope) wantsAccount(a whtAccount) bool {
	if s.allForms {
		return true
	}
	for _, f := range a.forms {
		if s.formTypes["PND"+f] {
			return true
		}
	}
	return false
}

// buildWithholdingReport - คำนวณรายการหัก ณ ที่จ่ายทั้งงวดจาก GL จริง (แยกออกจาก handler เพื่อทดสอบ integration ตรงกับ *sql.DB)
//
// หลักการอ่านข้อมูล (ตรวจกับโครง runtime จริง backend/internal/generalledger/schema.sql):
//   - อ่านใบที่ผ่านรายการ (posted) และใบที่ถูกกลับรายการ (reversed) ตามกติกา whtReversalMonthSQL; ไม่อ่านใบกลับรายการเอง
//     ใบยอดยกมา และใบปิดบัญชี (whtNonTaxJournalKinds) — กลับรายการภายในเดือนที่ยื่นเดียวกัน = หักล้างกันหมด ไม่ยื่นทั้งคู่,
//     กลับรายการในเดือนหลัง = ต้นฉบับยังอยู่ในเดือนของตัวเอง (ยื่นไปแล้ว) พร้อม ReversedMonth และเดือนที่กลับไม่มีแถวติดลบ
//   - รายการภาษีหักที่ผู้ใช้บันทึก (details.withholdings) = หลักฐานรายฉบับ ใช้ฐาน/ยอดหัก/อัตราที่บันทึกตรง ๆ หนึ่งแถวต่อรายการ
//     และเข้างวดตาม "เดือนที่จ่ายเงินได้" (payment_date) ตามกำหนดยื่นของกรมสรรพากร (docs/kms/21-thai-tax-form-references.md §9)
//   - ใบที่ไม่ได้บันทึกรายการทิศทางนี้เลย ประมาณจากบรรทัดบัญชีภาษีหัก (inferred ตามวันที่ใบ)
//   - ใบที่บันทึกไว้บางส่วน: ยอดในบัญชีภาษีหักที่เกินยอดที่บันทึก (ทุกวันที่จ่าย) เป็นแถว inferred ฐานไม่ทราบ ให้ยอดรวมตรงกับบัญชี
//     (withholdingRemainders) — บันทึกครบแล้วไม่มีแถวประมาณซ้ำ แม้แบบที่บันทึกจะต่างจากแบบของบัญชี หรือวันที่จ่ายอยู่คนละเดือน
//   - บัญชีภาษีหัก = บัญชี liability ที่ชื่อภาษาไทยอ้างแบบยื่น เช่น "ภ.ง.ด.53"; ฝั่ง received คือบัญชีที่ชื่อมีคำว่า
//     "ภาษีถูกหัก" หรือ "ถูกหัก ณ ที่จ่าย" — ชั่วคราวจนกว่าจะมีตาราง wht_records ตาม mydocs/datamodels/gl/wht.sql
//   - ยอดหักเป็นยอดที่บันทึกจริงเสมอ ไม่คำนวณย้อนจากอัตรา; ทุกยอดเป็น decimal ไม่มี float
func buildWithholdingReport(ctx context.Context, db *sql.DB, company string, year, month int, direction string, forms []string) (taxWithholdingReport, error) {
	report := taxWithholdingReport{Rows: []TaxWithholdingRow{}, Summary: TaxWithholdingSummary{ByRate: []TaxWithholdingRateGroup{}}}
	var formPatterns []string
	for _, f := range defaultWhtForms {
		formPatterns = append(formPatterns, whtFormPrefix+f)
	}
	// บัญชีภาษีหักของทั้งสองทิศทาง: ทิศทางที่ขอใช้หาบรรทัดภาษี ทั้งหมดใช้ตัดออกจากฐานภาษีของแถว inferred
	paidAccounts, err := findWithholdingAccounts(ctx, db, company, "liability", formPatterns)
	if err != nil {
		return report, err
	}
	receivedAccounts, err := findWithholdingAccounts(ctx, db, company, "", whtReceivedPatterns)
	if err != nil {
		return report, err
	}
	scope := whtScope{direction: direction, want: 1, accounts: paidAccounts, formTypes: map[string]bool{}}
	for _, a := range append(append([]whtAccount{}, paidAccounts...), receivedAccounts...) {
		scope.taxAccounts = append(scope.taxAccounts, a.code)
	}
	var formTypes []string
	if direction == "received" {
		scope.want, scope.accounts, scope.allForms = 2, receivedAccounts, true
	} else {
		if len(forms) == 0 {
			forms = defaultWhtForms
		}
		for _, f := range forms {
			if !validWhtForms[f] {
				return report, fmt.Errorf("invalid form %q", f)
			}
			formTypes = append(formTypes, "PND"+f)
			scope.formTypes["PND"+f] = true
		}
		scope.allForms = len(scope.formTypes) == len(defaultWhtForms)
	}

	recorded, err := recordedWithholdingRows(ctx, db, company, year, month, scope.want, formTypes)
	if err != nil {
		return report, err
	}
	inferred, unknownForm, err := inferredWithholdingRows(ctx, db, company, year, month, scope)
	if err != nil {
		return report, err
	}
	report.UnknownForm = unknownForm
	report.Rows = append(recorded, inferred...)
	if len(report.Rows) > whtReportMaxRows {
		return report, fmt.Errorf("wht rows exceed %d for %04d-%02d", whtReportMaxRows, year, month)
	}
	// เรียงตามวันที่จ่ายเงิน (วันที่ใบเมื่อไม่มี) — ใบแนบ ภ.ง.ด. พิมพ์วันที่จ่ายตามลำดับนี้ (UAT S10/S12 2026-09-24)
	sort.SliceStable(report.Rows, func(a, b int) bool {
		ra, rb := report.Rows[a], report.Rows[b]
		if dateA, dateB := firstNonEmpty(ra.PaidDate, ra.DocDate), firstNonEmpty(rb.PaidDate, rb.DocDate); dateA != dateB {
			return dateA < dateB
		}
		if ra.DocDate != rb.DocDate {
			return ra.DocDate < rb.DocDate
		}
		return ra.JournalID < rb.JournalID
	})
	requested := 0
	for _, a := range scope.accounts {
		if scope.wantsAccount(a) {
			requested++
		}
	}
	if requested == 0 && len(report.Rows) == 0 {
		report.NoteKey = "tax_wht_note_no_accounts"
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
		row.PartnerCode, row.Description, row.TaxBaseSource = item.PartnerCode, item.Description, "recorded"
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

// inferredWithholdingRows - แถวประมาณจากบรรทัดบัญชีภาษีหัก ตามวันที่ใบ:
//   - ใบที่ไม่ได้บันทึกรายการภาษีหักทิศทางนี้เลย: หนึ่งแถวต่อบัญชีภาษีหักต่อใบ (หลายบรรทัดบัญชีเดียวกันรวมเป็นแถวเดียว)
//     ใบที่มีภาษีหักมากกว่าหนึ่งบัญชี แบ่งฐานให้แต่ละบัญชีไม่ได้โดยไม่เดา → ฐาน 0 อัตราว่าง
//     (หมายเหตุ tax_form_note_inferred_base บอกให้บันทึกรายละเอียดภาษีหัก)
//     ใบที่ฝั่งตรงข้ามมีแต่บัญชีภาษี (เช่น โอนยอดจากบัญชี ภ.ง.ด.3 ไปบัญชี ภ.ง.ด.53) = ย้ายยอดระหว่างบัญชีภาษี ไม่ใช่การหักภาษีใหม่
//     → ไม่มีแถว (UAT S25/S26 2026-09-24: ใบโอนยอดเคยเป็นแถวฐาน 0 ทำให้ยอดภาษีในแบบนับซ้ำ)
//   - ใบที่บันทึกรายการไว้แล้ว: เฉพาะส่วนที่บัญชีภาษีหักเกินยอดที่บันทึก (withholdingRemainders) ฐานไม่ทราบ ไม่มีคู่ค้า
//     — เดิมทิ้งส่วนที่ยังไม่บันทึกไปเงียบ ๆ ยอดในแบบจึงน้อยกว่าบัญชี
func inferredWithholdingRows(ctx context.Context, db *sql.DB, company string, year, month int, s whtScope) ([]TaxWithholdingRow, int, error) {
	out := []TaxWithholdingRow{}
	unknownForm := 0
	if len(s.accounts) == 0 {
		return out, unknownForm, nil
	}
	byCode := map[string]whtAccount{}
	codes := make([]string, 0, len(s.accounts))
	for _, a := range s.accounts {
		byCode[a.code] = a
		codes = append(codes, a.code)
	}
	suppress, opposite := "credit", "debit"
	if s.direction == "received" {
		suppress, opposite = "debit", "credit"
	}
	rows, err := db.QueryContext(ctx, `
SELECT l.journal_id, l.doc_no, TO_CHAR(l.entry_date, 'YYYY-MM-DD'), l.account_code, l.`+suppress+`, l.description,
  COALESCE(r.payload->'details'->'withholdings' @> jsonb_build_array(jsonb_build_object('wht_direction', $5::int)), false),
  CASE WHEN r.payload->>'status' = 'reversed' THEN rev.reversal_month ELSE '' END
FROM gl_lines l
JOIN gl_records r ON r.company = l.company AND r.kind = 'journals' AND r.id = l.journal_id
CROSS JOIN LATERAL (SELECT `+whtReversalMonthSQL+` AS reversal_month) rev
WHERE l.company = $1
  AND (r.payload->>'status' = 'posted' OR (r.payload->>'status' = 'reversed' AND rev.reversal_month > TO_CHAR(l.entry_date, 'YYYY-MM')))
  AND NOT (COALESCE(r.payload->>'kind', '') = ANY($7::text[]))
  AND l.account_code = ANY($2)
  AND l.`+suppress+` > 0
  AND EXTRACT(YEAR FROM l.entry_date) = $3
  AND EXTRACT(MONTH FROM l.entry_date) = $4
ORDER BY l.entry_date, l.journal_id, l.line_no
LIMIT $6`, company, pqArray(codes), year, month, s.want, whtReportMaxRows+1, pqArray(whtNonTaxJournalKinds))
	if err != nil {
		return nil, 0, fmt.Errorf("read wht lines: %w", err)
	}
	type journalLines struct {
		journalID, docNo, docDate string
		reversedMonth             string   // กลับรายการในเดือนหลัง (whtReversalMonthSQL)
		recorded                  bool     // ใบบันทึกรายการภาษีหักทิศทางนี้ไว้แล้ว (อย่างน้อยหนึ่งรายการ)
		accounts                  []string // ตามลำดับบรรทัดแรกที่พบ
		amount                    map[string]decimal.Decimal
		description               map[string]string
	}
	var order []*journalLines
	byJournal := map[string]*journalLines{}
	count := 0
	for rows.Next() {
		var journalID, docNo, docDate, account, description, reversedMonth string
		var amount decimal.Decimal
		var recorded bool
		if err := rows.Scan(&journalID, &docNo, &docDate, &account, &amount, &description, &recorded, &reversedMonth); err != nil {
			rows.Close()
			return nil, 0, fmt.Errorf("scan wht line: %w", err)
		}
		count++
		j := byJournal[journalID]
		if j == nil {
			j = &journalLines{journalID: journalID, docNo: docNo, docDate: docDate, reversedMonth: reversedMonth, recorded: recorded,
				amount: map[string]decimal.Decimal{}, description: map[string]string{}}
			byJournal[journalID] = j
			order = append(order, j)
		}
		if _, seen := j.amount[account]; !seen {
			j.accounts = append(j.accounts, account)
			j.description[account] = description
		}
		j.amount[account] = j.amount[account].Add(amount)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("read wht lines: %w", err)
	}
	if count > whtReportMaxRows {
		return nil, 0, fmt.Errorf("wht lines exceed %d for %04d-%02d", whtReportMaxRows, year, month)
	}
	var plainIDs, recordedIDs []string
	for _, j := range order {
		if j.recorded {
			recordedIDs = append(recordedIDs, j.journalID)
		} else {
			plainIDs = append(plainIDs, j.journalID)
		}
	}
	counters, err := nonTaxCounterTotals(ctx, db, company, plainIDs, opposite, s.taxAccounts)
	if err != nil {
		return nil, 0, err
	}
	recordedTax, err := recordedTaxByForm(ctx, db, company, recordedIDs, s.want)
	if err != nil {
		return nil, 0, err
	}
	// ทิศทางเราหัก: ยอดที่ไม่รู้แบบ (ชื่อบัญชีอ้างหลายแบบ/แบ่งให้แบบใดไม่ได้) ห้ามเติมลงแบบใดเอง — นับให้ผู้ใช้บันทึกรายละเอียดและเลือกแบบ
	// ฝั่งถูกหักไม่มีแบบยื่นของเรา แบบว่างจึงเป็นเรื่องปกติ ไม่นับ
	noteUnknown := func(form string, relevant bool) {
		if form == "" && s.want == 1 && relevant {
			unknownForm++
		}
	}
	for _, j := range order {
		if j.recorded {
			// ยอดบัญชีแยกตามแบบของบัญชี (ฝั่ง received ไม่มีแบบ → รวมเป็นกลุ่มเดียว "")
			credit := map[string]decimal.Decimal{}
			relevant := false
			for _, code := range j.accounts {
				form := byCode[code].formType()
				credit[form] = credit[form].Add(j.amount[code])
				relevant = relevant || s.wantsAccount(byCode[code])
			}
			remainders := withholdingRemainders(credit, recordedTax[j.journalID])
			emitted := map[string]bool{}
			emit := func(form, description string) {
				amount, ok := remainders[form]
				if !ok || emitted[form] {
					return
				}
				emitted[form] = true
				noteUnknown(form, relevant)
				if !s.wantsForm(form) {
					return
				}
				out = append(out, finishWithholdingRow(TaxWithholdingRow{JournalID: j.journalID, DocNo: j.docNo, DocDate: j.docDate,
					Description: description, TaxBaseSource: "inferred", FormType: form, ReversedMonth: j.reversedMonth, wht: amount}, ""))
			}
			for _, code := range j.accounts {
				emit(byCode[code].formType(), j.description[code])
			}
			emit("", j.description[j.accounts[0]])
			continue
		}
		counter := counters[j.journalID]
		if counter.movement.Sign() <= 0 {
			continue // ย้ายยอดระหว่างบัญชีภาษี (สินทรัพย์/หนี้สิน) ไม่ใช่การหักภาษีใหม่
		}
		for _, code := range j.accounts {
			account := byCode[code]
			// แบบของแถว = แบบที่ชื่อบัญชีอ้างเพียงแบบเดียว; บัญชีที่ชื่ออ้างหลายแบบ (เช่น "ภ.ง.ด.3/ภ.ง.ด.53") แบบว่าง
			// → แสดงเฉพาะรายงานทุกแบบ ไม่เติมลงทั้ง ภ.ง.ด.3 และ ภ.ง.ด.53 (ยื่นภาษีซ้ำสองแบบ — review 2026-09-24)
			form := account.formType()
			noteUnknown(form, s.wantsAccount(account))
			if !s.wantsForm(form) {
				continue
			}
			row := TaxWithholdingRow{JournalID: j.journalID, DocNo: j.docNo, DocDate: j.docDate, Description: j.description[code],
				TaxBaseSource: "inferred", FormType: form, ReversedMonth: j.reversedMonth, wht: j.amount[code]}
			if err := fillWithholdingEvidence(ctx, db, company, j.journalID, s.direction, len(j.accounts) == 1, counter.base, &row); err != nil {
				return nil, 0, err
			}
			out = append(out, finishWithholdingRow(row, ""))
		}
	}
	return out, unknownForm, nil
}

// recordedTaxByForm - ยอดภาษีที่บันทึกในรายละเอียดของแต่ละใบ (ทุกวันที่จ่าย) แยกตามแบบ — อ่านครั้งเดียวทุกใบ ไม่ query ทีละใบ
// ฝั่ง received (want 2) รวมเป็นกลุ่มเดียว ("") เพราะบัญชีภาษีถูกหักไม่มีแบบยื่นของเรา; รวมยอดใน PostgreSQL เป็น numeric (ไม่มี float)
func recordedTaxByForm(ctx context.Context, db *sql.DB, company string, journalIDs []string, want int) (map[string]map[string]decimal.Decimal, error) {
	out := map[string]map[string]decimal.Decimal{}
	if len(journalIDs) == 0 {
		return out, nil
	}
	rows, err := db.QueryContext(ctx, `
SELECT r.id, CASE WHEN $3::int = 2 THEN '' ELSE COALESCE(w.item->>'form_type', '') END AS form,
  COALESCE(SUM(NULLIF(w.item->>'tax_amount', '')::numeric), 0)
FROM gl_records r
CROSS JOIN LATERAL jsonb_array_elements(CASE WHEN jsonb_typeof(r.payload->'details'->'withholdings') = 'array'
  THEN r.payload->'details'->'withholdings' ELSE '[]'::jsonb END) AS w(item)
WHERE r.company = $1 AND r.kind = 'journals' AND r.id = ANY($2)
  AND w.item @> jsonb_build_object('wht_direction', $3::int)
GROUP BY r.id, form`, company, pqArray(journalIDs), want)
	if err != nil {
		return nil, fmt.Errorf("read recorded wht by form: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var journalID, form string
		var tax decimal.Decimal
		if err := rows.Scan(&journalID, &form, &tax); err != nil {
			return nil, fmt.Errorf("scan recorded wht by form: %w", err)
		}
		if out[journalID] == nil {
			out[journalID] = map[string]decimal.Decimal{}
		}
		out[journalID][form] = tax
	}
	return out, rows.Err()
}

// withholdingRemainders - ภาษีในบัญชีภาษีหักของใบที่เกินยอดที่บันทึกไว้ (= ส่วนที่ยังไม่ได้บันทึกรายละเอียด) แยกตามแบบ
// credit/recorded key = แบบยื่น (PND3 ฯลฯ; "" = ไม่รู้แบบ) — ไม่เดาฐาน ไม่เดาแบบ:
//   - ยอดบัญชีรวมไม่เกินยอดที่บันทึกรวม → ไม่มีส่วนเหลือ (รวมกรณีบันทึก ภ.ง.ด.53 แต่ลงบัญชี ภ.ง.ด.3 — UAT S25 ห้ามนับซ้ำ)
//   - ส่วนเกินรายแบบรวมได้เท่าส่วนเกินรวม → แยกตามแบบได้ชัด
//   - มีแบบเดียวที่ยอดบัญชีเกินยอดที่บันทึก → ส่วนเกินรวมเป็นของแบบนั้น (ส่วนที่เหลือของแบบนั้นถูกบันทึกเป็นแบบอื่น)
//   - มากกว่านั้นแบ่งให้แบบใดไม่ได้ → แถวเดียว แบบว่าง (แสดงเฉพาะรายงานทุกแบบ) ให้ผู้ใช้บันทึกรายละเอียดเอง
func withholdingRemainders(credit, recorded map[string]decimal.Decimal) map[string]decimal.Decimal {
	total := decimal.Zero
	for _, v := range credit {
		total = total.Add(v)
	}
	for _, v := range recorded {
		total = total.Sub(v)
	}
	if !total.IsPositive() {
		return nil
	}
	short, sum := map[string]decimal.Decimal{}, decimal.Zero
	for form, v := range credit {
		if excess := v.Sub(recorded[form]); excess.IsPositive() {
			short[form], sum = excess, sum.Add(excess)
		}
	}
	if sum.Equal(total) {
		return short
	}
	if len(short) == 1 {
		for form := range short {
			return map[string]decimal.Decimal{form: total}
		}
	}
	return map[string]decimal.Decimal{"": total}
}

// whtCounter - ยอดฝั่งตรงข้ามของบรรทัดภาษีหักในใบหนึ่ง (เดบิตของใบจ่าย / เครดิตของใบรับ):
//   - base: ไม่รวมบัญชีภาษีทุกชนิด (ภาษีหักที่ค้นพบ taxAccounts + ชื่อมีคำใน whtAccountWords ทุกประเภทบัญชี + บัญชี VAT
//     ที่เป็นสินทรัพย์/หนี้สิน) = ฐานภาษีเมื่อไม่มีการตัดยอด
//   - movement: ไม่รวมเฉพาะบัญชีภาษีที่เป็นสินทรัพย์/หนี้สิน (taxAccounts + สินทรัพย์/หนี้สินที่ชื่อมีคำภาษีหักหรือ VAT)
//     ≤ 0 = ใบย้ายยอดระหว่างบัญชีภาษีเท่านั้น (เช่น โอนจากบัญชี ภ.ง.ด.3 ไป ภ.ง.ด.53) ไม่ใช่การหักภาษีใหม่
//     ใบที่ฝั่งตรงข้ามเป็นค่าใช้จ่ายภาษีที่บริษัทออกให้ (เงื่อนไข 2/3 เช่น "ภาษีเงินได้หัก ณ ที่จ่ายออกแทน") มี movement > 0
//     จึงยังเป็นแถว (ฐานไม่ทราบ) — เดิมถูกตัดทิ้งเพราะชื่อค่าใช้จ่ายมีคำ "หัก ณ ที่จ่าย" (review 2026-09-24)
type whtCounter struct {
	base, movement decimal.Decimal
}

// nonTaxCounterTotals - whtCounter ต่อใบ — อ่านครั้งเดียวทุกใบของงวด ไม่ query ทีละใบ; ใบที่ไม่มีบรรทัดฝั่งตรงข้ามเลยไม่อยู่ใน map (= 0)
// แยกบัญชีภาษีด้วยชื่อ + ประเภทบัญชีในผังเท่านั้น (ห้ามใช้รหัสบัญชีเป็นเงื่อนไข — ผังแต่ละบริษัทต่างกัน)
func nonTaxCounterTotals(ctx context.Context, db *sql.DB, company string, journalIDs []string, opposite string, taxAccounts []string) (map[string]whtCounter, error) {
	totals := map[string]whtCounter{}
	if len(journalIDs) == 0 {
		return totals, nil
	}
	rows, err := db.QueryContext(ctx, `
SELECT l.journal_id,
  COALESCE(SUM(l.`+opposite+`) FILTER (WHERE NOT tax.discovered AND NOT tax.wht_named AND NOT (tax.balance_sheet AND tax.vat_named)), 0),
  COALESCE(SUM(l.`+opposite+`) FILTER (WHERE NOT tax.discovered AND NOT (tax.balance_sheet AND (tax.wht_named OR tax.vat_named))), 0)
FROM gl_lines l
CROSS JOIN LATERAL (
  SELECT l.account_code = ANY(COALESCE($5::text[], '{}'::text[])) AS discovered,
    COALESCE(bool_or(EXISTS (SELECT 1 FROM unnest($3::text[]) f WHERE strpos(n->>'name', f) > 0)), false) AS wht_named,
    COALESCE(bool_or(EXISTS (SELECT 1 FROM unnest($4::text[]) f WHERE strpos(n->>'name', f) > 0)), false) AS vat_named,
    COALESCE(bool_or(a.payload->>'accounttype' IN ('asset', 'liability')), false) AS balance_sheet
  FROM gl_records a
  CROSS JOIN LATERAL jsonb_array_elements(CASE WHEN jsonb_typeof(a.payload->'names') = 'array' THEN a.payload->'names' ELSE '[]'::jsonb END) n
  WHERE a.company = l.company AND a.kind = 'accounts' AND a.payload->>'accountcode' = l.account_code
) tax
WHERE l.company = $1 AND l.journal_id = ANY($2)
GROUP BY l.journal_id`, company, pqArray(journalIDs), pqArray(whtAccountWords), pqArray(vatAccountWords), pqArray(taxAccounts))
	if err != nil {
		return nil, fmt.Errorf("wht counter side: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var journalID string
		var c whtCounter
		if err := rows.Scan(&journalID, &c.base, &c.movement); err != nil {
			return nil, fmt.Errorf("scan wht counter side: %w", err)
		}
		totals[journalID] = c
	}
	return totals, rows.Err()
}

// findWithholdingAccounts - บัญชีภาษีหักจากผังบัญชีจริง (ค้นจากชื่อบัญชี) พร้อมแบบยื่นที่ชื่ออ้างถึง (คำ "ภ.ง.ด.<แบบ>" ที่ตรง)
func findWithholdingAccounts(ctx context.Context, db *sql.DB, company, accountType string, patterns []string) ([]whtAccount, error) {
	rows, err := db.QueryContext(ctx, `
SELECT a.payload->>'accountcode', array_agg(DISTINCT f.pattern ORDER BY f.pattern)
FROM gl_records a
CROSS JOIN LATERAL jsonb_array_elements(CASE WHEN jsonb_typeof(a.payload->'names') = 'array' THEN a.payload->'names' ELSE '[]'::jsonb END) n
CROSS JOIN LATERAL unnest($3::text[]) AS f(pattern)
WHERE a.company = $1 AND a.kind = 'accounts'
  AND (a.payload->>'accounttype' = $2 OR $2 = '')
  AND strpos(n->>'name', f.pattern) > 0
GROUP BY a.payload->>'accountcode'
ORDER BY 1`, company, accountType, pqArray(patterns))
	if err != nil {
		return nil, fmt.Errorf("find wht accounts: %w", err)
	}
	defer rows.Close()
	accounts := []whtAccount{}
	for rows.Next() {
		var code string
		var matched pq.StringArray
		if err := rows.Scan(&code, &matched); err != nil {
			return nil, fmt.Errorf("scan wht account: %w", err)
		}
		account := whtAccount{code: code}
		for _, pattern := range matched {
			if form, ok := strings.CutPrefix(pattern, whtFormPrefix); ok {
				account.forms = append(account.forms, form)
			}
		}
		accounts = append(accounts, account)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read wht accounts: %w", err)
	}
	return accounts, nil
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

// fillWithholdingEvidence - โยงคู่ค้าและฐานภาษีของใบจากหลักฐานประกอบจริง
// baseKnown = false (ใบมีภาษีหักหลายบัญชี) → ไม่ใส่ฐาน: ยอดฝั่งตรงข้ามของทั้งใบแบ่งให้แต่ละบัญชีไม่ได้โดยไม่เดา
// counterBase = ยอดฝั่งตรงข้ามที่ไม่ใช่บัญชีภาษีของใบ (nonTaxCounterTotals) ใช้เป็นฐานเมื่อไม่มีการตัดยอด
func fillWithholdingEvidence(ctx context.Context, db *sql.DB, company, journalID, direction string, baseKnown bool, counterBase decimal.Decimal, row *TaxWithholdingRow) error {
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
	switch {
	case !baseKnown:
	case settleBase.Sign() > 0:
		row.base = settleBase
	default:
		// ไม่มีการตัดยอด: ใช้ผลรวมฝั่งตรงข้ามของบรรทัดภาษีหักเป็นฐาน (เช่น เดบิตค่าใช้จ่ายในใบจ่าย)
		// ไม่รวมบัญชีภาษีหัก/ถูกหัก และบัญชีภาษีมูลค่าเพิ่ม — แยกจากชื่อ + ประเภทบัญชีในผังเท่านั้น (nonTaxCounterTotals)
		row.base = counterBase
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

// fillPartnerMaster - ชื่อ/เลขผู้เสียภาษี/ที่อยู่ (+คำนำหน้า อำเภอ จังหวัด รหัสไปรษณีย์ ถ้ามี) ของคู่ค้าจากทะเบียนคู่ค้าของห้องบัญชี
func fillPartnerMaster(ctx context.Context, db *sql.DB, company, partnerCode string, row *TaxWithholdingRow) error {
	row.PartnerCode = partnerCode
	err := db.QueryRowContext(ctx, `SELECT COALESCE(payload->>'name_th',''), COALESCE(payload->>'tax_id',''), COALESCE(payload->>'address',''),
COALESCE(payload->>'title_name',''), COALESCE(payload->>'addr_district',''), COALESCE(payload->>'addr_province',''), COALESCE(payload->>'addr_postcode',''),
COALESCE(payload->>'tax_branch_no','')
FROM gl_subledger_partners WHERE company=$1 AND code=$2`, company, partnerCode).Scan(&row.PartnerName, &row.TaxID, &row.Address,
		&row.Title, &row.District, &row.Province, &row.Postcode, &row.BranchNo)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("load partner master: %w", err)
	}
	return nil
}

// finishWithholdingRow - ชื่อแสดงผล (คำนำหน้า + ชื่อ) และยอดข้อความ/สุทธิ/อัตรา; อัตราที่บันทึกไว้ชนะอัตราที่คำนวณย้อนจากยอด
func finishWithholdingRow(row TaxWithholdingRow, recordedRate string) TaxWithholdingRow {
	row.PartnerFullName = generalledger.PartnerFullName(row.Title, row.PartnerName)
	row.WhtAmount, row.BaseAmount = moneyText(row.wht), moneyText(row.base)
	row.WhtText = whtcert.BahtText(row.wht)
	row.NetAmount = ""
	if withholdingNetKnown(row) {
		row.NetAmount = moneyText(row.base.Sub(row.wht))
	}
	switch {
	case recordedRate != "":
		row.RatePercent = recordedRate
	case row.base.Sign() > 0:
		row.RatePercent = moneyText(row.wht.Mul(decimal.NewFromInt(100)).Div(row.base))
	}
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
