package generalledger

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// ลำดับชุดงบตามแบบ 2 ประกาศกรมพัฒนาธุรกิจการค้า พ.ศ. 2566 (หน้า 2-1 / 2-4 / 2-20 / 2-23) — frontend ตรวจบรรทัดเดียวกันซ้ำ
func TestStatementSetOrderFollowsDBDForm2(t *testing.T) {
	if want := []string{"balance_sheet", "pnl", "equity", "cash_flow"}; !reflect.DeepEqual(statementSetOrder, want) {
		t.Fatalf("order = %v, want %v", statementSetOrder, want)
	}
	for _, other := range []string{"production_cost", "custom", ""} {
		if statementSetRank(other) != len(statementSetOrder) {
			t.Fatalf("%q must sort after the DBD statements", other)
		}
	}
}

func setUserError(t *testing.T, err error, code string) *UserError {
	t.Helper()
	user, ok := AsUserError(err)
	if !ok || user.Code != code || user.Status != 400 {
		t.Fatalf("want 400 %s, got %v", code, err)
	}
	return user
}

func TestParseStatementSetTemplates(t *testing.T) {
	codes, err := ParseStatementSetTemplates(" PL-1 ,BS-1,,PL-1, บ.ฐานะ ")
	if err != nil || !reflect.DeepEqual(codes, []string{"PL-1", "BS-1", "บ.ฐานะ"}) {
		t.Fatalf("codes = %v, %v", codes, err)
	}
	// ค่าว่าง = ไม่พิมพ์งบ (ไม่ใช่ชุดเริ่มต้น) — ต้องได้ slice ว่างที่ไม่ใช่ nil
	if codes, err = ParseStatementSetTemplates(""); err != nil || codes == nil || len(codes) != 0 {
		t.Fatalf("empty = %#v, %v", codes, err)
	}
	many := make([]string, statementSetMaxTemplates)
	for i := range many {
		many[i] = fmt.Sprintf("T%02d", i)
	}
	if codes, err = ParseStatementSetTemplates(strings.Join(many, ",") + ",T00"); err != nil || len(codes) != statementSetMaxTemplates {
		t.Fatalf("limit (duplicates not counted) = %d, %v", len(codes), err)
	}
	_, err = ParseStatementSetTemplates(strings.Join(many, ",") + ",T99")
	user := setUserError(t, err, "statement_set_too_many")
	if user.Field != "templates" || !reflect.DeepEqual(user.Args, []string{"20", "21"}) || !strings.Contains(user.Message, "20") {
		t.Fatalf("too many = %+v", user)
	}
}

func TestParseStatementSetNotes(t *testing.T) {
	for raw, want := range map[string]bool{"": true, "true": true, "false": false} {
		if got, err := ParseStatementSetNotes(raw); err != nil || got != want {
			t.Fatalf("notes %q = %v, %v", raw, got, err)
		}
	}
	for _, raw := range []string{"1", "TRUE", "no"} {
		_, err := ParseStatementSetNotes(raw)
		setUserError(t, err, "statement_set_notes_invalid")
	}
}

func statementSetCodes(templates []Master) []string {
	codes := []string{}
	for _, template := range templates {
		codes = append(codes, template.Code)
	}
	return codes
}

func TestSelectStatementSetTemplates(t *testing.T) {
	all := []Master{
		{Code: "ZZ", StatementType: "custom", IsActive: true},
		{Code: "CF-1", StatementType: "cash_flow", IsActive: true},
		{Code: "PL-2", StatementType: "pnl", IsActive: true},
		{Code: "PL-1", StatementType: "pnl", IsActive: true},
		{Code: "PL-0", StatementType: "pnl", IsActive: false},
		{Code: "PC-1", StatementType: "production_cost", IsActive: true},
		{Code: "BS-1", StatementType: "balance_sheet", IsActive: true},
		{Code: "AA", StatementType: "custom", IsActive: true},
	}
	// ชุดเริ่มต้น: รูปแบบแรกตามรหัสที่เปิดใช้ของงบในแบบ 2 เท่านั้น (ไม่มีงบการเปลี่ยนแปลงส่วนของผู้ถือหุ้น = ข้าม)
	chosen, err := selectStatementSetTemplates(all, nil)
	if err != nil || !reflect.DeepEqual(statementSetCodes(chosen), []string{"BS-1", "PL-1", "CF-1"}) {
		t.Fatalf("default = %v, %v", statementSetCodes(chosen), err)
	}
	// ลำดับที่ส่งมาไม่มีผล: แบบ 2 ก่อน แล้วชนิดอื่นตามรหัส
	chosen, err = selectStatementSetTemplates(all, []string{"ZZ", "CF-1", "PC-1", "AA", "PL-2", "BS-1"})
	if err != nil || !reflect.DeepEqual(statementSetCodes(chosen), []string{"BS-1", "PL-2", "CF-1", "AA", "PC-1", "ZZ"}) {
		t.Fatalf("given = %v, %v", statementSetCodes(chosen), err)
	}
	if chosen, err = selectStatementSetTemplates(all, []string{}); err != nil || len(chosen) != 0 {
		t.Fatalf("none = %v, %v", statementSetCodes(chosen), err)
	}
	_, err = selectStatementSetTemplates(all, []string{"BS-1", "PL-0"})
	if user := setUserError(t, err, "statement_set_template_inactive"); !reflect.DeepEqual(user.Args, []string{"PL-0"}) || !strings.Contains(user.Message, "PL-0") {
		t.Fatalf("inactive = %+v", user)
	}
	_, err = selectStatementSetTemplates(all, []string{"NOPE"})
	if user := setUserError(t, err, "statement_set_template_not_found"); !reflect.DeepEqual(user.Args, []string{"NOPE"}) || !strings.Contains(user.Message, "NOPE") {
		t.Fatalf("not found = %+v", user)
	}
	if _, err = selectStatementSetTemplates(nil, nil); err != nil {
		t.Fatalf("no templates at all: %v", err)
	}
}

// ชื่อ/ทศนิยม/คอลัมน์หมายเหตุที่ส่งให้จอพิมพ์ต้องเป็นกติกาเดียวกับที่งบใช้คำนวณ
func TestStatementSetSectionDisplayRules(t *testing.T) {
	hidden := false
	if statementScale(Master{}) != 2 || statementScale(Master{GlobalStyle: &StatementGlobalStyle{Scale: 0}}) != 2 || statementScale(Master{GlobalStyle: &StatementGlobalStyle{Scale: 4}}) != 4 {
		t.Fatal("scale default")
	}
	if !statementShowsNoteColumn(Master{}) || !statementShowsNoteColumn(Master{GlobalStyle: &StatementGlobalStyle{}}) || statementShowsNoteColumn(Master{GlobalStyle: &StatementGlobalStyle{ShowNoteColumn: &hidden}}) {
		t.Fatal("note column default")
	}
}
