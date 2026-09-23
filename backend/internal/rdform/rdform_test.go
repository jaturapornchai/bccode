package rdform

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/api"

	"smlcloudplatform/internal/pdftext"
)

// sampleValue - ค่าตัวอย่างตามชนิดช่อง ให้ทุกช่องมีข้อความสำหรับเปิดดูตำแหน่งบนแบบ
func sampleValue(typ, key string, cells int) string {
	switch typ {
	case "taxid":
		return "0105556012345"
	case "digits":
		if cells == 0 {
			cells = 5
		}
		return strings.Repeat("0", max(0, cells-1)) + "1"
	case "money":
		return "1234567.89"
	case "int":
		return "12"
	case "check":
		return "1"
	}
	switch {
	case strings.Contains(key, "name"):
		return "บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด"
	case strings.Contains(key, "date"), strings.Contains(key, "day"):
		return "15"
	}
	return "ถนนพหลโยธิน"
}

func sampleDocument(t *testing.T, code string) (Document, int) {
	t.Helper()
	s, err := Lookup(code)
	if err != nil {
		t.Fatal(err)
	}
	doc := Document{Values: map[string]string{}}
	fill := func(fs []Field) {
		for _, f := range fs {
			if f.Type == "choice" {
				doc.Values[f.Key] = f.Options[len(f.Options)-1].Value
				continue
			}
			if _, ok := doc.Values[f.Key]; !ok {
				doc.Values[f.Key] = sampleValue(f.Type, f.Key, len(f.Cells))
			}
		}
	}
	fill(s.Fields)
	attach := Attachment(code)
	if attach == nil {
		return doc, 0
	}
	rowChoice := map[string]string{}
	for _, f := range attach.Fields {
		if attach.Table != nil && f.Type == "choice" && s.Field(f.Key) == nil {
			rowChoice[f.Key] = f.Options[0].Value // ค่ารายการ (แยกแผ่นตามค่านี้) ไม่ใช่ค่าหัวแบบ
			continue
		}
		fill([]Field{f})
	}
	if attach.Table == nil {
		doc.Sheets = []map[string]string{{"name": "สาขาลาดหลุมแก้ว"}, {"name": "สาขาบางนา"}}
		return doc, 2
	}
	n := len(attach.Table.Rows) + 1 // เกินหนึ่งแผ่น → ต้องได้ 2 แผ่น
	for i := 0; i < n; i++ {
		row := map[string]string{}
		for _, c := range attach.Table.Columns {
			cells := len(attach.Table.Rows[0][c.Key].Cells)
			row[c.Key] = sampleValue(c.Type, c.Key, cells)
		}
		if _, ok := row["seq"]; ok {
			row["seq"] = "" // ให้ตัวกรอกเรียงลำดับเอง
		}
		for k, v := range rowChoice {
			row[k] = v
		}
		doc.Rows = append(doc.Rows, row)
	}
	return doc, 2
}

func TestRenderEveryForm(t *testing.T) {
	dir := os.Getenv("RDFORM_SAMPLE_DIR") // ตั้งเพื่อบันทึก PDF ตัวอย่างไว้เปิดดูตำแหน่ง
	codes, err := Forms()
	if err != nil {
		t.Fatal(err)
	}
	if len(codes) == 0 {
		t.Fatal("no forms embedded")
	}
	for _, code := range codes {
		t.Run(code, func(t *testing.T) {
			doc, sheets := sampleDocument(t, code)
			if err := Validate(code, doc); err != nil {
				t.Fatalf("sample document invalid: %v", err)
			}
			pdf, err := Render(code, doc)
			if err != nil {
				t.Fatalf("render: %v", err)
			}
			if err := api.Validate(bytes.NewReader(pdf), pdftext.Config()); err != nil {
				t.Fatalf("invalid pdf: %v", err)
			}
			cover, _ := Lookup(code)
			want := len(cover.Pages)
			if a := Attachment(code); a != nil {
				want += sheets * len(a.Pages)
			}
			pages, err := api.PageCount(bytes.NewReader(pdf), pdftext.Config())
			if err != nil {
				t.Fatal(err)
			}
			if pages != want {
				t.Fatalf("pages = %d, want %d", pages, want)
			}
			if fields, err := api.FormFields(bytes.NewReader(pdf), pdftext.Config()); err == nil && len(fields) > 0 {
				t.Fatalf("output still has %d AcroForm fields", len(fields))
			}
			if dir != "" {
				if err := os.WriteFile(filepath.Join(dir, code+".pdf"), pdf, 0o644); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func TestPaginate(t *testing.T) {
	attach := &Spec{
		Code:   "x_attach",
		Fields: []Field{{Key: "income_type", Type: "choice"}, {Key: "page_total_amount", Type: "money"}, {Key: "page_total_tax", Type: "money"}},
		Table: &Table{
			Columns: []Column{{Key: "seq", Type: "int"}, {Key: "l1_amount", Type: "money"}, {Key: "l2_amount", Type: "money"}, {Key: "l1_tax", Type: "money"}},
			Rows:    []map[string]Box{{}, {}},
		},
	}
	rows := []map[string]string{
		{"income_type": "interest", "l1_amount": "0.1", "l2_amount": "0.2", "l1_tax": "1,000.005"},
		{"income_type": "dividend", "l1_amount": "5"},
		{"income_type": "interest", "l1_amount": "7"},
		{"income_type": "interest", "l1_amount": "9", "seq": "99"},
	}
	sheets, err := Paginate(attach, Document{Values: map[string]string{"tax_id": "1"}, Rows: rows})
	if err != nil {
		t.Fatal(err)
	}
	// interest: 3 แถว → 2 แผ่น, dividend: 1 แผ่น — ประเภทเงินได้ต่างกันห้ามอยู่แผ่นเดียวกัน
	if len(sheets) != 3 {
		t.Fatalf("sheets = %d, want 3", len(sheets))
	}
	got := []string{}
	for _, s := range sheets {
		seqs := []string{}
		for _, r := range s.Rows {
			seqs = append(seqs, r["seq"])
		}
		got = append(got, s.Values["income_type"]+":"+s.Values["sheet_no"]+"/"+s.Values["sheet_total"]+":"+strings.Join(seqs, ",")+":"+s.Values["page_total_amount"]+":"+s.Values["page_total_tax"]+":"+s.Values["tax_id"])
	}
	want := []string{"interest:1/3:1,2:7.30:1000.01:1", "interest:2/3:99:9.00:0.00:1", "dividend:3/3:4:5.00:0.00:1"}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("got  %v\nwant %v", got, want)
	}
	if rows[0]["seq"] != "" {
		t.Fatal("Paginate must not mutate the caller's rows")
	}
}

func TestPaginateRejectsBadMoney(t *testing.T) {
	attach := &Spec{
		Fields: []Field{{Key: "page_total_amount", Type: "money"}},
		Table:  &Table{Columns: []Column{{Key: "l1_amount", Type: "money"}}, Rows: []map[string]Box{{}}},
	}
	_, err := Paginate(attach, Document{Rows: []map[string]string{{"l1_amount": "12a"}}})
	if !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("err = %v, want ErrInvalidValue", err)
	}
}

func TestThousands(t *testing.T) {
	for in, want := range map[string]string{"0": "0", "999": "999", "1000": "1,000", "1234567": "1,234,567", "-1234": "-1,234"} {
		if got := thousands(in); got != want {
			t.Errorf("thousands(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestUnknownForm(t *testing.T) {
	if _, err := Render("pnd1", Document{}); !errors.Is(err, ErrUnknownForm) {
		t.Fatalf("err = %v, want ErrUnknownForm (ภ.ง.ด.1 = เงินเดือน อยู่นอกขอบเขต)", err)
	}
}

func TestSpecsIntegrity(t *testing.T) {
	codes, _ := Forms()
	for _, code := range codes {
		for _, s := range []*Spec{mustLookup(t, code), Attachment(code)} {
			if s == nil {
				continue
			}
			for _, f := range s.Fields {
				boxes := []Box{f.Box}
				if f.Type == "choice" {
					boxes = boxes[:0]
					for _, o := range f.Options {
						boxes = append(boxes, o.Box)
					}
				}
				for _, b := range boxes {
					ok := false
					for _, p := range s.Pages {
						ok = ok || p == b.Page
					}
					if !ok {
						t.Errorf("%s.%s on page %d outside %v", s.Code, f.Key, b.Page, s.Pages)
					}
				}
				if f.Type == "taxid" && len(f.Cells) != 13 && f.Comb == 0 {
					t.Errorf("%s.%s: tax id box without 13 cells or comb", s.Code, f.Key)
				}
			}
		}
	}
	if len(codes) < 5 {
		t.Fatalf("only %d forms: %s", len(codes), strconv.Quote(strings.Join(codes, ",")))
	}
}

func mustLookup(t *testing.T, code string) *Spec {
	t.Helper()
	s, err := Lookup(code)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestSchemaEveryForm(t *testing.T) {
	codes, _ := Forms()
	for _, code := range codes {
		s, err := SchemaFor(code)
		if err != nil {
			t.Fatal(err)
		}
		seen := map[string]bool{}
		for _, f := range s.Fields {
			if seen[f.Key] || f.Label == "" {
				t.Errorf("%s: field %q duplicated or unlabelled", code, f.Key)
			}
			seen[f.Key] = true
			if autoKey(f.Key) {
				t.Errorf("%s: auto key %q exposed to the editor", code, f.Key)
			}
		}
		if a := s.Attachment; a != nil && len(a.Columns) == 0 {
			t.Errorf("%s: attachment without editable columns", code)
		}
	}
	pnd2, _ := SchemaFor("pnd2")
	if pnd2.Attachment == nil || pnd2.Attachment.Columns[0].Key != "income_type" || len(pnd2.Attachment.Columns[0].Options) != 5 {
		t.Fatalf("pnd2 attachment must start with the per-sheet income type choice: %+v", pnd2.Attachment)
	}
	pnd50, _ := SchemaFor("pnd50")
	keys := map[string]string{}
	for _, f := range pnd50.Fields {
		keys[f.Key] = f.Type
	}
	if keys["overpaid_tax"] != "money" || keys["overpaid_tax_baht"] != "" || keys["overpaid_tax_satang"] != "" {
		t.Fatalf("pnd50 baht/satang boxes must be one money field in the editor")
	}
	pbt40, _ := SchemaFor("pbt40")
	if pbt40.Attachment == nil || pbt40.Attachment.Mode != "sheets" {
		t.Fatalf("pbt40 attachment must be one sheet per establishment")
	}
}

func TestSplitMoneyPairs(t *testing.T) {
	s := mustLookup(t, "pnd51")
	v := map[string]string{"extra_tax_paid": "1234.5", "fx_tax_thb": "-7"}
	if err := splitMoneyPairs(s, v); err != nil {
		t.Fatal(err)
	}
	if v["extra_tax_paid_baht"] != "1234" || v["extra_tax_paid_satang"] != "50" || v["fx_tax_thb_baht"] != "-7" {
		t.Fatalf("digits pair split wrong: %v", v)
	}
	s50 := mustLookup(t, "pnd50")
	v = map[string]string{"overpaid_tax": "1234567.891"}
	if err := splitMoneyPairs(s50, v); err != nil {
		t.Fatal(err)
	}
	if v["overpaid_tax_baht"] != "1,234,567" || v["overpaid_tax_satang"] != "89" {
		t.Fatalf("int pair split wrong: %v", v)
	}
	if err := splitMoneyPairs(s50, map[string]string{"overpaid_tax": "x"}); !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("err = %v", err)
	}
}

func TestValidateRejects(t *testing.T) {
	cases := map[string]Document{
		"unknown key":       {Values: map[string]string{"no_such_box": "1"}},
		"bad money":         {Values: map[string]string{"total_tax": "12,3x"}},
		"bad choice":        {Values: map[string]string{"month": "13"}},
		"short tax id":      {Values: map[string]string{"tax_id": "01055"}},
		"newline":           {Values: map[string]string{"name": "a\nb"}},
		"bad row money":     {Rows: []map[string]string{{"l1_amount": "abc"}}},
		"sheets on a table": {Sheets: []map[string]string{{"name": "x"}}},
	}
	for name, doc := range cases {
		if err := Validate("pnd53", doc); !errors.Is(err, ErrInvalidValue) {
			t.Errorf("%s: err = %v, want ErrInvalidValue", name, err)
		}
	}
	ok := Document{Values: map[string]string{"tax_id": "0-1055-56012-34-5", "total_tax": "1,234.50", "month": "9"}, Rows: []map[string]string{{"l1_amount": "100"}}}
	if err := Validate("pnd53", ok); err != nil {
		t.Fatalf("valid document rejected: %v", err)
	}
	if err := Validate("pnd2", Document{Rows: []map[string]string{{"income_type": "interest"}}}); err != nil {
		t.Fatalf("pnd2 per-row income type rejected: %v", err)
	}
}

func TestDigitGroups(t *testing.T) {
	s, _ := SchemaFor("pp30")
	found := false
	for _, c := range s.Attachment.Columns {
		if strings.HasPrefix(c.Key, "branch_d") {
			t.Fatalf("per-digit column %s exposed", c.Key)
		}
		found = found || (c.Key == "branch" && c.Type == "digits")
	}
	if !found {
		t.Fatal("pp30 attachment must expose one branch column")
	}
	row := expandDigitGroups(mustLookup(t, "pp30_attach").Table.DigitGroups(), map[string]string{"branch": "12"})
	if row["branch_d4"] != "1" || row["branch_d5"] != "2" || row["branch_d1"] != "" {
		t.Fatalf("branch digits not right-aligned: %v", row)
	}
	if err := Validate("pp30", Document{Rows: []map[string]string{{"branch": "00001", "sales_amount": "10"}}}); err != nil {
		t.Fatal(err)
	}
}
