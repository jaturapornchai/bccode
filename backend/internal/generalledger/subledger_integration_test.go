//go:build integration

package generalledger

import (
	"encoding/json"
	"strings"
	"testing"
)

func subledgerFixtures(t *testing.T) (*pgIntegrityFixture, Journal, Journal) {
	t.Helper()
	f := newPGIntegrityFixture(t)
	bank := Account{AccountCode: "1100", AccountType: "asset", NormalBalance: "debit", AllowPosting: true, IsActive: true, Names: []Name{{Code: "th", Name: "bank"}}}
	f.run(Command{Resource: "accounts", Action: "create", Account: &bank})
	debt := Journal{DocNo: "DEBT", Date: "2026-01-10", BookCode: "JV", FiscalYear: "2026", Kind: "manual", Description: "evidence", BranchCode: "B1", Lines: []Line{{AccountCode: "1000", Debit: "0.30", Credit: "0"}, {AccountCode: "4000", Debit: "0", Credit: "0.30"}}, Details: &JournalDetails{
		Partners:    []SubledgerPartner{{Code: "P", Name: "customer", IsCustomer: true, IsSupplier: true, IsActive: true}},
		Documents:   []SubledgerDocument{{ID: "D1", Ledger: "ar", PartnerCode: "P", DocumentNo: "INV1", Date: "2026-01-10", BranchCode: "B1", Kind: 1, Side: 1, Amount: "0.10", ControlAccountCode: "1000"}, {ID: "D2", Ledger: "ar", PartnerCode: "P", DocumentNo: "INV2", Date: "2026-01-10", BranchCode: "B1", Kind: 1, Side: 1, Amount: "0.20", ControlAccountCode: "1000"}},
		Allocations: []SubledgerAllocation{{ID: "A1", Ledger: "ar", DocumentID: "D1", LineNumber: 1, Amount: "0.10"}, {ID: "A2", Ledger: "ar", DocumentID: "D2", LineNumber: 1, Amount: "0.20"}}}}
	r := f.run(Command{Resource: "journals", Action: "create", Journal: &debt})
	debt = f.post(f.journal(r.ID))
	payment := Journal{DocNo: "PAY", Date: "2026-01-11", BookCode: "JV", FiscalYear: "2026", Kind: "manual", Description: "evidence", BranchCode: "B1", Lines: []Line{{AccountCode: "1100", Debit: "0.15", Credit: "0"}, {AccountCode: "1100", Debit: "0.15", Credit: "0"}, {AccountCode: "1000", Debit: "0", Credit: "0.30"}}, Details: &JournalDetails{
		Documents:      []SubledgerDocument{{ID: "P1", Ledger: "ar", PartnerCode: "P", DocumentNo: "PAY1", Date: "2026-01-11", BranchCode: "B1", Kind: 2, Side: 2, Amount: "0.15", ControlAccountCode: "1000"}, {ID: "P2", Ledger: "ar", PartnerCode: "P", DocumentNo: "PAY2", Date: "2026-01-11", BranchCode: "B1", Kind: 2, Side: 2, Amount: "0.15", ControlAccountCode: "1000"}},
		Allocations:    []SubledgerAllocation{{ID: "AP1", Ledger: "ar", DocumentID: "P1", LineNumber: 3, Amount: "0.15"}, {ID: "AP2", Ledger: "ar", DocumentID: "P2", LineNumber: 3, Amount: "0.15"}},
		BankAccounts:   []SubledgerBankAccount{{Code: "BANK", BankName: "bank", AccountNumber: "123", AccountName: "company", GLAccountCode: "1100", IsActive: true}},
		BankLines:      []SubledgerBankLine{{LineNumber: 1, BankAccountCode: "BANK", Direction: 1}, {LineNumber: 2, BankAccountCode: "BANK", Direction: 1}},
		StatementLines: []SubledgerStatementLine{{ID: "S1", BankAccountCode: "BANK", SourceKey: "FILE-1", Date: "2026-01-11", Direction: 1, Amount: "0.10"}, {ID: "S2", BankAccountCode: "BANK", SourceKey: "FILE-2", Date: "2026-01-11", Direction: 1, Amount: "0.20"}},
		Settlements:    []SubledgerSettlement{{ID: "SET1", Ledger: "ar", PartnerCode: "P", DebtDocumentID: "D1", PaymentDocumentID: "P1", Date: "2026-01-11", Amount: "0.10"}, {ID: "SET2", Ledger: "ar", PartnerCode: "P", DebtDocumentID: "D2", PaymentDocumentID: "P1", Date: "2026-01-11", Amount: "0.05"}, {ID: "SET3", Ledger: "ar", PartnerCode: "P", DebtDocumentID: "D2", PaymentDocumentID: "P2", Date: "2026-01-11", Amount: "0.15"}},
		Matches:        []SubledgerMatch{{ID: "M1", StatementLineID: "S1", LineNumber: 1, Amount: "0.10"}, {ID: "M2", StatementLineID: "S2", LineNumber: 1, Amount: "0.05"}, {ID: "M3", StatementLineID: "S2", LineNumber: 2, Amount: "0.15"}}}}
	r = f.run(Command{Resource: "journals", Action: "create", Journal: &payment})
	return f, f.journal(debt.ID), f.journal(r.ID)
}

func TestPostgresSubledgerRevokedMastersCheckedWithoutInlineCopies(t *testing.T) {
	f, debt, payment := subledgerFixtures(t)
	payment = f.post(payment)
	f.run(subledgerReconcile(f, payment, JournalDetails{Withdrawals: []SubledgerWithdrawal{{Kind: "settlement", ID: "SET1", Reason: "correct"}, {Kind: "settlement", ID: "SET2", Reason: "correct"}, {Kind: "settlement", ID: "SET3", Reason: "correct"}, {Kind: "match", ID: "M3", Reason: "correct"}}}))
	payment = f.journal(payment.ID)
	debt = f.journal(debt.ID)
	setActive := func(table string, active bool) {
		t.Helper()
		if _, err := f.db.Exec(`UPDATE `+table+` SET payload=jsonb_set(payload,'{is_active}',to_jsonb($1::boolean)) WHERE company='C'`, active); err != nil {
			t.Fatal(err)
		}
	}
	denyContaining := func(cmd Command, text string) {
		t.Helper()
		_, err := f.execute(f.scope, cmd)
		if err == nil || !strings.Contains(err.Error(), text) {
			t.Fatalf("expected active-master denial %s, got %v", text, err)
		}
	}
	settlement := JournalDetails{Settlements: []SubledgerSettlement{{ID: "SET-NEW", Ledger: "ar", PartnerCode: "P", DebtDocumentID: "D1", PaymentDocumentID: "P1", Date: "2026-01-11", Amount: "0.10"}}}
	replacement := JournalDetails{Withdrawals: []SubledgerWithdrawal{{Kind: "allocation", ID: "A1", Reason: "correct"}}, Allocations: []SubledgerAllocation{{ID: "A1-REPLACED", Ledger: "ar", DocumentID: "D1", LineNumber: 1, Amount: "0.10"}}}
	setActive("gl_subledger_partners", false)
	denyContaining(subledgerReconcile(f, payment, settlement), "คู่ค้า")
	denyContaining(subledgerReconcile(f, debt, replacement), "คู่ค้า")
	setActive("gl_subledger_partners", true)
	f.run(subledgerReconcile(f, debt, replacement))
	payment = f.journal(payment.ID)
	f.run(subledgerReconcile(f, payment, settlement))
	payment = f.journal(payment.ID)
	match := JournalDetails{Matches: []SubledgerMatch{{ID: "M-NEW", StatementLineID: "S2", JournalID: payment.ID, LineNumber: 2, Amount: "0.15"}}}
	setActive("gl_subledger_bank_accounts", false)
	denyContaining(subledgerReconcile(f, payment, match), "ธนาคาร")
	setActive("gl_subledger_bank_accounts", true)
	f.run(subledgerReconcile(f, payment, match))
}

func subledgerReconcile(f *pgIntegrityFixture, j Journal, d JournalDetails) Command {
	return Command{Resource: "journals", Action: "reconcile", ID: j.ID, Version: j.Version, Journal: &Journal{Details: &d}, Reason: "reconcile evidence"}
}
func subledgerDocumentAt(t *testing.T, f *pgIntegrityFixture, id, asof string) SubledgerOpenDocument {
	t.Helper()
	page, err := f.store.pg.SubledgerList(f.ctx, f.scope, "documents", id, 1, 50, asof)
	if err != nil {
		t.Fatal(err)
	}
	for _, raw := range page.Items {
		var row SubledgerOpenDocument
		if err = json.Unmarshal(raw, &row); err != nil {
			t.Fatal(err)
		}
		if row.ID == id {
			return row
		}
	}
	t.Fatalf("document %s missing", id)
	return SubledgerOpenDocument{}
}
func assertSubledgerAmount(t *testing.T, got Amount, want string) {
	t.Helper()
	expected, err := ParseAmount(want)
	if err != nil || !got.Decimal().Equal(expected.Decimal()) {
		t.Fatalf("amount %s want %s", got, want)
	}
}

func TestPostgresSubledgerManyToManyAtomicLifecycle(t *testing.T) {
	f, debt, payment := subledgerFixtures(t)
	// Draft reservations cannot report as paid or matched.
	assertSubledgerAmount(t, subledgerDocumentAt(t, f, "D2", "").Remaining, "0.20")
	page, err := f.store.pg.SubledgerList(f.ctx, f.scope, "statements", "", 1, 1, "")
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 2 || len(page.Items) != 1 {
		t.Fatal("SQL pagination failed")
	}
	var statement SubledgerStatementBalance
	if err = json.Unmarshal(page.Items[0], &statement); err != nil {
		t.Fatal(err)
	}
	assertSubledgerAmount(t, statement.Matched, "0")
	payment = f.post(payment)
	updatedDebt := f.journal(debt.ID)
	if updatedDebt.Version <= debt.Version {
		t.Fatal("posting payment did not invalidate debt review")
	}
	assertSubledgerAmount(t, subledgerDocumentAt(t, f, "D2", "").Remaining, "0")
	report, err := f.store.pg.SubledgerReport(f.ctx, f.scope, "ar-outstanding", ReportQuery{To: "9999-12-30"})
	if err != nil {
		t.Fatal(err)
	}
	if report.TotalRows != 0 {
		t.Fatalf("settled report has %d rows", report.TotalRows)
	}
	for _, kind := range []string{"partners", "bank-accounts", "documents", "statements", "bank-lines", "allocations", "settlements", "matches"} {
		if _, err = f.store.pg.SubledgerList(f.ctx, f.scope, kind, "", 1, 2, ""); err != nil {
			t.Fatalf("%s: %v", kind, err)
		}
	}
	bad := JournalDetails{Settlements: []SubledgerSettlement{{ID: "OVER", Ledger: "ar", PartnerCode: "P", DebtDocumentID: "D1", PaymentDocumentID: "P1", Date: "2026-01-11", Amount: "0.01"}}}
	f.deny(f.scope, subledgerReconcile(f, payment, bad))
	bad = JournalDetails{Matches: []SubledgerMatch{{ID: "OVER-M", StatementLineID: "S1", JournalID: payment.ID, LineNumber: 1, Amount: "0.01"}}}
	f.deny(f.scope, subledgerReconcile(f, payment, bad))
	f.deny(f.scope, Command{Resource: "journals", Action: "reverse", ID: payment.ID, Version: payment.Version, DocNo: "REV-PAY", Date: "2026-01-12", Reason: "correction"})
	// A failed mixed command leaves no imported statement or audit entry.
	var before int
	f.db.QueryRow(`SELECT COUNT(*) FROM gl_subledger_audit WHERE company='C'`).Scan(&before)
	bad.StatementLines = []SubledgerStatementLine{{ID: "ROLLBACK", BankAccountCode: "BANK", SourceKey: "rollback", Date: "2026-01-11", Direction: 1, Amount: "0.10"}}
	f.deny(f.scope, subledgerReconcile(f, payment, bad))
	var count int
	f.db.QueryRow(`SELECT COUNT(*) FROM gl_subledger_statements WHERE id='ROLLBACK'`).Scan(&count)
	if count != 0 {
		t.Fatal("partial import committed")
	}
	f.db.QueryRow(`SELECT COUNT(*) FROM gl_subledger_audit WHERE company='C'`).Scan(&count)
	if count != before {
		t.Fatal("failed command audit committed")
	}
	withdrawals := []SubledgerWithdrawal{{Kind: "settlement", ID: "SET1", Reason: "correction"}, {Kind: "settlement", ID: "SET2", Reason: "correction"}, {Kind: "settlement", ID: "SET3", Reason: "correction"}, {Kind: "match", ID: "M1", Reason: "correction"}, {Kind: "match", ID: "M2", Reason: "correction"}, {Kind: "match", ID: "M3", Reason: "correction"}}
	f.run(subledgerReconcile(f, payment, JournalDetails{Withdrawals: withdrawals}))
	payment = f.journal(payment.ID)
	assertSubledgerAmount(t, subledgerDocumentAt(t, f, "D2", "").Remaining, "0.20")
	f.run(Command{Resource: "journals", Action: "reverse", ID: payment.ID, Version: payment.Version, DocNo: "REV-PAY", Date: "2026-01-12", Reason: "correction"})
	if _, err = f.db.Exec(`UPDATE gl_subledger_audit SET action='tamper' WHERE company='C'`); err == nil {
		t.Fatal("audit mutation accepted")
	}
}

func TestPostgresSubledgerAllocationReplacementAndScope(t *testing.T) {
	f, debt, payment := subledgerFixtures(t)
	payment = f.post(payment)
	// Withdraw settlements before correcting debt evidence.
	f.run(subledgerReconcile(f, payment, JournalDetails{Withdrawals: []SubledgerWithdrawal{{Kind: "settlement", ID: "SET1", Reason: "correct"}, {Kind: "settlement", ID: "SET2", Reason: "correct"}, {Kind: "settlement", ID: "SET3", Reason: "correct"}}}))
	debt = f.journal(debt.ID)
	withdrawal := SubledgerWithdrawal{Kind: "allocation", ID: "A1", Reason: "reallocate"}
	f.deny(f.scope, subledgerReconcile(f, debt, JournalDetails{Withdrawals: []SubledgerWithdrawal{withdrawal}}))
	var reversed bool
	f.db.QueryRow(`SELECT reversed_at IS NOT NULL FROM gl_subledger_allocations WHERE id='A1'`).Scan(&reversed)
	if reversed {
		t.Fatal("failed withdrawal committed")
	}
	replacement := SubledgerAllocation{ID: "A1-NEW", Ledger: "ar", DocumentID: "D1", LineNumber: 1, Amount: "0.10"}
	f.run(subledgerReconcile(f, debt, JournalDetails{Withdrawals: []SubledgerWithdrawal{withdrawal}, Allocations: []SubledgerAllocation{replacement}}))
	debt = f.journal(debt.ID)
	assertSubledgerAmount(t, subledgerDocumentAt(t, f, "D1", "").Posted, "0.10")
	other := f.scope
	other.Branch = "B2"
	f.deny(other, subledgerReconcile(f, debt, JournalDetails{Withdrawals: []SubledgerWithdrawal{{Kind: "allocation", ID: "A1-NEW", Reason: "attack"}}}))
	for _, kind := range []string{"documents", "statements", "bank-lines", "allocations", "settlements", "matches"} {
		page, err := f.store.pg.SubledgerList(f.ctx, other, kind, "", 1, 50, "")
		if err != nil {
			t.Fatal(err)
		}
		if page.Total != 0 {
			t.Fatalf("%s leaked branch", kind)
		}
	}
	cross := f.scope
	cross.Company = "OTHER"
	page, err := f.store.pg.SubledgerList(f.ctx, cross, "documents", "D1", 1, 50, "")
	if err != nil || page.Total != 0 {
		t.Fatal("company isolation failed", err)
	}
}

func TestPostgresSubledgerStatementDedupAndHistory(t *testing.T) {
	f, _, payment := subledgerFixtures(t)
	payment = f.post(payment)
	identical := payment.Details.StatementLines[0]
	identical.ID = "DIFFERENT-ID"
	f.run(subledgerReconcile(f, payment, JournalDetails{StatementLines: []SubledgerStatementLine{identical}}))
	payment = f.journal(payment.ID)
	var count int
	f.db.QueryRow(`SELECT COUNT(*) FROM gl_subledger_statements WHERE source_key='FILE-1'`).Scan(&count)
	if count != 1 {
		t.Fatal("duplicate statement source")
	}
	identical.Amount = "0.11"
	f.deny(f.scope, subledgerReconcile(f, payment, JournalDetails{StatementLines: []SubledgerStatementLine{identical}}))
	// Isolated fixture timestamps model evidence known before and after cutoff.
	if _, err := f.db.Exec(`UPDATE gl_records SET payload=jsonb_set(payload,'{postedat}','"2026-01-11T00:00:00Z"'::jsonb) WHERE kind='journals' AND payload->>'status'='posted'; UPDATE gl_subledger_allocations SET created_at='2026-01-10'; UPDATE gl_subledger_settlements SET created_at='2026-02-01'; UPDATE gl_subledger_matches SET created_at='2026-02-01'; UPDATE gl_subledger_statements SET created_at='2026-01-11'`); err != nil {
		t.Fatal(err)
	}
	assertSubledgerAmount(t, subledgerDocumentAt(t, f, "D1", "2026-01-31").Settled, "0")
	assertSubledgerAmount(t, subledgerDocumentAt(t, f, "D1", "2026-02-28").Settled, "0.10")
	if _, err := f.db.Exec(`UPDATE gl_subledger_settlements SET reversed_at='2026-03-01',reversed_by='test',reversal_reason='historical' WHERE id='SET1'`); err != nil {
		t.Fatal(err)
	}
	assertSubledgerAmount(t, subledgerDocumentAt(t, f, "D1", "2026-02-28").Settled, "0.10")
	assertSubledgerAmount(t, subledgerDocumentAt(t, f, "D1", "2026-03-31").Settled, "0")
}

func TestPostgresSubledgerPartialJournalsAndConcurrentCaps(t *testing.T) {
	f := newPGIntegrityFixture(t)
	journal := func(no, amount string) *Journal {
		return &Journal{DocNo: no, Date: "2026-01-10", Description: "partial", BookCode: "JV", FiscalYear: "2026", Kind: "manual", BranchCode: "B1", Lines: []Line{{AccountCode: "1000", Debit: Amount(amount), Credit: "0"}, {AccountCode: "4000", Debit: "0", Credit: Amount(amount)}}}
	}
	first := journal("PART1", "60")
	first.Details = &JournalDetails{Partners: []SubledgerPartner{{Code: "P", Name: "customer", IsCustomer: true, IsActive: true}}, Documents: []SubledgerDocument{{ID: "SHARED", Ledger: "ar", PartnerCode: "P", DocumentNo: "SHARED", Date: "2026-01-10", BranchCode: "B1", Kind: 1, Side: 1, Amount: "100", ControlAccountCode: "1000"}}, Allocations: []SubledgerAllocation{{ID: "PART-A1", Ledger: "ar", DocumentID: "SHARED", LineNumber: 1, Amount: "60"}}}
	r := f.run(Command{Resource: "journals", Action: "create", Journal: first})
	firstID := r.ID
	second := journal("PART2", "40")
	second.Details = &JournalDetails{Allocations: []SubledgerAllocation{{ID: "PART-A2", Ledger: "ar", DocumentID: "SHARED", LineNumber: 1, Amount: "40"}}}
	r = f.run(Command{Resource: "journals", Action: "create", Journal: second})
	secondID := r.ID
	f.post(f.journal(firstID))
	row := subledgerDocumentAt(t, f, "SHARED", "")
	assertSubledgerAmount(t, row.Allocated, "100")
	assertSubledgerAmount(t, row.Posted, "60")
	payment := journal("PAY-OVER", "80")
	payment.Lines = []Line{{AccountCode: "1000", Debit: "0", Credit: "80"}, {AccountCode: "4000", Debit: "80", Credit: "0"}}
	payment.Details = &JournalDetails{Documents: []SubledgerDocument{{ID: "PAY-PART", Ledger: "ar", PartnerCode: "P", DocumentNo: "PAY-PART", Date: "2026-01-10", BranchCode: "B1", Kind: 2, Side: 2, Amount: "80", ControlAccountCode: "1000"}}, Allocations: []SubledgerAllocation{{ID: "PART-PAY", Ledger: "ar", DocumentID: "PAY-PART", LineNumber: 1, Amount: "80"}}, Settlements: []SubledgerSettlement{{ID: "UNPOSTED-FUND", Ledger: "ar", PartnerCode: "P", DebtDocumentID: "SHARED", PaymentDocumentID: "PAY-PART", Date: "2026-01-10", Amount: "80"}}}
	f.deny(f.scope, Command{Resource: "journals", Action: "create", Journal: payment})
	secondStored := f.journal(secondID)
	f.run(Command{Resource: "journals", Action: "delete", ID: secondID, Version: secondStored.Version})
	row = subledgerDocumentAt(t, f, "SHARED", "")
	assertSubledgerAmount(t, row.Allocated, "60")
	assertSubledgerAmount(t, row.Posted, "60")
	// Two commands racing for the last 40 cannot both reserve it.
	results := make(chan error, 2)
	for _, suffix := range []string{"X", "Y"} {
		go func(suffix string) {
			j := journal("RACE-"+suffix, "40")
			j.Details = &JournalDetails{Allocations: []SubledgerAllocation{{ID: "RACE-" + suffix, Ledger: "ar", DocumentID: "SHARED", LineNumber: 1, Amount: "40"}}}
			_, err := f.execute(f.scope, Command{Resource: "journals", Action: "create", Journal: j})
			results <- err
		}(suffix)
	}
	success := 0
	for i := 0; i < 2; i++ {
		if <-results == nil {
			success++
		}
	}
	if success != 1 {
		t.Fatalf("concurrent caps accepted %d", success)
	}
	assertSubledgerAmount(t, subledgerDocumentAt(t, f, "SHARED", "").Allocated, "100")
}

func TestPostgresSubledgerAPNetAndBangkokCutoff(t *testing.T) {
	f := newPGIntegrityFixture(t)
	account := Account{AccountCode: "2000", AccountType: "liability", NormalBalance: "credit", AllowPosting: true, IsActive: true, Names: []Name{{Code: "th", Name: "AP"}}}
	f.run(Command{Resource: "accounts", Action: "create", Account: &account})
	j := Journal{DocNo: "AP-DEBT", Date: "2026-01-10", Description: "ap", BookCode: "JV", FiscalYear: "2026", Kind: "manual", BranchCode: "B1", Lines: []Line{{AccountCode: "5000", Debit: "100", Credit: "0"}, {AccountCode: "2000", Debit: "0", Credit: "100"}}, Details: &JournalDetails{Partners: []SubledgerPartner{{Code: "SUP", Name: "supplier", IsSupplier: true, IsActive: true}}, Documents: []SubledgerDocument{{ID: "AP-D", Ledger: "ap", PartnerCode: "SUP", DocumentNo: "AP-D", Date: "2026-01-10", Kind: 1, Side: 1, Amount: "100", ControlAccountCode: "2000"}}, Allocations: []SubledgerAllocation{{ID: "AP-AD", Ledger: "ap", DocumentID: "AP-D", LineNumber: 2, Amount: "100"}}}}
	r := f.run(Command{Resource: "journals", Action: "create", Journal: &j})
	f.post(f.journal(r.ID))
	j.DocNo = "AP-PAY"
	j.Lines = []Line{{AccountCode: "2000", Debit: "20", Credit: "0"}, {AccountCode: "1000", Debit: "0", Credit: "20"}}
	j.Details = &JournalDetails{Documents: []SubledgerDocument{{ID: "AP-P", Ledger: "ap", PartnerCode: "SUP", DocumentNo: "AP-P", Date: "2026-01-10", Kind: 2, Side: 2, Amount: "20", ControlAccountCode: "2000"}}, Allocations: []SubledgerAllocation{{ID: "AP-AP", Ledger: "ap", DocumentID: "AP-P", LineNumber: 1, Amount: "20"}}}
	r = f.run(Command{Resource: "journals", Action: "create", Journal: &j})
	f.post(f.journal(r.ID))
	report, err := f.store.pg.SubledgerReport(f.ctx, f.scope, "ap-outstanding", ReportQuery{To: "9999-12-30", Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	assertSubledgerAmount(t, Amount(report.Totals["remaining_amount"]), "80")
	if report.TotalRows != 2 || len(report.Rows) != 1 {
		t.Fatal("report pagination/totals failed")
	}
	// 17:00 UTC is midnight of the next accounting day in Bangkok.
	if _, err = f.db.Exec(`UPDATE gl_records SET payload=jsonb_set(payload,'{postedat}','"2026-01-11T00:00:00Z"'::jsonb) WHERE kind='journals' AND payload->>'status'='posted'; UPDATE gl_subledger_allocations SET created_at='2026-01-31 16:59:59+00' WHERE id='AP-AD'; UPDATE gl_subledger_allocations SET created_at='2026-01-31 17:00:00+00' WHERE id='AP-AP'`); err != nil {
		t.Fatal(err)
	}
	assertSubledgerAmount(t, subledgerDocumentAt(t, f, "AP-D", "2026-01-31").Posted, "100")
	assertSubledgerAmount(t, subledgerDocumentAt(t, f, "AP-P", "2026-01-31").Posted, "0")
	assertSubledgerAmount(t, subledgerDocumentAt(t, f, "AP-P", "2026-02-01").Posted, "20")
}

func TestPostgresSubledgerDraftRemovalKeepsBankEvidence(t *testing.T) {
	f, _, payment := subledgerFixtures(t)
	moved := payment
	moved.BranchCode = "B2"
	f.deny(f.scope, Command{Resource: "journals", Action: "update", ID: payment.ID, Version: payment.Version, Journal: &moved})
	withoutDetails := payment
	withoutDetails.Details = nil
	f.run(Command{Resource: "journals", Action: "update", ID: payment.ID, Version: payment.Version, Journal: &withoutDetails})
	payment = f.journal(payment.ID)
	moved = payment
	moved.BranchCode = "B2"
	f.deny(f.scope, Command{Resource: "journals", Action: "update", ID: payment.ID, Version: payment.Version, Journal: &moved})
	f.run(Command{Resource: "journals", Action: "delete", ID: payment.ID, Version: payment.Version})
	var orphan int
	if err := f.db.QueryRow(`SELECT COUNT(*) FROM gl_subledger_documents WHERE id IN('P1','P2')`).Scan(&orphan); err != nil {
		t.Fatal(err)
	}
	if orphan != 0 {
		t.Fatal("abandoned draft documents remain")
	}
	page, err := f.store.pg.SubledgerList(f.ctx, f.scope, "statements", "", 1, 50, "")
	if err != nil || page.Total != 2 {
		t.Fatal("bank evidence disappeared with importing draft", err)
	}
	payment.Identity = Identity{}
	payment.DocNo = "PAY-REENTERED"
	payment.Details = &JournalDetails{BankLines: []SubledgerBankLine{{LineNumber: 1, BankAccountCode: "BANK", Direction: 1}, {LineNumber: 2, BankAccountCode: "BANK", Direction: 1}}, StatementLines: []SubledgerStatementLine{{ID: "ALIAS", BankAccountCode: "BANK", SourceKey: "FILE-1", Date: "2026-01-11", Direction: 1, Amount: "0.10"}}, Matches: []SubledgerMatch{{ID: "RE-MATCH", StatementLineID: "ALIAS", LineNumber: 1, Amount: "0.10"}}}
	r := f.run(Command{Resource: "journals", Action: "create", Journal: &payment})
	newPayment := f.post(f.journal(r.ID))
	page, err = f.store.pg.SubledgerList(f.ctx, f.scope, "bank-lines", newPayment.ID+":1", 1, 50, "")
	if err != nil || page.Total != 1 {
		t.Fatal("composite bank line hydration failed", err)
	}
	page, err = f.store.pg.SubledgerList(f.ctx, f.scope, "statements", "S1", 1, 50, "")
	if err != nil || page.Total != 1 {
		t.Fatal("reimport duplicated immutable evidence", err)
	}
	var statement SubledgerStatementBalance
	if err = json.Unmarshal(page.Items[0], &statement); err != nil {
		t.Fatal(err)
	}
	assertSubledgerAmount(t, statement.Matched, "0.10")
}

// เอกสารลูกหนี้/เจ้าหนี้ในเซสชันระดับบริษัทต้องอยู่สาขาเดียวกับใบสำคัญ — เดิมรับสาขาที่พิมพ์มั่ว ๆ (adversarial review 2026-09-24)
func TestPostgresSubledgerDocumentBranchFollowsJournal(t *testing.T) {
	f := newPGIntegrityFixture(t)
	details := func(branch string) *JournalDetails {
		return &JournalDetails{
			Partners:  []SubledgerPartner{{Code: "CUST1", Name: "บริษัท ก่อสร้างไทย จำกัด", IsCustomer: true, IsActive: true}},
			Documents: []SubledgerDocument{{ID: "AR1", Ledger: "ar", PartnerCode: "CUST1", DocumentNo: "INV6901-001", Date: "2026-01-10", BranchCode: branch, Kind: 1, Side: 1, Amount: "100", Currency: "THB", ControlAccountCode: "1000"}},
		}
	}
	j := Journal{DocNo: "SV6901-001", Date: "2026-01-10", BookCode: "JV", FiscalYear: "2026", Description: "ขายวัสดุก่อสร้างเป็นเงินเชื่อ", Kind: "manual", BranchCode: "B1",
		Lines:   []Line{{AccountCode: "1000", Debit: "100", Credit: "0"}, {AccountCode: "4000", Debit: "0", Credit: "100"}},
		Details: details("ZZZ")}
	_, err := f.execute(f.scope, Command{Resource: "journals", Action: "create", Journal: &j})
	if user, ok := AsUserError(err); !ok || user.Code != "subledger_document_branch_mismatch" || user.Field != "branch_code" {
		t.Fatalf("made-up document branch: %v", err)
	}
	var count int
	if err = f.db.QueryRow(`SELECT COUNT(*) FROM gl_subledger_documents WHERE company='C'`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("rejected document stored: %d (%v)", count, err)
	}
	j.Details = details(" B1 ")
	r := f.run(Command{Resource: "journals", Action: "create", Journal: &j})
	var branch string
	if err = f.db.QueryRow(`SELECT branch_code FROM gl_subledger_documents WHERE company='C' AND id='AR1'`).Scan(&branch); err != nil || branch != "B1" || r.ID == "" {
		t.Fatalf("document branch stored = %q (%v)", branch, err)
	}
}
