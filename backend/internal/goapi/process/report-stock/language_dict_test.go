package reportstock

import "testing"

func TestReportLanguageUsesSharedDictionary(t *testing.T) {
	if got := GetReportName("product_stock_movement", "ko"); got != "상품 재고 이동 보고서" {
		t.Fatalf("stock movement report name = %q", got)
	}
	if got := GetReportName("product_balance_by_warehouse", "zh"); got != "按仓库库存余额报告" {
		t.Fatalf("warehouse balance report name = %q", got)
	}
	if got := GetHeaderText("report_as_of", "jp"); got != "レポート日時" {
		t.Fatalf("report_as_of jp = %q", got)
	}
	if got := GetColumnText("increase_qty", "ko"); got != "증가 수량" {
		t.Fatalf("increase_qty ko = %q", got)
	}
}

func TestReportLanguageDictionaryShape(t *testing.T) {
	dict := GetReportLanguageDict("ko")
	columns, ok := dict["columns"].(map[string]string)
	if !ok {
		t.Fatalf("columns is %T", dict["columns"])
	}
	if got := columns["barcodelist"]; got != "바코드" {
		t.Fatalf("barcodelist ko = %q", got)
	}
}
