package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"smlcloudplatform/internal/doctype"
	gl "smlcloudplatform/internal/generalledger"
	"smlcloudplatform/pkg/apperr"
)

func dec(t *testing.T, value string) decimal.Decimal {
	t.Helper()
	d, err := decimal.NewFromString(value)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

// amt คือยอดทศนิยมคงที่ในตารางทดสอบ (ค่าผิดรูป = panic ทันที)
func amt(value string) decimal.Decimal { return decimal.RequireFromString(value) }

// ยอดเงินของเอกสารเป็น decimal ตั้งแต่ JSON — เข้าฝั่งบัญชีปัดเป็นทศนิยมสองตำแหน่งที่ระดับเอกสาร (Rule 20)
func TestPostingAmountsAreExactTwoDecimals(t *testing.T) {
	cases := []struct {
		name                      string
		doc                       ErpTransactionDoc
		wantTotal, wantVat, goods string
	}{
		{"0.1+0.2 stays exact", ErpTransactionDoc{TotalAmount: amt("0.1").Add(amt("0.2"))}, "0.30", "0.00", "0.30"},
		{"goods = total - vat exactly", ErpTransactionDoc{TotalAmount: amt("0.3"), TotalVatValue: amt("0.1")}, "0.30", "0.10", "0.20"},
		{"vat = total - before exactly", ErpTransactionDoc{TotalAmount: amt("0.3"), TotalBeforeVat: amt("0.1")}, "0.30", "0.20", "0.10"},
		{"rounds half up at 2dp", ErpTransactionDoc{TotalAmount: amt("107.005"), TotalVatValue: amt("7")}, "107.01", "7.00", "100.01"},
		{"just under the half stays down", ErpTransactionDoc{TotalAmount: amt("107.00499999"), TotalVatValue: amt("7")}, "107.00", "7.00", "100.00"},
		{"vat from before-vat difference", ErpTransactionDoc{TotalAmount: amt("5350"), TotalBeforeVat: amt("5000")}, "5350.00", "350.00", "5000.00"},
		{"explicit vat wins", ErpTransactionDoc{TotalAmount: amt("5350"), TotalVatValue: amt("350"), TotalBeforeVat: amt("5000")}, "5350.00", "350.00", "5000.00"},
		{"zero total", ErpTransactionDoc{}, "0.00", "0.00", "0.00"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := postingAmounts(&tc.doc)
			if err != nil {
				t.Fatal(err)
			}
			if got.total.StringFixed(2) != tc.wantTotal || got.vat.StringFixed(2) != tc.wantVat || got.goods.StringFixed(2) != tc.goods {
				t.Errorf("amounts = total %s vat %s goods %s; want %s %s %s",
					got.total.StringFixed(2), got.vat.StringFixed(2), got.goods.StringFixed(2), tc.wantTotal, tc.wantVat, tc.goods)
			}
			if !got.total.Equal(got.vat.Add(got.goods)) {
				t.Errorf("total %s must equal vat %s + goods %s", got.total, got.vat, got.goods)
			}
		})
	}
}

// ขอบการปัดที่ .005: ปัดครึ่งออกจากศูนย์ — ค่าที่ float64 เก็บไม่ตรง (1.005, 2.675) ต้องได้ผลเหมือนเลขฐานสิบ
func TestMoneyRoundsHalfAwayFromZeroAtTwoDecimals(t *testing.T) {
	for in, want := range map[string]string{
		"1.005": "1.01", "1.0049999": "1.00", "2.675": "2.68", "0.125": "0.13", "-1.005": "-1.01", "1070": "1070.00",
	} {
		if got := money(amt(in)).StringFixed(2); got != want {
			t.Errorf("money(%s) = %s, want %s", in, got, want)
		}
	}
}

func TestPostingAmountsRejectNonsense(t *testing.T) {
	for name, doc := range map[string]ErpTransactionDoc{
		"negative total": {TotalAmount: amt("-1")},
		"negative vat":   {TotalAmount: amt("100"), TotalVatValue: amt("-7")},
		"vat over total": {TotalAmount: amt("100"), TotalVatValue: amt("107")},
	} {
		doc := doc
		t.Run(name, func(t *testing.T) {
			_, err := postingAmounts(&doc)
			if userErrorCode(err) != "erp_amount_invalid" {
				t.Fatalf("postingAmounts error = %v, want erp_amount_invalid", err)
			}
		})
	}
}

// สัญญา JSON: ยอดเงินตอบเป็นข้อความทศนิยม; คำขอรับได้ทั้งตัวเลขและข้อความโดยไม่ผ่าน float; ช่องเสริมที่เป็นศูนย์ละไว้เหมือนเดิม
func TestDocJSONAmountsAreDecimalStrings(t *testing.T) {
	var doc ErpTransactionDoc
	body := `{"totalamount":0.30000000000000004,"totalbeforevat":"0.1","totalvatvalue":0.2,"vatrate":"7",
		"details":[{"qty":3,"price":"0.1","sumamount":0.3,"discountamount":"0"}]}`
	if err := json.Unmarshal([]byte(body), &doc); err != nil {
		t.Fatal(err)
	}
	if doc.TotalAmount.String() != "0.30000000000000004" {
		t.Errorf("number literal must be parsed exactly, got %s", doc.TotalAmount)
	}
	if !doc.TotalBeforeVat.Add(doc.TotalVatValue).Equal(amt("0.3")) {
		t.Errorf("0.1 + 0.2 = %s, want exactly 0.3", doc.TotalBeforeVat.Add(doc.TotalVatValue))
	}
	if !doc.Details[0].Qty.Mul(doc.Details[0].Price).Equal(doc.Details[0].SumAmount) {
		t.Errorf("3 x 0.1 = %s, want exactly %s", doc.Details[0].Qty.Mul(doc.Details[0].Price), doc.Details[0].SumAmount)
	}

	out, err := json.Marshal(ErpTransactionDoc{TotalAmount: amt("1070.00000000"), TotalVatValue: amt("70"),
		Details: []ErpDetailItem{{Qty: amt("10"), Price: amt("100.00"), SumAmount: amt("1000.00")}}})
	if err != nil {
		t.Fatal(err)
	}
	text := string(out)
	for _, want := range []string{`"totalamount":"1070"`, `"totalvatvalue":"70"`, `"qty":"10"`, `"price":"100"`, `"sumamount":"1000"`} {
		if !strings.Contains(text, want) {
			t.Errorf("JSON %s must contain %s", text, want)
		}
	}
	for _, absent := range []string{"totalbeforevat", "totalvalue", "totaldiscount", "totalaftervat", "vatrate", "discountamount"} {
		if strings.Contains(text, `"`+absent+`"`) {
			t.Errorf("zero %s must be omitted like before, JSON %s", absent, text)
		}
	}

	for _, bad := range []string{`{"totalamount":"NaN"}`, `{"totalamount":"12,000"}`, `{"totalamount":""}`, `{"details":[{"qty":"abc"}]}`} {
		var rejected ErpTransactionDoc
		if err := json.Unmarshal([]byte(bad), &rejected); err == nil {
			t.Errorf("%s must be rejected, got %+v", bad, rejected.TotalAmount)
		}
	}
}

// ขอบเขตตัวเลข: น้อยกว่า 1e10 (numeric(18,8)) ทศนิยมไม่เกิน 20 ตำแหน่ง — เลขชี้กำลังมหาศาลต้องถูกปฏิเสธโดยไม่ขยายค่า
func TestAmountInRangeBoundaries(t *testing.T) {
	for _, ok := range []string{"0", "0.1", "-9999999999.99999999", "9999999999.99999999", "1e9", "0.00000000000000000001"} {
		if !amountInRange(amt(ok)) {
			t.Errorf("%s must be accepted", ok)
		}
	}
	if !amountInRange(decimal.Decimal{}) {
		t.Error("zero value must be accepted")
	}
	for _, bad := range []string{"10000000000", "-10000000000", "1e10", "1e2000000000", "0e2000000000", "1e-2000000000",
		"0.000000000000000000001", "123456789012345678901234567890123456789"} {
		if amountInRange(amt(bad)) {
			t.Errorf("%s must be rejected", bad)
		}
	}
}

// คำขอที่มีตัวเลขเลขชี้กำลังมหาศาล (JSON ตัวเลขหรือข้อความ) ต้องได้ 400 ทันทีก่อนเปิดธุรกรรม — เดิมไหลไปถึง String()/Round
// ซึ่งขยายเป็น 10^2e9 จนเครื่องค้าง/หน่วยความจำหมด
func TestSaveDocRejectsHugeExponentAmountsFast(t *testing.T) {
	cases := []struct{ name, body, field string }{
		{"header-string", `{"totalamount":"1e2000000000"}`, "totalamount"},
		{"header-number", `{"totalamount":1e2000000000}`, "totalamount"},
		{"tiny-vatrate", `{"vatrate":"1e-2000000000"}`, "vatrate"},
		{"tiny-qty", `{"totalamount":"5","details":[{"qty":"1","price":"5","sumamount":"5"},{"qty":"1e-2000000000"}]}`, "details[1].qty"},
		{"price-over-limit", `{"details":[{"qty":"1","price":"99999999999","sumamount":"5"}]}`, "details[0].price"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, _ := newRecordedMock(t)
			var doc ErpTransactionDoc
			if err := json.NewDecoder(strings.NewReader(tc.body)).Decode(&doc); err != nil {
				t.Fatalf("decode: %v", err)
			}
			done := make(chan error, 1)
			go func() {
				_, err := (&Store{}).SaveDoc(context.Background(), db, testScope(), doctype.All()[0], &doc, false)
				done <- err
			}()
			var err error
			select {
			case err = <-done:
			case <-time.After(5 * time.Second):
				t.Fatal("SaveDoc hung on a huge exponent")
			}
			user, ok := gl.AsUserError(err)
			if !ok || user.Code != "erp_amount_out_of_range" || user.Field != tc.field || user.HTTPStatus() != http.StatusBadRequest {
				t.Fatalf("err = %v, want erp_amount_out_of_range on %s with 400", err, tc.field)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("must fail before opening a transaction: %v", err)
			}
			request := newPermContext("sale-invoice", "?lang=en")
			if docFailure(request, err, false); request.status != http.StatusBadRequest {
				t.Errorf("status = %d, want 400", request.status)
			}
			if response, ok := request.body.(apperr.Response); !ok || response.Message == "" || response.Message == "gl_err_erp_amount_out_of_range" {
				t.Errorf("response must carry translated text, got %#v", request.body)
			}
		})
	}
}

// รูปแบบลงบัญชีของลูกหนี้ (เดบิตลูกหนี้ = ยอดรวม, เครดิตรายได้ = สินค้า, เครดิตภาษีขาย = ภาษี) ต้องเดบิต = เครดิตเสมอ
func TestMappingLinesBalanceAndSkipZeroLines(t *testing.T) {
	rules := []gl.MappingRule{
		{AccountCode: "A-CONTROL", Side: gl.MappingSideDebit, Source: gl.MappingSourceTotal},
		{AccountCode: "A-REVENUE", Side: gl.MappingSideCredit, Source: gl.MappingSourceGoods},
		{AccountCode: "A-VAT", Side: gl.MappingSideCredit, Source: gl.MappingSourceVat},
	}
	withVat := docAmounts{total: dec(t, "5350.00"), vat: dec(t, "350.00"), goods: dec(t, "5000.00")}
	lines := mappingLines(rules, withVat, "ขายสินค้า")
	if len(lines) != 3 {
		t.Fatalf("lines = %+v, want 3", lines)
	}
	var debit, credit decimal.Decimal
	for _, line := range lines {
		debit = debit.Add(dec(t, string(line.Debit)))
		credit = credit.Add(dec(t, string(line.Credit)))
		if (line.Debit != "0") == (line.Credit != "0") {
			t.Errorf("a line is either debit or credit, got %+v", line)
		}
	}
	if !debit.Equal(credit) || debit.StringFixed(2) != "5350.00" {
		t.Errorf("debit %s credit %s, want both 5350.00", debit, credit)
	}

	noVat := docAmounts{total: dec(t, "100.00"), vat: decimal.Zero, goods: dec(t, "100.00")}
	if got := mappingLines(rules, noVat, ""); len(got) != 2 {
		t.Errorf("a zero VAT line must be skipped, got %+v", got)
	}
	if got := mappingLines(rules, docAmounts{}, ""); len(got) != 0 {
		t.Errorf("a zero document posts no lines, got %+v", got)
	}
}

func TestControlAccountComesFromTheMappingNotTheCode(t *testing.T) {
	rules := []gl.MappingRule{
		{AccountCode: "X1", Side: gl.MappingSideCredit, Source: gl.MappingSourceGoods},
		{AccountCode: "X2", Side: gl.MappingSideDebit, Source: gl.MappingSourceTotal},
	}
	if got := controlAccount(rules, gl.MappingSideDebit); got != "X2" {
		t.Errorf("debit control = %q, want X2", got)
	}
	if got := controlAccount(rules, gl.MappingSideCredit); got != "" {
		t.Errorf("a credit goods line is not a control line, got %q", got)
	}
}

// ทุกประเภทที่มีบัญชีย่อยต้องเป็นประเภทที่ลงบัญชีด้วย และรหัสเอกสารบัญชีย่อยต้องไม่ชนกัน
func TestSubledgerSpecsAreSubsetOfPostingTypesAndDistinct(t *testing.T) {
	posting := map[string]bool{}
	for _, key := range doctype.PostingKeys() {
		posting[key] = true
	}
	prefixes := map[string]string{}
	for key, spec := range subledgerSpecs {
		if !posting[key] {
			t.Errorf("subledger spec for %q, which does not post to the ledger", key)
		}
		if other, dup := prefixes[spec.docPrefix]; dup {
			t.Errorf("prefix %q is used by both %s and %s", spec.docPrefix, other, key)
		}
		prefixes[spec.docPrefix] = key
		if spec.ledger != "ar" && spec.ledger != "ap" {
			t.Errorf("%s: ledger %q", key, spec.ledger)
		}
		if spec.side != 1 && spec.side != 2 {
			t.Errorf("%s: side %d", key, spec.side)
		}
		if spec.settleSide != 0 && spec.settlePrefix == "" {
			t.Errorf("%s settles a debt but has no settlement prefix", key)
		}
	}
}

func TestGLRequestIDIsStableDistinctAndWithinLedgerLimits(t *testing.T) {
	a := glRequestID("01", 44, "INV6910-0001", 1, "create")
	if a != glRequestID("01", 44, "INV6910-0001", 1, "create") {
		t.Error("the same step must give the same request id so a retry returns the earlier result")
	}
	seen := map[string]bool{a: true}
	for _, id := range []string{
		glRequestID("02", 44, "INV6910-0001", 1, "create"),
		glRequestID("01", 48, "INV6910-0001", 1, "create"),
		glRequestID("01", 44, "INV6910-0002", 1, "create"),
		glRequestID("01", 44, "INV6910-0001", 2, "create"),
		glRequestID("01", 44, "INV6910-0001", 1, "post"),
		glRequestID("01", 44, "INV6910-0001", 1, "reverse"),
	} {
		if seen[id] {
			t.Errorf("request id %s repeats across different steps", id)
		}
		seen[id] = true
	}
	if len(a) < 16 || len(a) > 80 {
		t.Errorf("request id %q must be 16–80 chars", a)
	}
}

// สายสต็อกต้นทุนถัวเฉลี่ยเคลื่อนที่: รับเข้าใช้ราคาเอกสาร จ่ายออกใช้ต้นทุนเฉลี่ยก่อนหน้า มูลค่าไม่ติดลบ
func TestNextStockBalanceMovingAverage(t *testing.T) {
	zero := stockBalance{}
	// รับ 10 หน่วย × 100 = 1,000
	unit, amount, bal := nextStockBalance(1, dec(t, "10"), dec(t, "100"), dec(t, "1000"), zero)
	if !unit.Equal(dec(t, "100")) || !amount.Equal(dec(t, "1000")) || !bal.qty.Equal(dec(t, "10")) || !bal.avg.Equal(dec(t, "100")) {
		t.Fatalf("receipt: unit %s amount %s bal %+v", unit, amount, bal)
	}
	// รับเพิ่ม 10 หน่วย × 120 → ถัวเฉลี่ย 110
	_, _, bal = nextStockBalance(1, dec(t, "10"), dec(t, "120"), dec(t, "1200"), bal)
	if !bal.qty.Equal(dec(t, "20")) || !bal.amount.Equal(dec(t, "2200")) || !bal.avg.Equal(dec(t, "110")) {
		t.Fatalf("second receipt: %+v", bal)
	}
	// จ่าย 5 หน่วย (ราคาขาย 150) ตัดที่ต้นทุนเฉลี่ย 110
	unit, amount, bal = nextStockBalance(-1, dec(t, "5"), dec(t, "150"), dec(t, "750"), bal)
	if !unit.Equal(dec(t, "110")) || !amount.Equal(dec(t, "550")) || !bal.qty.Equal(dec(t, "15")) || !bal.amount.Equal(dec(t, "1650")) || !bal.avg.Equal(dec(t, "110")) {
		t.Fatalf("issue: unit %s amount %s bal %+v", unit, amount, bal)
	}
	// จ่ายเกินคงเหลือ: มูลค่าไม่ติดลบ
	_, _, bal = nextStockBalance(-1, dec(t, "100"), dec(t, "150"), dec(t, "15000"), bal)
	if bal.amount.IsNegative() {
		t.Errorf("stock value must floor at zero, got %s", bal.amount)
	}
	// จ่ายออกจากสายที่ไม่มีต้นทุน ใช้ราคาของเอกสาร
	unit, _, _ = nextStockBalance(-1, dec(t, "1"), dec(t, "80"), dec(t, "80"), zero)
	if !unit.Equal(dec(t, "80")) {
		t.Errorf("issue without cost must use the document price, got %s", unit)
	}
}

// ข้อความที่ผู้ใช้เห็นจากฝั่ง ERP ต้องมีแถวใน languages.tsv ครบ 13 คอลัมน์ มีไทย+อังกฤษ (ไม่งั้นได้ key ดิบ)
func TestERPMessagesHaveLanguageRows(t *testing.T) {
	raw, err := os.ReadFile("../../../assets/language/languages.tsv")
	if err != nil {
		t.Fatal(err)
	}
	rows := map[string][]string{}
	for _, line := range strings.Split(string(raw), "\n") {
		cells := strings.Split(strings.TrimRight(line, "\r"), "\t")
		rows[cells[0]] = cells
	}

	keys := map[string]bool{"erp_gl_journal_desc": true, "erp_gl_reverse_edit": true, "erp_gl_reverse_delete": true,
		"erp_gl_notice_posted": true, "erp_gl_notice_no_mapping": true, "erp_gl_notice_no_amount": true}
	source, err := os.ReadFile("posting.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range regexp.MustCompile(`erpError\("(erp_[a-z_]+)"`).FindAllStringSubmatch(string(source), -1) {
		keys["gl_err_"+m[1]] = true
	}
	if len(keys) < 9 {
		t.Fatalf("expected at least 3 ERP error codes in posting.go, found keys %v", keys)
	}
	for _, dt := range doctype.All() {
		if doctype.PostsToGL(dt.Flag) {
			keys["gl_mapping_doc_"+strings.ReplaceAll(dt.Key, "-", "_")] = true
		}
	}

	sorted := make([]string, 0, len(keys))
	for key := range keys {
		sorted = append(sorted, key)
	}
	sort.Strings(sorted)
	for _, key := range sorted {
		cells, ok := rows[key]
		if !ok {
			t.Errorf("languages.tsv has no row %q", key)
			continue
		}
		if len(cells) != 13 {
			t.Errorf("row %q has %d columns, want 13", key, len(cells))
			continue
		}
		for i, cell := range cells {
			if strings.TrimSpace(cell) == "" {
				t.Errorf("row %q column %d is empty", key, i)
			}
		}
	}
}
