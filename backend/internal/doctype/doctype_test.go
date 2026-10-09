package doctype

import (
	"reflect"
	"regexp"
	"sort"
	"testing"
)

var (
	keyPattern       = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)
	prefixPattern    = regexp.MustCompile(`^[A-Z][A-Z-]{1,6}$`)
	newPrefixPattern = regexp.MustCompile(`^[A-Z]{2,4}$`)
)

func TestRegistryKeysFlagsAndPrefixesAreUnique(t *testing.T) {
	keys := map[string]bool{}
	flags := map[int]string{}
	prefixes := map[string]string{}
	for _, dt := range registry {
		if !keyPattern.MatchString(dt.Key) {
			t.Errorf("key %q must be lowercase ascii", dt.Key)
		}
		if keys[dt.Key] {
			t.Errorf("duplicate key %q", dt.Key)
		}
		keys[dt.Key] = true
		if dt.Flag <= 0 {
			t.Errorf("%s: flag %d must be positive", dt.Key, dt.Flag)
		}
		if other, dup := flags[dt.Flag]; dup {
			t.Errorf("flag %d shared by %s and %s", dt.Flag, other, dt.Key)
		}
		flags[dt.Flag] = dt.Key
		if !prefixPattern.MatchString(dt.Prefix) {
			t.Errorf("%s: prefix %q must be uppercase ascii", dt.Key, dt.Prefix)
		}
		if other, dup := prefixes[dt.Prefix]; dup {
			t.Errorf("prefix %q shared by %s and %s", dt.Prefix, other, dt.Key)
		}
		prefixes[dt.Prefix] = dt.Key
	}
}

func TestNewFlagsAreContiguousFrom400WithShortPrefixes(t *testing.T) {
	var newFlags []int
	for _, dt := range registry {
		if dt.Flag < 400 {
			continue
		}
		if dt.Flag > 499 {
			t.Errorf("%s: new flag %d outside 400-499", dt.Key, dt.Flag)
		}
		// chequeissued ย้ายเลขมาจาก 62 จึงคง prefix เดิม
		if dt.Key != "chequeissued" && !newPrefixPattern.MatchString(dt.Prefix) {
			t.Errorf("%s: new prefix %q must be 2-4 uppercase letters", dt.Key, dt.Prefix)
		}
		newFlags = append(newFlags, dt.Flag)
	}
	sort.Ints(newFlags)
	for i, flag := range newFlags {
		if flag != 400+i {
			t.Fatalf("new flags must run 400..%d without gaps, got %v", 400+len(newFlags)-1, newFlags)
		}
	}
	if len(newFlags) != 34 {
		t.Errorf("expected 34 new document types, got %d", len(newFlags))
	}
}

func TestGuardedAndRemovedKeysAreNotRegistered(t *testing.T) {
	absent := []string{
		// ยังไม่เปิดเป็นเอกสาร
		"etax", "withholdingtax", "purchasereceiptadjustment", "petty-cash-limit", "pettycashlimit",
		"contacts", "bankcontacts", "bank-contacts",
		// ชื่อซ้ำ/ชื่อผิดที่ลบแล้ว
		"receipt", "payment", "purchasereturn", "salereturn", "creditnote", "debitnote",
		"paidadvance", "paidadvancerefund", "chequereceive", "chequepayment", "bank",
		"depositrefund", "receivedepositrefund", "stock-prickup-product",
	}
	for _, key := range absent {
		if _, ok := ByKey(key); ok {
			t.Errorf("key %q must not be registered", key)
		}
	}
}

func TestNamedConstantsBelongToTheirKeys(t *testing.T) {
	want := map[string]int{
		"purchase-order":       PurchaseOrder,
		"purchase":             Purchase,
		"purchase-return":      PurchaseReturn,
		"stock-pickup-product": StockPickup,
		"expense":              Expense,
		"paid":                 Paid,
		"pay":                  Pay,
		"sale-order":           SaleOrder,
		"sale-invoice":         SaleInvoice,
		"sale-invoice-return":  SaleInvoiceReturn,
		"stock-transfer":       StockTransfer,
		"goodsreceive":         GoodsReceive,
	}
	for key, flag := range want {
		dt, ok := ByKey(key)
		if !ok || dt.Flag != flag {
			t.Errorf("ByKey(%q) = %+v,%v; want flag %d", key, dt, ok, flag)
		}
		back, ok := ByFlag(flag)
		if !ok || back.Key != key {
			t.Errorf("ByFlag(%d) = %q; want %q", flag, back.Key, key)
		}
	}
}

func TestKeptLegacyFlagNumbers(t *testing.T) {
	// เลขเดิมคงไว้ เลขที่เคยชนกันให้ประเภทที่มีก่อนคงเลขเดิม
	want := map[string]int{
		"purchase-requisition": 2, "purchase-order": 6, "purchase": 12, "purchase-return": 16,
		"stock-pickup-product": 20, "stock-return-product": 21, "deposit": 22, "expense": 24,
		"billingnote": 26, "apbillingreceipt": 27, "paid": 28, "pay": 29, "advancepayment": 30,
		"advancepaymentrefund": 31, "receivedeposit": 32, "quotation": 34, "receiveadvancerefund": 35,
		"sale-order": 36, "sale-invoice": 44, "sale-invoice-return": 48, "ardebitnote": 50,
		"debtorbeginningbalance": 52, "creditorbeginningbalance": 53, "receivableother": 54,
		"stock-balance": 56, "chequereceived": 61, "stock-transfer": 62, "stock-receive-product": 72,
		"stock-adjustment": 310,
		// ประเภทที่เคยใช้เลขชนกับประเภทอื่น ย้ายไปช่วงใหม่
		"receiveadvance": 408, "temporaryreceipt": 409, "chequeissued": 418,
	}
	for key, flag := range want {
		if dt, ok := ByKey(key); !ok || dt.Flag != flag {
			t.Errorf("ByKey(%q).Flag = %d,%v; want %d", key, dt.Flag, ok, flag)
		}
	}
}

func TestStockEffectsOnlyOnStockDocuments(t *testing.T) {
	want := map[string]StockEffect{
		"stock-balance":         StockIn,
		"stock-receive-product": StockIn,
		"stock-return-product":  StockIn,
		"purchase":              StockIn,
		"goodsreceive":          StockIn,
		"sale-invoice-return":   StockIn,
		"stock-pickup-product":  StockOut,
		"sale-invoice":          StockOut,
		"purchase-return":       StockOut,
		"stock-transfer":        StockByLine,
		"stock-adjustment":      StockByLine,
	}
	var wantFlags []int
	for _, dt := range registry {
		effect, listed := want[dt.Key]
		if !listed {
			effect = StockNone
		} else {
			wantFlags = append(wantFlags, dt.Flag)
		}
		if dt.Stock != effect {
			t.Errorf("%s: stock effect %d, want %d", dt.Key, dt.Stock, effect)
		}
	}
	sort.Ints(wantFlags)
	if got := StockFlags(); !reflect.DeepEqual(got, wantFlags) {
		t.Errorf("StockFlags() = %v, want %v", got, wantFlags)
	}
}

func TestByDocNoUsesLongestPrefix(t *testing.T) {
	cases := map[string]string{
		"PRET6910-0001":   "purchase-return",
		"PR6910-0001":     "purchase-requisition",
		"PUR6910-0001":    "purchase",
		"PO6910-0001":     "purchase-order",
		"PO-001":          "purchase-order",
		"RET6910-0001":    "stock-return-product",
		"RE6910-0001":     "paid",
		"RDEP6910-0001":   "receivedeposit",
		"OUTR6910-0001":   "stock-issue-request",
		"OUT6910-0001":    "stock-pickup-product",
		"CCS6910-0001":    "creditcardsettle",
		"CC6910-0001":     "creditcard",
		"AR-BAL6910-0001": "debtorbeginningbalance",
		"CHQ-OUT6910-001": "chequeissued",
		"INV":             "sale-invoice",
	}
	for docNo, key := range cases {
		dt, ok := ByDocNo(docNo)
		if !ok || dt.Key != key {
			t.Errorf("ByDocNo(%q) = %q,%v; want %q", docNo, dt.Key, ok, key)
		}
	}
	for _, docNo := range []string{"", "6910-0001", "XYZ6910-0001", "PRX6910-0001"} {
		if dt, ok := ByDocNo(docNo); ok {
			t.Errorf("ByDocNo(%q) = %q; want no match", docNo, dt.Key)
		}
	}
}

func TestAllReturnsCopy(t *testing.T) {
	all := All()
	if len(all) != len(registry) {
		t.Fatalf("All() returned %d rows, want %d", len(all), len(registry))
	}
	all[0].Flag = -1
	if registry[0].Flag == -1 {
		t.Fatal("All() must not expose the registry slice")
	}
}

// ทุกประเภทต้องผูกกับ menu id เพื่อให้ backend บังคับสิทธิ์รายเมนู; menu id ซ้ำกันข้ามประเภทไม่ได้
// (จะเปิดสองประเภทด้วยสิทธิ์เดียว) ยกเว้น alias ที่ประกาศไว้
func TestEveryTypeHasItsOwnPermissionScreen(t *testing.T) {
	owner := map[string]string{}
	for _, dt := range registry {
		if !keyPattern.MatchString(dt.Screen) {
			t.Errorf("%s: screen %q must be a lowercase menu id", dt.Key, dt.Screen)
		}
		screens := Screens(dt.Key)
		if len(screens) == 0 || screens[0] != dt.Screen {
			t.Errorf("%s: Screens=%v want primary %q first", dt.Key, screens, dt.Screen)
		}
		for _, screen := range screens {
			if other, dup := owner[screen]; dup {
				t.Errorf("screen %q shared by %s and %s", screen, other, dt.Key)
			}
			owner[screen] = dt.Key
		}
	}
	if got := Screens("stock-count"); !reflect.DeepEqual(got, []string{"stock-count", "stock-count-sheet"}) {
		t.Errorf("stock-count screens = %v", got)
	}
	if Screens("not-a-module") != nil {
		t.Error("unknown module must have no screens")
	}
}
