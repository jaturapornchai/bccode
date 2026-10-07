//go:build integration

package generalledger

import "testing"

// closing/reversal journals are produced only by year-end and reverse; "mapping" has no producer.
// A client that sends them on create would hide its amounts from the P&L and the VAT/WHT reports
// (those reports skip closing/opening and treat reversals as cancellations).
func TestPostgresCreateRejectsSystemOnlyJournalKinds(t *testing.T) {
	f := newPGIntegrityFixture(t)
	for _, kind := range []string{"closing", "reversal", "mapping"} {
		j := Journal{DocNo: "JV-" + kind, Date: "2026-01-10", BookCode: "JV", FiscalYear: "2026", Description: "forged " + kind, Kind: kind, BranchCode: "B1",
			Lines: []Line{{AccountCode: "1000", Debit: Amount("100"), Credit: Amount("0")}, {AccountCode: "4000", Debit: Amount("0"), Credit: Amount("100")}}}
		_, err := f.execute(f.scope, Command{Resource: "journals", Action: "create", Journal: &j})
		user, ok := AsUserError(err)
		if !ok || user.Code != "journal_kind_not_allowed" || user.Field != "kind" {
			t.Fatalf("create kind=%s: err = %#v, want journal_kind_not_allowed on field kind", kind, err)
		}
		var count int
		if err := f.db.QueryRow(`SELECT COUNT(*) FROM gl_records WHERE company='C' AND kind='journals' AND code=$1`, j.DocNo).Scan(&count); err != nil || count != 0 {
			t.Fatalf("rejected %s journal stored: count=%d err=%v", kind, count, err)
		}
	}
	f.draft("JV-OK", "B1") // manual still works
}
