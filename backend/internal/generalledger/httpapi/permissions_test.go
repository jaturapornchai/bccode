package httpapi

import "testing"

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
	if reportScreens["gljournal"] != "gl-daily-report" {
		t.Fatalf("expected reportScreens[gljournal] to be gl-daily-report, got %s", reportScreens["gljournal"])
	}
	if reportScreens["budgetcomparison"] != "budget-comparison-report" {
		t.Fatalf("expected reportScreens[budgetcomparison] to be budget-comparison-report, got %s", reportScreens["budgetcomparison"])
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
	for _, resource := range []string{"accounts", "fiscal-years", "journal-books"} {
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
	}
	if anyLedger(map[string]bool{"jv-journal": true}) || anyLedger(map[string]bool{"gl-unpost": true}) {
		t.Fatal("legacy screen ids must not grant ledger lookups")
	}
}
