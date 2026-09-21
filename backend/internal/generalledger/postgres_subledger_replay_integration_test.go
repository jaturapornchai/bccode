//go:build integration

package generalledger

import (
	"encoding/json"
	"testing"
)

func TestPostgresSubledgerSourceReplayPreservesEvidence(t *testing.T) {
	f := newPGIntegrityFixture(t)
	bankAccount := Account{AccountCode: "1100", AccountType: "asset", NormalBalance: "debit", AllowPosting: true, IsActive: true, Names: []Name{{Code: "th", Name: "ธนาคาร"}}}
	f.run(Command{Resource: "accounts", Action: "create", Account: &bankAccount})
	j := Journal{DocNo: "EVIDENCE", Date: "2026-01-10", BookCode: "JV", FiscalYear: "2026", Description: "source evidence", Kind: "manual", BranchCode: "B1", SourceType: 2, SourceSystem: "office-import", SourceRecordID: "source-1", Lines: []Line{{AccountCode: "1000", Debit: Amount("100"), Credit: Amount("0")}, {AccountCode: "1100", Debit: Amount("100"), Credit: Amount("0")}, {AccountCode: "4000", Debit: Amount("0"), Credit: Amount("200")}}, Details: &JournalDetails{
		Partners:       []SubledgerPartner{{Code: "CUST1", Name: "ลูกค้า", IsCustomer: true, IsActive: true}},
		Documents:      []SubledgerDocument{{ID: "AR1", Ledger: "ar", PartnerCode: "CUST1", DocumentNo: "BILL1", Date: "2026-01-10", BranchCode: "B1", Kind: 1, Side: 1, Amount: Amount("100"), ControlAccountCode: "1000"}},
		Allocations:    []SubledgerAllocation{{ID: "ALLOC1", Ledger: "ar", DocumentID: "AR1", LineNumber: 1, Amount: Amount("100")}},
		BankAccounts:   []SubledgerBankAccount{{Code: "BANK1", BankName: "ธนาคาร", AccountNumber: "123", AccountName: "บริษัท", GLAccountCode: "1100", IsActive: true}},
		BankLines:      []SubledgerBankLine{{LineNumber: 2, BankAccountCode: "BANK1", Direction: 1}},
		StatementLines: []SubledgerStatementLine{{ID: "STMT1", BankAccountCode: "BANK1", SourceKey: "bank-row1", Date: "2026-01-10", Direction: 1, Amount: Amount("100")}},
		Matches:        []SubledgerMatch{{ID: "MATCH1", StatementLineID: "STMT1", LineNumber: 2, Amount: Amount("100")}},
	}}
	before, _ := json.Marshal(j)
	result := f.run(Command{Resource: "journals", Action: "create", Journal: &j})
	after, _ := json.Marshal(j)
	if string(before) != string(after) {
		t.Fatal("normalization mutated source request")
	}
	posted := f.post(f.journal(result.ID))
	retry := f.run(Command{Resource: "journals", Action: "create", Journal: &j})
	if retry.ID != result.ID {
		t.Fatal("source retry created duplicate")
	}
	tables := []string{"gl_source_journals", "gl_subledger_partners", "gl_subledger_bank_accounts", "gl_subledger_documents", "gl_subledger_allocations", "gl_subledger_bank_lines", "gl_subledger_statements", "gl_subledger_matches", "gl_subledger_audit"}
	snapshots := map[string]string{}
	for _, table := range tables {
		var snapshot string
		if err := f.db.QueryRow(`SELECT COALESCE(jsonb_agg(to_jsonb(t) ORDER BY to_jsonb(t)::text),'[]'::jsonb)::text FROM ` + table + ` t WHERE company='C'`).Scan(&snapshot); err != nil {
			t.Fatal(err)
		}
		snapshots[table] = snapshot
	}
	for _, action := range []string{"recalculate", "reprocess"} {
		f.run(Command{Resource: "processes", Action: action, ID: "2026", Reason: "rebuild projection"})
		for _, table := range tables {
			var snapshot string
			if err := f.db.QueryRow(`SELECT COALESCE(jsonb_agg(to_jsonb(t) ORDER BY to_jsonb(t)::text),'[]'::jsonb)::text FROM ` + table + ` t WHERE company='C'`).Scan(&snapshot); err != nil {
				t.Fatal(err)
			}
			if snapshot != snapshots[table] {
				t.Fatalf("replay changed authoritative evidence %s", table)
			}
		}
		current := f.journal(posted.ID)
		if current.Version != posted.Version || current.Details == nil {
			t.Fatal("replay changed journal revision/details")
		}
	}
}

func TestPostgresSubledgerFutureReversalKeepsPriorPeriodEvidence(t *testing.T) {
	f := newPGIntegrityFixture(t)
	j := Journal{DocNo: "HISTORY", Date: "2026-01-10", BookCode: "JV", FiscalYear: "2026", Description: "historical debt", Kind: "manual", BranchCode: "B1", Lines: []Line{{AccountCode: "1000", Debit: Amount("100"), Credit: Amount("0")}, {AccountCode: "4000", Debit: Amount("0"), Credit: Amount("100")}}, Details: &JournalDetails{
		Partners:    []SubledgerPartner{{Code: "CUST1", Name: "ลูกค้า", IsCustomer: true, IsActive: true}},
		Documents:   []SubledgerDocument{{ID: "HISTORY-AR", Ledger: "ar", PartnerCode: "CUST1", DocumentNo: "BILL1", Date: "2026-01-10", BranchCode: "B1", Kind: 1, Side: 1, Amount: Amount("100"), ControlAccountCode: "1000"}},
		Allocations: []SubledgerAllocation{{ID: "HISTORY-ALLOC", Ledger: "ar", DocumentID: "HISTORY-AR", LineNumber: 1, Amount: Amount("100")}},
	}}
	created := f.run(Command{Resource: "journals", Action: "create", Journal: &j})
	posted := f.post(f.journal(created.ID))
	f.run(Command{Resource: "journals", Action: "reverse", ID: posted.ID, Version: posted.Version, DocNo: "FUTURE-REV", Date: "2027-01-01", Reason: "future effective reversal"})
	for _, tc := range []struct{ asof, want string }{{"2026-12-31", "100"}, {"2027-01-01", "0"}} {
		page, err := f.store.pg.SubledgerList(f.ctx, f.scope, "documents", "HISTORY-AR", 1, 50, tc.asof)
		if err != nil {
			t.Fatal(err)
		}
		if len(page.Items) != 1 {
			t.Fatalf("missing history %s", tc.asof)
		}
		var document SubledgerOpenDocument
		if err = json.Unmarshal(page.Items[0], &document); err != nil {
			t.Fatal(err)
		}
		want := Amount(tc.want).Decimal()
		if !document.Posted.Decimal().Equal(want) || !document.Allocated.Decimal().Equal(want) {
			t.Fatalf("%s allocated=%s posted=%s want=%s", tc.asof, document.Allocated, document.Posted, tc.want)
		}
	}
}

func TestPostgresSubledgerPartialPostingDoesNotFundUnpostedDebt(t *testing.T) {
	f := newPGIntegrityFixture(t)
	makeDebt := func(doc string, amount Amount, details *JournalDetails) Journal {
		return Journal{DocNo: doc, Date: "2026-01-10", BookCode: "JV", FiscalYear: "2026", Description: "partial document allocation", Kind: "manual", BranchCode: "B1", Lines: []Line{{AccountCode: "1000", Debit: amount, Credit: Amount("0")}, {AccountCode: "4000", Debit: Amount("0"), Credit: amount}}, Details: details}
	}
	first := makeDebt("PART1", Amount("60"), &JournalDetails{Partners: []SubledgerPartner{{Code: "C1", Name: "ลูกค้า", IsCustomer: true, IsActive: true}}, Documents: []SubledgerDocument{{ID: "PARTIAL-AR", Ledger: "ar", PartnerCode: "C1", DocumentNo: "BILL100", Date: "2026-01-10", BranchCode: "B1", Kind: 1, Side: 1, Amount: Amount("100"), ControlAccountCode: "1000"}}, Allocations: []SubledgerAllocation{{ID: "PART-ALLOC1", Ledger: "ar", DocumentID: "PARTIAL-AR", LineNumber: 1, Amount: Amount("60")}}})
	firstResult := f.run(Command{Resource: "journals", Action: "create", Journal: &first})
	second := makeDebt("PART2", Amount("40"), &JournalDetails{Allocations: []SubledgerAllocation{{ID: "PART-ALLOC2", Ledger: "ar", DocumentID: "PARTIAL-AR", LineNumber: 1, Amount: Amount("40")}}})
	f.run(Command{Resource: "journals", Action: "create", Journal: &second})
	posted := f.post(f.journal(firstResult.ID))
	readDebt := func() SubledgerOpenDocument {
		t.Helper()
		page, err := f.store.pg.SubledgerList(f.ctx, f.scope, "documents", "PARTIAL-AR", 1, 20, "")
		if err != nil || len(page.Items) != 1 {
			t.Fatalf("debt query: %v %+v", err, page)
		}
		var d SubledgerOpenDocument
		if err = json.Unmarshal(page.Items[0], &d); err != nil {
			t.Fatal(err)
		}
		return d
	}
	d := readDebt()
	if !d.Allocated.Decimal().Equal(Amount("100").Decimal()) || !d.Posted.Decimal().Equal(Amount("60").Decimal()) {
		t.Fatalf("draft funding counted as posted: %+v", d)
	}
	zero := int64(0)
	f.run(Command{Resource: "journals", Action: "review", ID: posted.ID, Version: posted.Version, Review: &ReviewInput{Status: 3, ExpectedEventNo: &zero}})
	cash := Account{AccountCode: "1100", AccountType: "asset", NormalBalance: "debit", AllowPosting: true, IsActive: true, Names: []Name{{Code: "th", Name: "ธนาคาร"}}}
	f.run(Command{Resource: "accounts", Action: "create", Account: &cash})
	payment := Journal{DocNo: "PAY", Date: "2026-01-11", BookCode: "RV", FiscalYear: "2026", Description: "payment", Kind: "manual", BranchCode: "B1", Lines: []Line{{AccountCode: "1100", Debit: Amount("80"), Credit: Amount("0")}, {AccountCode: "1000", Debit: Amount("0"), Credit: Amount("80")}}, Details: &JournalDetails{Documents: []SubledgerDocument{{ID: "PAY-DOC", Ledger: "ar", PartnerCode: "C1", DocumentNo: "RECEIPT", Date: "2026-01-11", BranchCode: "B1", Kind: 2, Side: 2, Amount: Amount("80"), ControlAccountCode: "1000"}}, Allocations: []SubledgerAllocation{{ID: "PAY-ALLOC", Ledger: "ar", DocumentID: "PAY-DOC", LineNumber: 2, Amount: Amount("80")}}, Settlements: []SubledgerSettlement{{ID: "PAY-SETTLE", Ledger: "ar", PartnerCode: "C1", DebtDocumentID: "PARTIAL-AR", PaymentDocumentID: "PAY-DOC", Date: "2026-01-11", Amount: Amount("80")}}}}
	f.deny(f.scope, Command{Resource: "journals", Action: "create", Journal: &payment})
	payment.Lines[0].Debit = Amount("60")
	payment.Lines[1].Credit = Amount("60")
	payment.Details.Documents[0].Amount = Amount("60")
	payment.Details.Allocations[0].Amount = Amount("60")
	payment.Details.Settlements[0].Amount = Amount("60")
	payResult := f.run(Command{Resource: "journals", Action: "create", Journal: &payment})
	d = readDebt()
	if !d.Settled.Decimal().IsZero() {
		t.Fatalf("draft reservation counted as settled: %+v", d)
	}
	review, err := f.store.pg.JournalReview(f.ctx, f.scope, posted.ID)
	if err != nil || review.Status != 1 {
		t.Fatalf("related review remained valid: %+v %v", review, err)
	}
	f.post(f.journal(payResult.ID))
	d = readDebt()
	if !d.Settled.Decimal().Equal(Amount("60").Decimal()) || !d.Remaining.Decimal().IsZero() {
		t.Fatalf("posted partial settlement incorrect: %+v", d)
	}
}
