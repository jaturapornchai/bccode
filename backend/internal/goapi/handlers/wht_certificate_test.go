package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	authmodels "smlcloudplatform/internal/authentication/models"
	"smlcloudplatform/internal/generalledger"
	orgpolicy "smlcloudplatform/internal/organization/access"
	"smlcloudplatform/internal/whtcert"
	msmodels "smlcloudplatform/pkg/microservice/models"
)

const whtCertBody = `{"holdingcode":"h1","businesscode":"b1","certificate":{
  "runno":"2569/0001","sequenceno":"1","form":"53","condition":"withhold","issuedate":"2026-01-15",
  "payer":{"name":"พิมพ์แทนชื่อบริษัท","address":"เลขที่ 88/12 ถนนลาดหลุมแก้ว จังหวัดปทุมธานี 12140","taxid":"0000000000000"},
  "payee":{"name":"บริษัท สยามขนส่งด่วน จำกัด","address":"เลขที่ 159 แขวงบางนาใต้ เขตบางนา กรุงเทพมหานคร 10260","taxid":"0105562045671"},
  "incomes":[{"type":"3_tres","paiddate":"2026-01-15","amount":"25000.00","tax":"750.00"}]}}`

// stubWhtAccess - แทนการตรวจสมาชิกภาพในฐานควบคุมกลาง: scopes = สิทธิ์ของผู้เรียก, err = ผลตรวจ
func stubWhtAccess(t *testing.T, scopes []authmodels.AccessScope, err error) {
	t.Helper()
	original := whtCallerAccess
	whtCallerAccess = func(_ context.Context, _ msmodels.UserInfo, company string) (whtAccess, error) {
		if err != nil {
			return whtAccess{}, err
		}
		return whtAccess{company: company, scopes: scopes}, nil
	}
	t.Cleanup(func() { whtCallerAccess = original })
}

var whtCompanyWide = []authmodels.AccessScope{{ScopeType: "company", CompanyUID: "B1", AllBranches: true}}

// stubWhtCompany - ทะเบียนบริษัท + ผู้เรียกมีสิทธิ์ทั้งบริษัท B1 (เทสสิทธิ์ใช้ stubWhtAccess ทับ)
func stubWhtCompany(t *testing.T, header CompanyHeader) *CompanyHeader {
	t.Helper()
	stubWhtAccess(t, whtCompanyWide, nil)
	seen := &CompanyHeader{}
	original := whtCompanyHeader
	whtCompanyHeader = func(_ context.Context, holding, business string) (CompanyHeader, error) {
		*seen = CompanyHeader{Code: holding + "/" + business}
		return header, nil
	}
	t.Cleanup(func() { whtCompanyHeader = original })
	return seen
}

func TestWhtCertificateHandler_RendersPDFWithRegisteredPayer(t *testing.T) {
	// ทะเบียนบริษัทมีชื่อ+เลขผู้เสียภาษี → ต้องทับค่าที่หน้าจอส่งมา (เลข 0000000000000 ผิด checksum ถ้าไม่ถูกทับจะ 400)
	seen := stubWhtCompany(t, CompanyHeader{Name: "บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด", TaxID: "0105558012349"})
	rec := callTaxReportHandler(t, WhtCertificateHandler, whtCertBody, &taxReportTestUser)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body=%s", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/pdf" {
		t.Fatalf("content-type %q", ct)
	}
	if !bytes.HasPrefix(rec.Body.Bytes(), []byte("%PDF-")) {
		t.Fatal("body is not a PDF")
	}
	if got := rec.Header().Get("Content-Disposition"); got != `inline; filename="50tawi-2569-0001.pdf"` {
		t.Fatalf("content-disposition %q", got)
	}
	if seen.Code != "h1/B1" {
		t.Fatalf("company lookup scope = %q", seen.Code)
	}
}

func TestWhtCertificateHandler_ValidationErrorIsLanguageKey(t *testing.T) {
	stubWhtCompany(t, CompanyHeader{}) // ทะเบียนยังไม่มีเลข → ใช้ค่าที่ส่งมา ซึ่ง checksum ผิด
	rec := callTaxReportHandler(t, WhtCertificateHandler, whtCertBody, &taxReportTestUser)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d body=%s", rec.Code, rec.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["code"] != "wht_cert_taxid_checksum" || body["field"] != "payer.taxid" || body["message"] == "" {
		t.Fatalf("body %v", body)
	}
}

func TestWhtCertificateHandler_Unauthorized(t *testing.T) {
	stubWhtCompany(t, CompanyHeader{})
	if rec := callTaxReportHandler(t, WhtCertificateHandler, whtCertBody, nil); rec.Code != http.StatusUnauthorized {
		t.Fatalf("status %d", rec.Code)
	}
}

// stubWhtRecorded - แทนการอ่านรายการภาษีหักจากฐานข้อมูลกลุ่มกิจการ (ใบสำคัญสาขา 00001); เก็บ argument ที่ handler ส่งมา
func stubWhtRecorded(t *testing.T, item generalledger.SubledgerWithholding, partner generalledger.SubledgerPartner, err error) *[]string {
	t.Helper()
	seen := &[]string{}
	original := whtRecordedItem
	whtRecordedItem = func(_ context.Context, holding, company, journalID, itemID string) (whtRecord, error) {
		*seen = []string{holding, company, journalID, itemID}
		return whtRecord{item: item, partner: partner, branch: "00001"}, err
	}
	t.Cleanup(func() { whtRecordedItem = original })
	return seen
}

func whtCertBodyFor(journalID, itemID string) string {
	return strings.Replace(whtCertBody, `"certificate":{`, `"journalid":"`+journalID+`","withholdingid":"`+itemID+`","certificate":{`, 1)
}

// save-audit 2026-09-24: 50 ทวิ ใช้ผู้จ่าย/ผู้รับเงินที่บันทึกในรายการก่อน ช่องที่บันทึกว่างใช้ทะเบียนปัจจุบัน แล้วจึงค่าที่หน้าจอส่งมา
func TestApplyRecordedWithholdingPrefersSnapshotThenRegisters(t *testing.T) {
	company := CompanyHeader{Name: "บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด", TaxID: "0105558012349"}
	partner := generalledger.SubledgerPartner{Code: "TRANS", Name: "บริษัท ขนส่งไทยเร็ว จำกัด (ชื่อใหม่)", TaxID: "0105562045671", Address: "159 บางนา กรุงเทพมหานคร 10260"}
	screen := whtcert.Certificate{BookNo: "7", RunNo: "จอ-001",
		Payer: whtcert.Party{Name: "ชื่อจากจอ", Address: "ที่อยู่ผู้หักจากจอ", TaxID: "1111111111111"},
		Payee: whtcert.Party{Name: "ผู้รับจากจอ", Address: "ที่อยู่ผู้รับจากจอ", TaxID: "2222222222222"}}

	// ทิศทาง 1: ผู้รับเงิน (คู่ค้า) มี snapshot ณ วันบันทึก → ใช้ snapshot แม้ทะเบียนคู่ค้าเปลี่ยนชื่อแล้ว; ผู้จ่าย (บริษัท) ว่าง → ทะเบียนบริษัท
	item := generalledger.SubledgerWithholding{Direction: 1, BookNo: "001", CertificateNo: "0012/2569",
		PayeeName: "บริษัท ขนส่งไทยเร็ว จำกัด", PayeeTaxID: "0105562045671", PayeeAddress: "88/12 ถนนลาดหลุมแก้ว ปทุมธานี 12140"}
	got := applyRecordedWithholding(screen, item, partner, company)
	if got.Payee.Name != "บริษัท ขนส่งไทยเร็ว จำกัด" || got.Payee.Address != "88/12 ถนนลาดหลุมแก้ว ปทุมธานี 12140" || got.Payee.TaxID != "0105562045671" {
		t.Fatalf("payee must come from the stored snapshot: %+v", got.Payee)
	}
	if got.Payer.Name != company.Name || got.Payer.TaxID != company.TaxID || got.Payer.Address != "ที่อยู่ผู้หักจากจอ" {
		t.Fatalf("blank payer snapshot must fall back to company register then screen: %+v", got.Payer)
	}
	if got.BookNo != "001" || got.RunNo != "0012/2569" {
		t.Fatalf("book/run no = %q/%q", got.BookNo, got.RunNo)
	}

	// snapshot ผู้รับว่างทั้งชุด → ทะเบียนคู่ค้าปัจจุบัน; เล่มที่/เลขที่ว่าง → ค่าจากจอ
	got = applyRecordedWithholding(screen, generalledger.SubledgerWithholding{Direction: 1}, partner, company)
	if got.Payee.Name != partner.Name || got.Payee.TaxID != partner.TaxID || got.Payee.Address != partner.Address || got.BookNo != "7" || got.RunNo != "จอ-001" {
		t.Fatalf("blank snapshot fallback: %+v", got)
	}

	// ทิศทาง 2 ผู้จ่ายหักภาษีเรา: ผู้จ่าย = คู่ค้า, ผู้รับ = บริษัท; snapshot ผู้จ่ายชนะทะเบียน
	item = generalledger.SubledgerWithholding{Direction: 2, PayerName: "ลูกค้าตามหนังสือรับรอง", PayerTaxID: "0105562045671"}
	got = applyRecordedWithholding(screen, item, partner, company)
	if got.Payer.Name != "ลูกค้าตามหนังสือรับรอง" || got.Payer.Address != partner.Address || got.Payee.Name != company.Name || got.Payee.TaxID != company.TaxID || got.Payee.Address != "ที่อยู่ผู้รับจากจอ" {
		t.Fatalf("direction 2 merge: payer %+v payee %+v", got.Payer, got.Payee)
	}

	// adversarial review 2026-09-24: snapshot ฝั่งบริษัท (แก้ได้ทาง API/reconcile หรือค้างจากตอนทะเบียนว่าง) ต้องไม่ชนะทะเบียนบริษัท
	forged := generalledger.SubledgerWithholding{Direction: 1, PayerName: "บริษัท อื่น จำกัด", PayerTaxID: "0105599999999", PayerAddress: "ที่อยู่ตาม snapshot"}
	got = applyRecordedWithholding(screen, forged, partner, company)
	if got.Payer.Name != company.Name || got.Payer.TaxID != company.TaxID || got.Payer.Address != "ที่อยู่ตาม snapshot" {
		t.Fatalf("direction 1 company side must follow the register: %+v", got.Payer)
	}
	forged = generalledger.SubledgerWithholding{Direction: 2, PayeeName: "บริษัท อื่น จำกัด", PayeeTaxID: "0105599999999"}
	if got = applyRecordedWithholding(screen, forged, partner, company); got.Payee.Name != company.Name || got.Payee.TaxID != company.TaxID {
		t.Fatalf("direction 2 company side must follow the register: %+v", got.Payee)
	}
	// ทะเบียนบริษัทยังว่าง → snapshot ฝั่งบริษัทใช้ได้ (ก่อนค่าจากจอ)
	if got = applyRecordedWithholding(screen, generalledger.SubledgerWithholding{Direction: 1, PayerName: "บริษัท รุ่งเรือง (snapshot)"}, partner, CompanyHeader{}); got.Payer.Name != "บริษัท รุ่งเรือง (snapshot)" || got.Payer.TaxID != "1111111111111" {
		t.Fatalf("empty register: %+v", got.Payer)
	}
}

// review 2026-09-24: 50 ทวิ ที่อ้างรายการที่บันทึก พิมพ์ยอด/ภาษี/วันที่/ประเภท/แบบ/เงื่อนไขตามรายการเสมอ — ค่าที่คำขอส่งมา
// (ว่าง ต่าง หรือหลายบรรทัด) ใช้ไม่ได้; ตัวเลขในคำขอใช้เฉพาะหนังสือรับรองที่ไม่อ้างรายการ
func TestApplyRecordedFiguresRendersFromRecord(t *testing.T) {
	tax := generalledger.Amount("300")
	item := generalledger.SubledgerWithholding{FormType: "PND53", Condition: 1, IncomeType: "3_tres", PaymentDate: "2026-01-15",
		BaseAmount: "10000", TaxAmount: &tax, CertificateDate: "2026-01-16"}
	want := whtcert.Income{Type: "3_tres", PaidDate: "2026-01-15", Amount: "10000", Tax: "300"}
	for name, cert := range map[string]whtcert.Certificate{
		"blank request": {},
		"other figures": {Form: "3", Condition: "always", IssueDate: "2026-02-01",
			Incomes: []whtcert.Income{{Type: "40_2", PaidDate: "2026-02-01", Amount: "50000", Tax: "1500", Note: "ค่านายหน้า"}}},
		"extra lines":    {Incomes: []whtcert.Income{{Type: "3_tres", Amount: "1"}, {Type: "40_2", Amount: "2", Tax: "0"}}},
		"other + note":   {Condition: "other", ConditionNote: "ตามสัญญา", FundGPF: "100", FundSSF: "750", FundPVD: "300"},
		"equal decimals": {Form: "53", Condition: "withhold", Incomes: []whtcert.Income{{Type: "3_tres", PaidDate: "2026-01-15", Amount: "10000.00", Tax: "300.00"}}},
	} {
		got := applyRecordedFigures(cert, item)
		if len(got.Incomes) != 1 || got.Incomes[0] != want || got.Form != "53" || got.Condition != "withhold" || got.ConditionNote != "" ||
			got.FundGPF != "" || got.FundSSF != "" || got.FundPVD != "" || got.IssueDate != "2026-01-16" {
			t.Fatalf("%s: %+v", name, got)
		}
	}
	// ไม่ได้บันทึกวันที่ออกหนังสือรับรอง → ใช้วันที่จากหน้าจอ
	item.CertificateDate = ""
	if got := applyRecordedFigures(whtcert.Certificate{IssueDate: "2026-01-20"}, item); got.IssueDate != "2026-01-20" {
		t.Fatalf("issue date fallback = %q", got.IssueDate)
	}
	// บรรทัดที่มีช่อง "ระบุ": ข้อความจากหน้าจอของประเภทเดียวกันก่อน แล้วจึงคำอธิบายเงินได้ที่บันทึก
	other := generalledger.SubledgerWithholding{FormType: "PND3", Condition: 1, IncomeType: whtcert.IncomeOther, PaymentDate: "2026-01-15",
		BaseAmount: "5000", TaxAmount: &tax, Description: "ค่ารางวัลการแข่งขัน"}
	if got := applyRecordedFigures(whtcert.Certificate{}, other); got.Incomes[0].Note != "ค่ารางวัลการแข่งขัน" {
		t.Fatalf("note from record = %q", got.Incomes[0].Note)
	}
	screen := whtcert.Certificate{Incomes: []whtcert.Income{{Type: whtcert.IncomeOther, Note: "ค่ารางวัลประกวดออกแบบ"}}}
	if got := applyRecordedFigures(screen, other); got.Incomes[0].Note != "ค่ารางวัลประกวดออกแบบ" {
		t.Fatalf("note from screen = %q", got.Incomes[0].Note)
	}
	if got := applyRecordedFigures(whtcert.Certificate{Incomes: []whtcert.Income{{Type: "3_tres", Note: "ไม่มีช่อง"}}}, item); got.Incomes[0].Note != "" {
		t.Fatalf("a line without a note box must stay blank: %q", got.Incomes[0].Note)
	}
}

func TestWhtCertificateHandler_UsesRecordedWithholding(t *testing.T) {
	stubWhtCompany(t, CompanyHeader{Name: "บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด", TaxID: "0105558012349"})
	// snapshot ผู้รับเงินเลขถูกต้อง แทนเลขในคำขอ — ถ้า handler ไม่ใช้ snapshot เลขในคำขอ (0000000000000) จะ checksum ผิด 400
	recordedTax := generalledger.Amount("750")
	seen := stubWhtRecorded(t, generalledger.SubledgerWithholding{ID: "W1", Direction: 1, CertificateNo: "0012/2569",
		FormType: "PND53", Condition: 1, IncomeType: "3_tres", PaymentDate: "2026-01-15", BaseAmount: "25000", TaxAmount: &recordedTax,
		PayeeName: "บริษัท สยามขนส่งด่วน จำกัด", PayeeTaxID: "0105562045671", PayeeAddress: "เลขที่ 159 แขวงบางนาใต้ เขตบางนา กรุงเทพมหานคร 10260"},
		generalledger.SubledgerPartner{}, nil)
	body := strings.Replace(whtCertBodyFor("J-1", "W1"), `"taxid":"0105562045671"`, `"taxid":"0000000000000"`, 1)
	rec := callTaxReportHandler(t, WhtCertificateHandler, body, &taxReportTestUser)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body=%s", rec.Code, rec.Body.String())
	}
	if strings.Join(*seen, "|") != "h1|B1|J-1|W1" {
		t.Fatalf("recorded lookup args = %v", *seen)
	}
	if got := rec.Header().Get("Content-Disposition"); got != `inline; filename="50tawi-0012-2569.pdf"` {
		t.Fatalf("stored certificate no must name the file: %q", got)
	}
}

func TestWhtCertificateHandler_RecordedWithholdingNotFound(t *testing.T) {
	stubWhtCompany(t, CompanyHeader{})
	stubWhtRecorded(t, generalledger.SubledgerWithholding{}, generalledger.SubledgerPartner{}, generalledger.ErrNotFound)
	for name, tc := range map[string]struct {
		body   string
		status int
	}{
		"not posted or missing": {whtCertBodyFor("J-1", "W9"), http.StatusNotFound},
		"journal without item":  {whtCertBodyFor("J-1", ""), http.StatusBadRequest},
	} {
		rec := callTaxReportHandler(t, WhtCertificateHandler, tc.body, &taxReportTestUser)
		var got map[string]any
		_ = json.Unmarshal(rec.Body.Bytes(), &got)
		if rec.Code != tc.status || got["code"] != "wht_cert_record_not_found" || got["field"] != "withholdingid" {
			t.Fatalf("%s: status %d body %v", name, rec.Code, got)
		}
	}
}

// review 2026-09-24: คำขอส่งยอด 25,000/750 แต่รายการบันทึก 10,000/300 → PDF พิมพ์ตามรายการ (ไม่ปฏิเสธ ไม่ใช้ยอดในคำขอ)
func TestWhtCertificateHandler_RecordedFiguresOverrideRequest(t *testing.T) {
	stubWhtCompany(t, CompanyHeader{Name: "บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด", TaxID: "0105558012349"})
	recordedTax := generalledger.Amount("300")
	stubWhtRecorded(t, generalledger.SubledgerWithholding{ID: "W1", Direction: 1, FormType: "PND53", Condition: 1, IncomeType: "3_tres", PaymentDate: "2026-01-15", BaseAmount: "10000", TaxAmount: &recordedTax,
		PayeeName: "บริษัท สยามขนส่งด่วน จำกัด", PayeeTaxID: "0105562045671"}, generalledger.SubledgerPartner{}, nil)
	rec := callTaxReportHandler(t, WhtCertificateHandler, whtCertBodyFor("J-1", "W1"), &taxReportTestUser)
	if rec.Code != http.StatusOK || !bytes.HasPrefix(rec.Body.Bytes(), []byte("%PDF-")) {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
}

// review 2026-09-24: สิทธิ์ตรวจจากสมาชิกภาพทุกครั้ง — หมดอายุ/ไม่ใช่สมาชิก/บริษัทนอกสิทธิ์ = 403 และใบสำคัญต้องอยู่ในสาขาที่มีสิทธิ์
func TestWhtCertificateHandler_EnforcesCompanyAndBranchScope(t *testing.T) {
	recordedTax := generalledger.Amount("750")
	recorded := generalledger.SubledgerWithholding{ID: "W1", Direction: 1, FormType: "PND53", Condition: 1, IncomeType: "3_tres",
		PaymentDate: "2026-01-15", BaseAmount: "25000", TaxAmount: &recordedTax,
		PayeeName: "บริษัท สยามขนส่งด่วน จำกัด", PayeeTaxID: "0105562045671", PayeeAddress: "เลขที่ 159 แขวงบางนาใต้ เขตบางนา กรุงเทพมหานคร 10260"}
	branch := func(code string) []authmodels.AccessScope {
		return []authmodels.AccessScope{{ScopeType: "branch", CompanyUID: "B1", BranchUID: code}}
	}
	for name, tc := range map[string]struct {
		scopes []authmodels.AccessScope
		err    error
		body   string
		status int
		code   string
		field  string
	}{
		"expired membership":   {err: orgpolicy.ErrAccessExpired, body: whtCertBody, status: http.StatusForbidden, code: "user_access_expired"},
		"not a member":         {err: orgpolicy.ErrActiveMembershipRequired, body: whtCertBody, status: http.StatusForbidden, code: "wht_cert_scope_denied"},
		"company out of scope": {err: errWhtScopeDenied, body: whtCertBody, status: http.StatusForbidden, code: "wht_cert_scope_denied"},
		"voucher of another branch": {scopes: branch("00002"), body: whtCertBodyFor("J-1", "W1"),
			status: http.StatusForbidden, code: "wht_cert_scope_denied", field: "journalid"},
		"voucher of own branch": {scopes: branch("00001"), body: whtCertBodyFor("J-1", "W1"), status: http.StatusOK},
	} {
		t.Run(name, func(t *testing.T) {
			stubWhtCompany(t, CompanyHeader{Name: "บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด", TaxID: "0105558012349"})
			stubWhtAccess(t, tc.scopes, tc.err)
			stubWhtRecorded(t, recorded, generalledger.SubledgerPartner{}, nil)
			rec := callTaxReportHandler(t, WhtCertificateHandler, tc.body, &taxReportTestUser)
			if rec.Code != tc.status {
				t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
			}
			if tc.code == "" {
				return
			}
			var got map[string]any
			_ = json.Unmarshal(rec.Body.Bytes(), &got)
			if got["code"] != tc.code || got["field"] != tc.field || got["message"] == "" {
				t.Fatalf("body %v", got)
			}
		})
	}
}

func TestWhtAccessAllowsBranch(t *testing.T) {
	wide := whtAccess{company: "B1", scopes: whtCompanyWide}
	limited := whtAccess{company: "B1", scopes: []authmodels.AccessScope{{ScopeType: "branch", CompanyUID: "B1", BranchUID: "00001"}}}
	if !wide.allowsBranch("00002") || !wide.allowsBranch("") {
		t.Fatal("company-wide access must cover every branch and company-level vouchers")
	}
	if !limited.allowsBranch("00001") || limited.allowsBranch("00002") || limited.allowsBranch("") {
		t.Fatal("branch access must cover only its branch (a voucher without a branch needs company-wide access)")
	}
}
