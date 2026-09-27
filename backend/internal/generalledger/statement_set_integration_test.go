//go:build integration

package generalledger

import (
	"reflect"
	"testing"
)

// ชุดงบการเงิน (reports/statement-set) บน PostgreSQL จริง: ชุดเริ่มต้นข้ามรูปแบบที่ปิดใช้/ลบแล้ว, ลำดับตามแบบ 2 ไม่ขึ้นกับลำดับที่ส่ง,
// ทุกงบเท่ากับ reports/statement ทีละงบทุกค่า (ตัวเลข คำเตือน ผลตรวจ), หมายเหตุของปีบัญชีมาด้วย, งบที่คำนวณไม่สำเร็จไม่ล้มทั้งชุด
// แม้ล้มระดับ SQL (SAVEPOINT) และรหัสที่ใช้ไม่ได้ถูกปฏิเสธทั้งคำขอ
func TestStatementSetIntegration(t *testing.T) {
	f := newPGIntegrityFixture(t)
	post := func(doc, date, debitAccount, creditAccount, amount string) {
		j := Journal{DocNo: doc, Date: date, BookCode: "JV", FiscalYear: "2026", Description: "รายการทดสอบชุดงบการเงิน", Kind: "manual", BranchCode: "B1",
			Lines: []Line{{AccountCode: debitAccount, Debit: Amount(amount), Credit: Amount("0")}, {AccountCode: creditAccount, Debit: Amount("0"), Credit: Amount(amount)}}}
		f.post(f.journal(f.run(Command{Resource: "journals", Action: "create", Journal: &j}).ID))
	}
	post("JV26-1", "2026-01-05", "1000", "3000", "500")
	post("JV26-2", "2026-02-10", "1000", "4000", "100")
	post("JV26-3", "2026-03-01", "5000", "1000", "30.25")

	hidden := false
	position := Master{Kind: "statement-templates", Code: "BS-A", Name: "งบแสดงฐานะการเงิน", StatementType: "balance_sheet", IsActive: true,
		GlobalStyle: &StatementGlobalStyle{Scale: 2, ComparisonType: "previous_year"},
		Rows: []StatementRow{
			{ID: "a", RowNo: 10, RowType: "account", Title: "เงินสดและรายการเทียบเท่าเงินสด", NoteNo: "5", AccountCodes: []string{"1000"}},
			{ID: "b", RowNo: 20, RowType: "account", Title: "ทุนที่ชำระแล้ว", AccountCodes: []string{"3000"}},
			{ID: "c", RowNo: 30, RowType: "account", Title: "กำไร (ขาดทุน) สุทธิประจำงวด", AccountCodes: []string{statementCurrentEarnings}},
			{ID: "d", RowNo: 40, RowType: "subtotal", Title: "รวมส่วนของผู้ถือหุ้น", Formula: "R20 + R30"},
		}}
	positionB := position
	positionB.Code, positionB.Name = "BS-B", "งบแสดงฐานะการเงิน (ย่อ)"
	positionB.Rows = []StatementRow{{ID: "a", RowNo: 10, RowType: "account", Title: "เงินสด", NoteNo: "9", AccountCodes: []string{"1000"}}}
	deleted := position
	deleted.Code = "BS-0"
	income := Master{Kind: "statement-templates", Code: "PL-1", Name: "งบกำไรขาดทุน", StatementType: "pnl", IsActive: true,
		GlobalStyle: &StatementGlobalStyle{Scale: 2, ComparisonType: "none", ShowNoteColumn: &hidden},
		Rows: []StatementRow{
			{ID: "a", RowNo: 10, RowType: "account", Title: "รายได้จากการขาย", AccountCodes: []string{"4000"}},
			{ID: "b", RowNo: 20, RowType: "account", Title: "ค่าใช้จ่ายในการขายและบริหาร", AccountCodes: []string{"5000"}},
			{ID: "c", RowNo: 30, RowType: "formula", Title: "กำไร (ขาดทุน) สุทธิ", Formula: "R10 - R20"},
		}}
	inactive := income
	inactive.Code, inactive.IsActive = "PL-0", false
	equity := Master{Kind: "statement-templates", Code: "EQ-1", Name: "งบการเปลี่ยนแปลงส่วนของผู้ถือหุ้น", StatementType: "equity", IsActive: true,
		GlobalStyle: &StatementGlobalStyle{Scale: 2, ComparisonType: "none"},
		Columns: []StatementColumn{
			{ID: "cap", Title: "ทุนที่ชำระแล้ว", AccountCodes: []string{"3000"}},
			{ID: "re", Title: "กำไร (ขาดทุน) สะสม", AccountCodes: []string{"3100", "3200", statementCurrentEarnings}},
		},
		Rows: []StatementRow{
			{ID: "a", RowNo: 10, RowType: "account", Title: "ยอดคงเหลือ ณ ต้นงวด", AmountBasis: "opening"},
			{ID: "b", RowNo: 20, RowType: "account", Title: "การเพิ่ม (ลด) หุ้นสามัญ", AccountCodes: []string{"3000"}},
			{ID: "c", RowNo: 30, RowType: "account", Title: "กำไร (ขาดทุน) สุทธิ", AccountCodes: []string{statementCurrentEarnings}},
			{ID: "d", RowNo: 40, RowType: "account", Title: "ยอดคงเหลือ ณ ปลายงวด", AmountBasis: "closing"},
		}}
	// งบการเปลี่ยนแปลงส่วนของผู้ถือหุ้นที่ยังไม่ผูกบัญชีกับคอลัมน์ = คำนวณไม่สำเร็จ (ข้อความเดียวกับ reports/statement)
	equityEmpty := equity
	equityEmpty.Code, equityEmpty.Columns = "EQ-EMPTY", []StatementColumn{{ID: "cap", Title: "ทุนที่ชำระแล้ว"}}
	cashFlow := Master{Kind: "statement-templates", Code: "CF-1", Name: "งบกระแสเงินสด", StatementType: "cash_flow", IsActive: true,
		GlobalStyle: &StatementGlobalStyle{Scale: 2, ComparisonType: "none"},
		Rows: []StatementRow{
			{ID: "a", RowNo: 10, RowType: "account", Title: "เงินสดต้นงวด", AccountCodes: []string{"1000"}, AmountBasis: "opening"},
			{ID: "b", RowNo: 20, RowType: "account", Title: "เงินสดเพิ่มขึ้น (ลดลง) สุทธิ", AccountCodes: []string{"1000"}},
			{ID: "c", RowNo: 30, RowType: "formula", Title: "เงินสดปลายงวด", Formula: "R10 + R20"},
		}}
	production := Master{Kind: "statement-templates", Code: "PC-1", Name: "งบต้นทุนผลิต", StatementType: "production_cost", IsActive: true,
		GlobalStyle: &StatementGlobalStyle{Scale: 4},
		Rows:        []StatementRow{{ID: "a", RowNo: 10, RowType: "account", Title: "ต้นทุนขาย", AccountCodes: []string{"5000"}}}}
	custom := Master{Kind: "statement-templates", Code: "ZZ", Name: "งบกำหนดเอง", StatementType: "custom", IsActive: true,
		Rows: []StatementRow{{ID: "a", RowNo: 10, RowType: "account", Title: "เงินสด", AccountCodes: []string{"1000"}}}}
	for _, m := range []Master{custom, production, cashFlow, equityEmpty, equity, inactive, income, positionB, position} {
		m := m
		f.run(Command{Resource: "statement-templates", Action: "create", Master: &m})
	}
	created := f.run(Command{Resource: "statement-templates", Action: "create", Master: &deleted})
	f.run(Command{Resource: "statement-templates", Action: "delete", ID: created.ID, Version: created.Version})
	notes := []StatementNote{{NoteNo: "1", Title: "ข้อมูลทั่วไป", Body: "บริษัทจดทะเบียนในประเทศไทย"}, {NoteNo: "5", Title: "เงินสดและรายการเทียบเท่าเงินสด", Body: "เงินสดในมือและเงินฝากธนาคาร"}}
	f.run(Command{Resource: "statement-notes", Action: "create", Master: &Master{Code: "2026", Name: "หมายเหตุประกอบงบการเงิน 2026", IsActive: true, Notes: notes}})

	projection := f.store.Projection()
	set := func(templates []string, withNotes bool) (StatementSet, error) {
		t.Helper()
		return projection.StatementSet(f.ctx, f.scope, StatementSetQuery{Report: ReportQuery{FiscalYear: "2026"}, Templates: templates, Notes: withNotes})
	}
	codes := func(s StatementSet) []string {
		out := []string{}
		for _, section := range s.Sections {
			out = append(out, section.Code)
		}
		return out
	}
	// ทุกงบในชุดต้องเท่ากับ reports/statement ทีละงบทุกค่า (หรือล้มด้วยข้อความเดียวกัน)
	sameAsSingle := func(section StatementSetSection) {
		t.Helper()
		single, singleErr := f.store.Report(f.ctx, f.scope, "statement", ReportQuery{FiscalYear: "2026", Template: section.Code})
		if singleErr != nil || section.Err != nil {
			if singleErr == nil || section.Err == nil || singleErr.Error() != section.Err.Error() || section.Report != nil {
				t.Fatalf("%s: set err %v / report %v, single err %v", section.Code, section.Err, section.Report != nil, singleErr)
			}
			return
		}
		if section.Report == nil || !reflect.DeepEqual(*section.Report, single) {
			t.Fatalf("%s differs from reports/statement:\nset    %+v\nsingle %+v", section.Code, section.Report, single)
		}
	}

	// ชุดเริ่มต้น: BS-0 ลบแล้ว / PL-0 ปิดใช้ ถูกข้าม; งบต้นทุนผลิตและงบกำหนดเองไม่อยู่ในชุดเริ่มต้น; หมายเหตุมาด้วย
	def, err := set(nil, true)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"BS-A", "PL-1", "EQ-1", "CF-1"}; !reflect.DeepEqual(codes(def), want) {
		t.Fatalf("default set = %v, want %v", codes(def), want)
	}
	if def.FiscalYear != "2026" || def.From != "2026-01-01" || def.To != "2026-12-31" || len(def.Notes) != 2 || def.Notes[1].NoteNo != "5" || def.Notes[1].Body != notes[1].Body {
		t.Fatalf("set header/notes = %s %s %s %+v", def.FiscalYear, def.From, def.To, def.Notes)
	}
	for _, section := range def.Sections {
		sameAsSingle(section)
	}
	pl := def.Sections[1]
	if pl.Name != "งบกำไรขาดทุน" || pl.StatementType != "pnl" || pl.ShowNoteColumn || pl.Scale != 2 || !def.Sections[0].ShowNoteColumn {
		t.Fatalf("section display = %+v", pl)
	}
	for _, row := range pl.Report.Rows {
		if row["rowno"] == "30" && row["amount"] != "69.75" {
			t.Fatalf("net profit = %v", row)
		}
	}

	// รหัสที่ส่งมาเรียงตามแบบ 2 เสมอ แล้วชนิดอื่นตามรหัส; EQ-EMPTY คำนวณไม่สำเร็จแต่งบอื่นยังได้ผล; notes=false ไม่อ่านหมายเหตุ
	given, err := set([]string{"ZZ", "EQ-EMPTY", "PC-1", "CF-1", "BS-B", "PL-1"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"BS-B", "PL-1", "EQ-EMPTY", "CF-1", "PC-1", "ZZ"}; !reflect.DeepEqual(codes(given), want) {
		t.Fatalf("given set = %v, want %v", codes(given), want)
	}
	if len(given.Notes) != 0 || given.Notes == nil {
		t.Fatalf("notes not requested = %#v", given.Notes)
	}
	if given.Sections[2].Err == nil || given.Sections[3].Report == nil || given.Sections[4].Scale != 4 {
		t.Fatalf("section error isolation = %+v", given.Sections)
	}
	for _, section := range given.Sections {
		sameAsSingle(section)
	}
	// คำเตือนของงบมาครบ: BS-B อ้างหมายเหตุ 9 ที่ปี 2026 ไม่มี
	if w := given.Sections[0].Report.Warnings; len(w) == 0 || w[0] != "หมายเหตุ 9 ที่อ้างในงบยังไม่มีในหมายเหตุประกอบงบการเงินปี 2026" {
		t.Fatalf("BS-B warnings = %v", w)
	}

	// หมายเหตุอย่างเดียว (templates ว่าง)
	onlyNotes, err := set([]string{}, true)
	if err != nil || len(onlyNotes.Sections) != 0 || len(onlyNotes.Notes) != 2 {
		t.Fatalf("notes only = %+v, %v", onlyNotes, err)
	}

	// รหัสที่ใช้ไม่ได้ = ทั้งคำขอไม่ผ่าน พร้อมรหัสที่ผิด
	for _, tc := range []struct {
		templates []string
		notes     bool
		code      string
	}{
		{[]string{"BS-A", "PL-0"}, true, "statement_set_template_inactive"},
		{[]string{"BS-0"}, true, "statement_set_template_not_found"},
		{[]string{"NOPE"}, true, "statement_set_template_not_found"},
		{[]string{}, false, "statement_set_empty"},
	} {
		_, err := set(tc.templates, tc.notes)
		user, ok := AsUserError(err)
		if !ok || user.Code != tc.code || user.Status != 400 {
			t.Fatalf("%v: want %s, got %v", tc.templates, tc.code, err)
		}
		if len(tc.templates) == 1 && (len(user.Args) != 1 || user.Args[0] != tc.templates[0]) {
			t.Fatalf("%v: args = %v", tc.templates, user.Args)
		}
	}

	// งบที่ล้มระดับ SQL (ปีบัญชีที่วันที่เสีย ทำให้หาปีก่อนเปรียบเทียบไม่ได้) ต้องไม่ทำให้ transaction ใช้ต่อไม่ได้:
	// งบถัดไปและหมายเหตุยังอ่านได้ใน snapshot เดียวกัน
	if _, err = f.db.ExecContext(f.ctx, `INSERT INTO gl_records(company,kind,id,code,version,payload) VALUES ($1,'fiscal-years','broken-year','BROKEN',1,'{"code":"BROKEN","startdate":"x","enddate":"not-a-date"}')`, f.scope.Company); err != nil {
		t.Fatal(err)
	}
	broken, err := set([]string{"PL-1", "BS-A"}, true)
	if err != nil {
		t.Fatalf("SQL failure of one statement failed the whole set: %v", err)
	}
	if broken.Sections[0].Code != "BS-A" || broken.Sections[0].Err == nil || broken.Sections[1].Report == nil || len(broken.Notes) != 2 {
		t.Fatalf("SQL failure not isolated: %+v / notes %d", broken.Sections, len(broken.Notes))
	}
	sameAsSingle(broken.Sections[1])
}
