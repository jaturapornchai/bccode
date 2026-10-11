package cheque

import (
	"reflect"
	"strings"
	"testing"

	"github.com/shopspring/decimal"
)

const (
	chequeA = "11111111-1111-4111-8111-111111111111"
	chequeB = "22222222-2222-4222-8222-222222222222"
)

func intRef(v int) *int { return &v }

func TestNormalizeDocRejectsInvalidInputInScreenOrder(t *testing.T) {
	line := []DocLineInput{{ChequeID: chequeA}}
	for _, tc := range []struct {
		name  string
		in    DocInput
		key   string
		field string
	}{
		{"unknown transflag", DocInput{Flag: 44, DocDate: "2026-10-11", Lines: line}, "cheque_err_transflag_invalid", "transflag"},
		{"cancel type outside 423", DocInput{Flag: FlagReceivedClear, CancelType: intRef(0), DocDate: "2026-10-11", BookCode: "SCB01", Lines: line}, "cheque_err_cancel_type_invalid", "cancel_type"},
		{"423 without cancel type", DocInput{Flag: FlagReceivedCancel, DocDate: "2026-10-11", Lines: line}, "cheque_err_cancel_type_invalid", "cancel_type"},
		{"423 cancel type out of range", DocInput{Flag: FlagReceivedCancel, CancelType: intRef(4), DocDate: "2026-10-11", Lines: line}, "cheque_err_cancel_type_invalid", "cancel_type"},
		{"doc date missing", DocInput{Flag: FlagDeposit, BookCode: "SCB01", Lines: line}, "cheque_err_required_field", "doc_date"},
		{"doc date not a date", DocInput{Flag: FlagDeposit, DocDate: "2026-02-30", BookCode: "SCB01", Lines: line}, "cheque_err_date_invalid", "doc_date"},
		{"deposit without book", DocInput{Flag: FlagDeposit, DocDate: "2026-10-11", Lines: line}, "cheque_err_book_required", "book_code"},
		{"redeposit without book", DocInput{Flag: FlagRedeposit, DocDate: "2026-10-11", Lines: line}, "cheque_err_book_required", "book_code"},
		{"cancel passed issued cheque without book", DocInput{Flag: FlagIssuedCancel, CancelPassed: true, DocDate: "2026-10-11", Lines: line}, "cheque_err_book_required", "book_code"},
		{"cancel on-hand issued cheque without book", DocInput{Flag: FlagIssuedCancel, DocDate: "2026-10-11", Lines: line}, "cheque_err_book_required", "book_code"},
		{"clear received cheque without book", DocInput{Flag: FlagReceivedClear, DocDate: "2026-10-11", Lines: line}, "cheque_err_book_required", "book_code"},
		{"return received cheque without book", DocInput{Flag: FlagReceivedReturn, DocDate: "2026-10-11", Lines: line}, "cheque_err_book_required", "book_code"},
		{"clear issued cheque without book", DocInput{Flag: FlagIssuedClear, DocDate: "2026-10-11", Lines: line}, "cheque_err_book_required", "book_code"},
		{"no lines", DocInput{Flag: FlagReceivedReturn, DocDate: "2026-10-11", BookCode: "SCB01"}, "cheque_err_no_lines", "lines"},
		{"line is not a uuid", DocInput{Flag: FlagReceivedReturn, DocDate: "2026-10-11", BookCode: "SCB01", Lines: []DocLineInput{{ChequeID: "1; DROP"}}}, "cheque_err_not_found", "lines[0].cheque_id"},
		{"duplicate cheque", DocInput{Flag: FlagReceivedReturn, DocDate: "2026-10-11", BookCode: "SCB01", Lines: []DocLineInput{{ChequeID: chequeA}, {ChequeID: strings.ToUpper(chequeA)}}}, "cheque_err_duplicate_line", "lines[1].cheque_id"},
		{"negative fee", DocInput{Flag: FlagReceivedReturn, DocDate: "2026-10-11", BookCode: "SCB01", Lines: []DocLineInput{{ChequeID: chequeA, FeeAmount: "-5"}}}, "cheque_err_fee_invalid", "lines[0].fee_amount"},
		{"fee with 3 decimals", DocInput{Flag: FlagReceivedReturn, DocDate: "2026-10-11", BookCode: "SCB01", Lines: []DocLineInput{{ChequeID: chequeA, FeeAmount: "1.005"}}}, "cheque_err_fee_invalid", "lines[0].fee_amount"},
		{"deposit has no fee column", DocInput{Flag: FlagDeposit, DocDate: "2026-10-11", BookCode: "SCB01", Lines: []DocLineInput{{ChequeID: chequeA, FeeAmount: "10.00"}}}, "cheque_err_fee_invalid", "lines[0].fee_amount"},
		{"totals not a decimal", DocInput{Flag: FlagReceivedReturn, DocDate: "2026-10-11", BookCode: "SCB01", Lines: line, Totals: &DocTotals{Total: "1e3", Fee: "0", Net: "1000"}}, "cheque_err_totals_mismatch", "totals"},
		{"description with control character", DocInput{Flag: FlagReceivedReturn, DocDate: "2026-10-11", BookCode: "SCB01", Description: "ค่า\x00", Lines: line}, "cheque_err_text_invalid", "description"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := normalizeDoc(tc.in)
			if e := wantErrorKey(t, err, tc.key); e.Field != tc.field {
				t.Fatalf("field = %q, want %q", e.Field, tc.field)
			}
		})
	}
}

func TestNormalizeDocKeepsExactDecimalsAndDropsBookOfOnHandCancel(t *testing.T) {
	v, err := normalizeDoc(DocInput{
		Flag: FlagReceivedClear, DocNo: " CQIP6910-0001 ", DocDate: "2026-10-11", BookCode: "SCB01",
		Lines:  []DocLineInput{{ChequeID: strings.ToUpper(chequeA), FeeAmount: "0.10"}, {ChequeID: chequeB}},
		Totals: &DocTotals{Total: "1000.30", Fee: "0.10", Net: "1000.20"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if v.docNo != "CQIP6910-0001" || v.book != "SCB01" || len(v.lines) != 2 || v.lines[0].chequeID != chequeA {
		t.Fatalf("values = %+v", v)
	}
	if !v.lines[0].fee.Equal(decimal.RequireFromString("0.1")) || !v.lines[1].fee.IsZero() {
		t.Fatalf("fees = %s, %s", v.lines[0].fee, v.lines[1].fee)
	}
	if v.totals == nil || v.totals[0].String() != "1000.3" || v.totals[2].String() != "1000.2" {
		t.Fatalf("totals = %v", v.totals)
	}

	// ยกเลิกเช็คในมือ (423 ชนิด 0) ไม่ใช้สมุด แม้หน้าจอส่งมา
	v, err = normalizeDoc(DocInput{Flag: FlagReceivedCancel, CancelType: intRef(CancelOnHand), DocDate: "2026-10-11", BookCode: "SCB01", Lines: []DocLineInput{{ChequeID: chequeA}}})
	if err != nil {
		t.Fatal(err)
	}
	if v.book != "" || !v.cancelTypeSet || v.cancelType != CancelOnHand {
		t.Fatalf("on-hand cancel = %+v", v)
	}
	// cancel_passed มีผลเฉพาะ 426
	v, err = normalizeDoc(DocInput{Flag: FlagIssuedClear, CancelPassed: true, DocDate: "2026-10-11", BookCode: "SCB01", Lines: []DocLineInput{{ChequeID: chequeA}}})
	if err != nil || v.cancelPassd {
		t.Fatalf("issued clear cancel_passed = %v, err %v", v.cancelPassd, err)
	}
}

func TestSelectableStatusesFollowChampScreens(t *testing.T) {
	for _, tc := range []struct {
		flag, cancelType int
		passed           bool
		want             []int
	}{
		{FlagDeposit, -1, false, []int{InOnHand}},
		{FlagReceivedClear, -1, false, []int{InOnHand, InDeposited}},
		{FlagDiscount, -1, false, []int{InOnHand, InDeposited}},
		{FlagReceivedReturn, -1, false, []int{InCleared, InRedeposited}},
		{FlagRedeposit, -1, false, []int{InReturned}},
		{FlagReceivedCancel, CancelOnHand, false, []int{InOnHand}},
		{FlagReceivedCancel, CancelCleared, false, []int{InCleared}},
		{FlagReceivedCancel, CancelDeposited, false, []int{InDeposited}},
		{FlagReceivedCancel, CancelRedeposited, false, []int{InRedeposited}},
		{FlagReceivedCancel, -1, false, nil},
		{FlagIssuedClear, -1, false, []int{OutOnHand}},
		{FlagIssuedCancel, -1, false, []int{OutOnHand}},
		{FlagIssuedCancel, -1, true, []int{OutCleared}},
	} {
		if got := selectableStatuses(tc.flag, tc.cancelType, tc.passed); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("selectableStatuses(%d,%d,%v) = %v, want %v", tc.flag, tc.cancelType, tc.passed, got, tc.want)
		}
	}
}

func TestCheckSelectableOrder(t *testing.T) {
	onHand := lineCheque{id: chequeA, chequeNo: "10234567", docDate: "2026-10-01", dueDate: "2026-10-15", chequeType: TypeReceived, status: InOnHand}
	deposited := onHand
	deposited.status, deposited.book = InDeposited, "SCB01"
	returned := onHand
	returned.status, returned.book = InReturned, "KBANK02"
	returnedCleared := onHand
	returnedCleared.status, returnedCleared.book = InCleared, "KBANK02"
	issued := lineCheque{id: chequeB, chequeNo: "50001234", docDate: "2026-10-01", dueDate: "2026-10-20", chequeType: TypeIssued, status: OutOnHand, book: "SCB01"}

	for _, tc := range []struct {
		name      string
		flag, ct  int
		date, bk  string
		c         lineCheque
		wantKey   string
		wantArgs  []string
		wantClean bool
	}{
		{"wrong status", FlagRedeposit, -1, "2026-10-20", "KBANK02", onHand, "cheque_err_status_not_selectable", []string{"10234567"}, false},
		{"issued cheque on a received doc", FlagReceivedClear, -1, "2026-10-20", "SCB01", issued, "cheque_err_status_not_selectable", []string{"50001234"}, false},
		{"clear before due date", FlagReceivedClear, -1, "2026-10-14", "SCB01", deposited, "cheque_err_not_due", []string{"10234567", "15/10/2569"}, false},
		{"clear on due date", FlagReceivedClear, -1, "2026-10-15", "SCB01", deposited, "", nil, true},
		{"deposit before received", FlagDeposit, -1, "2026-09-30", "SCB01", onHand, "cheque_err_received_after_doc", []string{"10234567", "01/10/2569"}, false},
		{"deposit to any book", FlagDeposit, -1, "2026-10-01", "SCB01", onHand, "", nil, true},
		{"redeposit into another book", FlagRedeposit, -1, "2026-10-20", "SCB01", returned, "cheque_err_book_mismatch", []string{"10234567", "SCB01"}, false},
		{"redeposit same book", FlagRedeposit, -1, "2026-10-20", "KBANK02", returned, "", nil, true},
		{"clear a cheque of another book", FlagReceivedClear, -1, "2026-10-20", "KBANK02", deposited, "cheque_err_book_mismatch", []string{"10234567", "KBANK02"}, false},
		{"return a cheque of another book", FlagReceivedReturn, -1, "2026-10-20", "SCB01", returnedCleared, "cheque_err_book_mismatch", []string{"10234567", "SCB01"}, false},
		{"discount without book takes any book", FlagDiscount, -1, "2026-10-20", "", deposited, "", nil, true},
		{"discount filters by book when given", FlagDiscount, -1, "2026-10-20", "KBANK02", deposited, "cheque_err_book_mismatch", []string{"10234567", "KBANK02"}, false},
		{"issued clear of another book", FlagIssuedClear, -1, "2026-10-20", "KBANK02", issued, "cheque_err_book_mismatch", []string{"50001234", "KBANK02"}, false},
		{"issued clear checks due date", FlagIssuedClear, -1, "2026-10-19", "SCB01", issued, "cheque_err_not_due", []string{"50001234", "20/10/2569"}, false},
		{"issued cancel ignores dates", FlagIssuedCancel, -1, "2026-09-01", "", issued, "", nil, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := checkSelectable(tc.flag, tc.ct, false, tc.date, tc.bk, tc.c, "lines[0].cheque_id")
			if tc.wantClean {
				if err != nil {
					t.Fatalf("err = %v", err)
				}
				return
			}
			e := wantErrorKey(t, err, tc.wantKey)
			if !reflect.DeepEqual(e.Args, tc.wantArgs) || e.Field != "lines[0].cheque_id" {
				t.Fatalf("args = %v field %q, want %v", e.Args, e.Field, tc.wantArgs)
			}
		})
	}
}

func TestBookAfterAndExpectedBook(t *testing.T) {
	if got := bookAfter(FlagDeposit, "SCB01", ""); got != "SCB01" {
		t.Fatalf("deposit book = %q", got)
	}
	if got := bookAfter(FlagReceivedCancel, "", "SCB01"); got != "SCB01" {
		t.Fatalf("on-hand cancel keeps book, got %q", got)
	}
	if got := bookAfter(FlagIssuedClear, "KBANK02", "SCB01"); got != "SCB01" {
		t.Fatalf("issued cheque keeps its own book, got %q", got)
	}
	for _, tc := range []struct {
		chequeType, stored, derived int
		current, eventBook, want    string
	}{
		{TypeReceived, InOnHand, InOnHand, "SCB01", "", "SCB01"}, // เช็คในมือที่ระบุสมุดตอนบันทึก คงสมุดไว้
		{TypeReceived, InDeposited, InOnHand, "SCB01", "", ""},   // ย้อนกลับเป็นในมือ ล้างสมุดแบบ Champ
		{TypeReceived, InDeposited, InDeposited, "SCB01", "KBANK02", "SCB01"},
		{TypeReceived, InDeposited, InDeposited, "", "KBANK02", "KBANK02"},
		{TypeIssued, OutOnHand, OutOnHand, "SCB01", "", "SCB01"},
	} {
		if got := expectedBook(tc.chequeType, tc.stored, tc.derived, tc.current, tc.eventBook); got != tc.want {
			t.Errorf("expectedBook(%d,%d,%d,%q,%q) = %q, want %q", tc.chequeType, tc.stored, tc.derived, tc.current, tc.eventBook, got, tc.want)
		}
	}
}

func TestDocRegistryCoversEightFlags(t *testing.T) {
	if got := DocFlags(); !reflect.DeepEqual(got, []int{419, 420, 421, 422, 423, 424, 425, 426}) {
		t.Fatalf("DocFlags = %v", got)
	}
	want := []string{"chequereceivedclear", "chequereceivedreturn", "chequeredeposit", "chequereceivedcancel", "chequediscount", "chequeissuedclear", "chequeissuedcancel"}
	if got := MappingSourceDocs(); !reflect.DeepEqual(got, want) {
		t.Fatalf("MappingSourceDocs = %v (deposit 419 posts nothing)", got)
	}
	if ct, ok := DocChequeType(FlagIssuedCancel); !ok || ct != TypeIssued {
		t.Fatalf("DocChequeType(426) = %d, %v", ct, ok)
	}
	if _, ok := DocChequeType(400); ok {
		t.Fatal("400 is not a cheque status document")
	}
}

// OwnsDocFlag = ทะเบียนเช็ค 2 ชนิด + เอกสารสถานะ 8 ชนิด (รวม 10) และไม่รวมเอกสารอื่นที่เลขใกล้กัน
func TestOwnsDocFlagCoversExactlyTheTenChequeDocTypes(t *testing.T) {
	for _, flag := range []int{FlagRegisterReceived, FlagRegisterIssued, 419, 420, 421, 422, 423, 424, 425, 426} {
		if !OwnsDocFlag(flag) {
			t.Errorf("OwnsDocFlag(%d) = false", flag)
		}
	}
	for _, flag := range []int{0, 60, 62, 417, 427, 44} {
		if OwnsDocFlag(flag) {
			t.Errorf("OwnsDocFlag(%d) = true", flag)
		}
	}
}

func TestNewBookNoGLAccountErrorNamesTheBook(t *testing.T) {
	e := wantErrorKey(t, NewBookNoGLAccountError(""), "cheque_err_book_no_gl_account")
	if !reflect.DeepEqual(e.Args, []string{"-"}) || e.Status != 400 {
		t.Fatalf("error = %+v", e)
	}
	e = wantErrorKey(t, NewBookNoGLAccountError("SCB01"), "cheque_err_book_no_gl_account")
	if e.Args[0] != "SCB01" {
		t.Fatalf("args = %v", e.Args)
	}
}
