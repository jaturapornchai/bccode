package handlers

import (
	"context"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"smlcloudplatform/internal/generalledger"
	"smlcloudplatform/internal/goapi/language"
	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/mypg"

	"github.com/labstack/echo/v4"
	"github.com/shopspring/decimal"
)

// POST /api/report/tax/vat-summary — รายงานสรุปยอดภาษี (Champ เมนู 5539 `GLRepSumTaxView`: ภาษีซื้อ UNION ภาษีขาย + ยอดคงเหลือสะสม)
// อ่านชุดเดียวกับ ภ.พ.30 ของงวด (VatRecordsForPeriod: ภาษีขายของงวด + ภาษีซื้อที่ใช้สิทธิในงวด) และปัดต่อรายการแบบ sumPP30
// ยอดรวมจึงเท่ากับ ภ.พ.30 ข้อ 5/7 และข้อ 8/9 (label ใน internal/rdform/specs/pp30.json) เสมอ
// ต่างจาก Champ โดยตั้งใจ: ยอดคงเหลือของบรรทัดรวมรายวัน/รวมทั้งงวด = ภาษีซื้อ − ภาษีขายของช่วงนั้น
// (Champ บวกยอดคงเหลือสะสมของทุกแถวเข้าด้วยกัน ได้ตัวเลขที่ไม่มีความหมาย)

// vatSummaryMaxRows - เพดานรายการต่องวด (SME หลักสิบถึงหลักร้อยใบ) — ยอดคงเหลือสะสมต้องคำนวณทั้งงวดจึงไม่แบ่งหน้า
const vatSummaryMaxRows = 5000

// TaxVatSummaryRequest - งวดภาษี + การเรียง (แบบ Champ: วันที่ใบกำกับ / เลขที่ใบกำกับ / เลขที่เอกสาร)
type TaxVatSummaryRequest struct {
	HoldingCode  string `json:"holdingcode"`
	BusinessCode string `json:"businesscode"`
	Year         int    `json:"year"`
	Month        int    `json:"month"`
	Sort         string `json:"sort,omitempty"` // "date" (ค่าเริ่มต้น) | "taxno" | "docno"
}

// TaxVatSummaryRow - หนึ่งรายการภาษี: ภาษีซื้อหรือภาษีขายช่องใดช่องหนึ่ง ("" อีกช่อง); ใบลดหนี้ติดลบ
type TaxVatSummaryRow struct {
	No           int    `json:"no"`
	TaxDate      string `json:"taxdate"`
	TaxInvoiceNo string `json:"taxinvoiceno"`
	DocNo        string `json:"docno"`
	JournalID    string `json:"journalid"`
	Description  string `json:"description"`
	TaxIn        string `json:"taxin"`
	TaxOut       string `json:"taxout"`
	Balance      string `json:"balance"` // ภาษีซื้อ − ภาษีขาย สะสมถึงแถวนี้ (ติดลบ = ภาษีขายมากกว่า)
}

// TaxVatSummaryDay - รวมรายวัน (เฉพาะเรียงตามวันที่ แบบ Champ): Net = ภาษีซื้อ − ภาษีขายของวันนั้น
type TaxVatSummaryDay struct {
	Date   string `json:"date"`
	Count  int    `json:"count"`
	TaxIn  string `json:"taxin"`
	TaxOut string `json:"taxout"`
	Net    string `json:"net"`
}

// TaxVatSummaryTotals - ยอดทั้งงวด: TaxPayable = ภ.พ.30 ข้อ 8 (ข้อ 5 มากกว่าข้อ 7), TaxExcess = ข้อ 9 (ข้อ 5 น้อยกว่าข้อ 7)
type TaxVatSummaryTotals struct {
	Count      int    `json:"count"`
	TaxIn      string `json:"taxin"`
	TaxOut     string `json:"taxout"`
	Net        string `json:"net"`
	TaxPayable string `json:"taxpayable"`
	TaxExcess  string `json:"taxexcess"`
}

var vatSummarySorts = map[string]bool{"": true, "date": true, "taxno": true, "docno": true}

// TaxVatSummaryHandler - POST /api/report/tax/vat-summary
func TaxVatSummaryHandler(c echo.Context) error {
	var req TaxVatSummaryRequest
	if err := c.Bind(&req); err != nil {
		return taxReportFail(c, http.StatusBadRequest, "INVALID_PAYLOAD", "tax_form_payload_invalid")
	}
	holdingCode, businessCode, scopeErr := authenticatedCompanyContext(c, req.HoldingCode, req.BusinessCode)
	if scopeErr != nil {
		return taxScopeFail(c, scopeErr)
	}
	if !vatSummarySorts[req.Sort] {
		return taxReportFail(c, http.StatusBadRequest, "INVALID_SORT", "tax_report_sort_invalid")
	}
	if !isValidReportPeriod(req.Year, req.Month) {
		return taxReportFail(c, http.StatusBadRequest, "INVALID_PERIOD", "tax_form_period_invalid")
	}
	db, err := mypg.PgSqlFastConnect(holdingCode)
	if err != nil {
		logger.Error("TaxVatSummary: db: %v", err)
		return taxReportFail(c, http.StatusInternalServerError, "DB_CONNECTION_ERROR", "tax_report_failed")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := generalledger.EnsureSchema(ctx, db); err != nil {
		logger.Error("TaxVatSummary: ensure GL schema: %v", err)
		return taxReportFail(c, http.StatusInternalServerError, "QUERY_ERROR", "tax_report_failed")
	}
	purchases, err := generalledger.VatRecordsForPeriod(ctx, db, businessCode, req.Year, req.Month, 1)
	if err != nil {
		logger.Error("TaxVatSummary: purchases: %v", err)
		return taxReportFail(c, http.StatusInternalServerError, "QUERY_ERROR", "tax_report_failed")
	}
	sales, err := generalledger.VatRecordsForPeriod(ctx, db, businessCode, req.Year, req.Month, 2)
	if err != nil {
		logger.Error("TaxVatSummary: sales: %v", err)
		return taxReportFail(c, http.StatusInternalServerError, "QUERY_ERROR", "tax_report_failed")
	}
	if len(purchases)+len(sales) > vatSummaryMaxRows {
		return c.JSON(http.StatusUnprocessableEntity, map[string]any{"success": false, "code": "TOO_MANY_ROWS",
			"message": strings.ReplaceAll(language.Text("tax_report_too_many_rows", taxRequestLanguage(c)), "{max}", strconv.Itoa(vatSummaryMaxRows))})
	}
	rows, days, totals := buildVatSummary(purchases, sales, req.Sort, taxRequestLanguage(c))
	return c.JSON(http.StatusOK, map[string]any{"status": "success", "data": rows, "days": days, "summary": totals})
}

// vatSummaryItem - แถวก่อนจัดรูป: ยอดภาษีมีเครื่องหมายตามประเภทเอกสาร (ใบลดหนี้ติดลบ) ปัด 2 ตำแหน่งต่อรายการ
type vatSummaryItem struct {
	record generalledger.VatRecord
	input  bool
	amount decimal.Decimal
}

// buildVatSummary - รวมภาษีซื้อ + ภาษีขาย เรียงตาม sort แล้วคำนวณยอดคงเหลือสะสม รวมรายวัน และยอดทั้งงวด (decimal ทั้งหมด)
func buildVatSummary(purchases, sales []generalledger.VatRecord, sortBy, lang string) ([]TaxVatSummaryRow, []TaxVatSummaryDay, TaxVatSummaryTotals) {
	items := make([]vatSummaryItem, 0, len(purchases)+len(sales))
	for _, r := range purchases {
		items = append(items, vatSummaryItem{record: r, input: true, amount: vatAmountOf(r, vatSign(r.DocumentType))})
	}
	for _, r := range sales {
		items = append(items, vatSummaryItem{record: r, input: false, amount: vatAmountOf(r, vatSign(r.DocumentType))})
	}
	sort.SliceStable(items, func(i, j int) bool { return vatSummaryLess(items[i], items[j], sortBy) })

	rows := make([]TaxVatSummaryRow, 0, len(items))
	days := []TaxVatSummaryDay{}
	var balance, totalIn, totalOut, dayIn, dayOut decimal.Decimal
	dayCount := 0
	byDate := sortBy == "" || sortBy == "date"
	closeDay := func(date string) {
		if byDate && dayCount > 0 {
			days = append(days, TaxVatSummaryDay{Date: date, Count: dayCount, TaxIn: moneyText(dayIn), TaxOut: moneyText(dayOut), Net: moneyText(dayIn.Sub(dayOut))})
		}
		dayIn, dayOut, dayCount = decimal.Zero, decimal.Zero, 0
	}
	for i, it := range items {
		if i > 0 && it.record.TaxInvoiceDate != items[i-1].record.TaxInvoiceDate {
			closeDay(items[i-1].record.TaxInvoiceDate)
		}
		row := TaxVatSummaryRow{
			No:           i + 1,
			TaxDate:      it.record.TaxInvoiceDate,
			TaxInvoiceNo: it.record.TaxInvoiceNo,
			DocNo:        it.record.DocNo,
			JournalID:    it.record.JournalID,
			Description:  vatSummaryDescription(it.record, lang),
		}
		if it.input {
			balance, totalIn, dayIn = balance.Add(it.amount), totalIn.Add(it.amount), dayIn.Add(it.amount)
			row.TaxIn = moneyText(it.amount)
		} else {
			balance, totalOut, dayOut = balance.Sub(it.amount), totalOut.Add(it.amount), dayOut.Add(it.amount)
			row.TaxOut = moneyText(it.amount)
		}
		dayCount++
		row.Balance = moneyText(balance)
		rows = append(rows, row)
	}
	if len(items) > 0 {
		closeDay(items[len(items)-1].record.TaxInvoiceDate)
	}
	payable, excess := totalOut.Sub(totalIn), decimal.Zero
	if payable.IsNegative() {
		payable, excess = decimal.Zero, payable.Neg()
	}
	return rows, days, TaxVatSummaryTotals{
		Count: len(rows), TaxIn: moneyText(totalIn), TaxOut: moneyText(totalOut), Net: moneyText(totalIn.Sub(totalOut)),
		TaxPayable: moneyText(payable), TaxExcess: moneyText(excess),
	}
}

// vatSummaryLess - ลำดับแบบ Champ (GLRepSumTaxDlg ORDER BY): วันที่ → เลขที่ใบกำกับ → เลขที่เอกสาร / เลขที่ใบกำกับ → วันที่ → เอกสาร /
// เลขที่เอกสาร → วันที่ → เลขที่ใบกำกับ; ค่าเท่ากันทั้งหมดให้ภาษีซื้อก่อน แล้วตามใบสำคัญ + รายการ (ผลคงที่ทุกครั้ง)
func vatSummaryLess(a, b vatSummaryItem, sortBy string) bool {
	keys := func(it vatSummaryItem) []string {
		r := it.record
		switch sortBy {
		case "taxno":
			return []string{r.TaxInvoiceNo, r.TaxInvoiceDate, r.DocNo}
		case "docno":
			return []string{r.DocNo, r.TaxInvoiceDate, r.TaxInvoiceNo}
		}
		return []string{r.TaxInvoiceDate, r.TaxInvoiceNo, r.DocNo}
	}
	ka, kb := keys(a), keys(b)
	for i := range ka {
		if ka[i] != kb[i] {
			return ka[i] < kb[i]
		}
	}
	if a.input != b.input {
		return a.input
	}
	if a.record.JournalID != b.record.JournalID {
		return a.record.JournalID < b.record.JournalID
	}
	return a.record.ID < b.record.ID
}

// vatSummaryDescription - ชื่อคู่ค้า นำหน้าด้วยประเภทเอกสารเมื่อไม่ใช่ใบกำกับภาษี (Champ: ShortTaxDesc + ชื่อลูกหนี้/เจ้าหนี้)
func vatSummaryDescription(r generalledger.VatRecord, lang string) string {
	switch r.DocumentType {
	case 2:
		return strings.TrimSpace(language.Text("gl_vat_document_debit_note", lang) + " " + r.PartnerName)
	case 3:
		return strings.TrimSpace(language.Text("gl_vat_document_credit_note", lang) + " " + r.PartnerName)
	}
	return r.PartnerName
}
