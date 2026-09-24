package rdfile

import (
	"archive/zip"
	"bytes"
	"flag"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "เขียน testdata/*.golden ใหม่จากผลลัพธ์ปัจจุบัน")

// เลขประจำตัวผู้เสียภาษีที่ผ่าน mod 11 (สมมุติ ไม่ใช่ของบุคคลจริง)
const (
	companyTaxID = "0105558012349"
	payee53A     = "0105562045671"
	payee53B     = "0105535141967"
	person3A     = "3101701291901"
	person3B     = "3850100327556"
)

func baseHeader(month, yearBE string) Header {
	return Header{SenderID: SenderIDMedia, SenderNID: companyTaxID, SenderBranch: "000000", SenderRole: SenderRoleSelf,
		NID: companyTaxID, BranchNo: "000000", DeptName: HeadOfficeDept, LTO: "0", TaxMonth: month, TaxYear: yearBE,
		BranchType: "", FormType: "00", SurAmt: "0.00", UserID: "ZXC1234567", FormFlag: FormFlagMedia, SubmissionNo: "00"}
}

func item(date, rate, amount, tax, incType, cond string) Item {
	return Item{PaidDate: date, TaxRate: rate, PaidAmt: amount, TaxAmt: tax, IncType: incType, PayCon: cond}
}

func pnd53Sample() (Header, []Detail) {
	h := baseHeader("09", "2569")
	h.Sections = [3]string{"1", "0", "0"}
	ds := []Detail{
		{Row: 1, Seq: "1", BranchNo: "000000", NID: payee53A, TIN: NoTIN, Title: "บริษัท", FirstName: "สยามขนส่งด่วน จำกัด",
			Items: [3]Item{item("15092569", "3.00", "25000.00", "750.00", "ค่าขนส่ง", "1"), item("20092569", "5.00", "10000.00", "500.00", "ค่าเช่าอาคาร", "1")}},
		{Row: 2, Seq: "2", BranchNo: "000000", NID: payee53B, TIN: NoTIN, Title: "ห้างหุ้นส่วนจำกัด", FirstName: "รุ่งเรืองการค้าไทย",
			Items: [3]Item{item("30092569", "1.00", "0.10", "0.00", "ค่าโฆษณา", "2")}},
	}
	return Summarize(PND53, h, ds), ds
}

func pnd3Sample() (Header, []Detail) {
	h := baseHeader("09", "2569")
	h.Sections = [3]string{"1", "0", "0"}
	h.LTO = ""
	addr := Address{Amphur: "เขตบางนา", Province: "กรุงเทพมหานคร", PostalCode: "10260"}
	ds := []Detail{
		{Row: 1, Seq: "1", BranchNo: "000000", NID: person3A, TIN: NoTIN, Title: "นาย", FirstName: "สมชาย", LastName: "ใจดี",
			Items: [3]Item{item("05092569", "3.00", "0.20", "0.00", "ค่าจ้างทำของ", "1")}, Address: addr},
		{Row: 2, Seq: "2", BranchNo: "000000", NID: person3B, TIN: NoTIN, Title: "นางสาว", FirstName: "สมหญิง", LastName: "ศรีสุข",
			Items: [3]Item{item("10092569", "5.00", "12000.00", "600.00", "ค่าเช่าอาคาร", "3")}, Address: Address{Amphur: "เมืองนนทบุรี", Province: "นนทบุรี", PostalCode: "11000"}},
	}
	return Summarize(PND3, h, ds), ds
}

func pnd2Sample() (Header, []Detail) {
	h := baseHeader("06", "2569")
	h.BranchType = "V"
	ds := []Detail{
		{Row: 1, Seq: "1", BranchNo: "000000", NID: person3A, TIN: NoTIN, AccNo: "0002025280", Title: "นาย", FirstName: "ศิริวัฒน์", LastName: "ราชสีห์",
			Items: [3]Item{item("02062569", "15.00", "1000.00", "150.00", "2", "1")}},
		{Row: 2, Seq: "2", BranchNo: "000000", NID: "0000000000000", TIN: NoTIN, AccNo: "0002031353", Title: "นาง", FirstName: "นัดดา", LastName: "จันทร์แสงศรี",
			Items: [3]Item{item("30062569", "10.00", "500.00", "50.00", "2", "1")}},
		{Row: 3, Seq: "3", BranchNo: "000000", NID: person3B, TIN: NoTIN, AccNo: "0002045817", Title: "นาย", FirstName: "วรพงศ์", LastName: "จักรเสน",
			Items: [3]Item{item("15062569", "10.00", "2500.00", "250.00", "3", "1")}},
	}
	return Summarize(PND2, h, ds), ds
}

// TestGolden - ไฟล์สังเคราะห์ของทั้ง 3 แบบต้องผ่าน Check และตรงไฟล์ golden ทีละ byte (BOM, CRLF, ไม่มี CRLF ท้าย, จำนวนช่อง)
func TestGolden(t *testing.T) {
	cases := []struct {
		kind           Kind
		build          func() (Header, []Detail)
		header, detail int
	}{
		{PND53, pnd53Sample, headerFieldsWHT, detailFieldsWHT},
		{PND3, pnd3Sample, headerFieldsWHT, detailFieldsWHT},
		{PND2, pnd2Sample, headerFieldsPND2, detailFieldsPND2},
	}
	for _, tc := range cases {
		t.Run(string(tc.kind), func(t *testing.T) {
			h, ds := tc.build()
			if issues := Check(tc.kind, h, ds); len(issues) != 0 {
				t.Fatalf("sample must pass Check: %+v", issues)
			}
			got := Encode(Lines(tc.kind, h, ds))
			assertLayout(t, got, tc.header, tc.detail, len(ds))
			golden := filepath.Join("testdata", strings.ToLower(string(tc.kind))+".golden")
			if *update {
				if err := os.WriteFile(golden, got, 0o644); err != nil {
					t.Fatal(err)
				}
			}
			want, err := os.ReadFile(golden)
			if err != nil {
				t.Fatalf("%v (สร้างด้วย go test ./internal/rdfile -run TestGolden -update แล้วตรวจเนื้อไฟล์ด้วยตา)", err)
			}
			if !bytes.Equal(got, want) {
				t.Fatalf("golden mismatch\n got: %q\nwant: %q", got, want)
			}
		})
	}
}

func assertLayout(t *testing.T, file []byte, headerFields, detailFields, details int) {
	t.Helper()
	if !bytes.HasPrefix(file, utf8BOM) {
		t.Fatal("missing UTF-8 BOM")
	}
	if bytes.HasSuffix(file, []byte("\n")) || bytes.HasSuffix(file, []byte("\r")) {
		t.Fatal("file must not end with CRLF")
	}
	body := string(file[len(utf8BOM):])
	if strings.Count(body, "\n") != strings.Count(body, "\r\n") {
		t.Fatal("line breaks must be CRLF")
	}
	lines := strings.Split(body, "\r\n")
	if len(lines) != details+1 {
		t.Fatalf("lines = %d, want 1 H + %d D", len(lines), details)
	}
	for i, line := range lines {
		want, tag := detailFields, "D"
		if i == 0 {
			want, tag = headerFields, "H"
		}
		f := strings.Split(line, "|")
		if len(f) != want || f[0] != tag {
			t.Fatalf("line %d: %d fields tag %q, want %d %q", i+1, len(f), f[0], want, tag)
		}
	}
}

func TestSummarizeTotalsAndCount(t *testing.T) {
	h, ds := pnd3Sample()
	// 0.20 + 12000.00 = 12000.20 (decimal ไม่ใช่ float), ภาษี 600.00 + เงินเพิ่ม
	if h.TotNum != "2" || h.TotAmt != "12000.20" || h.TotTax != "600.00" || h.GTotTax != "600.00" || h.TransAmt != "0.00" {
		t.Fatalf("header totals = %+v", h)
	}
	h.SurAmt = "12.34"
	h = Summarize(PND3, h, ds)
	if h.GTotTax != "612.34" {
		t.Fatalf("GTOT_TAX = %s", h.GTotTax)
	}
	// กฎข้อ 20: 0.10 + 0.20 ต้องได้ 0.30 พอดี
	sum := Summarize(PND53, Header{}, []Detail{{Items: [3]Item{{PaidAmt: "0.10", TaxAmt: "0.10"}, {PaidAmt: "0.20", TaxAmt: "0.20"}}}})
	if sum.TotAmt != "0.30" || sum.TotTax != "0.30" || sum.TotNum != "1" {
		t.Fatalf("0.10+0.20 = %+v", sum)
	}
}

func TestEmptyItemsWrittenAsZero(t *testing.T) {
	_, ds := pnd53Sample()
	f := DetailFields(PND53, ds[1])
	// รายการที่ 2 และ 3 ว่าง → 00000000|0.00|0.00|0.00||
	if got := strings.Join(f[14:26], "|"); got != "00000000|0.00|0.00|0.00|||00000000|0.00|0.00|0.00||" {
		t.Fatalf("empty items = %q", got)
	}
	if f[4] != NoTIN || f[7] != "" {
		t.Fatalf("TIN/SNAME = %q/%q", f[4], f[7])
	}
}

func TestFileName(t *testing.T) {
	h := baseHeader("08", "2569")
	h.NID = "0105555555554"
	if got := FileName(PND53, h); got != "PND53_0105555555554_000000_2569_08_00_00.txt" {
		t.Fatalf("file name = %s", got)
	}
	h.FormType, h.SubmissionNo, h.BranchNo = "03", "12", "000001"
	if got := FileName(PND2, h); got != "PND2_0105555555554_000001_2569_08_03_12.txt" {
		t.Fatalf("file name = %s", got)
	}
}

func issueKeys(issues []Issue) []string {
	out := make([]string, 0, len(issues))
	for _, i := range issues {
		out = append(out, i.Key+"@"+i.Field)
	}
	return out
}

func hasIssue(issues []Issue, key, field string, row int) bool {
	for _, i := range issues {
		if i.Key == key && i.Field == field && i.Row == row {
			return true
		}
	}
	return false
}

func TestCheckForbiddenCharacters(t *testing.T) {
	for _, ch := range append(strings.Split(forbiddenChars, ""), "\t") {
		h, ds := pnd53Sample()
		ds[0].FirstName = "สยาม" + ch + "ขนส่ง"
		issues := Check(PND53, h, ds)
		want := ch
		if ch == "\t" {
			want = "U+0009"
		}
		if len(issues) != 1 || !hasIssue(issues, "tax_rdfile_forbidden_char", "name", 1) || issues[0].Args["char"] != want {
			t.Fatalf("char %q → %+v", ch, issues)
		}
	}
	// ช่องอื่น ๆ ก็ตรวจ: ประเภทเงินได้ของรายการที่ 2, ชื่อแผนก, เลขอ้างอิง
	h, ds := pnd53Sample()
	ds[0].Items[1].IncType = "ค่าบริการ/ค่าแรง"
	h.DeptName = "ฝ่ายบัญชี & การเงิน"
	h.UserID = "REF,1"
	issues := Check(PND53, h, ds)
	for _, want := range []struct {
		field string
		row   int
	}{{"l2_income_type", 1}, {"dept_name", 0}, {"media_ref_no", 0}} {
		if !hasIssue(issues, "tax_rdfile_forbidden_char", want.field, want.row) {
			t.Fatalf("missing forbidden char on %s: %v", want.field, issueKeys(issues))
		}
	}
	// จุด ขีด วงเล็บ และช่องว่าง ไม่ต้องห้าม (ไฟล์ตัวอย่างทางการมี "หจก." และ "เลขที่ 1")
	h, ds = pnd53Sample()
	ds[0].FirstName = "ส.ขนส่ง (ประเทศไทย)-1 จำกัด"
	if issues := Check(PND53, h, ds); len(issues) != 0 {
		t.Fatalf("allowed punctuation flagged: %+v", issues)
	}
}

func TestCheckLengthCountsRunes(t *testing.T) {
	h, ds := pnd53Sample()
	ds[0].FirstName = strings.Repeat("ก", 100) // 100 ตัวอักษร = 300 byte ต้องผ่าน
	if issues := Check(PND53, h, ds); len(issues) != 0 {
		t.Fatalf("100 runes flagged: %+v", issues)
	}
	ds[0].FirstName = strings.Repeat("ก", 101)
	issues := Check(PND53, h, ds)
	if len(issues) != 1 || !hasIssue(issues, "tax_rdfile_too_long", "name", 1) || issues[0].Args["max"] != "100" {
		t.Fatalf("101 runes → %+v", issues)
	}
	h.UserID = strings.Repeat("9", 21)
	if !hasIssue(Check(PND53, h, ds), "tax_rdfile_too_long", "media_ref_no", 0) {
		t.Fatal("USER_ID > 20 not flagged")
	}
}

func TestCheckTaxIDs(t *testing.T) {
	h, ds := pnd53Sample()
	h.NID = "0105558012340" // หลักตรวจสอบผิด
	ds[1].NID = "010553514196"
	issues := Check(PND53, h, ds)
	if !hasIssue(issues, "tax_rdfile_company_taxid", "tax_id", 0) || !hasIssue(issues, "tax_rdfile_taxid_invalid", "tax_id", 2) || len(issues) != 2 {
		t.Fatalf("issues = %v", issueKeys(issues))
	}
	// ภ.ง.ด.2: เลขศูนย์ 13 หลักได้เฉพาะรหัส 2 (ดอกเบี้ย 40(4)(ก))
	h, ds = pnd2Sample()
	if issues := Check(PND2, h, ds); len(issues) != 0 {
		t.Fatalf("zero PIN with interest flagged: %+v", issues)
	}
	ds[1].Items[0].IncType = "1"
	if !hasIssue(Check(PND2, h, ds), "tax_rdfile_taxid_invalid", "tax_id", 2) {
		t.Fatal("zero PIN with royalty must be rejected")
	}
}

func TestCheckAmounts(t *testing.T) {
	for _, bad := range []string{"-1.00", "1.000", "1,000.00", "100", "1234567890123456.00", ""} {
		h, ds := pnd53Sample()
		ds[0].Items[0].PaidAmt = bad
		if !hasIssue(Check(PND53, h, ds), "tax_rdfile_amount_invalid", "l1_amount", 1) {
			t.Fatalf("amount %q not flagged", bad)
		}
	}
	h, ds := pnd53Sample()
	ds[0].Items[0].PaidAmt = "123456789012345.00" // 18 ตัวอักษรพอดี
	ds[0].Items[0].TaxAmt = "0.00"
	h = Summarize(PND53, h, ds)
	for _, i := range Check(PND53, h, ds) {
		if i.Field == "l1_amount" || i.Field == "l1_tax" {
			t.Fatalf("18-char amount flagged: %+v", i)
		}
	}
	h.SurAmt = "-5.00"
	if !hasIssue(Check(PND53, h, ds), "tax_rdfile_amount_invalid", "surcharge", 0) {
		t.Fatal("negative surcharge not flagged")
	}
}

func TestCheckDatesAndItems(t *testing.T) {
	cases := map[string]string{"31022569": "tax_rdfile_date_invalid", "15082569": "tax_rdfile_date_outside_month",
		"15092568": "tax_rdfile_date_outside_month", "1509256": "tax_rdfile_date_invalid", "00000000": "tax_rdfile_date_invalid"}
	for date, key := range cases {
		h, ds := pnd53Sample()
		ds[0].Items[0].PaidDate = date
		if issues := Check(PND53, h, ds); len(issues) != 1 || !hasIssue(issues, key, "l1_date", 1) {
			t.Fatalf("date %s → %+v", date, issues)
		}
	}
	h, ds := pnd53Sample()
	ds[0].Items[0].PaidDate = "29022568" // 2568 = ค.ศ. 2025 ไม่ใช่ปีอธิกสุรทิน
	if !hasIssue(Check(PND53, h, ds), "tax_rdfile_date_invalid", "l1_date", 1) {
		t.Fatal("29/02/2568 must be invalid")
	}
	// รายการ 1 ว่างแต่รายการ 2 มีค่า → ต้องมีรายการที่ 1
	h, ds = pnd53Sample()
	ds[0].Items[0] = Item{}
	if issues := Check(PND53, h, ds); len(issues) != 1 || !hasIssue(issues, "tax_rdfile_item1_required", "l1_date", 1) {
		t.Fatalf("item 1 empty → %+v", issues)
	}
	// รายการที่มีข้อมูลบางช่องต้องครบทุกช่อง
	h, ds = pnd53Sample()
	ds[1].Items[2] = Item{TaxRate: "abc", PayCon: "4"}
	issues := Check(PND53, h, ds)
	for _, field := range []string{"l3_date", "l3_rate", "l3_amount", "l3_tax", "l3_income_type", "l3_condition"} {
		found := false
		for _, i := range issues {
			found = found || (i.Field == field && i.Row == 2)
		}
		if !found {
			t.Fatalf("partial item 3 missing %s: %v", field, issueKeys(issues))
		}
	}
	h, ds = pnd53Sample()
	ds[0].Items[0].TaxRate = "100.01"
	if !hasIssue(Check(PND53, h, ds), "tax_rdfile_rate_invalid", "l1_rate", 1) {
		t.Fatal("rate > 100 not flagged")
	}
}

func TestCheckHeaderRequirements(t *testing.T) {
	h, ds := pnd53Sample()
	h.BranchNo, h.DeptName, h.UserID, h.SubmissionNo, h.LTO, h.BranchType = "00000", "", "", "1", "2", "X"
	h.Sections = [3]string{"0", "0", "0"}
	issues := Check(PND53, h, ds)
	for _, want := range [][2]string{{"tax_rdfile_branch_required", "branch_no"}, {"tax_rdfile_dept_name_required", "dept_name"},
		{"tax_rdfile_user_id_required", "media_ref_no"}, {"tax_rdfile_submission_invalid", "submission_no"},
		{"tax_rdfile_section_required", "tax_section"}, {"tax_rdfile_value_invalid", "lto"}, {"tax_rdfile_value_invalid", "branch_type"}} {
		if !hasIssue(issues, want[0], want[1], 0) {
			t.Fatalf("missing %v in %v", want, issueKeys(issues))
		}
	}
	if issues := Check(PND53, h, nil); !hasIssue(issues, "tax_rdfile_no_rows", "", 0) {
		t.Fatal("no rows not flagged")
	}
	// ภ.ง.ด.2 ไม่มีธงมาตรา
	h2, ds2 := pnd2Sample()
	if issues := Check(PND2, h2, ds2); len(issues) != 0 {
		t.Fatalf("pnd2 without sections flagged: %+v", issues)
	}
	if issues := Check(Kind("PND1"), h2, ds2); len(issues) != 1 || issues[0].Key != "tax_rdfile_unsupported_form" {
		t.Fatalf("unsupported kind → %+v", issues)
	}
}

func TestCheckNamesAndAddress(t *testing.T) {
	h, ds := pnd3Sample()
	ds[0].Title, ds[0].FirstName = "", " "
	ds[1].Address = Address{PostalCode: "1100"}
	issues := Check(PND3, h, ds)
	for _, want := range []struct {
		key, field string
		row        int
	}{{"tax_rdfile_title_required", "title", 1}, {"tax_rdfile_name_required", "name", 1},
		{"tax_rdfile_address_required", "addr_district", 2}, {"tax_rdfile_address_required", "addr_province", 2},
		{"tax_rdfile_postcode_invalid", "addr_postcode", 2}} {
		if !hasIssue(issues, want.key, want.field, want.row) {
			t.Fatalf("missing %+v in %v", want, issueKeys(issues))
		}
	}
	// ภ.ง.ด.3 คำนำหน้า "-" ผ่าน; ภ.ง.ด.53 ไม่บังคับที่อยู่
	h, ds = pnd3Sample()
	ds[0].Title = "-"
	if issues := Check(PND3, h, ds); len(issues) != 0 {
		t.Fatalf("title '-' flagged: %+v", issues)
	}
	h, ds = pnd53Sample()
	if issues := Check(PND53, h, ds); len(issues) != 0 || ds[0].Address != (Address{}) {
		t.Fatalf("pnd53 address must be optional: %+v", issues)
	}
	h, ds = pnd3Sample()
	ds[1].Address.PostalCode = ""
	if !hasIssue(Check(PND3, h, ds), "tax_rdfile_address_required", "addr_postcode", 2) {
		t.Fatal("pnd3 postcode required")
	}
}

// TestOfficialSampleRoundTrip - ไฟล์ตัวอย่างทางการของกรมสรรพากร (swc_Data_Test.zip) แยกช่องแล้วเขียนกลับต้องได้ byte เท่ากันทุกไฟล์
// ภ.ง.ด.53/3/2 ทั้งแบบดิบ และผ่าน Header/Detail (พิสูจน์ลำดับช่องของ HeaderFields/DetailFields) — ตัวอย่างมี NID "TTTT…" และช่องว่างนำหน้า
// จึงไม่ผ่าน Check (ตั้งใจ) ใช้เฉพาะตัวเขียน
//
//	BC_RDFILE_SAMPLE_ZIP=C:/tmp/rdfile/swc_Data_Test.zip go test ./internal/rdfile -run TestOfficialSampleRoundTrip -v
func TestOfficialSampleRoundTrip(t *testing.T) {
	zipPath := os.Getenv("BC_RDFILE_SAMPLE_ZIP")
	if zipPath == "" {
		t.Skip("set BC_RDFILE_SAMPLE_ZIP (https://www.rd.go.th/fileadmin/user_upload/WHT/Download/swc_Data_Test.zip)")
	}
	z, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	seen := map[Kind]bool{}
	for _, entry := range z.File {
		name := path.Base(entry.Name)
		var kind Kind
		for _, k := range []Kind{PND53, PND3, PND2} {
			if strings.HasPrefix(name, string(k)+"_") {
				kind = k
			}
		}
		if kind == "" {
			continue
		}
		rc, err := entry.Open()
		if err != nil {
			t.Fatal(err)
		}
		raw, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.HasPrefix(raw, utf8BOM) {
			t.Fatalf("%s: official sample has no BOM", name)
		}
		var lines [][]string
		for _, line := range strings.Split(string(raw[len(utf8BOM):]), "\r\n") {
			lines = append(lines, strings.Split(line, "|"))
		}
		if got := Encode(lines); !bytes.Equal(got, raw) {
			t.Fatalf("%s: raw round trip differs", name)
		}
		h := headerFromFields(t, kind, lines[0])
		var ds []Detail
		for _, f := range lines[1:] {
			ds = append(ds, detailFromFields(t, kind, f))
		}
		if got := Encode(Lines(kind, h, ds)); !bytes.Equal(got, raw) {
			t.Fatalf("%s: struct round trip differs\n got: %q\nwant: %q", name, got, raw)
		}
		wantH, wantD := headerFieldsWHT, detailFieldsWHT
		if kind == PND2 {
			wantH, wantD = headerFieldsPND2, detailFieldsPND2
		}
		assertLayout(t, raw, wantH, wantD, len(ds))
		// TOT_NUM ของไฟล์ตัวอย่าง = จำนวนบรรทัด D (เหมือน Summarize)
		if h.TotNum != Summarize(kind, h, ds).TotNum {
			t.Fatalf("%s: TOT_NUM %s ≠ D lines %d", name, h.TotNum, len(ds))
		}
		seen[kind] = true
		t.Logf("%s: %d D lines, %d bytes identical", name, len(ds), len(raw))
	}
	for _, k := range []Kind{PND53, PND3, PND2} {
		if !seen[k] {
			t.Fatalf("sample for %s not found in %s", k, zipPath)
		}
	}
}

// headerFromFields - อ่านบรรทัด H ตามลำดับช่องในเอกสาร Format (ชื่อช่องทางการกำกับทุกตำแหน่ง)
func headerFromFields(t *testing.T, kind Kind, f []string) Header {
	t.Helper()
	want := headerFieldsWHT
	if kind == PND2 {
		want = headerFieldsPND2
	}
	if len(f) != want || f[0] != "H" || f[5] != string(kind) {
		t.Fatalf("%s header: %d fields %q", kind, len(f), f)
	}
	h := Header{SenderID: f[1], SenderNID: f[2], SenderBranch: f[3], SenderRole: f[4], NID: f[6], BranchNo: f[7], DeptName: f[8]}
	rest := f[9:]
	if kind != PND2 {
		h.Sections = [3]string{f[9], f[10], f[11]} // SECTION3 / SECTION65|48 / SECTION69|50
		rest = f[12:]
	}
	// LTO, TAX_MONTH, TAX_YEAR, BRANCH_TYPE, FORM_TYPE, TOT_NUM, TOT_AMT, TOT_TAX, SUR_AMT, GTOT_TAX, TRANS_AMT, USER_ID, FORM_FLAG
	h.LTO, h.TaxMonth, h.TaxYear, h.BranchType, h.FormType, h.TotNum = rest[0], rest[1], rest[2], rest[3], rest[4], rest[5]
	h.TotAmt, h.TotTax, h.SurAmt, h.GTotTax, h.TransAmt, h.UserID, h.FormFlag = rest[6], rest[7], rest[8], rest[9], rest[10], rest[11], rest[12]
	return h
}

// detailFromFields - อ่านบรรทัด D: DETAIL, SEQ_NO, BRANCH_NO, NID|PIN, TIN, [ACC_NO], TITLE_NAME, FNAME, SNAME,
// รายการเงินได้ (PAID_DATE, TAX_RATE, PAID_AMT, TAX_AMT, INC_TYPE, PAY_CON) × 3 (ภ.ง.ด.2 × 1), ที่อยู่ 12 ช่อง
func detailFromFields(t *testing.T, kind Kind, f []string) Detail {
	t.Helper()
	want, items := detailFieldsWHT, 3
	if kind == PND2 {
		want, items = detailFieldsPND2, 1
	}
	if len(f) != want || f[0] != "D" {
		t.Fatalf("%s detail: %d fields %q", kind, len(f), f)
	}
	d := Detail{Seq: f[1], BranchNo: f[2], NID: f[3], TIN: f[4]}
	rest := f[5:]
	if kind == PND2 {
		d.AccNo, rest = rest[0], rest[1:]
	}
	d.Title, d.FirstName, d.LastName, rest = rest[0], rest[1], rest[2], rest[3:]
	for i := 0; i < items; i++ {
		d.Items[i] = Item{PaidDate: rest[0], TaxRate: rest[1], PaidAmt: rest[2], TaxAmt: rest[3], IncType: rest[4], PayCon: rest[5]}
		rest = rest[6:]
	}
	d.Address = Address{Building: rest[0], Room: rest[1], Floor: rest[2], Village: rest[3], No: rest[4], Moo: rest[5], Soi: rest[6],
		Street: rest[7], Tambon: rest[8], Amphur: rest[9], Province: rest[10], PostalCode: rest[11]}
	return d
}

// ภ.ง.ด.2 มีรายการเงินได้เดียวต่อบรรทัด D: ยอดหัวแบบนับเฉพาะ Items[0] ให้ตรงกับเนื้อไฟล์ (Items[1] ที่หลงมาไม่ถูกนับ)
func TestSummarizePND2FirstItemOnly(t *testing.T) {
	h, ds := pnd2Sample()
	ds[0].Items[1] = item("02062569", "15.00", "999.00", "149.85", "2", "1")
	h = Summarize(PND2, h, ds)
	if h.TotAmt != "4000.00" || h.TotTax != "450.00" || h.GTotTax != "450.00" {
		t.Fatalf("PND2 header totals must follow the D lines: %+v", h)
	}
	// ภ.ง.ด.53 ยังรวมทั้ง 3 รายการของบรรทัด
	h53, _ := pnd53Sample()
	if h53.TotAmt != "35000.10" || h53.TotTax != "1250.00" {
		t.Fatalf("PND53 totals = %+v", h53)
	}
}

// [F2] D6: เงินปันผล 40(4)(ข) (รหัส 3) ต้องมีเลขที่บัญชีเงินฝาก; ดอกเบี้ย 40(4)(ก) (รหัส 2) ไม่บังคับ
// เพราะรหัส 2 รวมดอกเบี้ยพันธบัตร/ตั๋วเงินที่ไม่มีบัญชีเงินฝาก
func TestCheckPND2AccountNumber(t *testing.T) {
	h, ds := pnd2Sample()
	ds[2].AccNo = ""
	if issues := Check(PND2, h, ds); !hasIssue(issues, "tax_rdfile_account_required", "bank_account_no", 3) || len(issues) != 1 {
		t.Fatalf("dividend without account: %+v", issues)
	}
	h, ds = pnd2Sample()
	ds[0].AccNo = ""
	if issues := Check(PND2, h, ds); len(issues) != 0 {
		t.Fatalf("interest without account must pass: %+v", issues)
	}
}

// ไม่มีรายการ: ยื่นหัวแบบอย่างเดียว (TOT_NUM 0) ได้เฉพาะฉบับที่ผู้เรียกยืนยันว่าเป็นฉบับไม่มีเงินได้ (NilReturn)
func TestCheckNilReturnHeaderOnly(t *testing.T) {
	h := baseHeader("09", "2569")
	h.Sections = [3]string{"1", "0", "0"}
	h = Summarize(PND53, h, nil)
	if issues := Check(PND53, h, nil); !hasIssue(issues, "tax_rdfile_no_rows", "", 0) {
		t.Fatalf("no rows without NilReturn: %+v", issues)
	}
	h.NilReturn = true
	if issues := Check(PND53, h, nil); len(issues) != 0 {
		t.Fatalf("nil return must pass: %+v", issues)
	}
	file := Encode(Lines(PND53, h, nil))
	assertLayout(t, file, headerFieldsWHT, detailFieldsWHT, 0)
	fields := strings.Split(string(file[len(utf8BOM):]), "|")
	if fields[17] != "0" || fields[18] != "0.00" || fields[19] != "0.00" || fields[21] != "0.00" {
		t.Fatalf("header-only totals: TOT_NUM=%s TOT_AMT=%s TOT_TAX=%s GTOT_TAX=%s", fields[17], fields[18], fields[19], fields[21])
	}
}

// ช่องว่างหัว/ท้ายถูกตัดก่อน Check/Encode: ช่องที่มีแต่ช่องว่างเป็นช่องว่างจริง ("||") ช่องว่างกลางข้อความคงไว้
func TestTrimFields(t *testing.T) {
	h, ds := pnd3Sample()
	h.DeptName, h.UserID = "  สำนักงานใหญ่ ", " ZXC1234567 "
	ds[0].LastName, ds[0].FirstName, ds[0].Title = "  ", " สมชาย  ใจงาม ", " นาย"
	ds[0].Address.Amphur, ds[0].Items[0].IncType = " เขตบางนา ", "ค่าจ้างทำของ "
	h, ds = Trim(h, ds)
	if issues := Check(PND3, h, ds); len(issues) != 0 {
		t.Fatalf("trimmed sample must pass: %+v", issues)
	}
	f := DetailFields(PND3, ds[0])
	if f[5] != "นาย" || f[6] != "สมชาย  ใจงาม" || f[7] != "" || f[12] != "ค่าจ้างทำของ" || f[35] != "เขตบางนา" {
		t.Fatalf("trimmed D fields = %q", f)
	}
	if h.DeptName != "สำนักงานใหญ่" || h.UserID != "ZXC1234567" {
		t.Fatalf("trimmed header = %+v", h)
	}
	if !strings.Contains(string(Encode(Lines(PND3, h, ds))), "|สมชาย  ใจงาม||") {
		t.Fatal("blank surname must be written as an empty field")
	}
}
