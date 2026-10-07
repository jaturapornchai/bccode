//go:build integration

package generalledger

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

// กลุ่มผังบัญชี (Champ 5497) ใช้ในแถวงบ: สมาชิกเป็นรหัสเดียวหรือช่วงตามลำดับ byte (COLLATE "C"), บัญชีอยู่ได้หลายกลุ่ม (ซ้อนทับไม่นับซ้ำ),
// ตัวตรวจ "บัญชีไม่อยู่ในงบ" เห็นบัญชีของกลุ่ม, และกลุ่มที่แม่แบบอ้างอยู่ลบ/เปลี่ยนรหัสไม่ได้ — บน PostgreSQL จริง
func TestStatementAccountGroupIntegration(t *testing.T) {
	f := newPGIntegrityFixture(t)
	chart := []struct {
		code    string
		posting bool
	}{{"1100", false}, {"1110", true}, {"1120", true}, {"1130", true}, {"1140", true}, {"1210", true}, {"11-25", true}, {"1300", true}}
	made := map[string]Account{}
	for _, row := range chart {
		made[row.code] = f.suggestAccountFixture(row.code, "asset", "", "บัญชี "+row.code, row.posting, nil)
	}
	post := func(doc, debitAccount, amount string) {
		j := Journal{DocNo: doc, Date: "2026-03-10", BookCode: "JV", FiscalYear: "2026", Description: "รายการทดสอบกลุ่มผังบัญชี", Kind: "manual", BranchCode: "B1",
			Lines: []Line{{AccountCode: debitAccount, Debit: Amount(amount), Credit: Amount("0")}, {AccountCode: "3000", Debit: Amount("0"), Credit: Amount(amount)}}}
		f.post(f.journal(f.run(Command{Resource: "journals", Action: "create", Journal: &j}).ID))
	}
	post("JV-G1", "1110", "100")
	post("JV-G2", "1120", "20.50")
	post("JV-G3", "1130", "3.25")
	post("JV-G4", "1140", "7")
	post("JV-G5", "1210", "40")
	post("JV-G6", "11-25", "9")

	group := func(code string, entries ...AccountEntry) Result {
		return f.run(Command{Resource: "account-groups", Action: "create", Master: &Master{Code: code, Name: "กลุ่ม " + code, IsActive: true, AccountEntries: entries}})
	}
	// ช่วง 1110–1130 + รหัสเดียวที่ซ้อนอยู่ในช่วง และรายการซ้ำ/ช่วงปลายเท่ากัน ถูกยุบเป็นรายการเดียว
	overlap := group("G-OVERLAP", AccountEntry{From: "1110", To: "1130"}, AccountEntry{From: "1120"}, AccountEntry{From: "1120"}, AccountEntry{From: "1210", To: "1210"})
	var stored Master
	raw, err := f.store.Get(f.ctx, f.scope, "account-groups", overlap.ID)
	if err != nil || json.Unmarshal(raw, &stored) != nil {
		t.Fatalf("read group: %v", err)
	}
	if want := []AccountEntry{{From: "1110", To: "1130"}, {From: "1120"}, {From: "1210"}}; !reflect.DeepEqual(stored.AccountEntries, want) {
		t.Fatalf("stored entries = %+v, want %+v", stored.AccountEntries, want)
	}
	rangeOnly := group("G-RANGE", AccountEntry{From: "1110", To: "1130"})
	group("G-SINGLES", AccountEntry{From: "1210"}, AccountEntry{From: "1140"})
	group("G-EMPTY")

	// ตรวจตอนบันทึกกลุ่ม: ช่วงกลับด้าน, รหัสเดียวเป็นบัญชีหัวข้อ/ไม่มีในผัง, รหัสมีช่องว่าง
	user := f.failCode(Command{Resource: "account-groups", Action: "create", Master: &Master{Code: "G-BAD", Name: "ช่วงกลับด้าน", IsActive: true, AccountEntries: []AccountEntry{{From: "1130", To: "1110"}}}}, "account_group_entry_reversed", "accountentries[0].to")
	if !reflect.DeepEqual(user.Args, []string{"1130", "1110"}) {
		t.Fatalf("reversed args = %+v", user.Args)
	}
	user = f.failCode(Command{Resource: "account-groups", Action: "create", Master: &Master{Code: "G-BAD", Name: "บัญชีหัวข้อ", IsActive: true, AccountEntries: []AccountEntry{{From: "1100"}, {From: "NOPE"}, {From: "1110"}}}}, "account_group_accounts_invalid", "accountentries")
	if !reflect.DeepEqual(user.Args, []string{"1100, NOPE"}) {
		t.Fatalf("invalid single codes = %+v", user.Args)
	}
	f.failCode(Command{Resource: "account-groups", Action: "create", Master: &Master{Code: "G-BAD", Name: "รหัสมีช่องว่าง", IsActive: true, AccountEntries: []AccountEntry{{From: "11 10", To: "1130"}}}}, "code_has_space", "accountentries[0].from")
	// รายการบัญชีมีเฉพาะกลุ่มผังบัญชี — ข้อมูลหลักชนิดอื่นไม่เก็บ
	other := f.run(Command{Resource: "product-account-groups", Action: "create", Master: &Master{Code: "P-1", Name: "กลุ่มสินค้า", IsActive: true, AccountEntries: []AccountEntry{{From: "1110"}}}})
	raw, err = f.store.Get(f.ctx, f.scope, "product-account-groups", other.ID)
	if err != nil || strings.Contains(string(raw), "accountentries") {
		t.Fatalf("product group kept account entries: %v %s", err, raw)
	}

	// แม่แบบงบ: แถวที่อ้างกลุ่มต้องมีกลุ่มอยู่จริง และเลือกบัญชีเองซ้อนไม่ได้
	f.failCode(Command{Resource: "statement-templates", Action: "create", Master: suggestBalanceSheet("FS-BAD",
		StatementRow{ID: "r1", RowNo: 10, RowType: "account", Title: "กลุ่มที่ไม่มี", AccountGroup: "NOPE"})}, "statement_template_row_group_missing", "rows[0].accountgroup")
	user = f.failCode(Command{Resource: "statement-templates", Action: "create", Master: suggestBalanceSheet("FS-BAD",
		StatementRow{ID: "r1", RowNo: 10, RowType: "account", Title: "เลือกซ้อน", AccountGroup: "G-RANGE", AccountCodes: []string{"1140"}})}, "statement_template_row_group_and_accounts", "rows[0].accountgroup")
	if !reflect.DeepEqual(user.Args, []string{"10", "เลือกซ้อน"}) {
		t.Fatalf("mixed row args = %+v", user.Args)
	}

	template := suggestBalanceSheet("FS-GRP",
		StatementRow{ID: "r1", RowNo: 10, RowType: "account", Title: "ช่วงรหัส", AccountGroup: "G-RANGE"},
		StatementRow{ID: "r2", RowNo: 20, RowType: "account", Title: "ซ้อนทับ", AccountGroup: "G-OVERLAP"},
		StatementRow{ID: "r3", RowNo: 30, RowType: "account", Title: "รหัสเดียวหลายตัว", AccountGroup: "G-SINGLES"},
		StatementRow{ID: "r4", RowNo: 40, RowType: "account", Title: "กลุ่มว่าง", AccountGroup: "G-EMPTY"},
		StatementRow{ID: "r5", RowNo: 50, RowType: "account", Title: "เลือกบัญชีเอง", AccountCodes: []string{"1300"}},
		StatementRow{ID: "r6", RowNo: 60, RowType: "subtotal", Title: "รวม", Formula: "R10 + R30"},
		// แถวชนิดอื่นที่ค้างรหัสกลุ่มไว้ ไม่ถูกอ่านและไม่ถูกตรวจ
		StatementRow{ID: "r7", RowNo: 70, RowType: "header", Title: "หัวข้อ", AccountGroup: "GONE"})
	tpl := f.run(Command{Resource: "statement-templates", Action: "create", Master: template})
	report, err := f.store.Report(f.ctx, f.scope, "statement", ReportQuery{FiscalYear: "2026", Template: "FS-GRP"})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"10": "123.75", "20": "163.75", "30": "47.00", "40": "0.00", "50": "0.00", "60": "170.75"}
	for _, row := range report.Rows {
		if w, ok := want[row["rowno"]]; ok && row["amount"] != w {
			t.Errorf("row %s = %s, want %s", row["rowno"], row["amount"], w)
		}
	}
	// ช่วงเทียบแบบ byte ไม่ขึ้นกับ locale ของฐานข้อมูล: "11-25" ขึ้นต้นด้วย 11- ซึ่งน้อยกว่า 1110 จึงอยู่นอกช่วง — ตัวตรวจบัญชีหลุดงบต้องเห็นมัน
	unassigned := map[string]bool{}
	for _, item := range report.Unassigned {
		unassigned[item.AccountCode] = true
	}
	for _, code := range []string{"1110", "1120", "1130", "1140", "1210"} {
		if unassigned[code] {
			t.Errorf("account %s of a group row reported as unassigned: %+v", code, report.Unassigned)
		}
	}
	if !unassigned["11-25"] {
		t.Errorf("11-25 must be outside the 1110–1130 range and unassigned: %+v", report.Unassigned)
	}
	empty := strings.Join(report.Warnings, "\n")
	if !strings.Contains(empty, "G-EMPTY") || strings.Contains(empty, "GONE") || strings.Contains(empty, "G-RANGE") {
		t.Errorf("warnings = %q, want only the empty group G-EMPTY", empty)
	}

	// แก้รายการของกลุ่มแล้วงบเปลี่ยนตามทันที (ไม่ต้องแก้แม่แบบ)
	edited := Master{Code: "G-RANGE", Name: "กลุ่ม G-RANGE", IsActive: true, AccountEntries: []AccountEntry{{From: "1110", To: "1140"}}}
	f.run(Command{Resource: "account-groups", Action: "update", ID: rangeOnly.ID, Version: rangeOnly.Version, Master: &edited})
	report, err = f.store.Report(f.ctx, f.scope, "statement", ReportQuery{FiscalYear: "2026", Template: "FS-GRP"})
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range report.Rows {
		if row["rowno"] == "10" && row["amount"] != "130.75" {
			t.Errorf("edited range row 10 = %s, want 130.75", row["amount"])
		}
	}

	// กลุ่มที่แม่แบบอ้างอยู่ ลบ/เปลี่ยนรหัสไม่ได้ แต่แก้ชื่อได้
	var current Master
	raw, err = f.store.Get(f.ctx, f.scope, "account-groups", rangeOnly.ID)
	if err != nil || json.Unmarshal(raw, &current) != nil {
		t.Fatalf("read group: %v", err)
	}
	recoded := current
	recoded.Code = "G-RANGE2"
	if user = f.failCode(Command{Resource: "account-groups", Action: "update", ID: current.ID, Version: current.Version, Master: &recoded}, "account_group_in_use_code", "code"); user.HTTPStatus() != 409 || !reflect.DeepEqual(user.Args, []string{"G-RANGE"}) {
		t.Fatalf("recode status = %d args = %+v", user.HTTPStatus(), user.Args)
	}
	if user = f.failCode(Command{Resource: "account-groups", Action: "delete", ID: current.ID, Version: current.Version}, "account_group_in_use_delete", "code"); user.HTTPStatus() != 409 {
		t.Fatalf("delete status = %d", user.HTTPStatus())
	}
	renamed := current
	renamed.Name = "กลุ่มเงินสดและเงินฝาก"
	renameResult := f.run(Command{Resource: "account-groups", Action: "update", ID: current.ID, Version: current.Version, Master: &renamed})

	// รหัสเดี่ยวในกลุ่มกันการลบบัญชี (บัญชีที่ยังไม่มีรายการบัญชีก็ลบไม่ได้ ตราบที่กลุ่มยังมีรหัสนี้)
	loose := made["1300"]
	holder := group("G-HOLD", AccountEntry{From: "1300"})
	f.failCode(Command{Resource: "accounts", Action: "delete", ID: loose.ID, Version: loose.Version}, CodeReferenced, "")
	f.run(Command{Resource: "statement-templates", Action: "delete", ID: tpl.ID, Version: tpl.Version})
	f.failCode(Command{Resource: "accounts", Action: "delete", ID: loose.ID, Version: loose.Version}, CodeReferenced, "")
	f.run(Command{Resource: "account-groups", Action: "delete", ID: holder.ID, Version: holder.Version})
	f.run(Command{Resource: "accounts", Action: "delete", ID: loose.ID, Version: loose.Version})
	// ไม่มีแม่แบบอ้างแล้ว กลุ่มลบได้
	f.run(Command{Resource: "account-groups", Action: "delete", ID: current.ID, Version: renameResult.Version})
}

// รหัสกลุ่มที่แม่แบบอื่นอ้างอยู่แต่แถวนั้นไม่ใช่ชนิด account ไม่กันการลบ (ไม่ถูกอ่านตอนคำนวณ)
func TestStatementAccountGroupHeaderRowDoesNotBlockDeleteIntegration(t *testing.T) {
	f := newPGIntegrityFixture(t)
	f.suggestAccountFixture("1110", "asset", "", "เงินสด", true, nil)
	g := f.run(Command{Resource: "account-groups", Action: "create", Master: &Master{Code: "G-1", Name: "กลุ่ม", IsActive: true, AccountEntries: []AccountEntry{{From: "1110"}}}})
	f.run(Command{Resource: "statement-templates", Action: "create", Master: suggestBalanceSheet("FS-H",
		StatementRow{ID: "h", RowNo: 10, RowType: "header", Title: "หัวข้อ", AccountGroup: "G-1"})})
	f.run(Command{Resource: "account-groups", Action: "delete", ID: g.ID, Version: g.Version})
}
