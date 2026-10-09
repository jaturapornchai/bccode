package generalledger

import "testing"

// รูปแบบลงบัญชีต้องได้เดบิต = เครดิตกับทุกยอดเอกสาร: ยอดรวม = มูลค่าสินค้า + ภาษี
func TestMappingBalances(t *testing.T) {
	rule := func(side, source string) MappingRule { return MappingRule{AccountCode: "1000", Side: side, Source: source} }
	for _, tc := range []struct {
		name  string
		rules []MappingRule
		want  bool
	}{
		{"sale: Dr total / Cr goods + Cr vat", []MappingRule{rule("debit", "total"), rule("credit", "goods"), rule("credit", "vat")}, true},
		{"purchase: Dr goods + Dr vat / Cr total", []MappingRule{rule("debit", "goods"), rule("debit", "vat"), rule("credit", "total")}, true},
		{"no vat: Dr total / Cr total", []MappingRule{rule("debit", "total"), rule("credit", "total")}, true},
		{"vat line missing on credit side", []MappingRule{rule("debit", "total"), rule("credit", "goods")}, false},
		{"goods twice on one side", []MappingRule{rule("debit", "total"), rule("credit", "goods"), rule("credit", "goods"), rule("credit", "vat")}, false},
		{"only one side", []MappingRule{rule("debit", "total"), rule("debit", "total")}, false},
		{"empty", nil, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := mappingBalances(tc.rules); got != tc.want {
				t.Fatalf("mappingBalances = %v, want %v", got, tc.want)
			}
		})
	}
}
