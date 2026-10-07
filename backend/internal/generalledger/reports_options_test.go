package generalledger

import (
	"context"
	"strings"
	"testing"
)

// Champ "จากรหัสบัญชี–ถึงรหัสบัญชี" (docs/kms/champ/09-gl.md §5.4-5.5): a reversed range is the user's
// input error — rejected with a translatable code before any query runs (nil tx).
func TestReportAccountRangeRejectsReversedRange(t *testing.T) {
	_, err := newReportContext(context.Background(), nil, Scope{Holding: "H", Company: "A"}, ReportQuery{AccountFrom: " 5000 ", AccountTo: "1000"})
	if user := setUserError(t, err, "report_account_range_invalid"); user.Field != "accountfrom" {
		t.Fatalf("reversed range = %+v", user)
	}
}

// Every query that runs with r.args must reference $10 and $11, otherwise PostgreSQL cannot type the
// unused parameters. Byte-order (COLLATE "C") keeps the range identical on every database locale.
func TestAccountRangeFilterUsesBothParametersByteOrder(t *testing.T) {
	sql := accountRangeFilter("l.account_code")
	for _, want := range []string{`($10='' OR l.account_code COLLATE "C">=$10)`, `($11='' OR l.account_code COLLATE "C"<=$11)`} {
		if !strings.Contains(sql, want) {
			t.Fatalf("range filter %q misses %q", sql, want)
		}
	}
	r := reportContext{filter: "company=$1" + sql}
	if !strings.Contains(r.base(), "WITH RECURSIVE filtered AS") || !strings.Contains(r.base(), "$10") {
		t.Fatalf("base CTE = %q", r.base())
	}
}
