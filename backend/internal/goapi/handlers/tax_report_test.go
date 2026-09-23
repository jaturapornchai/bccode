package handlers

import (
	"github.com/shopspring/decimal"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"smlcloudplatform/internal/generalledger"
	"smlcloudplatform/internal/whtcert"
	msmodels "smlcloudplatform/pkg/microservice/models"

	"github.com/labstack/echo/v4"
)

// -------------------- isValidReportPeriod --------------------

func TestIsValidReportPeriod(t *testing.T) {
	cases := []struct {
		name  string
		year  int
		month int
		want  bool
	}{
		{"valid", 2026, 9, true},
		{"month zero", 2026, 0, false},
		{"month 13", 2026, 13, false},
		{"month negative", 2026, -1, false},
		{"year too small", 1999, 1, false},
		{"year too large", 2101, 1, false},
		{"boundary month 1", 2026, 1, true},
		{"boundary month 12", 2026, 12, true},
		{"zero year and month", 0, 0, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isValidReportPeriod(tc.year, tc.month); got != tc.want {
				t.Errorf("isValidReportPeriod(%d,%d) = %v, want %v", tc.year, tc.month, got, tc.want)
			}
		})
	}
}

// -------------------- normalizeVatRegisterPaging --------------------

func TestNormalizeVatRegisterPaging(t *testing.T) {
	cases := []struct {
		name       string
		limit      int
		offset     int
		wantLimit  int
		wantOffset int
	}{
		{"defaults when zero", 0, 0, taxRegisterDefaultLimit, 0},
		{"defaults when negative limit", -5, 0, taxRegisterDefaultLimit, 0},
		{"defaults when over max", 9999, 0, taxRegisterDefaultLimit, 0},
		{"within bounds kept as-is", 50, 100, 50, 100},
		{"exactly max limit kept", taxRegisterMaxLimit, 0, taxRegisterMaxLimit, 0},
		{"negative offset clamped to zero", 50, -10, 50, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotLimit, gotOffset := normalizeVatRegisterPaging(tc.limit, tc.offset)
			if gotLimit != tc.wantLimit || gotOffset != tc.wantOffset {
				t.Errorf("normalizeVatRegisterPaging(%d,%d) = (%d,%d), want (%d,%d)",
					tc.limit, tc.offset, gotLimit, gotOffset, tc.wantLimit, tc.wantOffset)
			}
		})
	}
}

func TestThaiBahtText(t *testing.T) {
	cases := map[string]string{
		"0":          "ศูนย์บาทถ้วน",
		"1":          "หนึ่งบาทถ้วน",
		"11":         "สิบเอ็ดบาทถ้วน",
		"21":         "ยี่สิบเอ็ดบาทถ้วน",
		"101":        "หนึ่งร้อยเอ็ดบาทถ้วน",
		"1000000":    "หนึ่งล้านบาทถ้วน",
		"10000000":   "สิบล้านบาทถ้วน",
		"1250500.00": "หนึ่งล้านสองแสนห้าหมื่นห้าร้อยบาทถ้วน",
		"0.5":        "ห้าสิบสตางค์",
		"0.01":       "หนึ่งสตางค์",
		"1234.75":    "หนึ่งพันสองร้อยสามสิบสี่บาทเจ็ดสิบห้าสตางค์",
		"3210":       "สามพันสองร้อยสิบบาทถ้วน",
		"-500":       "ลบห้าร้อยบาทถ้วน",
	}
	for in, want := range cases {
		if got := whtcert.BahtText(decimal.RequireFromString(in)); got != want {
			t.Errorf("whtcert.BahtText(%s) = %q, want %q", in, got, want)
		}
	}
}

// -------------------- VAT register / ภ.พ.30 จากรายละเอียดภาษีมูลค่าเพิ่มในใบสำคัญ --------------------

func vatRecord(doc, invoice string, documentType int, base, zero, exempt, vat string) generalledger.VatRecord {
	amount := generalledger.Amount(vat)
	return generalledger.VatRecord{JournalID: "J-" + doc, DocNo: doc, DocDate: "2026-09-10", SubledgerVat: generalledger.SubledgerVat{
		ID: doc + invoice, TaxType: 2, DocumentType: documentType, TaxInvoiceNo: invoice, TaxInvoiceDate: "2026-09-10",
		PartnerName: "บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด", PartnerTaxID: "0105558012349", PartnerBranchNo: "00000",
		BaseAmount: generalledger.Amount(base), ZeroRateAmount: generalledger.Amount(zero), ExemptAmount: generalledger.Amount(exempt),
		Rate: "7", VatAmount: &amount,
	}}
}

// ใบลดหนี้หักออก ใบเพิ่มหนี้บวกเพิ่ม และยอดรวมท้ายรายงาน = ผลบวกของแถว (ปัด 2 ตำแหน่งต่อรายการ)
func TestBuildVatRegisterSignsAndTotals(t *testing.T) {
	records := []generalledger.VatRecord{
		vatRecord("UV1", "IV001", 1, "1000", "0", "0", "70"),
		vatRecord("UV2", "EX001", 1, "0", "500", "0", "0"),
		vatRecord("UV3", "CN001", 3, "100", "0", "0", "7"),
		vatRecord("UV4", "DN001", 2, "0.105", "0", "0", "0.0074"),
	}
	rows, summary := buildVatRegister(records)
	if len(rows) != 4 {
		t.Fatalf("rows = %+v", rows)
	}
	credit := rows[2]
	if credit.TaxInvoiceNo != "CN001" || credit.AmountBeforeVat != "-100.00" || credit.VatAmount != "-7.00" || credit.TotalAmount != "-107.00" {
		t.Fatalf("credit note row = %+v", credit)
	}
	if debit := rows[3]; debit.AmountBeforeVat != "0.11" || debit.VatAmount != "0.01" || debit.TotalAmount != "0.12" {
		t.Fatalf("debit note row = %+v", debit)
	}
	if rows[0].DocDate != "2026-09-10" || rows[0].CounterpartyName == "" || rows[0].TaxID != "0105558012349" || rows[0].BranchNo != "00000" {
		t.Fatalf("invoice row = %+v", rows[0])
	}
	if summary.AmountBeforeVat != "1400.11" || summary.VatAmount != "63.01" || summary.TotalAmount != "1463.12" {
		t.Fatalf("summary = %+v", summary)
	}
	if page := pageVatRegisterRows(rows, 2, 3); len(page) != 1 || page[0].TaxInvoiceNo != "DN001" {
		t.Fatalf("page = %+v", page)
	}
	if page := pageVatRegisterRows(rows, 2, 9); len(page) != 0 {
		t.Fatalf("offset past rows = %+v", page)
	}
	if empty, s := buildVatRegister(nil); len(empty) != 0 || s.TotalAmount != "0.00" {
		t.Fatalf("empty register = %+v %+v", empty, s)
	}
}

// ภ.พ.30: ข้อ 2/3/4 แยกจากยอดที่บันทึก, ข้อ 5/7 = VAT ที่บันทึก (ไม่คำนวณจากฐาน x 7%), ใบลดหนี้หักออก
func TestSumPP30FromRecordedVat(t *testing.T) {
	sales := []generalledger.VatRecord{
		vatRecord("UV1", "IV001", 1, "1000", "0", "0", "70.01"),
		vatRecord("UV2", "EX001", 1, "0", "500", "250", "0"),
		vatRecord("UV3", "CN001", 3, "100", "0", "0", "7"),
	}
	purchases := []generalledger.VatRecord{
		vatRecord("SV1", "PI001", 1, "400", "0", "0", "28"),
		vatRecord("SV2", "PC001", 3, "40", "0", "0", "2.80"),
	}
	totals := sumPP30(sales, purchases)
	got := map[string]string{"taxable": moneyText(totals.salesTaxable), "zero": moneyText(totals.salesZeroRated), "exempt": moneyText(totals.salesExempt),
		"output": moneyText(totals.outputVat), "purchase": moneyText(totals.purchaseTaxable), "input": moneyText(totals.inputVat)}
	want := map[string]string{"taxable": "900.00", "zero": "500.00", "exempt": "250.00", "output": "63.01", "purchase": "360.00", "input": "25.20"}
	for key, value := range want {
		if got[key] != value {
			t.Fatalf("%s = %s want %s (all %v)", key, got[key], value, got)
		}
	}
	missing := vatRecord("UV9", "IV009", 1, "10", "0", "0", "0")
	missing.VatAmount = nil
	if total := sumPP30([]generalledger.VatRecord{missing}, nil); !total.outputVat.IsZero() {
		t.Fatalf("missing vat amount must count as zero, got %s", total.outputVat)
	}
}

// -------------------- HTTP-level validation (pre-DB) --------------------

var taxReportTestUser = msmodels.UserInfo{Username: "tester", HoldingCode: "h1", BusinessCode: "b1", UID: "u1"}

func callTaxReportHandler(t *testing.T, handler echo.HandlerFunc, body string, user *msmodels.UserInfo) *httptest.ResponseRecorder {
	t.Helper()
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	if user != nil {
		ctx.Set("UserInfo", *user)
	}
	if err := handler(ctx); err != nil {
		t.Fatalf("handler error: %v", err)
	}
	return rec
}

func TestTaxVatRegisterHandler_Unauthorized(t *testing.T) {
	rec := callTaxReportHandler(t, TaxVatRegisterHandler, `{"year":2026,"month":9,"type":"sale"}`, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestTaxVatRegisterHandler_InvalidType(t *testing.T) {
	rec := callTaxReportHandler(t, TaxVatRegisterHandler, `{"year":2026,"month":9,"type":"refund"}`, &taxReportTestUser)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "INVALID_TYPE") {
		t.Errorf("expected INVALID_TYPE code, got body=%s", rec.Body.String())
	}
}

func TestTaxVatRegisterHandler_InvalidPeriod(t *testing.T) {
	rec := callTaxReportHandler(t, TaxVatRegisterHandler, `{"year":2026,"month":13,"type":"sale"}`, &taxReportTestUser)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "INVALID_PERIOD") {
		t.Errorf("expected INVALID_PERIOD code, got body=%s", rec.Body.String())
	}
}

func TestTaxVatRegisterHandler_InvalidPayload(t *testing.T) {
	rec := callTaxReportHandler(t, TaxVatRegisterHandler, `{invalid-json`, &taxReportTestUser)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "INVALID_PAYLOAD") {
		t.Errorf("expected INVALID_PAYLOAD code, got body=%s", rec.Body.String())
	}
}
