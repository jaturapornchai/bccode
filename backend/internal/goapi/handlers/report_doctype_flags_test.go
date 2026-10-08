package handlers

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"smlcloudplatform/internal/doctype"
)

// รายงานขายต้องอ่านเฉพาะใบขายจากทะเบียนประเภทเอกสาร ทั้งเงื่อนไขหลักและเงื่อนไขกรองสินค้า
func TestSalesReportQueriesFilterSaleInvoiceFromRegistry(t *testing.T) {
	from := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, 0)
	products := []string{"P-1001"}
	main := fmt.Sprintf("WHERE p.transflag = %d\n", doctype.SaleInvoice)
	sub := fmt.Sprintf("WHERE sub.transflag = %d\n", doctype.SaleInvoice)

	header, _ := buildSalesReportHeaderQuery(from, to, "01", nil, products, true, 20, 0)
	detail, _ := buildSalesReportDetailQuery(from, to, "01", nil, products, true, 20, 0)
	summary, _ := buildSalesReportSummaryQuery(from, to, "01", nil, nil)
	for name, query := range map[string]string{"header": header, "detail": detail, "summary": summary} {
		if !strings.Contains(query, main) {
			t.Errorf("%s query missing %q", name, main)
		}
	}
	for name, query := range map[string]string{"header": header, "detail": detail} {
		if !strings.Contains(query, sub) {
			t.Errorf("%s query missing product filter %q", name, sub)
		}
	}
}

// ความเคลื่อนไหวลูกหนี้/เจ้าหนี้แสดงประเภทเอกสารตามทะเบียน: ตั้งหนี้จากใบขาย/ใบซื้อ ชำระจากใบเสร็จ/ใบจ่ายชำระหนี้
func TestDebtMovementReportsLabelDocumentTypesFromRegistry(t *testing.T) {
	cases := map[string]string{
		"ar_movement": fmt.Sprintf("CASE WHEN d.side = 1 THEN %d ELSE %d END AS transflag", doctype.SaleInvoice, doctype.Paid),
		"ap_movement": fmt.Sprintf("CASE WHEN d.side = 2 THEN %d ELSE %d END AS transflag", doctype.Purchase, doctype.Pay),
	}
	for reportCode, want := range cases {
		query, _, err := BuildDebtReportQuery(reportCode, "01", time.Time{}, time.Time{}, "", 20, 0)
		if err != nil {
			t.Fatalf("%s: %v", reportCode, err)
		}
		if !strings.Contains(query, want) {
			t.Errorf("%s query missing %q", reportCode, want)
		}
	}
}
