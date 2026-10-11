//go:build integration

package generalledger

import (
	"strings"
	"testing"
)

// A posted AP debt reserved by an active payment preparation cannot be reversed from the GL screen while the preparation
// holds it; once the preparation is cancelled the reversal goes through (design ap-payment-prepare §4.1). Reconcile cannot
// lower a document's posted amount at all (completeDocuments), so reversal is the GL path that needs the guard.
func TestPostgresAPPrepareReservationGuardsGLDebt(t *testing.T) {
	f := newPGIntegrityFixture(t)
	account := Account{AccountCode: "2000", AccountType: "liability", NormalBalance: "credit", AllowPosting: true, IsActive: true, Names: []Name{{Code: "th", Name: "AP"}}}
	f.run(Command{Resource: "accounts", Action: "create", Account: &account})
	j := Journal{DocNo: "JV-AP1", Date: "2026-01-10", Description: "ซื้อวัตถุดิบเชื่อ", BookCode: "JV", FiscalYear: "2026", Kind: "manual", BranchCode: "B1",
		Lines: []Line{{AccountCode: "5000", Debit: "100", Credit: "0"}, {AccountCode: "2000", Debit: "0", Credit: "100"}},
		Details: &JournalDetails{Partners: []SubledgerPartner{{Code: "SUP", Name: "หจก.ไทยรุ่งเรืองค้าวัสดุ", IsSupplier: true, IsActive: true}},
			Documents:   []SubledgerDocument{{ID: "AP-D", Ledger: "ap", PartnerCode: "SUP", DocumentNo: "INV-6901", Date: "2026-01-10", Kind: 1, Side: 1, Amount: "100", ControlAccountCode: "2000"}},
			Allocations: []SubledgerAllocation{{ID: "AP-AD", Ledger: "ap", DocumentID: "AP-D", LineNumber: 2, Amount: "100"}}}}
	r := f.run(Command{Resource: "journals", Action: "create", Journal: &j})
	debt := f.post(f.journal(r.ID))
	if _, err := f.db.Exec(`INSERT INTO ap_payment_prepare (company,docno,paydate,partner_code,total_amount,total_home_amount,net_amount,created_at,created_by,modified_at,modified_by)
		VALUES ('C','PP-6901','2026-01-20','SUP',60,60,60,now(),'tester',now(),'tester');
		INSERT INTO ap_payment_prepare_lines (company,docno,linenumber,debt_document_id,document_no,document_date,debt_amount,billtype,invbalance,payamount)
		VALUES ('C','PP-6901',1,'AP-D','INV-6901','2026-01-10',100,4,100,60)`); err != nil {
		t.Fatal(err)
	}
	reserved := func(cmd Command) {
		t.Helper()
		_, err := f.execute(f.scope, cmd)
		user, ok := AsUserError(err)
		if !ok || user.Code != "ap_prepare_reserved" || len(user.Args) != 2 || user.Args[0] != "INV-6901" || user.Args[1] != "PP-6901" {
			t.Fatalf("error = %v, want ap_prepare_reserved for INV-6901/PP-6901", err)
		}
		if !strings.Contains(user.Message, "PP-6901") || !strings.Contains(user.Message, "ยกเลิกใบเตรียมจ่าย") {
			t.Fatalf("message must tell how to fix it, got %q", user.Message)
		}
	}
	reverse := Command{Resource: "journals", Action: "reverse", ID: debt.ID, Version: debt.Version, DocNo: "REV-AP1", Date: "2026-01-12", Reason: "บันทึกผิด"}
	reserved(reverse)
	// The guard rolled back the whole command: the voucher is still posted and the debt fully open.
	if got := f.journal(debt.ID).Status; got != "posted" {
		t.Fatalf("status after refused reversal = %q, want posted", got)
	}
	assertSubledgerAmount(t, subledgerDocumentAt(t, f, "AP-D", "").Posted, "100")
	if _, err := f.db.Exec(`UPDATE ap_payment_prepare SET iscancel=true, cancel_reason='เลื่อนจ่าย' WHERE company='C' AND docno='PP-6901'`); err != nil {
		t.Fatal(err)
	}
	f.run(reverse)
	assertSubledgerAmount(t, subledgerDocumentAt(t, f, "AP-D", "").Posted, "0")
}
