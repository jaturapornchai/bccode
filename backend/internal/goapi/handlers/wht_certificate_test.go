package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

const whtCertBody = `{"holdingcode":"h1","businesscode":"b1","certificate":{
  "runno":"2569/0001","sequenceno":"1","form":"53","condition":"withhold","issuedate":"2026-01-15",
  "payer":{"name":"พิมพ์แทนชื่อบริษัท","address":"เลขที่ 88/12 ถนนลาดหลุมแก้ว จังหวัดปทุมธานี 12140","taxid":"0000000000000"},
  "payee":{"name":"บริษัท สยามขนส่งด่วน จำกัด","address":"เลขที่ 159 แขวงบางนาใต้ เขตบางนา กรุงเทพมหานคร 10260","taxid":"0105562045671"},
  "incomes":[{"type":"3_tres","paiddate":"2026-01-15","amount":"25000.00","tax":"750.00"}]}}`

func stubWhtCompany(t *testing.T, header CompanyHeader) *CompanyHeader {
	t.Helper()
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
