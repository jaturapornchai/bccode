package report

import (
	"bytes"
	"strings"
	"testing"
)

func TestWriteCSVQuotesGuardsAndLabels(t *testing.T) {
	resp := RunResponse{
		Columns: []ColumnView{{Key: "docno", Label: "เลขที่เอกสาร", Type: ColText}, {Key: "description", Label: "คำอธิบาย", Type: ColText},
			{Key: "debit", Label: "เดบิต", Type: ColAmount, Group: "ยอดประจำงวด"}},
		Rows: []map[string]string{
			{KindKey: KindDetail, "docno": "JV001", "description": `ขาย "สด"`, "debit": "-100.00"},
			{KindKey: KindDetail, "docno": "=HYPERLINK(\"x\")", "description": "+cmd", "debit": "=1+1"},
			{KindKey: KindDetail, "docno": "@SUM(A1)", "description": "\tx", "debit": "0.00"},
			{KindKey: KindGroupTotal, "rowtype": "total", "docno": "", "description": "", "debit": "100.00"},
			{KindKey: KindGroupTotal, LabelKey: "รวมสินทรัพย์ 2 รายการ", "docno": "", "debit": ""},
		},
		Footer: []map[string]string{{"label": "รวมทั้งสิ้น 3 รายการ", "debit": "100.00"}},
	}
	var out bytes.Buffer
	if err := WriteCSV(&out, resp, "th"); err != nil {
		t.Fatal(err)
	}
	text := out.String()
	if !strings.HasPrefix(text, "\uFEFF") || !strings.HasSuffix(text, "\r\n") {
		t.Fatal("csv must start with a BOM and end lines with CRLF")
	}
	lines := strings.Split(strings.TrimSuffix(strings.TrimPrefix(text, "\uFEFF"), "\r\n"), "\r\n")
	want := []string{
		`"เลขที่เอกสาร","คำอธิบาย","ยอดประจำงวด เดบิต"`,
		`"JV001","ขาย ""สด""","-100.00"`,
		`"'=HYPERLINK(""x"")","'+cmd","'=1+1"`,
		"\"'@SUM(A1)\",\"'\tx\",\"0.00\"",
		`"","ยอดรวม","100.00"`,
		`"รวมสินทรัพย์ 2 รายการ","",""`,
		`"รวมทั้งสิ้น 3 รายการ","","100.00"`,
	}
	if len(lines) != len(want) {
		t.Fatalf("lines = %q", lines)
	}
	for i := range want {
		if lines[i] != want[i] {
			t.Errorf("line %d = %s, want %s", i, lines[i], want[i])
		}
	}
}

func TestContentDispositionEncodesThaiName(t *testing.T) {
	got := ContentDisposition("report-trial-balance.csv", "งบทดลอง; x.csv")
	if !strings.HasPrefix(got, `attachment; filename="report-trial-balance.csv"; filename*=UTF-8''%E0%B8%87`) || strings.Contains(got, "; x") {
		t.Fatalf("header = %s", got)
	}
}
