package cheque

import (
	"strings"
	"testing"
)

func validRegisterInput(chequeType int) RegisterInput {
	in := RegisterInput{
		ChequeType: chequeType, DocDate: "2026-10-01", DueDate: "2026-10-15", PartnerCode: "AR-0012",
		ChequeNo: "10234567", BankCode: "BBL", BankBranchCode: "0123", Amount: "15250.75",
		Description: "รับชำระค่าเหล็กเส้น", SaleCode: "S01",
	}
	if chequeType == TypeIssued {
		in.PartnerCode, in.BookCode = "AP-0007", "SCB01"
	}
	return in
}

func TestParseMoneyAcceptsOnlyPlainDecimalStrings(t *testing.T) {
	for value, want := range map[string]string{"0": "0", "15250.75": "15250.75", " 1.5 ": "1.5", "9999999999999999.99": "9999999999999999.99"} {
		got, ok := parseMoney(value)
		if !ok || got.String() != want {
			t.Errorf("parseMoney(%q) = %s, %v; want %s", value, got, ok, want)
		}
	}
	for _, value := range []string{"", "-1", "+1", "1,000", "1e3", "1.005", ".5", "NaN", "0x10", "10000000000000000"} {
		if _, ok := parseMoney(value); ok {
			t.Errorf("parseMoney(%q) accepted", value)
		}
	}
}

func TestNormalizeRegisterRejectsInvalidInput(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*RegisterInput)
		key    string
		field  string
	}{
		{"unknown type", func(in *RegisterInput) { in.ChequeType = 2 }, "cheque_err_type_invalid", "cheque_type"},
		{"missing due date", func(in *RegisterInput) { in.DueDate = " " }, "cheque_err_required_field", "due_date"},
		{"bad doc date", func(in *RegisterInput) { in.DocDate = "01/10/2569" }, "cheque_err_date_invalid", "doc_date"},
		{"missing partner", func(in *RegisterInput) { in.PartnerCode = "" }, "cheque_err_required_field", "partner_code"},
		{"cheque number too long", func(in *RegisterInput) { in.ChequeNo = strings.Repeat("9", 21) }, "cheque_err_too_long", "cheque_no"},
		{"zero amount", func(in *RegisterInput) { in.Amount = "0.00" }, "cheque_err_amount_invalid", "amount"},
		{"float-like amount", func(in *RegisterInput) { in.Amount = "1.5e3" }, "cheque_err_amount_invalid", "amount"},
		{"foreign currency", func(in *RegisterInput) { in.CurrencyCode = "usd" }, "cheque_err_currency_not_supported", "currency_code"},
		{"rate other than 1", func(in *RegisterInput) { in.ExchangeRate = "35.25" }, "cheque_err_currency_not_supported", "exchange_rate"},
		{"rate not a number", func(in *RegisterInput) { in.ExchangeRate = "abc" }, "cheque_err_rate_invalid", "exchange_rate"},
		{"control character", func(in *RegisterInput) { in.Description = "ชำระ​\x07" }, "cheque_err_text_invalid", "description"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in := validRegisterInput(TypeReceived)
			tc.mutate(&in)
			_, err := normalizeRegister(in)
			if e := wantErrorKey(t, err, tc.key); e.Field != tc.field {
				t.Fatalf("field = %q, want %q", e.Field, tc.field)
			}
		})
	}

	issued := validRegisterInput(TypeIssued)
	issued.BookCode = ""
	_, err := normalizeRegister(issued)
	wantErrorKey(t, err, "cheque_err_book_required")

	issued = validRegisterInput(TypeIssued)
	issued.DocDate, issued.DueDate = "2026-10-20", "2026-10-15"
	_, err = normalizeRegister(issued)
	wantErrorKey(t, err, "cheque_err_due_before_issue")

	// เช็ครับที่ถึงกำหนดแล้ว (ลงวันที่ก่อนวันรับ) บันทึกได้
	received := validRegisterInput(TypeReceived)
	received.DocDate, received.DueDate = "2026-10-20", "2026-10-15"
	if _, err = normalizeRegister(received); err != nil {
		t.Fatalf("received post-dated: %v", err)
	}
}

func TestNormalizeRegisterComputesHomeAmountAndDropsIssuedSaleCode(t *testing.T) {
	v, err := normalizeRegister(validRegisterInput(TypeReceived))
	if err != nil {
		t.Fatal(err)
	}
	if v.amount.String() != "15250.75" || v.rate.String() != "1" || money2(v.home) != "15250.75" || v.sale != "S01" {
		t.Fatalf("received values = %+v", v)
	}
	in := validRegisterInput(TypeIssued)
	in.CurrencyCode, in.ExchangeRate = "thb", "1.0000"
	if v, err = normalizeRegister(in); err != nil {
		t.Fatal(err)
	}
	if v.sale != "" || v.book != "SCB01" {
		t.Fatalf("issued values = %+v", v)
	}
}

func TestRegisterStatusAndBookRules(t *testing.T) {
	for _, tc := range []struct {
		chequeType, status int
		creating, want     bool
	}{
		{TypeReceived, InOnHand, true, true},
		{TypeReceived, InRedeposited, true, true},
		{TypeReceived, 7, false, false},
		{TypeReceived, -1, false, false},
		{TypeIssued, OutOnHand, true, true},
		{TypeIssued, OutPrepared, true, true},
		{TypeIssued, OutCleared, true, false},
		{TypeIssued, OutCleared, false, true},
		{TypeIssued, 4, false, false},
	} {
		if got := validRegisterStatus(tc.chequeType, tc.status, tc.creating); got != tc.want {
			t.Errorf("validRegisterStatus(%d,%d,%v) = %v", tc.chequeType, tc.status, tc.creating, got)
		}
	}

	if book, err := registerBook(TypeReceived, InOnHand, "SCB01"); err != nil || book != "SCB01" {
		t.Fatalf("on-hand received cheque keeps the book it was given: %q, %v", book, err)
	}
	if _, err := registerBook(TypeReceived, InDeposited, ""); err == nil {
		t.Fatal("deposited received cheque needs a book")
	} else {
		wantErrorKey(t, err, "cheque_err_book_required")
	}
	if book, err := registerBook(TypeReceived, InCancelled, ""); err != nil || book != "" {
		t.Fatalf("cancelled on-hand cheque needs no book: %q, %v", book, err)
	}
	if book, err := registerBook(TypeIssued, OutCleared, "SCB01"); err != nil || book != "SCB01" {
		t.Fatalf("issued cheque keeps its book: %q, %v", book, err)
	}
}
