package httpapi

import "testing"

func TestLedgerPermissionsSeparateEntryFromActions(t *testing.T) {
	p := map[string]bool{"gl-post": true, "financial-close": true}
	if !allowed(p, "gl-post", "") || allowed(p, "gl-post", "update") {
		t.Fatal("screen entry granted mutation")
	}
	p["gl-post:update"] = true
	if !allowed(p, "gl-post", "update") || allowed(p, "jv-journal", "create") {
		t.Fatal("action scope bypass")
	}
	if !canReadReport(p, "workingpaper") || canReadReport(p, "pnl") {
		t.Fatal("process preview grant escaped report scope")
	}
	if !anyLedger(p) || anyLedger(map[string]bool{"sales-order": true}) {
		t.Fatal("lookup scope bypass")
	}
	if journalScreen("", "", "draft") != "gl-post" || journalScreen("", "", "posted") != "gl-unpost" || journalScreen("JV", "opening", "") != "gl-opening-balance" {
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
	for _, screen := range []string{"jv-journal", "uv-journal", "sv-journal", "rv-journal", "pv-journal", "gl-opening-balance", "gl-post", "gl-unpost", "*"} {
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
	for _, p := range []map[string]bool{nil, {"sales-order": true}, {"jv-journal": false}, {"jv-journal:update": true}} {
		if canReadJournalSupport(p) {
			t.Fatal("unrelated or action-only permission granted evidence read")
		}
	}
}
