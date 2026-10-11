//go:build integration

package generalledger

import (
	"context"
	"testing"

	"github.com/shopspring/decimal"
)

// post_status (Champ "สถานะ" of reports 5532-5535/5506): posted = posted journals only, unposted = drafts only,
// all = both, summed from the stored lines without rounding in between
func TestPostgresReportPostStatus(t *testing.T) {
	p, _ := glpgTestDB(t)
	ctx := context.Background()
	glpgSeed(t, p, "A")
	scope := Scope{Holding: "H", Company: "A"}
	glpgSendJournal(t, p, 2, glpgJournal("A", "POSTED", "2026-01-02", "manual", "posted", "1000", "4000", "0.10"))
	glpgSendJournal(t, p, 3, glpgJournal("A", "DRAFT", "2026-01-03", "manual", "draft", "1000", "4000", "0.20"))
	base := ReportQuery{FiscalYear: "2026", From: "2026-01-01", To: "2026-01-31"}

	docs := func(status string) map[string]bool {
		q := base
		q.PostStatus = status
		rep, err := p.Report(ctx, scope, "gljournal", q)
		if err != nil {
			t.Fatalf("gljournal %s: %v", status, err)
		}
		out := map[string]bool{}
		for _, row := range rep.Rows {
			if row["docno"] != "" {
				out[row["docno"]] = true
			}
		}
		return out
	}
	if got := docs(PostStatusPosted); !got["POSTED"] || got["DRAFT"] {
		t.Fatalf("posted = %v", got)
	}
	if got := docs(PostStatusUnposted); got["POSTED"] || !got["DRAFT"] {
		t.Fatalf("unposted = %v", got)
	}
	if got := docs(PostStatusAll); !got["POSTED"] || !got["DRAFT"] {
		t.Fatalf("all = %v", got)
	}

	debit := func(status string) decimal.Decimal {
		q := base
		q.PostStatus = status
		rep, err := p.Report(ctx, scope, "trialbalance", q)
		if err != nil {
			t.Fatalf("trialbalance %s: %v", status, err)
		}
		d, err := decimal.NewFromString(rep.Totals["debit"])
		if err != nil {
			t.Fatalf("debit total %q: %v", rep.Totals["debit"], err)
		}
		return d
	}
	posted, all := debit(PostStatusPosted), debit(PostStatusAll)
	if !posted.Equal(decimal.RequireFromString("0.10")) || !all.Sub(posted).Equal(decimal.RequireFromString("0.20")) {
		t.Fatalf("trial balance debit posted=%s all=%s", posted, all)
	}
	// ledger 5534 (by date): the movements follow the status, the opening of every status reads every line before the start
	// (Champ GLRepGL2Dlg.cpp:438-444 by date = bctranssub without IsConfirm; GLRepGL2View.cpp:388-394 takes that row as
	// the opening, GLRepGL2View.cpp:757-766 filters only the movements by IsConfirm)
	ledger := func(status, from string) (string, string) {
		q := ReportQuery{FiscalYear: "2026", From: from, To: "2026-01-31", AccountCode: "1000", PostStatus: status}
		rep, err := p.Report(ctx, scope, "ledger", q)
		if err != nil {
			t.Fatalf("ledger %s from %s: %v", status, from, err)
		}
		return rep.Totals["opening"], rep.Totals["balance"]
	}
	for _, c := range []struct{ status, from, opening, balance string }{
		{PostStatusPosted, "2026-01-03", "0.10", "0.10"}, {PostStatusUnposted, "2026-01-03", "0.10", "0.30"}, {PostStatusAll, "2026-01-03", "0.10", "0.30"},
		// the draft of 2026-01-03 lies before the start: "posted" still opens with it, like Champ
		{PostStatusPosted, "2026-01-04", "0.30", "0.30"}, {PostStatusUnposted, "2026-01-04", "0.30", "0.30"}, {PostStatusAll, "2026-01-04", "0.30", "0.30"},
	} {
		opening, balance := ledger(c.status, c.from)
		if !decimal.RequireFromString(opening).Equal(decimal.RequireFromString(c.opening)) || !decimal.RequireFromString(balance).Equal(decimal.RequireFromString(c.balance)) {
			t.Fatalf("ledger %s from %s opening=%s balance=%s, want %s/%s", c.status, c.from, opening, balance, c.opening, c.balance)
		}
	}
	if _, err := p.Report(ctx, scope, "gljournal", ReportQuery{FiscalYear: "2026", From: "2026-01-01", To: "2026-01-31", PostStatus: "deleted"}); err == nil {
		t.Fatal("an unknown post status must be refused")
	}
}
