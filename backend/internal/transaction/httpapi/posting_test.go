package httpapi

import (
	"math"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/shopspring/decimal"

	"smlcloudplatform/internal/doctype"
	gl "smlcloudplatform/internal/generalledger"
)

func dec(t *testing.T, value string) decimal.Decimal {
	t.Helper()
	d, err := decimal.NewFromString(value)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

// ยอดเงินของเอกสารรับมาเป็น float64 จาก JSON — เข้าฝั่งบัญชีต้องกลายเป็นทศนิยมสองตำแหน่งที่แน่นอนทันที (Rule 20)
func TestPostingAmountsAreExactTwoDecimals(t *testing.T) {
	cases := []struct {
		name                      string
		doc                       ErpTransactionDoc
		wantTotal, wantVat, goods string
	}{
		{"0.1+0.2 stays exact", ErpTransactionDoc{TotalAmount: 0.1 + 0.2, TotalVatValue: 0.0}, "0.30", "0.00", "0.30"},
		{"rounds half up at 2dp", ErpTransactionDoc{TotalAmount: 107.005, TotalVatValue: 7.0}, "107.01", "7.00", "100.01"},
		{"vat from before-vat difference", ErpTransactionDoc{TotalAmount: 5350, TotalBeforeVat: 5000}, "5350.00", "350.00", "5000.00"},
		{"explicit vat wins", ErpTransactionDoc{TotalAmount: 5350, TotalVatValue: 350, TotalBeforeVat: 5000}, "5350.00", "350.00", "5000.00"},
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

func TestPostingAmountsRejectNonsense(t *testing.T) {
	for name, doc := range map[string]ErpTransactionDoc{
		"negative total": {TotalAmount: -1},
		"negative vat":   {TotalAmount: 100, TotalVatValue: -7},
		"vat over total": {TotalAmount: 100, TotalVatValue: 107},
		"NaN":            {TotalAmount: math.NaN()},
		"Inf":            {TotalAmount: math.Inf(1)},
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
