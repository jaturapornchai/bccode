package handlers

import (
	"encoding/json"
	"github.com/shopspring/decimal"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"smlcloudplatform/internal/generalledger"
	"smlcloudplatform/internal/goapi/language"
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
		{"clamps to max when over max", 9999, 0, taxRegisterMaxLimit, 0},
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

// callTaxHandlerLang - เรียก handler พร้อมภาษา (query ?lang หรือ Accept-Language) เพื่อตรวจข้อความตามภาษาผู้ใช้
func callTaxHandlerLang(t *testing.T, handler echo.HandlerFunc, target, acceptLanguage, body string, user *msmodels.UserInfo) *httptest.ResponseRecorder {
	t.Helper()
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	if acceptLanguage != "" {
		req.Header.Set("Accept-Language", acceptLanguage)
	}
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

// TestTaxReportErrorsFollowLanguage - ข้อความผิดพลาดมาจาก languages.tsv ตามภาษาผู้ใช้ (รหัสและ HTTP status เดิม)
func TestTaxReportErrorsFollowLanguage(t *testing.T) {
	cases := []struct {
		name, target, lang, body string
		handler                  echo.HandlerFunc
		user                     *msmodels.UserInfo
		status                   int
		code, key                string
		want                     string
	}{
		{"vat type en", "/", "en-US,en;q=0.9", `{"year":2026,"month":9,"type":"refund"}`, TaxVatRegisterHandler, &taxReportTestUser, 400, "INVALID_TYPE", "tax_report_type_invalid", "en"},
		{"vat period th default", "/", "", `{"year":2026,"month":13,"type":"sale"}`, TaxVatRegisterHandler, &taxReportTestUser, 400, "INVALID_PERIOD", "tax_form_period_invalid", "th"},
		{"wht direction query lang wins", "/?lang=en", "th", `{"year":2026,"month":9,"direction":"sideways"}`, TaxWithholdingHandler, &taxReportTestUser, 400, "INVALID_TYPE", "tax_report_direction_invalid", "en"},
		// ภ.ง.ด.1 = เงินเดือน อยู่นอกขอบเขตผลิตภัณฑ์ (AGENTS.md) ต้องถูกปฏิเสธ
		{"wht payroll form rejected", "/", "th", `{"year":2026,"month":9,"direction":"paid","forms":["1"]}`, TaxWithholdingHandler, &taxReportTestUser, 400, "INVALID_FORM", "tax_report_form_invalid", "th"},
		{"wht payload en", "/", "en", `{bad`, TaxWithholdingHandler, &taxReportTestUser, 400, "INVALID_PAYLOAD", "tax_form_payload_invalid", "en"},
		{"unauthorized en", "/", "en", `{"year":2026,"month":9,"type":"sale"}`, TaxVatRegisterHandler, nil, 401, "UNAUTHORIZED", "unauthorized", "en"},
		{"form compute unauthorized th", "/", "th", `{"code":"pnd53"}`, TaxFormComputeHandler, nil, 401, "UNAUTHORIZED", "unauthorized", "th"},
		// 50 ทวิ: ข้อความขอบเขตบริษัทเดิมเป็นภาษาเดียว และอ่านแค่ Accept-Language — ต้องแปลและให้ ?lang ชนะแบบรายงานอื่น
		{"wht certificate unauthorized en", "/", "en", whtCertBody, WhtCertificateHandler, nil, 401, "UNAUTHORIZED", "unauthorized", "en"},
		{"wht certificate payload query lang wins", "/?lang=en", "th", `{bad`, WhtCertificateHandler, &taxReportTestUser, 400, "wht_cert_payload_invalid", "wht_cert_payload_invalid", "en"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := callTaxHandlerLang(t, tc.handler, tc.target, tc.lang, tc.body, tc.user)
			var body struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			want := language.Text(tc.key, tc.want)
			if rec.Code != tc.status || body.Code != tc.code || body.Message != want || want == tc.key {
				t.Fatalf("status=%d code=%s message=%q, want %d %s %q", rec.Code, body.Code, body.Message, tc.status, tc.code, want)
			}
		})
	}
	if language.Text("tax_report_form_invalid", "en") == language.Text("tax_report_form_invalid", "th") {
		t.Fatal("en and th texts must differ (real translation)")
	}
}

// TestTaxFormComputeReturnsRecomputedTotals - /compute ใช้ prepareTaxDocument ตัวเดียวกับ save/pdf
func TestTaxFormComputeReturnsRecomputedTotals(t *testing.T) {
	body := `{"code":"pnd53","document":{"values":{"total_tax":"1.00"},"rows":[{"l1_amount":"100","l1_tax":"3"},{"l1_amount":"200.50","l1_tax":"6.02"}]}}`
	rec := callTaxHandlerLang(t, TaxFormComputeHandler, "/", "th", body, &taxReportTestUser)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"total_tax":"9.02"`) {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

// UAT S14/S25 2026-09-24: แถวที่ไม่รู้ฐาน (ฐาน 0) ต้องไม่มียอดสุทธิติดลบ และยอดสุทธิรวมนับเฉพาะแถวที่รู้ฐาน
func TestWithholdingNetBlankWhenBaseUnknown(t *testing.T) {
	known := finishWithholdingRow(TaxWithholdingRow{JournalID: "J1", base: decimal.RequireFromString("10000"), wht: decimal.RequireFromString("300")}, "")
	unknown := finishWithholdingRow(TaxWithholdingRow{JournalID: "J2", base: decimal.Zero, wht: decimal.RequireFromString("200")}, "")
	inconsistent := finishWithholdingRow(TaxWithholdingRow{JournalID: "J3", base: decimal.RequireFromString("100"), wht: decimal.RequireFromString("150")}, "3")
	if known.NetAmount != "9700.00" || known.RatePercent != "3.00" {
		t.Fatalf("known = %+v", known)
	}
	if unknown.NetAmount != "" || unknown.BaseAmount != "0.00" || unknown.RatePercent != "" {
		t.Fatalf("unknown base = %+v", unknown)
	}
	if inconsistent.NetAmount != "" {
		t.Fatalf("wht above base must not give a negative net: %+v", inconsistent)
	}
	s := summarizeWithholding([]TaxWithholdingRow{known, unknown, inconsistent})
	if s.BaseTotal != "10100.00" || s.WhtTotal != "650.00" || s.NetTotal != "9700.00" {
		t.Fatalf("summary = %+v", s)
	}
}

// ใบกำกับที่ซ้ำกับใบสำคัญอื่น: แถวบอกเลขที่ใบสำคัญอื่น, ไม่ซ้ำ = [] (ไม่ใช่ null — จอวนลูปได้เลย), สรุปนับจำนวนแถวที่ซ้ำ
func TestBuildVatRegisterDuplicateWarnings(t *testing.T) {
	dup := vatRecord("SV1", "IV001", 1, "1000", "0", "0", "70")
	dup.DuplicateDocNos = []string{"SV7"}
	rows, summary := buildVatRegister([]generalledger.VatRecord{dup, vatRecord("SV2", "IV002", 1, "500", "0", "0", "35")})
	if len(rows) != 2 || strings.Join(rows[0].DuplicateDocNos, ",") != "SV7" || rows[1].DuplicateDocNos == nil || len(rows[1].DuplicateDocNos) != 0 {
		t.Fatalf("rows = %+v", rows)
	}
	if summary.DuplicateCount != 1 {
		t.Fatalf("duplicate count = %d", summary.DuplicateCount)
	}
	raw, err := json.Marshal(map[string]any{"row": rows[1], "summary": summary})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"duplicatedocnos":[]`) || !strings.Contains(string(raw), `"duplicatecount":1`) {
		t.Fatalf("contract JSON = %s", raw)
	}
}

// ส่วนที่บัญชีภาษีหักเกินยอดที่บันทึก: ไม่เดาฐาน ไม่เดาแบบ และห้ามนับซ้ำเมื่อบันทึกครบแต่ลงบัญชีคนละแบบ (UAT S25)
func TestWithholdingRemainders(t *testing.T) {
	d := func(v string) decimal.Decimal { return decimal.RequireFromString(v) }
	show := func(m map[string]decimal.Decimal) string {
		out := []string{}
		for _, form := range []string{"", "PND2", "PND3", "PND53"} {
			if v, ok := m[form]; ok {
				out = append(out, form+"="+moneyText(v))
			}
		}
		return strings.Join(out, ",")
	}
	cases := []struct {
		name             string
		credit, recorded map[string]decimal.Decimal
		want             string
	}{
		{"fully recorded", map[string]decimal.Decimal{"PND53": d("400")}, map[string]decimal.Decimal{"PND53": d("400")}, ""},
		{"partly recorded same form", map[string]decimal.Decimal{"PND53": d("400")}, map[string]decimal.Decimal{"PND53": d("300")}, "PND53=100.00"},
		{"recorded PND53 but posted to PND3 account", map[string]decimal.Decimal{"PND3": d("200")}, map[string]decimal.Decimal{"PND53": d("200")}, ""},
		{"two forms each short", map[string]decimal.Decimal{"PND3": d("200"), "PND53": d("400")}, map[string]decimal.Decimal{"PND53": d("300")}, "PND3=200.00,PND53=100.00"},
		{"cross-posted plus one short form", map[string]decimal.Decimal{"PND3": d("300")}, map[string]decimal.Decimal{"PND53": d("200")}, "PND3=100.00"},
		{"ambiguous split", map[string]decimal.Decimal{"PND3": d("300"), "PND53": d("200")}, map[string]decimal.Decimal{"PND2": d("250")}, "=250.00"},
		{"received side (no form)", map[string]decimal.Decimal{"": d("600")}, map[string]decimal.Decimal{"": d("450.50")}, "=149.50"},
		{"over-recorded", map[string]decimal.Decimal{"PND53": d("100")}, map[string]decimal.Decimal{"PND53": d("150")}, ""},
	}
	for _, tc := range cases {
		if got := show(withholdingRemainders(tc.credit, tc.recorded)); got != tc.want {
			t.Errorf("%s: got %q want %q", tc.name, got, tc.want)
		}
	}
}

// ชื่อประเภทเงินได้มาจาก languages.tsv ตามภาษาผู้ใช้ — คำอธิบายที่ผู้ใช้บันทึกชนะเสมอ
func TestIncomeTypeTextFollowsLanguage(t *testing.T) {
	row := TaxWithholdingRow{IncomeType: "40_2"}
	if th, en := incomeTypeText(row, "th"), incomeTypeText(row, "en"); th != "ค่านายหน้า 40(2)" || en != "Commission 40(2)" {
		t.Fatalf("40_2 th=%q en=%q", th, en)
	}
	if got := incomeTypeText(TaxWithholdingRow{IncomeType: "40_4b_1_1"}, "en"); got != "Dividend 40(4)(b)" {
		t.Fatalf("dividend sub-type = %q", got)
	}
	if got := incomeTypeText(TaxWithholdingRow{IncomeType: "3_tres", Description: "ค่าขนส่ง"}, "en"); got != "ค่าขนส่ง" {
		t.Fatalf("recorded description must win, got %q", got)
	}
	if got := incomeTypeText(TaxWithholdingRow{IncomeType: "other"}, "th"); got != "" {
		t.Fatalf("unknown code = %q", got)
	}
}

// รายงานภาษีหักของใบที่บันทึก snapshot ต้องใช้ชื่อ/เลขภาษี/ที่อยู่ตามที่บันทึกก่อนทะเบียนคู่ค้าปัจจุบัน (ทีละช่อง)
func TestApplyWithholdingPartySnapshotPrefersRecordedParty(t *testing.T) {
	registry := TaxWithholdingRow{PartnerName: "ชื่อในทะเบียนใหม่", TaxID: "0105599999999", Address: "ที่อยู่ในทะเบียนใหม่"}

	paid := registry
	applyWithholdingPartySnapshot(&paid, generalledger.SubledgerWithholding{Direction: 1, PayeeName: "ห้างหุ้นส่วนจำกัด รุ่งเรืองการค้าไทย", PayeeTaxID: "0103555012345", PayerName: "บริษัทเรา"})
	if paid.PartnerName != "ห้างหุ้นส่วนจำกัด รุ่งเรืองการค้าไทย" || paid.TaxID != "0103555012345" || paid.Address != "ที่อยู่ในทะเบียนใหม่" {
		t.Fatalf("direction 1 must use the payee snapshot per field, got %+v", paid)
	}

	received := registry
	applyWithholdingPartySnapshot(&received, generalledger.SubledgerWithholding{Direction: 2, PayerTaxID: "0105561234567", PayeeName: "บริษัทเรา"})
	if received.PartnerName != "ชื่อในทะเบียนใหม่" || received.TaxID != "0105561234567" {
		t.Fatalf("direction 2 must use the payer snapshot per field, got %+v", received)
	}

	blank := registry
	applyWithholdingPartySnapshot(&blank, generalledger.SubledgerWithholding{Direction: 1})
	if blank != registry {
		t.Fatalf("an empty snapshot must keep the registry values, got %+v", blank)
	}
}

// adversarial review 2026-09-24: ทะเบียนคู่ค้าบุคคลธรรมดาย้ายที่อยู่/เปลี่ยนชื่อภายหลัง — ใบแนบ ภ.ง.ด.3 ห้ามผสมที่อยู่เดิมกับจังหวัด/รหัสไปรษณีย์ใหม่
func TestApplyWithholdingPartySnapshotDoesNotMixRegistryAddressParts(t *testing.T) {
	moved := TaxWithholdingRow{PartnerName: "สมชาย ใจดี", Title: "นาย", Address: "99 ถนนห้วยแก้ว", District: "เมืองเชียงใหม่", Province: "เชียงใหม่", Postcode: "50200"}
	applyWithholdingPartySnapshot(&moved, generalledger.SubledgerWithholding{Direction: 1, PayeeName: "สมชาย ใจดี", PayeeAddress: "12/3 ถนนงามวงศ์วาน"})
	if moved.Address != "12/3 ถนนงามวงศ์วาน" || moved.District != "" || moved.Province != "" || moved.Postcode != "" || moved.Title != "นาย" {
		t.Fatalf("old snapshot address must not take the new province/postcode: %+v", moved)
	}

	renamed := TaxWithholdingRow{PartnerName: "สมหญิง ใจดี", Title: "นาง", Address: "12/3 ถนนงามวงศ์วาน", District: "เมืองนนทบุรี", Province: "นนทบุรี", Postcode: "11000"}
	applyWithholdingPartySnapshot(&renamed, generalledger.SubledgerWithholding{Direction: 1, PayeeName: "สมหญิง รักไทย", PayeeAddress: " 12/3  ถนนงามวงศ์วาน"})
	if renamed.Title != "" || renamed.PartnerName != "สมหญิง รักไทย" {
		t.Fatalf("title of the new registry name must not prefix the old recorded name: %+v", renamed)
	}
	if renamed.District != "เมืองนนทบุรี" || renamed.Province != "นนทบุรี" || renamed.Postcode != "11000" {
		t.Fatalf("unchanged address keeps its registry parts: %+v", renamed)
	}
}

// เดือนที่กลับรายการในหมายเหตุ: ภาษาไทยใช้ปี พ.ศ. แบบเอกสารภาษีไทย, อังกฤษใช้ ค.ศ.; รูปแบบผิดคืนค่าเดิม
func TestTaxMonthLabel(t *testing.T) {
	for _, tc := range []struct{ in, lang, want string }{
		{"2026-10", "th", language.Text("month_october", "th") + " 2569"},
		{"2026-01", "en", language.Text("month_january", "en") + " 2026"},
		{"2026-13", "th", "2026-13"},
		{"", "th", ""},
	} {
		if got := taxMonthLabel(tc.in, tc.lang); got != tc.want {
			t.Errorf("taxMonthLabel(%q,%q) = %q, want %q", tc.in, tc.lang, got, tc.want)
		}
	}
}

// หมายเหตุรายงาน: ยอดไม่รู้แบบ + กลับรายการเดือนหลัง นับจำนวนแทน {count}; ไม่มีเรื่องให้เตือน = ไม่มีหมายเหตุ
func TestWithholdingReportNotes(t *testing.T) {
	if notes := withholdingReportNotes(taxWithholdingReport{Rows: []TaxWithholdingRow{{}}}, "th"); len(notes) != 0 {
		t.Fatalf("no notes expected: %v", notes)
	}
	report := taxWithholdingReport{UnknownForm: 2, Rows: []TaxWithholdingRow{{ReversedMonth: "2026-11"}, {}, {ReversedMonth: "2026-12"}}}
	notes := withholdingReportNotes(report, "th")
	if len(notes) != 2 || strings.Contains(strings.Join(notes, ""), "{count}") {
		t.Fatalf("notes = %v", notes)
	}
	unknown := strings.ReplaceAll(language.Text("tax_wht_note_form_unknown", "th"), "{count}", "2")
	reversed := strings.ReplaceAll(language.Text("tax_wht_note_reversed_later", "th"), "{count}", "2")
	if notes[0] != unknown || notes[1] != reversed {
		t.Fatalf("notes = %v, want [%s %s]", notes, unknown, reversed)
	}
	// แบบ ภ.ง.ด.: หมายเหตุกลับรายการเดือนหลังเป็น key + จำนวน (frontend แปลตามภาษา)
	got := withholdingNotes(report.Rows)
	if len(got) != 2 || got[1].Key != "tax_form_note_wht_reversed_later" || got[1].Count != 2 {
		t.Fatalf("form notes = %+v", got)
	}
}

// UAT 2026-09-24: snapshot ใหม่เก็บชื่อ/ที่อยู่เต็ม — ใบแนบ ภ.ง.ด.3/ไฟล์ยื่นต้องยังได้คำนำหน้าและอำเภอ/จังหวัดแยกช่อง (ไม่ซ้อนในชื่อ)
// และรายงานต้องมีเลขสาขาของคู่ค้า (เดิมว่างเสมอ)
func TestApplyWithholdingPartySnapshotSplitsFullSnapshot(t *testing.T) {
	registry := TaxWithholdingRow{PartnerName: "สมชาย รับเหมาดี", Title: "นาย", Address: "เลขที่ 12 หมู่ 3 ตำบลคูบางหลวง",
		District: "อำเภอลาดหลุมแก้ว", Province: "ปทุมธานี", Postcode: "12140", BranchNo: "00000"}
	row := registry
	applyWithholdingPartySnapshot(&row, generalledger.SubledgerWithholding{Direction: 1, PayeeName: "นาย สมชาย รับเหมาดี",
		PayeeAddress: "เลขที่ 12 หมู่ 3 ตำบลคูบางหลวง อำเภอลาดหลุมแก้ว ปทุมธานี 12140", PayeeBranchNo: "00000"})
	if row != registry {
		t.Fatalf("full snapshot equal to the registry must split back into registry fields, got %+v", row)
	}
	branch := TaxWithholdingRow{PartnerName: "บริษัท ปูนไทย จำกัด"}
	applyWithholdingPartySnapshot(&branch, generalledger.SubledgerWithholding{Direction: 1, PayeeBranchNo: "00002"})
	if branch.BranchNo != "00002" {
		t.Fatalf("recorded branch must win: %+v", branch)
	}
}

// ทะเบียนภาษีหัก ณ ที่จ่าย: จอ/CSV/พิมพ์แสดงคำนำหน้า + ชื่อ (partnerfullname) แต่ partnername/title ยังแยกช่องสำหรับใบแนบ/ไฟล์ยื่น
func TestWithholdingRowPartnerFullName(t *testing.T) {
	cases := []struct{ title, name, want string }{
		{"นางสาว", "วิไลวรรณ ศรีสุข", "นางสาว วิไลวรรณ ศรีสุข"},
		{"บริษัท", "บริษัท สยามวัสดุ จำกัด", "บริษัท สยามวัสดุ จำกัด"}, // ชื่อบริษัทมีคำนำหน้าอยู่แล้ว ไม่เติมซ้ำ
		{"-", "ร้านทองดีการช่าง", "ร้านทองดีการช่าง"},
		{"", "สมชาย ใจดี", "สมชาย ใจดี"},
		{"นาย", "", ""},
	}
	for _, c := range cases {
		row := finishWithholdingRow(TaxWithholdingRow{Title: c.title, PartnerName: c.name}, "")
		if row.PartnerFullName != c.want || row.PartnerName != c.name || row.Title != c.title {
			t.Fatalf("title=%q name=%q: got full=%q name=%q title=%q, want full=%q", c.title, c.name, row.PartnerFullName, row.PartnerName, row.Title, c.want)
		}
	}
	// snapshot เต็มที่ตรงทะเบียน → แยกกลับ แต่ชื่อแสดงผลยังมีคำนำหน้า; snapshot ของชื่อเดิม (ทะเบียนถูกแก้ภายหลัง) → แสดงตาม snapshot
	split := TaxWithholdingRow{PartnerName: "วิไลวรรณ ศรีสุข", Title: "นางสาว"}
	applyWithholdingPartySnapshot(&split, generalledger.SubledgerWithholding{Direction: 1, PayeeName: "นางสาว วิไลวรรณ ศรีสุข"})
	split = finishWithholdingRow(split, "")
	if split.PartnerName != "วิไลวรรณ ศรีสุข" || split.Title != "นางสาว" || split.PartnerFullName != "นางสาว วิไลวรรณ ศรีสุข" {
		t.Fatalf("split snapshot = %+v", split)
	}
	renamed := TaxWithholdingRow{PartnerName: "วิไลวรรณ ศรีสุข", Title: "นาง"}
	applyWithholdingPartySnapshot(&renamed, generalledger.SubledgerWithholding{Direction: 1, PayeeName: "นางสาว วิไลวรรณ ทองคำ"})
	renamed = finishWithholdingRow(renamed, "")
	if renamed.PartnerFullName != "นางสาว วิไลวรรณ ทองคำ" || renamed.Title != "" {
		t.Fatalf("renamed snapshot = %+v", renamed)
	}
}
