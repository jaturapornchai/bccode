package httpapi

import (
	"strings"
	"testing"
)

func TestLedgerPermissionsSeparateEntryFromActions(t *testing.T) {
	p := map[string]bool{"gl-post": true, "financial-close": true}
	if !allowed(p, "gl-post", "") || allowed(p, "gl-post", "update") {
		t.Fatal("screen entry granted mutation")
	}
	p["gl-post:update"] = true
	if !allowed(p, "gl-post", "update") || allowed(p, "gl-journals", "create") {
		t.Fatal("action scope bypass")
	}
	if !canReadReport(p, "workingpaper") || canReadReport(p, "pnl") {
		t.Fatal("process preview grant escaped report scope")
	}
	if !anyLedger(p) || anyLedger(map[string]bool{"sales-order": true}) {
		t.Fatal("lookup scope bypass")
	}
	// สิทธิ์สมุดรายวันต้องไม่ผูกกับรหัสสมุด (สมุดเป็น master ที่ผู้ใช้เพิ่ม/แก้เองได้)
	if journalScreen("general") != "gl-journals" || journalScreen("") != "gl-journals" || journalScreen("opening") != "gl-opening-balance" {
		t.Fatal("journal permission routing")
	}
	if resourceScreens["journal-books"] != "gl-journal-books" {
		t.Fatalf("expected resourceScreens[journal-books] to be gl-journal-books, got %s", resourceScreens["journal-books"])
	}
	// แผนกและโครงการอยู่จอเดียว (กำหนดรหัสต่างๆ): สิทธิ์เดียวกัน อ่านได้ไม่ได้แปลว่าแก้ได้
	if resourceScreens["departments"] != "gl-dimensions" || resourceScreens["projects"] != "gl-dimensions" || allowed(map[string]bool{"gl-dimensions": true}, resourceScreens["departments"], "delete") || !allowed(map[string]bool{"gl-dimensions": true, "gl-dimensions:delete": true}, resourceScreens["projects"], "delete") {
		t.Fatalf("dimension screens = %q / %q", resourceScreens["departments"], resourceScreens["projects"])
	}
	// รูปแบบรายการรายวัน (journal-templates) ใช้สิทธิ์จอบันทึกรายวันเดียวกัน — ไม่มีเมนูใหม่
	if resourceScreens["journal-templates"] != "gl-journals" || !allowed(map[string]bool{"gl-journals": true, "gl-journals:create": true}, resourceScreens["journal-templates"], "create") || allowed(map[string]bool{"gl-journals": true}, resourceScreens["journal-templates"], "delete") {
		t.Fatalf("journal templates screen = %q", resourceScreens["journal-templates"])
	}
	// หมายเหตุประกอบงบการเงินอยู่ในจอออกแบบงบการเงิน (ไม่มีเมนูใหม่)
	if resourceScreens["statement-notes"] != "financial-statement-designer" || resourceScreens["statement-templates"] != "financial-statement-designer" {
		t.Fatalf("statement notes screen = %q", resourceScreens["statement-notes"])
	}
	if reportScreens["gljournal"] != "gl-daily-report" {
		t.Fatalf("expected reportScreens[gljournal] to be gl-daily-report, got %s", reportScreens["gljournal"])
	}
	if reportScreens["budgetcomparison"] != "budget-comparison-report" {
		t.Fatalf("expected reportScreens[budgetcomparison] to be budget-comparison-report, got %s", reportScreens["budgetcomparison"])
	}
	// Champ 5525 budget detail = print button of the budget screen: same permission as that screen
	if reportScreens["budgetdetail"] != "gl-budget" || !canReadReport(map[string]bool{"gl-budget": true}, "budgetdetail") || canReadReport(map[string]bool{"budget-comparison-report": true}, "budgetdetail") {
		t.Fatalf("budgetdetail screen = %q", reportScreens["budgetdetail"])
	}
}

func TestJournalSupportReadPermissions(t *testing.T) {
	for _, screen := range []string{"gl-journals", "gl-opening-balance", "gl-post", "*"} {
		p := map[string]bool{screen: true}
		if !canReadJournalSupport(p) {
			t.Fatalf("journal screen %s cannot read evidence", screen)
		}
		for _, report := range []string{"ar-outstanding", "ap-outstanding", "bank-unmatched"} {
			if !canReadReport(p, report) {
				t.Fatalf("journal screen %s cannot read %s", screen, report)
			}
		}
	}
	// สิทธิ์เฉพาะจอรายงานยอดคงค้าง เปิดได้เฉพาะรายงานของจอนั้น ไม่ลามไปอ่านหลักฐานรายวันหรือรายงานอื่น
	for report, screen := range map[string]string{"ar-outstanding": "gl-ar-outstanding", "ap-outstanding": "gl-ap-outstanding", "bank-unmatched": "gl-bank-unmatched"} {
		p := map[string]bool{screen: true}
		if !canReadReport(p, report) || canReadJournalSupport(p) {
			t.Fatalf("screen %s: report %s access wrong", screen, report)
		}
		for _, other := range []string{"ar-outstanding", "ap-outstanding", "bank-unmatched", "ledger"} {
			if other != report && canReadReport(p, other) {
				t.Fatalf("screen %s must not read %s", screen, other)
			}
		}
	}
	// สิทธิ์ตามรหัสสมุดแบบเก่า (jv-journal ฯลฯ) และ gl-unpost ไม่มีในเมนูแล้ว ต้องไม่เปิดสิทธิ์อ่าน
	for _, p := range []map[string]bool{nil, {"sales-order": true}, {"gl-journals": false}, {"gl-journals:update": true}, {"jv-journal": true}, {"sv-journal": true}, {"gl-unpost": true}} {
		if canReadJournalSupport(p) {
			t.Fatal("unrelated, legacy or action-only permission granted evidence read")
		}
	}
}

func TestJournalListPermissionsDoNotDependOnBookCode(t *testing.T) {
	journals := map[string]bool{"gl-journals": true}
	opening := map[string]bool{"gl-opening-balance": true}
	posting := map[string]bool{"gl-post": true}
	cases := []struct {
		name         string
		p            map[string]bool
		kind, status string
		want         bool
	}{
		{"journals any kind", journals, "general", "", true},
		{"journals opening", journals, "opening", "", true},
		{"opening screen opening kind", opening, "opening", "", true},
		{"opening screen cannot list general", opening, "general", "", false},
		{"opening screen cannot list all", opening, "", "", false},
		{"post screen drafts", posting, "", "draft", true},
		{"post screen posted", posting, "general", "posted", true},
		{"post screen all statuses", posting, "", "", false},
		{"post screen cancelled", posting, "", "cancelled", false},
		{"legacy book screen", map[string]bool{"jv-journal": true}, "general", "", false},
		{"action only", map[string]bool{"gl-journals:create": true}, "", "", false},
	}
	for _, c := range cases {
		if got := canListJournals(c.p, c.kind, c.status); got != c.want {
			t.Fatalf("%s: canListJournals=%v want %v", c.name, got, c.want)
		}
	}
	// ผู้ใช้สมุดรายวันทุกคนต้องอ่านรายการสมุดได้ (ไว้เลือกสมุดตอนคีย์) แต่แก้สมุดต้องมีสิทธิ์จอกำหนดสมุดรายวัน
	for _, resource := range []string{"accounts", "fiscal-years", "journal-books", "departments", "projects"} {
		if !ledgerLookup(resource) {
			t.Fatalf("%s must be a ledger lookup", resource)
		}
	}
	if ledgerLookup("journals") || ledgerLookup("mappings") {
		t.Fatal("non-lookup resource treated as lookup")
	}
	for _, p := range []map[string]bool{journals, opening, posting} {
		if !anyLedger(p) {
			t.Fatalf("journal user %v cannot read lookups", p)
		}
		if allowed(p, resourceScreens["journal-books"], "update") {
			t.Fatalf("journal user %v can edit journal books", p)
		}
		// the department / project pickers read the registers, but only the register screen may change them
		if allowed(p, resourceScreens["departments"], "create") || allowed(p, resourceScreens["projects"], "update") {
			t.Fatalf("journal user %v can edit the department or project register", p)
		}
	}
	if anyLedger(map[string]bool{"jv-journal": true}) || anyLedger(map[string]bool{"gl-unpost": true}) {
		t.Fatal("legacy screen ids must not grant ledger lookups")
	}
}

// พิมพ์/ส่งออกรายงานบัญชี (purpose=print|export) ต้องมีสิทธิ์เฉพาะแบบ Champ (BaseReportView.cpp:736-745,2406-2415):
// อ่านรายงานบนจอได้ไม่ได้แปลว่าพิมพ์หรือส่งออกได้ — ปฏิเสธที่ backend พร้อมข้อความบอกวิธีขอสิทธิ์
func TestReportPrintExportNeedSpecialRights(t *testing.T) {
	reader := map[string]bool{"profit-loss": true}
	if status, _ := reportPurposeRefusal(reader, "", "th"); status != 0 {
		t.Fatalf("screen read refused: %d", status)
	}
	status, body := reportPurposeRefusal(reader, "print", "th")
	if status != 403 || body.Code != "REPORT_PRINT_FORBIDDEN" || !strings.Contains(body.Message, "พิมพ์รายงานได้") {
		t.Fatalf("print without right = %d %+v", status, body)
	}
	status, body = reportPurposeRefusal(map[string]bool{"profit-loss": true, "report:print": true}, "export", "en")
	if status != 403 || body.Code != "REPORT_EXPORT_FORBIDDEN" || !strings.Contains(body.Message, "Can export report data") || !strings.Contains(body.MessageTH, "ส่งออกข้อมูลของรายงานได้") {
		t.Fatalf("export with print right only = %d %+v", status, body)
	}
	for _, p := range []map[string]bool{{"report:print": true, "report:export": true}, {"*": true}} {
		for _, purpose := range []string{"print", "export"} {
			if status, body := reportPurposeRefusal(p, purpose, "th"); status != 0 {
				t.Fatalf("%v %s refused: %+v", p, purpose, body)
			}
		}
	}
	if status, _ := reportPurposeRefusal(map[string]bool{"*": true}, "download", "th"); status != 400 {
		t.Fatalf("unknown purpose = %d", status)
	}
}
