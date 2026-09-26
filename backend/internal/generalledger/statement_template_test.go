package generalledger

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"
)

func TestStatementTemplateModel(t *testing.T) {
	showNotes := true
	tmpl := Master{
		Kind:          "statement-templates",
		Code:          "BS-DBD-01",
		Name:          "งบแสดงฐานะการเงิน (แบบ DBD)",
		StatementType: "balance_sheet",
		IsActive:      true,
		GlobalStyle: &StatementGlobalStyle{
			FontFamily:     "sarabun",
			FontSize:       "15px",
			Scale:          2,
			Compact:        false,
			ShowNoteColumn: &showNotes,
			ComparisonType: "none",
		},
		Rows: []StatementRow{
			{
				ID:      "row-1",
				RowNo:   10,
				RowType: "header",
				Title:   "สินทรัพย์",
				Style:   StatementStyle{FontWeight: "bold", Indent: 0},
			},
			{
				ID:            "row-2",
				RowNo:         20,
				RowType:       "account",
				Title:         "เงินสดและรายการเทียบเท่าเงินสด",
				NoteNo:        "3",
				AccountCodes:  []string{"1111-01", "1111-02"},
				NormalBalance: "debit",
				Style:         StatementStyle{Indent: 1},
			},
			{
				ID:      "row-3",
				RowNo:   30,
				RowType: "subtotal",
				Title:   "รวมสินทรัพย์",
				Formula: "R20",
				Style:   StatementStyle{FontWeight: "bold", Indent: 0, Underline: "double"},
			},
		},
	}

	if tmpl.CollectionName() != "gl_statement_templates" {
		t.Fatalf("CollectionName() = %q, want %q", tmpl.CollectionName(), "gl_statement_templates")
	}

	data, err := json.Marshal(tmpl)
	if err != nil {
		t.Fatalf("Marshal() err = %v", err)
	}

	var decoded Master
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal() err = %v", err)
	}

	if decoded.Code != "BS-DBD-01" {
		t.Fatalf("Code = %q, want BS-DBD-01", decoded.Code)
	}
	if decoded.StatementType != "balance_sheet" {
		t.Fatalf("StatementType = %q, want balance_sheet", decoded.StatementType)
	}
	if decoded.GlobalStyle == nil || decoded.GlobalStyle.FontFamily != "sarabun" {
		t.Fatalf("GlobalStyle.FontFamily = %v, want sarabun", decoded.GlobalStyle)
	}
	if len(decoded.Rows) != 3 {
		t.Fatalf("len(Rows) = %d, want 3", len(decoded.Rows))
	}
	if decoded.Rows[2].Style.Underline != "double" {
		t.Fatalf("Rows[2].Style.Underline = %q, want double", decoded.Rows[2].Style.Underline)
	}
}

// ไม่ติ๊ก "แสดงคอลัมน์หมายเหตุประกอบงบ" ต้องบันทึก false ได้จริง; รูปแบบเก่าที่ไม่เคยระบุยังเป็น nil (จอแสดงคอลัมน์ตามค่าเริ่ม)
func TestStatementGlobalStyleShowNoteColumnRoundTrip(t *testing.T) {
	for _, c := range []struct {
		payload string
		want    string
	}{
		{`{"shownotecolumn":false}`, "false"},
		{`{"shownotecolumn":true}`, "true"},
		{`{"fontsize":"15px"}`, "nil"},
	} {
		var style StatementGlobalStyle
		if err := json.Unmarshal([]byte(c.payload), &style); err != nil {
			t.Fatal(err)
		}
		got := "nil"
		if style.ShowNoteColumn != nil {
			got = strconv.FormatBool(*style.ShowNoteColumn)
		}
		data, err := json.Marshal(style)
		if err != nil {
			t.Fatal(err)
		}
		if got != c.want || c.want != "nil" && !strings.Contains(string(data), `"shownotecolumn":`+c.want) || c.want == "nil" && strings.Contains(string(data), "shownotecolumn") {
			t.Fatalf("%s: decoded %s, re-encoded %s", c.payload, got, data)
		}
	}
}
