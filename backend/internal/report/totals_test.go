package report

import "testing"

func TestRoundAmountOnceHalfAwayFromZero(t *testing.T) {
	for in, want := range map[string]string{
		"0.30": "0.30", "0.005": "0.01", "-0.005": "-0.01", "0.0049": "0.00", "1.004": "1.00", "2.675": "2.68",
		"123456789012.345": "123456789012.35", "": "0.00", "abc": "0.00", " 7 ": "7.00",
	} {
		if got := RoundAmount(in, amountScale); got != want {
			t.Errorf("RoundAmount(%q) = %s, want %s", in, got, want)
		}
	}
}

func TestGrandTotalsKeepsWhitelistOnly(t *testing.T) {
	spec := []TotalKey{{Key: "debit"}, {Key: "count", Int: true}}
	got := grandTotals(spec, map[string]string{"debit": "0.1000", "count": "12", "secret": "99"})
	if len(got) != 2 || got["debit"] != "0.10" || got["count"] != "12" {
		t.Fatalf("totals = %v", got)
	}
}

func detail(debit, credit string) map[string]string {
	return map[string]string{KindKey: KindDetail, "debit": debit, "credit": credit}
}

// page sums count detail rows only; carried = running total; the last carried equals the grand total
func TestPageTotalsSumDetailRowsAndCarry(t *testing.T) {
	rows := []map[string]string{
		{KindKey: KindOpening, "debit": "1000", "credit": "0"},
		detail("0.1", "0"), detail("0.2", "0"),
		{KindKey: KindGroupTotal, "debit": "0.3", "credit": "0"},
		detail("0", "0.3"), detail("-5.005", "0"), detail("10", "4.995"),
	}
	spec := []TotalKey{{Key: "debit", Page: true}, {Key: "credit", Page: true}, {Key: "count", Int: true, Page: true}, {Key: "balance"}}
	pages := pageTotals(rows, spec, 3)
	if len(pages) != 3 {
		t.Fatalf("pages = %d", len(pages))
	}
	if p := pages[0]; p.FirstRow != 1 || p.LastRow != 3 || p.Sums["debit"] != "0.30" || p.Sums["count"] != "2" || p.Carried["debit"] != "0.30" {
		t.Fatalf("page 1 = %+v", p)
	}
	if _, ok := pages[0].Sums["balance"]; ok {
		t.Fatal("a total without Page must not be summed per page")
	}
	if p := pages[1]; p.Sums["debit"] != "-5.01" || p.Sums["credit"] != "0.30" || p.Carried["debit"] != "-4.71" {
		t.Fatalf("page 2 = %+v", p)
	}
	last := pages[2]
	if last.FirstRow != 7 || last.LastRow != 7 || last.Carried["debit"] != "5.30" || last.Carried["credit"] != "5.30" || last.Carried["count"] != "5" {
		t.Fatalf("last page = %+v", last)
	}
	if got := pageTotals(rows, []TotalKey{{Key: "balance"}}, 3); len(got) != 0 {
		t.Fatal("no page totals without Page keys")
	}
}
