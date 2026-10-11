package glposting

import (
	"regexp"
	"testing"

	"github.com/shopspring/decimal"

	"smlcloudplatform/internal/cheque"
	"smlcloudplatform/internal/doctype"
	gl "smlcloudplatform/internal/generalledger"
)

func clearRules() []gl.MappingRule {
	return []gl.MappingRule{
		{AccountFrom: gl.MappingAccountFromBankBook, Side: gl.MappingSideDebit, Source: gl.MappingSourceChequeNet},
		{AccountCode: "53110", Side: gl.MappingSideDebit, Source: gl.MappingSourceChequeFee},
		{AccountCode: "11330", Side: gl.MappingSideCredit, Source: gl.MappingSourceChequeTotal},
	}
}

func posting(fee string) cheque.Posting {
	f := decimal.RequireFromString(fee)
	total := decimal.RequireFromString("15250.75")
	return cheque.Posting{Flag: cheque.FlagReceivedClear, DocNo: "CQIP6910-0001", BookCode: "SCB01", BookAccount: "11120", Total: total, Fee: f, Net: total.Sub(f)}
}

func TestMappingLinesUseBankBookAccountAndStayBalanced(t *testing.T) {
	lines, err := mappingLines(clearRules(), posting("0.10"), "บันทึกเช็ครับผ่าน เลขที่ CQIP6910-0001")
	if err != nil {
		t.Fatal(err)
	}
	want := []gl.Line{
		{AccountCode: "11120", Debit: "15250.65", Credit: "0"},
		{AccountCode: "53110", Debit: "0.10", Credit: "0"},
		{AccountCode: "11330", Debit: "0", Credit: "15250.75"},
	}
	if len(lines) != len(want) {
		t.Fatalf("lines = %+v", lines)
	}
	debit, credit := decimal.Zero, decimal.Zero
	for i, l := range lines {
		if l.AccountCode != want[i].AccountCode || l.Debit != want[i].Debit || l.Credit != want[i].Credit {
			t.Errorf("line %d = %+v, want %+v", i, l, want[i])
		}
		debit = debit.Add(decimal.RequireFromString(string(l.Debit)))
		credit = credit.Add(decimal.RequireFromString(string(l.Credit)))
	}
	if !debit.Equal(credit) {
		t.Fatalf("debit %s != credit %s", debit, credit)
	}
}

func TestMappingLinesSkipZeroFee(t *testing.T) {
	lines, err := mappingLines(clearRules(), posting("0"), "")
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 2 || lines[0].Debit != "15250.75" || lines[1].Credit != "15250.75" {
		t.Fatalf("lines = %+v", lines)
	}
}

func TestMappingLinesBankBookWithoutAccount(t *testing.T) {
	p := posting("0")
	p.BookAccount = ""
	_, err := mappingLines(clearRules(), p, "")
	e, ok := cheque.AsError(err)
	if !ok || e.Key != "cheque_err_book_no_gl_account" || e.Args[0] != "SCB01" {
		t.Fatalf("err = %v", err)
	}

	// ยกเลิกเช็ครับในมือไม่มีสมุด: ใช้บัญชีสำรองของรายการ (gltlsprocess.cpp:7824)
	rules := []gl.MappingRule{
		{AccountCode: "41900", AccountFrom: gl.MappingAccountFromBankBook, Side: gl.MappingSideDebit, Source: gl.MappingSourceChequeTotal},
		{AccountCode: "11330", Side: gl.MappingSideCredit, Source: gl.MappingSourceChequeTotal},
	}
	p = posting("0")
	p.Flag, p.BookCode, p.BookAccount, p.BookFallback = cheque.FlagReceivedCancel, "", "", true
	lines, err := mappingLines(rules, p, "")
	if err != nil || lines[0].AccountCode != "41900" {
		t.Fatalf("fallback lines = %+v, err %v", lines, err)
	}
	// ไม่มีบัญชีสำรองด้วย → ไม่ลงบัญชีครึ่งใบ
	rules[0].AccountCode = ""
	if _, err = mappingLines(rules, p, ""); err == nil {
		t.Fatal("rule without fallback account must fail when the document has no bank book")
	}
}

func TestRequestIDIsStablePerStep(t *testing.T) {
	a := requestID("01", cheque.FlagReceivedClear, "CQIP6910-0001", 1, actionCreate)
	if a != requestID("01", cheque.FlagReceivedClear, "CQIP6910-0001", 1, actionCreate) {
		t.Fatal("request id must be deterministic so a retry returns the first result")
	}
	if !regexp.MustCompile(`^chq-[0-9a-f]{40}$`).MatchString(a) {
		t.Fatalf("request id %q", a)
	}
	for _, other := range []string{
		requestID("01", cheque.FlagReceivedClear, "CQIP6910-0001", 2, actionCreate),
		requestID("01", cheque.FlagReceivedClear, "CQIP6910-0001", 1, actionPost),
		requestID("02", cheque.FlagReceivedClear, "CQIP6910-0001", 1, actionCreate),
		requestID("01", cheque.FlagDiscount, "CQIP6910-0001", 1, actionCreate),
	} {
		if other == a {
			t.Fatal("request ids of different steps collide")
		}
	}
}

func TestDocLabelIsThaiMenuName(t *testing.T) {
	if got := docLabel(mustDocType(t, cheque.FlagReceivedClear)); got == "" || got == "cheque_received_clear" {
		t.Fatalf("docLabel = %q", got)
	}
}

func mustDocType(t *testing.T, flag int) doctype.DocType {
	t.Helper()
	dt, ok := doctype.ByFlag(flag)
	if !ok {
		t.Fatalf("flag %d has no doctype", flag)
	}
	return dt
}
