//go:build integration

package httpapi

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"smlcloudplatform/internal/doctype"
	gl "smlcloudplatform/internal/generalledger"
)

// prepareFixture is a seeded ledger with supplier S101 and one posted purchase of 1,070.00 (10 × 100 + VAT 7%).
type prepareFixture struct {
	t                *testing.T
	ctx              context.Context
	db               *sql.DB
	store            *Store
	scope            gl.Scope
	payDay           string
	purchase, pay    doctype.DocType
	purchaseNo, debt string
}

func newPrepareFixture(t *testing.T) (*prepareFixture, *gl.Postgres) {
	db, projection := erpTestDB(t)
	seedLedger(t, db, projection)
	if _, err := db.Exec(`UPDATE gl_subledger_partners SET payload = payload || '{"is_supplier":true,"is_active":true}'::jsonb WHERE company=$1 AND code='S101'`, testCompany); err != nil {
		t.Fatal(err)
	}
	f := &prepareFixture{t: t, ctx: context.Background(), db: db, store: NewStore(), scope: erpScope(), payDay: erpDocDay.Format("2006-01-02"),
		purchase: mustDocType(t, "purchase"), pay: mustDocType(t, "pay")}
	f.purchaseNo, f.debt = f.buy("10")
	return f, projection
}

func (f *prepareFixture) buy(qty string) (string, string) {
	f.t.Helper()
	bought, err := f.store.SaveDoc(f.ctx, f.db, f.scope, f.purchase, purchaseDoc("", qty, "100"), false)
	if err != nil {
		f.t.Fatalf("purchase: %v", err)
	}
	return bought.DocNo, "AP-" + bought.DocNo
}

func (f *prepareFixture) prepare(scope gl.Scope, docNo, debtID, amount string, version int64, editNo string) (*apPrepareDoc, error) {
	return f.store.SavePrepare(f.ctx, f.db, scope, apPrepareInput{DocNo: docNo, PayDate: f.payDay, PartnerCode: "S101", Version: version,
		Lines: []apPrepareLineInput{{DebtDocumentID: debtID, PayAmount: amount}}}, editNo)
}

// payOn saves a payment voucher dated day; prepareNo "" = a payment not pulled from a preparation.
func (f *prepareFixture) payOn(day time.Time, prepareNo, debtID, purchaseNo, amount string) (*ErpTransactionDoc, error) {
	return f.store.SaveDoc(f.ctx, f.db, f.scope, f.pay, &ErpTransactionDoc{DocDateTime: day.Format(time.RFC3339), CustCode: "S101", TotalAmount: amt(amount),
		DocRef: purchaseNo, Details: []ErpDetailItem{{LineNumber: 1, DebtDocumentID: debtID, DocRef: purchaseNo, PayBillNo: prepareNo, SumAmount: amt(amount)}}}, false)
}

func (f *prepareFixture) outstanding(scope gl.Scope, excludeDocNo string) []apOutstandingRow {
	f.t.Helper()
	rows, _, err := f.store.PrepareOutstanding(f.ctx, f.db, scope, "S101", f.payDay, "", excludeDocNo, 50, 0)
	if err != nil {
		f.t.Fatalf("outstanding: %v", err)
	}
	return rows
}

func wantUserError(t *testing.T, label string, err error, code string, args ...string) {
	t.Helper()
	user, ok := gl.AsUserError(err)
	if !ok || user.Code != code {
		t.Fatalf("%s = %v, want %s", label, err, code)
	}
	if len(args) > 0 && fmt.Sprint(user.Args) != fmt.Sprint(args) {
		t.Fatalf("%s args = %q, want %q", label, user.Args, args)
	}
}

// Payment preparation against real PostgreSQL (design ap-payment-prepare §6): partial PV from the preparation, D11 edit
// rules (never below paid, above paid ok), fully paid = locked, the reserved purchase guarded on the ERP side, and Champ
// cancel semantics (ConfirmDeleteData: only a fully paid voucher is locked; a partly paid one keeps what was paid).
func TestIntegrationPaymentPrepareLifecycle(t *testing.T) {
	f, _ := newPrepareFixture(t)
	scope, debt, payDay := f.scope, f.debt, erpDocDay
	prepareFor := func(docNo, debtID, amount string, version int64, editNo string) (*apPrepareDoc, error) {
		return f.prepare(scope, docNo, debtID, amount, version, editNo)
	}
	wantCode := func(label string, err error, code string) {
		t.Helper()
		wantUserError(t, label, err, code)
	}

	prep, err := prepareFor("PP6910-0001", debt, "800.00", 0, "")
	if err != nil {
		t.Fatalf("create preparation: %v", err)
	}
	// The reserved purchase cannot be edited or deleted on the ERP side while the preparation holds it.
	_, err = f.store.SaveDoc(f.ctx, f.db, scope, f.purchase, purchaseDoc(f.purchaseNo, "5", "100"), true)
	wantCode("edit of reserved purchase", err, "ap_prepare_reserved")
	wantCode("delete of reserved purchase", f.store.DeleteDoc(f.ctx, f.db, scope, f.purchase, f.purchaseNo), "ap_prepare_reserved")

	if _, err = f.payOn(payDay, prep.DocNo, debt, f.purchaseNo, "400.00"); err != nil {
		t.Fatalf("partial payment from preparation: %v", err)
	}
	if n := scalarString(t, f.db, `SELECT count(*)::text FROM gl_subledger_settlements WHERE company=$1 AND debt_id=$2 AND payload->>'prepare_docno'=$3 AND reversed_at IS NULL`, testCompany, debt, prep.DocNo); n != "1" {
		t.Fatalf("settlements tagged with the preparation = %s, want 1", n)
	}
	requireDecimal(t, "debt after partial payment", subledgerOutstanding(t, f.db, debt), "670")
	// Design §4.5 item 5: a partial payment never closes the debt document it references (closed = remaining 0 in the ledger).
	if got := scalarString(t, f.db, `SELECT isclosed::text FROM doc WHERE businesscode=$1 AND docno=$2`, testCompany, f.purchaseNo); got != "false" {
		t.Fatalf("purchase isclosed after a partial payment = %s, want false", got)
	}

	// D11: partially paid stays editable, but never below what was paid.
	_, err = prepareFor(prep.DocNo, debt, "399.99", prep.Version, prep.DocNo)
	wantCode("edit below paid", err, "ap_prepare_below_paid")
	if prep, err = prepareFor(prep.DocNo, debt, "500.00", prep.Version, prep.DocNo); err != nil {
		t.Fatalf("edit above paid: %v", err)
	}
	requireDecimal(t, "prepared amount after edit", prep.Lines[0].PayAmount, "500")
	requireDecimal(t, "used amount after edit", prep.Lines[0].UsedAmount, "400")

	// Paying the rest of the preparation locks it for edit and cancel (Champ BillStatus 1).
	if _, err = f.payOn(payDay, prep.DocNo, debt, f.purchaseNo, "100.00"); err != nil {
		t.Fatalf("final payment from preparation: %v", err)
	}
	_, err = f.payOn(payDay, prep.DocNo, debt, f.purchaseNo, "0.01")
	if userErrorCode(err) == "" {
		t.Fatalf("payment beyond the preparation = %v, want a refusal", err)
	}
	_, err = prepareFor(prep.DocNo, debt, "600.00", prep.Version, prep.DocNo)
	wantCode("edit of fully paid preparation", err, "ap_prepare_paid_locked")
	wantCode("cancel of fully paid preparation", f.store.CancelPrepare(f.ctx, f.db, scope, prep.DocNo, "เลื่อนจ่าย", prep.Version), "ap_prepare_paid_locked")
	requireDecimal(t, "debt after preparation paid", subledgerOutstanding(t, f.db, debt), "570")

	// Cancel of a partly paid preparation (Champ BillStatus 0): reason required, version checked, the payment stays
	// settled and only the unpaid rest of the reservation is released.
	otherNo, otherDebt := f.buy("2")
	partly, err := prepareFor("PP6910-0002", otherDebt, "214.00", 0, "")
	if err != nil {
		t.Fatalf("second preparation: %v", err)
	}
	wantCode("cancel without reason", f.store.CancelPrepare(f.ctx, f.db, scope, partly.DocNo, " ", partly.Version), "ap_prepare_cancel_reason_required")
	wantCode("cancel with stale version", f.store.CancelPrepare(f.ctx, f.db, scope, partly.DocNo, "เลื่อนจ่าย", partly.Version-1), "ap_prepare_version_conflict")
	if _, err = f.payOn(payDay, partly.DocNo, otherDebt, otherNo, "100.00"); err != nil {
		t.Fatalf("partial payment from second preparation: %v", err)
	}
	if err = f.store.CancelPrepare(f.ctx, f.db, scope, partly.DocNo, "เลื่อนจ่าย", partly.Version); err != nil {
		t.Fatalf("cancel partly paid preparation: %v", err)
	}
	if got := scalarString(t, f.db, `SELECT iscancel::text||'|'||cancel_reason FROM ap_payment_prepare WHERE company=$1 AND docno=$2`, testCompany, partly.DocNo); got != "true|เลื่อนจ่าย" {
		t.Fatalf("cancelled preparation row = %q", got)
	}
	if n := scalarString(t, f.db, `SELECT count(*)::text FROM gl_subledger_settlements WHERE company=$1 AND payload->>'prepare_docno'=$2 AND reversed_at IS NULL`, testCompany, partly.DocNo); n != "1" {
		t.Fatalf("settlements of the cancelled preparation = %s, want 1 (paid part stays)", n)
	}
	requireDecimal(t, "second debt after cancel", subledgerOutstanding(t, f.db, otherDebt), "114")
	_, err = prepareFor("PP6910-0003", otherDebt, "114.01", 0, "")
	wantCode("new preparation above the released rest", err, "ap_prepare_over_balance")
	if _, err = prepareFor("PP6910-0003", otherDebt, "114.00", 0, ""); err != nil {
		t.Fatalf("new preparation of the released rest: %v", err)
	}
	_, err = prepareFor(partly.DocNo, otherDebt, "100.00", partly.Version+1, partly.DocNo)
	wantCode("edit of cancelled preparation", err, "ap_prepare_cancelled_locked")
	_, err = f.payOn(payDay, partly.DocNo, otherDebt, otherNo, "10.00")
	wantCode("payment from a cancelled preparation", err, "ap_prepare_cancelled_locked")
	// Champ PaymentProcessStatement (bcprocess.cpp:490-515) only updates the statement rows, so a payment that already pulled
	// from the preparation stays editable after the preparation is cancelled; a new pull from it is still refused.
	paid := scalarString(t, f.db, `SELECT docno FROM doc WHERE businesscode=$1 AND transflag=$2 AND docno IN (SELECT docno FROM docdetail WHERE businesscode=$1 AND paybillno=$3) ORDER BY docno LIMIT 1`, testCompany, f.pay.Flag, partly.DocNo)
	edited := &ErpTransactionDoc{DocNo: paid, DocDateTime: payDay.Format(time.RFC3339), CustCode: "S101", TotalAmount: amt("90.00"), DocRef: otherNo,
		Details: []ErpDetailItem{{LineNumber: 1, DebtDocumentID: otherDebt, DocRef: otherNo, PayBillNo: partly.DocNo, SumAmount: amt("90.00")}}}
	if _, err = f.store.SaveDoc(f.ctx, f.db, scope, f.pay, edited, true); err != nil {
		t.Fatalf("edit of a payment whose preparation was cancelled: %v", err)
	}
	requireDecimal(t, "second debt after editing the payment", subledgerOutstanding(t, f.db, otherDebt), "124")
	edited.Details = append(edited.Details, ErpDetailItem{LineNumber: 2, DebtDocumentID: debt, DocRef: f.purchaseNo, PayBillNo: partly.DocNo, SumAmount: amt("1.00")})
	edited.TotalAmount = amt("91.00")
	_, err = f.store.SaveDoc(f.ctx, f.db, scope, f.pay, edited, true)
	wantCode("new line from a cancelled preparation on edit", err, "ap_prepare_cancelled_locked")
	if n := scalarString(t, f.db, `SELECT count(*)::text FROM ap_payment_prepare_events WHERE company=$1`, testCompany); n != "5" {
		t.Fatalf("preparation events = %s, want 5 (create, update, create, cancel, create)", n)
	}
}

// Root cause A: the picker, the save and the edit limit read one balance — as-of = later of the date and today, debts
// of the preparation's own branch, and limit = posted debt − settlements and reservations of everything else.
func TestIntegrationPrepareLimitOneDefinition(t *testing.T) {
	f, projection := newPrepareFixture(t)
	today, err := dayStart(businessDay(time.Now()))
	if err != nil {
		t.Fatal(err)
	}
	later := time.Date(today.Year(), today.Month(), today.Day()+20, 9, 0, 0, 0, time.UTC)
	if later.Year() != erpDocDay.Year() {
		seedFiscalYear(t, projection, later.Year())
	}

	// (3) The purchase was posted today but is dated earlier: the picker as of the pay date lists it like the save does.
	rows := f.outstanding(f.scope, "")
	if len(rows) != 1 || rows[0].DebtDocumentID != f.debt {
		t.Fatalf("picker as of the pay date = %+v, want the purchase posted today", rows)
	}
	requireDecimal(t, "available in picker", rows[0].AvailableAmount, "1070")
	prep, err := f.prepare(f.scope, "PP6910-0101", f.debt, "1070.00", 0, "")
	if err != nil {
		t.Fatalf("create preparation: %v", err)
	}

	// (8) A user of another branch edits the preparation: the debts are read in the preparation's branch.
	elsewhere := f.scope
	elsewhere.Branch = "00009"
	rows = f.outstanding(elsewhere, prep.DocNo)
	if len(rows) != 1 || rows[0].DebtDocumentID != f.debt {
		t.Fatalf("picker of the preparation from another branch = %+v, want its debt", rows)
	}
	if prep, err = f.prepare(elsewhere, prep.DocNo, f.debt, "1070.00", prep.Version, prep.DocNo); err != nil {
		t.Fatalf("edit from another branch: %v", err)
	}

	// (1) A payment dated after today settles 500.00 from the preparation; the edit limit stays the posted debt.
	if _, err = f.payOn(later, prep.DocNo, f.debt, f.purchaseNo, "500.00"); err != nil {
		t.Fatalf("future-dated payment from preparation: %v", err)
	}
	_, err = f.prepare(f.scope, prep.DocNo, f.debt, "1570.00", prep.Version, prep.DocNo)
	wantUserError(t, "edit above the posted debt", err, "ap_prepare_over_balance", f.purchaseNo, "1070.00")
	if _, err = f.prepare(f.scope, prep.DocNo, f.debt, "1070.00", prep.Version, prep.DocNo); err != nil {
		t.Fatalf("edit at the posted debt: %v", err)
	}
}

// Root cause B: amounts leave the API as scale-2 strings that the save accepts as they are, and a limit that other
// reservations pushed below zero is shown as 0.00.
func TestIntegrationPrepareAmountsScaleTwo(t *testing.T) {
	f, _ := newPrepareFixture(t)
	rows := f.outstanding(f.scope, "")
	if len(rows) != 1 {
		t.Fatalf("picker rows = %+v, want 1", rows)
	}
	if got := [4]string{rows[0].DebtAmount, rows[0].RemainingAmount, rows[0].ReservedAmount, rows[0].AvailableAmount}; got != [4]string{"1070.00", "1070.00", "0.00", "1070.00"} {
		t.Fatalf("picker amounts = %q, want scale-2 strings", got)
	}
	first, err := f.prepare(f.scope, "PP6910-0201", f.debt, "500.00", 0, "")
	if err != nil {
		t.Fatalf("first preparation: %v", err)
	}
	rows = f.outstanding(f.scope, "")
	if len(rows) != 1 || rows[0].ReservedAmount != "500.00" || rows[0].AvailableAmount != "570.00" {
		t.Fatalf("picker after first preparation = %+v", rows)
	}
	if _, err = f.prepare(f.scope, "PP6910-0202", f.debt, rows[0].AvailableAmount, 0, ""); err != nil {
		t.Fatalf("save the picked amount as returned: %v", err)
	}
	// The two preparations hold the whole debt: a payment outside them may not pay what they reserved (Champ hides a bill
	// whose PayBillAmount reached its balance from the payment picker, apfrmappayment1view.cpp:1612-1617).
	_, err = f.payOn(erpDocDay, "", f.debt, f.purchaseNo, "600.00")
	wantUserError(t, "payment outside preparations over what they hold", err, "erp_settle_exceeds_remaining", f.purchaseNo, "0.00")
	_, err = f.prepare(f.scope, first.DocNo, f.debt, "500.01", first.Version, first.DocNo)
	wantUserError(t, "edit above what the other preparation leaves", err, "ap_prepare_over_balance", f.purchaseNo, "500.00")
}

// A payment dated after today that already settled the debt is seen by a preparation dated today: it cannot reserve
// what is no longer owed (the picker shows 0.00 available and the save refuses).
func TestIntegrationPrepareSeesFutureDatedPayment(t *testing.T) {
	f, projection := newPrepareFixture(t)
	today, err := dayStart(businessDay(time.Now()))
	if err != nil {
		t.Fatal(err)
	}
	later := time.Date(today.Year(), today.Month(), today.Day()+20, 9, 0, 0, 0, time.UTC)
	if later.Year() != erpDocDay.Year() {
		seedFiscalYear(t, projection, later.Year())
	}
	if _, err = f.payOn(later, "", f.debt, f.purchaseNo, "1070.00"); err != nil {
		t.Fatalf("future-dated payment: %v", err)
	}
	if rows := f.outstanding(f.scope, ""); len(rows) == 1 && rows[0].AvailableAmount != "0.00" {
		t.Fatalf("picker available after a future-dated payment = %s, want 0.00", rows[0].AvailableAmount)
	}
	_, err = f.prepare(f.scope, "PP6910-0301", f.debt, "1070.00", 0, "")
	wantUserError(t, "preparation of a debt paid by a future-dated payment", err, "ap_prepare_over_balance", f.purchaseNo, "0.00")
}

// seedFiscalYear opens the fiscal year of a payment dated past the year seedLedger opens.
func seedFiscalYear(t *testing.T, p *gl.Postgres, year int) {
	t.Helper()
	y := gl.FiscalYear{Identity: erpIdentity(fmt.Sprintf("year-%d", year)), Code: fmt.Sprint(year), StartDate: fmt.Sprintf("%d-01-01", year),
		EndDate: fmt.Sprintf("%d-12-31", year), IsActive: true, Scale: 2, RetainedEarningsAccount: "3000"}
	event := gl.Event{ID: "erp-seed-year", HoldingCode: erpHolding, BusinessCode: testCompany, Sequence: 2, RequestID: "erp-seed-year",
		Actor: "integration-test", Action: "test", OccurredAt: time.Date(2026, 1, 1, 0, 0, 2, 0, time.UTC), Changes: []gl.Change{erpChange(t, "fiscal-years", y.ID, y.Code, y)}}
	if err := p.Project(context.Background(), event); err != nil {
		t.Fatal(err)
	}
}
