package httpapi

import (
	"context"
	"database/sql"
	"net/http"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/shopspring/decimal"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	gl "smlcloudplatform/internal/generalledger"
	glhttp "smlcloudplatform/internal/generalledger/httpapi"
)

// ประเภทบรรทัด (คอลัมน์ S) มาจากประเภทเอกสารหนี้เท่านั้น: ERP ดู prefix ของ id บัญชีย่อย, จอบัญชีดู document_kind (คำตัดสิน D12)
// ลดหนี้เป็นยอดลบในรวมเป็นเงิน; ไม่รู้ประเภท = เจ้าหนี้อื่นๆ
func TestAPBillTypeFollowsDocumentType(t *testing.T) {
	for _, tc := range []struct {
		id             string
		fromGL         bool
		kind           int
		billType, sign int
	}{
		{"AP-PUR6910-0001", false, 0, billTypePurchase, 1},
		{"APDN-APDN6910-0001", false, 0, billTypeDebitNote, 1},
		{"APOD-APOD6910-0001", false, 0, billTypeOtherDebt, 1},
		{"GLAP-0007", true, 5, billTypeBeginning, 1},
		{"GLAP-0008", true, 3, billTypeDebitNote, 1},
		{"GLAP-0009", true, 4, billTypeCreditNote, -1},
		{"GLAP-0010", true, 1, billTypeOtherDebt, 1},
		{"GLAP-0011", true, 9, billTypeOtherDebt, 1},
	} {
		billType, sign := apBillType(tc.id, tc.fromGL, tc.kind)
		if billType != tc.billType || sign != tc.sign {
			t.Errorf("apBillType(%q, %v, %d) = %d,%d; want %d,%d", tc.id, tc.fromGL, tc.kind, billType, sign, tc.billType, tc.sign)
		}
		if billTypeKeys[billType] == "" {
			t.Errorf("billtype %d has no language key", billType)
		}
	}
}

func validPrepareInput() apPrepareInput {
	return apPrepareInput{PayDate: "2026-10-15", PartnerCode: " S0012 ", DiscountAmount: "150.25",
		Lines: []apPrepareLineInput{{DebtDocumentID: "AP-PUR6910-0001", PayAmount: "12500.50"}, {DebtDocumentID: "APDN-APDN6910-0002", PayAmount: "800"}}}
}

// ตรวจข้อมูลก่อนแตะฐานข้อมูล: ทุกข้อผิดพลาดเป็น 400 พร้อมช่องที่ต้องแก้; ยอดเงินรับเฉพาะสตริงทศนิยมไม่เกิน 2 ตำแหน่ง
func TestValidatePrepareInputRejectsBadInputWithField(t *testing.T) {
	for _, tc := range []struct {
		name, docNo, code, field string
		edit                     func(*apPrepareInput)
	}{
		{"docno required", "", "ap_prepare_docno_required", "docno", func(*apPrepareInput) {}},
		{"pay date invalid", "PP6910-0001", "source_date_invalid", "paydate", func(in *apPrepareInput) { in.PayDate = "15/10/2569" }},
		{"creditor required", "PP6910-0001", "ap_prepare_creditor_required", "partner_code", func(in *apPrepareInput) { in.PartnerCode = "  " }},
		{"discount three decimals", "PP6910-0001", "ap_prepare_discount_invalid", "discount_amount", func(in *apPrepareInput) { in.DiscountAmount = "1.234" }},
		{"discount negative", "PP6910-0001", "ap_prepare_discount_invalid", "discount_amount", func(in *apPrepareInput) { in.DiscountAmount = "-5" }},
		{"no lines", "PP6910-0001", "ap_prepare_no_lines", "lines", func(in *apPrepareInput) { in.Lines = nil }},
		{"debt missing", "PP6910-0001", "ap_prepare_debt_not_found", "lines[1].debt_document_id", func(in *apPrepareInput) { in.Lines[1].DebtDocumentID = " " }},
		{"duplicate debt", "PP6910-0001", "ap_prepare_line_duplicate", "lines[1].debt_document_id", func(in *apPrepareInput) { in.Lines[1].DebtDocumentID = " AP-PUR6910-0001" }},
		{"zero amount", "PP6910-0001", "ap_prepare_amount_invalid", "lines[0].payamount", func(in *apPrepareInput) { in.Lines[0].PayAmount = "0.00" }},
		{"negative amount", "PP6910-0001", "ap_prepare_amount_invalid", "lines[0].payamount", func(in *apPrepareInput) { in.Lines[0].PayAmount = "-1" }},
		{"three decimals", "PP6910-0001", "ap_prepare_amount_invalid", "lines[1].payamount", func(in *apPrepareInput) { in.Lines[1].PayAmount = "0.105" }},
		{"float notation", "PP6910-0001", "ap_prepare_amount_invalid", "lines[1].payamount", func(in *apPrepareInput) { in.Lines[1].PayAmount = "1e3" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in := validPrepareInput()
			tc.edit(&in)
			_, _, _, _, err := validatePrepareInput(in, tc.docNo)
			user, ok := gl.AsUserError(err)
			if !ok || user.Code != tc.code || user.Field != tc.field || user.Status != http.StatusBadRequest {
				t.Fatalf("validatePrepareInput error = %#v, want code %q field %q status 400", err, tc.code, tc.field)
			}
		})
	}
}

// เลขที่ใบเตรียมจ่ายเป็น segment ของ path API: ชุดอักขระเดียวกับ BFF และห้ามชนชื่อ path ของรายการ (list/outstanding/creditors);
// คำค้นตัดตามตัวอักษร ไม่ตัดกลางอักษรไทย (byte ไม่ครบ = PostgreSQL ปฏิเสธ)
func TestPrepareDocNoAndSearchStayRoutable(t *testing.T) {
	for _, docNo := range []string{"PP/6910", "PP 6910", "PP?1", "PP%1", "list", "outstanding", "creditors", ".", ".."} {
		_, _, _, _, err := validatePrepareInput(validPrepareInput(), docNo)
		user, ok := gl.AsUserError(err)
		if !ok || user.Code != "ap_prepare_docno_invalid" || user.Field != "docno" || user.Status != http.StatusBadRequest {
			t.Errorf("docno %q error = %v, want ap_prepare_docno_invalid", docNo, err)
		}
	}
	for _, docNo := range []string{"PP6910-0001", "ที่จ่าย_6910.01", "List"} {
		if _, _, _, _, err := validatePrepareInput(validPrepareInput(), docNo); err != nil {
			t.Errorf("docno %q error = %v, want accepted", docNo, err)
		}
	}
	got := gl.SearchPattern(" " + strings.Repeat("ก", 150) + strings.Repeat("ข", 100))
	if !utf8.ValidString(got) || got != "%"+strings.Repeat("ก", 150)+strings.Repeat("ข", 50)+"%" {
		t.Errorf("SearchPattern = %q, want 200 whole characters", got)
	}
	if got := gl.SearchPattern("  "); got != "" {
		t.Errorf("SearchPattern(blank) = %q, want no filter", got)
	}
}

func TestValidatePrepareInputKeepsExactAmounts(t *testing.T) {
	payDay, partner, discount, amounts, err := validatePrepareInput(validPrepareInput(), "PP6910-0001")
	if err != nil {
		t.Fatal(err)
	}
	if businessDay(payDay) != "2026-10-15" || partner != "S0012" || discount.StringFixed(2) != "150.25" {
		t.Errorf("header = %s %q %s", businessDay(payDay), partner, discount.StringFixed(2))
	}
	if len(amounts) != 2 || amounts[0].StringFixed(2) != "12500.50" || amounts[1].StringFixed(2) != "800.00" {
		t.Errorf("amounts = %v", amounts)
	}
}

// สิทธิ์ของจอเตรียมจ่ายใช้เมนู ap-payment-prepare; ยกเลิกใช้สิทธิ์ cancel หรือ delete ได้ (ปุ่ม Del ของ Champ)
func TestAllowsPrepareUsesItsOwnMenu(t *testing.T) {
	for _, tc := range []struct {
		name        string
		permissions []string
		action      string
		want        bool
	}{
		{"owner wildcard", []string{"*"}, actionCreate, true},
		{"menu entry reads", []string{prepareMenu}, actionRead, true},
		{"menu entry alone cannot create", []string{prepareMenu}, actionCreate, false},
		{"update granted", []string{prepareMenu, prepareMenu + ":update"}, actionUpdate, true},
		{"delete grants cancel", []string{prepareMenu, prepareMenu + ":delete"}, actionCancel, true},
		{"create does not grant cancel", []string{prepareMenu, prepareMenu + ":create"}, actionCancel, false},
		{"payment menu does not open preparation", []string{"pay", "pay:create"}, actionRead, false},
		{"no permissions", nil, actionRead, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := allowsPrepare(glhttp.NewSessionAccess(gl.Scope{}, tc.permissions...), tc.action); got != tc.want {
				t.Fatalf("allowsPrepare(%v, %q) = %v, want %v", tc.permissions, tc.action, got, tc.want)
			}
		})
	}
}

func beginMockTx(t *testing.T) (*sql.Tx, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, _ := newRecordedMock(t)
	mock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	return tx, mock
}

func payDocWithLines(amounts ...string) *ErpTransactionDoc {
	doc := &ErpTransactionDoc{CustCode: "S0012"}
	for i, a := range amounts {
		doc.Details = append(doc.Details, ErpDetailItem{LineNumber: i + 1, DebtDocumentID: "AP-PUR6910-000" + string(rune('1'+i)),
			DocRef: "PUR6910-000" + string(rune('1'+i)), PayBillNo: "PP6910-0001", SumAmount: amt(a)})
	}
	return doc
}

func settlePay(tx *sql.Tx, doc *ErpTransactionDoc, total string) error {
	return settleLines(context.Background(), tx, testScope(), subledgerSpecs["pay"], doc, "PV6910-0003", "PAY-PV6910-0003",
		"S0012", "21110", "2026-10-15", "J-77", amt(total), time.Date(2026, 10, 15, 3, 0, 0, 0, time.UTC))
}

// ยอดตัดหนี้รวมทุกบรรทัดเกินยอดใบจ่ายต้องหยุดก่อนแตะฐานข้อมูล (ไม่ล็อกแถวหนี้ ไม่เขียนการตัดหนี้)
func TestSettleLinesRejectsLinesOverPaymentTotal(t *testing.T) {
	tx, mock := beginMockTx(t)
	err := settlePay(tx, payDocWithLines("700.00", "300.01"), "1000.00")
	user, ok := gl.AsUserError(err)
	if !ok || user.Code != "ap_pay_lines_over_total" || user.Status != http.StatusBadRequest ||
		len(user.Args) != 2 || user.Args[0] != "1000.01" || user.Args[1] != "1000.00" {
		t.Fatalf("settleLines error = %#v, want ap_pay_lines_over_total 1000.01 > 1000.00", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestSettleLinesRejectsNonPositiveLine(t *testing.T) {
	tx, _ := beginMockTx(t)
	err := settlePay(tx, payDocWithLines("500.00", "0"), "1000.00")
	if user, ok := gl.AsUserError(err); !ok || user.Field != "details[1].sumamount" {
		t.Fatalf("settleLines error = %#v, want amount error on details[1].sumamount", err)
	}
}

// บรรทัดที่ดึงจากใบเตรียมจ่ายต้องมียอดมากกว่า 0 แม้ใบจ่ายไม่ลงบัญชี (ไม่ถึง settleLines): ยอดติดลบทำให้ยอดที่ใช้ของใบเตรียมจ่ายลดลง
// แล้วใบจ่ายอื่นดึงเกินที่เตรียมไว้ได้ — ปฏิเสธก่อนแตะฐานข้อมูล
func TestCheckPayPrepareLinesRejectsNonPositiveLine(t *testing.T) {
	tx, _ := beginMockTx(t)
	doc := &ErpTransactionDoc{Details: []ErpDetailItem{
		{LineNumber: 1, DebtDocumentID: "AP-PU6910-0001", PayBillNo: "PP6910-0001", SumAmount: decimal.RequireFromString("100.00")},
		{LineNumber: 2, DebtDocumentID: "AP-PU6910-0002", PayBillNo: "PP6910-0001", SumAmount: decimal.RequireFromString("-500.00")},
	}}
	err := checkPayPrepareLines(context.Background(), tx, "01", doc, "PV6910-0001", nil)
	if user, ok := gl.AsUserError(err); !ok || user.Code != "erp_amount_invalid" || user.Field != "details[1].sumamount" {
		t.Fatalf("checkPayPrepareLines error = %#v, want erp_amount_invalid on details[1].sumamount", err)
	}
}

// หนี้ที่ไม่ใช่ของเจ้าหนี้/บัญชีคุมของใบจ่าย (แถวกลับมาแต่ไม่ตรงเงื่อนไข) ต้องปฏิเสธพร้อมชี้บรรทัด
func TestSettleLinesRejectsDebtOfAnotherCreditor(t *testing.T) {
	tx, mock := beginMockTx(t)
	mock.ExpectQuery(`FROM gl_subledger_documents WHERE company=\$1 AND id=ANY\(\$2\) ORDER BY id FOR UPDATE`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "docno", "matches"}).
			AddRow("AP-PUR6910-0001", "PUR6910-0001", true).AddRow("AP-PUR6910-0002", "PUR6910-0002", false))
	err := settlePay(tx, payDocWithLines("500.00", "400.00"), "900.00")
	user, ok := gl.AsUserError(err)
	if !ok || user.Code != "ap_pay_settle_mismatch" || user.Field != "details[1].debt_document_id" || user.Status != http.StatusBadRequest {
		t.Fatalf("settleLines error = %#v, want ap_pay_settle_mismatch on details[1]", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

const testPrepareNo = "PP6910-0001"

func expectPrepareDebtLock(mock sqlmock.Sqlmock, id, docNo string) {
	mock.ExpectQuery(`FROM gl_subledger_documents WHERE company=\$1 AND id=ANY\(\$2\) ORDER BY id FOR UPDATE`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "docno", "matches"}).AddRow(id, docNo, true))
}

func expectPrepareUse(mock sqlmock.Sqlmock, debtID, payAmount, settled, pending string) {
	mock.ExpectQuery(`FROM ap_payment_prepare_lines l CROSS JOIN LATERAL`).
		WillReturnRows(sqlmock.NewRows([]string{"docno", "debt_document_id", "linenumber", "payamount", "settled", "pending",
			"document_no", "document_date", "debt_amount", "invbalance", "billtype", "sign", "description"}).
			AddRow(testPrepareNo, debtID, 1, payAmount, settled, pending, "PUR6910-0001", "2026-10-01", "15000.00", "15000.00", 0, 1, "ซื้อเหล็กเส้น"))
}

// expectOpenDebt: used = ยอดที่ใบเตรียมจ่ายที่กำลังแก้ตัดหนี้แล้ว (as-of เดียวกับ remaining); APOpenItems อ่านสองรอบ
// (as-of แล้วยอด ณ ปัจจุบัน) — ทั้งสองรอบคืนแถวเดียวกัน = ไม่มีใบจ่ายลงวันที่หลัง as-of
func expectOpenDebt(mock sqlmock.Sqlmock, id, remaining, available, used string) {
	for range 2 {
		mock.ExpectQuery(`posted_alloc AS`).
			WillReturnRows(sqlmock.NewRows([]string{"id", "docno", "date", "due", "description", "kind", "from_gl",
				"amount", "remaining", "reserved", "available", "used", "total"}).
				AddRow(id, "PUR6910-0001", "2026-10-01", "2026-10-31", "ซื้อเหล็กเส้น", 0, false, "15000.00", remaining, "0.00", available, used, 1))
	}
}

func checkLines(tx *sql.Tx, isEdit bool, ids []string, amounts ...string) ([]preparedLine, error) {
	inputs := make([]apPrepareLineInput, len(ids))
	values := make([]decimal.Decimal, len(ids))
	for i, id := range ids {
		inputs[i] = apPrepareLineInput{DebtDocumentID: id, PayAmount: amounts[i]}
		values[i] = amt(amounts[i])
	}
	return checkPrepareLines(context.Background(), tx, testScope(), testPrepareNo, "S0012", "2026-10-15", inputs, values, isEdit)
}

// ยอดตัดจ่ายเกินยอดที่จ่ายได้ (ยอดคงเหลือหักที่ใบเตรียมจ่ายอื่นจองไว้) ต้องบอกยอดสูงสุด
func TestCheckPrepareLinesRejectsOverAvailable(t *testing.T) {
	tx, mock := beginMockTx(t)
	expectPrepareDebtLock(mock, "AP-PUR6910-0001", "PUR6910-0001")
	expectOpenDebt(mock, "AP-PUR6910-0001", "15000.00", "9000.00", "0")
	_, err := checkLines(tx, false, []string{"AP-PUR6910-0001"}, "9000.01")
	user, ok := gl.AsUserError(err)
	if !ok || user.Code != "ap_prepare_over_balance" || user.Field != "lines[0].payamount" || len(user.Args) != 2 || user.Args[1] != "9000.00" {
		t.Fatalf("checkPrepareLines error = %#v, want ap_prepare_over_balance limit 9000.00", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

// คำตัดสิน D11: ใบที่จ่ายไปบางส่วนแก้ได้ แต่ลดยอดบรรทัดต่ำกว่ายอดที่ใบจ่ายใช้ไปแล้วไม่ได้
func TestCheckPrepareLinesKeepsPaidAmountOnEdit(t *testing.T) {
	tx, mock := beginMockTx(t)
	expectPrepareDebtLock(mock, "AP-PUR6910-0001", "PUR6910-0001")
	expectPrepareUse(mock, "AP-PUR6910-0001", "5000.00", "3000.00", "500.00")
	expectOpenDebt(mock, "AP-PUR6910-0001", "12000.00", "12000.00", "3000.00")
	_, err := checkLines(tx, true, []string{"AP-PUR6910-0001"}, "3499.99")
	user, ok := gl.AsUserError(err)
	if !ok || user.Code != "ap_prepare_below_paid" || user.Field != "lines[0].payamount" || len(user.Args) != 2 || user.Args[1] != "3500.00" {
		t.Fatalf("checkPrepareLines error = %#v, want ap_prepare_below_paid 3500.00", err)
	}
}

// ยอดที่ใบนี้ตัดหนี้แล้ว (as-of เดียวกับยอดคงเหลือ) บวกคืนเข้าเพดาน: ตัดแล้ว 3,000 + จ่ายได้ 12,000 = เพิ่มได้ถึง 15,000
func TestCheckPrepareLinesAllowsRaiseUpToSettledPlusAvailable(t *testing.T) {
	tx, mock := beginMockTx(t)
	expectPrepareDebtLock(mock, "AP-PUR6910-0001", "PUR6910-0001")
	expectPrepareUse(mock, "AP-PUR6910-0001", "5000.00", "3000.00", "0.00")
	expectOpenDebt(mock, "AP-PUR6910-0001", "12000.00", "12000.00", "3000.00")
	lines, err := checkLines(tx, true, []string{"AP-PUR6910-0001"}, "15000.00")
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 1 || lines[0].PayAmount != "15000.00" || lines[0].BillType != billTypePurchase || lines[0].Sign != 1 || lines[0].Description != "ซื้อเหล็กเส้น" {
		t.Errorf("prepared line = %+v", lines)
	}
}

// เอาบรรทัดที่ใบจ่ายใช้ไปแล้วออกไม่ได้ (เท่ากับลดยอดเป็น 0)
func TestCheckPrepareLinesRejectsRemovingPaidLine(t *testing.T) {
	tx, mock := beginMockTx(t)
	expectPrepareDebtLock(mock, "APOD-APOD6910-0004", "APOD6910-0004")
	expectPrepareUse(mock, "AP-PUR6910-0001", "5000.00", "1200.00", "0.00")
	_, err := checkLines(tx, true, []string{"APOD-APOD6910-0004"}, "800.00")
	if code := userErrorCode(err); code != "ap_prepare_below_paid" {
		t.Fatalf("checkPrepareLines error = %v, want ap_prepare_below_paid", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

// ใบที่จ่ายครบทุกบรรทัดแล้วแก้ไม่ได้
func TestCheckPrepareLinesLocksFullyPaid(t *testing.T) {
	tx, mock := beginMockTx(t)
	expectPrepareDebtLock(mock, "AP-PUR6910-0001", "PUR6910-0001")
	expectPrepareUse(mock, "AP-PUR6910-0001", "5000.00", "4000.00", "1000.00")
	_, err := checkLines(tx, true, []string{"AP-PUR6910-0001"}, "5000.00")
	if code := userErrorCode(err); code != "ap_prepare_paid_locked" {
		t.Fatalf("checkPrepareLines error = %v, want ap_prepare_paid_locked", err)
	}
}
