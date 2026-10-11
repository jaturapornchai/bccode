package fixedasset

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

// A company with no assets must get "rows":[] — JSON null crashed the PND50 depreciation tab
// (frontend read rows.length on null; found by `npm run uat`, 2026-10-11).
func TestReportsEmptyCompanyRowsAreJSONArray(t *testing.T) {
	connect, company := testConnector(t)
	ctx := context.Background()
	scope := Scope{Holding: "TEST_HOLDING", Company: company, Branch: "HQ", Actor: "CPA_AUDITOR"}
	reporter := NewReporter(connect)

	schedule, err := reporter.GetAssetScheduleReport(ctx, scope, "2026", 12, "")
	if err != nil {
		t.Fatalf("GetAssetScheduleReport: %v", err)
	}
	tax, err := reporter.GetTaxReconciliationReport(ctx, scope, "2026")
	if err != nil {
		t.Fatalf("GetTaxReconciliationReport: %v", err)
	}
	for name, report := range map[string]*ReportResult{"schedule": schedule, "tax": tax} {
		body, err := json.Marshal(report)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(body), `"rows":[]`) {
			t.Fatalf("%s report rows must be an empty JSON array: %s", name, body)
		}
	}
}
