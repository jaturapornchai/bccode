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
