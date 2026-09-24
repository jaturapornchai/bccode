package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"

	"smlcloudplatform/internal/rdfile"
	"smlcloudplatform/internal/rdform"
)

func TestRdFileFirstNameStripsTitleOnce(t *testing.T) {
	cases := []struct{ name, title, want string }{
		{"บริษัท สยาม จำกัด", "บริษัท", "สยาม จำกัด"},
		{"หจก.ก", "บริษัท", "หจก.ก"},
		{"นายสมชาย", "นาย", "สมชาย"},
		{"นางสาวิตรี", "นาง", "สาวิตรี"}, // ชื่อ "สาวิตรี" ขึ้นต้นด้วย "สาว" ตัดคำนำหน้าออกครั้งเดียวเท่านั้น
		{"สมชาย", "-", "สมชาย"},
		{" บริษัท  บริษัทไทย จำกัด ", "บริษัท", "บริษัทไทย จำกัด"},
		{"บริษัท", "บริษัท", ""},    // เหลือแต่คำนำหน้า → ชื่อว่าง ให้ Check แจ้ง
		{"นายิกา", "นาย", "นายิกา"}, // ตัดแล้วขึ้นต้นด้วยสระที่ขึ้นต้นคำไม่ได้ = ไม่ใช่คำนำหน้า (review 2026-09-24)
		{"นางสาวิตรี", "นางสาว", "นางสาวิตรี"},
	}
	for _, tc := range cases {
		if got := rdFileFirstName(tc.name, tc.title); got != tc.want {
			t.Errorf("rdFileFirstName(%q, %q) = %q, want %q", tc.name, tc.title, got, tc.want)
		}
	}
}

func TestRdFileValueConversions(t *testing.T) {
	for raw, want := range map[string]string{"1": "000001", "00000": "000000", "12345": "012345", "": "", "1234567": "1234567"} {
		if got := rdFileBranch(raw); got != want {
			t.Errorf("rdFileBranch(%q) = %q, want %q", raw, got, want)
		}
	}
	for raw, want := range map[string]string{"15/09/2569": "15092569", "5/9/2569": "05092569", "15092569": "15092569", "2026-09-15": "2026-09-15", "": ""} {
		if got := rdFileDate(raw); got != want {
			t.Errorf("rdFileDate(%q) = %q, want %q", raw, got, want)
		}
	}
	// ทศนิยมเกิน 2 ตำแหน่งที่มีค่าจริงไม่ปัด (ค่าเดิมไปให้ Check บล็อก); ติดลบแปลงรูปแบบแล้ว Check บล็อก
	for raw, want := range map[string]string{"3": "3.00", "1.5": "1.50", "1,234.5": "1234.50", "1.500": "1.50", "1.005": "1.005", "-5": "-5.00", "abc": "abc", "": ""} {
		if got := rdFileDecimal(raw); got != want {
			t.Errorf("rdFileDecimal(%q) = %q, want %q", raw, got, want)
		}
	}
}

func rdFileTestDoc() rdform.Document {
	return rdform.Document{
		Values: map[string]string{"tax_id": "0105558012349", "branch_no": "00000", "tax_section": "69tawi", "media_ref_no": "REF2569001", "surcharge": "10"},
		Rows: []map[string]string{
			{"seq": "", "tax_id": "0105562045671", "title": "บริษัท", "name": "บริษัท สยามขนส่งด่วน จำกัด", "surname": "ไม่ใช้",
				"l1_date": "15/09/2569", "l1_income_type": "ค่าขนส่ง", "l1_rate": "3", "l1_amount": "0.10", "l1_tax": "0.00", "l1_condition": "1",
				"l2_date": "16/09/2569", "l2_income_type": "ค่าขนส่ง", "l2_rate": "1", "l2_amount": "0.20", "l2_tax": "0.00", "l2_condition": "1"},
		},
	}
}

func TestRdFileContentPND53(t *testing.T) {
	h, ds := rdFileContent(rdfile.PND53, TaxFiling{Code: "pnd53", Year: 2026, Month: 9}, rdFileTestDoc(), rdFileOptions{}.normalized())
	if issues := rdfile.Check(rdfile.PND53, h, ds); len(issues) != 0 {
		t.Fatalf("issues: %+v", issues)
	}
	fields := rdfile.HeaderFields(rdfile.PND53, h)
	want := "H|0000|0105558012349|000000|1|PND53|0105558012349|000000|สำนักงานใหญ่|0|0|1||09|2569||00|1|0.30|0.00|10.00|10.00|0.00|REF2569001|1"
	if got := strings.Join(fields, "|"); got != want {
		t.Fatalf("header\n got %s\nwant %s", got, want)
	}
	d := strings.Join(rdfile.DetailFields(rdfile.PND53, ds[0]), "|")
	wantD := "D|1|000000|0105562045671|0000000000|บริษัท|สยามขนส่งด่วน จำกัด||15092569|3.00|0.10|0.00|ค่าขนส่ง|1|16092569|1.00|0.20|0.00|ค่าขนส่ง|1|00000000|0.00|0.00|0.00||||||||||||||"
	if d != wantD {
		t.Fatalf("detail\n got %s\nwant %s", d, wantD)
	}
	if name := rdfile.FileName(rdfile.PND53, h); name != "PND53_0105558012349_000000_2569_09_00_00.txt" {
		t.Fatalf("file name %s", name)
	}
}

func TestRdFileContentHeaderMapping(t *testing.T) {
	doc := rdFileTestDoc()
	// ภ.ง.ด.3: มาตรา 48 ทวิ = ธงที่ 2; ยื่นเพิ่มเติมครั้งที่ 3 → FORM_TYPE 03; สาขาอื่นต้องกรอกชื่อแผนกเอง
	doc.Values["tax_section"], doc.Values["branch_no"] = "48tawi", "00001"
	h, _ := rdFileContent(rdfile.PND3, TaxFiling{Code: "pnd3", Year: 2026, Month: 1, FilingSeq: 3}, doc, rdFileOptions{LTO: "1", BranchType: "V", SubmissionNo: "07"}.normalized())
	if h.Sections != [3]string{"0", "1", "0"} || h.FormType != "03" || h.TaxMonth != "01" || h.BranchNo != "000001" || h.DeptName != "" ||
		h.LTO != "1" || h.BranchType != "V" || h.SubmissionNo != "07" {
		t.Fatalf("header = %+v", h)
	}
	doc.Values["tax_section"] = ""
	h, ds := rdFileContent(rdfile.PND3, TaxFiling{Code: "pnd3", Year: 2026, Month: 9}, doc, rdFileOptions{DeptName: "ฝ่ายบัญชี"}.normalized())
	if h.Sections != [3]string{"0", "0", "0"} || h.DeptName != "ฝ่ายบัญชี" {
		t.Fatalf("header = %+v", h)
	}
	issues := rdfile.Check(rdfile.PND3, h, ds)
	for _, want := range []string{"tax_rdfile_section_required@tax_section", "tax_rdfile_address_required@addr_district"} {
		found := false
		for _, i := range issues {
			found = found || i.Key+"@"+i.Field == want
		}
		if !found {
			t.Fatalf("missing %s in %+v", want, issues)
		}
	}
	// ภ.ง.ด.53 ไม่มีนามสกุล แม้แถวจะมีค่า
	_, ds = rdFileContent(rdfile.PND53, TaxFiling{Code: "pnd53", Year: 2026, Month: 9}, rdFileTestDoc(), rdFileOptions{})
	if ds[0].LastName != "" {
		t.Fatalf("PND53 SNAME = %q", ds[0].LastName)
	}
}

func TestRdFileContentPND2(t *testing.T) {
	row := func(kind, taxID, date string) map[string]string {
		return map[string]string{"income_type": kind, "tax_id": taxID, "bank_account_no": "000-2025280", "title": "นาย", "name": "นายศิริวัฒน์", "surname": "ราชสีห์",
			"l1_date": date, "l1_rate": "15", "l1_amount": "1000", "l1_tax": "150", "l1_condition": "1"}
	}
	doc := rdform.Document{Values: map[string]string{"tax_id": "0105558012349", "branch_no": "00000", "media_ref_no": "REF1"},
		Rows: []map[string]string{row("interest", "3101701291901", "02/06/2569"), row("royalty", "3850100327556", "03/06/2569"),
			row("interest", "0000000000000", "04/06/2569"), row("other_404", "3101701291901", "05/06/2569")}}
	h, ds := rdFileContent(rdfile.PND2, TaxFiling{Code: "pnd2", Year: 2026, Month: 6}, doc, rdFileOptions{}.normalized())
	if issues := rdfile.Check(rdfile.PND2, h, ds); len(issues) != 0 {
		t.Fatalf("issues: %+v", issues)
	}
	// ลำดับเหมือนใบแนบที่พิมพ์: แยกตามประเภทเงินได้ที่พบก่อน (ดอกเบี้ย 2 แถว → ค่าสิทธิ → อื่นๆ) และ Row ชี้แถวเดิมในเอกสาร
	var got []string
	for _, d := range ds {
		got = append(got, d.Seq+":"+d.Items[0].IncType+":"+string(rune('0'+d.Row)))
	}
	if strings.Join(got, ",") != "1:2:1,2:2:3,3:1:2,4:5:4" {
		t.Fatalf("order = %v", got)
	}
	if h.TotNum != "4" || h.TotAmt != "4000.00" || h.TotTax != "600.00" || len(rdfile.HeaderFields(rdfile.PND2, h)) != 22 {
		t.Fatalf("header = %+v", h)
	}
	f := rdfile.DetailFields(rdfile.PND2, ds[0])
	if len(f) != 27 || f[5] != "0002025280" || f[6] != "นาย" || f[7] != "ศิริวัฒน์" || f[9] != "02062569" || f[13] != "2" {
		t.Fatalf("detail = %q", f)
	}
	// ประเภทเงินได้ทุกตัวเลือกของใบแนบ ภ.ง.ด.2 (ลำดับ 1-5 บนแบบ) → รหัส INC_TYPE_PND 1-5 ตาม [F2]
	for i, kind := range pnd2Kinds {
		if code := rdFilePND2IncomeTypes[kind]; code != string(rune('1'+i)) {
			t.Fatalf("%s → %q, want %d", kind, code, i+1)
		}
	}
	if len(rdFilePND2IncomeTypes) != len(pnd2Kinds) {
		t.Fatalf("map %v vs kinds %v", rdFilePND2IncomeTypes, pnd2Kinds)
	}
	// review 2026-09-24: ภาษีหัก ภ.ง.ด.2 ที่ประเภทเงินได้ไม่ใช่ 40(3)/40(4) (เช่น 40(2) ค่านายหน้า) ต้องไม่ถูกเขียนเป็น INC_TYPE 5 เงียบ ๆ
	commission := pnd2Rows([]TaxWithholdingRow{{TaxID: "3101701291901", PartnerName: "นายสมชาย ใจดี", Title: "นาย", IncomeType: "40_2", RatePercent: "15", BaseAmount: "1000.00", WhtAmount: "150.00", Condition: 1, PaidDate: "2026-09-16"}}, false)
	ch, cds := rdFileContent(rdfile.PND2, TaxFiling{Code: "pnd2", Year: 2026, Month: 9}, rdform.Document{Values: doc.Values, Rows: commission}, rdFileOptions{}.normalized())
	issues := rdfile.Check(rdfile.PND2, ch, cds)
	if len(issues) != 1 || issues[0].Key != "tax_rdfile_income_type_required" || issues[0].Field != "income_type" {
		t.Fatalf("40(2) in ภ.ง.ด.2 issues = %+v (INC_TYPE %q)", issues, cds[0].Items[0].IncType)
	}
	// ชื่อกลางอยู่ใน SNAME ([F2] ช่อง 9)
	foreign := pnd2Rows([]TaxWithholdingRow{{TaxID: "3101701291901", PartnerName: "Mr. John Michael Smith", Title: "Mr.", IncomeType: "40_4a", RatePercent: "15", BaseAmount: "1000.00", WhtAmount: "150.00", Condition: 1, PaidDate: "2026-09-16"}}, false)
	_, fds := rdFileContent(rdfile.PND2, TaxFiling{Code: "pnd2", Year: 2026, Month: 9}, rdform.Document{Values: doc.Values, Rows: foreign}, rdFileOptions{}.normalized())
	if fds[0].FirstName != "John" || fds[0].LastName != "Michael Smith" {
		t.Fatalf("FNAME/SNAME = %q / %q", fds[0].FirstName, fds[0].LastName)
	}
}

func TestRdFileIssuesResponse(t *testing.T) {
	e := echo.New()
	rec := httptest.NewRecorder()
	c := e.NewContext(httptest.NewRequest(http.MethodPost, "/", nil), rec)
	issues := make([]rdfile.Issue, 150)
	for i := range issues {
		issues[i] = rdfile.Issue{Key: "tax_rdfile_forbidden_char", Field: "name", Row: i + 1, Args: map[string]string{"char": "/"}}
	}
	issues[0] = rdfile.Issue{Key: "tax_form_row_label", Field: "name", Row: 1, Args: map[string]string{"n": "7"}} // key ที่มีอยู่แล้ว: "แถวที่ {n}"
	if err := (&taxFormCall{c: c, lang: "th"}).rdFileIssues(issues); err != nil {
		t.Fatal(err)
	}
	var body struct {
		Code   string `json:"code"`
		Total  int    `json:"total"`
		Issues []struct {
			Key, Field, Message string
			Row                 int
			Args                map[string]string
		} `json:"issues"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusBadRequest || body.Code != "tax_rdfile_invalid" || body.Total != 150 || len(body.Issues) != maxRdFileIssues {
		t.Fatalf("status=%d body=%+v", rec.Code, body)
	}
	if body.Issues[0].Message != "แถวที่ 7" || body.Issues[1].Args["char"] != "/" || body.Issues[1].Row != 2 {
		t.Fatalf("issues[0..1] = %+v", body.Issues[:2])
	}
}

func TestTaxFormRdFileHandlerGuards(t *testing.T) {
	if rec := callTaxReportHandler(t, TaxFormRdFileHandler, `{"id":1,"version":1}`, nil); rec.Code != http.StatusUnauthorized {
		t.Fatalf("no user: %d", rec.Code)
	}
	// ยังไม่บันทึก (ไม่มี id) → ไม่พบฉบับ ก่อนแตะฐานข้อมูล
	rec := callTaxReportHandler(t, TaxFormRdFileHandler, `{"version":1,"rdfile":{"submission_no":"00"}}`, &taxReportTestUser)
	if rec.Code != http.StatusNotFound || !strings.Contains(rec.Body.String(), "tax_form_not_found") {
		t.Fatalf("no id: %d %s", rec.Code, rec.Body.String())
	}
	rec = callTaxReportHandler(t, TaxFormRdFileHandler, `{"id":1,"rdfile":"x"}`, &taxReportTestUser)
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "tax_form_payload_invalid") {
		t.Fatalf("bad rdfile payload: %d %s", rec.Code, rec.Body.String())
	}
}

// ฉบับไม่มีใบแนบ: ยอดทุกช่องเป็นศูนย์ = ยื่นหัวแบบอย่างเดียวได้ (TOT_NUM 0); มียอดในหน้าแบบแต่ไม่มีใบแนบ = tax_rdfile_no_rows
// และช่องว่างหัว/ท้ายของค่าจากหัวแบบ/แผงสร้างไฟล์ถูกตัดก่อนตรวจ
func TestRdFileContentNilReturnAndTrim(t *testing.T) {
	doc := rdform.Document{Values: map[string]string{"tax_id": "0105558012349", "branch_no": "00000", "media_ref_no": " REF2569001 ",
		"tax_section": "3tres", "total_income": "0.00", "total_tax": "", "surcharge": "0"}}
	h, ds := rdFileContent(rdfile.PND53, TaxFiling{Code: "pnd53", Year: 2026, Month: 9}, doc, rdFileOptions{}.normalized())
	if issues := rdfile.Check(rdfile.PND53, h, ds); len(issues) != 0 || !h.NilReturn || h.TotNum != "0" || h.UserID != "REF2569001" {
		t.Fatalf("nil return: header=%+v issues=%+v", h, issues)
	}
	doc.Values["total_tax"] = "300.00"
	h, ds = rdFileContent(rdfile.PND53, TaxFiling{Code: "pnd53", Year: 2026, Month: 9}, doc, rdFileOptions{}.normalized())
	issues := rdfile.Check(rdfile.PND53, h, ds)
	if h.NilReturn || len(issues) != 1 || issues[0].Key != "tax_rdfile_no_rows" {
		t.Fatalf("totals without rows: header=%+v issues=%+v", h, issues)
	}
}

// แคตตาล็อกบอก frontend ว่าแบบไหนสร้างไฟล์ยื่นด้วยสื่อได้ (ภ.ง.ด.53/3/2) — ภ.ง.ด.2ก และแบบอื่นไม่มีธงนี้
func TestTaxFormCatalogRdFileFlag(t *testing.T) {
	rec := callTaxReportHandler(t, TaxFormCatalogHandler, `{}`, &taxReportTestUser)
	if rec.Code != http.StatusOK {
		t.Fatalf("catalog: %d %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Data []struct {
			Code   string `json:"code"`
			RdFile bool   `json:"rdfile"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	got := map[string]bool{}
	for _, e := range body.Data {
		got[e.Code] = e.RdFile
	}
	for code, want := range map[string]bool{"pnd53": true, "pnd3": true, "pnd2": true, "pnd2a": false, "pp30": false, "pnd51": false} {
		if got[code] != want {
			t.Fatalf("%s rdfile = %v, want %v (catalog %+v)", code, got[code], want, body.Data)
		}
	}
}
